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
// only well-keyed, they actually arrive as assignments marked Ours.
//
// Verifies: P-015, P-069. Criterion: AC-092.
func TestSupplementaryReachesTheDataset(t *testing.T) {
	sp := loadSpecies(t)
	want := 0
	for _, extras := range Supplementary {
		want += len(extras)
	}
	got := 0
	for _, s := range sp {
		for _, a := range append(append([]dataset.Assignment{}, s.Habitats...), s.Niches...) {
			if a.Origin == dataset.Ours {
				got++
			}
		}
	}
	if got != want {
		t.Errorf("%d assignments are marked Ours, the table defines %d", got, want)
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
	case dataset.Ours:
		if a.Why == "" {
			t.Errorf("%s: %s %q is ours but gives no reason", sci, what, a.Value)
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
