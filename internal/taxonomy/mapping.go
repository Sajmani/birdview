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
