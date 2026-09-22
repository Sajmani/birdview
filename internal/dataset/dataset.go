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

// Species is one compiled record.
type Species struct {
	ID         string `json:"id"`
	CommonName string `json:"commonName"`
	// BandingCode is the ABA four-letter alpha code, e.g. AMRO. Not to be
	// confused with the ABA rarity code, which is a scope filter and is
	// deliberately discarded (P-001).
	BandingCode    string `json:"bandingCode,omitempty"`
	ScientificName string `json:"scientificName"`
	Family         string `json:"family"`
	FamilyCommon   string `json:"familyCommon"`
	FamilyColor    string `json:"familyColor"`
	FamilyText     string `json:"familyTextColor"`
	Order          string `json:"order"`

	// The three main metrics shown on every tile (P-031).
	TotalLength Metric `json:"totalLength"`
	Wingspan    Metric `json:"wingspan"`
	Mass        Metric `json:"mass"`

	// Tarsus is the measured lower leg; Leg is the modeled visible leg derived
	// from it. Both are shown, so a reader can see the measurement and what was
	// inferred from it. Like Wingspan neither is part of the vertical sum, so
	// they live outside Segments, which must add up (P-041).
	Tarsus Metric `json:"tarsus"`
	Leg    Metric `json:"leg"`

	// Shown in the detail view (P-034).
	Segments Segments `json:"segments"`
}

// AllMetrics returns every metric the interface may present, so that checks
// over "every presented metric" (AC-040) cannot miss one by enumerating the
// struct by hand.
func (s *Species) AllMetrics() map[string]*Metric {
	return map[string]*Metric{
		"totalLength": &s.TotalLength,
		"wingspan":    &s.Wingspan,
		"mass":        &s.Mass,
		"tarsus":      &s.Tarsus,
		"leg":         &s.Leg,
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
