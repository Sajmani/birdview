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
	"testing"

	"birdview/internal/gbif"
)

// facet builds a raw facet for one source: name -> twelve monthly counts,
// with month 0 the year, as the resume cache holds it.
func facet(rows map[string][12]int64) map[int]map[string]int64 {
	out := map[int]map[string]int64{0: {}}
	for mo := 1; mo <= 12; mo++ {
		out[mo] = map[string]int64{}
	}
	for name, ms := range rows {
		for i, n := range ms {
			if n == 0 {
				continue
			}
			out[i+1][name] += n
			out[0][name] += n
		}
	}
	return out
}

func flat(n int64) [12]int64 {
	var v [12]int64
	for i := range v {
		v[i] = n
	}
	return v
}

// TestMeasureCountsEBirdAndFillsOnlyItsDeclaredGaps asserts decisions.countEbirdFirst's
// counting rule, one case per clause.
//
// Verifies: product.regionAndMonthFilter, product.shareOfRecords, product.namesRecordsFiledUnder, decisions.countEbirdFirst. Criterion: acceptance.ebirdCountedInatFillsGaps.
func TestMeasureCountsEBirdAndFillsOnlyItsDeclaredGaps(t *testing.T) {
	avi := map[string]string{
		"turdus migratorius": "Turdus migratorius", "geopelia cuneata": "Geopelia cuneata",
		"asio otus": "Asio otus", "passerina ciris": "Passerina ciris",
		"setophaga petechia": "Setophaga petechia", "setophaga aestiva": "Setophaga aestiva",
	}

	// New York, at its real size -- about 62 million eBird records, carried
	// here by the robin -- because the presence bar is relative and a small
	// fixture moves it: at 1.2 million records two Diamond Doves clear 1 per
	// million, which is how this fixture first failed against correct code.
	// Also an escaped Diamond Dove, and a Long-eared Owl eBird withholds
	// across the United States.
	ny := map[string]map[int]map[string]int64{
		// The Diamond Dove's numbers are New York's real ones: 2 on eBird, 1
		// on iNaturalist. A dove with no eBird records at all is never
		// counted and so never judged, which is how the first draft of
		// this test passed with the old presence rule put back.
		gbif.EBirdDatasetKey: facet(map[string][12]int64{
			"turdus migratorius": flat(5200000), "setophaga petechia": flat(500),
			"geopelia cuneata": {0, 0, 0, 0, 1, 1}}),
		iNatDatasetKey: facet(map[string][12]int64{
			"turdus migratorius": flat(1000), "geopelia cuneata": {0, 0, 0, 0, 0, 1},
			"asio otus":         {8, 6, 5, 2, 1, 1, 1, 1, 2, 3, 4, 5},
			"setophaga aestiva": flat(50)}),
	}
	m, err := measure(ny, "USA.33_1", avi)
	if err != nil {
		t.Fatal(err)
	}
	// eBird alone: iNaturalist's robins are not added.
	if got := m.counts["turdus migratorius"]; got != 62400000 {
		t.Errorf("robin counts %d, want 62,400,000 -- eBird's alone", got)
	}
	// The escape counts its two eBird records and is NOT present: two of 62
	// million is far under 1 per million, and the one iNaturalist photo that
	// cleared the old rule's 0.43-record bar no longer gets a vote.
	if got := m.counts["geopelia cuneata"]; got != 2 {
		t.Errorf("Diamond Dove counts %d, want eBird's 2", got)
	}
	if m.present["geopelia cuneata"] {
		t.Error("an escaped Diamond Dove is present in New York on one iNaturalist photo")
	}
	// The withheld owl is counted from iNaturalist, present on iNaturalist's
	// own total, and marked withheld all year.
	if got := m.counts["asio otus"]; got != 39 {
		t.Errorf("Long-eared Owl counts %d, want iNaturalist's 39 where eBird withholds it", got)
	}
	if !m.present["asio otus"] {
		t.Error("Long-eared Owl is not present in New York: eBird's silence is policy, not absence")
	}
	if m.withheld["asio otus"] != [12]bool{true, true, true, true, true, true, true, true, true, true, true, true} {
		t.Errorf("Long-eared Owl withheld = %v, want all year", m.withheld["asio otus"])
	}
	// Shared name: Northern Yellow Warbler takes petechia's eBird count,
	// and the total counts it once.
	if a, p := m.counts["setophaga aestiva"], m.counts["setophaga petechia"]; a != 6000 || p != 6000 {
		t.Errorf("aestiva %d, petechia %d; both should carry eBird's shared 6,000", a, p)
	}
	if !m.present["setophaga aestiva"] {
		t.Error("Northern Yellow Warbler is absent: its records are filed under petechia")
	}
	if want := int64(62400000 + 6000 + 39 + 2); m.total != want {
		t.Errorf("total %d, want %d: eBird's records plus the withheld owl, the shared count "+
			"once, and no iNaturalist robins or escapes", m.total, want)
	}

	// Florida: Painted Bunting, which eBird withholds October to mid-April.
	fl := map[string]map[int]map[string]int64{
		gbif.EBirdDatasetKey: facet(map[string][12]int64{
			"turdus migratorius": flat(50000),
			"passerina ciris":    {0, 0, 0, 12483, 2349, 1034, 905, 1680, 2759, 0, 0, 0}}),
		iNatDatasetKey: facet(map[string][12]int64{
			"passerina ciris": {40, 38, 45, 120, 60, 30, 25, 40, 55, 30, 35, 42}}),
	}
	m, err = measure(fl, "USA.10_1", avi)
	if err != nil {
		t.Fatal(err)
	}
	got := *m.monthly["passerina ciris"]
	want := [12]int64{40, 38, 45, 12483, 2349, 1034, 905, 1680, 2759, 30, 35, 42}
	if got != want {
		t.Errorf("Painted Bunting by month %v, want %v: iNaturalist fills the withheld "+
			"months and nothing else", got, want)
	}
}
