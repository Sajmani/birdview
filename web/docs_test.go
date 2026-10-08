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

package web

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"
)

// repoFile reads a file from the repository root. The docs are not embedded,
// so these read from disk; a missing one fails rather than skips, because a
// README that has gone away is itself the problem T-024 is about.
func repoFile(t *testing.T, rel string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("..", rel))
	if err != nil {
		t.Fatalf("reading %s: %v", rel, err)
	}
	return string(b)
}

// userFacingDocs are the files a reader meets before the code.
func userFacingDocs(t *testing.T) map[string]string {
	t.Helper()
	out := map[string]string{"README.md": repoFile(t, "README.md")}
	paths, err := filepath.Glob(filepath.Join("..", "docs", "*.md"))
	if err != nil {
		t.Fatalf("listing docs: %v", err)
	}
	if len(paths) == 0 {
		t.Fatal("no files under docs/; T-024 expects user-facing documentation to exist")
	}
	for _, p := range paths {
		b, err := os.ReadFile(p)
		if err != nil {
			t.Fatalf("reading %s: %v", p, err)
		}
		out["docs/"+filepath.Base(p)] = string(b)
	}
	return out
}

// datasetCounts reads the compiled dataset, which is a build artifact and is
// not committed (T-005), so a machine that has not run the ingest skips.
func datasetCounts(t *testing.T) (species, families int) {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("..", "web", "data", "species.json"))
	if err != nil {
		t.Skipf("no compiled dataset; run birdview-ingest first")
	}
	var d struct {
		Families []struct{} `json:"families"`
		Species  []struct{} `json:"species"`
	}
	if err := json.Unmarshal(b, &d); err != nil {
		t.Fatalf("reading the dataset: %v", err)
	}
	return len(d.Species), len(d.Families)
}

// ccBySources are the CC BY sources whose output reaches a page, each with
// the marker that counts as crediting it. Attribution is the licence
// condition these sources ask for in return, so a page that shows their work
// and omits the credit is a licence breach, not a styling lapse (T-024).
//
// `pages` names the pages a source's output actually reaches; nil means all
// of them. GBIF is the reason this field exists: region membership is derived
// from its occurrence records and is shown on the index and the detail page,
// but compare.html shows no regions, and demanding a credit there would be
// asking for attribution for something the page does not use.
//
// This list is hand-maintained, which is its weakness: it cannot catch a
// source nobody added to it. GBIF supplied served data from the day P-011
// shipped and was missing here until 2026-10-04, so the test passed while the
// obligation went unmet. Adding a source that reaches a page is part of
// adopting it, not a later tidy-up.
var ccBySources = []struct {
	name, marker, shows string
	pages               []string
}{
	{"AviList", "AviList", "every English name, scientific name, family and species code", nil},
	{"AVONET", "AVONET", "mass, wing chord and the measured segments", nil},
	{"Fox et al.", "Fox", "empirical wingspans", nil},
	{"GBIF", "GBIF", "which regions each bird is found in",
		[]string{"index.html", "detail.html"}},
}

// TestEveryPageCreditsWhatItShows asserts the licence half of T-024.
//
// Checked **per page**, because that is how attribution works and because
// the gap it was written for was per-page: the index carried full credits
// while detail.html and compare.html omitted AviList, on pages built out of
// AviList names from the title down.
//
// Verifies: T-024. Criterion: AC-090.
func TestEveryPageCreditsWhatItShows(t *testing.T) {
	for _, page := range []string{"index.html", "detail.html", "compare.html"} {
		b, err := Static.ReadFile("static/" + page)
		if err != nil {
			t.Fatalf("reading the embedded %s: %v", page, err)
		}
		body := string(b)
		for _, src := range ccBySources {
			if src.pages != nil && !slices.Contains(src.pages, page) {
				continue
			}
			if !strings.Contains(body, src.marker) {
				t.Errorf("%s does not credit %s, which supplies %s — "+
					"CC BY 4.0 asks for attribution in return",
					page, src.name, src.shows)
			}
		}
	}
}

// TestDocsCountsMatchTheDataset asserts the staleness half of T-024.
//
// **Only totality claims are checked**, in the two shapes the docs use to
// state the size of the dataset: "N species across/in M families", and a
// bolded "**N species**" opening a definition. Every number followed by the
// word "species" is deliberately *not* checked, because most are different
// claims that a lexical rule cannot tell apart -- "533 species without an
// empirical wingspan" is a subset, "442 of 841 species" is a record of a
// measurement taken when the dataset was that size, and Bohmer's "68
// families" is somebody else's study. A check that flagged those would be
// turned off within a week.
//
// Verifies: T-024. Criterion: AC-090.
func TestDocsCountsMatchTheDataset(t *testing.T) {
	species, families := datasetCounts(t)
	pair := regexp.MustCompile(`(\d{2,4}) species (?:across|in) (\d{2,3}) families`)
	bold := regexp.MustCompile(`\*\*(\d{2,4}) species\*\*`)

	for name, body := range userFacingDocs(t) {
		for _, m := range pair.FindAllStringSubmatch(body, -1) {
			if m[1] != itoa(species) || m[2] != itoa(families) {
				t.Errorf("%s says %q, the dataset has %d species in %d families",
					name, m[0], species, families)
			}
		}
		for _, m := range bold.FindAllStringSubmatch(body, -1) {
			if m[1] != itoa(species) {
				t.Errorf("%s says %q, the dataset has %d species", name, m[0], species)
			}
		}
	}
}

// TestNoDocCreditsASourceForWhatItNoLongerSupplies asserts the third kind
// of staleness: a source still named for a job it was relieved of.
//
// The failure it was written for: README.md listed the ABA Checklist as the
// source of "species list, names, families" after CR-022 moved names and
// families to AviList. The ABA checklist supplies scope and nothing else.
//
// **Scoped to the sentence, and skipped when AviList is named in it.** A
// proximity window is not enough: the first fix for this flagged its own
// correction, because "names come from AviList, not the ABA" mentions both
// the ABA and the word it was relieved of. A sentence that names the real
// source is disambiguating, not misattributing.
//
// Verifies: T-024. Criterion: AC-090.
func TestNoDocCreditsASourceForWhatItNoLongerSupplies(t *testing.T) {
	relieved := []string{"names", "families", "species codes", "alpha code"}
	// Sentence-ish: a full stop *followed by space*, or a line break. Table
	// rows have no full stop, so the line is the unit there. Splitting on a
	// bare "." cut "ABA Checklist 8.19" in two and hid the very claim this
	// looks for -- found by watching the check fail and seeing it not.
	split := regexp.MustCompile(`\.\s|\n`)

	for name, body := range userFacingDocs(t) {
		for _, seg := range split.Split(body, -1) {
			low := strings.ToLower(seg)
			if !strings.Contains(low, "aba checklist") {
				continue
			}
			if strings.Contains(low, "avilist") {
				continue // names the real source; this is a correction, not a claim
			}
			for _, w := range relieved {
				if strings.Contains(low, w) {
					t.Errorf("%s credits the ABA Checklist with %q: %q\n"+
						"CR-022 moved that to AviList; the ABA supplies scope alone",
						name, w, strings.TrimSpace(seg))
				}
			}
		}
	}
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b []byte
	for n > 0 {
		b = append([]byte{byte('0' + n%10)}, b...)
		n /= 10
	}
	return string(b)
}
