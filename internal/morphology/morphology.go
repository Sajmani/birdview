// Package morphology turns AVONET traits into the five diagram segments.
//
// Two of the five are measured; three are modeled. Every modeled value names
// the function that produced it, which is what makes presenting them
// defensible (P-062, P-045, CR-007).
//
// None of the models below is a published relation validated against real
// birds. They are documented shape functions, and the provenance strings say
// so. Do not quietly promote one to "measured".
//
// Verifies: P-040, P-041, P-045, P-062.
package morphology

import (
	"fmt"
	"math"
)

// DefaultNeckRatio applies to any family not in NeckRatios: the neck occupies
// this share of the modeled body axis.
const DefaultNeckRatio = 0.18

// NeckRatios is the per-family neck model of P-045. It is a drawing parameter
// with no measured basis -- no source measures avian neck length at this
// coverage -- and it is surfaced to the user with this value named, rather than
// hidden (CR-007).
//
// Families absent here get DefaultNeckRatio. The entries are the ones where
// neck length is the visually defining feature of the family's silhouette.
var NeckRatios = map[string]float64{
	// Long-necked waders and waterfowl.
	"Ardeidae":          0.85,
	"Ciconiidae":        0.80,
	"Threskiornithidae": 0.75,
	"Gruidae":           0.85,
	"Anhingidae":        0.90,
	"Phalacrocoracidae": 0.55,
	"Anatidae":          0.70,
	"Podicipedidae":     0.45,
	"Gaviidae":          0.50,
	"Phoenicopteridae":  0.95,
	// Shorebirds, gulls and rails: intermediate.
	"Scolopacidae":     0.35,
	"Charadriidae":     0.30,
	"Recurvirostridae": 0.45,
	"Laridae":          0.35,
	"Rallidae":         0.35,
	"Burhinidae":       0.35,
	"Haematopodidae":   0.30,
	// Compact: the neck is barely exposed.
	"Strigidae":   0.08,
	"Tytonidae":   0.08,
	"Trochilidae": 0.12,
	"Alcedinidae": 0.10,
	"Picidae":     0.14,
}

// NeckRatio returns the ratio for a family and whether it was explicitly
// listed. The caller needs the second value so the provenance string can say
// whether a family-specific model or the default was applied.
func NeckRatio(family string) (float64, bool) {
	r, ok := NeckRatios[family]
	if !ok {
		return DefaultNeckRatio, false
	}
	return r, true
}

// HeadLength models the cranial vault from bill depth at the nares.
//
// Modeled. The floor keeps very fine-billed birds from drawing a head smaller
// than their own skull.
func HeadLength(billDepthMm float64) (float64, string) {
	v := math.Max(12.0, billDepthMm*1.65)
	return v, "cranial vault modeled as max(12 mm, bill depth x 1.65)"
}

// BodyLength models the trunk from body mass by an allometric relation.
//
// Modeled. Mass is the only whole-animal measurement AVONET carries, which is
// why the trunk is derived from it rather than from a linear trait.
func BodyLength(massG float64) (float64, string) {
	m := math.Max(massG, 1.0)
	v := 18.0 * math.Pow(m, 0.31)
	return v, "trunk modeled as 18 x mass^0.31 (allometric, unvalidated)"
}

// NeckLength models the exposed neck as a share of the trunk.
//
// Modeled. The returned description names the family and the exact ratio, as
// P-045 requires -- a model string naming a function but not its value would
// tell the reader nothing they could check.
func NeckLength(bodyMm float64, family string) (float64, string) {
	ratio, explicit := NeckRatio(family)
	kind := "family default"
	if explicit {
		kind = "family-specific"
	}
	return bodyMm * ratio,
		fmt.Sprintf("neck modeled as %.2f x trunk (%s ratio for %s)", ratio, kind, family)
}

// Wingspan models span from the folded wing chord plus a mass-derived chest
// breadth.
//
// Modeled, and used only where no empirical wingspan exists. Wing chord is the
// folded wing from carpal flexure to the longest primary -- roughly 40% of a
// span -- so the doubling plus a body allowance is the shape of the estimate.
func Wingspan(wingChordMm, massG float64) (float64, string) {
	chest := 10.5 * math.Pow(math.Max(massG, 1.0), 0.33)
	v := 2.0*wingChordMm*1.15 + chest
	return v, "span modeled as 2 x wing chord x 1.15 + 10.5 x mass^0.33 (unvalidated)"
}

// Round1 rounds to 0.1 mm, the resolution BirdView presents.
func Round1(v float64) float64 { return math.Round(v*10) / 10 }

// BalanceToSum adjusts body so the five rounded segments sum exactly to the
// rounded total, which P-041 requires and AC-030 asserts without tolerance.
//
// Rounding each segment independently leaves the sum up to 0.2 mm adrift; an
// earlier design did exactly that and broke the stated invariant for 27.6% of
// its species. Body absorbs the residue because it is the largest modeled
// segment, so the adjustment is proportionally smallest there.
func BalanceToSum(bill, head, neck, body, tail float64) (b, h, n, bo, t, total float64) {
	total = Round1(bill + head + neck + body + tail)
	b, h, n, t = Round1(bill), Round1(head), Round1(neck), Round1(tail)
	bo = Round1(total - b - h - n - t)
	return
}
