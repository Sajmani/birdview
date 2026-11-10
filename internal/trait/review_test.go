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
	"strings"
	"testing"

	"birdview/internal/dataset"
)

// TestHabitatReviewParsesAndApplies asserts decisions.habitatReviewedPerSpecies's table: a reviewed row's
// additions reach a species' habitats as BirdView's, each with its reason; a proposed or pending
// row's do not; and a row that names a habitat outside AVONET's vocabulary, a reviewed row without
// a reviewer or date, an addition without a reason, or an unknown status is refused.
//
// Verifies: product.habitatAndDietFilters. Criterion: acceptance.habitatReviewComplete.
func TestHabitatReviewParsesAndApplies(t *testing.T) {
	const table = "# comment\n" +
		"Calidris minutilla\tLeast Sandpiper\tWetland\treviewed\tSameer\t2026-11-07\tCoastal: winters on mudflats\n" +
		"Calidris pusilla\tSemipalmated Sandpiper\tGrassland\tproposed\t\t\tCoastal: migrates on beaches\n" +
		"Turdus migratorius\tAmerican Robin\tWoodland\tpending\t\t\t\n"
	rev, err := ParseHabitatReview(strings.NewReader(table))
	if err != nil {
		t.Fatal(err)
	}
	if len(rev) != 3 || rev["Calidris minutilla"].AVONET != "Wetland" {
		t.Fatalf("read %+v", rev)
	}
	got := HabitatsFor("Calidris minutilla", "Wetland", rev)
	if len(got) != 2 || got[0].Origin != dataset.Published || got[1].Value != "Coastal" ||
		got[1].Origin != dataset.BirdView || got[1].Why != "winters on mudflats" {
		t.Errorf("Least Sandpiper, reviewed, has %+v; want AVONET's Wetland and BirdView's Coastal with its reason", got)
	}
	if got := HabitatsFor("Calidris pusilla", "Grassland", rev); len(got) != 1 {
		t.Errorf("Semipalmated Sandpiper, only proposed, has %+v; a proposal does not reach the page", got)
	}
	for _, bad := range []string{
		"Calidris minutilla\tLeast Sandpiper\tWetland\treviewed\tSameer\t2026-11-07\tBeach: on beaches\n",
		"Calidris minutilla\tLeast Sandpiper\tWetland\treviewed\t\t\tCoastal: winters on mudflats\n",
		"Calidris minutilla\tLeast Sandpiper\tWetland\treviewed\tSameer\t2026-11-07\tCoastal\n",
		"Calidris minutilla\tLeast Sandpiper\tWetland\tapproved\tSameer\t2026-11-07\t\n",
		"Calidris minutilla\tLeast Sandpiper\tWetland\n",
	} {
		if _, err := ParseHabitatReview(strings.NewReader(bad)); err == nil {
			t.Errorf("accepted the row %q", strings.TrimSpace(bad))
		}
	}
}

// TestHabitatGroupsPartitionAVONET asserts decisions.habitatGroupShortcuts and decisions.openOceanSplit:
// the five groups hold every one of AVONET's habitats exactly once, Marine as its two refinements, in
// the order and with the labels the requirement gives; the review accepts a refinement as an addition.
//
// Verifies: product.habitatAndDietFilters. Criterion: acceptance.habitatGroupsFilter.
func TestHabitatGroupsPartitionAVONET(t *testing.T) {
	var got []string
	seen := map[string]string{}
	for _, g := range HabitatGroups {
		got = append(got, g.Label+"="+strings.Join(g.Terms, ","))
		for _, term := range g.Terms {
			if other, dup := seen[term]; dup {
				t.Errorf("%s is in both %s and %s", term, other, g.Label)
			}
			seen[term] = g.Label
			if GroupOf(term) != g.Label {
				t.Errorf("GroupOf(%q) = %q, want %q", term, GroupOf(term), g.Label)
			}
		}
	}
	for _, h := range Habitats {
		if h.Value == "Marine" {
			if seen[h.Value] != "" {
				t.Error("Marine is offered in a group itself; it is offered as its refinements")
			}
			continue
		}
		if seen[h.Value] == "" {
			t.Errorf("AVONET's %s is in no group", h.Value)
		}
	}
	if LabelOf(HabitatTerms, "Marine-offshore") != "Open ocean" || LabelOf(HabitatTerms, "Marine-inshore") != "Sea near shore" {
		t.Error("the refinements are not labeled Open ocean and Sea near shore")
	}
	src := "# h\n# scientific name\tname\tAVONET habitat\tstatus\treviewer\tdate\tadditions\n" +
		"Xema sabini\tSabine's Gull\tMarine\treviewed\tSameer\t2026-11-09\tMarine-offshore: its shore records are breeding tundra and inland reservoirs\n"
	if _, err := ParseHabitatReview(strings.NewReader(src)); err != nil {
		t.Errorf("a reviewed Marine-offshore addition was refused: %v", err)
	}
	want := "Shore and sea=Coastal,Marine-inshore,Marine-offshore|Fresh water=Wetland,Riverine|Woods=Forest,Woodland|Open country=Grassland,Shrubland,Desert,Rock|Towns and farms=Human Modified"
	if strings.Join(got, "|") != want {
		t.Errorf("groups are %v, want %s", got, want)
	}
}

// TestDietGroupsPartitionNiches asserts product.habitatAndDietFilters' diet groups: AVONET's trophic
// levels, Plants and Animals, then Scavenges and Eats anything, holding every niche exactly once.
//
// Verifies: product.habitatAndDietFilters. Criterion: acceptance.habitatGroupsFilter.
func TestDietGroupsPartitionNiches(t *testing.T) {
	var got []string
	seen := map[string]bool{}
	for _, g := range DietGroups {
		got = append(got, g.Label+"="+strings.Join(g.Terms, ","))
		for _, term := range g.Terms {
			if seen[term] {
				t.Errorf("%s is in two diet groups", term)
			}
			seen[term] = true
		}
	}
	for _, n := range Niches {
		if !seen[n.Value] {
			t.Errorf("the niche %s is in no diet group", n.Value)
		}
	}
	want := "Plants=Granivore,Frugivore,Nectarivore,Herbivore terrestrial,Herbivore aquatic,Sap|Animals=Invertivore,Vertivore,Aquatic predator|Scavenges=Scavenger|Eats anything=Omnivore"
	if strings.Join(got, "|") != want {
		t.Errorf("diet groups are %v, want %s", got, want)
	}
}
