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

package taxonomy

import (
	"strings"
	"testing"
)

// TestOccurrenceNormalizesEveryKnownName asserts the single normalization
// both readers of the occurrence data now share (decisions.regionJoinNormalizes).
//
// The bug this closes was not a missing table entry — every entry below
// already existed — but a second reader that never consulted them.
// birdview-status normalized and birdview-regions did not, so eBird's
// 7,310,342 Dryobates villosus records matched no drawn species and went
// into the denominator of every share and no numerator at all.
//
// Verifies: product.regionAndMonthFilter, product.shareOfRecords, tech.joinsReportDrops.
func TestOccurrenceNormalizesEveryKnownName(t *testing.T) {
	for from, to := range AviListRenames {
		if got := Occurrence(from); got != strings.ToLower(to) {
			t.Errorf("Occurrence(%q) = %q, want %q", from, got, strings.ToLower(to))
		}
		// The occurrence data is lowercased before it gets here, so the
		// lowercase form has to work too — it is the only form that ever
		// arrives in practice.
		if got := Occurrence(strings.ToLower(from)); got != strings.ToLower(to) {
			t.Errorf("Occurrence(%q) = %q, want %q", strings.ToLower(from), got, strings.ToLower(to))
		}
	}
	for _, c := range CountedAs {
		from, to := c.Published, c.Species
		if got := Occurrence(strings.ToLower(from)); got != strings.ToLower(to) {
			t.Errorf("Occurrence(%q) = %q, want %q", from, got, strings.ToLower(to))
		}
	}
	// A name nobody renamed passes through, lowercased and trimmed.
	for _, n := range []string{"Turdus migratorius", "  turdus migratorius  "} {
		if got := Occurrence(n); got != "turdus migratorius" {
			t.Errorf("Occurrence(%q) = %q, want it unchanged", n, got)
		}
	}
	// A shared name is NOT normalized away: Setophaga petechia is a drawn
	// species in its own right, and decisions.sharedCountShownOnBoth discloses the overlap rather
	// than folding it (which would assert a division nobody measured).
	if got := Occurrence("setophaga petechia"); got != "setophaga petechia" {
		t.Errorf("Occurrence folded a shared name to %q; decisions.sharedCountShownOnBoth discloses those, "+
			"it does not reassign them", got)
	}
	if len(AviListRenames)+len(CountedAs) < 5 {
		t.Fatal("the naming tables are nearly empty; this check is asserting almost nothing")
	}
}

// TestBinomialReducesRankWithoutInventingSpecies asserts the rank reduction
// both readers of the occurrence data share, and the one ordering rule
// inside it that is load-bearing.
//
// iNaturalist publishes at the rank the observer reached, so 2.6% of its
// ABA-area records sit under a trinomial and 0.2% under a hybrid. Folding
// the first is the point. Folding the second is a bug that cannot be seen
// downstream: truncated to two tokens a hybrid becomes a perfectly ordinary
// species name, joins cleanly, and credits a real bird with records of
// something that is not it.
//
// Watched fail: see acceptance.occurrenceReadersAgree.
//
// Verifies: product.regionAndMonthFilter, tech.joinsReportDrops. Criterion: acceptance.occurrenceReadersAgree.
func TestBinomialReducesRankWithoutInventingSpecies(t *testing.T) {
	species := map[string]string{
		"turdus migratorius":                "turdus migratorius",
		"setophaga coronata auduboni":       "setophaga coronata",
		"columba livia domestica":           "columba livia",
		"dryobates villosus harrisi":        "dryobates villosus",
		"junco hyemalis oreganus something": "junco hyemalis",
	}
	for in, want := range species {
		got, ok := Binomial(in)
		if !ok {
			t.Errorf("Binomial(%q) identified no species, want %q", in, want)
			continue
		}
		if got != want {
			t.Errorf("Binomial(%q) = %q, want %q", in, got, want)
		}
	}

	// Each of these identifies no species, and the assertion is NOT that
	// the name survives -- it is that no species appears. A hybrid rejected
	// too late survives as its first parent, which is the failure that
	// matters and the one a check for the hybrid's own name walks past.
	notSpecies := map[string]string{
		"larus smithsonianus × hyperboreus":      "larus smithsonianus",
		"setophaga coronata coronata × auduboni": "setophaga coronata",
		"anas acuta × mareca americana":          "anas acuta",
		"selasphorus":                            "",
		"":                                       "",
	}
	for in, parent := range notSpecies {
		got, ok := Binomial(in)
		if ok {
			t.Errorf("Binomial(%q) = %q, want no species: it identifies none", in, got)
		}
		if parent != "" && got == parent {
			t.Errorf("Binomial(%q) = %q -- the hybrid test runs after the truncation, "+
				"so these records would be filed as %s's", in, got, parent)
		}
	}
}
