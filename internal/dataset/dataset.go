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

// Package dataset defines the compiled record BirdView serves, and the
// provenance vocabulary every presented value carries.
//
// Verifies: product.provenancePerMetric.
package dataset

// Kind classifies how well grounded a value is. Every metric BirdView shows
// carries one, and the interface distinguishes them (product.tileMarksModeled, product.metricProvenanceExplained).
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
// Verifies: product.provenancePerMetric.
// Origin says who decided a categorical assignment. Deliberately a separate
// vocabulary from Kind, which grades a *number*: "measured" and "published"
// are not the same claim, and one word for both would blur the only
// distinction product.categoryProvenance exists to draw.
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
// clothes, which is the failure product.categoryProvenance exists to prevent.
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
	// one that breaks (product.categoryProvenance, tech.selfContainedServer).
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
	// per-family function and the ratio applied to this bird's family (product.neckModelPerFamily).
	Model string `json:"model,omitempty"`
}

// Metric is a value in millimeters (or grams, for mass) plus its provenance.
// Min and Max carry a reported range where the source supplies one; Fox et al.
// report wingspan as a range, and product.lengthAndWingspanSources requires storing the triple rather
// than flattening it to a point.
type Metric struct {
	Value float64    `json:"value"`
	Min   float64    `json:"min,omitempty"`
	Max   float64    `json:"max,omitempty"`
	Prov  Provenance `json:"provenance"`
}

// Segments are the five vertical parts of the diagram, top to bottom. They abut
// exactly and sum to TotalLength (product.segmentsSumToLength), and each carries its own provenance
// because the detail view shows them individually (product.tileOpensDetail, product.metricProvenanceExplained).
type Segments struct {
	Bill Metric `json:"bill"`
	Head Metric `json:"head"`
	Neck Metric `json:"neck"`
	Body Metric `json:"body"`
	Tail Metric `json:"tail"`
}

// Locale selects which English name is served (product.englishNameChoice).
type Locale string

const (
	// LocaleNA is eBird/Clements usage -- Dovekie. The default, this being an
	// ABA-area dataset.
	LocaleNA Locale = "na"
	// LocaleWorld is AviList/IOC usage -- Little Auk.
	LocaleWorld Locale = "world"
)

// Commonest records where a species is the largest share of what a region
// reports, and the evidence that share rests on (product.shareOfRecords).
type Commonest struct {
	// Region is a region id, as Species.Regions uses.
	Region string `json:"region"`
	// Records is this species' records there; RegionRecords is every bird
	// record the region holds. The ratio is the share.
	Records       int64 `json:"records"`
	RegionRecords int64 `json:"regionRecords"`
}

// OneIn is the share expressed the way it is shown: one record in N. Zero
// when there is nothing to divide.
// WithheldIn returns the months eBird withholds the species in a region,
// January first (decisions.ebirdSilenceNotAbsence).
func (s *Species) WithheldIn(id string) [12]bool {
	var out [12]bool
	bits := s.RegionWithheld[id]
	for m := range out {
		out[m] = bits&(1<<m) != 0
	}
	return out
}

func (c *Commonest) OneIn() int {
	if c == nil || c.Records == 0 {
		return 0
	}
	return int(c.RegionRecords / c.Records)
}

// Species is one compiled record.
type Species struct {
	// ID is the eBird species code where AviList supplies one -- "amerob" --
	// and otherwise the scientific-name slug. It is the species URL and must
	// be stable across rebuilds (product.englishNameChoice, acceptance.stableSpeciesId).
	ID string `json:"id"`

	// Names carries BOTH English names, keyed by Locale, because the locale is
	// a request-time choice and re-ingesting per locale would be absurd. The
	// world name is always present; the North American one is absent for the
	// five species with no Clements concept.
	Names map[Locale]string `json:"names"`

	// SpeciesCode is the eBird code, e.g. "amerob", shown beside the name. It
	// replaced the ABA four-letter alpha code, which was locale-DEPENDENT --
	// derived from the English name, so Dovekie's DOVE would have to become
	// LIAU under the world locale (decisions.avilistReplacesAba).
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
	// (product.regionAndMonthFilter). Empty for the one species with too few records to place
	// anywhere, which the region filter must not silently swallow.
	Regions []string `json:"regions,omitempty"`

	// RegionCounts is how many records put the species in each of Regions,
	// in the same order and of the same length.
	//
	// Parallel to Regions rather than a map: the ids were already a list,
	// every consumer walks them in order, and a map would repeat each key as
	// a JSON string across 27,646 entries. The invariant that the two line
	// up is asserted, not assumed.
	RegionCounts []int64 `json:"regionCounts,omitempty"`

	// RegionMonths is RegionCounts split twelve ways, January first, and
	// parallel to Regions in the same way (product.regionAndMonthFilter, decisions.monthPresence).
	RegionMonths [][12]int64 `json:"regionMonths,omitempty"`

	// BelowRegions are the regions where the species has records but stays
	// under the presence bar, with BelowCounts and BelowMonths parallel as
	// above. Their records count in the species' figure; the species is not
	// listed under them, so nothing that filters by region reads these
	// (product.regionAndMonthFilter, decisions.belowBarCounted).
	BelowRegions []string    `json:"belowRegions,omitempty"`
	BelowCounts  []int64     `json:"belowCounts,omitempty"`
	BelowMonths  [][12]int64 `json:"belowMonths,omitempty"`

	// RegionWithheld maps a region id to the months eBird withholds this
	// species there under its Sensitive Species list, as a bitmask with
	// January in bit 0 (decisions.ebirdSilenceNotAbsence). A map rather than a slice parallel to
	// Regions because almost every cell has none: a few hundred, against
	// some 28,000.
	RegionWithheld map[string]uint16 `json:"regionWithheld,omitempty"`

	// Exotic is the declared reason this species is left off the index: its
	// records here are escapes or birds kept by people (product.statusIsPublished, product.escapedBirdsListed). Empty
	// for every other species. The detail page is still served.
	Exotic string `json:"exotic,omitempty"`

	// Commonest is where this species makes up the biggest SHARE of the bird
	// records a region holds, with both halves of the ratio kept (product.shareOfRecords).
	//
	// Share rather than record count, because a count says where the
	// birdwatchers are: California holds 35x the Sharp-shinned Hawk records
	// of Yukon and 118x the birdwatching, so the hawk is one record in 250
	// in Yukon and one in 835 in California. Both numbers are carried, not
	// the ratio, so a share resting on nine records can say so.
	Commonest *Commonest `json:"commonest,omitempty"`

	// INatTaxonID links the detail view to iNaturalist (product.inaturalistLink). Zero means no
	// confident match, and then no link is shown at all -- not a search URL,
	// not a guess (decisions.inaturalistLinkAdopted).
	INatTaxonID int `json:"inatTaxonId,omitempty"`

	// The three main metrics shown on every tile (product.tileContents).
	TotalLength Metric `json:"totalLength"`
	Wingspan    Metric `json:"wingspan"`
	Mass        Metric `json:"mass"`

	// WingChord is AVONET's Wing.Length: the ORNITHOLOGICAL chord, carpal joint
	// to the tip of the longest primary, measured along the wing. It is not the
	// aerodynamic chord (leading edge to trailing edge), which is a different
	// quantity on a different axis and which this diagram has no room for.
	//
	// Measured, and the input to the modeled wingspan (decisions.fittedWingspanModel). Drawn as the
	// solid outer portion of the crossbar (product.measuredVersusModeled).
	WingChord Metric `json:"wingChord,omitempty"`

	// Habitats and Niches are the categorical assignments of product.habitatAndDietFilters, each
	// carrying where it came from. Several are possible: AVONET records one
	// primary value, and BirdView may add its own, which is marked BirdView and
	// never silently mixed in with the published one (product.categoryProvenance).
	Habitats []Assignment `json:"habitats,omitempty"`
	Niches   []Assignment `json:"niches,omitempty"`

	// Shown in the detail view (product.tileOpensDetail).
	Segments Segments `json:"segments"`
}

// Name returns the English name for a locale, falling back to the world name.
//
// The fallback is not decoration: five species have no Clements concept and so
// no North American name, and a reader who asked for NA names must still see a
// bird rather than a blank (product.englishNameChoice).
func (s *Species) Name(l Locale) string {
	if n := s.Names[l]; n != "" {
		return n
	}
	return s.Names[LocaleWorld]
}

// AllMetrics returns every metric the interface may present, so that checks
// over "every presented metric" (acceptance.provenanceWellFormed) cannot miss one by enumerating the
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

// Undrawn is a species the status derivation classified and left out of
// scope: it has a URL and no drawing (product.undrawnSpeciesAnswer, decisions.speciesUrlOutlivesSpecies).
//
// Deliberately NOT a Species with empty metrics. A Species is the input to
// every drawing function in the program, and 334 of them carrying zero mass
// and no family color would be 334 chances for one of those functions to
// divide by nothing. The two types answer different questions -- what to draw,
// and what to say instead -- and the compiler should know which it has.
type Undrawn struct {
	// ID and Names are resolved exactly as a Species' are, so a URL that
	// worked before the species left scope still reaches this record.
	ID             string            `json:"id"`
	Names          map[Locale]string `json:"names"`
	ScientificName string            `json:"scientificName"`

	// Status is the verdict: vagrant, historical or absent. A string rather
	// than the status package's own type because that package imports this
	// one, and the dependency may not run both ways.
	Status string `json:"status"`

	// The evidence behind the verdict, from the region that decided it.
	//
	// All four are omitted together for `absent`, where the derivation never
	// measured any dates: writing the zeros would present "not measured" as
	// "measured zero", which is the confusion Provenance exists to prevent.
	Region   string `json:"region,omitempty"`
	Episodes int    `json:"episodes,omitempty"`
	Years    int    `json:"years,omitempty"`
	Records  int64  `json:"records,omitempty"`
	// LastSeen is an ISO date, and the whole content of a `historical`
	// verdict: Passenger Pigeon recurs across 89 episodes in 52 years, all of
	// which end on 1902-05-18.
	LastSeen string `json:"lastSeen,omitempty"`
}

// Measured reports whether the derivation has evidence to show for this
// verdict. False for `absent`, whose counts are absences of measurement
// rather than measurements of absence.
func (u *Undrawn) Measured() bool { return u.Region != "" }

// Name returns the English name for a locale, falling back to the world name,
// on the same rule as Species.Name.
func (u *Undrawn) Name(l Locale) string {
	if n := u.Names[l]; n != "" {
		return n
	}
	return u.Names[LocaleWorld]
}

// Family is one taxonomic family, with the colors its badge uses.
//
// Families are listed in AviList's linear sequence, which is phylogenetic --
// the order every field guide arranges birds in, and which alphabetizing
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
	// RegionTotals is every bird record a region holds, over every species
	// including those BirdView does not draw. It is the denominator when a
	// page says what share of the birds recorded in the regions on screen
	// are one species (product.shareOfRecords).
	RegionTotals map[string]int64 `json:"regionTotals,omitempty"`

	// RegionMonthlyTotals is the same denominator split twelve ways,
	// January first, for a figure scoped to months (product.regionAndMonthFilter, decisions.monthPresence). The
	// twelve sum to RegionTotals exactly -- every record carries a month,
	// and birdview-regions refuses to write a table where they do not.
	RegionMonthlyTotals map[string][12]int64 `json:"regionMonthlyTotals,omitempty"`

	// SourceVersions pins what this build was made from, so a served dataset
	// can be traced back to its inputs (goals.goodDataCitizen).
	SourceVersions map[string]string `json:"sourceVersions"`
	// Families in AviList's taxonomic sequence, grouped implicitly by Order.
	Families []Family  `json:"families"`
	Species  []Species `json:"species"`

	// Undrawn are the species the derivation classified and rejected, so a
	// URL that stops working can say why rather than 404 into silence
	// (product.undrawnSpeciesAnswer). Not species: nothing here is drawn, filtered, counted or
	// compared.
	Undrawn []Undrawn `json:"undrawn,omitempty"`
}
