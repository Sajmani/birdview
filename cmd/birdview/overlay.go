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

	BoxW, BoxH float64
	PxPerMm    float64
	GuideY     float64
	Align      string
	AlignLabel string
	Spread     bool
	Shift      bool
	Metric     bool
	NudgeMM    int
	SpreadURL  template.URL
	ShiftURL   template.URL
	ResetURL   template.URL
	Aligns     []overlayAlignChoice
	Layers     []overlayBird // sorted largest behind smallest (SVG paint order)
	Legend     []overlayBird // sorted by pinOrder (matching Pinned list / colors)
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

	BaseY  float64
	DX     float64
	Y      float64
	SpanPx float64
	DyMM   int
	DyText string
	UpURL  template.URL
	DnURL  template.URL
	Zero   template.URL
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

// overlayOffsets parses repeatable dy=<id>:<mm> query parameters for pinned
// birds, returning non-zero millimeter vertical offsets keyed by species ID.
func overlayOffsets(q url.Values, allowed map[string]bool) map[string]int {
	out := map[string]int{}
	for _, raw := range q["dy"] {
		id, val, ok := strings.Cut(raw, ":")
		if !ok || !allowed[id] {
			continue
		}
		mm, err := strconv.Atoi(val)
		if err != nil || mm == 0 {
			continue
		}
		if mm < -overlayMaxOffset {
			mm = -overlayMaxOffset
		} else if mm > overlayMaxOffset {
			mm = overlayMaxOffset
		}
		out[id] = mm
	}
	return out
}

func overlayWriteOffsets(q url.Values, dy map[string]int) {
	q.Del("dy")
	if len(dy) == 0 {
		return
	}
	ids := make([]string, 0, len(dy))
	for id, mm := range dy {
		if mm != 0 {
			ids = append(ids, id)
		}
	}
	sort.Strings(ids)
	for _, id := range ids {
		q.Add("dy", id+":"+strconv.Itoa(dy[id]))
	}
}

// overlaySetOffset returns the page URL with species id's vertical offset set
// to mm (or cleared when mm == 0), keeping only valid offsets for pinned birds.
func overlaySetOffset(q url.Values, allowed map[string]bool, id string, mm int) template.URL {
	nq := cloneValues(q)
	dy := overlayOffsets(nq, allowed)
	if mm == 0 {
		delete(dy, id)
	} else {
		if mm < -overlayMaxOffset {
			mm = -overlayMaxOffset
		} else if mm > overlayMaxOffset {
			mm = overlayMaxOffset
		}
		dy[id] = mm
	}
	overlayWriteOffsets(nq, dy)
	return nextURL(nq)
}

// overlayDropPin removes any dy=<id>:<mm> entry when id is unpinned.
func overlayDropPin(q url.Values, id string) {
	prefix := id + ":"
	var kept []string
	for _, raw := range q["dy"] {
		if !strings.HasPrefix(raw, prefix) {
			kept = append(kept, raw)
		}
	}
	if len(kept) == 0 {
		q.Del("dy")
	} else {
		q["dy"] = kept
	}
}

func formatOverlayShift(mm int, metric bool) string {
	if mm == 0 {
		return ""
	}
	if metric {
		return fmt.Sprintf("%+.1f cm", float64(mm)/10)
	}
	return fmt.Sprintf("%+.2f in", float64(mm)/25.4)
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

func (s *server) buildOverlay(q url.Values, pinOrder []string, opt options, u units) *overlayView {
	ov := &overlayView{
		PinCount: len(pinOrder),
		MaxPins:  overlayMaxPins,
		Metric:   opt.Metric3,
	}
	if len(pinOrder) < overlayMinPins {
		ov.TooFew = true
		return ov
	}
	if len(pinOrder) > overlayMaxPins {
		ov.TooMany = true
		return ov
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
		aq.Del("dy")
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

	allowed := make(map[string]bool, len(pinOrder))
	species := make([]*dataset.Species, 0, len(pinOrder))
	for _, id := range pinOrder {
		if sp, ok := s.byID[id]; ok {
			allowed[id] = true
			species = append(species, sp)
		}
	}
	dyMap := overlayOffsets(q, allowed)
	if len(dyMap) > 0 {
		rq := cloneValues(q)
		rq.Del("dy")
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
		dx       float64
		staggerY float64
		anchorY  float64
		dyMM     int
		shiftPx  float64
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
		dx := edgePad
		if spread {
			dx += stepOffset
		}
		var staggerY float64
		if shift {
			staggerY = stepOffset
		}
		dyMM := dyMap[sp.ID]
		items[i] = placed{
			sp:       sp,
			idx:      i,
			geom:     g,
			dx:       dx,
			staggerY: staggerY,
			anchorY:  anc,
			dyMM:     dyMM,
			shiftPx:  float64(dyMM) * g.PxPerMm,
		}
		ov.PxPerMm = g.PxPerMm
	}

	minY := 0.0
	maxY := overlayDrawH
	maxCubeHeight := 0.0
	for _, it := range items {
		baseY := maxAnchor - it.anchorY + it.staggerY
		y := baseY + it.shiftPx
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
	ov.BoxW = overlayBoxW
	ov.BoxH = boxH
	ov.GuideY = maxAnchor + topOffset

	birds := make([]overlayBird, len(items))
	for i, it := range items {
		sp := it.sp
		baseY := maxAnchor - it.anchorY + it.staggerY + topOffset
		y := baseY + it.shiftPx
		cube := newMassCube(sp, boxH, it.geom.PxPerMm)
		cube.X = 12
		var zeroURL template.URL
		if it.dyMM != 0 {
			zeroURL = overlaySetOffset(q, allowed, sp.ID, 0)
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
			BaseY:       baseY,
			DX:          it.dx,
			Y:           y,
			SpanPx:      it.geom.YTailTip,
			DyMM:        it.dyMM,
			DyText:      formatOverlayShift(it.dyMM, opt.Metric3),
			UpURL:       overlaySetOffset(q, allowed, sp.ID, it.dyMM-nudgeMM),
			DnURL:       overlaySetOffset(q, allowed, sp.ID, it.dyMM+nudgeMM),
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
