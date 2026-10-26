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

// Command birdview-regions builds the per-region presence table.
//
// For each US state and Canadian province it asks GBIF which bird names were
// reported there and how often, then writes the regions each of BirdView's
// species occurs in. The ingest reads the result; nothing queries the network
// at build or test time (tech.identifiedClient, and the no-network-in-tests binding).
//
// Two sources, because one is not enough:
//
//   - eBird, for depth: 1.24 billion records across the US and Canada.
//   - iNaturalist, because eBird's export WITHHOLDS conservation-sensitive
//     species entirely. Spotted Owl, Gyrfalcon, Northern Hawk Owl, and the
//     critically endangered Hawaiian endemics have zero eBird records and
//     hundreds in iNaturalist. An eBird-only table omits exactly the birds a
//     reader is most likely to look up. See river/requirements/imports.md, `gbif`.
//
// Names are faceted VERBATIM — as the publisher wrote them — not through
// GBIF's backbone, which drops what it cannot resolve and reported zero
// records for Cooper's Hawk.
//
// Verifies: product.regionAndMonthFilter.
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
	"birdview/internal/taxonomy"
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
// Relative, not absolute, for two reasons measured in the source manifest. eBird's
// volume doubled in five years, so any fixed count admits more species over
// time for no ecological reason. And regions differ hugely in effort —
// California has 114 million records against Hawaii's 2.5 million — so one
// count cannot mean the same thing in both.
//
// 1 ppm gives roughly 600 species for California, 390 for Pennsylvania and
// 267 for Hawaii, which are the right orders of magnitude for a state list.
const minPPM = 1.0

// maxOrphanFraction is how much of the occurrence record may belong to
// names matching no drawn species before the run refuses to write a table.
//
// Measured 2026-10-06: **0.0113%**, under 727 names, down from 0.0431%
// under 1,754 once taxonomy.Occurrence began reducing rank before it
// consults either naming table. Every one of the twenty largest is now a
// hybrid — `larus smithsonianus × hyperboreus`, `baeolophus bicolor ×
// atricristatus` — which is what should be left: a hybrid must never match
// a species, and what remains besides is subspecies of birds BirdView does
// not draw, which have no species here to fold into.
//
// The fold moved **387,515 records into 430 species** and added 114
// species-region presences while removing none, which is the shape the
// change has to have: a trinomial either joins a drawn species or stays
// exactly where it was, so nothing can lose by it. Region totals came back
// byte-identical, confirming this reattributes records rather than
// changing what they are divided by.
//
// The bar is 0.1%, now **8.8x the measured residue**. A bar set at the
// measured value fails the next time iNaturalist gains a subspecies, and a
// check that cries wolf gets deleted. What it exists to catch is the other
// order of magnitude: before decisions.regionJoinNormalizes this was 0.81%, and 9,954,362 of those
// records belonged to ten species BirdView draws.
//
// Deliberately a fraction of records rather than a count of names. A new
// vagrant adds a name and six records; a genus move adds one name and seven
// million.
const maxOrphanFraction = 0.001

func main() {
	var (
		dataPath = flag.String("data", "web/data/species.json", "compiled dataset, for the species to look up")
		outPath  = flag.String("out", "data/region-presence.tsv", "output table (committed)")
		marine   = flag.Bool("marine-only", false,
			"recompute only the sea rows and sea totals in -out, leaving the land half untouched")
		progress = flag.String("progress", "/tmp/birdview-regions-land.jsonl",
			"resumable cache of the raw land facets; delete to refetch")
	)
	flag.Parse()

	ds, err := loadSpecies(*dataPath)
	if err != nil {
		log.Fatalf("birdview-regions: %v", err)
	}
	log.Printf("birdview-regions: %d species to place", len(ds))

	// A sea-only refresh costs about 300 requests against the full run's
	// several thousand, and leaves every land row byte-identical -- so the
	// diff of a change to the marine rule is the marine rule, which is the
	// only way to review one (decisions.honestSeaDenominator).
	if *marine {
		if err := refreshMarine(*outPath, ds); err != nil {
			log.Fatalf("birdview-regions: %v", err)
		}
		return
	}

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
		// raw is this region's facets, from the cache or the wire. Carried
		// per result rather than merged into one shared map: the loop below
		// goes on reading that map while the workers it started write to
		// it, which is a race the detector caught and review did not.
		raw     map[string]map[int]map[string]int64
		present map[string]bool
		counts  map[string]int64
		// monthly is the same count split twelve ways.
		monthly map[string]*[12]int64
		// withheld marks the months eBird is confirmed to withhold each
		// species here (decisions.ebirdSilenceNotAbsence).
		withheld map[string][12]bool
		// total and monthlyTotal are every counted record in the region,
		// the denominator of every figure, taken before shared names are
		// copied so that a shared count is not divided into itself twice.
		total        int64
		monthlyTotal [12]int64
		err          error
	}
	results := make([]result, len(gids))

	// Resume what a previous run already fetched. The land pass is about to
	// go from 140 requests to 1,660 under decisions.monthPresence, one to three hours, and a
	// run that long will meet a dropped socket: GBIF reset two connections
	// during a 40-minute run on 2026-10-05 and timed out two more.
	//
	// The cache holds the RAW FACET, not presence or counts, for the reason
	// birdview-status caches raw dates rather than computed shapes. What is
	// expensive is asking GBIF; deciding what an answer means is free and
	// still being argued about -- decisions.monthPresence is changing that decision right
	// now. Caching the verdict would mean refetching every region the next
	// time the rule moved.
	done := map[string]landRec{}
	if f, err := os.Open(*progress); err == nil {
		dec := json.NewDecoder(f)
		for {
			var rec landRec
			if err := dec.Decode(&rec); err != nil {
				break
			}
			done[rec.GID] = rec
		}
		f.Close()
		if len(done) > 0 {
			log.Printf("birdview-regions: %d regions already fetched (%s)", len(done), *progress)
		}
	}
	pf, err := os.OpenFile(*progress, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		log.Fatalf("birdview-regions: %v", err)
	}
	defer pf.Close()
	penc := json.NewEncoder(pf)
	// One writer. Appending from the workers would interleave two JSON
	// objects into one unparseable line, and the file's whole purpose is to
	// still be readable after a crash.
	var progmu sync.Mutex

	var wg sync.WaitGroup
	sem := make(chan struct{}, c.Concurrency)
	for i, gid := range gids {
		if rec, ok := done[gid]; ok {
			results[i] = result{gid: gid, name: rec.Name, raw: rec.Raw}
			continue
		}
		wg.Add(1)
		go func(i int, gid string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			r := result{gid: gid}
			rec := landRec{GID: gid, Raw: map[string]map[int]map[string]int64{}}
			rec.Name, r.err = c.RegionName(ctx, gbif.EBirdDatasetKey, gid)
			if r.err != nil {
				results[i] = r
				return
			}
			for _, key := range []string{gbif.EBirdDatasetKey, iNatDatasetKey} {
				byMonth := map[int]map[string]int64{}
				// Thirteen: the whole year, and each month. The year is
				// fetched rather than summed so the two can be COMPARED --
				// decisions.monthPresence rests on every record carrying a month, and this
				// is where that claim is checked over the whole dataset
				// instead of the two regions it was spot-checked on.
				for m := 0; m <= 12; m++ {
					recs, err := c.NamesIn(ctx, key, gid, m)
					if err != nil {
						r.err = err
						results[i] = r
						return
					}
					raw := make(map[string]int64, len(recs))
					for _, v := range recs {
						raw[v.Name] = v.Records
					}
					byMonth[m] = raw
				}
				rec.Raw[key] = byMonth
			}
			// Written only when every dataset for this region succeeded.
			// A half-fetched region in the cache would be indistinguishable
			// from a region that really holds nothing for iNaturalist, and
			// the resume would skip it forever.
			progmu.Lock()
			err := penc.Encode(rec)
			progmu.Unlock()
			if err != nil {
				r.err = fmt.Errorf("writing progress: %w", err)
				results[i] = r
				return
			}
			r.name, r.raw = rec.Name, rec.Raw
			results[i] = r
			log.Printf("  %-28s %5d names", rec.Name, len(rec.Raw[gbif.EBirdDatasetKey][0]))
		}(i, gid)
	}
	wg.Wait()

	// Measure, from the cache, whatever its records came from. Separated
	// from the fetch so that a rule change costs a rerun of this loop and
	// not of the network.
	aviName := make(map[string]string, len(ds))
	for _, sp := range ds {
		aviName[strings.ToLower(sp)] = sp
	}
	for i := range results {
		if results[i].err != nil {
			continue
		}
		r := &results[i]
		if r.raw == nil {
			continue
		}
		m, err := measure(r.raw, r.gid, aviName)
		if err != nil {
			r.err = err
			continue
		}
		r.present, r.counts, r.monthly, r.withheld = m.present, m.counts, m.monthly, m.withheld
		r.total, r.monthlyTotal = m.total, m.monthlyTotal
	}
	for i := range results {
		if results[i].err != nil {
			log.Fatalf("birdview-regions: %v", results[i].err)
		}
	}

	// regionsOf maps a species' lowercased scientific name to the regions it
	// was reported in.
	regionsOf := map[string][]region.Reported{}
	// orphan counts records under names that match no drawn species, so the
	// join can say what it discarded instead of discarding it quietly.
	orphan := map[string]int64{}
	drawnName := make(map[string]bool, len(ds))
	for _, sp := range ds {
		drawnName[strings.ToLower(sp)] = true
	}
	names := map[string]string{}
	// Every bird record a region holds, over every species including the
	// ones BirdView does not draw. This is the denominator of the share a
	// page shows (product.shareOfRecords): dividing by all the birds the same observers saw
	// in the same place is what cancels out how hard they were looking.
	totals := map[string]int64{}
	monthlyTotals := map[string][12]int64{}
	for _, r := range results {
		if r.err != nil {
			log.Fatalf("birdview-regions: %s: %v", r.gid, r.err)
		}
		if r.name == "" {
			log.Printf("  WARNING: %s named nothing and is skipped", r.gid)
			continue
		}
		names[r.gid] = r.name
		totals[r.gid] = r.total
		monthlyTotals[r.gid] = r.monthlyTotal
		// Every region with records is written, and the ones under the bar
		// are marked rather than dropped: the bar decides where a species
		// is listed, not what it counts (product.regionAndMonthFilter, decisions.belowBarCounted).
		for _, sp := range ds {
			if lc := strings.ToLower(sp); r.counts[lc] > 0 {
				rep := region.Reported{ID: r.gid, Records: r.counts[lc], Below: !r.present[lc]}
				if m := r.monthly[lc]; m != nil {
					rep.Months = *m
				}
				rep.Withheld = r.withheld[lc]
				regionsOf[sp] = append(regionsOf[sp], rep)
			}
		}
		// Every name that reached the denominator and no species. These are
		// records counted as "some other bird" on every tile, which is
		// correct for a bird BirdView does not draw and a silent loss for
		// one it does (goals.honestData, product.joinsReportDrops).
		for n, c := range r.counts {
			if !drawnName[n] {
				orphan[n] += c
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
	// A sea counts only what is on OPEN WATER inside its box (decisions.honestSeaDenominator). The
	// boxes overlap the coast, so a plain box count attributes a bird
	// watched from the shore to the sea; GADM covers land, so a record
	// off land carries no country and the difference is the water.
	//
	// This logic lived only in the -marine-only path at first, and the full
	// run quietly reverted it -- caught by TestRegionCountsLineUp, which
	// asserts a sea's denominator is not the sum of its own numerators.
	// Both paths compute it the same way now.
	//
	// Monthly, in two requests per sea rather than twenty-four: GBIF ORs
	// repeated gadmGid values, so one month facet over the box and one
	// restricted to USA|CAN give total and on-land, and the difference is
	// water (decisions.monthPresence).
	for _, id := range region.MarineIDs() {
		minLat, maxLat, minLon, maxLon, _ := region.MarineBox(id)
		off, _, err := c.MonthlyOffLandInBox(ctx, gbif.EBirdDatasetKey, "",
			countries, minLat, maxLat, minLon, maxLon)
		if err != nil {
			log.Fatalf("birdview-regions: %s total: %v", id, err)
		}
		var year int64
		for _, n := range off {
			year += n
		}
		totals[id] = year
		monthlyTotals[id] = off
	}
	for _, sp := range pelagic {
		for _, id := range region.MarineIDs() {
			minLat, maxLat, minLon, maxLon, _ := region.MarineBox(id)
			off, _, err := c.MonthlyOffLandInBox(ctx, gbif.EBirdDatasetKey, strings.ToLower(sp),
				countries, minLat, maxLat, minLon, maxLon)
			if err != nil {
				log.Fatalf("birdview-regions: %s in %s: %v", sp, id, err)
			}
			var year int64
			for _, n := range off {
				year += n
			}
			if year > 0 {
				regionsOf[sp] = append(regionsOf[sp],
					region.Reported{ID: id, Records: year, Months: off})
			}
		}
	}
	for _, id := range region.MarineIDs() {
		r, _ := region.Get(id)
		names[id] = r.Name
	}

	placed := 0
	for _, v := range regionsOf {
		for _, r := range v {
			if !r.Below {
				placed++
				break
			}
		}
	}
	log.Printf("birdview-regions: %d of %d species placed in at least one region", placed, len(ds))

	// What the join discarded, named and counted (product.joinsReportDrops).
	//
	// A record under a name matching no drawn species still reaches the
	// denominator of every share, so it is never absent from the data --
	// only from any numerator. That is right for a bird BirdView does not
	// draw and a total loss for one it does, and the two are
	// indistinguishable unless the join says which names it dropped.
	//
	// Measured before this existed: 37 names, 10,015,688 records, of which
	// 9,954,362 belonged to ten species BirdView draws. Hairy Woodpecker
	// alone was 7,310,342 of them, leaving the bird showing 0.7% of its own
	// record (decisions.regionJoinNormalizes).
	var orphans []string
	var orphanTotal, everything int64
	for _, n := range totals {
		everything += n
	}
	for n, c := range orphan {
		orphans = append(orphans, fmt.Sprintf("%-32s %12d", n, c))
		orphanTotal += c
	}
	sort.Slice(orphans, func(i, j int) bool { return orphans[i][32:] > orphans[j][32:] })
	log.Printf("birdview-regions: %d names matched no drawn species, %d records (%.2f%% of all)",
		len(orphans), orphanTotal, 100*float64(orphanTotal)/float64(everything))
	for i, o := range orphans {
		if i == 20 {
			log.Printf("birdview-regions:   ... and %d more", len(orphans)-20)
			break
		}
		log.Printf("birdview-regions:   %s", o)
	}
	if f := float64(orphanTotal) / float64(everything); f > maxOrphanFraction {
		log.Fatalf("birdview-regions: %.2f%% of records match no drawn species, above the "+
			"%.2f%% bar. Every one of those is counted in the denominator of every share "+
			"and in no species' numerator. A name the occurrence data uses and "+
			"taxonomy.Occurrence does not know about belongs in AviListRenames or "+
			"CountedAs; refusing to write a table that loses them quietly (product.joinsReportDrops).",
			f*100, maxOrphanFraction*100)
	}

	if err := write(*outPath, ds, regionsOf, names, totals, monthlyTotals); err != nil {
		log.Fatalf("birdview-regions: %v", err)
	}
	log.Printf("birdview-regions: wrote %s", *outPath)
}

// measured is one region's counts, as BirdView counts them.
type measured struct {
	counts       map[string]int64
	monthly      map[string]*[12]int64
	present      map[string]bool
	withheld     map[string][12]bool
	total        int64
	monthlyTotal [12]int64
}

// measure turns one region's raw facets into the counts BirdView uses:
// eBird's records, plus iNaturalist's only where eBird is confirmed to be
// withholding the species (decisions.countEbirdFirst).
//
// **eBird alone, as a rule.** iNaturalist is 1.24% of land records, and its
// one real effect on the counts was to admit birds into regions on its own
// much smaller total: one photo of an escaped pet cleared a bar of 0.43
// records in New York. 985 of the 1,228 presences that rested on
// iNaturalist alone were escapes, domestic birds, and strays.
//
// **iNaturalist where eBird is silent by policy.** In a month eBird withholds
// a species in a region -- named by its Sensitive Species list and confirmed
// against the records (decisions.ebirdSilenceNotAbsence) -- iNaturalist's records are counted, because
// they are the only evidence there is. Without them Spotted Owl would read
// "0 in a billion" and appear in no state, which is eBird's privacy policy
// presented as a fact about the owl. Presence there is judged on
// iNaturalist's own total, since an eBird-sized denominator would make every
// iNaturalist count look forty-eight times rarer than it is.
//
// **Shared names.** Where eBird files two species under one name, both take
// that name's count (product.namesRecordsFiledUnder): eBird holds Northern Yellow Warbler's records
// only as Setophaga petechia, and without the copy the commonest warbler in
// North America would appear nowhere. The totals are taken before the copy,
// so a shared count is never divided into itself twice.
//
// aviName maps a lowercased name to the AviList spelling the Sensitive
// Species table is keyed on.
func measure(raw map[string]map[int]map[string]int64, gid string, aviName map[string]string) (measured, error) {
	m := measured{counts: map[string]int64{}, monthly: map[string]*[12]int64{},
		present: map[string]bool{}, withheld: map[string][12]bool{}}
	by := map[string]map[string]*[12]int64{}
	totals := map[string]int64{}
	for key, byMonth := range raw {
		// Every record carries a month, so the twelve must reconstruct the
		// year exactly. Checked over all 74 regions and both sources,
		// because decisions.monthPresence's whole model rests on it.
		var year, summed int64
		for _, n := range byMonth[0] {
			year += n
		}
		by[key] = map[string]*[12]int64{}
		for mo := 1; mo <= 12; mo++ {
			for name, n := range byMonth[mo] {
				summed += n
				k := taxonomy.Occurrence(name)
				if by[key][k] == nil {
					by[key][k] = &[12]int64{}
				}
				by[key][k][mo-1] += n
			}
		}
		if year != summed {
			return m, fmt.Errorf(
				"%s in %s: the year holds %d records and the twelve months hold %d, a "+
					"difference of %d. decisions.monthPresence assumes every record carries a month; it "+
					"does not, and a month filter would silently drop the difference",
				key, gid, year, summed, year-summed)
		}
		totals[key] = year
	}
	eb, in := by[gbif.EBirdDatasetKey], by[iNatDatasetKey]
	names := map[string]bool{}
	for k := range eb {
		names[k] = true
	}
	for k := range in {
		names[k] = true
	}
	for k := range names {
		var e, i [12]int64
		if v := eb[k]; v != nil {
			e = *v
		}
		if v := in[k]; v != nil {
			i = *v
		}
		var w [12]bool
		if sp, drawn := aviName[k]; drawn {
			if ws := taxonomy.Withholdings(sp, gid); len(ws) > 0 {
				w = taxonomy.ConfirmedMonths(ws, e, i)
			}
		}
		var eff [12]int64
		var eYear, iWithheld, year int64
		for mo := 0; mo < 12; mo++ {
			eff[mo] = e[mo]
			if w[mo] {
				eff[mo] += i[mo]
				iWithheld += i[mo]
			}
			eYear += e[mo]
			year += eff[mo]
		}
		if year == 0 {
			continue
		}
		m.counts[k] = year
		v := eff
		m.monthly[k] = &v
		m.total += year
		for mo := range eff {
			m.monthlyTotal[mo] += eff[mo]
		}
		if w != ([12]bool{}) {
			m.withheld[k] = w
		}
		// Present on eBird's own evidence against eBird's own total, or --
		// only where eBird is withholding it -- on iNaturalist's against
		// iNaturalist's. Never iNaturalist alone anywhere else.
		if t := totals[gbif.EBirdDatasetKey]; t > 0 && float64(eYear)/float64(t)*1e6 >= minPPM {
			m.present[k] = true
		}
		if t := totals[iNatDatasetKey]; t > 0 && iWithheld > 0 && float64(iWithheld)/float64(t)*1e6 >= minPPM {
			m.present[k] = true
		}
	}
	for _, sn := range taxonomy.SharedNames {
		src := strings.ToLower(sn.Source)
		for _, cover := range sn.Covers {
			c := strings.ToLower(cover)
			if c == src {
				continue
			}
			delete(m.counts, c)
			delete(m.monthly, c)
			delete(m.present, c)
			delete(m.withheld, c)
			if n, ok := m.counts[src]; ok {
				m.counts[c] = n
				v := *m.monthly[src]
				m.monthly[c] = &v
				m.present[c] = m.present[src]
				if w, ok := m.withheld[src]; ok {
					m.withheld[c] = w
				}
			}
		}
	}
	return m, nil
}

// landRec is one region's raw land facets, as cached for resume.
//
// Keyed by GADM id rather than by index, so adding a region to
// internal/region does not invalidate the regions already fetched.
type landRec struct {
	GID  string `json:"gid"`
	Name string `json:"name"`
	// Raw maps a GBIF dataset key to month to the verbatim facet it
	// returned: publisher's name, lowercased, to records. Month 0 is the
	// whole year, 1 to 12 the months. Unnormalized on purpose.
	Raw map[string]map[int]map[string]int64 `json:"raw"`
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

func write(path string, species []string, regionsOf map[string][]region.Reported,
	names map[string]string, totals map[string]int64, monthly map[string][12]int64) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	var b strings.Builder
	fmt.Fprintf(&b, "# BirdView per-region presence. Generated by cmd/birdview-regions.\n")
	fmt.Fprintf(&b, "# A species is listed in a region when it reaches %g record per million\n", minPPM)
	fmt.Fprintf(&b, "# of that region's bird records, in eBird or iNaturalist (imports.md, `gbif`).\n")
	fmt.Fprintf(&b, "#\n# region\tGADM id\tname\n")
	gids := make([]string, 0, len(names))
	for g := range names {
		gids = append(gids, g)
	}
	sort.Strings(gids)
	for _, g := range gids {
		fmt.Fprintf(&b, "#region\t%s\t%s\n", g, names[g])
	}
	// A region's own total, over every bird recorded there. Written as data
	// rather than as another comment: it is the denominator of the share
	// pages show, and a number a reader depends on should not live in a line
	// the parser skips.
	fmt.Fprintf(&b, "#\n# Counts are twelve, January first, separated by |. They sum to the\n")
	fmt.Fprintf(&b, "# year exactly: every record carries a month, and birdview-regions\n")
	fmt.Fprintf(&b, "# refuses to write unless the twelve reconstruct it (product.regionAndMonthFilter, decisions.monthPresence).\n")
	fmt.Fprintf(&b, "#\n# total\tregion id\tevery bird record there, by month\n")
	for _, g := range gids {
		fmt.Fprintf(&b, "total\t%s\t%s\n", g, joinMonths(monthly[g]))
	}
	fmt.Fprintf(&b, "#\n# A cell may end in ! and twelve 0/1 flags: the months eBird withholds the\n")
	fmt.Fprintf(&b, "# species there under its Sensitive Species list, confirmed against the\n")
	fmt.Fprintf(&b, "# records. A withheld month is missing by policy, not absent (decisions.ebirdSilenceNotAbsence).\n")
	fmt.Fprintf(&b, "#\n# A region id followed by %s is one where the species has records but stays\n", region.BelowMark)
	fmt.Fprintf(&b, "# under the bar: its records count, and it is not one of the species'\n")
	fmt.Fprintf(&b, "# regions (decisions.belowBarCounted).\n")
	fmt.Fprintf(&b, "#\n# scientific name\tregion[%s]:jan|feb|...|dec[!flags], comma separated\n", region.BelowMark)
	for _, sp := range species {
		rs := regionsOf[sp]
		sort.Slice(rs, func(i, j int) bool { return rs[i].ID < rs[j].ID })
		parts := make([]string, len(rs))
		for i, r := range rs {
			mark := ""
			if r.Below {
				mark = region.BelowMark
			}
			parts[i] = fmt.Sprintf("%s%s:%s%s", r.ID, mark, joinMonths(r.Months), withheldSuffix(r.Withheld))
		}
		fmt.Fprintf(&b, "%s\t%s\n", sp, strings.Join(parts, ","))
	}
	return os.WriteFile(path, []byte(b.String()), 0o644)
}

// withheldSuffix renders the months eBird is confirmed to withhold as
// `!` and twelve 0/1 flags, or nothing when it withholds none -- so the
// 99% of cells it never touches are unchanged byte for byte (decisions.ebirdSilenceNotAbsence).
func withheldSuffix(w [12]bool) string {
	var b strings.Builder
	set := false
	for _, on := range w {
		if on {
			set = true
			b.WriteByte('1')
		} else {
			b.WriteByte('0')
		}
	}
	if !set {
		return ""
	}
	return "!" + b.String()
}

// joinMonths renders twelve counts as `jan|feb|...|dec`.
func joinMonths(m [12]int64) string {
	var b strings.Builder
	for i, n := range m {
		if i > 0 {
			b.WriteByte('|')
		}
		b.WriteString(strconv.FormatInt(n, 10))
	}
	return b.String()
}

// refreshMarine recomputes the five sea rows and sea totals in place,
// leaving every land row exactly as it was (decisions.honestSeaDenominator).
//
// Two things change from what the full run writes today, and they are the
// whole of the fix:
//
//  1. A species' count in a sea becomes its OFF-LAND count in that box,
//     not its count in the box. The boxes overlap the coast -- all six of
//     the Black-capped Petrel's North Carolina records fall inside the
//     Atlantic box -- so the old number counted shore records as sea.
//
//  2. A sea's total becomes every bird record on water in that box, not the
//     sum over the species already judged pelagic. That sum was a
//     denominator over a restricted species set, which made a marine share
//     a different quantity from a land share while the pages divided by
//     both alike.
//
// The pelagic SET is taken from the table as it stands rather than
// re-measured, which is what keeps this cheap. A species that has newly
// become pelagic will not appear until the next full run; that is a
// deliberate trade and not an oversight, because re-measuring it costs a
// CountFor for every species in the dataset.
func refreshMarine(path string, species []string) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("reading the existing table: %w (run without -marine-only first)", err)
	}
	table, totals, monthly, err := region.LoadPresence(f)
	f.Close()
	if err != nil {
		return err
	}

	isSea := map[string]bool{}
	for _, id := range region.MarineIDs() {
		isSea[id] = true
	}
	// Who to ask about: whoever the table already places at sea.
	var pelagic []string
	for _, sp := range species {
		for _, r := range table[sp] {
			if isSea[r.ID] {
				pelagic = append(pelagic, sp)
				break
			}
		}
	}
	sort.Strings(pelagic)
	if len(pelagic) == 0 {
		return fmt.Errorf("no species is placed in a sea, so there is nothing to refresh; " +
			"run without -marine-only to build the marine rows first")
	}
	log.Printf("birdview-regions: refreshing %d sea-going species across %d seas",
		len(pelagic), len(isSea))

	c := gbif.NewClient()
	c.Concurrency = 2
	c.Interval = 400 * time.Millisecond
	ctx := context.Background()
	countries := []string{"US", "CA"}

	// Each sea's own denominator: every bird record on water in the box,
	// over all species rather than the pelagic ones.
	for _, id := range region.MarineIDs() {
		minLat, maxLat, minLon, maxLon, _ := region.MarineBox(id)
		offM, totalM, err := c.MonthlyOffLandInBox(ctx, gbif.EBirdDatasetKey, "",
			countries, minLat, maxLat, minLon, maxLon)
		if err != nil {
			return fmt.Errorf("%s total: %w", id, err)
		}
		var off, total int64
		for i := range offM {
			off += offM[i]
			total += totalM[i]
		}
		if off == 0 {
			return fmt.Errorf("%s holds no off-land bird records at all, which cannot be "+
				"right for an ocean; the gadmGid filter has probably changed shape", id)
		}
		log.Printf("birdview-regions:   %-9s %12d on water of %12d in box (%.1f%%)",
			id, off, total, 100*float64(off)/float64(total))
		totals[id] = off
		monthly[id] = offM
	}

	// Per species, off land, one request each.
	var mu sync.Mutex
	var wg sync.WaitGroup
	sem := make(chan struct{}, c.Concurrency)
	errs := map[string]error{}
	found := map[string][]region.Reported{}
	for _, sp := range pelagic {
		wg.Add(1)
		go func(sp string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			var rs []region.Reported
			for _, id := range region.MarineIDs() {
				minLat, maxLat, minLon, maxLon, _ := region.MarineBox(id)
				offM, _, err := c.MonthlyOffLandInBox(ctx, gbif.EBirdDatasetKey,
					strings.ToLower(sp), countries, minLat, maxLat, minLon, maxLon)
				if err != nil {
					mu.Lock()
					errs[sp] = err
					mu.Unlock()
					return
				}
				var off int64
				for _, n := range offM {
					off += n
				}
				if off > 0 {
					rs = append(rs, region.Reported{ID: id, Records: off, Months: offM})
				}
			}
			mu.Lock()
			found[sp] = rs
			mu.Unlock()
		}(sp)
	}
	wg.Wait()
	// Refuse the whole table rather than write one where a failed request is
	// indistinguishable from a bird that is not there -- the same rule
	// birdview-status applies to its date queries.
	if len(errs) > 0 {
		var lines []string
		for sp, err := range errs {
			lines = append(lines, fmt.Sprintf("%s: %v", sp, err))
		}
		sort.Strings(lines)
		return fmt.Errorf("%d species failed, refusing to write a table that would "+
			"report them absent from every sea:\n  %s", len(errs), strings.Join(lines, "\n  "))
	}

	// Swap the sea rows, keep the land ones.
	var dropped, kept int
	for sp, rs := range found {
		var merged []region.Reported
		for _, r := range table[sp] {
			if !isSea[r.ID] {
				merged = append(merged, r)
			}
		}
		merged = append(merged, rs...)
		if len(rs) == 0 {
			dropped++
		} else {
			kept += len(rs)
		}
		table[sp] = merged
	}
	log.Printf("birdview-regions: %d sea placements kept, %d species left no sea at all",
		kept, dropped)

	// Land names are carried over verbatim rather than regenerated. The full
	// run reads each one back out of the occurrence data, which is how a
	// wrong GADM id surfaces as a wrong name instead of as silence;
	// substituting BirdView's own names here would quietly disarm that check
	// and rewrite five outlying-island rows as a side effect of touching the
	// seas.
	names, err := readRegionNames(path)
	if err != nil {
		return err
	}
	for _, id := range region.MarineIDs() {
		if rg, ok := region.Get(id); ok {
			names[id] = rg.Name // nobody else names an ocean for us
		}
	}
	return write(path, species, table, names, totals, monthly)
}

// readRegionNames recovers the `#region` header lines from the table on disk.
// They are comments, so LoadPresence skips them, and they are the only record
// of the name GBIF reported for each region.
func readRegionNames(path string) (map[string]string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	out := map[string]string{}
	for _, line := range strings.Split(string(b), "\n") {
		rest, ok := strings.CutPrefix(line, "#region\t")
		if !ok {
			continue
		}
		id, name, ok := strings.Cut(rest, "\t")
		if !ok {
			return nil, fmt.Errorf("malformed #region line: %q", line)
		}
		out[id] = name
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("%s carries no #region lines to preserve", path)
	}
	return out, nil
}
