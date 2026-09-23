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

	// NeckFolded is set for the families that fly with the neck drawn back onto
	// the shoulders (P-040). Their neck is an S of two semicircles of NeckArcR
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
// extended -- the field mark that separates a heron from a crane, a stork or an
// ibis in the air (P-040, CR-011).
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
var foldedNecks = map[string]bool{
	"Ardeidae":    true,
	"Pelecanidae": true,
}

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

	// A folded neck is the same stroke in less height: two semicircles whose
	// combined path length is exactly the neck, so r = neck/2pi and the pair
	// spans 4r (P-040).
	neck := seg.Neck.Value * scale
	if g.NeckFolded = foldedNecks[s.Family]; g.NeckFolded {
		g.NeckArcR = neck / (2 * math.Pi)
		g.YShoulder = g.YNeckTop + 4*g.NeckArcR
	} else {
		g.YShoulder = g.YNeckTop + neck
	}

	g.YBodyEnd = g.YShoulder + seg.Body.Value*scale
	g.YTailTip = g.YBodyEnd + seg.Tail.Value*scale
	g.CrossbarHalfWidth = s.Wingspan.Value * scale / 2
	return g
}

// NeckMid is where a folded neck's S changes direction: the join between its
// two semicircles. Meaningless when the neck is straight.
func (g Geometry) NeckMid() float64 { return (g.YNeckTop + g.YShoulder) / 2 }

// TotalHeight is the drawn extent from bill tip to tail tip. Because the
// segments sum exactly to total length (P-041), this equals the bird's total
// length at the current scale -- which is what makes two diagrams at one scale
// directly comparable.
//
// Except when the neck is folded, where the S spans less height than the neck
// is long and the drawing is correspondingly shorter (CR-011). Length as data
// is unaffected: Sections and Scales read TotalLength, never this.
func (g Geometry) TotalHeight() float64 { return g.YTailTip - g.YBillTip }
