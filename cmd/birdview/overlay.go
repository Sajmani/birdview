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
	"fmt"
	"html/template"
	"math"
	"net/url"
	"sort"
	"strconv"
	"strings"

	"birdview/internal/dataset"
	"birdview/internal/view"
)

const (
	overlayMinPins = 2
	overlayMaxPins = 8

	overlayBoxW       = 640.0
	overlayDrawH      = 420.0
	overlayPad        = 24.0
	overlaySpreadStep = 14.0
	overlayMaxOffset  = 5000
)

// overlayView is the home page's show=overlay state: 2 to 8 pinned birds
// drawn together on one shared-scale canvas, each in its own color
// (decisions.overlayMode).
type overlayView struct {
	TooFew, TooMany bool
	PinCount        int
	MaxPins         int

	ViewX, ViewW float64
	BoxW, BoxH   float64
	PxPerMm      float64
	GuideY       float64
	Align        string
	AlignLabel   string
	Spread       bool
	Shift        bool
	NudgeMM      int
	SpreadURL    template.URL
	ShiftURL     template.URL
	ResetURL     template.URL
	Aligns       []overlayAlignChoice
	Layers       []overlayBird // sorted largest behind smallest (SVG paint order)
	Legend       []overlayBird // sorted by active Order by setting (matching Pinned list / colors)
}

type overlayAlignChoice struct {
	Key   string
	Label string
	On    bool
	URL   template.URL
}

// overlayBird is one pinned bird in both the SVG layers and the legend.
type overlayBird struct {
	ID         string
	Name       string
	Scientific string
	Anchor     string
	ColorIdx   int
	DetailURL  template.URL
	PinURL     template.URL

	Length string
	Wing   string
	Mass   string

	Geom        view.Geometry
	WingModeled bool
	MassCube    massCube

	BaseX   float64
	BaseY   float64
	X       float64
	Y       float64
	SpanPx  float64
	DragXMM int
	DragYMM int
	Zero    template.URL
}

var overlayAlignOptions = []struct {
	key, label string
}{
	{"wings", "Wings"},
	{"bill", "Bill"},
	{"tail", "Tail"},
	{"center", "Center"},
}

// overlayAlign returns the chosen alignment anchor ("wings", "bill", "tail", or
// "center"), defaulting to "wings".
func overlayAlign(q url.Values) string {
	switch q.Get("align") {
	case "bill", "tail", "center":
		return q.Get("align")
	default:
		return "wings"
	}
}

// overlaySpread reports whether horizontal centerline spread is active.
func overlaySpread(q url.Values) bool {
	return q.Get("spread") == "1"
}

// overlayShift reports whether automatic vertical staggering is active.
func overlayShift(q url.Values) bool {
	return q.Get("shift") == "1"
}

type overlayDrag struct {
	X, Y int
}

func clampOverlayOffset(mm int) int {
	if mm < -overlayMaxOffset {
		return -overlayMaxOffset
	}
	if mm > overlayMaxOffset {
		return overlayMaxOffset
	}
	return mm
}

// overlayDrags parses repeatable drag=<id>:<dx>,<dy> query parameters for
// pinned birds, returning non-zero millimeter offsets keyed by species ID.
func overlayDrags(q url.Values, allowed map[string]bool) map[string]overlayDrag {
	out := map[string]overlayDrag{}
	for _, raw := range q["drag"] {
		id, coords, ok := strings.Cut(raw, ":")
		if !ok || !allowed[id] {
			continue
		}
		xStr, yStr, ok := strings.Cut(coords, ",")
		if !ok {
			continue
		}
		dx, errX := strconv.Atoi(xStr)
		dy, errY := strconv.Atoi(yStr)
		if errX != nil || errY != nil {
			continue
		}
		dx, dy = clampOverlayOffset(dx), clampOverlayOffset(dy)
		if dx == 0 && dy == 0 {
			continue
		}
		out[id] = overlayDrag{X: dx, Y: dy}
	}
	return out
}

func overlayWriteDrags(q url.Values, drags map[string]overlayDrag) {
	q.Del("drag")
	if len(drags) == 0 {
		return
	}
	ids := make([]string, 0, len(drags))
	for id, d := range drags {
		if d.X != 0 || d.Y != 0 {
			ids = append(ids, id)
		}
	}
	sort.Strings(ids)
	for _, id := range ids {
		d := drags[id]
		q.Add("drag", fmt.Sprintf("%s:%d,%d", id, d.X, d.Y))
	}
}

// overlayClearOneDrag returns the page URL with species id's drag cleared,
// keeping valid drags for all other pinned birds.
func overlayClearOneDrag(q url.Values, allowed map[string]bool, id string) template.URL {
	nq := cloneValues(q)
	drags := overlayDrags(nq, allowed)
	delete(drags, id)
	overlayWriteDrags(nq, drags)
	return nextURL(nq)
}

// overlayClearVerticalDrags resets vertical drags on q to 0 while preserving
// any horizontal drags (used when switching Align presets).
func overlayClearVerticalDrags(q url.Values, allowed map[string]bool) {
	drags := overlayDrags(q, allowed)
	for id, d := range drags {
		if d.X == 0 {
			delete(drags, id)
		} else {
			drags[id] = overlayDrag{X: d.X, Y: 0}
		}
	}
	overlayWriteDrags(q, drags)
}

// overlayDropPin removes any drag=<id>:<dx>,<dy> entry when id is unpinned.
func overlayDropPin(q url.Values, id string) {
	prefix := id + ":"
	var kept []string
	for _, raw := range q["drag"] {
		if !strings.HasPrefix(raw, prefix) {
			kept = append(kept, raw)
		}
	}
	if len(kept) == 0 {
		q.Del("drag")
	} else {
		q["drag"] = kept
	}
}

func overlayAnchorY(g view.Geometry, align string) float64 {
	switch align {
	case "bill":
		return g.YBillTip
	case "tail":
		return g.YTailTip
	case "center":
		return g.YTailTip / 2
	default:
		return g.YShoulder
	}
}

func (s *server) buildOverlay(q url.Values, species []*dataset.Species, opt options, u units) *overlayView {
	ov := &overlayView{
		PinCount: len(species),
		MaxPins:  overlayMaxPins,
	}
	if len(species) < overlayMinPins {
		ov.TooFew = true
		return ov
	}
	if len(species) > overlayMaxPins {
		ov.TooMany = true
		return ov
	}

	allowed := make(map[string]bool, len(species))
	for _, sp := range species {
		allowed[sp.ID] = true
	}

	align := overlayAlign(q)
	spread := overlaySpread(q)
	shift := overlayShift(q)
	ov.Align = align
	ov.Spread = spread
	ov.Shift = shift
	for _, optAlign := range overlayAlignOptions {
		on := optAlign.key == align
		if on {
			ov.AlignLabel = optAlign.label
		}
		aq := cloneValues(q)
		overlayClearVerticalDrags(aq, allowed)
		if optAlign.key == "wings" {
			aq.Del("align")
		} else {
			aq.Set("align", optAlign.key)
		}
		ov.Aligns = append(ov.Aligns, overlayAlignChoice{
			Key:   optAlign.key,
			Label: optAlign.label,
			On:    on,
			URL:   nextURL(aq),
		})
	}

	sq := cloneValues(q)
	if spread {
		sq.Del("spread")
	} else {
		sq.Set("spread", "1")
	}
	ov.SpreadURL = nextURL(sq)

	vhq := cloneValues(q)
	if shift {
		vhq.Del("shift")
	} else {
		vhq.Set("shift", "1")
	}
	ov.ShiftURL = nextURL(vhq)

	drags := overlayDrags(q, allowed)
	if len(drags) > 0 {
		rq := cloneValues(q)
		rq.Del("drag")
		ov.ResetURL = nextURL(rq)
	}

	refLen, refSpan := view.Reference(species)
	nudgeMM := max(1, int(math.Round(refLen*0.05)))
	ov.NudgeMM = nudgeMM

	edgePad := math.Max(overlayPad, float64(len(species)-1)*overlaySpreadStep/2+16)
	layoutBox := view.Box{W: overlayBoxW - 2*edgePad, H: overlayDrawH}

	type placed struct {
		sp       *dataset.Species
		idx      int
		geom     view.Geometry
		baseX    float64
		staggerY float64
		anchorY  float64
		drag     overlayDrag
		dragXPx  float64
		dragYPx  float64
	}
	items := make([]placed, len(species))
	maxAnchor := 0.0
	for i, sp := range species {
		g := view.Layout(sp, refLen, refSpan, layoutBox)
		anc := overlayAnchorY(g, align)
		if anc > maxAnchor {
			maxAnchor = anc
		}
		stepOffset := (float64(i) - float64(len(species)-1)/2) * overlaySpreadStep
		baseX := edgePad
		if spread {
			baseX += stepOffset
		}
		var staggerY float64
		if shift {
			staggerY = stepOffset
		}
		dr := drags[sp.ID]
		items[i] = placed{
			sp:       sp,
			idx:      i,
			geom:     g,
			baseX:    baseX,
			staggerY: staggerY,
			anchorY:  anc,
			drag:     dr,
			dragXPx:  float64(dr.X) * g.PxPerMm,
			dragYPx:  float64(dr.Y) * g.PxPerMm,
		}
		ov.PxPerMm = g.PxPerMm
	}

	minX := 0.0
	maxX := overlayBoxW
	minY := 0.0
	maxY := overlayDrawH
	maxCubeHeight := 0.0
	for _, it := range items {
		x := it.baseX + it.dragXPx
		if left := x + it.geom.CX - it.geom.CrossbarHalfWidth - overlayPad; left < minX {
			minX = left
		}
		if right := x + it.geom.CX + it.geom.CrossbarHalfWidth + overlayPad; right > maxX {
			maxX = right
		}
		baseY := maxAnchor - it.anchorY + it.staggerY
		y := baseY + it.dragYPx
		if y < minY {
			minY = y
		}
		if bot := y + it.geom.YTailTip; bot > maxY {
			maxY = bot
		}
		if h := massCubeSide(it.sp.Mass.Value) * it.geom.PxPerMm * (1 + massCubeDepth); h > maxCubeHeight {
			maxCubeHeight = h
		}
	}
	topOffset := overlayPad - minY
	boxH := maxY + topOffset + overlayPad
	if minH := maxCubeHeight + 2*overlayPad; boxH < minH {
		boxH = minH
	}
	ov.ViewX = minX
	ov.ViewW = maxX - minX
	ov.BoxW = overlayBoxW
	ov.BoxH = boxH
	ov.GuideY = maxAnchor + topOffset

	birds := make([]overlayBird, len(items))
	for i, it := range items {
		sp := it.sp
		baseY := maxAnchor - it.anchorY + it.staggerY + topOffset
		x := it.baseX + it.dragXPx
		y := baseY + it.dragYPx
		cube := newMassCube(sp, boxH, it.geom.PxPerMm)
		cube.X = minX + 12
		var zeroURL template.URL
		if it.drag.X != 0 || it.drag.Y != 0 {
			zeroURL = overlayClearOneDrag(q, allowed, sp.ID)
		}
		birds[i] = overlayBird{
			ID:          sp.ID,
			Name:        sp.Name(opt.Locale),
			Scientific:  sp.ScientificName,
			Anchor:      pinnedAnchor(sp.ID),
			ColorIdx:    it.idx,
			DetailURL:   nextDetailLink(q, sp.ID),
			PinURL:      nextPin(q, sp.ID),
			Length:      u.format(sp.TotalLength.Value, view.ByLength),
			Wing:        u.format(sp.Wingspan.Value, view.ByWingspan),
			Mass:        u.format(sp.Mass.Value, view.ByMass),
			Geom:        it.geom,
			WingModeled: sp.Wingspan.Prov.Kind != dataset.Measured,
			MassCube:    cube,
			BaseX:       it.baseX,
			BaseY:       baseY,
			X:           x,
			Y:           y,
			SpanPx:      it.geom.YTailTip,
			DragXMM:     it.drag.X,
			DragYMM:     it.drag.Y,
			Zero:        zeroURL,
		}
	}

	ov.Legend = make([]overlayBird, len(birds))
	copy(ov.Legend, birds)

	ov.Layers = make([]overlayBird, len(birds))
	copy(ov.Layers, birds)
	sort.Slice(ov.Layers, func(i, j int) bool {
		li, lj := ov.Layers[i], ov.Layers[j]
		si, sj := s.byID[li.ID], s.byID[lj.ID]
		if si.TotalLength.Value != sj.TotalLength.Value {
			return si.TotalLength.Value > sj.TotalLength.Value
		}
		if si.Wingspan.Value != sj.Wingspan.Value {
			return si.Wingspan.Value > sj.Wingspan.Value
		}
		return li.ID < lj.ID
	})
	return ov
}
