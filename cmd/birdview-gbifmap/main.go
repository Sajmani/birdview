// Copyright 2026 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Command birdview-gbifmap builds the committed table that maps GBIF backbone
// scientific names onto AviList's.
//
// It is deliberately NOT part of birdview-ingest, for the reason tech.pacedInaturalistMapping gives
// for cmd/birdview-inatmap: a full run is about 2,800 requests to someone
// else's server, tests make no network request, and a build that needs the
// network is a build that fails on a plane. Run this when the bridge needs
// refreshing, read the report, commit the table.
//
// It also does not wire anything into the dataset. Whether BirdView's species
// scope should come from GBIF occurrence counts is an open product decision;
// this tool only establishes that the names can be joined, and measures how
// well.
//
// Verifies: tech.identifiedClient, tech.recordBeforeConsuming.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sort"
	"strings"
	"syscall"
	"time"

	"birdview/internal/avilist"
	"birdview/internal/gbif"
)

func main() {
	log.SetFlags(0)
	log.SetPrefix("birdview-gbifmap: ")

	var (
		avlPath     = flag.String("avilist", "third_party/avilist/AviList-v2025b-10Jun2026-extended.xlsx", "AviList checklist (extended release)")
		out         = flag.String("out", "data/gbif-taxa.tsv", "mapping table to write")
		dataset     = flag.String("dataset", gbif.EBirdDatasetKey, "GBIF occurrence dataset key")
		countryList = flag.String("countries", "US,CA", "comma-separated country codes to scope the species list to")
		base        = flag.String("api", gbif.DefaultBaseURL, "API base URL")
		interval    = flag.Duration("interval", gbif.MinInterval, "minimum delay between requests, across all workers")
		workers     = flag.Int("concurrency", gbif.DefaultConcurrency, "requests in flight at once")
	)
	flag.Parse()

	avl, err := avilist.Open(*avlPath)
	if err != nil {
		log.Fatalf("%v", err)
	}
	familyOf := make(map[string]string, len(avl))
	for name, sp := range avl {
		familyOf[name] = sp.Family
	}
	idx := gbif.NewIndex(familyOf)
	log.Printf("AviList: %d species", idx.Len())

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	c := gbif.NewClient()
	c.BaseURL, c.Interval, c.Concurrency = *base, *interval, *workers

	start := time.Now()
	countries := strings.Split(*countryList, ",")
	keys, records, err := c.SpeciesKeys(ctx, *dataset, countries)
	if err != nil {
		log.Fatalf("%v", err)
	}
	var faceted int64
	for _, k := range keys {
		faceted += k.Records
	}
	log.Printf("GBIF: %d species over %s records in %s (%s sit above species rank and are invisible here)",
		len(keys), commas(records), countries, commas(records-faceted))

	taxa, err := c.Taxa(ctx, keys)
	if err != nil {
		log.Fatalf("%v", err)
	}
	log.Printf("fetched %d backbone records in %s", len(taxa), time.Since(start).Round(time.Second))

	table := gbif.Table{}
	byTier := map[gbif.Tier][]gbif.Taxon{}
	for _, t := range taxa {
		e := gbif.Resolve(t, idx)
		if prev, dup := table[t.Name]; dup {
			log.Fatalf("GBIF returned %q twice, as keys %d and %d", t.Name, prev.GBIFKey, t.Key)
		}
		table[t.Name] = e
		byTier[e.Tier] = append(byTier[e.Tier], t)
	}

	f, err := os.Create(*out)
	if err != nil {
		log.Fatalf("creating %s: %v", *out, err)
	}
	if err := gbif.Write(f, table); err != nil {
		f.Close()
		log.Fatalf("writing %s: %v", *out, err)
	}
	if err := f.Close(); err != nil {
		log.Fatalf("closing %s: %v", *out, err)
	}

	report(*out, table, byTier)
}

// report prints the tier-by-tier contribution and the whole unresolved
// residue. The residue is the part worth reading: a fuzzy match that is wrong
// is worse than an honest gap, so the gaps are listed by name rather than
// summarized away.
func report(path string, table gbif.Table, byTier map[gbif.Tier][]gbif.Taxon) {
	var resolved int
	for _, e := range table {
		if e.Tier.Resolved() {
			resolved++
		}
	}
	fmt.Printf("\nwrote %s: %d of %d GBIF species resolved to an AviList species\n",
		path, resolved, len(table))
	for _, tier := range []gbif.Tier{gbif.TierExact, gbif.TierSynonym, gbif.TierEpithet, gbif.TierAmbiguous, gbif.TierNone} {
		fmt.Printf("  %-10s %4d\n", tier, len(byTier[tier]))
	}

	// Everything resolved below the exact tier, named, because those are the
	// rows a reviewer has to agree with.
	for _, tier := range []gbif.Tier{gbif.TierEpithet, gbif.TierAmbiguous, gbif.TierNone} {
		ts := byTier[tier]
		if len(ts) == 0 {
			continue
		}
		sort.Slice(ts, func(i, j int) bool { return ts[i].Records > ts[j].Records })
		fmt.Printf("\n%s (%d):\n", tier, len(ts))
		for _, t := range ts {
			fmt.Printf("  %-32s %-20s %12s records -> %s\n",
				t.Name, t.Family, commas(t.Records), table[t.Name].AviList)
		}
	}
}

func commas(n int64) string {
	s := fmt.Sprintf("%d", n)
	if n < 0 {
		return s
	}
	var b strings.Builder
	for i, r := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			b.WriteByte(',')
		}
		b.WriteRune(r)
	}
	return b.String()
}
