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

package morphology

import (
	"encoding/json"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// TestSegmentsSumExactly asserts product.segmentsSumToLength as stated: the five serialized segment
// values sum exactly to the serialized total, not within a tolerance.
//
// Property test over random but plausible inputs, because the failure this
// guards against is a rounding residue that appears for some values and not
// others -- an earlier design broke it for 27.6% of its species.
//
// Verifies: product.segmentsSumToLength. Criterion: acceptance.segmentsSumExactly.
func TestSegmentsSumExactly(t *testing.T) {
	r := rand.New(rand.NewSource(1))
	for i := 0; i < 20000; i++ {
		bill := r.Float64() * 200
		head := 12 + r.Float64()*60
		neck := r.Float64() * 300
		body := 10 + r.Float64()*400
		tail := r.Float64() * 350

		b, h, n, bo, tl, total := BalanceToSum(bill, head, neck, body, tail)
		if sum := b + h + n + bo + tl; math.Abs(sum-total) > 1e-9 {
			t.Fatalf("segments %v+%v+%v+%v+%v = %v, total %v (diff %g)",
				b, h, n, bo, tl, sum, total, sum-total)
		}
	}
}

// TestNeckModelNamesTheFamilyRatio asserts the part of product.neckModelPerFamily with teeth: the
// model string must name the numeric ratio applied to this bird's family, not
// merely say that a per-family ratio exists.
//
// Verifies: product.neckModelPerFamily. Criterion: acceptance.neckLengthSerialized.
func TestNeckModelNamesTheFamilyRatio(t *testing.T) {
	cases := []struct {
		family string
		want   string // the ratio, as it must appear in the model string
	}{
		{"Ardeidae", "0.85"},
		{"Strigidae", "0.08"},
		{"Scolopacidae", "0.35"},
		{"Parulidae", "0.18"}, // not listed -- default
	}
	for _, c := range cases {
		_, model := NeckLength(100, c.family)
		if !strings.Contains(model, c.want) {
			t.Errorf("%s: model %q does not name the ratio %s", c.family, model, c.want)
		}
		if !strings.Contains(model, c.family) {
			t.Errorf("%s: model %q does not name the family", c.family, model)
		}
	}
}

// TestNeckRatioDefaultIsFlagged checks that an unlisted family is described as
// taking the default rather than being presented as a considered choice.
//
// Verifies: product.neckModelPerFamily.
func TestNeckRatioDefaultIsFlagged(t *testing.T) {
	if _, model := NeckLength(100, "Parulidae"); !strings.Contains(model, "default") {
		t.Errorf("unlisted family should be described as taking the default, got %q", model)
	}
	if _, model := NeckLength(100, "Ardeidae"); strings.Contains(model, "default") {
		t.Errorf("listed family should not be described as taking the default, got %q", model)
	}
}

// TestShapeDifferentiation asserts product.proportionDifferentiates through orderings that come from
// ornithology, not from running the code. If one fails, the model is wrong.
//
// A heron's neck is the defining feature of its silhouette and an owl's is
// invisible; whatever the absolute numbers, that ordering must hold.
//
// Verifies: product.proportionDifferentiates. Criterion: acceptance.proportionsDifferentiate (partial -- the bill and tail orderings
// live in the ingest test, where real trait data is available).
func TestShapeDifferentiation(t *testing.T) {
	const body = 100
	heron, _ := NeckLength(body, "Ardeidae")
	sparrow, _ := NeckLength(body, "Passerellidae")
	owl, _ := NeckLength(body, "Strigidae")
	sandpiper, _ := NeckLength(body, "Scolopacidae")

	if !(heron > sandpiper && sandpiper > sparrow && sparrow > owl) {
		t.Errorf("neck ordering wrong: heron %.1f, sandpiper %.1f, sparrow %.1f, owl %.1f; "+
			"want heron > sandpiper > sparrow > owl", heron, sandpiper, sparrow, owl)
	}
}

// TestHeadScalesWithMass asserts the cranium is a function of the bird rather
// than a constant.
//
// The previous model, max(12 mm, bill depth x 1.65), returned exactly 12 mm for
// 442 of 841 species and for 86% of birds under 200 mm: half the dataset shared
// one head that varied with nothing. Isometric scaling has no floor, so a
// regression to a constant shows up as a flat curve here.
//
// Verifies: product.provenancePerMetric. Criterion: acceptance.craniumVaries.
func TestHeadScalesWithMass(t *testing.T) {
	// Expected values follow from the model definition, 7.5 * mass^(1/3), not
	// from running it: 8 g gives 7.5*2 = 15, 1000 g gives 7.5*10 = 75.
	for _, c := range []struct{ mass, want float64 }{{8, 15}, {1000, 75}, {27, 22.5}} {
		if got, _ := HeadLength(c.mass); math.Abs(got-c.want) > 0.05 {
			t.Errorf("HeadLength(%g) = %.2f, want %.2f", c.mass, got, c.want)
		}
	}
	// Strictly increasing across the real mass range, with no plateau.
	prev := 0.0
	for _, m := range []float64{2.5, 5, 10, 20, 50, 100, 500, 2000, 11000} {
		got, _ := HeadLength(m)
		if got <= prev {
			t.Errorf("head did not grow between the previous mass and %g g (%.1f then %.1f)",
				m, prev, got)
		}
		prev = got
	}
}

// Böhmer et al. 2019 measured 103 skeletons across 34 orders. Regressing log
// dimension on log mass gives, for the two MULTI-ELEMENT dimensions it reports:
//
//	neck length       0.380, 95% CI 0.338 - 0.421
//	total leg length  0.364, 95% CI 0.328 - 0.401
//
// product.trunkFromMass requires the trunk exponent to lie in the intersection. These bounds
// come from the published data, not from this package -- do not "fix" a failure
// by widening them.
const (
	minTrunkExponent = 0.338
	maxTrunkExponent = 0.401
)

// supersededExponent and supersededCoefficient are the invented constants
// decisions.trunkExponentCalibrated replaced. Kept because two assertions below are about them: that the
// calibrated band excludes the old exponent, and that the new curve still
// passes through the old one at the anchor.
const (
	supersededExponent    = 0.31
	supersededCoefficient = 18.0
)

// TestTrunkScalesLikeMeasuredSkeletons recovers the trunk model's own exponent
// by regression and checks it against an interval taken from measurement.
//
// The point of fitting rather than reading the constant is that this tests the
// function's behavior: a model that stopped being a clean power law, or that
// clamped, would fail here while a direct comparison against trunkExponent
// would still pass.
//
// Verifies: product.trunkFromMass. Criterion: acceptance.trunkScalesLikeSkeletons.
func TestTrunkScalesLikeMeasuredSkeletons(t *testing.T) {
	// The band must actually discriminate. A criterion whose bounds admit the
	// value it replaced cannot fail, which is the failure mode this project
	// keeps rediscovering -- so assert the band's sharpness before using it.
	if supersededExponent >= minTrunkExponent {
		t.Fatalf("the calibrated band [%g, %g] admits the superseded exponent %g, so it cannot discriminate",
			minTrunkExponent, maxTrunkExponent, supersededExponent)
	}
	if exaggerated := 0.45; exaggerated <= maxTrunkExponent {
		t.Fatalf("the calibrated band [%g, %g] admits %g, so it is too loose to catch an overcorrection",
			minTrunkExponent, maxTrunkExponent, exaggerated)
	}

	// Recover the exponent over the real ABA mass range: Calliope Hummingbird
	// at 2.6 g to Trumpeter Swan at 11071 g.
	var xs, ys []float64
	for i := 0; i <= 200; i++ {
		lm := math.Log(2.6) + (math.Log(11071)-math.Log(2.6))*float64(i)/200
		mass := math.Exp(lm)
		trunk, _ := BodyLength(mass)
		xs = append(xs, lm)
		ys = append(ys, math.Log(trunk))
	}
	got := slope(xs, ys)
	if got < minTrunkExponent || got > maxTrunkExponent {
		t.Errorf("trunk scales as mass^%.4f, outside the measured interval [%g, %g] product.trunkFromMass requires",
			got, minTrunkExponent, maxTrunkExponent)
	}

	// The anchor pins the coefficient, which no scaling assertion can see: the
	// right exponent with a wrong coefficient draws every bird wrong while
	// scaling perfectly.
	wantAtAnchor := supersededCoefficient * math.Pow(anchorMassG, supersededExponent)
	if gotAtAnchor, _ := BodyLength(anchorMassG); math.Abs(gotAtAnchor-wantAtAnchor) > 0.01 {
		t.Errorf("at the %g g anchor the trunk is %.3f mm, want %.3f mm — the recalibration is "+
			"supposed to change proportion, not overall scale",
			anchorMassG, gotAtAnchor, wantAtAnchor)
	}

	// Direction: lighter birds shrink, heavier birds grow, monotonically.
	prev := 0.0
	for _, mass := range []float64{2.6, 10, 30, anchorMassG, 150, 800, 4000, 11071} {
		now, _ := BodyLength(mass)
		was := supersededCoefficient * math.Pow(mass, supersededExponent)
		ratio := now / was
		if ratio <= prev {
			t.Errorf("at %g g the new/old trunk ratio is %.4f, not greater than the previous %.4f; "+
				"the correction must grow with mass", mass, ratio, prev)
		}
		switch {
		case mass < anchorMassG && ratio >= 1:
			t.Errorf("at %g g, below the anchor, the trunk grew (ratio %.4f)", mass, ratio)
		case mass > anchorMassG && ratio <= 1:
			t.Errorf("at %g g, above the anchor, the trunk shrank (ratio %.4f)", mass, ratio)
		}
		prev = ratio
	}
}

// slope returns the ordinary least-squares slope of y on x.
func slope(x, y []float64) float64 {
	n := float64(len(x))
	var mx, my float64
	for i := range x {
		mx += x[i] / n
		my += y[i] / n
	}
	var sxy, sxx float64
	for i := range x {
		sxy += (x[i] - mx) * (y[i] - my)
		sxx += (x[i] - mx) * (x[i] - mx)
	}
	return sxy / sxx
}

// TestModelsAreDescribed checks every model function returns a non-empty
// description. A modeled value with an empty model string would satisfy the
// schema while telling the reader nothing (product.provenancePerMetric).
func TestModelsAreDescribed(t *testing.T) {
	_, h := HeadLength(20)
	_, b := BodyLength(50)
	_, n := NeckLength(100, "Anatidae")
	for name, desc := range map[string]string{"head": h, "body": b, "neck": n} {
		if strings.TrimSpace(desc) == "" {
			t.Errorf("%s model returned an empty description", name)
		}
	}
}

// Every non-passerine family must be either given a neck ratio or explicitly
// recorded as reviewed and left on the default.
//
// The table used to say the quiet part itself: "only a review of the table
// against the full family list will catch the next one". It was right, and
// it caught nothing -- Stercorariidae drew half a gull's neck for however
// long, and Cathartidae was never looked at after AviList gave the New World
// vultures an order of their own. Absence from NeckRatios meant two
// different things, "considered and unremarkable" and "never considered",
// and nothing could tell them apart.
//
// Passeriformes is exempt as an order; see NeckDefaultReviewed.
//
// Watched fail: deleting Stercorariidae from NeckRatios reports it as
// unreviewed; so does removing Procellariidae from NeckDefaultReviewed.
//
// Verifies: product.neckModelPerFamily. Criterion: acceptance.neckRatiosReviewed.
func TestEveryFamilyHasBeenReviewed(t *testing.T) {
	ds := realFamilies(t)
	var missing []string
	for fam, order := range ds {
		if order == "Passeriformes" {
			continue
		}
		if _, ok := NeckRatios[fam]; ok {
			continue
		}
		if _, ok := NeckDefaultReviewed[fam]; ok {
			continue
		}
		missing = append(missing, fam+" ("+order+")")
	}
	sort.Strings(missing)
	if len(missing) > 0 {
		t.Errorf("%d families have neither a neck ratio nor a recorded review, so they "+
			"draw 0.18 because nobody looked:\n  %s", len(missing), strings.Join(missing, "\n  "))
	}

	// A family cannot be in both: the two maps mean opposite things.
	for fam := range NeckDefaultReviewed {
		if _, ok := NeckRatios[fam]; ok {
			t.Errorf("%s is both given a ratio and recorded as left on the default", fam)
		}
	}
	// Every reviewed entry must carry a reason, or it records nothing.
	for fam, why := range NeckDefaultReviewed {
		if strings.TrimSpace(why) == "" {
			t.Errorf("%s is recorded as reviewed with no reason given", fam)
		}
	}
}

// realFamilies maps family to order from the compiled dataset, or skips.
func realFamilies(t *testing.T) map[string]string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("..", "..", "web", "data", "species.json"))
	if err != nil {
		t.Skipf("no compiled dataset; run birdview-ingest first")
	}
	var d struct {
		Species []struct{ Family, Order string } `json:"species"`
	}
	if err := json.Unmarshal(b, &d); err != nil {
		t.Fatalf("reading the dataset: %v", err)
	}
	out := map[string]string{}
	for _, s := range d.Species {
		out[s.Family] = s.Order
	}
	return out
}
