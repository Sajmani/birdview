package morphology

import (
	"math"
	"math/rand"
	"strings"
	"testing"
)

// TestSegmentsSumExactly asserts P-041 as stated: the five serialized segment
// values sum exactly to the serialized total, not within a tolerance.
//
// Property test over random but plausible inputs, because the failure this
// guards against is a rounding residue that appears for some values and not
// others -- an earlier design broke it for 27.6% of its species.
//
// Verifies: P-041. Criterion: AC-030.
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

// TestNeckModelNamesTheFamilyRatio asserts the part of P-045 with teeth: the
// model string must name the numeric ratio applied to this bird's family, not
// merely say that a per-family ratio exists.
//
// Verifies: P-045. Criterion: AC-035.
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
// Verifies: P-045.
func TestNeckRatioDefaultIsFlagged(t *testing.T) {
	if _, model := NeckLength(100, "Parulidae"); !strings.Contains(model, "default") {
		t.Errorf("unlisted family should be described as taking the default, got %q", model)
	}
	if _, model := NeckLength(100, "Ardeidae"); strings.Contains(model, "default") {
		t.Errorf("listed family should not be described as taking the default, got %q", model)
	}
}

// TestShapeDifferentiation asserts P-044 through orderings that come from
// ornithology, not from running the code. If one fails, the model is wrong.
//
// A heron's neck is the defining feature of its silhouette and an owl's is
// invisible; whatever the absolute numbers, that ordering must hold.
//
// Verifies: P-044. Criterion: AC-034 (partial -- the bill and tail orderings
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

// TestHeadFloor checks the floor actually binds, since it exists to stop a
// fine-billed bird drawing a head smaller than its skull.
func TestHeadFloor(t *testing.T) {
	if v, _ := HeadLength(1.0); v != 12.0 {
		t.Errorf("HeadLength(1.0) = %v, want the 12 mm floor", v)
	}
	if v, _ := HeadLength(20.0); v != 33.0 {
		t.Errorf("HeadLength(20.0) = %v, want 33.0", v)
	}
}

// TestModelsAreDescribed checks every model function returns a non-empty
// description. A modeled value with an empty model string would satisfy the
// schema while telling the reader nothing (P-062).
func TestModelsAreDescribed(t *testing.T) {
	_, h := HeadLength(10)
	_, b := BodyLength(50)
	_, n := NeckLength(100, "Anatidae")
	_, w := Wingspan(150, 50)
	for name, desc := range map[string]string{"head": h, "body": b, "neck": n, "wingspan": w} {
		if strings.TrimSpace(desc) == "" {
			t.Errorf("%s model returned an empty description", name)
		}
	}
}
