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
//
// Two were missing and drew the default until a reader noticed a short-necked
// Limpkin. Aramidae flies neck-extended and is built like a small crane, and
// Pelecanidae was worse: it was listed as a family that FOLDS its neck while
// having no neck to fold, so a pelican drew a 5%-of-length stub tucked into an
// S. view.TestEveryFoldedFamilyHasALongNeck now makes the second class of that
// mistake impossible; the first was caught by eye, and only a review of the
// table against the family list will catch the next one.
var NeckRatios = map[string]float64{
	// Long-necked waders and waterfowl.
	"Ardeidae":          0.85,
	"Ciconiidae":        0.80,
	"Threskiornithidae": 0.75,
	"Gruidae":           0.85,
	"Anhingidae":        0.90,
	"Phalacrocoracidae": 0.55,
	// Gannets and boobies fly the "cigar" silhouette: the neck is genuinely
	// extended ahead of the wings, but less snaky than a cormorant's. Placed
	// between Laridae and Phalacrocoracidae, the two neighbours already here.
	"Sulidae":          0.45,
	"Anatidae":         0.70,
	"Podicipedidae":    0.45,
	"Gaviidae":         0.50,
	"Phoenicopteridae": 0.95,
	"Pelecanidae":      0.75,
	"Aramidae":         0.70,
	// Shorebirds, gulls and rails: intermediate.
	"Scolopacidae":     0.35,
	"Charadriidae":     0.30,
	"Recurvirostridae": 0.45,
	"Laridae":          0.35,
	// Jaegers and skuas are gulls in build, in the same order, and were
	// taking the default: a 474 mm Pomarine Jaeger drew a 28.4 mm neck where
	// a 474 mm California Gull drew 55.0 mm. That gap was an omission from
	// this table, not a judgement about birds, so this matches Laridae.
	"Stercorariidae": 0.35,
	"Rallidae":       0.35,
	"Burhinidae":     0.35,
	"Haematopodidae": 0.30,
	// Compact: the neck is barely exposed.
	"Strigidae":   0.08,
	"Tytonidae":   0.08,
	"Trochilidae": 0.12,
	"Alcedinidae": 0.10,
	"Picidae":     0.14,
	// Auks are the stubby extreme of Charadriiformes -- "no neck" is the
	// field impression of a murre going past -- which puts them with the
	// kingfishers rather than with their own order's gulls.
	"Alcidae": 0.10,
	// New World vultures soar with the small bare head drawn back against
	// the shoulders, so almost nothing projects. AviList gave them an order
	// of their own (CR-022) and they were among the families never reviewed
	// against this table.
	"Cathartidae": 0.10,

	// Seabirds that fly neck-extended, between the compact group above and
	// the waders at the top.
	//
	// A frigatebird's long slender neck is part of what identifies it in the
	// air, just below the gannet's cigar. An albatross projects its head and
	// neck well ahead of the wings -- more than a shearwater, less than a
	// gannet -- and Procellariidae stays on the default because a shearwater
	// does not.
	"Fregatidae":  0.40,
	"Diomedeidae": 0.30,
}

// NeckDefaultReviewed names the families looked at and left on
// DefaultNeckRatio, so "not in NeckRatios" stops meaning two different things
// at once: considered and found unremarkable, or never considered.
//
// This exists because the table's own comment admitted the gap and could only
// hope someone would close it -- "only a review of the table against the full
// family list will catch the next one". TestEveryFamilyHasBeenReviewed turns
// that hope into a build failure: a family in neither map fails, so the next
// one AviList adds has to be looked at rather than quietly drawing 0.18.
//
// **Passeriformes is exempt as an order**, and that is not laziness. All 43
// of its families here are small compact songbirds with no exposed neck, the
// default is a short stub, and writing the same sentence 43 times would bury
// the families where the question is live. The exemption is the claim being
// made: a new passerine family needs no review, a new family in any other
// order does.
//
// Verifies: P-045.
var NeckDefaultReviewed = map[string]string{
	// Landbirds with no exposed neck to speak of. The default is a short
	// stub, which is what these are.
	"Accipitridae":   "head projects modestly in flight; the default is right",
	"Pandionidae":    "built as an Accipitridae for this purpose",
	"Falconidae":     "compact; the head barely clears the shoulders",
	"Apodidae":       "no exposed neck at all",
	"Caprimulgidae":  "no exposed neck at all",
	"Columbidae":     "compact",
	"Cuculidae":      "compact",
	"Psittacidae":    "compact",
	"Psittaculidae":  "compact",
	"Pteroclidae":    "compact",
	"Trogonidae":     "compact",
	"Cracidae":       "gamebird; short neck, and the tail carries the silhouette",
	"Odontophoridae": "gamebird; short neck",
	"Phasianidae":    "gamebird; short neck",
	"Phaethontidae":  "moderate, between a gull and a shearwater; the tail streamers identify it, not the neck",
	"Procellariidae": "shearwaters and petrels fly head-close; unlike the albatrosses",
	"Hydrobatidae":   "too small for a neck to read at any scale this project draws",
	"Oceanitidae":    "too small for a neck to read at any scale this project draws",
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
//
// The exponent is NOT a free parameter. It is calibrated against measured
// skeletons and must stay inside the interval P-047 names; see trunkExponent.
//
// Verifies: P-047.
func BodyLength(massG float64) (float64, string) {
	m := math.Max(massG, 1.0)
	v := trunkCoefficient * math.Pow(m, trunkExponent)
	return v, fmt.Sprintf("trunk modeled as %g x mass^%g (allometric, calibrated against skeletal data)",
		trunkCoefficient, trunkExponent)
}

// trunkExponent is how fast the trunk grows with mass.
//
// It replaces an invented 0.31 that was too shallow, understating the large end
// of the ABA scope by about 1.65x (CR-021). Three independent measurements of
// how a linear dimension scales with mass agree far above it: measured wingspan
// 0.368 (n=301, Fox), neck length 0.380 and total leg length 0.364 (n=103,
// Bohmer et al. 2019). 0.37 is the round value nearest the mean of the two that
// carry no aerodynamic confound.
//
// The distinction that makes this evidence usable: SINGLE bones scale near
// isometry -- femur 0.325, central cervical vertebra 0.314 -- while CHAINS of
// them scale higher. A trunk is a chain. This is also why the cranium of P-046
// keeps its cube root and was deliberately left alone.
//
// P-047 requires this to lie in [0.338, 0.401]. Do not retune it by eye; that
// is what produced the value it replaced.
const trunkExponent = 0.37

// trunkCoefficient follows from trunkExponent and is not independently chosen.
//
// It is set so the model is unchanged at anchorMassG, which keeps the
// recalibration a change of proportion across the size range rather than a
// change of overall scale: the median bird is drawn as it was, small birds a
// little shorter, large birds substantially longer.
//
// Derivation: 18.0 * 63.4^0.31 / 63.4^0.37 = 14.0329. Carried to four decimals
// because AC-080 checks the anchor to 0.01 mm and 14.03 misses it by 0.013.
const trunkCoefficient = 14.0329

// anchorMassG is the mass at which the recalibration of CR-021 is a no-op: the
// median of the 841 ABA species when the exponent was calibrated.
//
// It is a FIXED CONSTANT, deliberately not recomputed from the dataset at build
// time. If scope changes the median moves, and the whole drawing must not move
// with it silently -- re-anchoring is a decision, and P-047 says it has to be
// recorded as one.
const anchorMassG = 63.4

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
