package region

import (
	"regexp"
	"strings"
	"testing"
)

// TestEachGroupingPartitionsTheRegions guards the invariant the filter UI
// needs: within one grouping, the zones partition the regions. A region in no
// zone of a grouping is unreachable from that block, and one in two zones of
// the same grouping appears twice in it.
//
// The invariant is per grouping, not global: a region is deliberately in two
// zones overall, one from each grouping (P-013).
//
// Verifies: P-011, P-013.
func TestEachGroupingPartitionsTheRegions(t *testing.T) {
	if len(Groupings) < 2 {
		t.Fatalf("expected at least the two groupings of P-013, got %d", len(Groupings))
	}
	for _, g := range Groupings {
		seen := map[string]int{}
		for _, z := range g.Zones {
			got := InZone(g.Name, z)
			if len(got) == 0 {
				t.Errorf("%s: zone %q holds no regions", g.Name, z)
			}
			for _, r := range got {
				seen[r.ID]++
			}
		}
		if len(seen) != Len() {
			t.Errorf("%s: zones cover %d regions, the table has %d", g.Name, len(seen), Len())
		}
		for id, n := range seen {
			if n != 1 {
				t.Errorf("%s: %s appears in %d zones", g.Name, id, n)
			}
		}
		// And every region's own zone for this grouping is one the UI
		// offers, so a typo in the table cannot hide a region from every
		// group at once.
		known := map[string]bool{}
		for _, z := range g.Zones {
			known[z] = true
		}
		for _, id := range append(LandIDs(), MarineIDs()...) {
			r, _ := Get(id)
			if !known[r.Zone(g.Name)] {
				t.Errorf("%s: %s names zone %q, which is not offered",
					g.Name, id, r.Zone(g.Name))
			}
		}
	}
}

// TestGroupingsAreDistinct checks that the two groupings actually group
// differently. Two identical groupings would render two identical blocks and
// satisfy every other test here, while telling the reader nothing -- and the
// whole point of CR-023 is that the measured grouping disagrees with the
// hand-drawn one.
//
// Verifies: P-013.
func TestGroupingsAreDistinct(t *testing.T) {
	// Alaska is the documented disagreement: the clustering files it with
	// British Columbia, geography files it with the Yukon.
	ak, _ := Get("USA.2_1")
	if ak.Zone(ByBirds) == ak.Zone(ByGeography) {
		t.Errorf("Alaska is in %q under both groupings; CR-023 records that they disagree",
			ak.Zone(ByBirds))
	}
	for _, g := range Groupings {
		if g.Note == "" {
			t.Errorf("grouping %q has no note saying whether it is measured or drawn", g.Name)
		}
	}
	if ZonesOf("no such grouping") != nil {
		t.Error("ZonesOf invented a grouping that does not exist")
	}
}

// TestCodesAreUniqueAndShort checks the chip labels. Two regions sharing a
// code is indistinguishable on screen, and a long one breaks the grid the
// family filter already set the width of.
//
// Verifies: P-011.
func TestCodesAreUniqueAndShort(t *testing.T) {
	seen := map[string]string{}
	for _, id := range append(LandIDs(), MarineIDs()...) {
		r, _ := Get(id)
		if other, dup := seen[r.Code]; dup {
			t.Errorf("%s and %s share the code %q", other, id, r.Code)
		}
		seen[r.Code] = id
		if r.Code == "" || r.Name == "" {
			t.Errorf("%s has an empty code or name", id)
		}
		if len(r.Code) > 10 {
			t.Errorf("%s has a %d-character code %q; chips are sized for short ones",
				id, len(r.Code), r.Code)
		}
	}
}

// TestLandIDsLookLikeGADM catches the mistake that costs most: a hand-typed
// GADM id that is syntactically fine and points at the wrong place. It cannot
// tell Ohio from Oklahoma -- only the generator's name read-back does that --
// but it does catch a malformed id before a network run burns ten minutes.
//
// Verifies: P-011.
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
// Verifies: P-011.
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

// TestPelagicIsTheLastZone pins the ordering P-011 asks for, in **every**
// grouping. The sea is the one group that is not somewhere a reader lives, so
// it sits after the places that are, rather than between them.
//
// Verifies: P-011, P-013.
func TestPelagicIsTheLastZone(t *testing.T) {
	for _, g := range Groupings {
		zs := g.Zones
		if n := len(zs); n == 0 || zs[n-1] != "Pelagic" {
			t.Fatalf("%s: zones end with %q, want Pelagic last; order is %v",
				g.Name, zs[len(zs)-1], zs)
		}
		// And it is exactly the marine regions, so no land region hides in it.
		for _, r := range InZone(g.Name, "Pelagic") {
			if r.Kind != Marine {
				t.Errorf("%s: %s is in the Pelagic zone but is a %s region",
					g.Name, r.ID, r.Kind)
			}
		}
		if n := len(InZone(g.Name, "Pelagic")); n != len(MarineIDs()) {
			t.Errorf("%s: the Pelagic zone holds %d regions, there are %d marine ones",
				g.Name, n, len(MarineIDs()))
		}
	}
}
