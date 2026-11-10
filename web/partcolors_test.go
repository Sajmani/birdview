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

package web

import (
	"math"
	"regexp"
	"testing"

	"birdview/internal/palette"
)

// part is one part of the diagram: the rule that strokes it, and the rule
// that colors its name on the tile, if it has one (product.oneDiagramVocabulary,
// product.labeledDetailDiagram). Apart marks a part drawn away from the figure, with its name
// beside it, which need hold apart from the others only under normal vision
// (decisions.massCubeSkyBlue).
type part struct {
	name, stroke, key string
	apart             bool
}

var parts = []part{
	{"bill", ".bill", ".k-bill", false},
	{"tail", ".tail", ".k-tail", false},
	{"crossbar", ".wing", ".k-wing", false},
	{"chord", ".chord", ".k-chord", false},
	{"mass cube", ".masscube polygon, .masscube rect", ".k-mass", true},
	{"head and neck", ".neck", "", false},
	{"body", ".body", "", false},
}

var reStroke = regexp.MustCompile(`(?:^|[;\s])stroke:\s*(#[0-9a-fA-F]{6})`)
var reTextColor = regexp.MustCompile(`(?:^|[;\s])color:\s*(#[0-9a-fA-F]{6})`)

// TestPartColorsHoldApart asserts the diagram's palette (product.oneDiagramVocabulary, decisions.partColorsByMeasurement): every
// colored part's stroke clears 7:1 on the page and the panel, so its name can be set in it;
// each name on the tile is set in exactly its part's color; and every two parts, grays
// included, differ by a ΔE00 of at least 10 under normal vision and under simulated
// protanopia, deuteranopia, and tritanopia (Machado, Oliveira, and Fernandes 2009, severity 1),
// except that the mass cube, drawn apart from the figure with its name beside it, is held apart
// only under normal vision (decisions.massCubeSkyBlue).
//
// Verifies: product.oneDiagramVocabulary, product.labeledDetailDiagram. Criterion: app.partColorsHoldApart.
func TestPartColorsHoldApart(t *testing.T) {
	type col struct {
		name  string
		rgb   [3]uint8
		apart bool
	}
	var cols []col
	page, panel := mustVar(t, "--bg"), mustVar(t, "--panel")
	for _, p := range parts {
		m := reStroke.FindStringSubmatch(rule(t, p.stroke))
		if m == nil {
			t.Errorf("%s: %s sets no stroke color", p.name, p.stroke)
			continue
		}
		r, g, b := hexRGB(t, m[1])
		cols = append(cols, col{p.name, [3]uint8{r, g, b}, p.apart})
		if p.key == "" {
			continue
		}
		l := palette.RelativeLuminance(r, g, b)
		for _, s := range []struct {
			name string
			rgb  [3]uint8
		}{{"page", page}, {"panel", panel}} {
			if c := palette.Contrast(l, palette.RelativeLuminance(s.rgb[0], s.rgb[1], s.rgb[2])); c < 7 {
				t.Errorf("%s %s is %.2f:1 on the %s; its name could not be set in it", p.name, m[1], c, s.name)
			}
		}
		if k := reTextColor.FindStringSubmatch(rule(t, p.key)); k == nil || k[1] != m[1] {
			t.Errorf("%s: its name is colored %v, its stroke %s; they should be the one color", p.name, k, m[1])
		}
	}
	for _, sim := range []string{"normal", "protanopia", "deuteranopia", "tritanopia"} {
		for i := range cols {
			for j := i + 1; j < len(cols); j++ {
				if sim != "normal" && (cols[i].apart || cols[j].apart) {
					continue
				}
				if d := deltaE2000(simulate(cols[i].rgb, sim), simulate(cols[j].rgb, sim)); d < 10 {
					t.Errorf("%s: %s and %s are ΔE00 %.1f apart; at least 10 keeps them distinct", sim, cols[i].name, cols[j].name, d)
				}
			}
		}
	}
}

func mustVar(t *testing.T, name string) [3]uint8 {
	t.Helper()
	m := regexp.MustCompile(regexp.QuoteMeta(name) + `:\s*(#[0-9a-fA-F]{6})`).FindStringSubmatch(string(mustReadStyles(t)))
	if m == nil {
		t.Fatalf("no %s in the stylesheet", name)
	}
	r, g, b := hexRGB(t, m[1])
	return [3]uint8{r, g, b}
}

func toLinear(c uint8) float64 {
	v := float64(c) / 255
	if v <= 0.04045 {
		return v / 12.92
	}
	return math.Pow((v+0.055)/1.055, 2.4)
}

func fromLinear(v float64) float64 {
	v = math.Max(0, math.Min(1, v))
	if v <= 0.0031308 {
		return 12.92 * v
	}
	return 1.055*math.Pow(v, 1/2.4) - 0.055
}

// cvd are Machado et al.'s 2009 matrices at full severity, on linear RGB.
var cvd = map[string][3][3]float64{
	"protanopia":   {{0.152286, 1.052583, -0.204868}, {0.114503, 0.786281, 0.099216}, {-0.003882, -0.048116, 1.051998}},
	"deuteranopia": {{0.367322, 0.860646, -0.227968}, {0.280085, 0.672501, 0.047413}, {-0.011820, 0.042940, 0.968881}},
	"tritanopia":   {{1.255528, -0.076749, -0.178779}, {-0.078411, 0.930809, 0.147602}, {0.004733, 0.691367, 0.303900}},
}

// simulate returns a color as a viewer with the named vision sees it, as sRGB in [0,1].
func simulate(c [3]uint8, kind string) [3]float64 {
	l := [3]float64{toLinear(c[0]), toLinear(c[1]), toLinear(c[2])}
	m, ok := cvd[kind]
	if !ok {
		return [3]float64{float64(c[0]) / 255, float64(c[1]) / 255, float64(c[2]) / 255}
	}
	var out [3]float64
	for i := 0; i < 3; i++ {
		out[i] = fromLinear(m[i][0]*l[0] + m[i][1]*l[1] + m[i][2]*l[2])
	}
	return out
}

func lab(c [3]float64) (float64, float64, float64) {
	lin := func(v float64) float64 {
		if v <= 0.04045 {
			return v / 12.92
		}
		return math.Pow((v+0.055)/1.055, 2.4)
	}
	r, g, b := lin(c[0]), lin(c[1]), lin(c[2])
	x := (0.4124*r + 0.3576*g + 0.1805*b) / 0.95047
	y := 0.2126*r + 0.7152*g + 0.0722*b
	z := (0.0193*r + 0.1192*g + 0.9505*b) / 1.08883
	f := func(t float64) float64 {
		if t > 0.008856 {
			return math.Cbrt(t)
		}
		return 7.787*t + 16.0/116
	}
	return 116*f(y) - 16, 500 * (f(x) - f(y)), 200 * (f(y) - f(z))
}

// deltaE2000 is the CIEDE2000 color difference between two sRGB colors.
func deltaE2000(c1, c2 [3]float64) float64 {
	rad, deg := math.Pi/180, 180/math.Pi
	L1, a1, b1 := lab(c1)
	L2, a2, b2 := lab(c2)
	cb := (math.Hypot(a1, b1) + math.Hypot(a2, b2)) / 2
	g := 0.5 * (1 - math.Sqrt(math.Pow(cb, 7)/(math.Pow(cb, 7)+math.Pow(25, 7))))
	a1p, a2p := (1+g)*a1, (1+g)*a2
	c1p, c2p := math.Hypot(a1p, b1), math.Hypot(a2p, b2)
	h := func(b, a float64) float64 { return math.Mod(math.Atan2(b, a)*deg+360, 360) }
	h1p, h2p := h(b1, a1p), h(b2, a2p)
	dL, dC := L2-L1, c2p-c1p
	dh := h2p - h1p
	switch {
	case c1p*c2p == 0:
		dh = 0
	case dh > 180:
		dh -= 360
	case dh < -180:
		dh += 360
	}
	dH := 2 * math.Sqrt(c1p*c2p) * math.Sin(dh/2*rad)
	lbp, cbp := (L1+L2)/2, (c1p+c2p)/2
	hbp := h1p + h2p
	if c1p*c2p != 0 {
		if math.Abs(h1p-h2p) <= 180 {
			hbp = (h1p + h2p) / 2
		} else if h1p+h2p < 360 {
			hbp = (h1p + h2p + 360) / 2
		} else {
			hbp = (h1p + h2p - 360) / 2
		}
	}
	tt := 1 - 0.17*math.Cos((hbp-30)*rad) + 0.24*math.Cos(2*hbp*rad) + 0.32*math.Cos((3*hbp+6)*rad) - 0.20*math.Cos((4*hbp-63)*rad)
	dth := 30 * math.Exp(-math.Pow((hbp-275)/25, 2))
	rc := 2 * math.Sqrt(math.Pow(cbp, 7)/(math.Pow(cbp, 7)+math.Pow(25, 7)))
	sl := 1 + 0.015*math.Pow(lbp-50, 2)/math.Sqrt(20+math.Pow(lbp-50, 2))
	sc, sh := 1+0.045*cbp, 1+0.015*cbp*tt
	rt := -math.Sin(2*dth*rad) * rc
	return math.Sqrt(math.Pow(dL/sl, 2) + math.Pow(dC/sc, 2) + math.Pow(dH/sh, 2) + rt*(dC/sc)*(dH/sh))
}

// TestDeltaE2000KnownValues pins the formula on two pairs whose difference is known by
// construction: a color against itself, and black against white, which differ only in
// lightness, by 100.
func TestDeltaE2000KnownValues(t *testing.T) {
	if d := deltaE2000([3]float64{0.5, 0.2, 0.7}, [3]float64{0.5, 0.2, 0.7}); d != 0 {
		t.Errorf("a color differs from itself by %.3f", d)
	}
	if d := deltaE2000([3]float64{0, 0, 0}, [3]float64{1, 1, 1}); math.Abs(d-100) > 0.01 {
		t.Errorf("black and white differ by %.3f, want 100", d)
	}
}
