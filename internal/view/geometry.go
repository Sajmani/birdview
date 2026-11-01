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

package view

import (
	"math"

	"birdview/internal/dataset"
)

// Box is the drawing area available to one diagram, in pixels.
type Box struct{ W, H float64 }

// Geometry is everything the renderer needs to draw one bird, in pixels
// relative to the box origin. Every vertical coordinate derives from the same
// scale, and so does the crossbar (product.crossbarAtVerticalScale).
type Geometry struct {
	PxPerMm float64

	CX                float64 // the vertical axis
	YBillTip          float64
	YCraniumTop       float64 // bill meets head
	YHeadCenter       float64
	HeadRx, HeadRy    float64
	YNeckTop          float64 // head meets neck
	YShoulder         float64 // neck meets body; the crossbar sits here
	YBodyEnd          float64 // body meets tail
	YTailTip          float64
	CrossbarHalfWidth float64

	// ChordTickX is how far from the axis the measured wing chord begins, in
	// px: CrossbarHalfWidth minus the scaled chord. Zero when unknown.
	//
	// The chord is the ORNITHOLOGICAL one -- carpal joint
	// to primary tip, along the wing -- so it occupies the OUTER part of each
	// half-span and the tick marks its inner end. It is not a boundary between
	// a measured and a modeled portion of the span: where the span is modeled
	// it is modeled in full, from this chord, multiplicatively. The tick says
	// "the measurement reaches to here", nothing more.
	ChordTickX float64

	// NeckFolded is set for the families that fly with the neck drawn back onto
	// the shoulders (product.segmentComposition). Their neck is an S of two semicircles of NeckArcR
	// rather than a straight stick, so it is the same length of stroke in less
	// height. NeckArcR is zero when the neck is straight.
	NeckFolded bool
	NeckArcR   float64
}

// headAspect is how wide the cranium oval is drawn relative to its height.
// A drawing convention with no measured basis, like the neck ratio -- but
// unlike the neck it produces no metric, so it carries no provenance record.
const headAspect = 0.72

// foldedNecks are the families that fly with the neck folded back rather than
// extended -- the field mark that separates a heron from a crane, a stork, or an
// ibis in the air (product.segmentComposition, decisions.foldedNeckAsS).
//
// This is a posture, not a measurement, and it produces no metric: the neck
// length is unchanged and is still reported in full. So like headAspect it
// carries no provenance record.
//
// Everything absent here flies neck-extended and is correct as a stick. That
// includes every other long-necked family -- Gruidae, Ciconiidae,
// Threskiornithidae, Phoenicopteridae, Anatidae, Gaviidae, Podicipedidae,
// Phalacrocoracidae, Anhingidae. Gulls and shorebirds draw the head in somewhat
// but do not fold, and their modeled necks are short enough that a stick is not
// misleading.
var foldedNecks = map[string]string{
	"Ardeidae":    "Herons",
	"Pelecanidae": "Pelicans",
}

// FoldedNeckGroup names the birds of a family that fly with the neck folded, for prose such as
// "Herons fly with the neck folded back", or returns "" for a family drawn with a straight neck.
func FoldedNeckGroup(family string) string { return foldedNecks[family] }

// neckFoldCompression is how much of its own length a folded neck spans
// vertically: two semicircles of radius r run 2*pi*r of stroke through 4r of
// height, so the ratio is 2/pi and nothing was chosen to make it so.
//
// Declared for the tests and the documentation to share. Layout does not use it
// -- it derives the radius from the length, which is the same statement the
// other way round.
const neckFoldCompression = 2 / math.Pi

// Layout places one bird inside a box at a shared scale.
//
// refLength and refWingspan are the largest total length and wingspan the
// scale must accommodate -- the section's, or the whole set's in global mode (product.scaleToggle).
// Passing the same references for every bird in a group is what makes them
// comparable; the scale is chosen so the largest of each fits, and every other
// bird is drawn smaller in proportion.
//
// Verifies: product.segmentComposition, product.segmentsSumToLength, product.scaleToggle, product.crossbarAtVerticalScale.
func Layout(s *dataset.Species, refLength, refWingspan float64, box Box) Geometry {
	scale := 0.0
	if refLength > 0 {
		scale = box.H / refLength
	}
	// The crossbar is drawn at the vertical scale (product.crossbarAtVerticalScale), so a wingspan of
	// 2.5x body length needs 2.5x the height in width. Where the box cannot
	// give that, the whole diagram shrinks rather than the crossbar alone --
	// rescaling one axis would silently misstate the proportion.
	if refWingspan > 0 {
		if w := box.W / refWingspan; w < scale || scale == 0 {
			scale = w
		}
	}

	seg := s.Segments
	g := Geometry{PxPerMm: scale, CX: box.W / 2}
	g.YBillTip = 0
	g.YCraniumTop = g.YBillTip + seg.Bill.Value*scale
	headH := seg.Head.Value * scale
	g.HeadRy = headH / 2
	g.HeadRx = g.HeadRy * headAspect
	g.YHeadCenter = g.YCraniumTop + g.HeadRy
	g.YNeckTop = g.YCraniumTop + headH

	// A folded neck is the same stroke in less height: two semicircles whose
	// combined path length is exactly the neck, so r = neck/2pi and the pair
	// spans 4r (product.segmentComposition).
	neck := seg.Neck.Value * scale
	if g.NeckFolded = foldedNecks[s.Family] != ""; g.NeckFolded {
		g.NeckArcR = neck / (2 * math.Pi)
		g.YShoulder = g.YNeckTop + 4*g.NeckArcR
	} else {
		g.YShoulder = g.YNeckTop + neck
	}

	g.YBodyEnd = g.YShoulder + seg.Body.Value*scale
	g.YTailTip = g.YBodyEnd + seg.Tail.Value*scale
	g.CrossbarHalfWidth = s.Wingspan.Value * scale / 2
	if c := s.WingChord.Value * scale; c > 0 && c < g.CrossbarHalfWidth {
		g.ChordTickX = g.CrossbarHalfWidth - c
	}
	return g
}

// NeckMid is where a folded neck's S changes direction: the join between its
// two semicircles. Meaningless when the neck is straight.
func (g Geometry) NeckMid() float64 { return (g.YNeckTop + g.YShoulder) / 2 }

// TotalHeight is the drawn extent from bill tip to tail tip. Because the
// segments sum exactly to total length (product.segmentsSumToLength), this equals the bird's total
// length at the current scale -- which is what makes two diagrams at one scale
// directly comparable.
//
// Except when the neck is folded, where the S spans less height than the neck
// is long and the drawing is correspondingly shorter (decisions.foldedNeckAsS). Length as data
// is unaffected: Sections and Scales read TotalLength, never this.
func (g Geometry) TotalHeight() float64 { return g.YTailTip - g.YBillTip }
