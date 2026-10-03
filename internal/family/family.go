// Package family groups the bird families for the selector.
//
// Two groupings sit side by side (P-014). **Birder groups** is this project's
// own, hand-drawn, and offered first: Shorebirds, Birds of Prey, Owls &
// Nightbirds, the names a field guide uses. **Taxonomic order** is AviList's,
// and the server builds it from the dataset rather than from this package,
// because the dataset already carries each family's order.
//
// Birder groups exist because the taxonomic grouping cannot navigate. It is
// not wrong, it is lopsided: 43 of 90 families sit in Passeriformes under one
// label and 15 orders hold exactly one family, so the whole problem is one
// order and no relabelling fixes it.
//
// **The grouping claims nothing about the birds.** "Shorebirds" says
// Scolopacidae is a family birders call a shorebird; it does not say every
// sandpiper lives on a shore. That distinction is what makes a hand-drawn
// grouping defensible where an ecological one is not: grouping families by
// habitat was measured and rejected, because habitat is a property of a
// species and assigning each family its majority habitat misfiles 23-35% of
// birds (CR-025).
//
// Verifies: P-014.
package family

import "sort"

// ByBirders is the hand-drawn grouping, and the name the UI labels it with.
const ByBirders = "Birder groups"

// Note is the one line shown under the heading. A hand-drawn grouping must
// say that it is hand-drawn, for the same reason the Geographic region
// grouping does.
const Note = "The names a field guide uses. Drawn by hand, not measured — a group says how birders talk, not where a bird lives."

// Groups in the order the selector offers them: roughly waterbirds, then
// landbirds, then songbirds, which is the sequence a field guide runs in and
// therefore the one a reader scans.
//
// **Keyed on the scientific family name.** CR-022 renamed 55 of 90 English
// family names and only 2 scientific ones, so the English name is the fragile
// key and must not be used here.
var Groups = []struct {
	Name     string
	Families []string
}{
	{"Waterfowl, Loons & Grebes", []string{"Anatidae", "Gaviidae", "Podicipedidae"}},
	{"Game Birds", []string{"Phasianidae", "Odontophoridae", "Cracidae", "Pteroclidae"}},
	{"Seabirds", []string{
		"Procellariidae", "Hydrobatidae", "Oceanitidae", "Diomedeidae",
		"Sulidae", "Phalacrocoracidae", "Pelecanidae", "Fregatidae", "Anhingidae", "Phaethontidae"}},
	{"Herons, Ibises & Storks", []string{
		"Ardeidae", "Threskiornithidae", "Ciconiidae", "Phoenicopteridae"}},
	{"Birds of Prey", []string{"Accipitridae", "Pandionidae", "Falconidae", "Cathartidae"}},
	{"Rails & Cranes", []string{"Rallidae", "Gruidae", "Aramidae"}},
	{"Shorebirds", []string{"Scolopacidae", "Charadriidae", "Haematopodidae", "Recurvirostridae"}},
	{"Gulls, Terns & Auks", []string{"Laridae", "Stercorariidae", "Alcidae"}},
	{"Pigeons & Doves", []string{"Columbidae"}},
	{"Cuckoos", []string{"Cuculidae"}},
	{"Parrots", []string{"Psittacidae", "Psittaculidae"}},
	// Nightjars belong with the owls, not with the pigeons and cuckoos they
	// were first filed beside: "what you look for after dark" is a group a
	// birder has, and "mid-sized landbirds that fit nowhere else" is not.
	{"Owls & Nightbirds", []string{"Strigidae", "Tytonidae", "Caprimulgidae"}},
	// Swifts go with swallows, not with the hummingbirds they are actually
	// related to. "Swift or swallow?" is the question a bird overhead
	// provokes; nobody has ever had to tell a hummingbird from a swift. The
	// shared order is a taxonomic fact, and the taxonomic grouping is where
	// a reader should find taxonomic facts.
	{"Swifts & Swallows", []string{"Apodidae", "Hirundinidae"}},
	{"Hummingbirds", []string{"Trochilidae"}},
	{"Woodpeckers, Kingfishers & Trogons", []string{"Picidae", "Alcedinidae", "Trogonidae"}},
	{"Flycatchers, Vireos & Shrikes", []string{
		"Tyrannidae", "Vireonidae", "Laniidae", "Monarchidae", "Tityridae", "Ptiliogonatidae"}},
	{"Jays & Crows", []string{"Corvidae"}},
	// The Wrentit sits here, not with the Old World songbirds its family
	// belongs to. It is a resident of west-coast chaparral that a birder
	// meets alongside bushtits and wrens, and Paradoxornithidae is next to
	// Aegithalidae in checklist sequence anyway -- so this costs nothing
	// taxonomically and saves a familiar bird from a label nobody would
	// look for it under.
	{"Chickadees, Nuthatches & Wrens", []string{
		"Paridae", "Troglodytidae", "Sittidae", "Polioptilidae", "Regulidae",
		"Certhiidae", "Remizidae", "Aegithalidae", "Paradoxornithidae"}},
	{"Thrushes, Mimids & Starlings", []string{
		"Turdidae", "Mimidae", "Muscicapidae", "Cinclidae", "Bombycillidae", "Sturnidae"}},
	// Nine birds of Eurasian lineage, which in this scope means Hawaii's
	// introductions and a couple of Alaskan strays. They are their own group
	// rather than being folded in with the thrushes: "both are Old World
	// passerines" is not something a birder thinks, it is what is left over.
	{"Old World Songbirds", []string{
		"Pycnonotidae", "Cettiidae", "Phylloscopidae", "Acrocephalidae",
		"Leiothrichidae", "Zosteropidae"}},
	{"New World Warblers", []string{"Parulidae", "Peucedramidae"}},
	// Larks join the longspurs and Snow Bunting they flock with in winter
	// stubble, rather than the swallows they merely share open country with.
	{"Sparrows, Buntings & Larks", []string{
		"Passerellidae", "Calcariidae", "Emberizidae", "Passeridae", "Alaudidae"}},
	{"Cardinals & Tanagers", []string{"Cardinalidae", "Thraupidae", "Spindalidae"}},
	{"Blackbirds & Orioles", []string{"Icteridae"}},
	{"Finches, Waxbills & Pipits", []string{"Fringillidae", "Estrildidae", "Motacillidae"}},
}

// groupOf is the reverse index, built once. A family listed in two groups is
// a programming error that would make one of them unreachable, so it panics
// here rather than letting the selector quietly disagree with itself.
var groupOf = map[string]string{}

func init() {
	for _, g := range Groups {
		for _, f := range g.Families {
			if other, dup := groupOf[f]; dup {
				panic("family: " + f + " is in both " + other + " and " + g.Name)
			}
			groupOf[f] = g.Name
		}
	}
}

// GroupOf returns the birder group a family belongs to, and whether the table
// knows the family at all. An unknown family is reported rather than filed
// under a default: AviList can add one, and a bird that silently disappears
// from the selector is worse than a failing test (AC-089).
func GroupOf(scientificName string) (string, bool) {
	g, ok := groupOf[scientificName]
	return g, ok
}

// Names lists the groups in offer order.
func Names() []string {
	out := make([]string, 0, len(Groups))
	for _, g := range Groups {
		out = append(out, g.Name)
	}
	return out
}

// InGroup lists the scientific family names in one group, sorted.
func InGroup(group string) []string {
	for _, g := range Groups {
		if g.Name == group {
			out := append([]string(nil), g.Families...)
			sort.Strings(out)
			return out
		}
	}
	return nil
}

// Len is how many families the table files.
func Len() int { return len(groupOf) }
