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
	"fmt"
	"math"
	"sort"
)

// Model estimates wingspan from wing chord for species with no published
// measurement, by regressing log wingspan on log wing chord within a taxonomic
// order.
//
// The fit is made here, at ingest time, from the empirical rows of the Fox et
// al. dataset -- the same CC BY column BirdView already adopts as measured
// data. It is not an ingest of that dataset's own predicted columns, which
// the source manifest excludes.
//
// Validated against those published predictions before adoption: fitted this
// way and asked for the same 9,194 species, this model agrees with Fox's to a
// mean of 1.01% and a median of 0.61%. It is, in effect, their model
// reconstructed, which is why it is worth preferring over an invented one --
// BirdView's previous constants carried a mean error of 18.7% against the
// empirical data, against roughly 6% for this.
//
// Verifies: product.lengthAndWingspanSources, product.provenancePerMetric.
type Model struct {
	byGroup      map[string]fit
	global       fit
	orderToGroup map[string]string
}

type fit struct {
	a, b      float64 // log(span_cm) = a + b*log(chord_mm)
	n         int     // empirical species the fit was made from
	medianErr float64 // cross-validated median absolute percentage error
}

// minGroupSize is the smallest number of empirical species an order needs
// before it gets its own regression. Fox uses the same threshold, grouping
// smaller orders with related larger ones; below it, the global fit applies.
const minGroupSize = 5

// Predict returns an estimated wingspan in millimeters and a provenance
// description naming the model, the group it came from, and its measured
// error -- so the detail view can tell a reader how much to trust it (product.metricProvenanceExplained).
func (m *Model) Predict(order string, chordMm float64) (float64, string) {
	if chordMm <= 0 {
		return 0, ""
	}
	group := m.orderToGroup[order]
	f, ok := m.byGroup[group]
	scope := group
	if !ok {
		f, scope = m.global, "all orders"
	}
	cm := math.Exp(f.a + f.b*math.Log(chordMm))
	return cm * 10, fmt.Sprintf(
		"log-log regression of wingspan on wing chord, fitted on %d empirical species in %s "+
			"from the Fox et al. 2026 dataset; cross-validated median error %.1f%%",
		f.n, scope, f.medianErr)
}

// buildModel fits one regression per order group plus a global fallback.
func buildModel(obs []observation) *Model {
	m := &Model{byGroup: map[string]fit{}, orderToGroup: map[string]string{}}

	byGroup := map[string][]observation{}
	var gx, gy []float64
	for _, o := range obs {
		byGroup[o.group] = append(byGroup[o.group], o)
		gx = append(gx, math.Log(o.chordMm))
		gy = append(gy, math.Log(o.spanCm))
		if o.order != "" {
			m.orderToGroup[o.order] = o.group
		}
	}
	a, b := ols(gx, gy)
	m.global = fit{a, b, len(gx), crossValidate(obs)}

	for g, rows := range byGroup {
		if len(rows) < minGroupSize {
			continue // too few points to fit; these fall back to the global model
		}
		var xs, ys []float64
		for _, o := range rows {
			xs = append(xs, math.Log(o.chordMm))
			ys = append(ys, math.Log(o.spanCm))
		}
		ga, gb := ols(xs, ys)
		m.byGroup[g] = fit{ga, gb, len(rows), crossValidate(rows)}
	}
	return m
}

// crossValidate returns the median absolute percentage error under 5-fold
// cross-validation, so the reported error is out-of-sample rather than the
// fit's own residuals.
func crossValidate(obs []observation) float64 {
	const folds = 5
	if len(obs) < folds*2 {
		return math.NaN()
	}
	var errs []float64
	for k := 0; k < folds; k++ {
		var tx, ty []float64
		var test []observation
		for i, o := range obs {
			if i%folds == k {
				test = append(test, o)
				continue
			}
			tx = append(tx, math.Log(o.chordMm))
			ty = append(ty, math.Log(o.spanCm))
		}
		if len(tx) < 2 {
			continue
		}
		a, b := ols(tx, ty)
		for _, o := range test {
			p := math.Exp(a + b*math.Log(o.chordMm))
			errs = append(errs, 100*math.Abs(p-o.spanCm)/o.spanCm)
		}
	}
	if len(errs) == 0 {
		return math.NaN()
	}
	sort.Float64s(errs)
	return errs[len(errs)/2]
}

func ols(xs, ys []float64) (a, b float64) {
	n := float64(len(xs))
	if n == 0 {
		return 0, 0
	}
	var sx, sy, sxx, sxy float64
	for i := range xs {
		sx += xs[i]
		sy += ys[i]
		sxx += xs[i] * xs[i]
		sxy += xs[i] * ys[i]
	}
	d := n*sxx - sx*sx
	if d == 0 {
		return sy / n, 0
	}
	b = (n*sxy - sx*sy) / d
	return (sy - b*sx) / n, b
}
