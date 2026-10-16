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

// Command birdview-inatmap resolves BirdView's species to iNaturalist taxon
// IDs and writes the committed mapping table.
//
// It is deliberately NOT part of birdview-ingest. At the one-request-per-second
// iNaturalist asks for (inat-api/R1), 841 species is about a quarter of an
// hour, and a build that needs the network is a build that fails on a plane.
// Run this when the mapping needs refreshing; commit the result; the ingest
// joins against it offline (tech.pacedInaturalistMapping, decisions.inaturalistLinkAdopted).
//
// Verifies: product.inaturalistLink, tech.pacedInaturalistMapping.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sort"
	"syscall"
	"time"

	"birdview/internal/dataset"
	"birdview/internal/inat"
	"birdview/internal/taxonomy"
)

func main() {
	log.SetFlags(0)
	log.SetPrefix("birdview-inatmap: ")

	var (
		in       = flag.String("data", "web/data/species.json", "compiled dataset to read species from")
		out      = flag.String("out", "data/inat-taxa.tsv", "mapping table to write")
		interval = flag.Duration("interval", inat.MinInterval, "minimum delay between requests (inat-api/R1)")
		base     = flag.String("api", inat.DefaultBaseURL, "API base URL")
	)
	flag.Parse()

	ds, err := load(*in)
	if err != nil {
		log.Fatalf("reading %s: %v (run birdview-ingest first)", *in, err)
	}

	// Ctrl-C writes out what has been resolved so far rather than discarding a
	// quarter of an hour of someone else's server time.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	c := inat.NewClient()
	c.BaseURL, c.Interval = *base, *interval

	species := make([]dataset.Species, len(ds.Species))
	copy(species, ds.Species)
	sort.Slice(species, func(i, j int) bool { return species[i].ScientificName < species[j].ScientificName })

	table := inat.Table{}
	var missed []string
	start := time.Now()

	for i := range species {
		s := &species[i]
		if ctx.Err() != nil {
			log.Printf("interrupted after %d of %d; writing what resolved", i, len(species))
			break
		}
		e, ok, err := c.Lookup(ctx, s.ScientificName, s.Name(dataset.LocaleNA))
		// iNaturalist predates AviList's renames, as AVONET and the GBIF
		// occurrence data both do: it has no Pica nuttallii and no Icterus
		// bullockiorum, only the spellings those were renamed from. Retry
		// under the older label rather than record a confident non-match,
		// which is what two of 1,083 species got on the first full run.
		if err == nil && !ok {
			if from, renamed := taxonomy.AviListRenamedFrom[s.ScientificName]; renamed {
				e, ok, err = c.Lookup(ctx, from, s.Name(dataset.LocaleNA))
			}
		}
		switch {
		case err != nil:
			// A transport error is not a confident non-match, so it must not be
			// recorded as one. Report and carry on; a rerun fills the gap.
			log.Printf("  error  %-28s %v", s.ScientificName, err)
			missed = append(missed, s.ScientificName+" (error)")
		case !ok:
			missed = append(missed, s.ScientificName)
		default:
			table[s.ScientificName] = e
		}
		if n := i + 1; n%100 == 0 {
			log.Printf("  %d/%d resolved=%d elapsed=%s", n, len(species), len(table),
				time.Since(start).Round(time.Second))
		}
	}

	f, err := os.Create(*out)
	if err != nil {
		log.Fatalf("creating %s: %v", *out, err)
	}
	if err := inat.Write(f, table); err != nil {
		f.Close()
		log.Fatalf("writing %s: %v", *out, err)
	}
	if err := f.Close(); err != nil {
		log.Fatalf("closing %s: %v", *out, err)
	}

	log.Printf("wrote %d of %d species to %s in %s", len(table), len(species), *out,
		time.Since(start).Round(time.Second))
	if len(missed) > 0 {
		log.Printf("%d without a confident match; these get no link (product.inaturalistLink):", len(missed))
		for _, m := range missed {
			fmt.Fprintf(os.Stderr, "    %s\n", m)
		}
	}
}

// load reads the compiled dataset. The server has its own copy of this with
// different error handling; sharing it would couple a build tool to the
// serving path for three lines.
func load(path string) (*dataset.Dataset, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var ds dataset.Dataset
	if err := json.NewDecoder(f).Decode(&ds); err != nil {
		return nil, err
	}
	return &ds, nil
}
