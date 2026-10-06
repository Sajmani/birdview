package taxonomy

import (
	"strings"
	"testing"
)

// TestOccurrenceNormalizesEveryKnownName asserts the single normalization
// both readers of the occurrence data now share (CR-035).
//
// The bug this closes was not a missing table entry — every entry below
// already existed — but a second reader that never consulted them.
// birdview-status normalized and birdview-regions did not, so eBird's
// 7,310,342 Dryobates villosus records matched no drawn species and went
// into the denominator of every share and no numerator at all.
//
// Verifies: P-011, P-074, T-027.
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
	for from, to := range LumpedInto {
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
	// species in its own right, and CR-034 discloses the overlap rather
	// than folding it (which would assert a division nobody measured).
	if got := Occurrence("setophaga petechia"); got != "setophaga petechia" {
		t.Errorf("Occurrence folded a shared name to %q; CR-034 discloses those, "+
			"it does not reassign them", got)
	}
	if len(AviListRenames)+len(LumpedInto) < 5 {
		t.Fatal("the naming tables are nearly empty; this check is asserting almost nothing")
	}
}
