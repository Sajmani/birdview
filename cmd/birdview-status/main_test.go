package main

import (
	"testing"

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
