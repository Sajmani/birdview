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
	// The owls appear here **and** under Owls & Nightbirds. They are birds of
	// prey by any definition, and a reader who wants to compare a Great
	// Horned Owl against a Red-tailed Hawk should not have to know that this
	// project filed them by hour rather than by trade. See the note on Owls
	// & Nightbirds for the other half of the argument; both are true, so the
	// grouping stops insisting on exactly one (CR-026).
	{"Birds of Prey", []string{
		"Accipitridae", "Pandionidae", "Falconidae", "Cathartidae", "Strigidae", "Tytonidae"}},
	{"Rails & Cranes", []string{"Rallidae", "Gruidae", "Aramidae"}},
	{"Shorebirds", []string{"Scolopacidae", "Charadriidae", "Haematopodidae", "Recurvirostridae"}},
	{"Gulls, Terns & Auks", []string{"Laridae", "Stercorariidae", "Alcidae"}},
	{"Pigeons & Doves", []string{"Columbidae"}},
	{"Cuckoos", []string{"Cuculidae"}},
	{"Parrots", []string{"Psittacidae", "Psittaculidae"}},
	// Nightjars belong with the owls, not with the pigeons and cuckoos they
	// were first filed beside: "what you look for after dark" is a group a
	// birder has, and "mid-sized landbirds that fit nowhere else" is not.
	//
	// **The owls are also under Birds of Prey**, and that is the point. In
	// practice "raptor" means the diurnal ones -- a hawk watch counts hawks,
	// eagles, falcons and vultures, and nobody reports six raptors meaning
	// owls -- while owling is its own activity at its own hour, and
	// nightjars are what else it turns up. Both are true.
	//
	// Owls sit on the seam in the whole scheme: some names say what a bird
	// is (Birds of Prey, Waterfowl) and some say when or where you meet it
	// (Shorebirds, Seabirds). Rather than pick a side and be half wrong,
	// they are in both (CR-026).
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
	// The small brown birds of open ground, which is a confusion set rather
	// than a lineage. Larks join the longspurs and Snow Bunting they flock
	// with in winter stubble rather than the swallows they merely share open
	// country with, and the pipits join them for the same reason -- they had
	// been filed with the finches, which is wrong twice over: a pipit is an
	// insectivore that walks, and a finch is a seedeater in a tree.
	{"Sparrows, Larks & Pipits", []string{
		"Passerellidae", "Calcariidae", "Emberizidae", "Passeridae", "Alaudidae", "Motacillidae"}},
	{"Cardinals & Tanagers", []string{"Cardinalidae", "Thraupidae", "Spindalidae"}},
	{"Blackbirds & Orioles", []string{"Icteridae"}},
	// Not merged with the sparrows, though eight families would be no larger
	// than Seabirds' ten. "Finches" and "Sparrows" are two of the strongest
	// names a birder has, and lumping them trades two words people say for
	// one they do not.
	{"Finches & Waxbills", []string{"Fringillidae", "Estrildidae"}},
}

// groupsOf is the reverse index, built once.
//
// **A family may be in more than one group.** The groups cover every family
// but do not partition them: an owl is a bird of prey and a bird of the
// night, and a reader looking under either should find it (CR-026). The
// duplication is deliberate and rare, so the index is a slice rather than a
// single value, and a family listed twice in the *same* group is still a
// mistake and still panics.
var groupsOf = map[string][]string{}

func init() {
	for _, g := range Groups {
		for _, f := range g.Families {
			for _, already := range groupsOf[f] {
				if already == g.Name {
					panic("family: " + f + " is listed twice in " + g.Name)
				}
			}
			groupsOf[f] = append(groupsOf[f], g.Name)
		}
	}
}

// GroupsOf returns every birder group a family belongs to, in offer order.
// An unknown family returns nothing rather than being filed under a default:
// AviList can add one, and a bird that silently disappears from the selector
// is worse than a failing test (AC-089).
func GroupsOf(scientificName string) []string { return groupsOf[scientificName] }

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

// Len is how many distinct families the table files -- not how many
// placements it makes, which is larger wherever a family is in two groups.
func Len() int { return len(groupsOf) }
