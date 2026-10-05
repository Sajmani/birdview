package main

import (
	"testing"
	"time"

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
// Verifies: P-001. Criterion: AC-093.
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
	if _, ok := taxonomy.LumpedInto["Setophaga petechia"]; ok {
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
// It matters more now that the table carries a date: P-075 puts that date on
// a page as the reason a bird is not drawn, and a reason that changes when
// nothing changed is not a reason.
//
// Watched fail: restoring the bare map walk reports several different
// regions across the runs below. One run cannot catch it -- a single call is
// self-consistent whatever order it saw -- which is why this calls many
// times and compares.
//
// Verifies: P-075, CR-028.
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
