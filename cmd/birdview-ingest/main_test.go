package main

import (
	"encoding/json"
	"math"
	"os"
	"strings"
	"testing"

	"birdview/internal/dataset"
	"birdview/internal/taxonomy"
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

// Verifies: P-001. Criterion: AC-011.
func TestRarityCodeNeverReachesTheDataset(t *testing.T) {
	ds := load(t)
	blob := mustJSON(t, ds.Species[0])
	for _, banned := range []string{"rarity", "abaCode", "code"} {
		if strings.Contains(strings.ToLower(blob), strings.ToLower(banned)) {
			t.Errorf("serialized record mentions %q; the rarity code is a scope filter, not data", banned)
		}
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
