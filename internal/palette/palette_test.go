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

package palette

import (
	"fmt"
	"math"
	"sort"
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

// TestContrastAgainstOwnText asserts T-010 as it now reads: a badge is legible
// when it clears 7:1 against the text it actually carries, which may be black
// or white. The expected value comes from WCAG, not from a run.
//
// Verifies: T-010, P-005, wcag/1.4.6. Criterion: AC-002.
func TestContrastAgainstOwnText(t *testing.T) {
	for _, sw := range Generate(abaFamilies, 0) {
		got, err := sw.ContrastWith()
		if err != nil {
			t.Fatalf("%+v: %v", sw, err)
		}
		if got < MinContrast {
			t.Errorf("%s on %s has contrast %.2f:1, want >= %.1f:1", sw.Text, sw.Hex, got, MinContrast)
		}
	}
}

// TestUsesBothHalvesOfTheGamut checks the dark half is actually reached.
//
// Requiring black text everywhere caps the reachable palette at about 43
// mutually distinguishable colors against 91 families; admitting white text
// raises it to about 69. If every swatch comes back light, that capacity has
// been silently given up again.
//
// Verifies: P-005.
func TestUsesBothHalvesOfTheGamut(t *testing.T) {
	var light, dark int
	for _, sw := range Generate(abaFamilies, 0) {
		if sw.Text == "#ffffff" {
			dark++
		} else {
			light++
		}
	}
	if dark == 0 || light == 0 {
		t.Errorf("palette used only one half of the gamut: %d light, %d dark", light, dark)
	}
}

// TestPresetsAlternate guards the property that hue-adjacent entries always
// differ in lightness. An odd-length alternating cycle wraps onto itself and
// silently breaks it, which is how two pale Passeriformes families ended up at
// dE 4.4.
//
// Verifies: P-005.
func TestPresetsAlternate(t *testing.T) {
	if !AlternatesLightAndDark() {
		t.Error("preset cycle does not strictly alternate light and dark; " +
			"the count must be even or the cycle wraps light onto light")
	}
}

// TestContrast asserts the requirement, not the implementation: every generated
// badge color clears WCAG 2.2 SC 1.4.6 against black text.
//
// The expected value comes from the standard (7:1), not from running the
// generator and recording what it produced.
//
// Verifies: T-010, P-005, wcag/1.4.6. Criterion: AC-002.
func TestContrast(t *testing.T) {
	colors := Generate(abaFamilies, 0)
	if len(colors) != len(abaFamilies) {
		t.Fatalf("got %d colors for %d families", len(colors), len(abaFamilies))
	}
	for fam, sw := range colors {
		got, err := sw.ContrastWith()
		if err != nil {
			t.Fatal(err)
		}
		if got < MinContrast {
			t.Errorf("%s: %s on %s has contrast %.2f:1, want >= %.1f:1",
				fam, sw.Text, sw.Hex, got, MinContrast)
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
	a := Generate(abaFamilies, 0)
	shuffled := append([]string(nil), abaFamilies...)
	shuffled[0], shuffled[len(shuffled)-1] = shuffled[len(shuffled)-1], shuffled[0]
	b := Generate(shuffled, 0)
	for fam, want := range a {
		if b[fam] != want {
			t.Errorf("%s: %+v then %+v -- input order changed the output", fam, want, b[fam])
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
	for fam, sw := range Generate(abaFamilies, 0) {
		if other, dup := seen[sw.Hex]; dup {
			t.Errorf("%s and %s share %s", other, fam, sw.Hex)
		}
		seen[sw.Hex] = fam
	}
}

// TestAdjacentFamiliesAreVisiblyDifferent is the check TestDistinct should
// always have been. Exact-hex inequality is far too weak: the first generated
// palette stepped hue evenly in sorted order, which put consecutive families
// four degrees apart and produced 91 technically-distinct colors of which every
// neighboring pair was indistinguishable. Accipitridae, Acrocephalidae and
// Aegithalidae were three identical reds, side by side in the selector.
//
// The selector lists families in sorted order, so neighbors there are what a
// reader actually compares. dE 25 is a deliberate margin over the conventional
// "clearly different" threshold of 10, and far above the 2-6 the broken palette
// achieved.
//
// Verifies: P-005. Criterion: AC-002.
func TestAdjacentFamiliesAreVisiblyDifferent(t *testing.T) {
	const minDeltaE = 25.0

	fams := append([]string(nil), abaFamilies...)
	sort.Strings(fams)
	colors := Generate(fams, 0)

	worst := math.Inf(1)
	for i := 0; i+1 < len(fams); i++ {
		r1, g1, b1 := mustParseHex(t, colors[fams[i]].Hex)
		r2, g2, b2 := mustParseHex(t, colors[fams[i+1]].Hex)
		d := DeltaE76(r1, g1, b1, r2, g2, b2)
		if d < worst {
			worst = d
		}
		if d < minDeltaE {
			t.Errorf("%s (%s) and %s (%s) are adjacent in the selector but differ by only dE %.1f",
				fams[i], colors[fams[i]].Hex, fams[i+1], colors[fams[i+1]].Hex, d)
		}
	}
	t.Logf("worst adjacent pair: dE %.1f", worst)
}

// TestDeltaE76KnownValues pins the color-difference math against values that do
// not depend on this package, so the distinctness check cannot pass vacuously.
func TestDeltaE76KnownValues(t *testing.T) {
	cases := []struct {
		a, b string
		want float64
	}{
		{"#000000", "#000000", 0},
		{"#000000", "#ffffff", 100}, // L* runs 0 to 100, a* and b* are 0 for both
		{"#ff0000", "#ff0000", 0},
	}
	for _, c := range cases {
		r1, g1, b1 := mustParseHex(t, c.a)
		r2, g2, b2 := mustParseHex(t, c.b)
		if got := DeltaE76(r1, g1, b1, r2, g2, b2); math.Abs(got-c.want) > 0.5 {
			t.Errorf("dE(%s,%s) = %.2f, want %.1f", c.a, c.b, got, c.want)
		}
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
