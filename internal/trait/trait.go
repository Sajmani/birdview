// Package trait holds the habitat and niche vocabularies the species filter
// offers, and the handful of assignments BirdView adds to AVONET's.
//
// **The classification is AVONET's, not this project's** (P-015). Every term
// here is one AVONET uses, and every description is AVONET's own definition,
// quoted. That is deliberate and it is the lesson of CR-024 and CR-025: the
// categories this project invented had to be withdrawn one after another,
// while the postal codes and published names it merely cited did not.
//
// What BirdView adds is a plain-English **label** for each term and, for a
// few species, a **second** habitat or niche AVONET has no room for. AVONET
// records one primary value per species; a bird that genuinely lives in two
// places gets one of them. Those additions are marked BirdView and are listed in
// Supplementary below, each with a reason (P-069).
//
// Verifies: P-015, P-069.
package trait

import "birdview/internal/dataset"

// Source is cited on every published assignment.
const Source = "AVONET v7 (Tobias et al. 2022), CC BY 4.0"

// Term is one option in a vocabulary.
type Term struct {
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
	{"Desert", "Desert", "Drylands and other open arid habitats, often sandy with very sparse vegetation."},
	{"Rock", "Rock and scree", "Rocky substrate typically with no or very little vegetation, including rocky outcrops, rocky coastlines, arid stony steppes, rocky mountaintops and mountain slopes."},
	{"Grassland", "Grassland", "Open dry to moist grass-dominated landscapes, at all elevations."},
	{"Shrubland", "Shrubland", "Low stature bushy habitats, included thornscrub, thorny or arid savanna, caatinga, xerophytic shrubland and coastal scrub."},
	{"Woodland", "Woodland", "Medium stature tree-dominated habitats, including Acacia woodland, riparian woodlands, mangrove forests, forest edges, also more open parkland with scattered taller trees."},
	{"Forest", "Forest", "Tall tree-dominated vegetation with more or less closed canopy, including palm forest."},
	{"Human Modified", "Towns and farmland", "Urban landscapes, intensive agriculture, gardens."},
	{"Wetland", "Marsh and lake", "Wide range of freshwater aquatic habitats including lakes, marshes, swamps and reedbeds."},
	{"Riverine", "River and stream", "Associated with rivers and streams at all elevations."},
	{"Coastal", "Shore and estuary", "Intertidal zones within immediate vicinity of beaches, estuaries, brackish to salty marshes, including mudflats, lagoons, alkaline wetlands, coastal dunes and harbours."},
	{"Marine", "Open sea", "Pelagic, on sea near coasts, including species in the intertidal zone on beaches, and those pelagic species nesting near the sea on cliffs, islets and islands."},
}

// Niches in the order the control offers them: plant food, then animal food,
// then the two that are neither.
var Niches = []Term{
	{"Granivore", "Eats seeds", "Species obtaining at least 60% of food resources from seeds or nuts."},
	{"Frugivore", "Eats fruit", "Species obtaining at least 60% of food resources from fruit."},
	{"Nectarivore", "Drinks nectar", "Species obtaining at least 60% of food resources from nectar."},
	{"Herbivore terrestrial", "Eats plants", "Species obtaining at least 60% of food resources from other plant materials in non-aquatic systems, including leaves, buds, whole flowers etc."},
	{"Herbivore aquatic", "Eats water plants", "Species obtaining at least 60% of food resources from plant materials in aquatic systems, including algae and aquatic plant leaves."},
	{"Invertivore", "Eats insects", "Species obtaining at least 60% of food resources from invertebrates in terrestrial systems, including insects, worms, arachnids, etc."},
	{"Vertivore", "Hunts animals", "Species obtaining at least 60% of food resources from vertebrate animals in terrestrial systems, including mammals, birds, reptiles etc."},
	{"Aquatic predator", "Hunts in water", "Species obtaining at least 60% of food resources from vertebrate and invertebrate animals in aquatic systems, including fish, crustacea, molluscs, etc."},
	{"Scavenger", "Scavenges", "Species obtaining at least 60% of food resources from carrion, offal or refuse."},
	{"Omnivore", "Eats anything", "Species using multiple niches, within or across trophic levels, in relatively equal proportions."},
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
// which is the habit CR-025 was written to break. An entry earns its place
// only when AVONET's single value hides something a reader filtering by
// habitat would certainly expect to find.
//
// Every entry is rendered as BirdView, never merged into AVONET's (P-069).
var Supplementary = map[string][]extra{
	// AVONET codes these by where they nest or mainly feed, but each is a
	// bird a reader filtering for the other habitat would expect to see.
	"Pelecanus erythrorhynchos": {{Habitat: "Coastal", Why: "winters on sheltered coasts and estuaries, though AVONET codes its primary habitat as inland wetland"}},
	"Nannopterum auritum":       {{Habitat: "Wetland", Why: "abundant on inland lakes and rivers, though AVONET codes its primary habitat as marine"}},
	"Ardea herodias":            {{Habitat: "Coastal", Why: "forages on tidal flats and estuaries as readily as on fresh water"}},
	"Pandion haliaetus":         {{Habitat: "Coastal", Why: "fishes estuaries and sheltered salt water as well as fresh"}},
	"Larus smithsonianus":       {{Habitat: "Human Modified", Why: "feeds routinely at landfill, docks and town rooftops"}},
}

// Published builds the assignment for a value AVONET supplies.
func Published(value string) dataset.Assignment {
	return dataset.Assignment{Value: value, Origin: dataset.Published, Source: Source}
}

// BirdView builds the assignment for a value BirdView adds, with its reason and
// a pointer to the document that explains the practice.
func BirdView(value, why string) dataset.Assignment {
	return dataset.Assignment{Value: value, Origin: dataset.BirdView, Why: why, Doc: "data"}
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

// NichesFor is HabitatsFor for the feeding niche.
func NichesFor(scientificName, avonetNiche string) []dataset.Assignment {
	var out []dataset.Assignment
	if avonetNiche != "" {
		out = append(out, Published(avonetNiche))
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
