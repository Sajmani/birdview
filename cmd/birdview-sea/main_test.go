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
	"bytes"
	"strings"
	"testing"

	"birdview/internal/gbif"
	"birdview/internal/geo"
)

// TestMeasureBandsWaterRecords asserts tech.seaDistanceTable's method on a synthetic island in the
// North Atlantic box: eBird records are banded by their cell center's distance to shore, a cell
// whose center lies on land is dropped, iNaturalist rows and cells outside every sea box are left
// out, names no drawn species has are counted apart, and the table is written with its source.
//
// Verifies: tech.seaDistanceTable. Criterion: dataset.seaDistanceTable.
func TestMeasureBandsWaterRecords(t *testing.T) {
	island := [][][][2]float64{{{{-60, 40}, {-59, 40}, {-59, 41}, {-60, 41}, {-60, 40}}}}
	shore := geo.NewShore(island, -30)
	table := "datasetkey\tlat10\tlon10\tmonth\tv_scientificname\tspecieskey\tn\n" +
		gbif.EBirdDatasetKey + "\t405\t-590\t6\tCalonectris borealis\t1\t3\n" + // 0.05° east of shore: near
		gbif.EBirdDatasetKey + "\t405\t-588\t6\tCalonectris borealis\t1\t5\n" + // 0.25°, about 21 km: 10 to 30
		gbif.EBirdDatasetKey + "\t405\t-570\t6\tCalonectris borealis\t1\t7\n" + // 2°: 30 or more
		gbif.EBirdDatasetKey + "\t405\t-596\t6\tCalonectris borealis\t1\t11\n" + // on the island: dropped
		"50c9509d-22c7-4a22-a47d-8c48425ef4a7\t405\t-570\t6\tCalonectris borealis\t1\t13\n" + // iNaturalist
		gbif.EBirdDatasetKey + "\t50\t-570\t6\tCalonectris borealis\t1\t17\n" + // outside every box
		gbif.EBirdDatasetKey + "\t405\t-570\t6\tAves incognita\t9\t19\n"
	rows, err := parseCells(strings.NewReader(table))
	if err != nil {
		t.Fatal(err)
	}
	n := names{byKey: map[string]string{"1": "Calonectris borealis"}, byName: map[string]string{}}
	bands, unmatched := measure(rows, shore, n)
	if got := bands["Calonectris borealis"]; got != [3]int64{3, 5, 7} {
		t.Errorf("Cory's Shearwater's bands are %v, want [3 5 7]", got)
	}
	if unmatched != 19 {
		t.Errorf("%d records went unmatched, want 19", unmatched)
	}
	var out bytes.Buffer
	if err := write(&out, bands, "0025653-261023074918325", "10.15468/dl.4jngjn"); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"doi:10.15468/dl.4jngjn", "scientific name\tunder 10 km\t10 to 30 km\t30 km or more\n",
		"Calonectris borealis\t3\t5\t7\n"} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("the table lacks %q:\n%s", want, out.String())
		}
	}
	if _, err := parseCells(strings.NewReader("datasetkey\tlevel2gid\tmonth\n")); err == nil {
		t.Error("a download without the offshore-cells columns was accepted")
	}
}
