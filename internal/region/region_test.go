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

package region

import (
	"net/url"
	"regexp"
	"strings"
	"testing"
)

// TestZonesPartitionTheRegions checks the invariant the Region setting needs: the zones partition
// the regions. A region in no zone is unreachable, and one in two zones appears twice.
//
// Verifies: product.regionAndMonthFilter.
func TestZonesPartitionTheRegions(t *testing.T) {
	seen := map[string]int{}
	for _, z := range Zones {
		got := InZone(z)
		if len(got) == 0 {
			t.Errorf("zone %q holds no regions", z)
		}
		for _, r := range got {
			seen[r.ID]++
		}
	}
	if len(seen) != Len() {
		t.Errorf("zones cover %d regions, the table has %d", len(seen), Len())
	}
	for id, n := range seen {
		if n != 1 {
			t.Errorf("%s appears in %d zones", id, n)
		}
	}
	// And every region's own zone is one the setting offers, so a typo in the table cannot hide a
	// region.
	known := map[string]bool{}
	for _, z := range Zones {
		known[z] = true
	}
	for _, id := range append(LandIDs(), MarineIDs()...) {
		if r, _ := Get(id); !known[r.Zone()] {
			t.Errorf("%s names zone %q, which is not offered", id, r.Zone())
		}
	}
}

// TestCodesAreStandardOrAbsent checks the chip labels. A code is a postal
// abbreviation and is therefore exactly two letters, or it is absent — there
// is no third case, and **absent is the assertion that matters**: it is what
// stops an invented code being slipped in beside 64 official ones to make a
// tidier chip (decisions.regionUrlIsPostalCode).
//
// Two regions sharing a code would be indistinguishable on screen, so
// non-empty codes must still be unique. Empty ones collide by design.
//
// Verifies: product.regionAndMonthFilter.
func TestCodesAreStandardOrAbsent(t *testing.T) {
	seen := map[string]string{}
	coded := 0
	for _, id := range append(LandIDs(), MarineIDs()...) {
		r, _ := Get(id)
		if r.Name == "" {
			t.Errorf("%s has no name", id)
		}
		if r.Code == "" {
			continue
		}
		coded++
		if other, dup := seen[r.Code]; dup {
			t.Errorf("%s and %s share the code %q", other, id, r.Code)
		}
		seen[r.Code] = id
		if len(r.Code) != 2 {
			t.Errorf("%s has the %d-character code %q; a postal abbreviation is two "+
				"letters, and anything else is this project inventing one",
				id, len(r.Code), r.Code)
		}
	}
	// The ten islands and oceans have no standard abbreviation, so exactly
	// the 64 states and provinces may carry a code.
	if want := Len() - 10; coded != want {
		t.Errorf("%d regions carry a code, want %d; only the states and provinces "+
			"have a postal standard to cite", coded, want)
	}
}

// TestLandIDsLookLikeGADM catches the mistake that costs most: a hand-typed
// GADM id that is syntactically fine and points at the wrong place. It cannot
// tell Ohio from Oklahoma -- only the generator's name read-back does that --
// but it does catch a malformed id before a network run burns ten minutes.
//
// Verifies: product.regionAndMonthFilter.
func TestLandIDsLookLikeGADM(t *testing.T) {
	re := regexp.MustCompile(`^(USA|CAN|UMI)\.\d+_1$`)
	for _, id := range LandIDs() {
		if !re.MatchString(id) {
			t.Errorf("%q is not a GADM level-1 id", id)
		}
	}
	for _, id := range MarineIDs() {
		if !strings.HasPrefix(id, "SEA.") {
			t.Errorf("marine region %q does not use the SEA. prefix", id)
		}
		r, ok := Get(id)
		if !ok || r.Kind != Marine {
			t.Errorf("%q is not registered as a marine region", id)
		}
	}
}

// TestMarineBoxesAreDisjoint pins the property the ocean assignment rests on:
// a record falls in at most one ocean, so "every box with records" is a sound
// way to decide which seas a bird uses.
//
// This replaces an assertion that Arctic must be tested first, which encoded a
// claim that was simply false — the boxes abut at 66°N rather than overlapping,
// so no ordering matters. The test caught the error in its own comment.
//
// Verifies: product.regionAndMonthFilter.
func TestMarineBoxesAreDisjoint(t *testing.T) {
	ids := MarineIDs()
	if len(ids) < 2 {
		t.Fatalf("only %d marine boxes; disjointness is vacuous", len(ids))
	}
	type box struct {
		id                     string
		lat0, lat1, lon0, lon1 float64
	}
	var bs []box
	for _, id := range ids {
		a, b, c, d, ok := MarineBox(id)
		if !ok {
			t.Fatalf("%s has no box", id)
		}
		if a >= b || c >= d {
			t.Errorf("%s is degenerate: lat %g-%g, lon %g-%g", id, a, b, c, d)
		}
		bs = append(bs, box{id, a, b, c, d})
	}
	// Overlap means sharing interior area. Touching at an edge is allowed:
	// the ranges are inclusive, so a record exactly on a boundary counts
	// twice, which is one record in a million and not worth a gap.
	for i := range bs {
		for j := i + 1; j < len(bs); j++ {
			x, y := bs[i], bs[j]
			latOverlap := min(x.lat1, y.lat1) - max(x.lat0, y.lat0)
			lonOverlap := min(x.lon1, y.lon1) - max(x.lon0, y.lon0)
			if latOverlap > 0 && lonOverlap > 0 {
				t.Errorf("%s and %s overlap by %g° of latitude and %g° of longitude; "+
					"a record there would be counted in both oceans",
					x.id, y.id, latOverlap, lonOverlap)
			}
		}
	}
}

// TestPelagicIsTheLastZone pins the ordering product.regionAndMonthFilter asks for. The sea is the
// one zone that is not somewhere a reader lives, so it sits after the places that are.
//
// Verifies: product.regionAndMonthFilter.
func TestPelagicIsTheLastZone(t *testing.T) {
	if n := len(Zones); n == 0 || Zones[n-1] != "Pelagic" {
		t.Fatalf("zones end with %q, want Pelagic last; order is %v", Zones[len(Zones)-1], Zones)
	}
	// And it is exactly the marine regions, so no land region hides in it.
	for _, r := range InZone("Pelagic") {
		if r.Kind != Marine {
			t.Errorf("%s is in the Pelagic zone but is a %s region", r.ID, r.Kind)
		}
	}
	if n := len(InZone("Pelagic")); n != len(MarineIDs()) {
		t.Errorf("the Pelagic zone holds %d regions, there are %d marine ones", n, len(MarineIDs()))
	}
}

// TestSlugsAreUniqueAndResolve is the invariant every region link rests on.
// Two regions sharing a URL value makes one unreachable from every choice, and the
// collision is silent: the page still renders, the wrong birds appear.
//
// Verifies: product.regionAndMonthFilter. Criterion: app.regionUrlRoundTrips.
func TestSlugsAreUniqueAndResolve(t *testing.T) {
	seen := map[string]string{}
	for _, id := range append(LandIDs(), MarineIDs()...) {
		r, _ := Get(id)
		slug := r.Slug()
		if slug == "" {
			t.Errorf("%s has an empty URL value", id)
			continue
		}
		if other, dup := seen[slug]; dup {
			t.Errorf("%s and %s share the URL value %q", other, id, slug)
		}
		seen[slug] = id

		// URL-safe without escaping: a value that needs encoding is one that
		// will be mis-typed and mis-copied.
		if url.QueryEscape(slug) != slug {
			t.Errorf("%s has the URL value %q, which does not survive a URL unescaped", id, slug)
		}

		// Round-trip: what is rendered must resolve, and to the same region.
		back, ok := BySlug(slug)
		if !ok {
			t.Errorf("%s renders the URL value %q, which resolves to nothing", id, slug)
			continue
		}
		if back.ID != id {
			t.Errorf("URL value %q round-trips to %s, not %s", slug, back.ID, id)
		}
	}
	if len(seen) != Len() {
		t.Errorf("%d distinct URL values for %d regions", len(seen), Len())
	}
}

// TestSlugIsDataCode pins the rule (decisions.regionCodesFromData): a land region's URL value
// is its GADM code, the code its records are filed under; a sea box, which no source codes yet,
// keeps its name in PascalCase until Marine Regions replaces it.
//
// Verifies: product.regionAndMonthFilter. Criterion: app.regionUrlRoundTrips.
func TestSlugIsDataCode(t *testing.T) {
	for _, tc := range []struct{ id, want string }{
		{"USA.5_1", "USA.5_1"},       // a state is its GADM code
		{"CAN.2_1", "CAN.2_1"},       // so is a province
		{"UMI.4_1", "UMI.4_1"},       // and an island
		{"SEA.GULF", "GulfOfMexico"}, // a sea box keeps its name
		{"SEA.PAC", "NorthPacific"},  // without the gloss
	} {
		r, ok := Get(tc.id)
		if !ok {
			t.Fatalf("%s is not a region", tc.id)
		}
		if got := r.Slug(); got != tc.want {
			t.Errorf("%s (%s) has the URL value %q, want %q", tc.id, r.Name, got, tc.want)
		}
	}
	for _, id := range MarineIDs() {
		if _, ok := BySlug(id); ok {
			t.Errorf("the internal sea id %q resolves in a URL; it is BirdView's key, not a code any source assigns", id)
		}
	}
}

// TestRetiredValuesResolve keeps every region URL BirdView wrote before decisions.regionCodesFromData
// working: each postal code and each island name resolves, as retired, to its region, and no
// retired value is also a current one, which would make the redirect ambiguous.
//
// Verifies: tech.stableUrls, product.regionAndMonthFilter. Criterion: app.regionUrlRoundTrips.
func TestRetiredValuesResolve(t *testing.T) {
	for old, id := range map[string]string{
		"CA": "USA.5_1", "NY": "USA.33_1", "BC": "CAN.2_1", "ON": "CAN.9_1", "DC": "USA.9_1",
		"JohnstonAtoll": "UMI.4_1", "WakeIsland": "UMI.9_1",
	} {
		r, ok := Retired(old)
		if !ok || r.ID != id {
			t.Errorf("the retired value %q resolves to %v (%v), want %s", old, r.ID, ok, id)
		}
		if _, current := BySlug(old); current {
			t.Errorf("%q is both retired and current", old)
		}
	}
	n := 0
	for _, id := range LandIDs() {
		r, _ := Get(id)
		if r.Code != "" {
			if back, ok := Retired(r.Code); !ok || back.ID != id {
				t.Errorf("%s's postal code %q does not resolve as retired", id, r.Code)
			}
			n++
		}
	}
	if n < 60 {
		t.Errorf("only %d land regions have postal codes; the check cannot see the old URLs", n)
	}
	if _, ok := Retired("GulfOfMexico"); ok {
		t.Error("GulfOfMexico resolves as retired, but the sea boxes keep their names until Marine Regions replaces them")
	}
}

// TestPresenceReadsWithheldFlags asserts the table's optional withheld
// suffix round-trips, and that a malformed one is refused rather than read
// as "nothing withheld".
//
// Verifies: decisions.ebirdSilenceNotAbsence. Criterion: app.withheldMonthsExplained.
func TestPresenceReadsWithheldFlags(t *testing.T) {
	const table = "total\tUSA.10_1\t1|1|1|1|1|1|1|1|1|1|1|1\n" +
		"Passerina ciris\tUSA.10_1:0|0|0|12483|2349|1034|905|1680|2759|0|0|0!111000000111\n"
	p, _, _, err := LoadPresence(strings.NewReader(table))
	if err != nil {
		t.Fatal(err)
	}
	rs := p["Passerina ciris"]
	if len(rs) != 1 {
		t.Fatalf("got %d cells", len(rs))
	}
	want := [12]bool{true, true, true, false, false, false, false, false, false, true, true, true}
	if rs[0].Withheld != want {
		t.Errorf("withheld = %v, want %v", rs[0].Withheld, want)
	}
	if rs[0].Records != 12483+2349+1034+905+1680+2759 {
		t.Errorf("records = %d: the flags must not disturb the counts", rs[0].Records)
	}
	for _, bad := range []string{"!11100", "!11100000011x"} {
		_, _, _, err := LoadPresence(strings.NewReader("total\tUSA.10_1\t1|1|1|1|1|1|1|1|1|1|1|1\n" +
			"Passerina ciris\tUSA.10_1:0|0|0|1|1|1|1|1|1|0|0|0" + bad + "\n"))
		if err == nil {
			t.Errorf("flags %q were accepted; a malformed mask must not read as nothing withheld", bad)
		}
	}
}

// TestCountyParentFromCode asserts a county's state is read from its code alone, so a county is
// recognized whether or not it is registered (decisions.countyListingRule).
//
// Verifies: product.regionAndMonthFilter. Criterion: app.countyListingApplied.
func TestCountyParentFromCode(t *testing.T) {
	for id, want := range map[string]string{"USA.33.24_1": "USA.33_1", "CAN.9.12_1": "CAN.9_1", "USA.33_1": "", "SEA.GULF": "", "GulfOfMexico": ""} {
		got, ok := CountyParent(id)
		if got != want || ok != (want != "") {
			t.Errorf("CountyParent(%q) = %q, %v; want %q", id, got, ok, want)
		}
	}
}
