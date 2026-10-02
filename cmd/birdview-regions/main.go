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
	"birdview/internal/region"
)

// iNatDatasetKey is the iNaturalist research-grade observations dataset.
const iNatDatasetKey = "50c9509d-22c7-4a22-a47d-8c48425ef4a7"

// usStates and caProvinces are GADM level-1 counts. GADM numbers them from 1
// in alphabetical order, so the ids are generated rather than listed; each
// one's real name is read back from the data and written into the table, so a
// wrong id shows up as a wrong name rather than as silence.
// minPelagicShare is how much of a species' records must fall outside EVERY
// land region before it is treated as a bird of the open sea.
//
// The measure separates cleanly, which is why a single threshold works:
// Bermuda Petrel 100%, Craveri's Murrelet 98.9%, Sooty Shearwater 79.1%,
// Tufted Puffin 53.9% -- against American Robin 0.9% and Herring Gull 1.1%.
// There is nothing between 54% and 1%, so the exact cut is not load-bearing.
const minPelagicShare = 0.5

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
	countries := []string{"US", "CA"}

	gids := region.LandIDs()

	type result struct {
		gid, name string
		present   map[string]bool
		counts    map[string]int64
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
			r := result{gid: gid, present: map[string]bool{}, counts: map[string]int64{}}
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
					r.counts[rec.Name] += rec.Records
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

	// How much of each species sits outside every land region. A land bird is
	// near zero; a petrel is near one. The denominator is the species' whole
	// ABA-area count, which the land facets cannot supply.
	landTotal := map[string]int64{}
	for _, r := range results {
		for _, sp := range ds {
			landTotal[sp] += r.counts[strings.ToLower(sp)]
		}
	}

	log.Printf("birdview-regions: measuring how much of each species is at sea")
	var pelagic []string
	var pmu sync.Mutex
	var pwg sync.WaitGroup
	psem := make(chan struct{}, c.Concurrency)
	for _, sp := range ds {
		pwg.Add(1)
		go func(sp string) {
			defer pwg.Done()
			psem <- struct{}{}
			defer func() { <-psem }()
			total, err := c.CountFor(ctx, gbif.EBirdDatasetKey, strings.ToLower(sp), countries)
			if err != nil || total == 0 {
				return
			}
			if 1-float64(landTotal[sp])/float64(total) >= minPelagicShare {
				pmu.Lock()
				pelagic = append(pelagic, sp)
				pmu.Unlock()
			}
		}(sp)
	}
	pwg.Wait()
	sort.Strings(pelagic)
	log.Printf("birdview-regions: %d species are mostly at sea", len(pelagic))

	// Which ocean. Tested in the order internal/region declares, Arctic first,
	// because it is a latitude band that overlaps both oceans in longitude.
	for _, sp := range pelagic {
		for _, id := range region.MarineIDs() {
			minLat, maxLat, minLon, maxLon, _ := region.MarineBox(id)
			n, err := c.CountInBox(ctx, gbif.EBirdDatasetKey, strings.ToLower(sp),
				countries, minLat, maxLat, minLon, maxLon)
			if err != nil {
				log.Fatalf("birdview-regions: %s in %s: %v", sp, id, err)
			}
			if n > 0 {
				regionsOf[sp] = append(regionsOf[sp], id)
			}
		}
	}
	for _, id := range region.MarineIDs() {
		r, _ := region.Get(id)
		names[id] = r.Name
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
