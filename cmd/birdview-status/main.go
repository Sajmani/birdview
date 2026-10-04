// Command birdview-status derives BirdView's own occurrence status for every
// species the ABA Area reports, and writes the committed status table.
//
// It replaces the ABA rarity codes, which supplied scope until 2026-10-04 and
// could not be redistributed (P-001, CR-028). What it measures is not how
// MANY times a bird was seen but whether the sightings RECUR: a record count
// tracks observer effort, so it under-reports pelagic species and island
// endemics and over-reports vagrants, and those biases pull in opposite
// directions. Recurrence is insensitive to both.
//
// Two passes, which is what keeps the run small:
//
//  1. One verbatimScientificName facet per region — about 70 queries — giving
//     the record count for every (species, region) pair at once.
//  2. An eventDate facet for each species in its top regions by record count.
//     If a species does not recur where it is seen most, it does not recur
//     anywhere, so there is no reason to ask about the other sixty.
//
// Both eBird and iNaturalist are queried, because eBird withholds
// conservation-sensitive species outright and iNaturalist is the only open
// source that has them (sources.md, `gbif`).
//
// Nothing queries the network at build or test time; the output is committed
// (T-005's sibling rule for generated tables, as for data/gbif-taxa.tsv).
//
// Verifies: P-001, P-070, T-006.
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
	"birdview/internal/status"
)

// iNatDatasetKey is the iNaturalist research-grade observations dataset.
const iNatDatasetKey = "50c9509d-22c7-4a22-a47d-8c48425ef4a7"

// topRegions is how many of a species' regions get the date query, ordered by
// record count.
//
// Three, not one: the region where a species has the most records is usually
// where it is most established, but not always — a bird can be commonest in
// the state with the most birders rather than the state it breeds in. Three
// covers that without multiplying the run by sixty.
const topRegions = 3

// minRecords is the floor below which a date query is not worth making. A
// species with fewer than this in a region cannot reach MinEpisodes however
// its dates fall, so the query would only confirm a vagrant.
const minRecords = 3

type pair struct {
	name    string // verbatim scientific name, lowercased
	gid     string
	records int64
}

func main() {
	var (
		outPath  = flag.String("out", "data/status.tsv", "output table (committed)")
		dump     = flag.String("dump", "", "also write the raw per-region shapes here, for calibration")
		limit    = flag.Int("limit", 0, "stop after this many species (0 = all), for a smoke run")
		cache    = flag.String("cache", "/tmp/birdview-status-pass1.json", "pass 1 cache; reused if present")
		workers  = flag.Int("workers", 2, "concurrent requests")
		progress = flag.String("progress", "/tmp/birdview-status-pass2.jsonl",
			"pass 2 progress log; completed species are skipped on a rerun")
	)
	flag.Parse()

	ctx := context.Background()
	c := gbif.NewClient()
	// Gentler than the package defaults. These are whole-region facets over a
	// billion records, not taxon lookups, and GBIF sheds load on them: at the
	// default 100ms across three workers the first run took a 429 that
	// outlasted all four retries. The courtesy is ours to set, not a limit
	// GBIF publishes.
	c.Interval = 250 * time.Millisecond
	c.Backoff = 5 * time.Second
	gids := region.LandIDs()

	// Pass 1: record counts for every (species, region) pair, both datasets.
	// Cached to disk because it is the slow half, and a throttle in the
	// fortieth region should not discard the first thirty-nine.
	counts := map[string]map[string]int64{} // name -> gid -> records
	if b, err := os.ReadFile(*cache); err == nil {
		if err := json.Unmarshal(b, &counts); err != nil {
			log.Fatalf("birdview-status: cache %s: %v", *cache, err)
		}
		log.Printf("birdview-status: pass 1 read from %s, %d names", *cache, len(counts))
	}
	var mu sync.Mutex
	var wg sync.WaitGroup
	sem := make(chan struct{}, *workers)
	if len(counts) == 0 {
		log.Printf("birdview-status: pass 1, %d regions x 2 datasets", len(gids))
		var failed []string
		for _, gid := range gids {
			for _, ds := range []string{gbif.EBirdDatasetKey, iNatDatasetKey} {
				wg.Add(1)
				go func(gid, ds string) {
					defer wg.Done()
					sem <- struct{}{}
					defer func() { <-sem }()
					recs, err := c.NamesIn(ctx, ds, gid)
					mu.Lock()
					defer mu.Unlock()
					if err != nil {
						// Collected rather than fatal: dying here throws away
						// every region that did answer.
						failed = append(failed, gid+": "+err.Error())
						return
					}
					for _, r := range recs {
						if counts[r.Name] == nil {
							counts[r.Name] = map[string]int64{}
						}
						counts[r.Name][gid] += r.Records
					}
				}(gid, ds)
			}
		}
		wg.Wait()
		if len(failed) > 0 {
			// A missing region silently shrinks every species' evidence, so a
			// partial pass is refused rather than written out looking whole.
			log.Fatalf("birdview-status: %d region queries failed, refusing to write "+
				"a partial table:\n  %s", len(failed), strings.Join(failed, "\n  "))
		}
		b, err := json.Marshal(counts)
		if err != nil {
			log.Fatalf("birdview-status: %v", err)
		}
		if err := os.WriteFile(*cache, b, 0o644); err != nil {
			log.Fatalf("birdview-status: %v", err)
		}
		log.Printf("birdview-status: pass 1 done, %d names, cached to %s", len(counts), *cache)
	}

	counts = toBinomials(counts)
	names := make([]string, 0, len(counts))
	for n := range counts {
		names = append(names, n)
	}
	sort.Strings(names)
	if *limit > 0 && *limit < len(names) {
		names = names[:*limit]
	}

	// Pass 2: dates for each species in its top regions.
	type result struct {
		name   string
		shapes map[string]status.Shape
		best   status.Status
	}
	results := make([]result, len(names))
	var failed2 []string
	nDone := 0

	// Pass 2 is resumable, for the same reason pass 1 is cached and then
	// some: an overnight run hung for nine hours in a read on a socket the
	// laptop's sleep had killed, and everything it had already measured was
	// in memory and lost. Each species is appended as it completes, so a
	// rerun pays only for what is missing.
	done := map[string]map[string]status.Shape{}
	if f, err := os.Open(*progress); err == nil {
		dec := json.NewDecoder(f)
		for {
			var rec struct {
				Name   string                  `json:"n"`
				Shapes map[string]status.Shape `json:"s"`
			}
			if err := dec.Decode(&rec); err != nil {
				break
			}
			done[rec.Name] = rec.Shapes
		}
		f.Close()
		log.Printf("birdview-status: resuming, %d species already measured", len(done))
	}
	pf, err := os.OpenFile(*progress, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		log.Fatalf("birdview-status: %v", err)
	}
	defer pf.Close()
	penc := json.NewEncoder(pf)

	log.Printf("birdview-status: pass 2, %d species x up to %d regions", len(names), topRegions)
	for i, n := range names {
		if sh, ok := done[n]; ok {
			var sts []status.Status
			for _, s := range sh {
				sts = append(sts, status.Classify(s))
			}
			results[i] = result{n, sh, status.Best(sts)}
			continue
		}
		wg.Add(1)
		go func(i int, n string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			var ps []pair
			for gid, recs := range counts[n] {
				ps = append(ps, pair{n, gid, recs})
			}
			sort.Slice(ps, func(a, b int) bool {
				if ps[a].records != ps[b].records {
					return ps[a].records > ps[b].records
				}
				return ps[a].gid < ps[b].gid // stable, so a rerun matches
			})
			if len(ps) > topRegions {
				ps = ps[:topRegions]
			}
			shapes := map[string]status.Shape{}
			var sts []status.Status
			for _, p := range ps {
				if p.records < minRecords {
					continue
				}
				sh, err := shapeOf(ctx, c, p)
				if err != nil {
					// Collected, not fatal. The first run died in a goroutine
					// on one unqueryable name and discarded ninety seconds of
					// completed work; a failure here is refused at the end
					// instead, where the whole list can be shown at once.
					mu.Lock()
					failed2 = append(failed2, err.Error())
					mu.Unlock()
					continue
				}
				shapes[p.gid] = sh
				sts = append(sts, status.Classify(sh))
			}
			results[i] = result{n, shapes, status.Best(sts)}
			// Recorded under the same lock as failed2, which also serializes
			// the encoder: two goroutines writing a JSON line at once would
			// interleave and make the file unreadable on resume.
			mu.Lock()
			if err := penc.Encode(struct {
				Name   string                  `json:"n"`
				Shapes map[string]status.Shape `json:"s"`
			}{n, shapes}); err != nil {
				log.Printf("birdview-status: progress write: %v", err)
			}
			nDone++
			if nDone%100 == 0 {
				log.Printf("birdview-status:   %d/%d measured", nDone, len(names)-len(done))
			}
			mu.Unlock()
		}(i, n)
	}
	wg.Wait()
	if len(failed2) > 0 {
		// Same reason as pass 1: a species whose dates never arrived would be
		// filed as a vagrant on no evidence, which is indistinguishable in the
		// output from a measured one.
		log.Fatalf("birdview-status: %d date queries failed, refusing to write "+
			"a table that would call them vagrants:\n  %s",
			len(failed2), strings.Join(failed2, "\n  "))
	}

	if err := os.MkdirAll(filepath.Dir(*outPath), 0o755); err != nil {
		log.Fatalf("birdview-status: %v", err)
	}
	f, err := os.Create(*outPath)
	if err != nil {
		log.Fatalf("birdview-status: %v", err)
	}
	defer f.Close()
	fmt.Fprintf(f, "# BirdView's own occurrence status, derived from GBIF (P-001, CR-028).\n")
	fmt.Fprintf(f, "# Generated by cmd/birdview-status. Not a copy of anything published.\n")
	fmt.Fprintf(f, "# verbatim name<TAB>status<TAB>best region<TAB>episodes<TAB>years<TAB>records\n#\n")
	fmt.Fprintf(f, "# Status is recurrence, not abundance: a species is established where its\n")
	fmt.Fprintf(f, "# records fall in repeated episodes across years, however few they are.\n")
	fmt.Fprintf(f, "# Sources: EOD - eBird Observation Dataset, doi:10.15468/aomfnb, and the\n")
	fmt.Fprintf(f, "# iNaturalist research-grade dataset. Both CC BY 4.0.\n")

	tally := map[status.Status]int{}
	for _, r := range results {
		if r.name == "" {
			continue
		}
		tally[r.best]++
		gid, sh := bestRegion(r.shapes)
		fmt.Fprintf(f, "%s\t%s\t%s\t%d\t%d\t%d\n",
			r.name, r.best, gid, sh.Episodes, sh.Years, sh.Records)
	}
	log.Printf("birdview-status: wrote %s", *outPath)
	for _, s := range []status.Status{status.Established, status.Rare, status.Vagrant, status.Absent} {
		log.Printf("birdview-status:   %-12s %d", s, tally[s])
	}

	if *dump != "" {
		d, err := os.Create(*dump)
		if err != nil {
			log.Fatalf("birdview-status: %v", err)
		}
		defer d.Close()
		fmt.Fprintf(d, "name\tgid\trecords\tdates\tepisodes\tyears\tspan\ttruncated\tstatus\n")
		for _, r := range results {
			for gid, sh := range r.shapes {
				fmt.Fprintf(d, "%s\t%s\t%d\t%d\t%d\t%d\t%d\t%v\t%s\n",
					r.name, gid, sh.Records, sh.Dates, sh.Episodes, sh.Years,
					sh.SpanDays, sh.Truncated, status.Classify(sh))
			}
		}
		log.Printf("birdview-status: wrote %s", *dump)
	}
}

// toBinomials reduces the verbatim facet to species.
//
// iNaturalist publishes each identification at the rank the observer reached,
// so the facet mixes binomials with subspecies trinomials, domestic forms,
// hybrids and bare genera. eBird does not — every one of its 1,382 ABA-area
// names is already a binomial — so this only ever touches the iNaturalist
// side.
//
// Hybrids are dropped because they are not one species, and a bare genus
// because it was never identified to one. Anything longer is truncated to its
// first two tokens, which folds `setophaga coronata auduboni` into its
// species and `columba livia domestica` into Rock Pigeon.
//
// Measured 2026-10-04: 97.1% of iNaturalist's ABA-area records are already
// bare binomials, 2.6% sit under trinomials, 0.2% are hybrids and 0.1% are
// genus-only. Only 7 species have records exclusively under a trinomial and
// the largest has 6, so nothing here can change a classification. The dates
// query below asks for the binomial alone and therefore misses that 2.6%;
// that is accepted, because those records can only ever make an established
// bird look more established.
func toBinomials(in map[string]map[string]int64) map[string]map[string]int64 {
	out := map[string]map[string]int64{}
	for name, regions := range in {
		if strings.ContainsRune(name, '×') {
			continue
		}
		f := strings.Fields(name)
		if len(f) < 2 {
			continue
		}
		k := f[0] + " " + f[1]
		if out[k] == nil {
			out[k] = map[string]int64{}
		}
		for gid, n := range regions {
			out[k][gid] += n
		}
	}
	return out
}

// shapeOf asks both datasets for the dates behind one (species, region) pair
// and measures the union. The union is right because the two datasets record
// different observers, not the same ones twice: a day either had a sighting
// or it did not.
func shapeOf(ctx context.Context, c *gbif.Client, p pair) (status.Shape, error) {
	all := map[string]int64{}
	var total int64
	var trunc bool
	for _, ds := range []string{gbif.EBirdDatasetKey, iNatDatasetKey} {
		dates, n, t, err := c.DatesIn(ctx, ds, p.gid, p.name)
		if err != nil {
			return status.Shape{}, fmt.Errorf("dates for %q in %s: %w", p.name, p.gid, err)
		}
		for d, k := range dates {
			all[d] += k
		}
		total += n
		trunc = trunc || t
	}
	var ts []time.Time
	for d := range all {
		t, err := time.Parse("2006-01-02", d)
		if err != nil {
			continue // a range rather than a day; DatesIn filters these already
		}
		ts = append(ts, t)
	}
	return status.Measure(total, ts, trunc), nil
}

// bestRegion names the region behind a species' status, so the table carries
// its own evidence and a verdict can be argued with.
func bestRegion(shapes map[string]status.Shape) (string, status.Shape) {
	var bg string
	var bs status.Shape
	rank := map[status.Status]int{status.Absent: 0, status.Vagrant: 1, status.Rare: 2, status.Established: 3}
	for gid, sh := range shapes {
		if bg == "" || rank[status.Classify(sh)] > rank[status.Classify(bs)] ||
			(rank[status.Classify(sh)] == rank[status.Classify(bs)] && sh.Episodes > bs.Episodes) {
			bg, bs = gid, sh
		}
	}
	if bg == "" {
		return "-", status.Shape{}
	}
	return bg, bs
}
