package main

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"strings"
	"testing"

	"birdview/internal/dataset"
	"birdview/internal/morphology"
	"birdview/internal/taxonomy"
	"birdview/internal/view"
)

const (
	abaFile = "../../fetched/ABA_Checklist-8.19.csv"
	avoFile = "../../AVONET Supplementary dataset 1.xlsx"
	foxFile = "../../wingspan_fox2026_AppendixS1.xlsx"
)

// built caches the compiled dataset across tests: the AVONET workbook is 21 MB
// and parsing it three times per test would dominate the suite.
var built *dataset.Dataset

func load(t *testing.T) *dataset.Dataset {
	t.Helper()
	if built != nil {
		return built
	}
	if _, err := os.Stat(abaFile); err != nil {
		// The ABA checklist is all rights reserved and never committed (T-004),
		// so a clean checkout legitimately lacks it. Skip loudly rather than
		// pass silently.
		t.Skip("skipping: fetched/ABA_Checklist-8.19.csv is absent; run the fetch step first")
	}
	ds, _, err := build(abaFile, avoFile, foxFile)
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	built = ds
	return ds
}

// Verifies: P-001. Criterion: AC-010.
func TestScopeIsABACodes1To3(t *testing.T) {
	ds := load(t)
	// 841 is counted from ABA Checklist 8.19 itself -- rarity codes 1, 2 and 3
	// number 490, 258 and 93 -- not from a previous run of this code.
	const want = 841
	if got := len(ds.Species); got != want {
		t.Errorf("got %d species, want %d for ABA checklist 8.19", got, want)
	}
}

// TestRarityCodeNeverReachesTheDataset checks the ABA rarity code is used as a
// scope filter and then discarded (P-001).
//
// This inspects field NAMES rather than searching the whole record for "code",
// which an earlier version did. That was too broad: it would now reject the
// four-letter banding code, which is a legitimate part of a species' identity
// and which P-004 requires. The requirement is about rarity specifically.
//
// Verifies: P-001. Criterion: AC-011.
func TestRarityCodeNeverReachesTheDataset(t *testing.T) {
	ds := load(t)
	var fields map[string]any
	if err := json.Unmarshal([]byte(mustJSON(t, ds.Species[0])), &fields); err != nil {
		t.Fatal(err)
	}
	for name := range fields {
		if strings.Contains(strings.ToLower(name), "rarity") {
			t.Errorf("serialized record has field %q; the rarity code is a scope filter, not data", name)
		}
	}
	// And no field should hold a bare 1-6 that could be one.
	if v, ok := fields["abaCode"]; ok {
		t.Errorf("serialized record carries abaCode = %v", v)
	}
}

// TestBandingCodesArePresent checks P-004's newest requirement: the tile shows
// the four-letter code, so the ingest must carry it.
//
// Verifies: P-004. Criterion: AC-013.
func TestBandingCodesArePresent(t *testing.T) {
	ds := load(t)
	var missing int
	for _, s := range ds.Species {
		if s.BandingCode == "" {
			missing++
		}
	}
	// A handful of ABA rows genuinely have no alpha code; a large share missing
	// would mean the column moved.
	if frac := float64(missing) / float64(len(ds.Species)); frac > 0.05 {
		t.Errorf("%d of %d species (%.1f%%) have no banding code", missing, len(ds.Species), frac*100)
	}
}

// Verifies: P-004. Criterion: AC-013.
func TestNamesArePopulatedNotDefaulted(t *testing.T) {
	ds := load(t)
	for _, s := range ds.Species {
		switch {
		case s.CommonName == "":
			t.Errorf("%s: empty common name", s.ScientificName)
		case s.Family == "" || s.FamilyCommon == "":
			t.Errorf("%s: empty family (%q / %q)", s.ScientificName, s.Family, s.FamilyCommon)
		case s.CommonName == s.ScientificName:
			// The tell of a silent fallback: an earlier design shipped this for
			// every one of its species.
			t.Errorf("%s: common name equals the scientific name", s.ScientificName)
		case s.FamilyColor == "":
			t.Errorf("%s: no family color", s.ScientificName)
		}
	}
}

// Verifies: P-041. Criterion: AC-030.
func TestSegmentsSumToTotalExactly(t *testing.T) {
	ds := load(t)
	for _, s := range ds.Species {
		sum := s.Segments.Bill.Value + s.Segments.Head.Value + s.Segments.Neck.Value +
			s.Segments.Body.Value + s.Segments.Tail.Value
		if math.Abs(sum-s.TotalLength.Value) > 1e-9 {
			t.Errorf("%s: segments sum to %.4f, total is %.4f", s.ScientificName, sum, s.TotalLength.Value)
		}
	}
}

// Verifies: P-062. Criterion: AC-040.
func TestEveryMetricHasWellFormedProvenance(t *testing.T) {
	ds := load(t)
	for i := range ds.Species {
		s := &ds.Species[i]
		for name, m := range s.AllMetrics() {
			p := m.Prov
			switch p.Kind {
			case dataset.Measured, dataset.NearMatch:
				if p.Source == "" {
					t.Errorf("%s/%s: %s value with no source", s.ScientificName, name, p.Kind)
				}
				if p.SourceTaxon == "" {
					t.Errorf("%s/%s: %s value with no sourceTaxon", s.ScientificName, name, p.Kind)
				}
			case dataset.Modeled:
				if strings.TrimSpace(p.Model) == "" {
					t.Errorf("%s/%s: modeled value with no model named", s.ScientificName, name)
				}
			default:
				t.Errorf("%s/%s: unknown provenance kind %q", s.ScientificName, name, p.Kind)
			}
			// near-match means the measurement belongs to a DIFFERENT taxon.
			// If the taxon matches the species, the label is wrong.
			if p.Kind == dataset.NearMatch && p.SourceTaxon == s.ScientificName {
				t.Errorf("%s/%s: near-match names its own taxon", s.ScientificName, name)
			}
		}
	}
}

// Verifies: P-060, P-062. Criterion: AC-041.
func TestProvenanceMatchesTheSourcesUsed(t *testing.T) {
	ds := load(t)
	var measuredSpan, nearSpan, modeledSpan int
	for i := range ds.Species {
		s := &ds.Species[i]
		if s.TotalLength.Prov.Kind != dataset.Modeled {
			t.Errorf("%s: length is %q; no measured length source exists (CR-002)",
				s.ScientificName, s.TotalLength.Prov.Kind)
		}
		switch s.Wingspan.Prov.Kind {
		case dataset.Measured:
			measuredSpan++
		case dataset.NearMatch:
			nearSpan++
		default:
			modeledSpan++
		}
	}
	// 301 is the empirical-wingspan count measured against the Fox dataset and
	// AVONET's crosswalk before this code existed, and recorded in
	// spec/sources/fox-wingspan/PROVENANCE.md.
	if measuredSpan != 301 {
		t.Errorf("measured wingspans: got %d, want 301", measuredSpan)
	}
	if measuredSpan+nearSpan+modeledSpan != len(ds.Species) {
		t.Error("wingspan kinds do not partition the dataset")
	}
}

// Verifies: P-045, P-062. Criterion: AC-035.
func TestNeckIsModeledAndNamesItsFamilyRatio(t *testing.T) {
	ds := load(t)
	for i := range ds.Species {
		s := &ds.Species[i]
		p := s.Segments.Neck.Prov
		if p.Kind != dataset.Modeled {
			t.Fatalf("%s: neck kind is %q, want modeled", s.ScientificName, p.Kind)
		}
		if !strings.Contains(p.Model, s.Family) {
			t.Errorf("%s: neck model %q does not name the family", s.ScientificName, p.Model)
		}
		if !strings.ContainsAny(p.Model, "0123456789") {
			t.Errorf("%s: neck model %q names no ratio", s.ScientificName, p.Model)
		}
	}
}

// Verifies: P-060, P-041. Criterion: AC-044.
func TestTotalLengthSaysItIsASum(t *testing.T) {
	ds := load(t)
	m := ds.Species[0].TotalLength.Prov.Model
	for _, want := range []string{"sum", "measured", "modeled"} {
		if !strings.Contains(strings.ToLower(m), want) {
			t.Errorf("total length model %q does not mention %q; a reader seeing a measured "+
				"bill inside a modeled total needs it to explain how both are true", m, want)
		}
	}
}

// Verifies: P-006, P-007. Criterion: AC-014.
func TestNearMatchesAreExactlyTheMappedOnes(t *testing.T) {
	ds := load(t)
	got := map[string]bool{}
	for i := range ds.Species {
		s := &ds.Species[i]
		if s.Segments.Bill.Prov.Kind == dataset.NearMatch {
			got[s.ScientificName] = true
		}
	}
	for name := range taxonomy.NearMatches {
		if !got[name] {
			t.Errorf("%s is mapped as a near-match but its traits are not labeled so", name)
		}
		delete(got, name)
	}
	for name := range got {
		t.Errorf("%s is labeled near-match but appears in no mapping", name)
	}
}

func mustJSON(t *testing.T, v any) string {
	t.Helper()
	b, err := jsonMarshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func jsonMarshal(v any) ([]byte, error) { return json.Marshal(v) }

// TestHeadIsNotAConstant guards the regression that prompted the model change:
// the cranium must vary with the bird, not sit on a floor.
//
// The previous model pinned 442 of 841 species at exactly 12 mm, so half the
// dataset drew an identical head that consumed a fixed share of every small
// bird's length and differentiated nothing.
//
// Verifies: P-062. Criterion: AC-065.
func TestHeadIsNotAConstant(t *testing.T) {
	ds := load(t)
	counts := map[float64]int{}
	for i := range ds.Species {
		counts[ds.Species[i].Segments.Head.Value]++
	}
	var worst float64
	var worstN int
	for v, n := range counts {
		if n > worstN {
			worst, worstN = v, n
		}
	}
	// A floor shows up as a large pile on one value. Some sharing is expected,
	// since mass is reported to limited precision.
	if frac := float64(worstN) / float64(len(ds.Species)); frac > 0.05 {
		t.Errorf("%d of %d species (%.0f%%) share head length %.1f mm; the cranium has "+
			"collapsed to a constant", worstN, len(ds.Species), frac*100, worst)
	}
	// And it must track mass: the heaviest bird needs the largest head.
	var heaviest, biggestHead *dataset.Species
	for i := range ds.Species {
		s := &ds.Species[i]
		if heaviest == nil || s.Mass.Value > heaviest.Mass.Value {
			heaviest = s
		}
		if biggestHead == nil || s.Segments.Head.Value > biggestHead.Segments.Head.Value {
			biggestHead = s
		}
	}
	if heaviest.ScientificName != biggestHead.ScientificName {
		t.Errorf("heaviest bird is %s but the largest head belongs to %s",
			heaviest.ScientificName, biggestHead.ScientificName)
	}
}

// TestLegIsModeledFromTheMeasuredTarsus asserts P-040's leg model end to end on
// the shipped dataset: every leg is derived from the measured tarsus by a named
// per-family factor, and says so.
//
// The factors are invented -- no source measures exposed leg length -- so the
// thing that makes serializing them defensible is the same thing that makes the
// neck defensible (P-045, CR-007): the value is tagged modeled, the model string
// names the factor, the family and the tarsus it was applied to, and the tarsus
// itself is carried alongside so a reader can recover the measurement.
//
// Verifies: P-040, P-062. Criterion: AC-066.
func TestLegIsModeledFromTheMeasuredTarsus(t *testing.T) {
	ds := load(t)
	var lifted int
	for i := range ds.Species {
		s := &ds.Species[i]

		if s.Leg.Prov.Kind != dataset.Modeled {
			t.Errorf("%s: leg is %q; the exposed leg is never measured by any source",
				s.ScientificName, s.Leg.Prov.Kind)
		}
		if s.Tarsus.Prov.Kind == dataset.Modeled {
			t.Errorf("%s: tarsus is modeled; it is the measurement the leg is built from",
				s.ScientificName)
		}
		// Recompute the factor rather than trusting the value: this is the one
		// check that fails if the table and the ingest disagree.
		k, ok := morphology.LegFactors[s.Family]
		if !ok {
			k = morphology.DefaultLegFactor
		}
		if want := morphology.Round1(s.Tarsus.Value * k); math.Abs(s.Leg.Value-want) > 1e-9 {
			t.Errorf("%s: leg is %.1f mm, but %.1f x the %.1f mm tarsus is %.1f",
				s.ScientificName, s.Leg.Value, k, s.Tarsus.Value, want)
		}
		// The model string has to let a reader undo the model.
		for _, want := range []string{s.Family, fmt.Sprintf("%.1f", k), fmt.Sprintf("%.1f mm", s.Tarsus.Value)} {
			if !strings.Contains(s.Leg.Prov.Model, want) {
				t.Errorf("%s: leg model %q does not name %q", s.ScientificName, s.Leg.Prov.Model, want)
			}
		}
		if s.Leg.Value < s.Tarsus.Value {
			t.Errorf("%s: modeled leg %.1f is shorter than the tarsus %.1f it contains",
				s.ScientificName, s.Leg.Value, s.Tarsus.Value)
		}
		if s.Leg.Value > s.Tarsus.Value {
			lifted++
		}

		// The tile's width reference is the wingspan alone. That is sound only
		// while a pair of legs spreads no wider than the wings, which is
		// structural -- a bird's wingspan dwarfs its legs -- but nothing
		// enforces it, so say so here rather than carry a branch for a case
		// that cannot arise.
		if spread := 2 * s.Leg.Value * view.LegCos; spread > s.Wingspan.Value {
			t.Errorf("%s: legs spread %.1f mm against a %.1f mm wingspan; the tile width "+
				"reference no longer covers the drawing", s.ScientificName, spread, s.Wingspan.Value)
		}
	}
	if lifted == 0 {
		t.Error("no species has a leg longer than its tarsus; the family factors are doing nothing")
	}
}

// TestFeetBelowTheTailAreConfinedToWaders checks the consequence of the leg
// model that the renderer has to cope with, on real data.
//
// Modeling the exposed leg made "the feet never fall below the tail tip" false
// for the first time, which is why the renderer centers on the drawn extent
// rather than on bill-to-tail. Two things must hold. It has to happen at all,
// or that renderer path is dead code no test exercises. And it has to stay
// confined to the long-legged families -- a hummingbird standing lower than its
// own tail would mean the factors had escaped, and it would look absurd.
//
// Verifies: P-040. Criterion: AC-066.
func TestFeetBelowTheTailAreConfinedToWaders(t *testing.T) {
	ds := load(t)
	var below []string
	for i := range ds.Species {
		s := &ds.Species[i]
		if view.VerticalExtent(s) <= s.TotalLength.Value {
			continue
		}
		below = append(below, s.CommonName)
		// Long-legged means a family the table singles out, not the default.
		if k, ok := morphology.LegFactors[s.Family]; !ok || k < 1.3 {
			t.Errorf("%s (%s, factor %.1f) stands lower than its tail tip; only waders and "+
				"cranes should", s.CommonName, s.Family, k)
		}
	}
	if len(below) == 0 {
		t.Error("no species reaches below its tail tip, so the renderer's extra room for feet " +
			"is never exercised by real data")
	}
}
