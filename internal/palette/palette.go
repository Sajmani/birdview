// Package palette generates a badge color per family.
//
// Verifies: P-005, T-010, wcag/1.4.6.
package palette

import (
	"fmt"
	"math"
	"sort"
)

// MinContrast is WCAG 2.2 success criterion 1.4.6 (Contrast Enhanced) for
// normal text. BirdView targets it rather than 1.4.3's 4.5:1.
//
// Verifies: T-010.
const MinContrast = 7.0

// goldenAngle spaces successive hues as far apart as possible: 360 degrees
// divided by the golden ratio. Stepping by it means consecutive families land
// on opposite sides of the wheel while the whole set still covers it evenly.
const goldenAngle = 137.50776405003785

// satCycle and lightCycle add a second and third axis of separation, so that
// two families which happen to land on a similar hue still differ in how
// washed-out or bright they look.
var (
	satCycle   = []float64{0.88, 0.55, 0.97}
	lightCycle = []float64{0.72, 0.80, 0.66}
)

// Generate assigns a color to every family.
//
// Hues step by the golden angle in sorted family order, rather than marching
// evenly around the wheel. Even spacing seems the obvious choice and is wrong
// for this use: with 91 ABA families it puts consecutive hues 4 degrees apart,
// and the family selector lists families in exactly that order -- so
// Accipitridae, Acrocephalidae and Aegithalidae came out as three
// indistinguishable reds sitting side by side. Golden-angle stepping keeps the
// full-wheel coverage while making neighbors in the list maximally unalike.
//
// Saturation and lightness cycle with period 3, so families that do land near
// each other in hue are still separated on another axis.
//
// Lightness is then raised per color until it clears MinContrast against black.
// That loop is not decoration: relative luminance is strongly hue-dependent --
// a saturated blue at the same HSL lightness as a yellow has roughly a third of
// its luminance -- so one fixed lightness that works for yellow fails for blue.
//
// Verifies: P-005, T-010.
func Generate(families []string) map[string]string {
	sorted := append([]string(nil), families...)
	sort.Strings(sorted)

	out := make(map[string]string, len(sorted))
	for i, fam := range sorted {
		hue := math.Mod(float64(i)*goldenAngle, 360)
		out[fam] = colorAtLeast(hue, satCycle[i%len(satCycle)],
			lightCycle[i%len(lightCycle)], MinContrast)
	}
	return out
}

// colorAtLeast returns the hex color at the given hue and saturation whose
// contrast against black is at least want, raising lightness from start in
// small steps until it is. It terminates: at lightness 1.0 the color is white,
// whose contrast against black is 21:1.
func colorAtLeast(hue, sat, start, want float64) string {
	for light := start; light < 1.0; light += 0.01 {
		r, g, b := hslToRGB(hue, sat, light)
		if ContrastOnBlack(r, g, b) >= want {
			return fmt.Sprintf("#%02x%02x%02x", r, g, b)
		}
	}
	return "#ffffff"
}

// ContrastOnBlack returns the WCAG 2.2 contrast ratio of an sRGB color against
// pure black, which is (L+0.05)/0.05 where L is relative luminance.
//
// Verifies: T-010, wcag/1.4.6.
func ContrastOnBlack(r, g, b uint8) float64 {
	return (RelativeLuminance(r, g, b) + 0.05) / 0.05
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

// DeltaE76 returns the CIE76 color difference between two sRGB colors.
//
// The scale is conventional and citable, which is why it is used here rather
// than an ad-hoc RGB distance: below about 2.3 a difference is only just
// noticeable, and 10 or more is clearly different at a glance.
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
	// sRGB to XYZ, normalized to the D65 white point.
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
