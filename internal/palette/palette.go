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

// Generate assigns a color to every family, evenly spaced around the hue wheel
// in sorted family order so the result is deterministic for a given set of
// families and adjacent families are maximally separated in hue.
//
// Lightness starts high enough that a pastel clears MinContrast against black,
// and is raised per color until it actually does. The loop is not decoration:
// relative luminance is strongly hue-dependent -- a saturated blue at the same
// HSL lightness as a yellow has roughly a third of its luminance -- so a single
// fixed lightness that works for yellow fails for blue.
//
// Verifies: P-005, T-010.
func Generate(families []string) map[string]string {
	sorted := append([]string(nil), families...)
	sort.Strings(sorted)

	out := make(map[string]string, len(sorted))
	for i, fam := range sorted {
		hue := 360.0 * float64(i) / math.Max(1, float64(len(sorted)))
		out[fam] = colorAtLeast(hue, 0.82, MinContrast)
	}
	return out
}

// colorAtLeast returns the hex color at the given hue and saturation whose
// contrast against black is at least want, raising lightness in small steps
// until it is. It terminates: at lightness 1.0 the color is white, whose
// contrast against black is 21:1.
func colorAtLeast(hue, sat, want float64) string {
	for light := 0.72; light < 1.0; light += 0.01 {
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
