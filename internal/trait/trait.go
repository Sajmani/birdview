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

// Package trait holds the habitat and niche vocabularies the species filter
// offers, and the handful of assignments BirdView adds to AVONET's.
//
// **The classification is AVONET's, not this project's** (product.habitatAndDietFilters). Every term
// here is one AVONET uses, and every description is AVONET's own definition,
// quoted. That is deliberate and it is the lesson of decisions.regionUrlIsPostalCode and decisions.familiesTwoWays: the
// categories this project invented had to be withdrawn one after another,
// while the postal codes and published names it merely cited did not.
//
// What BirdView adds is a plain-English **label** for each term and, for a
// few species, a **second** habitat or niche AVONET has no room for. AVONET
// records one primary value per species; a bird that genuinely lives in two
// places gets one of them. Those additions are marked BirdView and are listed in
// Supplementary below, each with a reason (product.categoryProvenance).
//
// Verifies: product.habitatAndDietFilters, product.categoryProvenance.
package trait

import (
	"strings"

	"birdview/internal/dataset"
)

// Source is cited on every published assignment.
const Source = "AVONET v7 (Tobias et al. 2022), CC BY 4.0"

// Term is one option in a vocabulary.
type Term struct {
	// Coined marks a term BirdView added because AVONET has none that fits.
	// **The bar is "no term fits", not "no term fits well".** A bird that
	// eats berries is a Frugivore even if the word is clumsy; a bird that
	// drinks tree sap is nothing in AVONET's list at all. Without that
	// limit the vocabulary drifts from a published classification into this
	// project's own, which is what decisions.familiesTwoWays was written about.
	Coined bool
	// Value is AVONET's own string, as it appears in the data. It is the
	// filter's value and must not be prettified -- the data is the key.
	Value string
	// Label is the plain-English name shown on the control. AVONET's terms
	// are precise and several are jargon: "Invertivore" and "Insessorial"
	// are not words a birder uses. Translating the word while keeping the
	// classification is a far lighter liberty than inventing the
	// classification.
	Label string
	// Description is AVONET's definition, quoted from the workbook's
	// Metadata sheet so the meaning is the publisher's and not a paraphrase
	// of mine.
	Description string
}

// Habitats in the order the control offers them: land cover from dry to wet,
// then fresh water, then salt. A reader scanning for "where would I be
// standing" moves along one axis rather than hopping.
var Habitats = []Term{
	{Value: "Desert", Label: "Desert", Description: "Drylands and other open arid habitats, often sandy with very sparse vegetation."},
	{Value: "Rock", Label: "Rock and scree", Description: "Rocky substrate typically with no or very little vegetation, including rocky outcrops, rocky coastlines, arid stony steppes, rocky mountaintops and mountain slopes."},
	{Value: "Grassland", Label: "Grassland", Description: "Open dry to moist grass-dominated landscapes, at all elevations."},
	{Value: "Shrubland", Label: "Shrubland", Description: "Low stature bushy habitats, included thornscrub, thorny or arid savanna, caatinga, xerophytic shrubland and coastal scrub."},
	{Value: "Woodland", Label: "Woodland", Description: "Medium stature tree-dominated habitats, including Acacia woodland, riparian woodlands, mangrove forests, forest edges, also more open parkland with scattered taller trees."},
	{Value: "Forest", Label: "Forest", Description: "Tall tree-dominated vegetation with more or less closed canopy, including palm forest."},
	{Value: "Human Modified", Label: "Towns and farmland", Description: "Urban landscapes, intensive agriculture, gardens."},
	{Value: "Wetland", Label: "Marsh and lake", Description: "Wide range of freshwater aquatic habitats including lakes, marshes, swamps and reedbeds."},
	{Value: "Riverine", Label: "River and stream", Description: "Associated with rivers and streams at all elevations."},
	{Value: "Coastal", Label: "Shore and estuary", Description: "Intertidal zones within immediate vicinity of beaches, estuaries, brackish to salty marshes, including mudflats, lagoons, alkaline wetlands, coastal dunes and harbours."},
	{Value: "Marine", Label: "Open sea", Description: "Pelagic, on sea near coasts, including species in the intertidal zone on beaches, and those pelagic species nesting near the sea on cliffs, islets and islands."},
}

// Niches in the order the control offers them: plant food, then animal food,
// then the two that are neither.
var Niches = []Term{
	{Value: "Granivore", Label: "Eats seeds", Description: "Species obtaining at least 60% of food resources from seeds or nuts."},
	{Value: "Frugivore", Label: "Eats fruit", Description: "Species obtaining at least 60% of food resources from fruit."},
	{Value: "Nectarivore", Label: "Drinks nectar", Description: "Species obtaining at least 60% of food resources from nectar."},
	{Value: "Herbivore terrestrial", Label: "Eats plants", Description: "Species obtaining at least 60% of food resources from other plant materials in non-aquatic systems, including leaves, buds, whole flowers etc."},
	{Value: "Herbivore aquatic", Label: "Eats water plants", Description: "Species obtaining at least 60% of food resources from plant materials in aquatic systems, including algae and aquatic plant leaves."},
	{Value: "Invertivore", Label: "Eats insects", Description: "Species obtaining at least 60% of food resources from invertebrates in terrestrial systems, including insects, worms, arachnids, etc."},
	{Value: "Vertivore", Label: "Hunts animals", Description: "Species obtaining at least 60% of food resources from vertebrate animals in terrestrial systems, including mammals, birds, reptiles etc."},
	{Value: "Aquatic predator", Label: "Hunts in water", Description: "Species obtaining at least 60% of food resources from vertebrate and invertebrate animals in aquatic systems, including fish, crustacea, molluscs, etc."},
	{Value: "Scavenger", Label: "Scavenges", Description: "Species obtaining at least 60% of food resources from carrion, offal or refuse."},
	{Value: "Omnivore", Label: "Eats anything", Description: "Species using multiple niches, within or across trophic levels, in relatively equal proportions."},
	// BirdView's own. AVONET's ten niches have no category for exudates, so
	// the four sapsuckers -- a genus named for this -- are filed under
	// Invertivore or Omnivore and the thing they are known for disappears.
	{Value: "Sap", Label: "Drinks sap", Coined: true,
		Description: "Species obtaining a large part of its food from tree sap, drilling and " +
			"maintaining wells in bark and returning to them. AVONET has no niche for exudates, " +
			"so this category is BirdView's own."},
}

// extra is one assignment BirdView adds beyond AVONET's primary value.
type extra struct {
	Habitat string
	Niche   string
	Why     string
}

// Supplementary is where BirdView disagrees with, or rather adds to, AVONET
// — keyed by scientific name.
//
// **Kept deliberately short.** AVONET's primary value is right about the
// bird's main haunt in almost every case, and a long list here would be this
// project quietly substituting its own classification for a published one,
// which is the habit decisions.familiesTwoWays was written to break. An entry earns its place
// only when AVONET's single value hides something a reader filtering by
// habitat would certainly expect to find.
//
// Every entry is rendered as BirdView, never merged into AVONET's (product.categoryProvenance).
var Supplementary = map[string][]extra{
	// AVONET codes these by where they nest or mainly feed, but each is a
	// bird a reader filtering for the other habitat would expect to see.
	"Pelecanus erythrorhynchos": {{Habitat: "Coastal", Why: "winters on sheltered coasts and estuaries, though AVONET codes its primary habitat as inland wetland"}},
	"Nannopterum auritum":       {{Habitat: "Wetland", Why: "abundant on inland lakes and rivers, though AVONET codes its primary habitat as marine"}},
	"Ardea herodias":            {{Habitat: "Coastal", Why: "forages on tidal flats and estuaries as readily as on fresh water"}},
	"Pandion haliaetus": {
		{Habitat: "Coastal", Why: "fishes estuaries and sheltered salt water as well as fresh"},
		{Habitat: "Human Modified", Why: "nests routinely on poles, pylons and purpose-built platforms"},
	},
	"Larus smithsonianus": {{Habitat: "Human Modified", Why: "feeds routinely at landfill, docks and town rooftops"}},

	// Birds whose everyday haunt is not the one AVONET records. Chimney
	// Swift and Barn Swallow were on this list too and came off it: AVONET
	// already codes both Human Modified, and an addition that restates the
	// published value adds nothing.
	"Branta canadensis":  {{Habitat: "Human Modified", Why: "parks, playing fields and golf courses are now its commonest North American habitat"}},
	"Anas platyrhynchos": {{Habitat: "Human Modified", Why: "town ponds and park lakes throughout the range"}},
	"Columba livia":      {{Habitat: "Rock", Why: "a cliff-nester by ancestry, still using cliffs and quarries away from towns"}},
	"Falco peregrinus":   {{Habitat: "Human Modified", Why: "city buildings and bridges are a primary nest site"}},

	// The sapsuckers. AVONET files them under Invertivore or Omnivore,
	// which is true and misses the point: Sphyrapicus drills wells and
	// lives off sap, and the genus is named for it.
	"Sphyrapicus varius":     {{Niche: "Sap", Why: "drills and tends sap wells; AVONET has no niche for exudates"}},
	"Sphyrapicus nuchalis":   {{Niche: "Sap", Why: "drills and tends sap wells; AVONET has no niche for exudates"}},
	"Sphyrapicus ruber":      {{Niche: "Sap", Why: "drills and tends sap wells; AVONET has no niche for exudates"}},
	"Sphyrapicus thyroideus": {{Niche: "Sap", Why: "drills and tends sap wells; AVONET has no niche for exudates"}},
}

// mixedDiet holds the two rules that read AVONET's own disagreement with
// itself.
//
// AVONET records `Trophic.Level` as well as `Trophic.Niche`. Where the level
// is **Omnivore** but the niche names one specific food, AVONET is saying
// the diet is mixed and the niche only names the larger half. That happens
// for 64 of the 839 species here, and they are the seasonal switchers: nine
// chickadees and titmice, seven thrushes, the sparrows and longspurs.
//
// The rules are written as rules, not as a list of 64 species keys. A typed
// list would need every scientific name to be right, and the first five
// hand-typed entries in Supplementary had two names AviList does not use
// (acceptance.habitatDietFilterPerSpecies).

// plantFoodOf names the plant half for a family AVONET codes Invertivore at
// an Omnivore level. **That there is a plant half is AVONET's statement;
// which one it is, is BirdView's**, so these arrive tagged birdview.
var plantFoodOf = map[string]string{
	// Cachers and seed-takers.
	"Paridae":       "Granivore", // chickadees and titmice, through winter
	"Sittidae":      "Granivore", // nuthatches cache seeds and nuts
	"Passerellidae": "Granivore", // Cassin's, Grasshopper, Nelson's Sparrow
	"Phasianidae":   "Granivore", // Lesser Prairie-Chicken
	// Berry-takers.
	"Turdidae":          "Frugivore", // robins and thrushes switch to fruit
	"Mimidae":           "Frugivore", // catbird and thrashers
	"Vireonidae":        "Frugivore", // Red-eyed Vireo is strongly frugivorous in autumn
	"Tyrannidae":        "Frugivore", // Great Crested, Scissor-tailed
	"Parulidae":         "Frugivore", // Cape May, Bay-breasted, Black-throated Green
	"Paradoxornithidae": "Frugivore", // Wrentit
	"Cardinalidae":      "Frugivore", // Flame-colored Tanager
}

// plantFoodOfSpecies overrides the family where a family does not agree
// with itself. Icteridae splits -- orioles take fruit, the meadowlark and
// the blackbirds take grain -- and Picidae splits between fruit-takers and
// Williamson's Sapsucker, which takes sap and is handled by its own
// category rather than filed under fruit.
var plantFoodOfSpecies = map[string]string{
	"Icterus galbula":        "Frugivore", // Baltimore Oriole
	"Icterus pectoralis":     "Frugivore", // Spot-breasted Oriole
	"Molothrus bonariensis":  "Granivore", // Shiny Cowbird
	"Agelaius tricolor":      "Granivore", // Tricolored Blackbird
	"Sturnella neglecta":     "Granivore", // Western Meadowlark
	"Colaptes chrysoides":    "Frugivore", // Gilded Flicker
	"Melanerpes lewis":       "Frugivore", // Lewis's Woodpecker
	"Sphyrapicus thyroideus": "",          // Williamson's Sapsucker: sap, not fruit
}

// MixedDietNiche returns the niche AVONET's two diet columns imply but do
// not state, and whether BirdView decided which it is.
//
// Returns "" when the columns agree, which is the usual case.
func MixedDietNiche(scientificName, family, niche, level string) (value string, ours bool) {
	if level != "Omnivore" || niche == "Omnivore" || niche == "" {
		return "", false
	}
	switch niche {
	case "Granivore":
		// The animal half of a small seed-eater's mixed diet is
		// invertebrates: it is not taking carrion or vertebrates. Forced by
		// AVONET's own definitions, so this one is derived, not ours.
		return "Invertivore", false
	case "Invertivore":
		if v, ok := plantFoodOfSpecies[scientificName]; ok {
			return v, v != ""
		}
		if v, ok := plantFoodOf[family]; ok {
			return v, true
		}
	}
	return "", false
}

// Derived builds the assignment for a value computed from published data by
// a stated rule.
func Derived(value, why string) dataset.Assignment {
	return dataset.Assignment{Value: value, Origin: dataset.Derived, Why: why, Doc: "data#habitat-and-diet"}
}

// Published builds the assignment for a value AVONET supplies.
func Published(value string) dataset.Assignment {
	return dataset.Assignment{Value: value, Origin: dataset.Published, Source: Source}
}

// BirdView builds the assignment for a value BirdView adds, with its reason and
// a pointer to the document that explains the practice.
func BirdView(value, why string) dataset.Assignment {
	return dataset.Assignment{Value: value, Origin: dataset.BirdView, Why: why, Doc: "data#habitat-and-diet"}
}

// HabitatsFor returns every habitat assigned to a species: AVONET's primary
// value first, then anything BirdView adds.
func HabitatsFor(scientificName, avonetHabitat string) []dataset.Assignment {
	var out []dataset.Assignment
	if avonetHabitat != "" {
		out = append(out, Published(avonetHabitat))
	}
	for _, e := range Supplementary[scientificName] {
		if e.Habitat != "" && e.Habitat != avonetHabitat {
			out = append(out, BirdView(e.Habitat, e.Why))
		}
	}
	return out
}

// NichesFor is HabitatsFor for the feeding niche, plus the mixed-diet rule.
func NichesFor(scientificName, family, avonetNiche, avonetLevel string) []dataset.Assignment {
	var out []dataset.Assignment
	if avonetNiche != "" {
		out = append(out, Published(avonetNiche))
	}
	if v, ours := MixedDietNiche(scientificName, family, avonetNiche, avonetLevel); v != "" {
		why := "AVONET's trophic level is Omnivore though its niche names " +
			strings.ToLower(LabelOf(Niches, avonetNiche)[5:]) + ", so the diet is mixed"
		if ours {
			why += "; which plant food is BirdView's reading of the family"
		} else {
			why += ", and the other half of a seed-eater's diet is invertebrates"
		}
		out = append(out, func() dataset.Assignment {
			if ours {
				return BirdView(v, why)
			}
			return Derived(v, why)
		}())
	}
	for _, e := range Supplementary[scientificName] {
		if e.Niche != "" && e.Niche != avonetNiche {
			out = append(out, BirdView(e.Niche, e.Why))
		}
	}
	return out
}

// LabelOf returns the plain-English label for a value, or the value itself
// if the vocabulary does not know it -- which a test catches rather than the
// reader seeing a blank chip.
func LabelOf(terms []Term, value string) string {
	for _, t := range terms {
		if t.Value == value {
			return t.Label
		}
	}
	return value
}
