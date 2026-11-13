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
	"archive/zip"
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"birdview/internal/gbif"
)

// TestCountyListingRule asserts decisions.countyListingRule on rows shaped like the county download: a
// species is listed in a county only when its state lists it and its county records reach both
// one per million of the county's records and 5; every other species with records there is
// written under the bar, so its records still count; and the county's total counts every name.
//
// Verifies: product.regionAndMonthFilter. Criterion: app.countyListingApplied.
func TestCountyListingRule(t *testing.T) {
	const small, big = "USA.33.24_1", "USA.33.32_1"
	eb := gbif.EBirdDatasetKey
	var rows []countyRow
	add := func(gid, name string, perMonth int64, months int) {
		for m := 1; m <= months; m++ {
			rows = append(rows, countyRow{Dataset: eb, GID: gid, Month: m, Name: name, N: perMonth})
		}
	}
	add(small, "Turdus migratorius", 100, 12)  // common, on the state list
	add(small, "Setophaga petechia", 1, 5)     // 5 records, on the state list
	add(small, "Spinus pinus", 1, 4)           // 4 records, on the state list
	add(small, "Nymphicus hollandicus", 50, 2) // plenty, but not on the state list
	add(small, "Sturnus vulgaris", 5000, 12)   // a name BirdView does not draw: counted in the total only
	// A county so heavily birded that one per million is 12 records.
	add(big, "Turdus migratorius", 1_000_000, 12)
	add(big, "Setophaga petechia", 1, 10) // 10 records: past the floor, under one per million
	drawn := []string{"Turdus migratorius", "Setophaga petechia", "Spinus pinus", "Nymphicus hollandicus"}
	stateListed := map[string]bool{"Turdus migratorius": true, "Setophaga petechia": true, "Spinus pinus": true}

	got, totals, monthly, err := countyPresence(rows, stateListed, drawn)
	if err != nil {
		t.Fatal(err)
	}
	listed := func(sp, gid string) (bool, bool) {
		for _, r := range got[sp] {
			if r.ID == gid {
				return !r.Below, true
			}
		}
		return false, false
	}
	for _, c := range []struct {
		sp, gid      string
		want, exists bool
	}{
		{"Turdus migratorius", small, true, true},
		{"Setophaga petechia", small, true, true},     // exactly 5
		{"Spinus pinus", small, false, true},          // 4: under the floor, still counted
		{"Nymphicus hollandicus", small, false, true}, // not on the state list
		{"Turdus migratorius", big, true, true},
		{"Setophaga petechia", big, false, true}, // under one per million of a busy county
	} {
		on, ok := listed(c.sp, c.gid)
		if ok != c.exists || on != c.want {
			t.Errorf("%s in %s: listed %v (written %v), want listed %v", c.sp, c.gid, on, ok, c.want)
		}
	}
	if want := int64(100*12 + 5 + 4 + 100 + 5000*12); totals[small] != want {
		t.Errorf("the small county's total is %d, want %d, every name counted", totals[small], want)
	}
	var sum int64
	for _, n := range monthly[small] {
		sum += n
	}
	if sum != totals[small] {
		t.Errorf("the small county's months sum to %d, its total %d", sum, totals[small])
	}
	if _, ok := got["Sturnus vulgaris"]; ok {
		t.Error("a name BirdView does not draw was written as a species")
	}
}

// TestCountyRowsFromDownload reads the county download's file as GBIF writes it: a zip holding one
// tab-separated table with a header.
//
// Verifies: tech.gbifDownloadsRecorded. Criterion: app.countyListingApplied.
func TestCountyRowsFromDownload(t *testing.T) {
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	w, _ := zw.Create("0001234-261103000000000.csv")
	w.Write([]byte("datasetkey\tlevel2gid\tmonth\tv_scientificname\tspecieskey\tn\n" +
		gbif.EBirdDatasetKey + "\tUSA.33.24_1\t7\tTurdus migratorius\t9705453\t19\n"))
	zw.Close()
	path := filepath.Join(t.TempDir(), "d.zip")
	if err := os.WriteFile(path, buf.Bytes(), 0o644); err != nil {
		t.Fatal(err)
	}
	rows, err := readCountyRows(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || rows[0] != (countyRow{Dataset: gbif.EBirdDatasetKey, GID: "USA.33.24_1", Month: 7, Name: "Turdus migratorius", N: 19}) {
		t.Errorf("read %+v", rows)
	}
}

// TestCountiesSplitByState splits one download covering several states and provinces into each
// state's rows, by the state each county's GADM code names, so one download builds every state's
// table.
//
// Verifies: tech.gbifDownloadsRecorded. Criterion: app.countyListingApplied.
func TestCountiesSplitByState(t *testing.T) {
	rows := []countyRow{
		{Dataset: gbif.EBirdDatasetKey, GID: "USA.33.24_1", Month: 7, Name: "Turdus migratorius", N: 19},
		{Dataset: gbif.EBirdDatasetKey, GID: "USA.44.101_1", Month: 1, Name: "Turdus migratorius", N: 4},
		{Dataset: gbif.EBirdDatasetKey, GID: "CAN.9.12_1", Month: 5, Name: "Turdus migratorius", N: 8},
		{Dataset: gbif.EBirdDatasetKey, GID: "USA.33.42_1", Month: 7, Name: "Corvus brachyrhynchos", N: 3},
	}
	got, err := countiesByState(rows)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string][]string{
		"USA.33_1": {"USA.33.24_1", "USA.33.42_1"},
		"USA.44_1": {"USA.44.101_1"},
		"CAN.9_1":  {"CAN.9.12_1"},
	}
	if len(got) != len(want) {
		t.Errorf("split into %d states, want %d: %v", len(got), len(want), got)
	}
	for st, gids := range want {
		var have []string
		for _, r := range got[st] {
			have = append(have, r.GID)
		}
		if strings.Join(have, ",") != strings.Join(gids, ",") {
			t.Errorf("%s holds %v, want %v", st, have, gids)
		}
	}
	if _, err := countiesByState([]countyRow{{GID: "USA.33_1"}}); err == nil {
		t.Error("a state's code in the county column was accepted as a county")
	}
}
