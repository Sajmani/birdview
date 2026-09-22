package view

import (
	"math"

	"birdview/internal/dataset"
)

// Box is the drawing area available to one diagram, in pixels.
type Box struct{ W, H float64 }

// Geometry is everything the renderer needs to draw one bird, in pixels
// relative to the box origin. Every vertical coordinate derives from the same
// scale, and so does the crossbar (P-043).
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

	// Legs attach at the body's midpoint and rake back toward the tail at 45
	// degrees, so each has equal horizontal and vertical extent. Drawn from the
	// measured tarsus -- the lower leg of a standing bird -- which makes them,
	// with bill and tail, one of the three parts of the diagram that is not
	// modeled.
	YLegTop      float64
	LegDX, LegDY float64
}

// headAspect is how wide the cranium oval is drawn relative to its height.
// A drawing convention with no measured basis, like the neck ratio -- but
// unlike the neck it produces no metric, so it carries no provenance record.
const headAspect = 0.72

// Layout places one bird inside a box at a shared scale.
//
// refLength and refWingspan are the largest total length and wingspan the scale
// must accommodate -- the section's, or the whole set's in global mode (P-042).
// Passing the same references for every bird in a group is what makes them
// comparable; the scale is chosen so the largest of each fits, and every other
// bird is drawn smaller in proportion.
//
// Verifies: P-040, P-041, P-042, P-043.
func Layout(s *dataset.Species, refLength, refWingspan float64, box Box) Geometry {
	scale := 0.0
	if refLength > 0 {
		scale = box.H / refLength
	}
	// The crossbar is drawn at the vertical scale (P-043), so a wingspan of
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
	g.YShoulder = g.YNeckTop + seg.Neck.Value*scale
	g.YBodyEnd = g.YShoulder + seg.Body.Value*scale
	g.YTailTip = g.YBodyEnd + seg.Tail.Value*scale
	g.CrossbarHalfWidth = s.Wingspan.Value * scale / 2

	g.YLegTop = (g.YShoulder + g.YBodyEnd) / 2
	// At 45 degrees the horizontal and vertical components are equal, each the
	// tarsus over root two.
	leg := s.Tarsus.Value * scale / math.Sqrt2
	g.LegDX, g.LegDY = leg, leg
	return g
}

// TotalHeight is the drawn extent from bill tip to tail tip. Because the
// segments sum exactly to total length (P-041), this equals the bird's total
// length at the current scale -- which is what makes two diagrams at one scale
// directly comparable.
func (g Geometry) TotalHeight() float64 { return g.YTailTip - g.YBillTip }
