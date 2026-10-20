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

// Package avilist reads names, family and order, and the eBird species code
// from the AviList global checklist.
//
// AviList replaced the ABA checklist as BirdView's name source because it is
// CC BY 4.0 and the ABA checklist is not: serving BirdView reproduces whatever
// supplies its names, so an all-rights-reserved list blocked deployment
// outright (decisions.avilistReplacesAba).
//
// Two English names are read, not one. AviList's own name is IOC usage and the
// extended release also carries the eBird/Clements name, which is North
// American usage -- Little Auk and Dovekie for the same bird. product.englishNameChoice serves
// whichever the reader asks for, and that is only possible because both are in
// this one file.
//
// Verifies: product.namesFromAvilist, product.englishNameChoice.
package avilist

import (
	"fmt"
	"strconv"
	"strings"

	"birdview/internal/xlsx"
)

// Source names this dataset in provenance records and in the page footer.
const Source = "AviList v2025b (CC BY 4.0)"

// sheetPrefix finds the taxonomy tab. The workbook has six; the rest are
// decision summaries, citations and usage notes.
//
// Matched by prefix rather than named exactly, because the tab carries the
// version AND the release in its name -- "AviList v2025b extended" -- so a
// literal name breaks on every annual update and on the short release too.
const sheetPrefix = "AviList"

// Species is one AviList row, reduced to what BirdView serves.
type Species struct {
	Scientific string
	// NameNA is the eBird/Clements English name -- North American usage.
	// Empty where AviList's species concept has no 1:1 Clements counterpart.
	NameNA string
	// NameWorld is AviList's own English name, following IOC. Always present.
	NameWorld string
	// Code is the eBird species code, e.g. "amerob". It is BirdView's species
	// id and the code shown on a tile. Empty on the same concept mismatches
	// that leave NameNA empty.
	//
	// NOT fixed-width, despite being called a six-letter code: over BirdView's
	// scope they run 4 to 8 characters, and 104 of them carry a disambiguating
	// digit. Anything laying them out must not assume six.
	Code string

	Family        string
	FamilyEnglish string
	Order         string

	// Seq is AviList's linear sequence, which is phylogenetic rather than
	// alphabetical. BirdView orders families by the lowest Seq of any species
	// in them, so the family list reads the way a field guide does.
	Seq int
}

// Table is keyed on scientific name, which is how the ingest joins AviList to
// AVONET. Genus and epithet changes defeat that join and are handled by the
// synonym table in internal/taxonomy, not here -- this package reports what
// the file says and nothing more.
type Table map[string]Species

// required names the columns the reader cannot work without. The SHORT release
// of AviList omits two of them, so a build pointed at the wrong file fails here
// with a column name rather than silently producing 11,131 nameless species.
var required = []string{
	"Sequence", "Taxon_rank", "Scientific_name", "English_name_AviList",
	"English_name_Clements_v2025", "Species_code_Cornell_Lab",
	"Family", "Family_English_name", "Order",
}

// Open reads the species rows of an AviList workbook.
func Open(path string) (Table, error) {
	f, err := xlsx.Open(path)
	if err != nil {
		return nil, fmt.Errorf("avilist.Open %s: %w", path, err)
	}
	defer f.Close()

	sheet := ""
	for _, n := range f.SheetNames() {
		if strings.HasPrefix(n, sheetPrefix) {
			sheet = n
			break
		}
	}
	if sheet == "" {
		return nil, fmt.Errorf("avilist: %s has no sheet named %s...; tabs are %v",
			path, sheetPrefix, f.SheetNames())
	}

	rows, err := f.Rows(sheet)
	if err != nil {
		return nil, fmt.Errorf("avilist: sheet %q: %w", sheet, err)
	}
	defer rows.Close()

	out := Table{}
	var idx map[string]int
	for rows.Next() {
		cells, err := rows.Columns()
		if err != nil {
			return nil, fmt.Errorf("avilist: %w", err)
		}
		if idx == nil {
			idx = map[string]int{}
			for i, c := range cells {
				idx[strings.TrimSpace(c)] = i
			}
			for _, want := range required {
				if _, ok := idx[want]; !ok {
					return nil, fmt.Errorf(
						"avilist: %s has no %q column — this is the SHORT release; "+
							"BirdView needs the extended one", path, want)
				}
			}
			continue
		}
		if at(cells, idx["Taxon_rank"]) != "species" {
			continue // subspecies, genera, families, orders
		}
		sci := at(cells, idx["Scientific_name"])
		if sci == "" {
			continue
		}
		seq, _ := strconv.Atoi(at(cells, idx["Sequence"]))
		out[sci] = Species{
			Scientific:    sci,
			Seq:           seq,
			NameNA:        at(cells, idx["English_name_Clements_v2025"]),
			NameWorld:     at(cells, idx["English_name_AviList"]),
			Code:          at(cells, idx["Species_code_Cornell_Lab"]),
			Family:        at(cells, idx["Family"]),
			FamilyEnglish: at(cells, idx["Family_English_name"]),
			Order:         at(cells, idx["Order"]),
		}
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("avilist: %s yielded no species rows", path)
	}
	return out, nil
}

func at(cells []string, i int) string {
	if i < 0 || i >= len(cells) {
		return ""
	}
	return strings.TrimSpace(cells[i])
}
