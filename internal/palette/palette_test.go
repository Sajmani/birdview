package palette

import (
	"fmt"
	"testing"
)

// abaFamilies is a representative slice of the real ABA family list, chosen to
// span the hue wheel and to include the families whose hues are hardest to make
// legible (blues and violets have low luminance at a given HSL lightness).
var abaFamilies = []string{
	"Anatidae", "Odontophoridae", "Phasianidae", "Podicipedidae", "Columbidae",
	"Cuculidae", "Caprimulgidae", "Apodidae", "Trochilidae", "Rallidae",
	"Gruidae", "Recurvirostridae", "Charadriidae", "Scolopacidae", "Laridae",
	"Gaviidae", "Procellariidae", "Ardeidae", "Threskiornithidae", "Cathartidae",
	"Pandionidae", "Accipitridae", "Tytonidae", "Strigidae", "Alcedinidae",
	"Picidae", "Falconidae", "Tyrannidae", "Vireonidae", "Corvidae",
	"Paridae", "Hirundinidae", "Regulidae", "Sittidae", "Certhiidae",
	"Troglodytidae", "Sturnidae", "Turdidae", "Mimidae", "Bombycillidae",
	"Passeridae", "Fringillidae", "Passerellidae", "Icteridae", "Parulidae",
	"Cardinalidae", "Thraupidae", "Alaudidae", "Motacillidae", "Calcariidae",
}

// TestContrast asserts the requirement, not the implementation: every generated
// badge color clears WCAG 2.2 SC 1.4.6 against black text.
//
// The expected value comes from the standard (7:1), not from running the
// generator and recording what it produced.
//
// Verifies: T-010, P-005, wcag/1.4.6. Criterion: AC-002.
func TestContrast(t *testing.T) {
	colors := Generate(abaFamilies)
	if len(colors) != len(abaFamilies) {
		t.Fatalf("got %d colors for %d families", len(colors), len(abaFamilies))
	}
	for fam, hex := range colors {
		r, g, b := mustParseHex(t, hex)
		if got := ContrastOnBlack(r, g, b); got < MinContrast {
			t.Errorf("%s: %s has contrast %.2f:1 against black, want >= %.1f:1",
				fam, hex, got, MinContrast)
		}
	}
}

// TestKnownContrastValues pins the luminance math against values computed from
// the WCAG formula independently of this package, so a bug in RelativeLuminance
// cannot make TestContrast pass vacuously.
//
// Verifies: T-010.
func TestKnownContrastValues(t *testing.T) {
	cases := []struct {
		hex  string
		want float64 // (L+0.05)/0.05 per WCAG 2.2
	}{
		{"#ffffff", 21.0},
		{"#000000", 1.0},
		// Mid gray. Worked through by hand from the WCAG formula rather than
		// recalled: s = 128/255 = 0.50196, linear = ((s+0.055)/1.055)^2.4 =
		// 0.21586, and since the three channels are equal and the coefficients
		// sum to 1, L = 0.21586. Contrast = (L+0.05)/0.05 = 5.317.
		// An earlier draft of this table said 4.6 from memory, and this test
		// caught it -- which is the point of deriving expectations from the
		// standard instead of from a run.
		{"#808080", 5.317},
		{"#5eead4", 14.2}, // teal, from the earlier hand-curated palette
		{"#fef08a", 18.0}, // pale yellow
	}
	for _, c := range cases {
		r, g, b := mustParseHex(t, c.hex)
		got := ContrastOnBlack(r, g, b)
		if diff := got - c.want; diff > 0.1 || diff < -0.1 {
			t.Errorf("%s: contrast %.2f, want %.1f", c.hex, got, c.want)
		}
	}
}

// TestDeterministic asserts the same families always yield the same colors, so
// a rebuild does not reshuffle badges for no reason.
//
// Verifies: P-005.
func TestDeterministic(t *testing.T) {
	a := Generate(abaFamilies)
	shuffled := append([]string(nil), abaFamilies...)
	shuffled[0], shuffled[len(shuffled)-1] = shuffled[len(shuffled)-1], shuffled[0]
	b := Generate(shuffled)
	for fam, want := range a {
		if b[fam] != want {
			t.Errorf("%s: %s then %s -- input order changed the output", fam, want, b[fam])
		}
	}
}

// TestDistinct checks families do not share a badge color. The earlier
// hand-curated palette gave Falconidae and Icteridae the same hex, in a table
// whose whole purpose was telling families apart.
//
// Verifies: P-005.
func TestDistinct(t *testing.T) {
	seen := map[string]string{}
	for fam, hex := range Generate(abaFamilies) {
		if other, dup := seen[hex]; dup {
			t.Errorf("%s and %s share %s", other, fam, hex)
		}
		seen[hex] = fam
	}
}

func mustParseHex(t *testing.T, hex string) (uint8, uint8, uint8) {
	t.Helper()
	var r, g, b uint8
	if _, err := fmt.Sscanf(hex, "#%02x%02x%02x", &r, &g, &b); err != nil {
		t.Fatalf("bad hex %q: %v", hex, err)
	}
	return r, g, b
}
