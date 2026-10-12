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

// TestEBirdSensitiveIsWellFormed checks the restated list without trusting
// the scrape that produced it: every window parses, every region is a GADM
// id or a country, and WithheldByEBird -- now derived from the list -- holds
// exactly the species it should.
//
// Verifies: CR-037. Criterion: AC-113.
func TestEBirdSensitiveIsWellFormed(t *testing.T) {
	if len(EBirdSensitive) < 60 {
		t.Fatalf("only %d entries; the table is nearly empty and this asserts almost nothing",
			len(EBirdSensitive))
	}
	global := map[string]bool{}
	for _, w := range EBirdSensitive {
		if (w.From == "") != (w.To == "") {
			t.Errorf("%s in %s: a window needs both ends, got %q-%q", w.Species, w.Where, w.From, w.To)
		}
		if w.From != "" && (strings.Count(w.From, " ") != 1 || strings.Count(w.To, " ") != 1) {
			t.Errorf("%s in %s: window %q-%q is not in eBird's form", w.Species, w.Where, w.From, w.To)
		}
		switch {
		case w.Region == "":
			global[w.Species] = true
		case w.Region == "USA" || w.Region == "CAN":
		case strings.HasPrefix(w.Region, "USA.") || strings.HasPrefix(w.Region, "CAN."):
		default:
			t.Errorf("%s: region %q is neither a GADM id nor a country", w.Species, w.Region)
		}
	}
	// WithheldByEBird is derived: the nine global species, plus Black Rail,
	// which eBird lists year-round for both the US and Canada. Not Long-eared
	// Owl, listed year-round for the US but only seasonally in Canada; and
	// not Desertas Petrel, which the hand-kept table held and which is on no
	// entry of eBird's list.
	for sp := range global {
		if !WithheldByEBird[sp] {
			t.Errorf("%s is on eBird's global list but not in WithheldByEBird", sp)
		}
	}
	for sp, want := range map[string]bool{
		"Laterallus jamaicensis": true, "Asio otus": false, "Pterodroma deserta": false,
	} {
		if WithheldByEBird[sp] != want {
			t.Errorf("WithheldByEBird[%s] = %v, want %v", sp, !want, want)
		}
	}
	if len(WithheldByEBird) != len(global)+1 {
		t.Errorf("WithheldByEBird holds %d species, want the %d global plus Black Rail",
			len(WithheldByEBird), len(global))
	}
}

// TestWindowsCoverWholeMonthsOnly asserts the window arithmetic, including
// the two shapes that are easy to get wrong: a window that wraps the year
// end, and one that covers half a month.
//
// Painted Bunting's Florida listing is the pair Jan 01-Apr 15 and Oct 01-Dec
// 31. April is half covered and still holds 12,483 eBird records, so it must
// NOT be marked; marking it would put "withheld by eBird" on a month eBird
// plainly published.
//
// Verifies: CR-037. Criterion: AC-113.
func TestWindowsCoverWholeMonthsOnly(t *testing.T) {
	months := func(w Withholding) string {
		var b strings.Builder
		for _, on := range w.FullMonths() {
			if on {
				b.WriteByte('+')
			} else {
				b.WriteByte('-')
			}
		}
		return b.String()
	}
	cases := []struct {
		w    Withholding
		want string // Jan..Dec
	}{
		{Withholding{}, "++++++++++++"},
		{Withholding{From: "Jan 01", To: "Apr 15"}, "+++---------"},
		{Withholding{From: "Oct 01", To: "Dec 31"}, "---------+++"},
		{Withholding{From: "Apr 15", To: "Jul 15"}, "----++------"},
		// Wraps the year end: Nov 01 to Feb 28 is Nov, Dec, Jan, Feb.
		{Withholding{From: "Nov 01", To: "Feb 28"}, "++--------++"},
		// Shorter than any whole month: covers none.
		{Withholding{From: "Jun 10", To: "Jun 20"}, "------------"},
	}
	for _, c := range cases {
		if got := months(c.w); got != c.want {
			t.Errorf("%q-%q covers %s, want %s (Jan..Dec)", c.w.From, c.w.To, got, c.want)
		}
	}

	// A country entry covers that country's regions and no other.
	us := Withholding{Region: "USA"}
	if !us.Covers("USA.33_1") || us.Covers("CAN.9_1") || us.Covers("UMI.6_1") {
		t.Error("a USA entry must cover US states and nothing else")
	}
	ny := Withholding{Region: "USA.33_1"}
	if !ny.Covers("USA.33_1") || ny.Covers("USA.3_1") {
		t.Error("a GADM entry must cover exactly that region")
	}
	// The motivating entries are present, so a regenerated table that lost
	// them fails here rather than on a live page.
	for _, c := range []struct{ sp, gid string }{
		{"Passerina ciris", "USA.10_1"}, {"Asio flammeus", "USA.33_1"}, {"Asio otus", "USA.33_1"},
	} {
		if len(Withholdings(c.sp, c.gid)) == 0 {
			t.Errorf("no entry for %s in %s", c.sp, c.gid)
		}
	}
}

// TestConfirmedMonthsBelievesTheRecordsOverTheList asserts that a listing is
// acted on only where the records show it applied, using the counts of the
// cases CR-037 names.
//
// Verifies: CR-037. Criterion: AC-113.
func TestConfirmedMonthsBelievesTheRecordsOverTheList(t *testing.T) {
	marks := func(m [12]bool) string {
		var b strings.Builder
		for _, on := range m {
			if on {
				b.WriteByte('+')
			} else {
				b.WriteByte('-')
			}
		}
		return b.String()
	}
	pabu := []Withholding{
		{Species: "Passerina ciris", Region: "USA.10_1", From: "Jan 01", To: "Apr 15"},
		{Species: "Passerina ciris", Region: "USA.10_1", From: "Oct 01", To: "Dec 31"},
	}
	cases := []struct {
		name        string
		entries     []Withholding
		ebird, inat [12]int64
		want        string
	}{
		// Florida: nothing October to March, thousands outside.
		{"Painted Bunting, Florida", pabu,
			[12]int64{0, 0, 0, 12483, 2349, 1034, 905, 1680, 2759, 0, 0, 0},
			[12]int64{40, 38, 45, 120, 60, 30, 25, 40, 55, 30, 35, 42},
			"+++------+++"},
		// Utah: the list says Feb-May; the export holds more inside than out.
		{"Greater Sage-Grouse, Utah",
			[]Withholding{{Species: "Centrocercus urophasianus", Region: "USA", From: "Feb 01", To: "May 31"}},
			[12]int64{60, 400, 450, 420, 419, 60, 60, 60, 60, 60, 60, 61},
			[12]int64{},
			"------------"},
		// New York, whole year: eBird silent, iNaturalist 39.
		{"Long-eared Owl, New York",
			[]Withholding{{Species: "Asio otus", Region: "USA"}},
			[12]int64{},
			[12]int64{8, 6, 5, 2, 1, 1, 1, 1, 2, 3, 4, 5},
			"++++++++++++"},
		// Vermont, whole year: the listing never reached the export.
		{"Spruce Grouse, Vermont",
			[]Withholding{{Species: "Canachites canadensis", Region: "USA.46_1"}},
			[12]int64{100, 90, 100, 110, 120, 110, 110, 110, 120, 110, 110, 113},
			[12]int64{2, 2, 3, 3, 4, 4, 4, 4, 3, 3, 3, 2},
			"------------"},
		// No records from either source: nothing to confirm.
		{"no records", []Withholding{{Species: "x", Region: "USA"}},
			[12]int64{}, [12]int64{}, "------------"},
	}
	for _, c := range cases {
		if got := marks(ConfirmedMonths(c.entries, c.ebird, c.inat)); got != c.want {
			t.Errorf("%s: withheld %s, want %s (Jan..Dec)", c.name, got, c.want)
		}
	}
}
