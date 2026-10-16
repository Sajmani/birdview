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

package main

import (
	"strings"
	"testing"
	"time"

	"birdview/internal/gbif"
	"birdview/internal/status"
	"birdview/internal/taxonomy"
)

// The verbatim facet is not a species list, and treating it as one is what
// killed the first full run: a date query for "anas acuta × mareca americana"
// returned 404 and took the whole pass down with it.
//
// Watched fail: deleting the hybrid guard reports the hybrid's records
// arriving as Northern Pintail's; admitting single-token names lets the bare
// genus through. The first version of this test checked only for the hybrid's
// own name and PASSED against the broken code, because the failure is not
// that the name survives -- it is that the name is truncated to its first
// parent. The assertion had to change, not the code.
func TestToBinomialsReducesTheFacetToSpecies(t *testing.T) {
	in := map[string]map[string]int64{
		// A hybrid: not one species, and unqueryable besides.
		"anas acuta × mareca americana": {"USA.3_1": 7},
		// Genus only: never identified to a species.
		"selasphorus": {"USA.3_1": 3610},
		// Two subspecies and the bare binomial, all one bird.
		"setophaga coronata coronata": {"USA.3_1": 10},
		"setophaga coronata auduboni": {"USA.3_1": 20},
		"setophaga coronata":          {"USA.3_1": 5, "USA.5_1": 100},
		// A domestic form folds into its wild species.
		"columba livia domestica": {"USA.3_1": 40},
		"columba livia":           {"USA.3_1": 2},
	}
	got := toBinomials(in)

	// Checking for the hybrid's own name would pass vacuously: without the
	// guard the name does not survive, it is TRUNCATED to its first parent,
	// so the hybrid's records silently become Northern Pintail's. The
	// assertion has to be that no pintail appeared at all.
	if _, ok := got["anas acuta × mareca americana"]; ok {
		t.Error("hybrid survived under its own name")
	}
	if n, ok := got["anas acuta"]; ok {
		t.Errorf("hybrid was folded into its first parent: anas acuta = %v, "+
			"want no such species -- these records are not Northern Pintails", n)
	}
	for k := range got {
		if k == "selasphorus" {
			t.Error("bare genus survived; it was never identified to a species")
		}
	}
	if n := got["setophaga coronata"]["USA.3_1"]; n != 35 {
		t.Errorf("Yellow-rumped Warbler in USA.3_1 = %d, want 35 (5 + 10 + 20)", n)
	}
	if n := got["setophaga coronata"]["USA.5_1"]; n != 100 {
		t.Errorf("a region with only the binomial = %d, want 100", n)
	}
	if n := got["columba livia"]["USA.3_1"]; n != 42 {
		t.Errorf("Rock Pigeon in USA.3_1 = %d, want 42 (2 + 40)", n)
	}
	if len(got) != 2 {
		t.Errorf("got %d species %v, want 2", len(got), keys(got))
	}
}

func keys(m map[string]map[string]int64) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}

// A lump the occurrence data still publishes must not become a species of
// its own alongside the half BirdView carries.
//
// The first complete run admitted both halves of all three documented pairs,
// so a table written from it would have listed Whimbrel twice. The two
// databases disagree completely rather than partially -- eBird holds zero
// records under the split name, iNaturalist zero under the lump -- so the
// evidence has to be added, not chosen between.
//
// Watched fail: dropping mergeLumps from prepare leaves both halves in and
// this reports the lump surviving. It calls prepare, not mergeLumps: the
// first version called mergeLumps directly and passed with the call removed
// from main, verifying the function and not the pipeline.
//
// Verifies: product.scopeIsRecurringAvifauna. Criterion: acceptance.onlyNarrowsOneGroup.
func TestLumpsMergeIntoTheSpeciesBirdViewCarries(t *testing.T) {
	// eBird files everything under the lump; iNaturalist under the split.
	in := map[string]map[string]int64{
		"numenius phaeopus":   {"USA.5_1": 556569, "USA.44_1": 100},
		"numenius hudsonicus": {"USA.5_1": 17923},
		"gygis alba":          {"USA.12_1": 16008},
		"turdus migratorius":  {"USA.39_1": 1498010},
	}
	got := prepare(in)

	for _, lump := range []string{"numenius phaeopus", "gygis alba"} {
		if _, ok := got[lump]; ok {
			t.Errorf("%q survived as a species; it is the AVONET donor for the half "+
				"BirdView carries, so the two would draw identically", lump)
		}
	}
	// The evidence is added, not discarded and not chosen between.
	if n := got["numenius hudsonicus"]["USA.5_1"]; n != 556569+17923 {
		t.Errorf("Whimbrel in USA.5_1 = %d, want %d (%d eBird + %d iNaturalist)",
			n, 556569+17923, 556569, 17923)
	}
	// A region the split name never reached still arrives.
	if n := got["numenius hudsonicus"]["USA.44_1"]; n != 100 {
		t.Errorf("Whimbrel in USA.44_1 = %d, want 100: the lump's regions must carry over", n)
	}
	// Gygis candida had no records of its own at all and must now exist.
	if n := got["gygis candida"]["USA.12_1"]; n != 16008 {
		t.Errorf("Blue-billed White-Tern = %d, want 16008", n)
	}
	// Everything else is untouched.
	if n := got["turdus migratorius"]["USA.39_1"]; n != 1498010 {
		t.Errorf("an unrelated species was disturbed: %d", n)
	}
	// Setophaga is deliberately NOT merged: both halves are carried species
	// with their own AVONET rows, and merging would delete Mangrove Warbler.
	if _, _, ok := taxonomy.SpeciesCountedAs("Setophaga petechia"); ok {
		t.Error("Setophaga petechia must not be merged; it is a species BirdView carries")
	}
}

// TestBestRegionIsDeterministic asserts that the region a verdict is
// attributed to does not depend on Go's map iteration order.
//
// The bug: bestRegion ranked on (status, episodes) and walked the shapes map
// directly, so two regions tied on both were separated by whichever came out
// first. Fifteen rows of the committed table named a different region on
// every run and carried their episode, year, record and last-seen figures
// with them. Cooper's Hawk recurs in 39 episodes in both Alaska and Arizona,
// and the table reported 24 years and 216,391 records on one run and 18 and
// 119,194 on the next -- a generated table two runs over identical inputs
// disagree about.
//
// It matters more now that the table carries a date: product.undrawnSpeciesAnswer puts that date on
// a page as the reason a bird is not drawn, and a reason that changes when
// nothing changed is not a reason.
//
// Watched fail: restoring the bare map walk reports several different
// regions across the runs below. One run cannot catch it -- a single call is
// self-consistent whatever order it saw -- which is why this calls many
// times and compares.
//
// Verifies: product.undrawnSpeciesAnswer, decisions.scopeFromRecurrence.
func TestBestRegionIsDeterministic(t *testing.T) {
	since := time.Date(2001, 1, 1, 0, 0, 0, 0, time.UTC)
	recent := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)

	// Three regions the old rule could not separate: same status, same
	// episode count. Years and records differ, so the declared tiebreak has
	// a defensible answer -- the region it recurred in across the most years
	// -- rather than merely a stable one.
	shapes := map[string]status.Shape{
		"USA.2_1":  {Records: 216391, Dates: 300, Episodes: 39, Years: 24, LastSeen: recent},
		"USA.3_1":  {Records: 119194, Dates: 300, Episodes: 39, Years: 18, LastSeen: recent},
		"USA.5_1":  {Records: 900000, Dates: 300, Episodes: 39, Years: 24, LastSeen: recent},
		"CAN.13_1": {Records: 12, Dates: 9, Episodes: 2, Years: 2, LastSeen: recent},
	}

	gid, sh := bestRegion(shapes, since)
	for i := 0; i < 200; i++ {
		g, s := bestRegion(shapes, since)
		if g != gid || s != sh {
			t.Fatalf("run %d chose %s (%d years, %d records), run 0 chose %s (%d years, %d records); "+
				"the committed table is not a function of its inputs",
				i, g, s.Years, s.Records, gid, sh.Years, sh.Records)
		}
	}

	// And the answer is the declared one, not merely a repeatable one.
	// Years before records, because the model is recurrence and not
	// abundance: USA.5_1 ties USA.2_1 on episodes and years and has four
	// times the records, so records is what separates them.
	if gid != "USA.5_1" {
		t.Errorf("chose %s; want USA.5_1 -- highest status, most episodes, most "+
			"years, and the most records among those tied", gid)
	}
	if sh.Years != 24 {
		t.Errorf("chose a region with %d years; a region recurring across fewer "+
			"years is weaker evidence whatever its record count", sh.Years)
	}
}

// TestBothReadersOfTheOccurrenceDataAgreeOnTheSpecies asserts the invariant
// decisions.regionJoinNormalizes turned out to be three separate violations of: a verbatim
// occurrence name reaches the SAME AviList species whichever command reads
// it.
//
// The three asymmetries were found one at a time over two days, because
// each was invisible from inside the command that had it right. Nothing
// compared the two readers, so each fix left the next gap in place:
// birdview-regions applied neither AviListRenames nor CountedAs, and then,
// once it did, still did not reduce a trinomial to its species. This checks
// the property the three share instead of the three bugs, so a fourth
// reader, or a fourth table, cannot open the same hole again.
//
// It deliberately spans the two namespaces end to end -- prepare and then
// status.Scope on one side, taxonomy.Occurrence on the other -- because the
// two commands are required to agree on the name a join finally uses, not
// at every step on the way there. birdview-status resolves the genus move
// when it reads the table and birdview-regions when it writes one; that is
// a real difference and not a defect, and a check pinned to an intermediate
// value would call it one.
//
// Watched fail: see acceptance.occurrenceReadersAgree.
//
// Verifies: product.regionAndMonthFilter, product.shareOfRecords, tech.joinsReportDrops. Criterion: acceptance.occurrenceReadersAgree.
func TestBothReadersOfTheOccurrenceDataAgreeOnTheSpecies(t *testing.T) {
	// One verbatim name of each shape the facet actually contains, with the
	// species each is a record of.
	cases := map[string]string{
		"turdus migratorius":          "turdus migratorius",      // nothing to do
		"dryobates villosus":          "leuconotopicus villosus", // a genus move
		"dryobates villosus harrisi":  "leuconotopicus villosus", // ... filed under a subspecies
		"setophaga coronata auduboni": "setophaga coronata",      // a subspecies needing no rename
		"columba livia domestica":     "columba livia",           // a domestic form
		"numenius phaeopus":           "numenius hudsonicus",     // a lump
		"icterus bullockii":           "icterus bullockiorum",    // in no region at all before decisions.regionJoinNormalizes
	}

	// Scope returns the declared product.declaredExceptions exceptions whatever it is given, so
	// the species a row actually contributes is the difference. Subtracted
	// rather than filtered by name, so this stays correct when the
	// exceptions change.
	declared := map[string]bool{}
	for _, n := range status.Scope(nil) {
		declared[n] = true
	}

	for verbatim, want := range cases {
		// What birdview-status ends up joining on: the reduction in
		// prepare, then the AviList resolution in Scope.
		in := map[string]map[string]int64{verbatim: {"USA.5_1": 1000}}
		var rows []status.Row
		for k := range prepare(in) {
			rows = append(rows, status.Row{Name: k, Status: status.Established})
		}
		var added []string
		for _, n := range status.Scope(rows) {
			if !declared[n] {
				added = append(added, strings.ToLower(n))
			}
		}
		if len(added) != 1 {
			t.Errorf("birdview-status turned %q into %v, want exactly one species",
				verbatim, added)
			continue
		}
		// What birdview-regions ends up joining on, from the same name.
		fromRegions := taxonomy.Occurrence(verbatim)

		if added[0] != want {
			t.Errorf("birdview-status files %q under %q, want %q",
				verbatim, added[0], want)
		}
		if fromRegions != want {
			t.Errorf("birdview-regions files %q under %q, want %q",
				verbatim, fromRegions, want)
		}
		if added[0] != fromRegions {
			t.Errorf("the two readers disagree about %q: birdview-status says %q, "+
				"birdview-regions says %q. One of the two tables is in a taxonomy "+
				"the other cannot join on, which is how 7,310,342 Hairy Woodpecker "+
				"records came to belong to no species (decisions.regionJoinNormalizes)",
				verbatim, added[0], fromRegions)
		}
	}

	// The one place the two may differ, and why. A hybrid carries no status,
	// so birdview-status drops it; it is still a bird somebody saw, and so
	// still part of what a region's share is divided by, so birdview-regions
	// keeps it. What neither may do is call it a species -- truncated to two
	// tokens it reads as one, and these records are not that bird's.
	const hybrid = "larus smithsonianus × hyperboreus"
	if got := prepare(map[string]map[string]int64{hybrid: {"USA.5_1": 1000}}); len(got) != 0 {
		t.Errorf("birdview-status kept %v from a hybrid; it has no status to carry", keys(got))
	}
	if got := taxonomy.Occurrence(hybrid); got != hybrid {
		t.Errorf("birdview-regions turned the hybrid into %q, want it left verbatim: "+
			"folded to its first parent it credits the American Herring Gull with "+
			"records of a bird that is not one, and dropped outright it leaves the "+
			"region's denominator short of birds that were genuinely seen", got)
	}
}

// TestScopeCountsEBirdAndFillsOnlyItsDeclaredGaps asserts decisions.countEbirdFirst where it
// decides which species BirdView draws: recurrence is measured on eBird's
// dates, plus iNaturalist's only in months eBird is confirmed to withhold.
//
// Verifies: product.scopeIsRecurringAvifauna, decisions.countEbirdFirst. Criterion: acceptance.ebirdCountedInatFillsGaps.
func TestScopeCountsEBirdAndFillsOnlyItsDeclaredGaps(t *testing.T) {
	eb, in := gbif.EBirdDatasetKey, iNatDatasetKey

	// An escape only iNaturalist photographed, in three different years.
	// Under the old union this recurred; under the rule it has no dates.
	got := countedDates("pavo cristatus", "USA.33_1", map[string]map[string]int64{
		in: {"2019-05-01": 1, "2021-06-01": 1, "2023-07-01": 1}})
	if len(got) != 0 {
		t.Errorf("an iNaturalist-only escape kept %d dates; it would recur on unvetted "+
			"evidence alone", len(got))
	}

	// Long-eared Owl in New York: eBird withholds it all year, so
	// iNaturalist's dates are the evidence.
	got = countedDates("asio otus", "USA.33_1", map[string]map[string]int64{
		in: {"2019-01-10": 1, "2021-02-11": 2, "2023-12-12": 1}})
	if len(got) != 3 {
		t.Errorf("Long-eared Owl in New York kept %d of iNaturalist's 3 dates; eBird's "+
			"silence there is policy", len(got))
	}

	// Painted Bunting in Florida: eBird's summer dates, and iNaturalist's
	// winter ones -- but NOT iNaturalist's summer ones, which eBird covers.
	by := map[string]map[string]int64{
		eb: {"2022-05-01": 3000, "2022-06-01": 2000, "2022-07-01": 2000, "2022-08-01": 2000},
		in: {"2022-01-15": 5, "2022-06-15": 4},
	}
	got = countedDates("passerina ciris", "USA.10_1", by)
	if got["2022-01-15"] != 5 {
		t.Error("Painted Bunting's January iNaturalist date was dropped; eBird withholds January")
	}
	if got["2022-06-15"] != 0 {
		t.Error("Painted Bunting's June iNaturalist date was counted; eBird publishes June")
	}
	if got["2022-05-01"] != 3000 {
		t.Error("eBird's own dates were not all counted")
	}
}
