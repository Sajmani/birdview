// Command birdview-regions builds the per-region presence table.
//
// For each US state and Canadian province it asks GBIF which bird names were
// reported there and how often, then writes the regions each of BirdView's
// species occurs in. The ingest reads the result; nothing queries the network
// at build or test time (T-006, and the no-network-in-tests binding).
//
// Two sources, because one is not enough:
//
//   - eBird, for depth: 1.24 billion records across the US and Canada.
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
// Relative, not absolute, for two reasons measured in sources.md. eBird's
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
// order of magnitude: before CR-035 this was 0.81%, and 9,954,362 of those
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
	// only way to review one (CR-031).
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
					// Normalized to the AviList name BEFORE matching, by the
					// same function birdview-status uses. Matching the raw
					// facet name is how eBird's 7,310,342 Dryobates villosus
					// records came to belong to no species at all (CR-035).
					n := taxonomy.Occurrence(rec.Name)
					r.counts[n] += rec.Records
					if float64(rec.Records)/float64(total)*1e6 >= minPPM {
						r.present[n] = true
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
	// page shows (P-074): dividing by all the birds the same observers saw
	// in the same place is what cancels out how hard they were looking.
	totals := map[string]int64{}
	for _, r := range results {
		if r.err != nil {
			log.Fatalf("birdview-regions: %s: %v", r.gid, r.err)
		}
		if r.name == "" {
			log.Printf("  WARNING: %s named nothing and is skipped", r.gid)
			continue
		}
		names[r.gid] = r.name
		for _, n := range r.counts {
			totals[r.gid] += n
		}
		for _, sp := range ds {
			if lc := strings.ToLower(sp); r.present[lc] {
				regionsOf[sp] = append(regionsOf[sp], region.Reported{
					ID: r.gid, Records: r.counts[lc]})
			}
		}
		// Every name that reached the denominator and no species. These are
		// records counted as "some other bird" on every tile, which is
		// correct for a bird BirdView does not draw and a silent loss for
		// one it does (G-006, T-027).
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
	// A sea counts only what is on OPEN WATER inside its box (CR-031). The
	// boxes overlap the coast, so a plain box count attributes a bird
	// watched from the shore to the sea; GADM covers land, so a record
	// off land carries no country and the difference is the water.
	//
	// This logic lived only in the -marine-only path at first, and the full
	// run quietly reverted it -- caught by TestRegionCountsLineUp, which
	// asserts a sea's denominator is not the sum of its own numerators.
	// Both paths compute it the same way now.
	for _, id := range region.MarineIDs() {
		minLat, maxLat, minLon, maxLon, _ := region.MarineBox(id)
		off, _, err := c.CountOffLandInBox(ctx, gbif.EBirdDatasetKey, "",
			countries, minLat, maxLat, minLon, maxLon)
		if err != nil {
			log.Fatalf("birdview-regions: %s total: %v", id, err)
		}
		totals[id] = off
	}
	for _, sp := range pelagic {
		for _, id := range region.MarineIDs() {
			minLat, maxLat, minLon, maxLon, _ := region.MarineBox(id)
			n, _, err := c.CountOffLandInBox(ctx, gbif.EBirdDatasetKey, strings.ToLower(sp),
				countries, minLat, maxLat, minLon, maxLon)
			if err != nil {
				log.Fatalf("birdview-regions: %s in %s: %v", sp, id, err)
			}
			if n > 0 {
				regionsOf[sp] = append(regionsOf[sp], region.Reported{ID: id, Records: n})
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

	// What the join discarded, named and counted (T-027).
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
	// record (CR-035).
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
			"CountedAs; refusing to write a table that loses them quietly (T-027).",
			f*100, maxOrphanFraction*100)
	}

	if err := write(*outPath, ds, regionsOf, names, totals); err != nil {
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

func write(path string, species []string, regionsOf map[string][]region.Reported,
	names map[string]string, totals map[string]int64) error {
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
	// A region's own total, over every bird recorded there. Written as data
	// rather than as another comment: it is the denominator of the share
	// pages show, and a number a reader depends on should not live in a line
	// the parser skips.
	fmt.Fprintf(&b, "#\n# total\tregion id\tevery bird record there\n")
	for _, g := range gids {
		fmt.Fprintf(&b, "total\t%s\t%d\n", g, totals[g])
	}
	fmt.Fprintf(&b, "#\n# scientific name\tregion:records, comma separated\n")
	for _, sp := range species {
		rs := regionsOf[sp]
		sort.Slice(rs, func(i, j int) bool { return rs[i].ID < rs[j].ID })
		parts := make([]string, len(rs))
		for i, r := range rs {
			parts[i] = fmt.Sprintf("%s:%d", r.ID, r.Records)
		}
		fmt.Fprintf(&b, "%s\t%s\n", sp, strings.Join(parts, ","))
	}
	return os.WriteFile(path, []byte(b.String()), 0o644)
}

// refreshMarine recomputes the five sea rows and sea totals in place,
// leaving every land row exactly as it was (CR-031).
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
	table, totals, err := region.LoadPresence(f)
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
		off, total, err := c.CountOffLandInBox(ctx, gbif.EBirdDatasetKey, "",
			countries, minLat, maxLat, minLon, maxLon)
		if err != nil {
			return fmt.Errorf("%s total: %w", id, err)
		}
		if off == 0 {
			return fmt.Errorf("%s holds no off-land bird records at all, which cannot be "+
				"right for an ocean; the gadmLevel0Gid facet has probably changed shape", id)
		}
		log.Printf("birdview-regions:   %-9s %12d on water of %12d in box (%.1f%%)",
			id, off, total, 100*float64(off)/float64(total))
		totals[id] = off
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
				off, _, err := c.CountOffLandInBox(ctx, gbif.EBirdDatasetKey,
					strings.ToLower(sp), countries, minLat, maxLat, minLon, maxLon)
				if err != nil {
					mu.Lock()
					errs[sp] = err
					mu.Unlock()
					return
				}
				if off > 0 {
					rs = append(rs, region.Reported{ID: id, Records: off})
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
	return write(path, species, table, names, totals)
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
