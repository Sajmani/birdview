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

// Command birdview-status derives BirdView's own occurrence status for every
// species its regions report, and writes the committed status table.
//
// It replaces the ABA rarity codes, which supplied scope until 2026-10-04 and
// could not be redistributed (product.scopeIsRecurringAvifauna, decisions.scopeFromRecurrence). What it measures is not how
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
// (tech.datasetNeverCommitted's sibling rule for generated tables, as for data/gbif-taxa.tsv).
//
// Verifies: product.scopeIsRecurringAvifauna, product.statusIsPublished, tech.identifiedClient.
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
	"strconv"
	"strings"
	"sync"
	"time"

	"birdview/internal/gbif"
	"birdview/internal/region"
	"birdview/internal/status"
	"birdview/internal/taxonomy"
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
		progress = flag.String("progress", "/tmp/birdview-status-pass2-bysource.jsonl",
			"pass 2 progress log; completed species are skipped on a rerun")
	)
	flag.Parse()

	ctx := context.Background()
	c := gbif.NewClient()
	// Gentler than the package defaults, and a starting point rather than a
	// setting: these are whole-region facets over a billion records, not
	// taxon lookups, and GBIF sheds load on them. 250ms throttled hard enough
	// to spend a run in backoff; 500ms measured clean over twenty consecutive
	// requests. The client widens this itself when GBIF says to (slowDown).
	c.Interval = 500 * time.Millisecond
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
					// Month 0: this pass classifies a species over the whole
					// year, and decisions.monthPresence's monthly rule is applied by the
					// server over the region table, not here.
					recs, err := c.NamesIn(ctx, ds, gid, 0)
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

	counts = prepare(counts)
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
	// The cache holds RAW DATES, not computed shapes, so that changing how a
	// shape is measured or classified costs nothing. It was shapes, and the
	// first change to the model -- adding LastSeen, once extinct birds turned
	// up classified as established -- would have meant fetching all 1,429
	// species a third time. Fetching is expensive and cacheable; measuring is
	// cheap and still being argued about. They belong apart.
	done := map[string]map[string]map[string]int64{}
	trunc := map[string]map[string]bool{}
	if f, err := os.Open(*progress); err == nil {
		dec := json.NewDecoder(f)
		for {
			var rec cacheRec
			if err := dec.Decode(&rec); err != nil {
				break
			}
			done[rec.Name] = countedAll(rec.Name, rec.BySource)
			trunc[rec.Name] = rec.Trunc
		}
		f.Close()
		log.Printf("birdview-status: %d species already fetched", len(done))
	}
	pf, err := os.OpenFile(*progress, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		log.Fatalf("birdview-status: %v", err)
	}
	defer pf.Close()
	penc := json.NewEncoder(pf)

	log.Printf("birdview-status: pass 2, %d species x up to %d regions", len(names), topRegions)
	for i, n := range names {
		if _, ok := done[n]; ok {
			continue // already fetched; measured below, once the cutoff is known
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
			dates := map[string]map[string]int64{}
			bySource := map[string]map[string]map[string]int64{}
			tr := map[string]bool{}
			partial := false
			for _, p := range ps {
				if p.records < minRecords {
					continue
				}
				d, t, err := datesOf(ctx, c, p)
				if err != nil {
					// Collected, not fatal. The first run died in a goroutine
					// on one unqueryable name and discarded ninety seconds of
					// completed work; a failure here is refused at the end
					// instead, where the whole list can be shown at once.
					partial = true
					mu.Lock()
					failed2 = append(failed2, err.Error())
					mu.Unlock()
					continue
				}
				bySource[p.gid] = d
				dates[p.gid] = countedDates(n, p.gid, d)
				tr[p.gid] = t
			}

			// A species missing one of its regions is NOT recorded as done.
			//
			// It was, and that quietly defeated the refusal above: the run
			// stopped without writing a table, but the three species whose
			// Florida or Alberta query had failed were already in the
			// progress file with the other two regions, so the resume took
			// them as measured and wrote the table anyway. Refusing at the
			// end and checkpointing partial work are only consistent if the
			// checkpoint holds complete species.
			if partial {
				return
			}
			// Recorded under the same lock as failed2, which also serializes
			// the encoder: two goroutines writing a JSON line at once would
			// interleave and make the file unreadable on resume.
			mu.Lock()
			done[n], trunc[n] = dates, tr
			if err := penc.Encode(cacheRec{n, bySource, tr}); err != nil {
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

	// Measure, now that every species is fetched. The recency cutoff is
	// anchored to the LATEST DATE IN THE CORPUS, not to time.Now(): the data
	// has its own end -- eBird's export runs to the end of 2024 -- and
	// anchoring to the clock would make the same inputs classify differently
	// next year, with no new evidence behind the change.
	var latest time.Time
	for n := range done {
		for gid, dates := range done[n] {
			sh := measure(dates, trunc[n][gid])
			if sh.LastSeen.After(latest) {
				latest = sh.LastSeen
			}
		}
	}
	since := latest.AddDate(-status.RecentYears, 0, 0)
	log.Printf("birdview-status: corpus ends %s; counting records since %s",
		latest.Format("2006-01-02"), since.Format("2006-01-02"))

	for i, n := range names {
		shapes := map[string]status.Shape{}
		var sts []status.Status
		for gid, dates := range done[n] {
			sh := measure(dates, trunc[n][gid])
			shapes[gid] = sh
			sts = append(sts, status.Classify(sh, since))
		}
		results[i] = result{n, shapes, status.Best(sts)}
	}

	if err := os.MkdirAll(filepath.Dir(*outPath), 0o755); err != nil {
		log.Fatalf("birdview-status: %v", err)
	}
	f, err := os.Create(*outPath)
	if err != nil {
		log.Fatalf("birdview-status: %v", err)
	}
	defer f.Close()
	fmt.Fprintf(f, "# BirdView's own occurrence status, derived from GBIF (product.scopeIsRecurringAvifauna, decisions.scopeFromRecurrence).\n")
	fmt.Fprintf(f, "# Generated by cmd/birdview-status. Not a copy of anything published.\n")
	fmt.Fprintf(f, "# name<TAB>status<TAB>region<TAB>episodes<TAB>years<TAB>records"+
		"<TAB>commonest region<TAB>its records there<TAB>all bird records there"+
		"<TAB>last seen\n#\n")
	fmt.Fprintf(f, "# The last three columns are where the species makes up the biggest SHARE of\n")
	fmt.Fprintf(f, "# what is recorded, which is not where it has the most records: California has\n")
	fmt.Fprintf(f, "# 35x the Sharp-shinned Hawk records of Yukon and 118x the birdwatching, so the\n")
	fmt.Fprintf(f, "# hawk is 1 record in 250 there against 1 in 835 in California. Dividing by the\n")
	fmt.Fprintf(f, "# region's own total cancels out how hard anyone was looking. Both numbers are\n")
	fmt.Fprintf(f, "# written rather than the ratio, so a share resting on nine records says so.\n")
	fmt.Fprintf(f, "# Status is recurrence, not abundance: a species is established where its\n")
	fmt.Fprintf(f, "# records fall in repeated episodes across years, however few they are.\n")
	fmt.Fprintf(f, "# Sources: EOD - eBird Observation Dataset, doi:10.15468/aomfnb, and the\n")
	fmt.Fprintf(f, "# iNaturalist research-grade dataset. Both CC BY 4.0.\n#\n")
	fmt.Fprintf(f, "# Corpus ends %s. A species last recorded before %s is historical,\n",
		latest.Format("2006-01-02"), since.Format("2006-01-02"))
	fmt.Fprintf(f, "# not present: GBIF carries museum specimens under their collection dates,\n")
	fmt.Fprintf(f, "# and Passenger Pigeon recurs across a century of them.\n")
	fmt.Fprintf(f, "# `last seen` is the most recent date in the region column, or - where no\n")
	fmt.Fprintf(f, "# date was measured. It is what makes a historical verdict checkable rather\n")
	fmt.Fprintf(f, "# than asserted, so a page can say WHEN (product.undrawnSpeciesAnswer).\n")

	// Where each species makes up the biggest share of what that region
	// records. Computed from the pass-1 counts, so it costs no requests.
	regionTotal := map[string]int64{}
	for _, regions := range counts {
		for gid, n := range regions {
			regionTotal[gid] += n
		}
	}
	type peak struct {
		gid     string
		records int64
		total   int64
	}
	peakOf := map[string]peak{}
	for name, regions := range counts {
		var best peak
		for gid, n := range regions {
			t := regionTotal[gid]
			if t == 0 {
				continue
			}
			if best.total == 0 || float64(n)/float64(t) > float64(best.records)/float64(best.total) {
				best = peak{gid, n, t}
			}
		}
		if best.total > 0 {
			peakOf[name] = best
		}
	}

	tally := map[status.Status]int{}
	for _, r := range results {
		if r.name == "" {
			continue
		}
		tally[r.best]++
		gid, sh := bestRegion(r.shapes, since)
		pk := peakOf[r.name]
		if pk.gid == "" {
			pk.gid = "-"
		}
		seen := "-"
		if !sh.LastSeen.IsZero() {
			seen = sh.LastSeen.Format(status.DateLayout)
		}
		fmt.Fprintf(f, "%s\t%s\t%s\t%d\t%d\t%d\t%s\t%d\t%d\t%s\n",
			r.name, r.best, gid, sh.Episodes, sh.Years, sh.Records,
			pk.gid, pk.records, pk.total, seen)
	}
	log.Printf("birdview-status: wrote %s", *outPath)
	for _, s := range []status.Status{status.Established, status.Rare, status.Vagrant,
		status.Historical, status.Absent} {
		log.Printf("birdview-status:   %-12s %d", s, tally[s])
	}

	if *dump != "" {
		d, err := os.Create(*dump)
		if err != nil {
			log.Fatalf("birdview-status: %v", err)
		}
		defer d.Close()
		fmt.Fprintf(d, "name\tgid\trecords\tdates\tepisodes\tyears\tspan\tlastseen\ttruncated\tstatus\n")
		for _, r := range results {
			for gid, sh := range r.shapes {
				fmt.Fprintf(d, "%s\t%s\t%d\t%d\t%d\t%d\t%d\t%s\t%v\t%s\n",
					r.name, gid, sh.Records, sh.Dates, sh.Episodes, sh.Years,
					sh.SpanDays, sh.LastSeen.Format("2006-01-02"), sh.Truncated,
					status.Classify(sh, since))
			}
		}
		log.Printf("birdview-status: wrote %s", *dump)
	}
}

// prepare turns the raw verbatim facet into the species the run classifies.
//
// One function rather than two calls at the point of use, so a test can
// exercise what main actually runs. The first version of the lump test called
// mergeLumps directly and passed happily with the call removed from main --
// it verified the function and not the pipeline, which is the shape of a
// check that tests nothing.
func prepare(in map[string]map[string]int64) map[string]map[string]int64 {
	return mergeCounted(toBinomials(in))
}

// mergeCounted folds a name the occurrence data publishes into the AviList
// species whose birds those records are (taxonomy.CountedAs).
//
// Without it the run admits both sides of a boundary disagreement as
// separate species. eBird holds 556,569 records under Numenius phaeopus and
// none under hudsonicus; iNaturalist holds 17,923 under hudsonicus and nine
// under phaeopus. Merged, that is one Whimbrel. Left apart it is two -- and
// because the published name is also hudsonicus's AVONET donor, the two
// would draw identically.
func mergeCounted(in map[string]map[string]int64) map[string]map[string]int64 {
	for _, c := range taxonomy.CountedAs {
		lump, into := c.Published, c.Species
		l, i := strings.ToLower(lump), strings.ToLower(into)
		regions, ok := in[l]
		if !ok {
			continue
		}
		if in[i] == nil {
			in[i] = map[string]int64{}
		}
		for gid, n := range regions {
			in[i][gid] += n
		}
		delete(in, l)
	}
	return in
}

// toBinomials reduces the verbatim facet to species, by taxonomy.Binomial —
// the same rule birdview-regions now applies through taxonomy.Occurrence.
//
// The rule lived here and only here until 2026-10-06, which is why the
// region pass never applied it: the last of the three normalization
// asymmetries decisions.regionJoinNormalizes found. What stays in this function is the part that is
// this command's own. A name identifying no species is DROPPED here, because
// this pass classifies species and a hybrid has no status to carry;
// birdview-regions keeps it instead, because a hybrid is still a bird
// somebody saw and still occupies a region's denominator.
//
// Only 7 species have records exclusively under a trinomial and the largest
// has 6, so folding them cannot change a classification. The dates query
// below asks for the binomial alone and therefore misses the 2.6% of
// iNaturalist records filed under trinomials; that is accepted, because
// those records can only ever make an established bird look more
// established.
func toBinomials(in map[string]map[string]int64) map[string]map[string]int64 {
	out := map[string]map[string]int64{}
	for name, regions := range in {
		k, ok := taxonomy.Binomial(name)
		if !ok {
			continue
		}
		if out[k] == nil {
			out[k] = map[string]int64{}
		}
		for gid, n := range regions {
			out[k][gid] += n
		}
	}
	return out
}

// cacheRec is one species' raw fetch: the dates each region reported, and
// whether that region's facet was truncated.
type cacheRec struct {
	Name string `json:"n"`
	// BySource maps a region to a GBIF dataset key to the dates that source
	// reported. Kept per source, not unioned, so the counting rule of
	// decisions.countEbirdFirst -- eBird's dates, and iNaturalist's only where eBird is
	// withholding the species -- can be applied, and changed, without
	// refetching.
	BySource map[string]map[string]map[string]int64 `json:"s"`
	Trunc    map[string]bool                        `json:"t"`
}

// datesOf asks both datasets for the dates behind one (species, region) pair
// and unions them. The union is right because the two datasets record
// different observers, not the same ones twice: a day either had a sighting
// or it did not.
func datesOf(ctx context.Context, c *gbif.Client, p pair) (map[string]map[string]int64, bool, error) {
	all := map[string]map[string]int64{}
	var trunc bool
	for _, name := range gbifNames(p.name) {
		for _, ds := range []string{gbif.EBirdDatasetKey, iNatDatasetKey} {
			dates, _, t, err := c.DatesIn(ctx, ds, p.gid, name)
			if err != nil {
				return nil, false, fmt.Errorf("dates for %q in %s: %w", name, p.gid, err)
			}
			if all[ds] == nil {
				all[ds] = map[string]int64{}
			}
			for d, k := range dates {
				all[ds][d] += k
			}
			trunc = trunc || t
		}
	}
	return all, trunc, nil
}

// countedDates applies decisions.countEbirdFirst's rule to one species in one region: every
// date eBird reported, plus iNaturalist's dates only in months eBird is
// confirmed to withhold the species there (decisions.ebirdSilenceNotAbsence).
//
// Scope is recurrence, so this decides which species BirdView draws at all.
// Before it, the two sources' dates were unioned, and a species iNaturalist
// alone photographed in repeated years -- feral peafowl, escaped parakeets
// -- recurred on that evidence. eBird's reviewers vet every record; the
// union let an unvetted second source decide scope on its own.
//
// name is the occurrence name, lowercased.
func countedDates(name, gid string, by map[string]map[string]int64) map[string]int64 {
	out := map[string]int64{}
	for d, k := range by[gbif.EBirdDatasetKey] {
		out[d] += k
	}
	ws := taxonomy.Withholdings(aviListSpelling(name), gid)
	if len(ws) == 0 {
		return out
	}
	monthly := func(dates map[string]int64) [12]int64 {
		var m [12]int64
		for d, k := range dates {
			if mo, ok := monthOf(d); ok {
				m[mo] += k
			}
		}
		return m
	}
	withheld := taxonomy.ConfirmedMonths(ws, monthly(by[gbif.EBirdDatasetKey]),
		monthly(by[iNatDatasetKey]))
	for d, k := range by[iNatDatasetKey] {
		if mo, ok := monthOf(d); ok && withheld[mo] {
			out[d] += k
		}
	}
	return out
}

// countedAll applies countedDates to every region of a cached species.
func countedAll(name string, by map[string]map[string]map[string]int64) map[string]map[string]int64 {
	out := make(map[string]map[string]int64, len(by))
	for gid, d := range by {
		out[gid] = countedDates(name, gid, d)
	}
	return out
}

// monthOf reads the 0-based month from an ISO date, "2024-06-03".
func monthOf(date string) (int, bool) {
	if len(date) < 7 {
		return 0, false
	}
	mo, err := strconv.Atoi(date[5:7])
	if err != nil || mo < 1 || mo > 12 {
		return 0, false
	}
	return mo - 1, true
}

// aviListSpelling maps a lowercased occurrence name onto the AviList name
// the Sensitive Species table is keyed on.
func aviListSpelling(name string) string {
	if name == "" {
		return ""
	}
	t := strings.ToUpper(name[:1]) + name[1:]
	if alt, ok := taxonomy.AviListRenames[t]; ok {
		return alt
	}
	return t
}

// gbifNames lists every name the occurrence data files this species under:
// its own, plus any lump the databases still publish it as.
//
// Merging the pass-1 counts is not enough on its own, and the first attempt
// did only that. It changed which regions were queried and left pass 2
// asking GBIF for "numenius hudsonicus" -- a name eBird holds **zero**
// records under -- so the merged species came back with iNaturalist's 169
// and none of eBird's half-million. The records exist only under the lump,
// and can only be fetched under it.
func gbifNames(species string) []string {
	out := []string{species}
	for _, c := range taxonomy.CountedAs {
		if strings.EqualFold(c.Species, species) {
			out = append(out, strings.ToLower(c.Published))
		}
	}
	sort.Strings(out) // stable, so a rerun matches
	return out
}

// measure turns one region's cached dates into a Shape.
func measure(dates map[string]int64, truncated bool) status.Shape {
	var ts []time.Time
	var records int64
	for d, n := range dates {
		t, err := time.Parse("2006-01-02", d)
		if err != nil {
			continue
		}
		ts = append(ts, t)
		records += n
	}
	return status.Measure(records, ts, truncated)
}

// bestRegion names the region behind a species' status, so the table carries
// its own evidence and a verdict can be argued with.
func bestRegion(shapes map[string]status.Shape, since time.Time) (string, status.Shape) {
	// Ordered, not a bare map walk. Ranking on (status, episodes) alone left
	// ties, and a tie was broken by Go's randomized map iteration: ten rows
	// of the committed table named a different region on every run, carrying
	// their episode, year, record and last-seen figures with them. Cooper's
	// Hawk recurs in 39 episodes in both Alaska and Arizona, and the table
	// reported 24 years and 216,391 records one run and 18 and 119,194 the
	// next. A generated table that two runs over identical inputs disagree
	// about is not evidence of anything (tech.datasetNeverCommitted, product.undrawnSpeciesAnswer).
	gids := make([]string, 0, len(shapes))
	for gid := range shapes {
		gids = append(gids, gid)
	}
	sort.Strings(gids)

	// Years before records, because the model is recurrence and not abundance
	// (decisions.scopeFromRecurrence): between two regions a species recurs in equally often, the one
	// it has come back to across more years is the better evidence, whatever
	// the crowds did. Records only separate what years cannot, and the region
	// id settles what records cannot -- arbitrary, but the same arbitrary
	// answer every run, which is the whole point.
	better := func(a, b status.Shape) bool {
		ra, rb := status.Rank(status.Classify(a, since)), status.Rank(status.Classify(b, since))
		switch {
		case ra != rb:
			return ra > rb
		case a.Episodes != b.Episodes:
			return a.Episodes > b.Episodes
		case a.Years != b.Years:
			return a.Years > b.Years
		default:
			return a.Records > b.Records
		}
	}
	var bg string
	var bs status.Shape
	for _, gid := range gids {
		if sh := shapes[gid]; bg == "" || better(sh, bs) {
			bg, bs = gid, sh
		}
	}
	if bg == "" {
		return "-", status.Shape{}
	}
	return bg, bs
}
