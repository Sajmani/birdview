package family

import (
	"encoding/json"
	"os"
	"sort"
	"testing"
)

// datasetFile is the compiled dataset. The family list comes from AviList, so
// the grouping is checked against **the real list** rather than a fixture: a
// fixture would let the table and the taxonomy drift apart silently, and
// CR-022 renamed 55 of 90 English family names in a single change.
//
// The dataset is a build artifact and is not committed (T-005), so a machine
// that has not run the ingest skips rather than fails.
const datasetFile = "../../web/data/species.json"

func datasetFamilies(t *testing.T) map[string]string {
	t.Helper()
	b, err := os.ReadFile(datasetFile)
	if err != nil {
		t.Skipf("no compiled dataset at %s; run birdview-ingest first", datasetFile)
	}
	var d struct {
		Families []struct {
			Name   string `json:"name"`
			Common string `json:"common"`
		} `json:"families"`
	}
	if err := json.Unmarshal(b, &d); err != nil {
		t.Fatalf("reading the dataset: %v", err)
	}
	out := map[string]string{}
	for _, f := range d.Families {
		out[f.Name] = f.Common
	}
	if len(out) == 0 {
		t.Fatal("the dataset lists no families")
	}
	return out
}

// TestEveryFamilyIsInAtLeastOneGroup is the invariant the selector needs:
// the birder groups **cover** the families. A family in no group is
// unreachable from that block.
//
// It is coverage, not a partition. A family may be in two groups on purpose
// — the owls are birds of prey and birds of the night, and a reader looking
// under either should find them (CR-026). What would be wrong is a family in
// *no* group, or the same family listed twice in one group, which panics at
// init.
//
// Checked against the dataset in both directions. A family AviList adds and
// the table has not been told about fails here, which is the point — the
// alternative is a bird quietly missing from a control nobody is watching.
//
// Verifies: P-014. Criterion: AC-089.
func TestEveryFamilyIsInAtLeastOneGroup(t *testing.T) {
	ds := datasetFamilies(t)

	for name, common := range ds {
		gs := GroupsOf(name)
		if len(gs) == 0 {
			t.Errorf("%s (%s) is in the dataset but in no birder group", name, common)
			continue
		}
		for _, g := range gs {
			if g == "" {
				t.Errorf("%s (%s) is filed under an empty group name", name, common)
			}
		}
	}
	for _, g := range Groups {
		if len(g.Families) == 0 {
			t.Errorf("group %q holds no families", g.Name)
		}
		for _, f := range g.Families {
			if _, known := ds[f]; !known {
				t.Errorf("group %q lists %s, which is not a family in the dataset", g.Name, f)
			}
		}
	}
	if Len() != len(ds) {
		t.Errorf("the table files %d families, the dataset has %d", Len(), len(ds))
	}
}

// TestGroupNamesAreUniqueAndOrdered checks the offer order is a real
// sequence: duplicate group names would render two blocks a reader cannot
// tell apart, and a toggle would select the wrong one.
//
// Verifies: P-014. Criterion: AC-089.
func TestGroupNamesAreUniqueAndOrdered(t *testing.T) {
	seen := map[string]bool{}
	for _, g := range Groups {
		if g.Name == "" {
			t.Error("a group has no name")
		}
		if seen[g.Name] {
			t.Errorf("two groups are both called %q", g.Name)
		}
		seen[g.Name] = true
	}
	names := Names()
	if len(names) != len(Groups) {
		t.Errorf("Names() lists %d groups, there are %d", len(names), len(Groups))
	}
	for i, g := range Groups {
		if names[i] != g.Name {
			t.Errorf("Names()[%d] is %q, want %q — the offer order is not the table order",
				i, names[i], g.Name)
		}
	}
	// A family in two groups is deliberate but must be rare; if it became
	// common the block would be a tag cloud rather than a grouping.
	multi := 0
	for _, g := range Groups {
		for _, f := range g.Families {
			if len(GroupsOf(f)) > 1 {
				multi++
			}
		}
	}
	if multi > 6 {
		t.Errorf("%d placements are duplicates; overlap is for the rare family that "+
			"genuinely belongs twice, not a general licence", multi)
	}

	// InGroup round-trips, and reports nothing for a group that does not exist.
	for _, g := range Groups {
		got := InGroup(g.Name)
		want := append([]string(nil), g.Families...)
		sort.Strings(want)
		if len(got) != len(want) {
			t.Errorf("InGroup(%q) returned %d families, want %d", g.Name, len(got), len(want))
			continue
		}
		for i := range got {
			if got[i] != want[i] {
				t.Errorf("InGroup(%q)[%d] = %q, want %q", g.Name, i, got[i], want[i])
			}
		}
	}
	if InGroup("no such group") != nil {
		t.Error("InGroup invented a group that does not exist")
	}
}

// TestBirderGroupsDisagreeWithTaxonomy is the case for having a second
// grouping at all. Two groupings that agreed would render two near-identical
// blocks and tell a reader nothing; P-014 rests on them disagreeing, and the
// measured disagreement is 76 of 90 families.
//
// The threshold is "most", not the measured number: pinning 76 would record
// what this taxonomy happens to do, and AviList moves families between orders
// (CR-022 moved four).
//
// Verifies: P-014. Criterion: AC-089.
func TestBirderGroupsDisagreeWithTaxonomy(t *testing.T) {
	b, err := os.ReadFile(datasetFile)
	if err != nil {
		t.Skipf("no compiled dataset at %s; run birdview-ingest first", datasetFile)
	}
	var d struct {
		Families []struct {
			Name  string `json:"name"`
			Order string `json:"order"`
		} `json:"families"`
	}
	if err := json.Unmarshal(b, &d); err != nil {
		t.Fatalf("reading the dataset: %v", err)
	}

	// A family is filed "the same way" when its order maps onto exactly one
	// birder group and that group draws from exactly one order.
	groupsOfOrder := map[string]map[string]bool{}
	ordersOfGroup := map[string]map[string]bool{}
	for _, f := range d.Families {
		for _, g := range GroupsOf(f.Name) {
			if groupsOfOrder[f.Order] == nil {
				groupsOfOrder[f.Order] = map[string]bool{}
			}
			if ordersOfGroup[g] == nil {
				ordersOfGroup[g] = map[string]bool{}
			}
			groupsOfOrder[f.Order][g] = true
			ordersOfGroup[g][f.Order] = true
		}
	}
	same := 0
	for _, f := range d.Families {
		gs := GroupsOf(f.Name)
		if len(gs) != 1 {
			continue // a family in two groups is by construction not filed "the same way"
		}
		if len(groupsOfOrder[f.Order]) == 1 && len(ordersOfGroup[gs[0]]) == 1 {
			same++
		}
	}
	if same*2 >= len(d.Families) {
		t.Errorf("the two groupings file %d of %d families identically; "+
			"a birder grouping that mostly repeats taxonomy earns nobody a second block",
			same, len(d.Families))
	}
}
