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

// Package gbif bridges the scientific names in GBIF's backbone taxonomy onto
// the names AviList uses, so a species list derived from GBIF occurrence
// records can be joined to BirdView's.
//
// The join does not work on names alone, because GBIF's backbone lags current
// taxonomy by several revisions. GBIF files the Double-crested Cormorant as
// Phalacrocorax auritus where AviList has Nannopterum auritum, the
// Ruby-crowned Kinglet as Regulus calendula where AviList has Corthylio
// calendula, and so on for 71 of the 1,378 species the eBird Observation
// Dataset reports from the United States and Canada.
//
// GBIF's /species/match endpoint does NOT solve this, and the temptation to
// reach for it should be resisted. It maps an old name onto the current
// backbone -- old to new -- where this bridge needs new to old. Asked for
// AviList's Astur cooperii on 2026-10-02 it answered matchType NONE under
// strict=true, and without it offered Aster cooperi, a shrub in the
// Asteraceae, at confidence 84.
//
// What does work is the backbone's own synonymy: GET /species/{key} gives the
// accepted name and GET /species/{key}/synonyms lists the names it absorbed,
// and Nannopterum auritum is in that list. Resolve applies that in three
// clearly separated tiers and refuses anything ambiguous, because a plausible
// wrong bird is worse than an honest gap.
//
// The mapping is a committed table rather than a build-time fetch, for the
// reason T-022 gives for the iNaturalist one: tests make no network request
// (tech.md bindings), and a build that needs the network is a build that fails
// on a plane. See cmd/birdview-gbifmap.
//
// Verifies: T-006, T-021.
package gbif

import (
	"bufio"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"
)

// Source names this input in provenance records.
const Source = "GBIF Backbone Taxonomy, doi:10.15468/39omei (CC BY 4.0)"

// Tier records how a name was resolved, in descending order of confidence.
// It is written into the committed table so the contribution of each strategy
// stays visible and a fuzzy match can never be mistaken for an exact one.
type Tier string

const (
	// TierExact: GBIF's accepted name is an AviList species name, verbatim.
	TierExact Tier = "exact"
	// TierSynonym: exactly one name GBIF files as a synonym of this taxon is
	// an AviList species name. This is the tier that does the real work.
	TierSynonym Tier = "synonym"
	// TierEpithet: no name matched, but within GBIF's family exactly one
	// AviList species shares the epithet once gender agreement and Latin
	// genitive endings are folded away.
	TierEpithet Tier = "epithet"
	// TierAmbiguous: more than one AviList species corresponded, so the match
	// was refused. Recorded rather than silently dropped -- it is a finding.
	TierAmbiguous Tier = "ambiguous"
	// TierNone: nothing in AviList corresponded.
	TierNone Tier = "none"
)

// Resolved reports whether a tier names an AviList species.
func (t Tier) Resolved() bool { return t == TierExact || t == TierSynonym || t == TierEpithet }

var tiers = []Tier{TierExact, TierSynonym, TierEpithet, TierAmbiguous, TierNone}

func validTier(s string) bool {
	for _, t := range tiers {
		if Tier(s) == t {
			return true
		}
	}
	return false
}

// Taxon is one GBIF backbone species as the mapping tool retrieves it.
type Taxon struct {
	Key      int      // GBIF species key
	Name     string   // accepted canonical name, e.g. "Phalacrocorax auritus"
	Family   string   // GBIF's family placement, which scopes the epithet tier
	Synonyms []string // species-rank synonyms GBIF files under this taxon
	Records  int64    // occurrences behind it; used only in the tool's report
}

// Entry is one row of the bridge.
type Entry struct {
	GBIFName string
	GBIFKey  int
	AviList  string // the AviList species name; empty unless Tier.Resolved()
	Tier     Tier
}

// Table maps a GBIF backbone scientific name onto an AviList one.
//
// Unresolved taxa are present, with an empty AviList name and a tier saying
// why. Absence would be indistinguishable from a taxon nobody looked at, and
// the residue is the part of this table worth reading.
type Table map[string]Entry

// Index is AviList arranged for the three tiers: a set of names for the first
// two, and a family-scoped epithet index for the third.
type Index struct {
	names    map[string]bool
	byFamily map[string]map[string][]string
}

// NewIndex builds the index from a map of AviList scientific name to family.
//
// It takes a plain map rather than an avilist.Table so that this package does
// not depend on a 9 MB spreadsheet reader, and so a test can state the three
// species a case needs instead of loading eleven thousand.
func NewIndex(familyOf map[string]string) *Index {
	idx := &Index{
		names:    make(map[string]bool, len(familyOf)),
		byFamily: map[string]map[string][]string{},
	}
	for name, family := range familyOf {
		idx.names[name] = true
		e := epithet(name)
		if e == "" || family == "" {
			continue
		}
		if idx.byFamily[family] == nil {
			idx.byFamily[family] = map[string][]string{}
		}
		k := fold(e)
		idx.byFamily[family][k] = append(idx.byFamily[family][k], name)
	}
	for _, m := range idx.byFamily {
		for k := range m {
			sort.Strings(m[k])
		}
	}
	return idx
}

// Len reports how many AviList species the index holds.
func (idx *Index) Len() int { return len(idx.names) }

// Resolve maps one GBIF taxon onto an AviList species.
//
// The tiers are tried in order and the first that produces exactly one AviList
// species wins. Every tier refuses ambiguity: two candidates mean the
// correspondence is unknown, not that either will do.
func Resolve(t Taxon, idx *Index) Entry {
	e := Entry{GBIFName: t.Name, GBIFKey: t.Key}

	// 1. The accepted name itself. Strongest evidence there is, and it covers
	// the large majority -- GBIF and AviList agree far more often than not.
	if idx.names[t.Name] {
		e.AviList, e.Tier = t.Name, TierExact
		return e
	}

	// 2. A synonym. GBIF keeps the pre-revision names under the accepted
	// taxon, and a name AviList has promoted is usually one of them:
	// Nannopterum auritum is a synonym of Phalacrocorax auritus.
	var hits []string
	for _, s := range t.Synonyms {
		if idx.names[s] {
			hits = appendUnique(hits, s)
		}
	}
	switch len(hits) {
	case 1:
		e.AviList, e.Tier = hits[0], TierSynonym
		return e
	case 0:
		// fall through to the epithet tier
	default:
		e.Tier = TierAmbiguous
		return e
	}

	// 3. The epithet, scoped to GBIF's family. This is the fallback, and it is
	// deliberately the narrowest thing that still helps: it recovers a change
	// of genus and a change of Latin case ending, and nothing else.
	//
	// Family scoping and the uniqueness rule are both load-bearing. Across
	// AviList there are 369 (family, folded epithet) buckets holding more than
	// one species -- Accipitridae alone has Astur gundlachi beside Buteogallus
	// gundlachii -- so an epithet match without a uniqueness check would be a
	// coin toss in those buckets. Measured over AviList v2025b, 2026-10-02.
	//
	// Candidate epithets come from the synonyms as well as the accepted name,
	// which is what resolves GBIF's misspelled Phylloscopus sibillatrix: the
	// accepted name carries a doubled L that AviList does not, but the synonym
	// Rhadina sibilatrix spells it the way AviList does.
	byEpithet := idx.byFamily[t.Family]
	if byEpithet == nil {
		e.Tier = TierNone
		return e
	}
	var cands []string
	for _, n := range append([]string{t.Name}, t.Synonyms...) {
		if ep := epithet(n); ep != "" {
			for _, c := range byEpithet[fold(ep)] {
				cands = appendUnique(cands, c)
			}
		}
	}
	switch len(cands) {
	case 1:
		e.AviList, e.Tier = cands[0], TierEpithet
	case 0:
		e.Tier = TierNone
	default:
		e.Tier = TierAmbiguous
	}
	return e
}

func appendUnique(xs []string, s string) []string {
	for _, x := range xs {
		if x == s {
			return xs
		}
	}
	return append(xs, s)
}

// epithet returns the specific epithet of a binomial, lowercased, or "" for
// anything that is not two words. GBIF's synonym lists contain the occasional
// piece of debris -- "Motacilla spec" is filed under the Wood Warbler -- and
// returning it is harmless, because it has to match an AviList species to
// matter.
func epithet(name string) string {
	f := strings.Fields(name)
	if len(f) != 2 {
		return ""
	}
	return strings.ToLower(f[1])
}

// genitive are the Latin genitive endings a patronymic epithet takes, longest
// first. Icterus bullockii and Icterus bullockiorum are the same oriole named
// for the same Bullocks, differing only in whether the name honors one of them
// or both; Pica nuttalli and Pica nuttallii differ only in how many i's the
// genitive of Nuttall is held to need.
var genitive = []string{"iorum", "orum", "arum", "iae", "ae", "ii", "i"}

// gendered are the adjectival endings that change when a species moves genus,
// because an epithet agrees in grammatical gender with its genus: Oressochen
// jubatus and Neochen jubata are one goose.
//
// The same list, for the same reason, is in internal/inat. They are kept apart
// because the rules around them differ -- that one has no genitive step, and
// pairs every match with an agreeing common name instead.
var gendered = []string{"rum", "um", "us", "is", "ra", "er", "a", "e"}

// fold reduces an epithet to a form comparable across a change of genus or of
// Latin case. A genitive is tried first, since "ii" is a genitive ending and
// "is" is a gendered one and they would otherwise fight over "bullockii".
//
// Short results are refused throughout: trimming four letters off a six-letter
// epithet leaves a stem that would match almost anything.
func fold(e string) string {
	for _, suf := range genitive {
		if len(e) > len(suf)+3 && strings.HasSuffix(e, suf) {
			return e[:len(e)-len(suf)]
		}
	}
	for _, suf := range gendered {
		if len(e) > len(suf)+3 && strings.HasSuffix(e, suf) {
			return e[:len(e)-len(suf)]
		}
	}
	return e
}

// noName is what the AviList column holds on an unresolved row. An explicit
// marker rather than an empty field, so a truncated line cannot read as a
// deliberate non-match.
const noName = "-"

// Load reads a mapping table. One tab-separated record per line,
// `gbif name<TAB>gbif species key<TAB>avilist name<TAB>tier`, with `#`
// comments and blank lines ignored.
func Load(r io.Reader) (Table, error) {
	t := Table{}
	claimed := map[string]string{} // avilist name -> the gbif name that claimed it
	sc := bufio.NewScanner(r)
	for line := 1; sc.Scan(); line++ {
		s := strings.TrimSpace(sc.Text())
		if s == "" || strings.HasPrefix(s, "#") {
			continue
		}
		f := strings.Split(s, "\t")
		if len(f) != 4 {
			return nil, fmt.Errorf("gbif table line %d: got %d fields, want 4", line, len(f))
		}
		key, err := strconv.Atoi(f[1])
		if err != nil {
			return nil, fmt.Errorf("gbif table line %d: species key %q: %w", line, f[1], err)
		}
		if key <= 0 {
			return nil, fmt.Errorf("gbif table line %d: species key %d is not positive", line, key)
		}
		if !validTier(f[3]) {
			return nil, fmt.Errorf("gbif table line %d: unknown tier %q", line, f[3])
		}
		tier := Tier(f[3])
		if tier.Resolved() == (f[2] == noName) {
			return nil, fmt.Errorf("gbif table line %d: tier %q with AviList name %q", line, tier, f[2])
		}
		if _, dup := t[f[0]]; dup {
			return nil, fmt.Errorf("gbif table line %d: %q appears twice", line, f[0])
		}
		e := Entry{GBIFName: f[0], GBIFKey: key, Tier: tier}
		if tier.Resolved() {
			e.AviList = f[2]
			// The bridge must be injective. Two GBIF taxa pointing at one
			// AviList species would merge two entries of a derived species
			// list into one, and the count that disappeared would belong to
			// whichever bird GBIF happened to list second.
			if prev, dup := claimed[e.AviList]; dup {
				return nil, fmt.Errorf("gbif table line %d: %q and %q both claim AviList %q",
					line, prev, f[0], e.AviList)
			}
			claimed[e.AviList] = f[0]
		}
		t[f[0]] = e
	}
	return t, sc.Err()
}

// Write serializes a table in Load's format, sorted by GBIF name, so that
// regenerating it produces a reviewable diff rather than a reshuffle.
func Write(w io.Writer, t Table) error {
	names := make([]string, 0, len(t))
	for k := range t {
		names = append(names, k)
	}
	sort.Strings(names)

	var resolved int
	for _, e := range t {
		if e.Tier.Resolved() {
			resolved++
		}
	}

	if _, err := fmt.Fprintf(w,
		"# GBIF backbone -> AviList name bridge. Generated by cmd/birdview-gbifmap.\n"+
			"# gbif scientific name<TAB>gbif species key<TAB>avilist scientific name<TAB>tier\n"+
			"#\n"+
			"# Scope: every species the eBird Observation Dataset (doi:10.15468/aomfnb)\n"+
			"# reports from the US and Canada. Names and synonymy from the GBIF Backbone\n"+
			"# Taxonomy (doi:10.15468/39omei). Both CC BY 4.0.\n"+
			"#\n"+
			"# tier=exact     GBIF's accepted name is an AviList species name\n"+
			"# tier=synonym   one GBIF synonym of it is\n"+
			"# tier=epithet   one AviList species in the same family shares the epithet,\n"+
			"#                after folding gender agreement and Latin genitive endings\n"+
			"# tier=ambiguous more than one corresponded, so none was chosen\n"+
			"# tier=none      nothing corresponded\n"+
			"#\n"+
			"# A resolved row says the two authorities use different names for one taxon.\n"+
			"# It does NOT say their species CONCEPTS agree. Where GBIF has not followed a\n"+
			"# split, its name maps to whichever half kept the name, and the occurrences\n"+
			"# behind it are mostly the other half: GBIF's Circus cyaneus maps to AviList's\n"+
			"# Circus cyaneus, the Eurasian Hen Harrier, and 3.8 million of those records\n"+
			"# are Northern Harriers. Nine such lumps touch BirdView's own species; they are\n"+
			"# named in river/sources.md under `gbif`.\n"+
			"#\n"+
			"# %d of %d resolved.\n", resolved, len(t)); err != nil {
		return err
	}
	for _, n := range names {
		e := t[n]
		avi := e.AviList
		if !e.Tier.Resolved() {
			avi = noName
		}
		if _, err := fmt.Fprintf(w, "%s\t%d\t%s\t%s\n", n, e.GBIFKey, avi, e.Tier); err != nil {
			return err
		}
	}
	return nil
}
