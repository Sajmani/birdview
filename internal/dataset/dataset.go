// Package dataset defines the compiled record BirdView serves, and the
// provenance vocabulary every presented value carries.
//
// Verifies: P-062.
package dataset

// Kind classifies how well grounded a value is. Every metric BirdView shows
// carries one, and the interface distinguishes them (P-061, P-063).
type Kind string

const (
	// Measured: a published measurement of this species, whatever name it was
	// filed under and whichever source sheet it sits in. A value read from
	// AVONET1_BirdLife, or found under a pre-revision name such as Charadrius
	// nivosus for Anarhynchus nivosus, is measured -- the specimens are the
	// right birds, and SourceTaxon records the name so the claim is auditable.
	Measured Kind = "measured"

	// NearMatch: a published measurement of a DIFFERENT taxon, used because
	// none exists for this species. Typically a pre-split parent, e.g. Larus
	// argentatus standing in for Larus vegae. Real data, wrong bird.
	NearMatch Kind = "near-match"

	// Modeled: computed rather than measured. Model names the computation.
	Modeled Kind = "modeled"
)

// Provenance records where one specific value came from.
//
// Verifies: P-062.
// Origin says who decided a categorical assignment. Deliberately a separate
// vocabulary from Kind, which grades a *number*: "measured" and "published"
// are not the same claim, and one word for both would blur the only
// distinction P-069 exists to draw.
type Origin string

const (
	// Published: a source says so, and Source names it.
	Published Origin = "published"
	// Derived: BirdView computed it from published data by a stated rule.
	Derived Origin = "derived"
	// BirdView: this project decided it and no source says it.
	BirdView Origin = "birdview"
)

// Assignment is one category a species is in, with its origin. The zero
// Origin is invalid: an unmarked category is an opinion wearing a finding's
// clothes, which is the failure P-069 exists to prevent.
type Assignment struct {
	Value  string `json:"value"`
	Origin Origin `json:"origin"`
	// Source names the publication, for Published.
	Source string `json:"source,omitempty"`
	// Why is one line saying how the value was arrived at, for Derived and
	// BirdView. Long explanations belong in a document, not here.
	Why string `json:"why,omitempty"`
	// Doc is a path BirdView itself serves, e.g. "data". There is no public
	// repository to link to, and a provenance link that needs the network is
	// one that breaks (P-069, T-002).
	Doc string `json:"doc,omitempty"`
}

type Provenance struct {
	Kind Kind `json:"kind"`

	// Source names the dataset and its pinned version, including the sheet
	// where that distinguishes rows. Empty only if Kind is Modeled and the
	// value derives from no source at all.
	Source string `json:"source,omitempty"`

	// SourceTaxon is the binomial the value was actually filed under. Required
	// for Measured and NearMatch. For NearMatch it necessarily differs from the
	// species' own scientific name -- that difference is the whole point.
	SourceTaxon string `json:"sourceTaxon,omitempty"`

	// Model names the computation that produced a Modeled value, specifically
	// enough that a reader can check it. For the neck this must name both the
	// per-family function and the ratio applied to this bird's family (P-045).
	Model string `json:"model,omitempty"`
}

// Metric is a value in millimeters (or grams, for mass) plus its provenance.
// Min and Max carry a reported range where the source supplies one; Fox et al.
// report wingspan as a range, and P-060 requires storing the triple rather
// than flattening it to a point.
type Metric struct {
	Value float64    `json:"value"`
	Min   float64    `json:"min,omitempty"`
	Max   float64    `json:"max,omitempty"`
	Prov  Provenance `json:"provenance"`
}

// Segments are the five vertical parts of the diagram, top to bottom. They abut
// exactly and sum to TotalLength (P-041), and each carries its own provenance
// because the detail view shows them individually (P-034, P-063).
type Segments struct {
	Bill Metric `json:"bill"`
	Head Metric `json:"head"`
	Neck Metric `json:"neck"`
	Body Metric `json:"body"`
	Tail Metric `json:"tail"`
}

// Locale selects which English name is served (P-049).
type Locale string

const (
	// LocaleNA is eBird/Clements usage -- Dovekie. The default, this being an
	// ABA-area dataset.
	LocaleNA Locale = "na"
	// LocaleWorld is AviList/IOC usage -- Little Auk.
	LocaleWorld Locale = "world"
)

// Species is one compiled record.
type Species struct {
	// ID is the eBird species code where AviList supplies one -- "amerob" --
	// and otherwise the scientific-name slug. It is the species URL and must
	// be stable across rebuilds (P-049, AC-084).
	ID string `json:"id"`

	// Names carries BOTH English names, keyed by Locale, because the locale is
	// a request-time choice and re-ingesting per locale would be absurd. The
	// world name is always present; the North American one is absent for the
	// five species with no Clements concept.
	Names map[Locale]string `json:"names"`

	// SpeciesCode is the eBird code, e.g. "amerob", shown beside the name. It
	// replaced the ABA four-letter alpha code, which was locale-DEPENDENT --
	// derived from the English name, so Dovekie's DOVE would have to become
	// LIAU under the world locale (CR-022).
	//
	// Not fixed-width despite the name: 4 to 8 characters in practice.
	SpeciesCode    string `json:"speciesCode,omitempty"`
	ScientificName string `json:"scientificName"`
	Family         string `json:"family"`
	FamilyCommon   string `json:"familyCommon"`
	FamilyColor    string `json:"familyColor"`
	FamilyText     string `json:"familyTextColor"`
	Order          string `json:"order"`

	// Regions are the places this species is reported from, as region ids
	// (P-011). Empty for the one species with too few records to place
	// anywhere, which the region filter must not silently swallow.
	Regions []string `json:"regions,omitempty"`

	// INatTaxonID links the detail view to iNaturalist (P-064). Zero means no
	// confident match, and then no link is shown at all -- not a search URL,
	// not a guess (CR-012).
	INatTaxonID int `json:"inatTaxonId,omitempty"`

	// The three main metrics shown on every tile (P-031).
	TotalLength Metric `json:"totalLength"`
	Wingspan    Metric `json:"wingspan"`
	Mass        Metric `json:"mass"`

	// WingChord is AVONET's Wing.Length: the ORNITHOLOGICAL chord, carpal joint
	// to the tip of the longest primary, measured along the wing. It is not the
	// aerodynamic chord (leading edge to trailing edge), which is a different
	// quantity on a different axis and which this diagram has no room for.
	//
	// Measured, and the input to the modeled wingspan (CR-008). Drawn as the
	// solid outer portion of the crossbar (P-065).
	WingChord Metric `json:"wingChord,omitempty"`

	// Habitats and Niches are the categorical assignments of P-015, each
	// carrying where it came from. Several are possible: AVONET records one
	// primary value, and BirdView may add its own, which is marked BirdView and
	// never silently mixed in with the published one (P-069).
	Habitats []Assignment `json:"habitats,omitempty"`
	Niches   []Assignment `json:"niches,omitempty"`

	// Shown in the detail view (P-034).
	Segments Segments `json:"segments"`
}

// Name returns the English name for a locale, falling back to the world name.
//
// The fallback is not decoration: five species have no Clements concept and so
// no North American name, and a reader who asked for NA names must still see a
// bird rather than a blank (P-049).
func (s *Species) Name(l Locale) string {
	if n := s.Names[l]; n != "" {
		return n
	}
	return s.Names[LocaleWorld]
}

// AllMetrics returns every metric the interface may present, so that checks
// over "every presented metric" (AC-040) cannot miss one by enumerating the
// struct by hand.
func (s *Species) AllMetrics() map[string]*Metric {
	return map[string]*Metric{
		"totalLength": &s.TotalLength,
		"wingspan":    &s.Wingspan,
		"mass":        &s.Mass,
		"bill":        &s.Segments.Bill,
		"head":        &s.Segments.Head,
		"neck":        &s.Segments.Neck,
		"body":        &s.Segments.Body,
		"tail":        &s.Segments.Tail,
	}
}

// Family is one taxonomic family, with the colors its badge uses.
//
// Families are listed in the ABA checklist's own taxonomic sequence, which is
// the order every field guide arranges birds in -- and which alphabetizing
// discards.
type Family struct {
	Name        string `json:"name"`        // scientific, e.g. Anatidae
	Common      string `json:"common"`      // e.g. Ducks, Geese, and Swans
	Order       string `json:"order"`       // e.g. Anseriformes
	OrderCommon string `json:"orderCommon"` // project-authored label, e.g. Waterfowl
	Color       string `json:"color"`
	TextColor   string `json:"textColor"`
	OrderColor  string `json:"orderColor"`
	OrderText   string `json:"orderTextColor"`
}

// Dataset is the whole compiled output.
type Dataset struct {
	// SourceVersions pins what this build was made from, so a served dataset
	// can be traced back to its inputs (G-004).
	SourceVersions map[string]string `json:"sourceVersions"`
	// Families in ABA taxonomic sequence, grouped implicitly by Order.
	Families []Family  `json:"families"`
	Species  []Species `json:"species"`
}
