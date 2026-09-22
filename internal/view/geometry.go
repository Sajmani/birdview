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

	// Legs attach at the body's midpoint and rake back toward the tail at
	// legRake. They are drawn from Species.Leg -- the modeled exposed leg, a
	// per-family multiple of the measured tarsus -- so unlike bill and tail
	// they carry a model, and the detail view shows the tarsus beside them.
	//
	// On a long-legged bird the feet reach below the tail tip, so DrawnHeight
	// rather than TotalHeight is what the renderer must make room for.
	YLegTop      float64
	LegDX, LegDY float64
}

// legRake is how steeply the legs rake back from the body, measured from the
// horizontal. The line is always exactly the modeled leg long; the rake only
// decides how that length is split between vertical and horizontal.
//
// 45 degrees, tried first, spends the same on each and makes a correct-length
// leg read short: a Black-necked Stilt dropped 10 px against a 42 px body. A
// standing bird's leg is near-vertical anyway, so a steeper rake is both more
// accurate and more legible. At 60 degrees the vertical share rises from 0.71
// to 0.87 of the leg, a 22% gain for nothing.
const legRake = 60 * math.Pi / 180

// legSin is the vertical share of a leg at legRake, precomputed so that
// VerticalExtent and Layout cannot drift apart.
var legSin = math.Sin(legRake)

// LegCos is the horizontal share of a leg at legRake. Exported so the sideways
// reach of a pair of legs can be checked against the wingspan that sets the
// tile width -- see TestLegIsModeledFromTheMeasuredTarsus.
var LegCos = math.Cos(legRake)

// headAspect is how wide the cranium oval is drawn relative to its height.
// A drawing convention with no measured basis, like the neck ratio -- but
// unlike the neck it produces no metric, so it carries no provenance record.
const headAspect = 0.72

// Layout places one bird inside a box at a shared scale.
//
// refLength and refWingspan are the largest vertical extent and wingspan the
// scale must accommodate -- the section's, or the whole set's in global mode (P-042).
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
	leg := s.Leg.Value * scale
	g.LegDX = leg * math.Cos(legRake)
	g.LegDY = leg * math.Sin(legRake)
	return g
}

// DrawnHeight is how far the drawing actually reaches, feet included.
//
// On a long-legged bird the feet fall below the tail tip -- 11 of 841 species,
// by up to 30% of their own height for an American Flamingo -- which is correct
// anatomy, so the renderer accommodates it rather than the model backing off.
func (g Geometry) DrawnHeight() float64 {
	if feet := g.YLegTop + g.LegDY; feet > g.YTailTip {
		return feet - g.YBillTip
	}
	return g.TotalHeight()
}

// TotalHeight is the drawn extent from bill tip to tail tip. Because the
// segments sum exactly to total length (P-041), this equals the bird's total
// length at the current scale -- which is what makes two diagrams at one scale
// directly comparable.
func (g Geometry) TotalHeight() float64 { return g.YTailTip - g.YBillTip }
