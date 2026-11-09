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

package wingspan

import (
	"math"
	"os"
	"testing"
)

const foxFile = "../../third_party/foxwingspan/wingspan_fox2026_AppendixS1.xlsx"

// TestModelAgreesWithFoxPublishedPredictions is the criterion that justifies
// fitting our own regression instead of ingesting Fox's predicted columns: if
// ours is their model reconstructed, the two must agree closely on the species
// they predicted.
//
// The threshold is derived from the comparison made before this code was
// written -- mean 1.01%, median 0.61%, p99 4.15% across 9,194 species -- not
// from running this implementation. A regression in the fitting code moves
// these numbers immediately.
//
// Verifies: product.lengthAndWingspanSources, product.provenancePerMetric. Criterion: acceptance.wingspanModelReconstructsFox.
func TestModelAgreesWithFoxPublishedPredictions(t *testing.T) {
	if _, err := os.Stat(foxFile); err != nil {
		t.Skip("skipping: Fox wingspan workbook not present")
	}
	_, model, err := Open(foxFile)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	preds, err := publishedPredictions(foxFile)
	if err != nil {
		t.Fatalf("publishedPredictions: %v", err)
	}
	if len(preds) < 9000 {
		t.Fatalf("expected ~9,194 published predictions, got %d", len(preds))
	}

	var sum float64
	var n, over10 int
	for _, p := range preds {
		got, desc := model.Predict(p.order, p.chordMm)
		if got <= 0 || desc == "" {
			t.Fatalf("%s: model returned no estimate", p.species)
		}
		diff := 100 * math.Abs(got/10-p.foxCm) / p.foxCm
		sum += diff
		n++
		if diff > 10 {
			over10++
		}
	}
	mean := sum / float64(n)
	if mean > 2.0 {
		t.Errorf("mean disagreement with Fox's published predictions is %.2f%%, want <= 2%%; "+
			"our fit is supposed to reconstruct their model", mean)
	}
	// A handful of extreme-aspect seabirds disagree by more; a large share
	// doing so would mean the grouping had broken.
	if frac := float64(over10) / float64(n); frac > 0.02 {
		t.Errorf("%.1f%% of species disagree by more than 10%%, want <= 2%%", frac*100)
	}
	t.Logf("agreement with Fox across %d species: mean %.2f%%, %d over 10%%", n, mean, over10)
}

// TestPredictAlwaysExplainsItself checks every estimate names its model, its
// scope, and its error -- the detail view has nothing else to show (product.metricProvenanceExplained).
func TestPredictAlwaysExplainsItself(t *testing.T) {
	if _, err := os.Stat(foxFile); err != nil {
		t.Skip("skipping: Fox wingspan workbook not present")
	}
	_, model, err := Open(foxFile)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	for _, order := range []string{"Passeriformes", "Galliformes", "", "Nonesuchiformes"} {
		v, desc := model.Predict(order, 120)
		if v <= 0 {
			t.Errorf("%q: no estimate", order)
		}
		for _, want := range []string{"regression", "error"} {
			if !contains(desc, want) {
				t.Errorf("%q: description %q lacks %q", order, desc, want)
			}
		}
	}
}

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

// TestEveryFitStatesItsError asserts product.lengthAndWingspanSources: every fit states an
// out-of-sample median error, by leave-one-out where a group is too small for five folds, so no
// modeled wingspan reads "NaN%" (as Piciformes and the loons and grebes did).
//
// Verifies: product.lengthAndWingspanSources, product.metricProvenanceExplained. Criterion: acceptance.wingspanErrorStated.
func TestEveryFitStatesItsError(t *testing.T) {
	var small []observation
	for i, chord := range []float64{90, 110, 130, 150, 170, 190, 210} {
		small = append(small, observation{chordMm: chord, spanCm: 0.4*chord*(1+0.03*float64(i%3-1)) + 2})
	}
	if e := crossValidate(small); math.IsNaN(e) || e <= 0 {
		t.Errorf("seven species give a cross-validated error of %v; want a finite out-of-sample error", e)
	}
	if _, err := os.Stat(foxFile); err != nil {
		t.Skip("skipping: Fox wingspan workbook not present")
	}
	_, model, err := Open(foxFile)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	for g, f := range model.byGroup {
		if math.IsNaN(f.medianErr) {
			t.Errorf("the fit for %s (%d species) states no error", g, f.n)
		}
	}
	for _, order := range []string{"Piciformes", "Gaviiformes", "Podicipediformes"} {
		if _, desc := model.Predict(order, 150); contains(desc, "NaN") || !contains(desc, "leave-one-out") {
			t.Errorf("%s's wingspan reads %q; want a leave-one-out error, not NaN", order, desc)
		}
	}
}
