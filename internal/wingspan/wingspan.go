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

// Package wingspan reads empirical wingspans from the Fox et al. 2026 dataset.
//
// Only rows flagged needs_prediction = 0 are read. The dataset's
// wingspan_predicted_* columns are regressions from AVONET wing chord;
// ingesting them would reintroduce modeled values wearing a measured source's
// name, which is the error decisions.noOpenLengthAndWingspan exists to prevent. BirdView models its own
// estimates and labels them.
//
// Verifies: product.lengthAndWingspanSources, product.provenancePerMetric.
package wingspan

import (
	"fmt"
	"strconv"
	"strings"

	"birdview/internal/xlsx"
)

// Source names this dataset in provenance records.
const Source = "Fox et al. 2026 avian wingspan database (CC BY 4.0)"

// Reported is an empirical wingspan with the range the source reports.
// product.lengthAndWingspanSources requires storing the triple rather than flattening to a point.
type Reported struct {
	Mid, Min, Max float64 // centimeters as published
	Taxon         string  // the BirdLife binomial it was filed under
}

// Table is keyed on BirdLife binomial (Species1), which is how the dataset
// identifies species -- not on eBird names. The join therefore runs through
// AVONET's own BirdLife-eBird crosswalk rather than matching names directly.
type Table map[string]Reported

// observation is one empirical row, used to fit the fallback model.
type observation struct {
	order, group string
	chordMm      float64
	spanCm       float64
}

// Open reads the empirical rows and fits the fallback model from them.
func Open(path string) (Table, *Model, error) {
	f, err := xlsx.Open(path)
	if err != nil {
		return nil, nil, fmt.Errorf("wingspan.Open %s: %w", path, err)
	}
	defer f.Close()

	rows, err := f.Rows("wingspan data")
	if err != nil {
		return nil, nil, fmt.Errorf("wingspan: %w", err)
	}
	defer rows.Close()

	out := Table{}
	var obs []observation
	var idx map[string]int
	for rows.Next() {
		cells, err := rows.Columns()
		if err != nil {
			return nil, nil, err
		}
		if idx == nil {
			idx = map[string]int{}
			for i, c := range cells {
				idx[strings.TrimSpace(c)] = i
			}
			for _, want := range []string{"Species1", "needs_prediction", "wingspan_reported_mid_all",
				"Wing.Length_AVONET", "Order1", "Order_model_grouped"} {
				if _, ok := idx[want]; !ok {
					return nil, nil, fmt.Errorf("wingspan: missing required column %q", want)
				}
			}
			continue
		}
		if !isEmpirical(at(cells, idx["needs_prediction"])) {
			continue // predicted only -- deliberately skipped
		}
		name := strings.TrimSpace(at(cells, idx["Species1"]))
		mid := num(at(cells, idx["wingspan_reported_mid_all"]))
		if name == "" || mid <= 0 {
			continue
		}
		out[name] = Reported{
			Mid:   mid,
			Min:   num(at(cells, idx["wingspan_reported_min_all"])),
			Max:   num(at(cells, idx["wingspan_reported_max_all"])),
			Taxon: name,
		}
		if chord := num(at(cells, idx["Wing.Length_AVONET"])); chord > 0 {
			obs = append(obs, observation{
				order:   strings.TrimSpace(at(cells, idx["Order1"])),
				group:   strings.TrimSpace(at(cells, idx["Order_model_grouped"])),
				chordMm: chord,
				spanCm:  mid,
			})
		}
	}
	return out, buildModel(obs), nil
}

// isEmpirical reports whether a needs_prediction cell means "we have a real
// measurement". The column is a spreadsheet boolean, which internal/xlsx renders as
// FALSE/TRUE while the underlying XML stores 0/1 -- so both spellings are
// accepted. Reading only "0" silently matched nothing and produced a dataset
// with zero measured wingspans, which looked like a coverage problem rather
// than a parsing one.
func isEmpirical(v string) bool {
	switch strings.ToUpper(strings.TrimSpace(v)) {
	case "FALSE", "0":
		return true
	}
	return false
}

func at(cells []string, i int) string {
	if i < 0 || i >= len(cells) {
		return ""
	}
	return cells[i]
}

func num(s string) float64 {
	v, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
	if err != nil {
		return 0
	}
	return v
}
