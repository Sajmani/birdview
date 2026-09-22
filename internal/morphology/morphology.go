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

// HeadLength models the cranium from body mass, isometrically: a linear
// dimension scales as the cube root of mass.
//
// Modeled, and unvalidated -- no source measures avian head length, so the
// coefficient is chosen to put a 20 g passerine at roughly 20 mm and is a shape
// function rather than a fit.
//
// It replaces a model based on bill depth, `max(12 mm, depth x 1.65)`, which
// failed in a way worth remembering: the floor bound for 442 of 841 species and
// 86% of birds under 200 mm, so most of the dataset shared one constant head
// that varied with nothing and consumed a fixed share of every small bird's
// length. Bill depth is in any case a poor predictor of skull length -- unfloored
// it gave a sandpiper a 7 mm head.
//
// There is deliberately no floor. A 2.5 g hummingbird lands near 10 mm, which is
// about right, so a floor would only hide a bad coefficient.
func HeadLength(massG float64) (float64, string) {
	m := math.Max(massG, 1.0)
	v := headCoefficient * math.Cbrt(m)
	return v, fmt.Sprintf("cranium modeled as %.1f x mass^(1/3) (isometric, unvalidated)",
		headCoefficient)
}

// headCoefficient sets the cranium scale. See HeadLength.
const headCoefficient = 7.5

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

// DefaultLegFactor is how much of a bird's visible leg the tarsus accounts for
// in families not listed in LegFactors.
const DefaultLegFactor = 1.1

// LegFactors scale the measured tarsus up to an approximate *visible* leg.
//
// The tarsus is the tarsometatarsus alone. Above it the tibiotarsus is feathered
// over in most perching birds but substantially exposed in waders and cranes, so
// the tarsus understates the leg -- and understates it most in the birds whose
// legs define them, which is the opposite of useful.
//
// **These factors are invented.** No source measures exposed leg length, and
// they are a drawing parameter of exactly the kind P-045 describes for the neck:
// named in the provenance of every value they produce, never presented as a
// measurement. The measured tarsus is carried alongside and shown separately, so
// a reader can always see what was measured and what was inferred from it.
var LegFactors = map[string]float64{
	// Tibia almost wholly exposed.
	"Phoenicopteridae":  2.0,
	"Ardeidae":          1.9,
	"Ciconiidae":        1.9,
	"Gruidae":           1.9,
	"Threskiornithidae": 1.8,
	"Aramidae":          1.8,
	// Long, clearly visible tibia.
	"Recurvirostridae": 1.7,
	"Burhinidae":       1.6,
	"Rallidae":         1.5,
	// Moderate.
	"Scolopacidae":   1.4,
	"Charadriidae":   1.4,
	"Haematopodidae": 1.4,
	"Phasianidae":    1.4,
	"Cracidae":       1.4,
	"Cathartidae":    1.4,
	// Slight.
	"Anatidae":          1.2,
	"Laridae":           1.2,
	"Stercorariidae":    1.2,
	"Podicipedidae":     1.2,
	"Gaviidae":          1.2,
	"Pelecanidae":       1.2,
	"Phalacrocoracidae": 1.2,
	"Odontophoridae":    1.2,
	// Effectively none: the leg is the tarsus and little else.
	"Apodidae":      1.0,
	"Trochilidae":   1.0,
	"Caprimulgidae": 1.0,
	"Alcedinidae":   1.0,
	"Fregatidae":    1.0,
}

// LegLength models the visible leg from the measured tarsus.
//
// The description names the factor, the family and the measurement it was
// applied to, so a reader can recover the tarsus from what is shown (P-062).
func LegLength(tarsusMm float64, family string) (float64, string) {
	k, explicit := LegFactors[family]
	kind := "family default"
	if !explicit {
		k = DefaultLegFactor
	} else {
		kind = "family-specific"
	}
	return tarsusMm * k, fmt.Sprintf(
		"exposed leg modeled as %.1f x the measured tarsus of %.1f mm (%s factor for %s)",
		k, tarsusMm, kind, family)
}
