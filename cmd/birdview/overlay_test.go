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
	_, ov := get(t, s, "/?show=overlay&align=tail&spread=1&shift=1&dy=amecro:-20&pin=amecro&pin=annhum")
	pinBtn := mustMatch(t, ov, `<a class="btn" href="([^"]+)"[^>]*>Pinned \(2\)</a>`, "Pinned button")
	href := html.UnescapeString(pinBtn)
	for _, drop := range []string{"align=", "spread=", "shift=", "dy="} {
		if strings.Contains(href, drop) {
			t.Errorf("Pinned button %q retained overlay param %q", href, drop)
		}
	}
}

// TestOverlayGeometryAlignmentAndURL checks shared scale, largest-behind-smallest
// layering, alignment presets, horizontal spread, millimeter dy= shifts, and
// preserved provenance cues.
//
// Verifies: app.overlayModeComparesPinnedBirds.
func TestOverlayGeometryAlignmentAndURL(t *testing.T) {
	s := overlayTestServer(t)

	// Three birds with distinct silhouettes and provenance: hummingbird (smallest),
	// crow (wing modeled, solid tail/bill, dashed body), heron (largest, folded neck).
	parseAttrs := func(body string) (layers []struct {
		id, cls string
		baseY   float64
		dx, y   float64
		dyMM    int
	}, guideY float64) {
		gm := mustMatch(t, body, `<line class="ov-guide"[^>]*y1="([^"]+)"`, "overlay guide")
		guideY, _ = strconv.ParseFloat(gm, 64)
		reBird := regexp.MustCompile(`<g class="ov-bird (ov-c\d)" data-id="([^"]+)" data-base-y="([^"]+)" data-dx="([^"]+)" data-dy="([^"]+)"[^>]*transform="translate\(([^,]+),([^)]+)\)"`)
		for _, m := range reBird.FindAllStringSubmatch(body, -1) {
			by, _ := strconv.ParseFloat(m[3], 64)
			dx, _ := strconv.ParseFloat(m[6], 64)
			y, _ := strconv.ParseFloat(m[7], 64)
			dy, _ := strconv.Atoi(m[5])
			layers = append(layers, struct {
				id, cls string
				baseY   float64
				dx, y   float64
				dyMM    int
			}{m[2], m[1], by, dx, y, dy})
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
	byID := map[string]struct{ baseY, dx, y float64 }{}
	for _, l := range layers {
		byID[l.id] = struct{ baseY, dx, y float64 }{l.baseY, l.dx, l.y}
	}
	// Non-spread: every bird shares the exact same horizontal dx.
	if byID["greblu"].dx != byID["amecro"].dx || byID["amecro"].dx != byID["annhum"].dx {
		t.Errorf("without spread=1, dx differed: %v", byID)
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

	// Spread=1 separates dx by pin order symmetrically while keeping y untouched.
	_, spBody := get(t, s, "/?show=overlay&align=bill&spread=1&pin=annhum&pin=amecro&pin=greblu")
	if !strings.Contains(spBody, ">Spread ↔</a>") || !strings.Contains(spBody, ">Spread ↕</a>") {
		t.Errorf("overlay bar missing Spread ↔ / Spread ↕ toggle labels")
	}
	spLayers, _ := parseAttrs(spBody)
	spByID := map[string]float64{}
	for _, l := range spLayers {
		spByID[l.id] = l.dx
	}
	if !(spByID["annhum"] < spByID["amecro"] && spByID["amecro"] < spByID["greblu"]) {
		t.Errorf("spread=1 dx not ordered by pin order: %v", spByID)
	}

	// Shift=1 staggers y vertically by exact overlaySpreadStep in pin order while keeping dx untouched.
	_, shfBody := get(t, s, "/?show=overlay&align=bill&shift=1&pin=annhum&pin=amecro&pin=greblu")
	shfLayers, _ := parseAttrs(shfBody)
	shfByID := map[string]struct{ dx, y float64 }{}
	for _, l := range shfLayers {
		shfByID[l.id] = struct{ dx, y float64 }{l.dx, l.y}
	}
	if shfByID["annhum"].dx != shfByID["amecro"].dx || shfByID["amecro"].dx != shfByID["greblu"].dx {
		t.Errorf("shift=1 changed horizontal dx: %v", shfByID)
	}
	dy01 := shfByID["amecro"].y - shfByID["annhum"].y
	dy12 := shfByID["greblu"].y - shfByID["amecro"].y
	if math.Abs(dy01-overlaySpreadStep) > 0.05 || math.Abs(dy12-overlaySpreadStep) > 0.05 {
		t.Errorf("shift=1 vertical steps = (%.2f, %.2f), want exact %.2f", dy01, dy12, overlaySpreadStep)
	}

	// Spread=1&shift=1 applies both horizontal and vertical offsets together.
	_, bothBody := get(t, s, "/?show=overlay&align=bill&spread=1&shift=1&pin=annhum&pin=amecro&pin=greblu")
	bothLayers, _ := parseAttrs(bothBody)
	bothByID := map[string]struct{ dx, y float64 }{}
	for _, l := range bothLayers {
		bothByID[l.id] = struct{ dx, y float64 }{l.dx, l.y}
	}
	if !(bothByID["annhum"].dx < bothByID["amecro"].dx && bothByID["amecro"].dx < bothByID["greblu"].dx) ||
		math.Abs((bothByID["amecro"].y-bothByID["annhum"].y)-overlaySpreadStep) > 0.05 {
		t.Errorf("spread=1&shift=1 did not apply both offsets: %v", bothByID)
	}

	// Millimeter shift dy=amecro:-100 shifts amecro up by 100 * PxPerMm without moving others.
	_, shBody := get(t, s, "/?show=overlay&pin=annhum&pin=amecro&pin=greblu&dy=amecro:-100")
	pxMatch := mustMatch(t, shBody, `data-px-per-mm="([^"]+)"`, "pxPerMm")
	pxPerMm, _ := strconv.ParseFloat(pxMatch, 64)
	shLayers, _ := parseAttrs(shBody)
	for _, l := range shLayers {
		if l.id == "amecro" {
			wantShift := -100 * pxPerMm
			gotShift := l.y - l.baseY
			if gotShift-wantShift > 0.2 || wantShift-gotShift > 0.2 {
				t.Errorf("dy=amecro:-100 shifted by %.2f px, want %.2f px (pxPerMm=%.4f)", gotShift, wantShift, pxPerMm)
			}
		} else if l.y != l.baseY {
			t.Errorf("dy=amecro:-100 moved unshifted bird %s: y=%.2f baseY=%.2f", l.id, l.y, l.baseY)
		}
	}
}

// TestOverlayInTheBrowser verifies interactive selection (legend click brings
// bird to front and dims others), vertical nudge updating transform and URL via
// history.replaceState, pointer drag on canvas, and 360px phone layout without
// horizontal overflow.
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
    var pickCrow = document.querySelector('.ov-row[data-id="amecro"] .ov-pick');
    var crowG = document.querySelector('.ov-bird[data-id="amecro"]');
    var upCrow = document.querySelector('.ov-row[data-id="amecro"] .ov-up');
    if (pickCrow && crowG) {
      pickCrow.click();
      var birds = document.querySelectorAll('.ov-birds .ov-bird');
      out.selectedLast = birds[birds.length - 1].getAttribute("data-id");
      out.hasSel = document.querySelector('svg.ov-dia').classList.contains("has-sel");
    }
    if (upCrow) {
      upCrow.click();
      out.afterNudgeURL = location.search;
      out.afterNudgeDy = crowG.getAttribute("data-dy");
      out.afterNudgeTransform = crowG.getAttribute("transform");
    }
    var svg = document.querySelector("svg.ov-dia");
    if (svg && crowG && typeof PointerEvent !== "undefined") {
      var r = crowG.getBoundingClientRect();
      var x = r.left + r.width / 2, y = r.top + r.height / 2;
      svg.dispatchEvent(new PointerEvent("pointerdown", {bubbles:true, clientX:x, clientY:y, button:0, pointerId:1}));
      svg.dispatchEvent(new PointerEvent("pointermove", {bubbles:true, clientX:x, clientY:y+30, pointerId:1}));
      svg.dispatchEvent(new PointerEvent("pointerup",   {bubbles:true, clientX:x, clientY:y+30, pointerId:1}));
      out.afterDragDy = crowG.getAttribute("data-dy");
      out.afterDragURL = location.search;
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
	type ovProbeResult struct {
		Width               int    `json:"width"`
		ScrollW             int    `json:"scrollW"`
		ShowTops            []int  `json:"showTops"`
		BarTops             []int  `json:"barTops"`
		OverflowRight       int    `json:"overflowRight"`
		SelectedLast        string `json:"selectedLast"`
		HasSel              bool   `json:"hasSel"`
		AfterNudgeURL       string `json:"afterNudgeURL"`
		AfterNudgeDy        string `json:"afterNudgeDy"`
		AfterNudgeTransform string `json:"afterNudgeTransform"`
		AfterDragDy         string `json:"afterDragDy"`
		AfterDragURL        string `json:"afterDragURL"`
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
	if !res.HasSel || res.SelectedLast != "amecro" {
		t.Errorf("clicking amecro legend row: hasSel=%v selectedLast=%q, want true/\"amecro\"", res.HasSel, res.SelectedLast)
	}
	if !strings.Contains(res.AfterNudgeURL, "dy=amecro%3A-") && !strings.Contains(res.AfterNudgeURL, "dy=amecro:-") {
		t.Errorf("after nudge URL = %q, want negative dy for amecro", res.AfterNudgeURL)
	}
	if res.AfterNudgeDy == "" || res.AfterNudgeDy == "0" {
		t.Errorf("after nudge data-dy = %q, want non-zero", res.AfterNudgeDy)
	}
	nudgeVal, _ := strconv.Atoi(res.AfterNudgeDy)
	dragVal, _ := strconv.Atoi(res.AfterDragDy)
	if dragVal <= nudgeVal {
		t.Errorf("after downward drag data-dy=%d did not increase from nudge=%d (url=%q)", dragVal, nudgeVal, res.AfterDragURL)
	}

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
}
