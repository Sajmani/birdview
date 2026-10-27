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
	"sort"
	"strings"

	"birdview/internal/dataset"
)

// The find box's matching rule (product.nextView). The rule lives here: the
// server serves each bird's searchable fields, normalized, in /birds.json,
// and next-find.js only checks that the typed text, normalized the same way,
// occurs in one of them. TestNextFindInTheBrowser holds the script to
// searchMatches.

// normalizeSearch lowercases s and turns every run of characters other than
// a-z and 0-9 into one space, so "Chuck-will's-widow" is "chuck will s widow"
// and punctuation never decides a match. next-find.js does the same.
func normalizeSearch(s string) string {
	var b strings.Builder
	space := true
	for _, r := range strings.ToLower(s) {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
			space = false
		} else if !space {
			b.WriteByte(' ')
			space = true
		}
	}
	return strings.TrimSpace(b.String())
}

// fieldSep separates the fields of a search text. A normalized query never
// contains it, so a match cannot run from one field into the next: "d hawk"
// finds "sharp shinned hawk" but not "...caprimulgidae|...nighthawk".
const fieldSep = "|"

// searchMatches reports whether query, normalized, occurs unbroken within one
// field of text, a bird's search text. An empty query matches everything.
func searchMatches(text, query string) bool {
	return strings.Contains(text, normalizeSearch(query))
}

// searchTexts builds every species' search text: its English names in both
// locales, its species code, its scientific name, and the scientific names
// of its family and order, each normalized, joined by fieldSep. Not the
// families' and orders' English descriptions, which list other birds:
// Accipitridae's "Kites, Old World Vultures, Eagles, and Hawks" would make
// "hawk" find every kite.
func searchTexts(ds *dataset.Dataset) map[string]string {
	out := make(map[string]string, len(ds.Species))
	for i := range ds.Species {
		sp := &ds.Species[i]
		var fields []string
		seen := map[string]bool{}
		names := make([]string, 0, len(sp.Names))
		for _, n := range sp.Names {
			names = append(names, n)
		}
		sort.Strings(names) // map order would make the page differ run to run
		for _, f := range append(names, sp.SpeciesCode, sp.ScientificName, sp.Family, sp.Order) {
			if n := normalizeSearch(f); n != "" && !seen[n] {
				seen[n] = true
				fields = append(fields, n)
			}
		}
		out[sp.ID] = strings.Join(fields, fieldSep)
	}
	return out
}

// hiddenField is one setting a home page form carries, so that Enter
// without JavaScript keeps every other setting.
type hiddenField struct{ Name, Value string }
