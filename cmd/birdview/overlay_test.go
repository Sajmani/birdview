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

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
	"math"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os/exec"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	"birdview/internal/dataset"
)

func overlayTestServer(t *testing.T) *server {
	t.Helper()
	ds := testDataset()
	base := ds.Species[1] // American Robin template for extra pins
	for _, extra := range []struct {
		id, name, sci string
		len, wing     float64
	}{
		{"sparrow1", "Song Sparrow", "Melospiza melodia", 150, 210},
		{"warbler1", "Yellow Warbler", "Setophaga petechia", 120, 190},
		{"hawk1", "Red-tailed Hawk", "Buteo jamaicensis", 500, 1200},
		{"duck1", "Mallard", "Anas platyrhynchos", 580, 900},
	} {
		sp := base
		sp.ID = extra.id
		sp.SpeciesCode = extra.id
		sp.ScientificName = extra.sci
		sp.Names = map[dataset.Locale]string{
			dataset.LocaleNA:    extra.name,
			dataset.LocaleWorld: extra.name,
		}
		sp.TotalLength.Value = extra.len
		sp.Wingspan.Value = extra.wing
		sp.Segments.Body.Value = extra.len - (sp.Segments.Bill.Value + sp.Segments.Head.Value + sp.Segments.Neck.Value + sp.Segments.Tail.Value)
		ds.Species = append(ds.Species, sp)
	}
	s, err := newServer(ds, nil)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

// TestOverlayShowRowAndPinBounds checks Overlay visibility and 2-8 pin bounds
// in the Show row and on direct ?show=overlay requests.
//
// Verifies: app.overlayModeComparesPinnedBirds.
func TestOverlayShowRowAndPinBounds(t *testing.T) {
	s := overlayTestServer(t)

	// 0 and 1 pins: Overlay absent from Show row; direct show=overlay asks for 2+.
	for _, q := range []string{"/", "/?pin=amecro"} {
		_, body := get(t, s, q)
		row := mustMatch(t, body, `(?s)<div class="showrow">(.*?)\n</div>`, "Show row")
		if strings.Contains(row, ">Overlay") {
			t.Errorf("%s Show row shows Overlay with fewer than 2 pins:\n%s", q, row)
		}
		_, ov := get(t, s, q+map[string]string{"/": "?show=overlay", "/?pin=amecro": "&show=overlay"}[q])
		if !strings.Contains(ov, "Pin at least 2 birds to overlay them") {
			t.Errorf("%s&show=overlay missing too-few notice", q)
		}
	}

	// 2 pins and 8 pins: Overlay link enabled in Show row with the pinned count.
	pins8 := []string{"annhum", "amerob", "amecro", "blasti", "greblu", "thimur", "sparrow1", "warbler1"}
	pinQuery := func(ids []string) string {
		v := url.Values{}
		for _, id := range ids {
			v.Add("pin", id)
		}
		return v.Encode()
	}
	for _, ids := range [][]string{{"amecro", "annhum"}, pins8} {
		_, body := get(t, s, "/?"+pinQuery(ids))
		row := mustMatch(t, body, `(?s)<div class="showrow">(.*?)\n</div>`, "Show row")
		if !strings.Contains(row, `href="/?`+pinQuery(ids)+`&amp;show=overlay"`) &&
			!strings.Contains(row, `show=overlay`) {
			t.Errorf("%d pins: Show row missing enabled Overlay link:\n%s", len(ids), row)
		}
		wantLabel := fmt.Sprintf(">Overlay (%d)</a>", len(ids))
		if !strings.Contains(row, wantLabel) {
			t.Errorf("%d pins: Show row missing %q:\n%s", len(ids), wantLabel, row)
		}
		if strings.Contains(row, `aria-disabled="true"`) {
			t.Errorf("%d pins: Overlay unexpectedly disabled:\n%s", len(ids), row)
		}
	}

	// 9 pins: Overlay disabled in Show row (still showing count), and direct show=overlay says at most 8.
	pins9 := append(append([]string(nil), pins8...), "hawk1")
	_, body9 := get(t, s, "/?"+pinQuery(pins9))
	row9 := mustMatch(t, body9, `(?s)<div class="showrow">(.*?)\n</div>`, "Show row")
	if !strings.Contains(row9, `<span class="btn disabled" aria-disabled="true"`) ||
		!strings.Contains(row9, `>Overlay (9)</span>`) {
		t.Errorf("9 pins: Show row did not disable Overlay (9):\n%s", row9)
	}
	_, ov9 := get(t, s, "/?show=overlay&"+pinQuery(pins9))
	if !strings.Contains(ov9, "Overlay compares up to 8 pinned birds (9 are pinned)") {
		t.Errorf("9 pins direct show=overlay missing cap notice")
	}

	// Switching away from show=overlay to Pinned drops overlay-only params.
	_, ov := get(t, s, "/?show=overlay&align=tail&spread=1&shift=1&drag=amecro:30,-20&pin=amecro&pin=annhum")
	pinBtn := mustMatch(t, ov, `<a class="btn" href="([^"]+)"[^>]*>Pinned \(2\)</a>`, "Pinned button")
	href := html.UnescapeString(pinBtn)
	for _, drop := range []string{"align=", "spread=", "shift=", "drag="} {
		if strings.Contains(href, drop) {
			t.Errorf("Pinned button %q retained overlay param %q", href, drop)
		}
	}
}

// TestOverlayGeometryAlignmentAndURL checks shared scale, largest-behind-smallest
// layering, alignment presets, horizontal spread, 2D drag= URL parameters, and
// preserved provenance cues.
//
// Verifies: app.overlayModeComparesPinnedBirds.
func TestOverlayGeometryAlignmentAndURL(t *testing.T) {
	s := overlayTestServer(t)

	// Three birds with distinct silhouettes and provenance: hummingbird (smallest),
	// crow (wing modeled, solid tail/bill, dashed body), heron (largest, folded neck).
	parseAttrs := func(body string) (layers []struct {
		id, cls                  string
		baseX, baseY, x, y, span float64
		dragX, dragY             int
	}, guideY float64) {
		gm := mustMatch(t, body, `<line class="ov-guide"[^>]*y1="([^"]+)"`, "overlay guide")
		guideY, _ = strconv.ParseFloat(gm, 64)
		reBird := regexp.MustCompile(`<g class="ov-bird (ov-c\d)" data-id="([^"]+)" data-base-x="([^"]+)" data-base-y="([^"]+)" data-drag-x="([^"]+)" data-drag-y="([^"]+)"[^>]*data-span="([^"]+)"[^>]*transform="translate\(([^,]+),([^)]+)\)"`)
		for _, m := range reBird.FindAllStringSubmatch(body, -1) {
			bx, _ := strconv.ParseFloat(m[3], 64)
			by, _ := strconv.ParseFloat(m[4], 64)
			dxMM, _ := strconv.Atoi(m[5])
			dyMM, _ := strconv.Atoi(m[6])
			span, _ := strconv.ParseFloat(m[7], 64)
			x, _ := strconv.ParseFloat(m[8], 64)
			y, _ := strconv.ParseFloat(m[9], 64)
			layers = append(layers, struct {
				id, cls                  string
				baseX, baseY, x, y, span float64
				dragX, dragY             int
			}{m[2], m[1], bx, by, x, y, span, dxMM, dyMM})
		}
		return layers, guideY
	}

	// Default (wings alignment): largest painted first (greblu -> amecro -> annhum),
	// no inline style/color attrs, crow wings dashed, heron folded neck path present.
	_, body := get(t, s, "/?show=overlay&pin=annhum&pin=amecro&pin=greblu")
	layers, guideY := parseAttrs(body)
	if len(layers) != 3 || layers[0].id != "greblu" || layers[1].id != "amecro" || layers[2].id != "annhum" {
		t.Fatalf("overlay layer paint order = %v, want [greblu amecro annhum]", layers)
	}
	if !strings.Contains(body, `class="wing est"`) || !strings.Contains(body, `class="neck" d="M `) {
		t.Errorf("overlay lost stroke provenance or folded-neck S-curve")
	}
	if strings.Contains(body, `class="ov-bird ov-c0" style=`) {
		t.Errorf("overlay bird carries inline style")
	}
	// Wing alignment puts every bird's shoulder (top + geom.YShoulder) on guideY.
	byID := map[string]struct{ baseY, x, y float64 }{}
	for _, l := range layers {
		byID[l.id] = struct{ baseY, x, y float64 }{l.baseY, l.x, l.y}
	}
	// Non-spread: every bird shares the exact same horizontal x.
	if byID["greblu"].x != byID["amecro"].x || byID["amecro"].x != byID["annhum"].x {
		t.Errorf("without spread=1, x differed: %v", byID)
	}
	if guideY <= 0 {
		t.Errorf("guideY = %v, want > 0", guideY)
	}

	// Bill alignment puts every bird's bill tip (YBillTip == 0 -> y) on guideY.
	_, billBody := get(t, s, "/?show=overlay&align=bill&pin=annhum&pin=amecro&pin=greblu")
	billLayers, billGuide := parseAttrs(billBody)
	for _, l := range billLayers {
		if l.y != billGuide {
			t.Errorf("align=bill: bird %s y=%.2f, guide=%.2f", l.id, l.y, billGuide)
		}
	}

	// Spread=1 separates x by active Order by order symmetrically while keeping y untouched;
	// spread buttons use matching monochrome SVG arrows, and Reset drags + prompt sit below canvas without ▲/▼ buttons.
	_, spBody := get(t, s, "/?show=overlay&by=length&align=bill&spread=1&pin=annhum&pin=amecro&pin=greblu")
	spBtn := mustMatch(t, spBody, `(?s)<a class="btn ov-spread[^"]*"[^>]*>(.*?)</a>`, "Spread horizontal button")
	shBtn := mustMatch(t, spBody, `(?s)<a class="btn ov-shift[^"]*"[^>]*>(.*?)</a>`, "Spread vertical button")
	if !strings.Contains(spBtn, `<svg class="ov-arrow"`) || strings.Contains(spBtn, "↔") ||
		!strings.Contains(shBtn, `<svg class="ov-arrow"`) || strings.Contains(shBtn, "↕") {
		t.Errorf("spread buttons do not use uniform inline SVG arrows: spread=%q shift=%q", spBtn, shBtn)
	}
	hint := mustMatch(t, spBody, `(?s)<p class="ov-hint">(.*?)</p>`, "overlay provenance hint")
	if !strings.Contains(hint, "Solid lines are measured; dashed lines and hollow heads are estimated.") ||
		strings.Contains(hint, "grab one") {
		t.Errorf("unexpected ov-hint copy: %q", hint)
	}
	legHead := mustMatch(t, spBody, `(?s)<div class="ov-leghead">(.*?)</div>`, "overlay legend header")
	if !strings.Contains(legHead, "Select a bird below to highlight and drag") || !strings.Contains(legHead, "Reset drags") {
		t.Errorf("ov-leghead missing prompt or Reset drags button: %q", legHead)
	}
	if strings.Contains(spBody, "ov-up") || strings.Contains(spBody, "ov-dn") {
		t.Errorf("overlay legend still contains up/down arrow buttons")
	}
	spLayers, _ := parseAttrs(spBody)
	spByID := map[string]float64{}
	for _, l := range spLayers {
		spByID[l.id] = l.x
	}
	if !(spByID["annhum"] < spByID["amecro"] && spByID["amecro"] < spByID["greblu"]) {
		t.Errorf("spread=1 x not ordered by length: %v", spByID)
	}

	// Shift=1 staggers y vertically by exact overlaySpreadStep in active Order by order while keeping x untouched.
	_, shfBody := get(t, s, "/?show=overlay&by=length&align=bill&shift=1&pin=annhum&pin=amecro&pin=greblu")
	shfLayers, _ := parseAttrs(shfBody)
	shfByID := map[string]struct{ x, y float64 }{}
	for _, l := range shfLayers {
		shfByID[l.id] = struct{ x, y float64 }{l.x, l.y}
	}
	if shfByID["annhum"].x != shfByID["amecro"].x || shfByID["amecro"].x != shfByID["greblu"].x {
		t.Errorf("shift=1 changed horizontal x: %v", shfByID)
	}
	dy01 := shfByID["amecro"].y - shfByID["annhum"].y
	dy12 := shfByID["greblu"].y - shfByID["amecro"].y
	if math.Abs(dy01-overlaySpreadStep) > 0.05 || math.Abs(dy12-overlaySpreadStep) > 0.05 {
		t.Errorf("shift=1 vertical steps = (%.2f, %.2f), want exact %.2f", dy01, dy12, overlaySpreadStep)
	}

	// Spread=1&shift=1 applies both horizontal and vertical offsets together.
	_, bothBody := get(t, s, "/?show=overlay&by=length&align=bill&spread=1&shift=1&pin=annhum&pin=amecro&pin=greblu")
	bothLayers, _ := parseAttrs(bothBody)
	bothByID := map[string]struct{ x, y float64 }{}
	for _, l := range bothLayers {
		bothByID[l.id] = struct{ x, y float64 }{l.x, l.y}
	}
	if !(bothByID["annhum"].x < bothByID["amecro"].x && bothByID["amecro"].x < bothByID["greblu"].x) ||
		math.Abs((bothByID["amecro"].y-bothByID["annhum"].y)-overlaySpreadStep) > 0.05 {
		t.Errorf("spread=1&shift=1 did not apply both offsets: %v", bothByID)
	}

	// Overlay legend, horizontal bottom cube row, spread=1, and shift=1 follow the Order by control
	// (and rev=1) just like Pinned tiles, while SVG paint order stays largest behind smallest
	// and each bird keeps its URL pin-order color class and 1-8 badge across every sort.
	reLegend := regexp.MustCompile(`<div class="ov-row"[^>]*data-id="([^"]+)"[\s\S]*?<span class="ov-num (ov-c\d)"[^>]*>([^<]+)</span>`)
	reTag := regexp.MustCompile(`<span class="ov-tag ov-wtag (ov-c\d)" data-id="([^"]+)"[^>]*>([^<]+)</span>`)
	reCubeTag := regexp.MustCompile(`<span class="ov-tag ov-ctag (ov-c\d)" data-id="([^"]+)"[^>]*>([^<]+)</span>`)
	reCube := regexp.MustCompile(`<g class="ov-cube (ov-c\d)" data-id="([^"]+)">[\s\S]*?<rect class="mcface" x="([^"]+)" y="([^"]+)"\s+width="([^"]+)" height="([^"]+)"/>`)
	wantPinBadge := map[string]struct{ cls, label string }{
		"thimur": {"ov-c0", "1"},
		"amecro": {"ov-c1", "2"},
		"blasti": {"ov-c2", "3"},
	}
	for _, ord := range []string{"", "&by=length", "&by=length&rev=1", "&by=mass", "&by=wingspan", "&by=bill"} {
		const pins = "&pin=thimur&pin=amecro&pin=blasti"
		pinnedPage := s.nextPage(httptest.NewRequest("GET", "/?show=pinned"+pins+ord, nil))
		var wantOrder []string
		for _, tl := range pinnedPage.GroupTiles {
			wantOrder = append(wantOrder, tl.S.ID)
		}
		_, ordBody := get(t, s, "/?show=overlay&align=bill&spread=1&shift=1"+pins+ord)
		var gotLegend []string
		for _, m := range reLegend.FindAllStringSubmatch(ordBody, -1) {
			gotLegend = append(gotLegend, m[1])
			if w := wantPinBadge[m[1]]; m[2] != w.cls || m[3] != w.label {
				t.Errorf("order %q legend badge %s = (%s,%q), want (%s,%q)", ord, m[1], m[2], m[3], w.cls, w.label)
			}
		}
		if !slices.Equal(gotLegend, wantOrder) {
			t.Errorf("order %q: legend = %v, want %v (matching Pinned)", ord, gotLegend, wantOrder)
		}
		tags := map[string]struct{ cls, label string }{}
		for _, m := range reTag.FindAllStringSubmatch(ordBody, -1) {
			tags[m[2]] = struct{ cls, label string }{m[1], m[3]}
		}
		if len(tags) != 3 {
			t.Errorf("order %q: found %d wingtip .ov-wtag badges, want 3", ord, len(tags))
		}
		for id, w := range wantPinBadge {
			if g := tags[id]; g != w {
				t.Errorf("order %q wingtip .ov-wtag %s = %+v, want %+v", ord, id, g, w)
			}
		}
		var cubeTagOrder []string
		for _, m := range reCubeTag.FindAllStringSubmatch(ordBody, -1) {
			cubeTagOrder = append(cubeTagOrder, m[2])
			if w := wantPinBadge[m[2]]; m[1] != w.cls || m[3] != w.label {
				t.Errorf("order %q cube badge %s = (%s,%q), want (%s,%q)", ord, m[2], m[1], m[3], w.cls, w.label)
			}
		}
		if !slices.Equal(cubeTagOrder, wantOrder) {
			t.Errorf("order %q: cube badge order = %v, want %v", ord, cubeTagOrder, wantOrder)
		}
		ordLayers, _ := parseAttrs(ordBody)
		var paintIDs []string
		posByID := map[string]struct{ x, y float64 }{}
		maxBirdTailY := 0.0
		for _, l := range ordLayers {
			paintIDs = append(paintIDs, l.id)
			posByID[l.id] = struct{ x, y float64 }{l.x, l.y}
			if tailY := l.y + l.span; tailY > maxBirdTailY {
				maxBirdTailY = tailY
			}
			if w := wantPinBadge[l.id]; l.cls != w.cls {
				t.Errorf("order %q SVG layer %s class = %s, want %s (locked to URL pin order)", ord, l.id, l.cls, w.cls)
			}
		}
		wantPaint := []string{"amecro", "thimur", "blasti"}
		if !slices.Equal(paintIDs, wantPaint) {
			t.Errorf("order %q: paint order = %v, want %v (largest behind smallest)", ord, paintIDs, wantPaint)
		}
		for i := 1; i < len(wantOrder); i++ {
			a, b := posByID[wantOrder[i-1]], posByID[wantOrder[i]]
			if !(a.x < b.x) || math.Abs((b.y-a.y)-overlaySpreadStep) > 0.05 {
				t.Errorf("order %q: spread order broke between %s (%v) and %s (%v)", ord, wantOrder[i-1], a, wantOrder[i], b)
			}
		}
		type cubeBox struct {
			id, cls    string
			x, y, side float64
		}
		var cubes []cubeBox
		for _, m := range reCube.FindAllStringSubmatch(ordBody, -1) {
			x, _ := strconv.ParseFloat(m[3], 64)
			y, _ := strconv.ParseFloat(m[4], 64)
			side, _ := strconv.ParseFloat(m[5], 64)
			cubes = append(cubes, cubeBox{m[2], m[1], x, y, side})
		}
		if len(cubes) != 3 {
			t.Errorf("order %q: got %d cubes, want 3", ord, len(cubes))
		} else {
			var cubeOrder []string
			floorY := cubes[0].y + cubes[0].side
			for i, c := range cubes {
				cubeOrder = append(cubeOrder, c.id)
				if w := wantPinBadge[c.id]; c.cls != w.cls {
					t.Errorf("order %q cube %s class = %s, want %s", ord, c.id, c.cls, w.cls)
				}
				if math.Abs((c.y+c.side)-floorY) > 0.2 {
					t.Errorf("order %q cube %s floor = %.2f, want shared %.2f", ord, c.id, c.y+c.side, floorY)
				}
				if cubeTop := c.y - c.side*massCubeDepth; cubeTop <= maxBirdTailY {
					t.Errorf("order %q cube %s top = %.2f overlaps bird tails maxY = %.2f", ord, c.id, cubeTop, maxBirdTailY)
				}
				if i > 0 && cubes[i-1].x+cubes[i-1].side*(1+massCubeDepth) >= c.x {
					t.Errorf("order %q cubes %s (x=%.2f,w=%.2f) and %s (x=%.2f) overlap horizontally",
						ord, cubes[i-1].id, cubes[i-1].x, cubes[i-1].side*(1+massCubeDepth), c.id, c.x)
				}
			}
			if !slices.Equal(cubeOrder, wantOrder) {
				t.Errorf("order %q: cube row order = %v, want %v", ord, cubeOrder, wantOrder)
			}
		}
	}

	// Millimeter drag=amecro:80,-100 moves amecro by (+80, -100) * PxPerMm without moving others;
	// Align preserves dx (clearing dy to 0), Spread keeps both axes, Reset clears all drags,
	// and per-bird × reset appears only for amecro.
	_, drBody := get(t, s, "/?show=overlay&pin=annhum&pin=amecro&pin=greblu&drag=amecro:80,-100")
	pxMatch := mustMatch(t, drBody, `data-px-per-mm="([^"]+)"`, "pxPerMm")
	pxPerMm, _ := strconv.ParseFloat(pxMatch, 64)
	drLayers, _ := parseAttrs(drBody)
	for _, l := range drLayers {
		if l.id == "amecro" {
			wantDX, wantDY := 80*pxPerMm, -100*pxPerMm
			gotDX, gotDY := l.x-l.baseX, l.y-l.baseY
			if math.Abs(gotDX-wantDX) > 0.2 || math.Abs(gotDY-wantDY) > 0.2 {
				t.Errorf("drag=amecro:80,-100 moved by (%.2f, %.2f) px, want (%.2f, %.2f) px (pxPerMm=%.4f)",
					gotDX, gotDY, wantDX, wantDY, pxPerMm)
			}
		} else if l.x != l.baseX || l.y != l.baseY {
			t.Errorf("drag=amecro:80,-100 moved undragged bird %s: (x,y)=(%.2f,%.2f) base=(%.2f,%.2f)",
				l.id, l.x, l.y, l.baseX, l.baseY)
		}
	}
	reRowZero := regexp.MustCompile(`<div class="ov-row"[^>]*data-id="([^"]+)"[\s\S]*?<a class="btn ov-zero"([^>]*)>`)
	zeroHidden := map[string]bool{}
	for _, m := range reRowZero.FindAllStringSubmatch(drBody, -1) {
		zeroHidden[m[1]] = strings.Contains(m[2], " hidden")
	}
	if zeroHidden["amecro"] || !zeroHidden["annhum"] || !zeroHidden["greblu"] {
		t.Errorf("per-bird reset hidden states = %v, want only amecro visible", zeroHidden)
	}
	billHref := html.UnescapeString(mustMatch(t, drBody, `<a class="btn" href="([^"]*align=bill[^"]*)"`, "Bill align link"))
	if bu, err := url.Parse(billHref); err != nil || !slices.Equal(bu.Query()["drag"], []string{"amecro:80,0"}) {
		t.Errorf("Align link drag query = %v (%q), want [amecro:80,0]", bu.Query()["drag"], billHref)
	}
	spreadHref := html.UnescapeString(mustMatch(t, drBody, `<a class="btn ov-spread[^"]*" href="([^"]+)"`, "Spread link"))
	if su, err := url.Parse(spreadHref); err != nil || !slices.Equal(su.Query()["drag"], []string{"amecro:80,-100"}) {
		t.Errorf("Spread link drag query = %v (%q), want [amecro:80,-100]", su.Query()["drag"], spreadHref)
	}
	resetTag := mustMatch(t, drBody, `<a class="btn reset ov-reset"([^>]*)>`, "Reset drags link")
	if strings.Contains(resetTag, " hidden") {
		t.Errorf("Reset drags stayed hidden when drag= was active: %q", resetTag)
	}
	resetHref := html.UnescapeString(mustMatch(t, resetTag, `href="([^"]+)"`, "Reset drags href"))
	if strings.Contains(resetHref, "drag=") {
		t.Errorf("Reset drags link retained drag=: %q", resetHref)
	}
}

// TestOverlayInTheBrowser verifies select-first interaction (dragging on canvas
// before selecting does nothing; legend click brings bird to front and dims
// others), 2D pointer drag and arrow-key nudges updating transform and drag= URL
// via history.replaceState (with Align preserving dx and clearing dy), Reset
// drags, and 360px phone layout without horizontal overflow.
//
// Verifies: app.overlayModeComparesPinnedBirds.
func TestOverlayInTheBrowser(t *testing.T) {
	chrome := findChrome()
	if chrome == "" {
		t.Skip("no Chrome or Chromium found (set BIRDVIEW_CHROME); Overlay browser interactions not tested")
	}

	s := overlayTestServer(t)
	probeJS := `
window.addEventListener("load", function () {
  var params = new URLSearchParams(location.hash.slice(1));
  if (!params.has("ovprobe")) return;
  setTimeout(function () {
    var out = {
      width: document.documentElement.clientWidth,
      scrollW: document.documentElement.scrollWidth,
      showTops: [],
      barTops: []
    };
    document.querySelectorAll(".showrow > *").forEach(function (e) {
      var r = e.getBoundingClientRect();
      out.showTops.push(Math.round(r.top));
      if (r.right > out.width + 1) out.overflowRight = Math.round(r.right);
    });
    document.querySelectorAll(".ov-bar .seg > .btn, .ov-bar .ov-spread, .ov-bar .ov-shift").forEach(function (e) {
      out.barTops.push(Math.round(e.getBoundingClientRect().top));
    });
    var svg = document.querySelector("svg.ov-dia");
    var canvas = document.querySelector(".ov-canvas");
    var pickCrow = document.querySelector('.ov-row[data-id="amecro"] .ov-pick');
    var crowG = document.querySelector('.ov-bird[data-id="amecro"]');
    var resetBtn = document.querySelector('.ov-reset');
    var zeroCrow = document.querySelector('.ov-row[data-id="amecro"] .ov-zero');
    function tagBox(sel, id) {
      var el = document.querySelector(sel + '[data-id="' + id + '"]');
      if (!el || !canvas) return null;
      var r = el.getBoundingClientRect(), cr = canvas.getBoundingClientRect();
      return {
        id: id,
        left: r.left, top: r.top, right: r.right, bottom: r.bottom,
        fontSize: parseFloat(getComputedStyle(el).fontSize),
        inside: r.left >= cr.left - 1 && r.right <= cr.right + 1 && r.top >= cr.top - 1 && r.bottom <= cr.bottom + 1
      };
    }
    function allTagBoxes(sel) {
      var list = [];
      document.querySelectorAll(sel + "[data-id]").forEach(function (el) {
        list.push(tagBox(sel, el.getAttribute("data-id")));
      });
      return list;
    }
    out.tagRects = allTagBoxes(".ov-tag.ov-wtag");
    out.cubeTagRects = allTagBoxes(".ov-tag.ov-ctag");
    out.tagBeforeDrag = tagBox(".ov-tag.ov-wtag", "amecro");
    if (svg && crowG && typeof PointerEvent !== "undefined") {
      var r0 = crowG.getBoundingClientRect();
      var x0 = r0.left + r0.width / 2, y0 = r0.top + r0.height / 2;
      crowG.dispatchEvent(new PointerEvent("pointerdown", {bubbles:true, clientX:x0, clientY:y0, button:0, pointerId:1}));
      crowG.dispatchEvent(new PointerEvent("pointermove", {bubbles:true, clientX:x0+40, clientY:y0+30, pointerId:1}));
      crowG.dispatchEvent(new PointerEvent("pointerup",   {bubbles:true, clientX:x0+40, clientY:y0+30, pointerId:1}));
      out.unselHasSel = svg.classList.contains("has-sel");
      out.unselDragX = crowG.getAttribute("data-drag-x");
      out.unselDragY = crowG.getAttribute("data-drag-y");
      out.unselURL = location.search;
    }
    if (pickCrow && crowG) {
      pickCrow.click();
      var birds = document.querySelectorAll('.ov-birds .ov-bird');
      out.selectedLast = birds[birds.length - 1].getAttribute("data-id");
      out.hasSel = svg && svg.classList.contains("has-sel");
    }
    if (svg && crowG && typeof PointerEvent !== "undefined") {
      var r = svg.getBoundingClientRect();
      var x = r.left + r.width / 2, y = r.top + r.height / 2;
      svg.dispatchEvent(new PointerEvent("pointerdown", {bubbles:true, clientX:x, clientY:y, button:0, pointerId:2}));
      svg.dispatchEvent(new PointerEvent("pointermove", {bubbles:true, clientX:x+40, clientY:y+30, pointerId:2}));
      svg.dispatchEvent(new PointerEvent("pointerup",   {bubbles:true, clientX:x+40, clientY:y+30, pointerId:2}));
      out.afterDragX = crowG.getAttribute("data-drag-x");
      out.afterDragY = crowG.getAttribute("data-drag-y");
      out.afterDragURL = location.search;
      out.resetHiddenAfterDrag = !!(resetBtn && resetBtn.hasAttribute("hidden"));
      out.zeroHiddenAfterDrag = !!(zeroCrow && zeroCrow.hasAttribute("hidden"));
      var billLink = document.querySelector('.ov-bar .seg a[href*="align=bill"]');
      out.billHrefAfterDrag = billLink ? billLink.getAttribute("href") : "";
      out.tagAfterDrag = tagBox(".ov-tag.ov-wtag", "amecro");
    }
    document.dispatchEvent(new KeyboardEvent("keydown", {key: "ArrowLeft", shiftKey: true, bubbles: true}));
    document.dispatchEvent(new KeyboardEvent("keydown", {key: "ArrowUp", shiftKey: true, bubbles: true}));
    if (crowG) {
      out.afterKeyX = crowG.getAttribute("data-drag-x");
      out.afterKeyY = crowG.getAttribute("data-drag-y");
    }
    if (resetBtn) {
      resetBtn.click();
      out.afterResetX = crowG ? crowG.getAttribute("data-drag-x") : "";
      out.afterResetY = crowG ? crowG.getAttribute("data-drag-y") : "";
      out.afterResetURL = location.search;
      out.resetHiddenAfterReset = resetBtn.hasAttribute("hidden");
      out.tagAfterReset = tagBox(".ov-tag.ov-wtag", "amecro");
    }
    document.documentElement.setAttribute("data-ovprobe", JSON.stringify(out));
  }, 150);
});`

	mux := instrument(s, probeJS, nil)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/__phone" {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			fmt.Fprintf(w, `<!doctype html><html><body style="margin:0">`+
				`<iframe id="f" src="%s" style="border:0;width:%dpx;height:900px"></iframe>`+
				`<script>setInterval(function () {`+
				`var d = document.getElementById('f').contentDocument;`+
				`var v = d && d.documentElement.getAttribute('data-ovprobe');`+
				`if (v) document.documentElement.setAttribute('data-ovprobe', v);`+
				`}, 80);</script></body></html>`,
				html.EscapeString(r.URL.Query().Get("u"))+"#ovprobe=1", phoneWidth)
			return
		}
		mux.ServeHTTP(w, r)
	}))
	defer ts.Close()

	reProbe := regexp.MustCompile(`data-ovprobe="([^"]*)"`)

	// Desktop interactive check (800px).
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	out, err := exec.CommandContext(ctx, chrome, "--headless", "--disable-gpu",
		"--no-first-run", "--no-default-browser-check", "--disable-background-networking",
		"--disable-component-update", "--disable-sync", "--disable-extensions",
		"--user-data-dir="+profileDir(t), "--window-size=800,900",
		"--virtual-time-budget=3000", "--dump-dom",
		ts.URL+"/?show=overlay&pin=annhum&pin=amecro&pin=greblu#ovprobe=1").Output()
	if err != nil {
		t.Fatalf("chrome desktop: %v", err)
	}
	m := reProbe.FindSubmatch(out)
	if m == nil {
		t.Fatal("desktop overlay probe missing")
	}
	type ovTagBox struct {
		ID       string  `json:"id"`
		Left     float64 `json:"left"`
		Top      float64 `json:"top"`
		Right    float64 `json:"right"`
		Bottom   float64 `json:"bottom"`
		FontSize float64 `json:"fontSize"`
		Inside   bool    `json:"inside"`
	}
	type ovProbeResult struct {
		Width                 int        `json:"width"`
		ScrollW               int        `json:"scrollW"`
		ShowTops              []int      `json:"showTops"`
		BarTops               []int      `json:"barTops"`
		OverflowRight         int        `json:"overflowRight"`
		UnselHasSel           bool       `json:"unselHasSel"`
		UnselDragX            string     `json:"unselDragX"`
		UnselDragY            string     `json:"unselDragY"`
		UnselURL              string     `json:"unselURL"`
		SelectedLast          string     `json:"selectedLast"`
		HasSel                bool       `json:"hasSel"`
		AfterDragX            string     `json:"afterDragX"`
		AfterDragY            string     `json:"afterDragY"`
		AfterDragURL          string     `json:"afterDragURL"`
		ResetHiddenAfterDrag  bool       `json:"resetHiddenAfterDrag"`
		ZeroHiddenAfterDrag   bool       `json:"zeroHiddenAfterDrag"`
		BillHrefAfterDrag     string     `json:"billHrefAfterDrag"`
		AfterKeyX             string     `json:"afterKeyX"`
		AfterKeyY             string     `json:"afterKeyY"`
		AfterResetX           string     `json:"afterResetX"`
		AfterResetY           string     `json:"afterResetY"`
		AfterResetURL         string     `json:"afterResetURL"`
		ResetHiddenAfterReset bool       `json:"resetHiddenAfterReset"`
		TagRects              []ovTagBox `json:"tagRects"`
		CubeTagRects          []ovTagBox `json:"cubeTagRects"`
		TagBeforeDrag         *ovTagBox  `json:"tagBeforeDrag"`
		TagAfterDrag          *ovTagBox  `json:"tagAfterDrag"`
		TagAfterReset         *ovTagBox  `json:"tagAfterReset"`
	}
	checkTags := func(label string, tags []ovTagBox, wantCount int) {
		t.Helper()
		if len(tags) != wantCount {
			t.Errorf("%s: got %d .ov-tag badges, want %d", label, len(tags), wantCount)
		}
		for _, tg := range tags {
			if tg.FontSize < 13 {
				t.Errorf("%s badge %s fontSize = %.1fpx, want >= 13px", label, tg.ID, tg.FontSize)
			}
			if !tg.Inside {
				t.Errorf("%s badge %s fell outside .ov-canvas: %+v", label, tg.ID, tg)
			}
		}
		for i := 0; i < len(tags); i++ {
			for j := i + 1; j < len(tags); j++ {
				a, b := tags[i], tags[j]
				if a.Left < b.Right && a.Right > b.Left && a.Top < b.Bottom && a.Bottom > b.Top {
					t.Errorf("%s badges %s (%+v) and %s (%+v) overlap", label, a.ID, a, b.ID, b)
				}
			}
		}
	}
	checkCubeTags := func(label string, tags []ovTagBox, wantOrder []string) {
		t.Helper()
		checkTags(label+" cubes", tags, len(wantOrder))
		if len(tags) != len(wantOrder) {
			return
		}
		for i, tg := range tags {
			if tg.ID != wantOrder[i] {
				t.Errorf("%s cube badge %d id = %q, want %q", label, i, tg.ID, wantOrder[i])
			}
			if i > 0 {
				if math.Abs(tg.Top-tags[i-1].Top) > 2 {
					t.Errorf("%s cube badges %s (top=%.1f) and %s (top=%.1f) not on one horizontal row",
						label, tags[i-1].ID, tags[i-1].Top, tg.ID, tg.Top)
				}
				if tg.Left <= tags[i-1].Left {
					t.Errorf("%s cube badges %s (left=%.1f) and %s (left=%.1f) not increasing left-to-right",
						label, tags[i-1].ID, tags[i-1].Left, tg.ID, tg.Left)
				}
			}
		}
	}
	pinnedOrder := func(query string) []string {
		p := s.nextPage(httptest.NewRequest("GET", query, nil))
		var ids []string
		for _, tl := range p.GroupTiles {
			ids = append(ids, tl.S.ID)
		}
		return ids
	}
	var res ovProbeResult
	if err := json.Unmarshal([]byte(html.UnescapeString(string(m[1]))), &res); err != nil {
		t.Fatal(err)
	}
	if len(res.BarTops) != 6 {
		t.Fatalf("desktop ov-bar has %d controls (4 align + Spread ↔ + Spread ↕), tops=%v", len(res.BarTops), res.BarTops)
	}
	for i, top := range res.BarTops {
		if top-res.BarTops[0] > 1 || res.BarTops[0]-top > 1 {
			t.Errorf("desktop ov-bar control %d top=%dpx, want aligned with first (%dpx): %v", i, top, res.BarTops[0], res.BarTops)
		}
	}
	if res.UnselHasSel || (res.UnselDragX != "" && res.UnselDragX != "0") || (res.UnselDragY != "" && res.UnselDragY != "0") || strings.Contains(res.UnselURL, "drag=") {
		t.Errorf("unselected drag moved or selected bird: hasSel=%v dx=%q dy=%q url=%q",
			res.UnselHasSel, res.UnselDragX, res.UnselDragY, res.UnselURL)
	}
	if !res.HasSel || res.SelectedLast != "amecro" {
		t.Errorf("clicking amecro legend row: hasSel=%v selectedLast=%q, want true/\"amecro\"", res.HasSel, res.SelectedLast)
	}
	dragX, _ := strconv.Atoi(res.AfterDragX)
	dragY, _ := strconv.Atoi(res.AfterDragY)
	if dragX <= 0 || dragY <= 0 {
		t.Errorf("after selected 2D drag (+40,+30 px): (dragX,dragY)=(%q,%q), want both > 0", res.AfterDragX, res.AfterDragY)
	}
	if !strings.Contains(res.AfterDragURL, "drag=amecro%3A") && !strings.Contains(res.AfterDragURL, "drag=amecro:") {
		t.Errorf("after 2D drag URL = %q, want drag=amecro:...", res.AfterDragURL)
	}
	if res.ResetHiddenAfterDrag || res.ZeroHiddenAfterDrag {
		t.Errorf("after 2D drag: resetHidden=%v zeroHidden=%v, want both visible", res.ResetHiddenAfterDrag, res.ZeroHiddenAfterDrag)
	}
	if wantBill := fmt.Sprintf("drag=amecro:%d,0", dragX); !strings.Contains(res.BillHrefAfterDrag, wantBill) &&
		!strings.Contains(res.BillHrefAfterDrag, url.QueryEscape(fmt.Sprintf("amecro:%d,0", dragX))) {
		t.Errorf("after 2D drag Bill align href = %q, want %q preserved", res.BillHrefAfterDrag, wantBill)
	}
	keyX, _ := strconv.Atoi(res.AfterKeyX)
	keyY, _ := strconv.Atoi(res.AfterKeyY)
	if keyX >= dragX || keyY >= dragY {
		t.Errorf("after ArrowLeft+ArrowUp: (keyX,keyY)=(%d,%d) did not decrease from (%d,%d)", keyX, keyY, dragX, dragY)
	}
	if res.AfterResetX != "0" || res.AfterResetY != "0" || strings.Contains(res.AfterResetURL, "drag=") || !res.ResetHiddenAfterReset {
		t.Errorf("after Reset drags: (x,y)=(%q,%q) url=%q hidden=%v, want (0,0), no drag=, hidden=true",
			res.AfterResetX, res.AfterResetY, res.AfterResetURL, res.ResetHiddenAfterReset)
	}
	checkTags("desktop [annhum,amecro,greblu]", res.TagRects, 3)
	checkCubeTags("desktop [annhum,amecro,greblu]", res.CubeTagRects, pinnedOrder("/?show=pinned&pin=annhum&pin=amecro&pin=greblu"))
	if res.TagBeforeDrag == nil || res.TagAfterDrag == nil || res.TagAfterReset == nil {
		t.Errorf("missing amecro .ov-tag position during drag lifecycle: before=%v after=%v reset=%v",
			res.TagBeforeDrag, res.TagAfterDrag, res.TagAfterReset)
	} else {
		if res.TagAfterDrag.Left <= res.TagBeforeDrag.Left+10 || res.TagAfterDrag.Top <= res.TagBeforeDrag.Top+10 {
			t.Errorf("amecro .ov-tag did not track 2D drag (+40,+30): before=%+v after=%+v", res.TagBeforeDrag, res.TagAfterDrag)
		}
		if math.Abs(res.TagAfterReset.Left-res.TagBeforeDrag.Left) > 2 || math.Abs(res.TagAfterReset.Top-res.TagBeforeDrag.Top) > 2 {
			t.Errorf("amecro .ov-tag did not return on Reset drags: before=%+v reset=%+v", res.TagBeforeDrag, res.TagAfterReset)
		}
	}

	// Close-wingtip de-overlap check (thimur 726mm & blasti 750mm with Spread ↕ off).
	closeOut, err := exec.CommandContext(ctx, chrome, "--headless", "--disable-gpu",
		"--no-first-run", "--no-default-browser-check", "--disable-background-networking",
		"--disable-component-update", "--disable-sync", "--disable-extensions",
		"--user-data-dir="+profileDir(t), "--window-size=800,900",
		"--virtual-time-budget=3000", "--dump-dom",
		ts.URL+"/?show=overlay&pin=thimur&pin=amecro&pin=blasti#ovprobe=1").Output()
	if err != nil {
		t.Fatalf("chrome close wingtips: %v", err)
	}
	cm := reProbe.FindSubmatch(closeOut)
	if cm == nil {
		t.Fatal("close-wingtip overlay probe missing")
	}
	var closeRes ovProbeResult
	if err := json.Unmarshal([]byte(html.UnescapeString(string(cm[1]))), &closeRes); err != nil {
		t.Fatal(err)
	}
	checkTags("close wingtips [thimur,amecro,blasti]", closeRes.TagRects, 3)
	checkCubeTags("close wingtips [thimur,amecro,blasti]", closeRes.CubeTagRects, pinnedOrder("/?show=pinned&pin=thimur&pin=amecro&pin=blasti"))

	// Phone 360px check: no horizontal overflow and Show row holds 5 controls.
	phoneOut, err := exec.CommandContext(ctx, chrome, "--headless", "--disable-gpu",
		"--no-first-run", "--no-default-browser-check", "--disable-background-networking",
		"--disable-component-update", "--disable-sync", "--disable-extensions",
		"--hide-scrollbars", "--user-data-dir="+profileDir(t), "--window-size=800,900",
		"--virtual-time-budget=5000", "--dump-dom",
		ts.URL+"/__phone?u="+url.QueryEscape("/?show=overlay&pin=annhum&pin=amecro&pin=greblu")).Output()
	if err != nil {
		t.Fatalf("chrome phone: %v", err)
	}
	pm := reProbe.FindSubmatch(phoneOut)
	if pm == nil {
		t.Fatal("phone overlay probe missing")
	}
	res = ovProbeResult{}
	if err := json.Unmarshal([]byte(html.UnescapeString(string(pm[1]))), &res); err != nil {
		t.Fatal(err)
	}
	if res.Width != phoneWidth {
		t.Fatalf("phone width = %d, want %d", res.Width, phoneWidth)
	}
	if res.ScrollW > res.Width {
		t.Errorf("360px overlay page scrollWidth=%d > clientWidth=%d", res.ScrollW, res.Width)
	}
	if len(res.ShowTops) != 5 {
		t.Errorf("Show row has %d controls at 3 pins, want 5 (First, All, Pinned, Overlay, Clear pins)", len(res.ShowTops))
	}
	if res.OverflowRight > 0 {
		t.Errorf("Show row control reached right=%dpx past %dpx", res.OverflowRight, res.Width)
	}
	checkCubeTags("phone 360px [annhum,amecro,greblu]", res.CubeTagRects, pinnedOrder("/?show=pinned&pin=annhum&pin=amecro&pin=greblu"))
}
