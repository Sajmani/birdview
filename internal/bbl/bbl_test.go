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

package bbl

import (
	"os"
	"testing"
)

// TestParseTheVendoredList reads the vendored list and checks that it yields
// the species rows, with known codes where they should be, and none of the
// rows that are not one species.
//
// Verifies: product.bandingCodes. Criterion: app.bandingCodesShown.
func TestParseTheVendoredList(t *testing.T) {
	b, err := os.ReadFile("../../third_party/bbl/speclist.html")
	if err != nil {
		t.Fatal(err)
	}
	es := Parse(string(b))
	if len(es) < 1100 {
		t.Fatalf("%d species rows; the list has about 1,200", len(es))
	}
	by := map[string]Entry{}
	for _, e := range es {
		if _, dup := by[e.Alpha]; dup {
			t.Errorf("code %s appears twice", e.Alpha)
		}
		by[e.Alpha] = e
	}
	for code, sci := range map[string]string{
		"AMRO": "Turdus migratorius", "BLJA": "Cyanocitta cristata", "WEGR": "Aechmophorus occidentalis",
		"BITH": "Catharus bicknelli",
	} {
		if by[code].Scientific != sci {
			t.Errorf("%s is %+v, want %s", code, by[code], sci)
		}
	}
	if _, ok := by["GCBT"]; ok {
		t.Error("the unidentified Gray-cheeked/Bicknell's Thrush row was kept")
	}
}

// TestParseSkipsWhatIsNotOneSpecies checks each rule that drops a row on its
// own: in the vendored list they overlap (most such rows also lack a
// scientific name), so a list built here gives each its own row.
//
// Verifies: product.bandingCodes. Criterion: app.bandingCodesShown.
func TestParseSkipsWhatIsNotOneSpecies(t *testing.T) {
	row := func(code, common, sci string) string {
		return "<tr><td>1</td><td>" + code + "</td><td>" + common + "</td><td>2</td><td>x</td><td>" +
			sci + "</td><td></td><td></td><td>1</td></tr>"
	}
	page := row("AMRO", "American Robin", "Turdus migratorius") +
		row("UNTH", "Unidentified Thrush", "Turdus sp") +
		row("PAIR", "Gray-cheeked/Bicknell's Thrush", "Catharus sp") +
		row("HYBR", "Hybrid Warbler", "Vermivora sp") +
		row("CROS", "Some Cross", "Vermivora a x Vermivora b") +
		row("NOSC", "No Scientific Name", "") +
		row("abcd", "Lowercase Code", "Lower case") +
		row("FIVEL", "Five Letters", "Five letters")
	es := Parse(page)
	if len(es) != 1 || es[0].Alpha != "AMRO" {
		t.Errorf("parsed %+v; want only AMRO", es)
	}
}
