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

package trait

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"birdview/internal/dataset"
)

const datasetFile = "../../web/data/species.json"

func loadSpecies(t *testing.T) []dataset.Species {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(datasetFile))
	if err != nil {
		t.Skipf("no compiled dataset; run birdview-ingest first")
	}
	var d struct {
		Species []dataset.Species `json:"species"`
	}
	if err := json.Unmarshal(b, &d); err != nil {
		t.Fatalf("reading the dataset: %v", err)
	}
	return d.Species
}

// TestSupplementaryKeysExist is the check this package most needs.
//
// An entry keyed on a name AviList does not use does **nothing**, silently:
// the map lookup misses, no assignment is added, no error is raised, and the
// table still looks right to a reader. Two of the first five entries were
// dead on arrival — `Phalacrocorax auritus` and `Larus argentatus`, where
// AviList says `Nannopterum auritum` and `Larus smithsonianus` — and the
// only symptom was a count being lower than intended.
//
// Verifies: P-015. Criterion: AC-092.
func TestSupplementaryKeysExist(t *testing.T) {
	sp := loadSpecies(t)
	known := map[string]bool{}
	for _, s := range sp {
		known[s.ScientificName] = true
	}
	for name := range Supplementary {
		if !known[name] {
			t.Errorf("Supplementary is keyed on %q, which is not a species in the dataset; "+
				"the entry does nothing and says nothing", name)
		}
	}
}

// TestSupplementaryReachesTheDataset closes the loop: the entries are not
// only well-keyed, each one actually arrives as an assignment.
//
// Checked **per entry**, not by a total. A total was the first version and
// broke the moment BirdView gained a second source of its own assignments
// — the mixed-diet rule — because it assumed every `birdview` tag came
// from this table. Per entry says what is actually meant: this row did
// something.
//
// Verifies: P-015, P-069. Criterion: AC-092.
func TestSupplementaryReachesTheDataset(t *testing.T) {
	byName := map[string]dataset.Species{}
	for _, s := range loadSpecies(t) {
		byName[s.ScientificName] = s
	}
	for name, extras := range Supplementary {
		sp, ok := byName[name]
		if !ok {
			continue // reported by TestSupplementaryKeysExist
		}
		for _, e := range extras {
			if e.Habitat != "" && !carries(sp.Habitats, e.Habitat) {
				t.Errorf("%s: the table adds habitat %q and the dataset does not have it",
					name, e.Habitat)
			}
			if e.Niche != "" && !carries(sp.Niches, e.Niche) {
				t.Errorf("%s: the table adds niche %q and the dataset does not have it",
					name, e.Niche)
			}
		}
	}
}

func carries(as []dataset.Assignment, value string) bool {
	for _, a := range as {
		if a.Value == value && a.Origin == dataset.BirdView {
			return true
		}
	}
	return false
}

// TestMixedDietRuleReachesTheDataset covers BirdView's other source of
// assignments: the rule that reads AVONET's trophic level against its
// niche. The rule is written as a rule rather than as 52 species keys, so
// what needs checking is that it fires and that both of its branches do.
//
// Verifies: P-015. Criterion: AC-092.
func TestMixedDietRuleReachesTheDataset(t *testing.T) {
	var derived, ours int
	for _, s := range loadSpecies(t) {
		for _, a := range s.Niches {
			if !strings.Contains(a.Why, "trophic level is Omnivore") {
				continue
			}
			switch a.Origin {
			case dataset.Derived:
				derived++
			case dataset.BirdView:
				ours++
			default:
				t.Errorf("%s: mixed-diet niche %q has origin %q", s.ScientificName, a.Value, a.Origin)
			}
		}
	}
	// Both branches must fire: the forced one, where a seed-eater's other
	// half is invertebrates, and the judged one, where BirdView picks which
	// plant food a family takes.
	if derived == 0 {
		t.Error("the derived branch of the mixed-diet rule reaches nothing")
	}
	if ours == 0 {
		t.Error("the judged branch of the mixed-diet rule reaches nothing")
	}
}

// TestEveryAssignmentHasAnOriginAndTheVocabularyKnowsIt guards the two ways
// a category can be dishonest: carrying no origin at all, which is an
// opinion wearing a finding's clothes (P-069), and carrying a value the
// control has no label for, which renders as a blank or raw-jargon chip.
//
// Verifies: P-015, P-069. Criterion: AC-092.
func TestEveryAssignmentHasAnOriginAndTheVocabularyKnowsIt(t *testing.T) {
	sp := loadSpecies(t)
	hab := map[string]bool{}
	for _, term := range Habitats {
		hab[term.Value] = true
	}
	nic := map[string]bool{}
	for _, term := range Niches {
		nic[term.Value] = true
	}

	for _, s := range sp {
		for _, a := range s.Habitats {
			checkAssignment(t, s.ScientificName, "habitat", a, hab)
		}
		for _, a := range s.Niches {
			checkAssignment(t, s.ScientificName, "niche", a, nic)
		}
	}
}

func checkAssignment(t *testing.T, sci, what string, a dataset.Assignment, vocab map[string]bool) {
	t.Helper()
	switch a.Origin {
	case dataset.Published:
		if a.Source == "" {
			t.Errorf("%s: published %s %q names no source", sci, what, a.Value)
		}
	case dataset.BirdView:
		if a.Why == "" {
			t.Errorf("%s: %s %q is birdview's but gives no reason", sci, what, a.Value)
		}
	case dataset.Derived:
		if a.Why == "" {
			t.Errorf("%s: derived %s %q states no rule", sci, what, a.Value)
		}
	default:
		t.Errorf("%s: %s %q has origin %q; an unmarked category is an opinion "+
			"presented as a finding", sci, what, a.Value, a.Origin)
	}
	if !vocab[a.Value] {
		t.Errorf("%s: %s %q is not in the control's vocabulary, so it can be "+
			"assigned but never selected", sci, what, a.Value)
	}
}

// TestLabelsAreNotJargon checks the one liberty this package takes. AVONET's
// terms are precise and several are unreadable; the label is a translation,
// so it must actually differ from the term where the term is jargon, and
// every description must be AVONET's own.
//
// Verifies: P-015. Criterion: AC-092.
func TestLabelsAreNotJargon(t *testing.T) {
	jargon := []string{"Invertivore", "Vertivore", "Granivore", "Frugivore",
		"Nectarivore", "Herbivore terrestrial", "Herbivore aquatic"}
	for _, j := range jargon {
		if LabelOf(Niches, j) == j {
			t.Errorf("%q is shown to the reader untranslated", j)
		}
	}
	for _, set := range [][]Term{Habitats, Niches} {
		for _, term := range set {
			if term.Label == "" || term.Description == "" {
				t.Errorf("%q has no label or no description", term.Value)
			}
			if !strings.HasSuffix(term.Description, ".") {
				t.Errorf("%q's description is not a sentence: %q", term.Value, term.Description)
			}
		}
	}
}

// TestCoinedTermsAreRareAndSaySo guards the one place BirdView adds to
// AVONET's vocabulary rather than to its assignments.
//
// The bar is **no AVONET term fits**, not "none fits well". A bird that
// eats berries is a Frugivore even if the word is clumsy; a bird that
// drinks tree sap is nothing in AVONET's list at all. Without that limit
// the vocabulary drifts from a published classification into this
// project's own, which is the habit CR-025 was written about — so the
// count is capped and a coined term must declare itself.
//
// Verifies: P-015. Criterion: AC-092.
func TestCoinedTermsAreRareAndSaySo(t *testing.T) {
	coined := 0
	for _, set := range [][]Term{Habitats, Niches} {
		for _, term := range set {
			if !term.Coined {
				continue
			}
			coined++
			if !strings.Contains(term.Description, "BirdView") {
				t.Errorf("%q is BirdView's own but its definition does not say so", term.Value)
			}
		}
	}
	if coined == 0 {
		t.Error("no coined term; this check is not exercising anything")
	}
	if coined > 2 {
		t.Errorf("%d coined terms; adding to a published vocabulary is for the gap no "+
			"term fills, not a habit", coined)
	}
}

// TestCoinedTermsAreNeverPublished closes the loop: a category of ours must
// never reach a species as though a source had said it.
//
// Verifies: P-015, P-069. Criterion: AC-092.
func TestCoinedTermsAreNeverPublished(t *testing.T) {
	ours := map[string]bool{}
	for _, set := range [][]Term{Habitats, Niches} {
		for _, term := range set {
			if term.Coined {
				ours[term.Value] = true
			}
		}
	}
	seen := 0
	for _, s := range loadSpecies(t) {
		for _, a := range append(append([]dataset.Assignment{}, s.Habitats...), s.Niches...) {
			if !ours[a.Value] {
				continue
			}
			seen++
			if a.Origin != dataset.BirdView {
				t.Errorf("%s carries the coined category %q with origin %q; a category "+
					"BirdView invented cannot be published by anyone",
					s.ScientificName, a.Value, a.Origin)
			}
		}
	}
	if seen == 0 {
		t.Error("no species carries a coined category; the term exists but reaches nothing")
	}
}
