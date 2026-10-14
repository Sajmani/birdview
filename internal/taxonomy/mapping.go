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

// Package taxonomy maps current AOS/NACC names onto the names AVONET filed its
// measurements under.
//
// ABA 8.19 follows AOS/NACC; AVONET2_eBird carries the eBird/Clements taxonomy
// of 2022. Four years of revision separate them, so 22 of BirdView's 841
// in-scope species do not join on name alone. This table is the project's own
// work and is committed (P-006).
//
// The distinction between the two maps below is not cosmetic. A Rename is the
// same birds under a different label, so the measurement stays `measured`. A
// NearMatch borrows a different taxon's measurement, so it must be labeled
// `near-match` and the borrowed taxon named (P-062, CR-006).
//
// Verifies: P-006, P-007.
package taxonomy

import (
	"sort"
	"strings"
)

// Renames map a current name to the pre-revision names AVONET files the same
// taxon under, in any of its three taxonomies. The specimens are the right
// birds; only the label moved, so values resolved this way stay `measured`.
//
// Several entries exist because AVONET's taxonomies disagree with each other as
// well as with AOS: Setophaga aestiva is absent from the eBird and BirdLife
// sheets but present in AVONET3_BirdTree as Dendroica aestiva.
var Renames = map[string][]string{
	"Anarhynchus mongolus": {"Charadrius mongolus"},
	"Anarhynchus montanus": {"Charadrius montanus"},
	"Anarhynchus nivosus":  {"Charadrius nivosus"},
	"Anarhynchus wilsonia": {"Charadrius wilsonia"},
	"Astur cooperii":       {"Accipiter cooperii"},
	"Botaurus exilis":      {"Ixobrychus exilis"},
	"Ramosomyia violiceps": {"Leucolia violiceps", "Amazilia violiceps"},
	"Spilopelia chinensis": {"Streptopelia chinensis"},
	"Setophaga aestiva":    {"Dendroica aestiva"},
}

// NearMatches map a species with no row of its own, in any AVONET taxonomy,
// onto the pre-split parent whose measurements stand in for it. Every value
// derived through this map is labeled near-match and names the parent.
//
// Larus vegae borrows from Larus smithsonianus, which no eBird row carries:
// HBW-BirdLife v5, the taxonomy of AVONET1_BirdLife, files vegae as a
// subspecies of it. American Herring Gull draws from the same row, so the two
// are drawn alike and both pages say so (CR-040, P-079).
var NearMatches = map[string]string{
	"Larus vegae":         "Larus smithsonianus",
	"Numenius hudsonicus": "Numenius phaeopus",
	"Sturnella lilianae":  "Sturnella magna",
	"Vireo swainsoni":     "Vireo gilvus",
	"Sula brewsteri":      "Sula leucogaster",
	"Anthus japonicus":    "Anthus rubescens",

	// Splits that look like renames, and are not. Each of these resolves to an
	// AVONET row representing the COMBINED pre-split taxon, so the measurement
	// is not of this species. They appear as near-matches because that is the
	// conservative and accurate label.
	//
	// CR-006 left their classification open pending a human taxonomic check and
	// said explicitly that each "needs a human taxonomic check before entering
	// the mapping table". Labeling them near-match takes the safe side without
	// closing that question; it does not resolve it.
	"Astur atricapillus": "Accipiter gentilis", // American from Eurasian Goshawk
	"Ardea ibis":         "Bubulcus ibis",      // Western from Eastern Cattle-Egret
	"Gygis candida":      "Gygis alba",         // split from White Tern
	"Tyto furcata":       "Tyto alba",          // American from Western Barn Owl
}

// NearMatchSheets names the AVONET sheet a near-match must borrow from, where
// one holds a narrower concept of the parent than the eBird sheet the lookup
// would otherwise reach first. The value is an avonet sheet name, and there is
// no fallback: if that sheet lacks the parent the build fails rather than
// quietly taking the broader row.
//
// AVONET2_eBird's Gygis alba still contains all three white terns AviList
// separates. AVONET1_BirdLife's excludes the small Little White Tern, so its
// bill and tail are nearer Blue-billed White-Tern's (CR-039).
//
// Verifies: P-060, P-062.
var NearMatchSheets = map[string]string{
	"Gygis candida": "AVONET1_BirdLife",
}

// OwnRowSheets names the AVONET sheet a species' own row must be read from,
// where the eBird sheet, which the lookup tries first, files a broader concept
// under the same name. As with NearMatchSheets there is no fallback.
//
// AVONET2_eBird's Larus argentatus is the 2022 herring gull, averaging 16
// specimens of which 6 are European; AVONET1_BirdLife's is those 6 (CR-041).
//
// Verifies: P-060, P-062.
var OwnRowSheets = map[string]string{
	"Larus argentatus": "AVONET1_BirdLife",
}

// AwaitingTaxonomicCheck lists the entries whose rename-versus-split status is
// still open, so the ingest can report them rather than let the question fade.
//
// Verifies: P-007.
var AwaitingTaxonomicCheck = []string{
	"Astur atricapillus", "Ardea ibis", "Gygis candida", "Tyto furcata",
}

// AviListRenames maps the name BirdView carries onto the name AviList files
// the SAME taxon under. These are one-to-one renamings -- a genus move or a
// corrected epithet -- so the name, code and family resolved through them
// belong to this bird and nothing is borrowed (CR-022).
//
// This is a second synonym table and not an extension of Renames above,
// because the two point in opposite directions: Renames reaches BACKWARD from
// a current name to the older label AVONET measured under, while this reaches
// FORWARD from BirdView's name to a newer label AviList adopted. Merging them
// would make the direction of each entry a matter of memory.
//
// Verifies: P-004.
var AviListRenames = map[string]string{
	"Coccothraustes vespertinus": "Hesperiphona vespertina",
	"Dryobates albolarvatus":     "Leuconotopicus albolarvatus",
	"Dryobates arizonae":         "Leuconotopicus arizonae",
	"Dryobates borealis":         "Leuconotopicus borealis",
	"Dryobates villosus":         "Leuconotopicus villosus",
	"Icterus bullockii":          "Icterus bullockiorum",
	"Pica nuttalli":              "Pica nuttallii",
}

// Counted is one binomial the OCCURRENCE DATA publishes whose records are,
// under the taxonomy BirdView serves, a different AviList species' birds.
type Counted struct {
	// Published is the binomial eBird or iNaturalist files them under.
	Published string
	// Species is the AviList species whose birds those records are.
	Species string
	// Why names the databases and says which way the disagreement runs,
	// for the page to show. Per entry, because the direction is not the
	// same in all of them.
	Why string
}

// CountedAs holds the names whose records BirdView counts as another
// species, because the sources and AviList draw the species boundary in
// different places.
//
// **Not "lumps", and the word is avoided deliberately.** A lump is a formal
// taxonomic act, the merging of two species into one, and only the third
// entry below is one — AviList declines to recognize Stejneger's Stonechat.
// The other two are the reverse: AviList SPLIT a species and eBird has not
// adopted the split, so eBird goes on filing the daughter's birds under the
// parent's name. Calling all three lumps names the wrong event for two of
// them, and calling the table LumpedInto said so in the code for months.
//
// What they share is not the taxonomy but the arithmetic: a name the
// occurrence data publishes, and an AviList species whose records those are.
//
// eBird is the deeper and better-curated source and supplies the evidence;
// AviList supplies the names (CR-022). This table is where the two meet.
//
// **Why these and not every boundary disagreement.** The first two published
// names are also the AVONET donor for the species BirdView carries -- see
// NearMatches -- so admitting them as species of their own would draw the
// same bird twice, with byte-identical length and mass, while implying there
// are two. That is not extra information; it reads as a bug. Where both
// halves have measurements of their own, both are kept.
//
// **Setophaga petechia is deliberately absent.** eBird's ABA-area records
// under it are overwhelmingly Setophaga aestiva, but Mangrove Yellow Warbler
// is a species BirdView carries with an AVONET row of its own, so merging
// would delete a real bird. The misattribution does not change either one's
// scope -- both clear the bar on their own records -- so it is recorded and
// left alone rather than fixed by guesswork. The same holds for Trogon:
// BirdView carries elegans, AviList splits ambiguus off it, and both would
// have their own measurements.
//
// Verifies: P-001, P-077.
var CountedAs = []Counted{{
	Published: "Numenius phaeopus",
	Species:   "Numenius hudsonicus",
	Why: "AviList splits the Whimbrel and eBird has not adopted the split: eBird files " +
		"556,569 of these birds under Numenius phaeopus, the name AviList keeps for the " +
		"Eurasian bird, and holds none at all under Numenius hudsonicus. iNaturalist " +
		"follows the split, with 17,923 records under hudsonicus and nine under phaeopus.",
}, {
	Published: "Gygis alba",
	Species:   "Gygis candida",
	Why: "AviList treats Gygis candida as a species separate from Gygis alba and eBird " +
		"does not: eBird files 16,008 of these birds under Gygis alba and holds none " +
		"under candida, while iNaturalist follows AviList and holds 1,445 under candida.",
}, {
	// This one runs the other way: AviList lumps where both databases split.
	// Without it the species is dropped for having no name BirdView may
	// publish, and maurus is left with the two records filed under it
	// directly, which is below any bar.
	Published: "Saxicola stejnegeri",
	Species:   "Saxicola maurus",
	Why: "AviList does not recognize Stejneger's Stonechat as a species separate from the " +
		"Siberian Stonechat, and both databases do: eBird holds 95 records under Saxicola " +
		"stejnegeri and iNaturalist 28, against two and none under Saxicola maurus. Under " +
		"AviList's taxonomy those are maurus's birds.",
}}

// SpeciesCountedAs returns the AviList species whose birds the records under
// a published binomial are. The key is matched case-insensitively, because
// the occurrence data arrives lowercased.
func SpeciesCountedAs(published string) (species, why string, ok bool) {
	for _, c := range CountedAs {
		if strings.EqualFold(c.Published, published) {
			return c.Species, c.Why, true
		}
	}
	return "", "", false
}

// CountedInto returns the published names whose records are counted as an
// AviList species, for the page to disclose. Sorted, because map order is
// not an order and a page that reshuffles between runs records nothing.
func CountedInto(aviList string) []Counted {
	var out []Counted
	for _, c := range CountedAs {
		if c.Species == aviList {
			out = append(out, c)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Published < out[j].Published })
	return out
}

// AviListRenamedFrom is AviListRenames inverted: the name AVONET and the
// occurrence data use, given the AviList name BirdView carries.
//
// Scope now arrives already mapped into AviList's namespace, so the AVONET
// lookup sees "Pica nuttallii" where the workbook has "Pica nuttalli" and
// finds nothing. It is the same bird; only the label moved forward, so a
// value resolved through here is still `measured`.
//
// Verifies: P-006.
var AviListRenamedFrom = func() map[string]string {
	m := make(map[string]string, len(AviListRenames))
	for from, to := range AviListRenames {
		m[to] = from
	}
	return m
}()

// NotInAviList lists species AviList does not recognize at all, because it
// lumps them into a taxon BirdView already carries. They are DROPPED from the
// dataset rather than kept under an ABA name (CR-022).
//
// That follows from the decision itself: if AviList supplies the names, it
// supplies the species concepts too, and keeping a bird AviList does not
// recognize would mean keeping the one name the project may not redistribute.
//
// Lumping them in place is not an option. Vireo swainsoni would resolve to
// Vireo gilvus, which is already in scope as the Eastern Warbling-Vireo, so
// two BirdView species would share the code warvir and therefore one URL.
//
// Listed rather than inferred, so the ingest asserts the count and a third
// species going quiet is a build failure rather than a silent deletion.
//
// Verifies: P-004.
var NotInAviList = map[string]string{
	"Porphyrio poliocephalus": "AviList lumps it into Porphyrio porphyrio, which is out of scope",
	"Vireo swainsoni":         "AviList lumps it into Vireo gilvus, already in scope as warvir",
}

// OrderCommonNames gives each order an English label for the family selector.
//
// **These are the project's own editorial labels, not sourced data.** No input
// BirdView uses publishes vernacular names for orders: the ABA checklist names
// families only, and AVONET carries bare scientific orders. They are written to
// be recognizable to a birder rather than to be taxonomically complete, so
// several are deliberately loose -- Charadriiformes really is shorebirds plus
// gulls plus auks, and no short phrase covers it exactly.
//
// Anything displayed from this table is a convenience label. Nothing is derived
// from it, and no measurement depends on it (G-004).
var OrderCommonNames = map[string]string{
	// Arrived with the derived scope (CR-028): one feral Emu population
	// recurs often enough to qualify. Named for the family a reader would
	// recognize rather than for the order's full worldwide content.
	"Casuariiformes":      "Emus and Cassowaries",
	"Anseriformes":        "Waterfowl",
	"Galliformes":         "Gamebirds",
	"Phoenicopteriformes": "Flamingos",
	"Podicipediformes":    "Grebes",
	"Pterocliformes":      "Sandgrouse",
	"Columbiformes":       "Pigeons and Doves",
	"Cuculiformes":        "Cuckoos",
	"Caprimulgiformes":    "Nightjars",
	"Apodiformes":         "Swifts and Hummingbirds",
	"Gruiformes":          "Cranes, Rails, and Allies",
	"Charadriiformes":     "Shorebirds, Gulls, and Auks",
	"Phaethontiformes":    "Tropicbirds",
	"Gaviiformes":         "Loons",
	"Procellariiformes":   "Tubenoses",
	"Ciconiiformes":       "Storks",
	"Suliformes":          "Boobies, Cormorants, and Allies",
	"Pelecaniformes":      "Pelicans, Herons, and Ibises",
	"Accipitriformes":     "Hawks and Eagles",
	// AviList separates the New World vultures into their own order, where
	// ABA and AVONET had them among the hawks (CR-022). Accipitriformes is
	// relabelled to match: it no longer contains a vulture.
	"Cathartiformes": "New World Vultures",
	"Strigiformes":   "Owls",
	"Trogoniformes":  "Trogons",
	"Coraciiformes":  "Kingfishers and Allies",
	"Piciformes":     "Woodpeckers",
	"Falconiformes":  "Falcons",
	"Psittaciformes": "Parrots",
	"Passeriformes":  "Perching Birds",
}

// SharedNames records a binomial that ONE occurrence source uses for what
// AviList treats as TWO species BirdView draws.
//
// This is a third kind of disagreement, and the dangerous one. An alias
// (AviListRenames) fails by missing: the join finds nothing and a species
// shows zero, which is visible. A name AviList does not carry (CountedAs)
// fails the same way. A shared name **fails by matching** -- the string is
// identical on both sides, the join succeeds, and every record lands on
// whichever daughter inherited the name.
//
// No name-matching check can catch it. data/gbif-taxa.tsv already files
// `Setophaga petechia -> Setophaga petechia` as tier `exact`, and it is
// exact: the strings agree and denote different birds. Finding these took a
// sweep of all 1,083 species against both databases, comparing what each
// reports under each name (CR-034).
//
// The records are NOT divided between the two species. eBird distinguishes
// none of them, so any division would be invented -- 6,245,227 records for
// the warblers alone, growing by about 827,000 a year. Both species carry
// the shared count and both say whose it also is (P-077, G-005, G-006).
//
// Verifies: P-077.
type SharedName struct {
	// Source is the binomial the occurrence data publishes.
	Source string
	// Covers are the AviList species it spans, in the order AviList lists
	// them. Both are drawn; neither owns the records alone.
	Covers [2]string
	// Why names the database and the disagreement, for the page to show.
	Why string
}

var SharedNames = []SharedName{{
	Source: "Setophaga petechia",
	Covers: [2]string{"Setophaga aestiva", "Setophaga petechia"},
	Why: "eBird has not adopted AviList's split of the Yellow Warbler and files both the " +
		"northern and the mangrove bird under Setophaga petechia, the pre-split name. It " +
		"holds no record at all under Setophaga aestiva, and filed 827,098 under the shared " +
		"name in 2024 alone.",
}, {
	Source: "Trogon elegans",
	Covers: [2]string{"Trogon ambiguus", "Trogon elegans"},
	Why: "eBird treats the Coppery-tailed Trogon as part of the Elegant Trogon and files " +
		"both under Trogon elegans. It holds no record under Trogon ambiguus; iNaturalist, " +
		"which follows the split, holds 1,180.",
}}

// SharedWith returns the other AviList species a shared name covers, and the
// reason, for an AviList name. Empty when the species' records are its own.
func SharedWith(aviList string) (other string, source string, why string, ok bool) {
	for _, s := range SharedNames {
		switch aviList {
		case s.Covers[0]:
			return s.Covers[1], s.Source, s.Why, true
		case s.Covers[1]:
			return s.Covers[0], s.Source, s.Why, true
		}
	}
	return "", "", "", false
}

// Binomial reduces a name the occurrence data publishes to the species it
// identifies, and reports whether it identifies one at all.
//
// iNaturalist publishes each identification at the rank the observer reached,
// so its facet mixes binomials with subspecies trinomials, domestic forms,
// hybrids and bare genera. eBird does not — every one of its 1,382 ABA-area
// names is already a binomial — so this only ever changes the iNaturalist
// side.
//
// A hybrid is not one species and a bare genus was never identified to one,
// so neither reduces. Anything longer is truncated to its first two tokens,
// which folds `setophaga coronata auduboni` into its species and `columba
// livia domestica` into Rock Pigeon.
//
// **The hybrid test runs before the truncation, and must.** Truncated to two
// tokens, `larus smithsonianus × hyperboreus` reads as a species BirdView
// draws, and would credit the American Herring Gull with records of a bird
// that is demonstrably not one.
//
// Measured 2026-10-04 over the ABA area: 97.1% of iNaturalist's records are
// already bare binomials, 2.6% sit under trinomials, 0.2% are hybrids and
// 0.1% are genus-only.
//
// Verifies: P-011, T-027.
func Binomial(name string) (string, bool) {
	if strings.ContainsRune(name, '×') {
		return "", false
	}
	f := strings.Fields(name)
	if len(f) < 2 {
		return "", false
	}
	return f[0] + " " + f[1], true
}

// Occurrence maps a name the occurrence data publishes onto the AviList name
// BirdView serves, lowercased for joining.
//
// One function so that every reader of the occurrence data normalizes the
// same way. birdview-status did this and birdview-regions did not, which is
// how the two committed tables came to use different taxonomies: eBird's
// 7,310,342 Dryobates villosus records matched no drawn species and were
// discarded, leaving the Hairy Woodpecker with iNaturalist's 0.7% of its own
// record (CR-035).
//
// The three steps run in one order and it is not arbitrary. Rank first, so a
// trinomial is a species before any table is consulted — both tables are
// keyed on binomials and would walk straight past `dryobates villosus
// harrisi`. Then the genus move, then the lump, each of which is keyed on
// the name the occurrence data publishes rather than on the other's output.
//
// Verifies: P-011, P-074.
func Occurrence(name string) string {
	lower := strings.ToLower(strings.TrimSpace(name))
	b, ok := Binomial(lower)
	if !ok {
		// A hybrid or a bare genus is a bird somebody saw and no species
		// BirdView draws, so it belongs in the region's denominator and in
		// nobody's numerator. Returned verbatim rather than dropped: that
		// keeps it counted, and puts it in birdview-regions' orphan report
		// under a name a reader can recognize as the deliberate loss it is.
		return lower
	}
	t := strings.ToUpper(b[:1]) + b[1:]
	if alt, ok := AviListRenames[t]; ok {
		return strings.ToLower(alt)
	}
	if alt, _, ok := SpeciesCountedAs(t); ok {
		return strings.ToLower(alt)
	}
	return b
}

// DomesticForms are species whose record counts are skewed because
// iNaturalist files the feral or domestic population under a trinomial and
// eBird does not (T-028).
//
// Not a naming disagreement to fix: both databases are right about their own
// records, and BirdView draws the wild species. Declared so the ratio check
// reads as a statement about taxonomy rather than a list of numbers to
// ignore.
var DomesticForms = map[string]string{
	"Columba livia":    "iNaturalist files feral Rock Pigeons under Columba livia domestica",
	"Cairina moschata": "iNaturalist files domestic Muscovy Ducks under Cairina moschata domestica",
}
