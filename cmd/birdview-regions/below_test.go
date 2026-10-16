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

package main

import (
	"os"
	"path/filepath"
	"testing"

	"birdview/internal/region"
)

// TestBelowBarCellsRoundTrip asserts acceptance.belowBarCountsKept's first assertion: a cell under
// the presence bar is written with its mark and its counts, and reads back
// as below; an unmarked cell reads back as present.
//
// Verifies: product.regionAndMonthFilter, decisions.belowBarCounted. Criterion: acceptance.belowBarCountsKept.
func TestBelowBarCellsRoundTrip(t *testing.T) {
	below := region.Reported{ID: "USA.44_1", Records: 19,
		Months: [12]int64{2, 1, 3, 1, 2, 1, 1, 2, 2, 1, 2, 1}, Below: true}
	present := region.Reported{ID: "USA.5_1", Records: 12,
		Months: [12]int64{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1}}
	path := filepath.Join(t.TempDir(), "presence.tsv")
	err := write(path, []string{"Dromaius novaehollandiae"},
		map[string][]region.Reported{"Dromaius novaehollandiae": {below, present}},
		map[string]string{"USA.44_1": "Texas", "USA.5_1": "California"},
		map[string]int64{"USA.44_1": 120, "USA.5_1": 120},
		map[string][12]int64{"USA.44_1": {10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10},
			"USA.5_1": {10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10}})
	if err != nil {
		t.Fatal(err)
	}
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	table, _, _, err := region.LoadPresence(f)
	if err != nil {
		t.Fatal(err)
	}
	got := map[string]region.Reported{}
	for _, r := range table["Dromaius novaehollandiae"] {
		got[r.ID] = r
	}
	if r, ok := got["USA.44_1"]; !ok || !r.Below || r.Records != 19 || r.Months != below.Months {
		t.Errorf("the below-bar cell read back as %+v (found %v), want %+v", r, ok, below)
	}
	if r, ok := got["USA.5_1"]; !ok || r.Below || r.Records != 12 {
		t.Errorf("the present cell read back as %+v (found %v), want present with 12 records", r, ok)
	}
}
