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

// Package palette generates badge colors.
//
// Two palettes are generated, not one: a color per taxonomic order, and a color
// per family *within* its order. That split is what makes the problem solvable.
// A colored badge must stay legible, which bounds the usable gamut at roughly
// 43 mutually distinguishable colors with black text, or 69 if white text is
// also allowed — against 91 ABA families. No assignment fits 91 into 43.
//
// Scoping family colors to their order reduces the requirement to the largest
// order, Passeriformes with 44 families, which does fit. Two families in
// different orders may share a color; the order badge tells them apart.
//
// Verifies: product.generatedBadgeColors, tech.badgeContrast, wcag/1.4.6.
package palette

import (
	"fmt"
	"math"
	"sort"
)

// MinContrast is WCAG 2.2 success criterion 1.4.6 (Contrast Enhanced) for
// normal text. BirdView targets it rather than 1.4.3's 4.5:1.
//
// Verifies: tech.badgeContrast.
const MinContrast = 7.0

const (
	black = "#000000"
	white = "#ffffff"
)

// Swatch is a background together with the text color that stays legible on it.
//
// Carrying the text color is what unlocks the dark half of the gamut: a badge
// is legible when it clears MinContrast against *its own* text, not
// specifically against black. Requiring black text discards every dark color
// and costs about a third of the reachable palette.
type Swatch struct {
	Hex  string `json:"hex"`
	Text string `json:"text"`
}

// invPhi is 1/golden ratio, used to pick how far apart consecutive entries sit.
const invPhi = 0.6180339887498949

// stride picks a step k, coprime to n, that walks the n evenly spaced hue slots
// in an order which keeps consecutive entries far apart.
//
// Both properties are needed and neither alone suffices. Spacing hues evenly
// maximizes the smallest gap between any two, which is what stops a large set
// from containing a near-identical pair -- golden-angle stepping does not
// guarantee this, and left Passeriformes with two families at dE 4.2. Walking
// those slots in sequence, though, gives neighbors adjacent hues, which is the
// original defect: a list is read in order. Striding by roughly n/phi slots
// delivers both.
func stride(n int) int {
	if n < 3 {
		return 1
	}
	k := int(math.Round(float64(n) * invPhi))
	for ; k > 1; k-- {
		if gcd(k, n) == 1 {
			return k
		}
	}
	return 1
}

func gcd(a, b int) int {
	for b != 0 {
		a, b = b, a%b
	}
	return a
}

// presets alternate light and dark backgrounds, so that hue-adjacent entries
// always differ in lightness too -- the largest available separation when hues
// are only a few degrees apart.
//
// The count must be EVEN. An odd-length alternating cycle wraps onto itself:
// with five presets the last light entry is followed by the first light entry,
// so one pair of adjacent hue slots shared a lightness. That is exactly how
// Phylloscopidae and Turdidae came out at dE 4.4, eight degrees apart in hue
// and both pale.
var presets = []struct {
	sat, start float64
	dark       bool
}{
	{0.88, 0.74, false},
	{0.80, 0.34, true},
	{0.55, 0.80, false},
	{0.62, 0.26, true},
	{0.97, 0.70, false},
	{0.95, 0.30, true},
}

// Generate assigns a swatch to each name, deterministically: hues occupy n
// evenly spaced slots, visited in a stride that keeps consecutive names far
// apart, with saturation and lightness cycling through presets.
//
// Callers pass a distinct seed per order, so families in different orders are
// not gratuitously given the same color even though they may collide.
//
// Verifies: product.generatedBadgeColors, tech.badgeContrast.
func Generate(names []string, seed float64) map[string]Swatch {
	sorted := append([]string(nil), names...)
	sort.Strings(sorted)

	n := len(sorted)
	k := stride(n)
	step := 360.0 / math.Max(1, float64(n))

	out := make(map[string]Swatch, n)
	for i, name := range sorted {
		slot := (i * k) % max(1, n)
		// Keyed on the slot, not the list index: the pairs at risk of looking
		// alike are the ones adjacent in HUE, so those are the ones that must
		// differ in lightness. Cycling by index instead leaves the closest hue
		// pair free to share a preset, which is what left two Passeriformes
		// families at dE 4.5.
		p := presets[slot%len(presets)]
		hue := math.Mod(seed+float64(slot)*step, 360)
		out[name] = swatchFor(hue, p.sat, p.start, p.dark)
	}
	return out
}

// swatchFor walks lightness until the color clears MinContrast against the text
// it will carry. It terminates: white reaches 21:1 against black text, and
// black reaches 21:1 against white text.
func swatchFor(hue, sat, start float64, dark bool) Swatch {
	if dark {
		for light := start; light > 0; light -= 0.01 {
			r, g, b := hslToRGB(hue, sat, light)
			if ContrastOnWhite(r, g, b) >= MinContrast {
				return Swatch{hex(r, g, b), white}
			}
		}
		return Swatch{black, white}
	}
	for light := start; light < 1.0; light += 0.01 {
		r, g, b := hslToRGB(hue, sat, light)
		if ContrastOnBlack(r, g, b) >= MinContrast {
			return Swatch{hex(r, g, b), black}
		}
	}
	return Swatch{white, black}
}

// Contrast returns the WCAG 2.2 contrast ratio between two relative luminances.
func Contrast(l1, l2 float64) float64 {
	if l1 < l2 {
		l1, l2 = l2, l1
	}
	return (l1 + 0.05) / (l2 + 0.05)
}

// ContrastOnBlack returns the contrast ratio of an sRGB color against black.
func ContrastOnBlack(r, g, b uint8) float64 { return Contrast(RelativeLuminance(r, g, b), 0) }

// ContrastOnWhite returns the contrast ratio of an sRGB color against white.
func ContrastOnWhite(r, g, b uint8) float64 { return Contrast(1, RelativeLuminance(r, g, b)) }

// ContrastWith returns the contrast ratio of a swatch's background against its
// own text color, which is the ratio tech.badgeContrast actually constrains.
func (s Swatch) ContrastWith() (float64, error) {
	r, g, b, err := ParseHex(s.Hex)
	if err != nil {
		return 0, err
	}
	tr, tg, tb, err := ParseHex(s.Text)
	if err != nil {
		return 0, err
	}
	return Contrast(RelativeLuminance(r, g, b), RelativeLuminance(tr, tg, tb)), nil
}

// RelativeLuminance implements the WCAG 2.x definition.
func RelativeLuminance(r, g, b uint8) float64 {
	lin := func(c uint8) float64 {
		s := float64(c) / 255
		if s <= 0.03928 {
			return s / 12.92
		}
		return math.Pow((s+0.055)/1.055, 2.4)
	}
	return 0.2126*lin(r) + 0.7152*lin(g) + 0.0722*lin(b)
}

// DeltaE76 returns the CIE76 color difference between two sRGB colors.
//
// The scale is conventional and citable, which is why it is used rather than an
// ad-hoc RGB distance: below about 2.3 a difference is only just noticeable,
// and 10 or more is clearly different at a glance.
func DeltaE76(r1, g1, b1, r2, g2, b2 uint8) float64 {
	l1, a1, bb1 := toLab(r1, g1, b1)
	l2, a2, bb2 := toLab(r2, g2, b2)
	return math.Sqrt((l1-l2)*(l1-l2) + (a1-a2)*(a1-a2) + (bb1-bb2)*(bb1-bb2))
}

func toLab(r8, g8, b8 uint8) (l, a, b float64) {
	lin := func(c uint8) float64 {
		s := float64(c) / 255
		if s <= 0.04045 {
			return s / 12.92
		}
		return math.Pow((s+0.055)/1.055, 2.4)
	}
	r, g, bl := lin(r8), lin(g8), lin(b8)
	x := (0.4124*r + 0.3576*g + 0.1805*bl) / 0.95047
	y := 0.2126*r + 0.7152*g + 0.0722*bl
	z := (0.0193*r + 0.1192*g + 0.9505*bl) / 1.08883
	f := func(t float64) float64 {
		if t > 0.008856 {
			return math.Cbrt(t)
		}
		return 7.787*t + 16.0/116
	}
	fx, fy, fz := f(x), f(y), f(z)
	return 116*fy - 16, 500 * (fx - fy), 200 * (fy - fz)
}

// ParseHex reads "#rrggbb".
func ParseHex(s string) (r, g, b uint8, err error) {
	_, err = fmt.Sscanf(s, "#%02x%02x%02x", &r, &g, &b)
	return
}

func hex(r, g, b uint8) string { return fmt.Sprintf("#%02x%02x%02x", r, g, b) }

func hslToRGB(h, s, l float64) (uint8, uint8, uint8) {
	c := (1 - math.Abs(2*l-1)) * s
	hp := math.Mod(h/60, 6)
	x := c * (1 - math.Abs(math.Mod(hp, 2)-1))
	var r, g, b float64
	switch int(hp) {
	case 0:
		r, g, b = c, x, 0
	case 1:
		r, g, b = x, c, 0
	case 2:
		r, g, b = 0, c, x
	case 3:
		r, g, b = 0, x, c
	case 4:
		r, g, b = x, 0, c
	default:
		r, g, b = c, 0, x
	}
	m := l - c/2
	to8 := func(v float64) uint8 { return uint8(math.Round(math.Max(0, math.Min(1, v+m)) * 255)) }
	return to8(r), to8(g), to8(b)
}

// AlternatesLightAndDark reports whether the preset cycle strictly alternates,
// which requires an even count. Exposed so a test can assert the property
// rather than re-deriving it.
func AlternatesLightAndDark() bool {
	if len(presets)%2 != 0 {
		return false
	}
	for i := range presets {
		if presets[i].dark == presets[(i+1)%len(presets)].dark {
			return false
		}
	}
	return true
}
