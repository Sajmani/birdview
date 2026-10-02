// Command birdview-regions builds the per-region presence table.
//
// For each US state and Canadian province it asks GBIF which bird names were
// reported there and how often, then writes the regions each of BirdView's
// species occurs in. The ingest reads the result; nothing queries the network
// at build or test time (T-006, and the no-network-in-tests binding).
//
// Two sources, because one is not enough:
//
//   - eBird, for depth: 1.24 billion records across the ABA Area.
//   - iNaturalist, because eBird's export WITHHOLDS conservation-sensitive
//     species entirely. Spotted Owl, Gyrfalcon, Northern Hawk Owl and the
//     critically endangered Hawaiian endemics have zero eBird records and
//     hundreds in iNaturalist. An eBird-only table omits exactly the birds a
//     reader is most likely to look up. See sources.md, `gbif`.
//
// Names are faceted VERBATIM — as the publisher wrote them — not through
// GBIF's backbone, which drops what it cannot resolve and reported zero
// records for Cooper's Hawk.
//
// Verifies: P-011.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"birdview/internal/gbif"
)

// iNatDatasetKey is the iNaturalist research-grade observations dataset.
const iNatDatasetKey = "50c9509d-22c7-4a22-a47d-8c48425ef4a7"

// usStates and caProvinces are GADM level-1 counts. GADM numbers them from 1
// in alphabetical order, so the ids are generated rather than listed; each
// one's real name is read back from the data and written into the table, so a
// wrong id shows up as a wrong name rather than as silence.
const (
	usStates    = 51 // 50 plus the District of Columbia
	caProvinces = 13 // 10 provinces, 3 territories
	// GADM files the Northwestern Hawaiian Islands under United States Minor
	// Outlying Islands, not under Hawaii, so Laysan Duck and Laysan Finch land
	// in no state and were left unplaced until this was added. A unit that
	// names nothing is skipped, so over-asking is free.
	umiUnits = 10
)

// minPPM is the presence bar: a species counts as present in a region when it
// reaches this share of that region's bird records.
//
// Relative, not absolute, for two reasons measured in sources.md. eBird's
// volume doubled in five years, so any fixed count admits more species over
// time for no ecological reason. And regions differ hugely in effort —
// California has 114 million records against Hawaii's 2.5 million — so one
// count cannot mean the same thing in both.
//
// 1 ppm gives roughly 600 species for California, 390 for Pennsylvania and
// 267 for Hawaii, which are the right orders of magnitude for a state list.
const minPPM = 1.0

func main() {
	var (
		dataPath = flag.String("data", "web/data/species.json", "compiled dataset, for the species to look up")
		outPath  = flag.String("out", "data/region-presence.tsv", "output table (committed)")
	)
	flag.Parse()

	ds, err := loadSpecies(*dataPath)
	if err != nil {
		log.Fatalf("birdview-regions: %v", err)
	}
	log.Printf("birdview-regions: %d species to place", len(ds))

	c := gbif.NewClient()
	// Gentler than the bridge's defaults. These facet queries are far heavier
	// than a taxon lookup -- each scans a region's whole occurrence set -- and
	// GBIF answered 429 at four workers on a 100 ms interval. Two workers at
	// 400 ms completes the run in well under a minute, which is not worth
	// optimizing against someone else's free service.
	c.Concurrency = 2
	c.Interval = 400 * time.Millisecond
	ctx := context.Background()

	var gids []string
	for i := 1; i <= usStates; i++ {
		gids = append(gids, fmt.Sprintf("USA.%d_1", i))
	}
	for i := 1; i <= caProvinces; i++ {
		gids = append(gids, fmt.Sprintf("CAN.%d_1", i))
	}
	for i := 1; i <= umiUnits; i++ {
		gids = append(gids, fmt.Sprintf("UMI.%d_1", i))
	}

	type result struct {
		gid, name string
		present   map[string]bool
		err       error
	}
	results := make([]result, len(gids))
	var wg sync.WaitGroup
	sem := make(chan struct{}, c.Concurrency)
	for i, gid := range gids {
		wg.Add(1)
		go func(i int, gid string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			r := result{gid: gid, present: map[string]bool{}}
			r.name, r.err = c.RegionName(ctx, gbif.EBirdDatasetKey, gid)
			if r.err != nil {
				results[i] = r
				return
			}
			for _, key := range []string{gbif.EBirdDatasetKey, iNatDatasetKey} {
				recs, err := c.NamesIn(ctx, key, gid)
				if err != nil {
					r.err = err
					break
				}
				var total int64
				for _, rec := range recs {
					total += rec.Records
				}
				if total == 0 {
					continue
				}
				for _, rec := range recs {
					if float64(rec.Records)/float64(total)*1e6 >= minPPM {
						r.present[rec.Name] = true
					}
				}
			}
			results[i] = r
			log.Printf("  %-28s %5d names", r.name, len(r.present))
		}(i, gid)
	}
	wg.Wait()

	// regionsOf maps a species' lowercased scientific name to the regions it
	// was reported in.
	regionsOf := map[string][]string{}
	names := map[string]string{}
	for _, r := range results {
		if r.err != nil {
			log.Fatalf("birdview-regions: %s: %v", r.gid, r.err)
		}
		if r.name == "" {
			log.Printf("  WARNING: %s named nothing and is skipped", r.gid)
			continue
		}
		names[r.gid] = r.name
		for _, sp := range ds {
			if r.present[strings.ToLower(sp)] {
				regionsOf[sp] = append(regionsOf[sp], r.gid)
			}
		}
	}

	placed := 0
	for _, v := range regionsOf {
		if len(v) > 0 {
			placed++
		}
	}
	log.Printf("birdview-regions: %d of %d species placed in at least one region", placed, len(ds))

	if err := write(*outPath, ds, regionsOf, names); err != nil {
		log.Fatalf("birdview-regions: %v", err)
	}
	log.Printf("birdview-regions: wrote %s", *outPath)
}

func loadSpecies(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("reading the dataset: %w (run birdview-ingest first)", err)
	}
	defer f.Close()
	var ds struct {
		Species []struct {
			ScientificName string `json:"scientificName"`
		} `json:"species"`
	}
	if err := json.NewDecoder(f).Decode(&ds); err != nil {
		return nil, err
	}
	out := make([]string, 0, len(ds.Species))
	for _, s := range ds.Species {
		out = append(out, s.ScientificName)
	}
	sort.Strings(out)
	return out, nil
}

func write(path string, species []string, regionsOf map[string][]string, names map[string]string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	var b strings.Builder
	fmt.Fprintf(&b, "# BirdView per-region presence. Generated by cmd/birdview-regions.\n")
	fmt.Fprintf(&b, "# A species is listed in a region when it reaches %g record per million\n", minPPM)
	fmt.Fprintf(&b, "# of that region's bird records, in eBird or iNaturalist (sources.md, `gbif`).\n")
	fmt.Fprintf(&b, "#\n# region\tGADM id\tname\n")
	gids := make([]string, 0, len(names))
	for g := range names {
		gids = append(gids, g)
	}
	sort.Strings(gids)
	for _, g := range gids {
		fmt.Fprintf(&b, "#region\t%s\t%s\n", g, names[g])
	}
	fmt.Fprintf(&b, "#\n# scientific name\tregion ids, comma separated\n")
	for _, sp := range species {
		rs := regionsOf[sp]
		sort.Strings(rs)
		fmt.Fprintf(&b, "%s\t%s\n", sp, strings.Join(rs, ","))
	}
	return os.WriteFile(path, []byte(b.String()), 0o644)
}
