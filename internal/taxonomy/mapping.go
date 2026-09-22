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
// Note that Larus smithsonianus and Larus vegae would both resolve to Larus
// argentatus; smithsonianus is absent here because AVONET1_BirdLife carries it
// as its own taxon, so it resolves as measured instead.
var NearMatches = map[string]string{
	"Larus vegae":         "Larus argentatus",
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

// AwaitingTaxonomicCheck lists the entries whose rename-versus-split status is
// still open, so the ingest can report them rather than let the question fade.
//
// Verifies: P-007.
var AwaitingTaxonomicCheck = []string{
	"Astur atricapillus", "Ardea ibis", "Gygis candida", "Tyto furcata",
}
