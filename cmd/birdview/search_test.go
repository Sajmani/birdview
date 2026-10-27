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
	"slices"
	"sort"
	"strings"
	"testing"
)

// searchCases are queries and the fixture birds each must leave showing:
// a name, a word shared by two names, the end of one word and the start of
// the next, a scientific name, a code in capitals, a name typed without its
// punctuation, a family; and three that must match nothing: words out of
// order, a query running from one field into the next, and nonsense.
var searchCases = []struct {
	q    string
	want []string
}{
	{"crow", []string{"amecro"}},
	{"american", []string{"amecro", "amerob"}},
	{"n robin", []string{"amerob"}},
	{"corvus brachy", []string{"amecro"}},
	{"AMECRO", []string{"amecro"}},
	{"black necked", []string{"blasti"}},
	{"ardeidae", []string{"greblu"}},
	{"heron great", nil},
	{"heron ardea", nil},
	{"zzz", nil},
}

// goMatches is the fixture birds the Go matcher keeps for q, sorted.
func goMatches(s *server, q string) []string {
	var out []string
	for i := range s.ds.Species {
		if id := s.ds.Species[i].ID; searchMatches(s.search[id], q) {
			out = append(out, id)
		}
	}
	sort.Strings(out)
	return out
}

// TestSearchMatchesNamesCodesAndTaxa checks the matching rule: every word
// must occur, ignoring case and punctuation, in a bird's names, code,
// scientific name, family, or order.
//
// The home page's find box matches by this rule (product.nextView).
//
// Verifies: product.nextView.
func TestSearchMatchesNamesCodesAndTaxa(t *testing.T) {
	s := newTestServer(t)
	for _, c := range searchCases {
		if got := goMatches(s, c.q); !slices.Equal(got, c.want) {
			t.Errorf("search %q keeps %v, want %v", c.q, got, c.want)
		}
	}
	if got := normalizeSearch("  Chuck-will's-widow  "); got != "chuck will s widow" {
		t.Errorf("normalizeSearch = %q", got)
	}
	// Families and orders are searched by their scientific names only: the
	// English descriptions list other birds. The fixture's are "<Family>
	// family" and "<Order> (group)", so neither word may be searchable.
	for id, text := range s.search {
		for _, w := range []string{"family", "group"} {
			if strings.Contains(text, w) {
				t.Errorf("%s's search text %q includes a family's or order's English description", id, text)
			}
		}
	}
	if got := goMatches(s, "passeriformes"); len(got) == 0 {
		t.Error("no bird matches its order's scientific name")
	}
}
