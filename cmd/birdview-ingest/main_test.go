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

package main

import (
	"encoding/json"
	"math"
	"net/url"
	"os"
	"sort"
	"strings"
	"testing"

	"birdview/internal/avonet"
	"birdview/internal/dataset"
	"birdview/internal/inat"
	"birdview/internal/region"
	"birdview/internal/status"
	"birdview/internal/taxonomy"
)

const (
	statusFile = "../../data/status.tsv"
	avoFile    = "../../third_party/avonet/AVONET Supplementary dataset 1.xlsx"
	foxFile    = "../../third_party/fox-wingspan/wingspan_fox2026_AppendixS1.xlsx"
	avlFile    = "../../third_party/avilist/AviList-v2025b-10Jun2026-extended.xlsx"
	inatFile   = "../../data/inat-taxa.tsv"
	regFile    = "../../data/region-presence.tsv"
)

// built caches the compiled dataset across tests: the AVONET workbook is 21 MB
// and parsing it three times per test would dominate the suite.
var built *dataset.Dataset

func load(t *testing.T) *dataset.Dataset {
	t.Helper()
	if built != nil {
		return built
	}
	// The status table IS committed, unlike the ABA checklist it replaced
	// (CR-028), so this no longer skips on a clean checkout. That is the
	// point of the change: the build depends on nothing it may not
	// redistribute.
	ds, _, err := build(statusFile, avoFile, foxFile, avlFile)
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	// main() applies these after build() -- both are enhancements it warns
	// about rather than dies on -- so a test loading only build() sees a
	// dataset no reader ever gets. Composed here the same way, and fatal
	// here because a test silently checking an unregioned dataset proves
	// nothing about the one that ships.
	if _, err := applyRegions(ds, regFile); err != nil {
		t.Fatalf("applyRegions: %v", err)
	}
	built = ds
	return ds
}

// Scope is now every species the status table says recurs, plus the declared
// exceptions, less what cannot be named or measured (P-001, CR-028).
//
// The expected count is derived from those inputs rather than from a run:
// count the in-scope rows, add the declared, subtract the names AviList has
// no concept for. Writing the total as a literal would make this test agree
// with whatever the code last produced.
//
// Verifies: P-001. Criterion: AC-010.
func TestScopeIsWhatTheStatusTableSaysRecurs(t *testing.T) {
	ds := load(t)
	f, err := os.Open(statusFile)
	if err != nil {
		t.Fatalf("opening the status table: %v", err)
	}
	defer f.Close()
	rows, err := status.Read(f)
	if err != nil {
		t.Fatalf("reading the status table: %v", err)
	}
	// Only the species AviList lacks that are actually in scope are
	// subtracted: Vireo swainsoni left scope under CR-038, and subtracting
	// it anyway would make this test demand one species fewer than exists.
	scope := status.Scope(rows)
	noConcept := 0
	for _, sci := range scope {
		if _, missing := taxonomy.NotInAviList[sci]; missing {
			noConcept++
		}
	}
	want := len(scope) - noConcept - unmeasurable
	if got := len(ds.Species); got != want {
		t.Errorf("got %d species, want %d (in scope, less %d with no AviList concept "+
			"and %d with no AVONET measurement)",
			got, want, noConcept, unmeasurable)
	}
	// The scope must actually be the derived one, not a leftover list.
	if len(ds.Species) < 1000 {
		t.Errorf("only %d species: the ABA checklist carried 839 and the derived "+
			"scope is larger, so this looks like the old list", len(ds.Species))
	}
}

// unmeasurable counts the in-scope species AVONET has no row for, in any of
// its three taxonomies, under any name internal/taxonomy knows. They are
// named rather than counted silently, because each is a bird BirdView admits
// and cannot draw.
var unmeasurableNames = []string{
	"Anarhynchus collaris",  // Collared Plover
	"Coloeus monedula",      // Western Jackdaw
	"Daptrius chimachima",   // Yellow-headed Caracara
	"Periporphyrus celaeno", // Crimson-collared Grosbeak
}

const unmeasurable = 4

// TestRarityCodeNeverReachesTheDataset checks the ABA rarity code is used as a
// scope filter and then discarded (P-001).
//
// This inspects field NAMES rather than searching the whole record for "code",
// which an earlier version did. That was too broad: it would now reject the
// four-letter banding code, which is a legitimate part of a species' identity
// and which P-004 requires. The requirement is about rarity specifically.
//
// Verifies: P-001. Criterion: AC-011.
func TestRarityCodeNeverReachesTheDataset(t *testing.T) {
	ds := load(t)
	var fields map[string]any
	if err := json.Unmarshal([]byte(mustJSON(t, ds.Species[0])), &fields); err != nil {
		t.Fatal(err)
	}
	for name := range fields {
		if strings.Contains(strings.ToLower(name), "rarity") {
			t.Errorf("serialized record has field %q; the rarity code is a scope filter, not data", name)
		}
	}
	// And no field should hold a bare 1-6 that could be one.
	if v, ok := fields["abaCode"]; ok {
		t.Errorf("serialized record carries abaCode = %v", v)
	}
}

// TestBandingCodesArePresent checks P-004's newest requirement: the tile shows
// the four-letter code, so the ingest must carry it.
//
// Verifies: P-004. Criterion: AC-013.
func TestSpeciesCodesArePresent(t *testing.T) {
	ds := load(t)
	var missing []string
	for _, s := range ds.Species {
		if s.SpeciesCode == "" {
			missing = append(missing, s.ScientificName)
		}
	}
	// Exactly three species have no eBird code, and they are named rather
	// than counted: each is a taxon AviList recognizes and Clements has no
	// 1:1 counterpart for, so Clements publishes no code. A fourth appearing
	// means the join moved, not that the world changed.
	want := map[string]bool{
		"Oxyura jamaicensis": true, "Pterodroma cervicalis": true, "Tyto furcata": true,
	}
	for _, m := range missing {
		if !want[m] {
			t.Errorf("%s has no eBird species code and is not one of the three known gaps", m)
		}
	}
	if len(missing) != len(want) {
		t.Errorf("%d species have no eBird code, want %d: %v", len(missing), len(want), missing)
	}
}

// TestSpeciesCodesAreUniqueAndURLSafe guards the thing the code now carries
// that an alpha code never did: it IS the URL (P-049).
//
// Verifies: P-049. Criterion: AC-084.
func TestSpeciesCodesAreUniqueAndURLSafe(t *testing.T) {
	ds := load(t)
	seen := map[string]string{}
	for _, s := range ds.Species {
		if other, dup := seen[s.ID]; dup {
			t.Errorf("%s and %s share the id %q, so they share one URL", other, s.ScientificName, s.ID)
		}
		seen[s.ID] = s.ScientificName
		if s.ID != url.PathEscape(s.ID) {
			t.Errorf("%s: id %q needs escaping in a URL", s.ScientificName, s.ID)
		}
		if s.ID == "" {
			t.Errorf("%s: empty id", s.ScientificName)
		}
	}
}

// Verifies: P-004. Criterion: AC-013.
func TestNamesArePopulatedNotDefaulted(t *testing.T) {
	ds := load(t)
	for _, s := range ds.Species {
		switch {
		case s.Name(dataset.LocaleWorld) == "":
			t.Errorf("%s: empty world name", s.ScientificName)
		case s.Name(dataset.LocaleNA) == "":
			t.Errorf("%s: empty North American name, and the world fallback is empty too", s.ScientificName)
		case s.Family == "" || s.FamilyCommon == "":
			t.Errorf("%s: empty family (%q / %q)", s.ScientificName, s.Family, s.FamilyCommon)
		case s.Name(dataset.LocaleNA) == s.ScientificName:
			// The tell of a silent fallback: an earlier design shipped this for
			// every one of its species.
			t.Errorf("%s: common name equals the scientific name", s.ScientificName)
		case s.FamilyColor == "":
			t.Errorf("%s: no family color", s.ScientificName)
		}
	}
}

// Verifies: P-041. Criterion: AC-030.
func TestSegmentsSumToTotalExactly(t *testing.T) {
	ds := load(t)
	for _, s := range ds.Species {
		sum := s.Segments.Bill.Value + s.Segments.Head.Value + s.Segments.Neck.Value +
			s.Segments.Body.Value + s.Segments.Tail.Value
		if math.Abs(sum-s.TotalLength.Value) > 1e-9 {
			t.Errorf("%s: segments sum to %.4f, total is %.4f", s.ScientificName, sum, s.TotalLength.Value)
		}
	}
}

// Verifies: P-062. Criterion: AC-040.
func TestEveryMetricHasWellFormedProvenance(t *testing.T) {
	ds := load(t)
	for i := range ds.Species {
		s := &ds.Species[i]
		for name, m := range s.AllMetrics() {
			p := m.Prov
			switch p.Kind {
			case dataset.Measured, dataset.NearMatch:
				if p.Source == "" {
					t.Errorf("%s/%s: %s value with no source", s.ScientificName, name, p.Kind)
				}
				if p.SourceTaxon == "" {
					t.Errorf("%s/%s: %s value with no sourceTaxon", s.ScientificName, name, p.Kind)
				}
			case dataset.Modeled:
				if strings.TrimSpace(p.Model) == "" {
					t.Errorf("%s/%s: modeled value with no model named", s.ScientificName, name)
				}
			default:
				t.Errorf("%s/%s: unknown provenance kind %q", s.ScientificName, name, p.Kind)
			}
			// near-match means the measurement belongs to a DIFFERENT taxon.
			// If the taxon matches the species, the label is wrong.
			if p.Kind == dataset.NearMatch && p.SourceTaxon == s.ScientificName {
				t.Errorf("%s/%s: near-match names its own taxon", s.ScientificName, name)
			}
		}
	}
}

// Verifies: P-060, P-062. Criterion: AC-041.
func TestProvenanceMatchesTheSourcesUsed(t *testing.T) {
	ds := load(t)
	var measuredSpan, nearSpan, modeledSpan int
	for i := range ds.Species {
		s := &ds.Species[i]
		if s.TotalLength.Prov.Kind != dataset.Modeled {
			t.Errorf("%s: length is %q; no measured length source exists (CR-002)",
				s.ScientificName, s.TotalLength.Prov.Kind)
		}
		switch s.Wingspan.Prov.Kind {
		case dataset.Measured:
			measuredSpan++
		case dataset.NearMatch:
			nearSpan++
		default:
			modeledSpan++
		}
	}
	// There used to be an absolute count here -- 301 less one, measured
	// against the Fox dataset for the ABA checklist's scope and recorded in
	// spec/sources/fox-wingspan/PROVENANCE.md. CR-028 replaced that scope, so
	// the number is a fact about a species list that no longer exists, and
	// restating it as whatever this run produces would be reading the
	// expectation off the code under test.
	//
	// What survives the change is the PROPERTY, which is what P-060 actually
	// requires: a wingspan is `measured` only if an empirical one was found,
	// and an empirical one is only ever Fox's. A regression that modeled a
	// span and called it measured fails here however many species are in
	// scope.
	for _, sp := range ds.Species {
		p := sp.Wingspan.Prov
		switch p.Kind {
		case dataset.Measured, dataset.NearMatch:
			if !strings.Contains(p.Source, "Fox") {
				t.Errorf("%s: wingspan is %s but cites %q; the only empirical "+
					"wingspans are Fox's", sp.ScientificName, p.Kind, p.Source)
			}
		case dataset.Modeled:
			if p.Model == "" {
				t.Errorf("%s: wingspan is modeled and names no model", sp.ScientificName)
			}
		}
	}
	if measuredSpan == 0 {
		t.Error("no measured wingspans at all; the Fox join has broken")
	}
	if measuredSpan+nearSpan+modeledSpan != len(ds.Species) {
		t.Error("wingspan kinds do not partition the dataset")
	}
}

// Verifies: P-045, P-062. Criterion: AC-035.
func TestNeckIsModeledAndNamesItsFamilyRatio(t *testing.T) {
	ds := load(t)
	for i := range ds.Species {
		s := &ds.Species[i]
		p := s.Segments.Neck.Prov
		if p.Kind != dataset.Modeled {
			t.Fatalf("%s: neck kind is %q, want modeled", s.ScientificName, p.Kind)
		}
		if !strings.Contains(p.Model, s.Family) {
			t.Errorf("%s: neck model %q does not name the family", s.ScientificName, p.Model)
		}
		if !strings.ContainsAny(p.Model, "0123456789") {
			t.Errorf("%s: neck model %q names no ratio", s.ScientificName, p.Model)
		}
	}
}

// Verifies: P-060, P-041. Criterion: AC-044.
func TestTotalLengthSaysItIsASum(t *testing.T) {
	ds := load(t)
	m := ds.Species[0].TotalLength.Prov.Model
	for _, want := range []string{"sum", "measured", "modeled"} {
		if !strings.Contains(strings.ToLower(m), want) {
			t.Errorf("total length model %q does not mention %q; a reader seeing a measured "+
				"bill inside a modeled total needs it to explain how both are true", m, want)
		}
	}
}

// Verifies: P-006, P-007. Criterion: AC-014.
func TestNearMatchesAreExactlyTheMappedOnes(t *testing.T) {
	ds := load(t)
	got := map[string]bool{}
	for i := range ds.Species {
		s := &ds.Species[i]
		if s.Segments.Bill.Prov.Kind == dataset.NearMatch {
			got[s.ScientificName] = true
		}
	}
	for name := range taxonomy.NearMatches {
		if _, dropped := taxonomy.NotInAviList[name]; dropped {
			// The species left the dataset with CR-022, so there is nothing
			// to label. The mapping stays: it records a real AVONET
			// relationship and would be needed again if AviList split the
			// taxon back out.
			continue
		}
		if !got[name] {
			t.Errorf("%s is mapped as a near-match but its traits are not labeled so", name)
		}
		delete(got, name)
	}
	for name := range got {
		t.Errorf("%s is labeled near-match but appears in no mapping", name)
	}
}

// Verifies: P-060, P-062. Criterion: AC-116.
func TestNearMatchBorrowsFromTheDeclaredSheet(t *testing.T) {
	wb, err := avonet.Open(avoFile)
	if err != nil {
		t.Fatal(err)
	}
	ds := load(t)

	// The expected values come from the workbook, and must differ from the
	// eBird row, or the check cannot tell the two sheets apart.
	want := wb.BirdLife["Gygis alba"]
	broad := wb.EBird["Gygis alba"]
	if want.Culmen == broad.Culmen || want.Tail == broad.Tail {
		t.Fatalf("BirdLife and eBird Gygis alba agree (culmen %v/%v, tail %v/%v); "+
			"this test can no longer distinguish them", want.Culmen, broad.Culmen, want.Tail, broad.Tail)
	}

	seen := 0
	for i := range ds.Species {
		s := &ds.Species[i]
		if s.Segments.Bill.Prov.Kind != dataset.NearMatch {
			continue
		}
		seen++
		// Written out rather than read from taxonomy.NearMatchSheets, so a
		// second species moved off the eBird sheet has to be decided here too.
		sheet := avonet.SheetEBird
		switch s.ScientificName {
		case "Gygis candida", "Larus vegae": // CR-039, CR-040
			sheet = avonet.SheetBirdLife
		}
		for _, m := range []struct {
			name string
			got  dataset.Metric
		}{{"bill", s.Segments.Bill}, {"tail", s.Segments.Tail}} {
			if !strings.HasSuffix(m.got.Prov.Source, "sheet "+sheet) {
				t.Errorf("%s %s comes from %q, want sheet %s", s.ScientificName, m.name, m.got.Prov.Source, sheet)
			}
		}
		if s.ScientificName != "Gygis candida" {
			continue
		}
		if s.Segments.Bill.Prov.SourceTaxon != "Gygis alba" {
			t.Errorf("Gygis candida borrows from %q, want Gygis alba", s.Segments.Bill.Prov.SourceTaxon)
		}
		if s.Segments.Bill.Value != want.Culmen || s.Segments.Tail.Value != want.Tail {
			t.Errorf("Gygis candida has bill %v, tail %v; BirdLife Gygis alba has %v, %v",
				s.Segments.Bill.Value, s.Segments.Tail.Value, want.Culmen, want.Tail)
		}
	}
	if seen == 0 {
		t.Fatal("no near-match species in the dataset; the check ran over nothing")
	}

	// A declared sheet that lacks the parent stops the build.
	if err := checkNearMatchSheets(wb, map[string]string{"Gygis candida": avonet.SheetBirdTree + "X"}); err == nil {
		t.Error("a near-match sheet that holds no row for the parent was accepted")
	}
}

func mustJSON(t *testing.T, v any) string {
	t.Helper()
	b, err := jsonMarshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func jsonMarshal(v any) ([]byte, error) { return json.Marshal(v) }

// TestHeadIsNotAConstant guards the regression that prompted the model change:
// the cranium must vary with the bird, not sit on a floor.
//
// The previous model pinned 442 of 841 species at exactly 12 mm, so half the
// dataset drew an identical head that consumed a fixed share of every small
// bird's length and differentiated nothing.
//
// Verifies: P-062. Criterion: AC-065.
func TestHeadIsNotAConstant(t *testing.T) {
	ds := load(t)
	counts := map[float64]int{}
	for i := range ds.Species {
		counts[ds.Species[i].Segments.Head.Value]++
	}
	var worst float64
	var worstN int
	for v, n := range counts {
		if n > worstN {
			worst, worstN = v, n
		}
	}
	// A floor shows up as a large pile on one value. Some sharing is expected,
	// since mass is reported to limited precision.
	if frac := float64(worstN) / float64(len(ds.Species)); frac > 0.05 {
		t.Errorf("%d of %d species (%.0f%%) share head length %.1f mm; the cranium has "+
			"collapsed to a constant", worstN, len(ds.Species), frac*100, worst)
	}
	// And it must track mass: the heaviest bird needs the largest head.
	var heaviest, biggestHead *dataset.Species
	for i := range ds.Species {
		s := &ds.Species[i]
		if heaviest == nil || s.Mass.Value > heaviest.Mass.Value {
			heaviest = s
		}
		if biggestHead == nil || s.Segments.Head.Value > biggestHead.Segments.Head.Value {
			biggestHead = s
		}
	}
	if heaviest.ScientificName != biggestHead.ScientificName {
		t.Errorf("heaviest bird is %s but the largest head belongs to %s",
			heaviest.ScientificName, biggestHead.ScientificName)
	}
}

// TestINatTableMatchesTheDataset guards the committed mapping against drift.
//
// The table is generated rarely and by hand, so it is exactly the kind of
// artifact that rots quietly: a species renamed by a taxonomic revision leaves a
// row behind that matches nothing, and the link simply stops appearing with
// nobody told. This makes that loud.
//
// It does NOT require full coverage. A species iNaturalist has not confidently
// matched is supposed to be absent (P-064) -- usually a recent split it has not
// followed yet -- so absence is correct and only a *stale* row is a fault.
//
// Verifies: P-064, T-022. Criterion: AC-068.
func TestINatTableMatchesTheDataset(t *testing.T) {
	f, err := os.Open(inatFile)
	if err != nil {
		t.Skipf("skipping: %s is absent; run cmd/birdview-inatmap", inatFile)
	}
	defer f.Close()
	table, err := inat.Load(f)
	if err != nil {
		t.Fatalf("the committed mapping does not parse: %v", err)
	}

	ds := load(t)
	known := map[string]string{}
	for i := range ds.Species {
		known[ds.Species[i].ScientificName] = ds.Species[i].Name(dataset.LocaleNA)
	}

	var stale []string
	for sci := range table {
		if _, ok := known[sci]; !ok {
			stale = append(stale, sci)
		}
	}
	if len(stale) > 0 {
		sort.Strings(stale)
		t.Errorf("%d rows name species not in the dataset, so their links are dead: %v",
			len(stale), stale)
	}

	// Every ID positive, and no two species pointing at the same taxon -- which
	// would mean the matcher accepted the same bird twice under different names.
	seen := map[int]string{}
	for sci, e := range table {
		if e.TaxonID <= 0 {
			t.Errorf("%s has taxon id %d", sci, e.TaxonID)
		}
		if prev, dup := seen[e.TaxonID]; dup {
			t.Errorf("taxon %d is claimed by both %s and %s", e.TaxonID, prev, sci)
		}
		seen[e.TaxonID] = sci
	}

	// Coverage is REQUIRED to be total (P-064, CR-016). Every one of the 30 that
	// once failed turned out to be a formatting or synonym artifact, not a real
	// disagreement, so an unresolved species now means something is wrong --
	// a taxonomic change, a normalizer regression, or a stale table -- rather
	// than a bird iNaturalist does not have.
	var missing []string
	for i := range ds.Species {
		if _, ok := table[ds.Species[i].ScientificName]; !ok {
			missing = append(missing, ds.Species[i].ScientificName)
		}
	}
	if len(missing) > 0 {
		sort.Strings(missing)
		t.Errorf("%d of %d species have no iNaturalist link; coverage must be total. "+
			"Re-run cmd/birdview-inatmap, and if any still fail, diagnose why before "+
			"accepting the gap: %v", len(missing), len(ds.Species), missing)
	}
}

// Region ids and their record counts travel as parallel slices, so the
// invariant that they line up has to be asserted rather than assumed.
//
// A map would make this impossible to get wrong and was rejected for size:
// it repeats each region id as a JSON string across 27,646 entries. That
// trade is only sound if the pairing is checked.
//
// Watched fail: appending one extra count in applyRegions reports the
// mismatch and names the species.
//
// Verifies: P-074, CR-030. Criterion: AC-096, AC-099.
func TestRegionCountsLineUp(t *testing.T) {
	ds := load(t)
	var withCounts int
	for _, sp := range ds.Species {
		if len(sp.Regions) != len(sp.RegionCounts) {
			t.Errorf("%s has %d regions and %d counts", sp.ScientificName,
				len(sp.Regions), len(sp.RegionCounts))
		}
		for i, n := range sp.RegionCounts {
			if n <= 0 {
				t.Errorf("%s is listed in %s with %d records; presence means it was "+
					"reported there", sp.ScientificName, sp.Regions[i], n)
			}
		}
		if len(sp.RegionCounts) > 0 {
			withCounts++
		}
	}
	if withCounts == 0 {
		t.Fatal("no species carries region counts; the share cannot be computed and " +
			"this test is checking nothing")
	}

	// Every region a species is placed in must have a total to divide by,
	// or its share is silently uncomputable.
	//
	// This briefly excluded the seas, under CR-030, because a sea's figure
	// summed only the species already judged pelagic -- a denominator built
	// out of its own numerators, which is not the quantity a land total is.
	// CR-031 fixed the number rather than hiding it: a sea's total is now
	// every bird record on open water in its box. So the assertion reverts,
	// and a second one below pins the thing that was actually wrong.
	if len(ds.RegionTotals) == 0 {
		t.Fatal("the dataset carries no region totals")
	}
	for _, sp := range ds.Species {
		for i, id := range sp.Regions {
			total, ok := ds.RegionTotals[id]
			if !ok {
				t.Errorf("%s is in %s, which has no total", sp.ScientificName, id)
				continue
			}
			if sp.RegionCounts[i] > total {
				t.Errorf("%s has %d records in %s, more than the %d the region holds",
					sp.ScientificName, sp.RegionCounts[i], id, total)
			}
		}
	}

	// A sea's denominator must not be the sum of its own numerators.
	//
	// This is the shape of the CR-030 bug, and it is the only part a test
	// can see offline. A sea's total is every bird on that water, while the
	// species BirdView places at sea are a few dozen specialists, so the
	// total runs 8x to 200x the sum -- the tightest is the North Pacific at
	// 11.9%. Equality means the pelagic-only sum is back, and equality is
	// exactly what the old table had, by construction.
	//
	// Land is deliberately exempt: there the drawn species really are
	// almost all the records (98-100%, and Navassa is 6 of 6), so no
	// margin exists to assert.
	summed := map[string]int64{}
	for _, sp := range ds.Species {
		for i, id := range sp.Regions {
			summed[id] += sp.RegionCounts[i]
		}
	}
	for _, id := range region.MarineIDs() {
		total, ok := ds.RegionTotals[id]
		if !ok || summed[id] == 0 {
			continue
		}
		if total <= summed[id] {
			t.Errorf("%s totals %d against %d summed over the species placed there; "+
				"a sea's denominator is every bird on that water, not a sum of the "+
				"birds we happen to draw (CR-031)", id, total, summed[id])
		}
	}
}

// Verifies: P-079, CR-040. Criterion: AC-117.
func TestVegaGullSharesAmericanHerringGullsRow(t *testing.T) {
	ds := load(t)
	got := map[string]*dataset.Species{}
	for i := range ds.Species {
		got[ds.Species[i].ScientificName] = &ds.Species[i]
	}
	vega, amhg := got["Larus vegae"], got["Larus smithsonianus"]
	if vega == nil || amhg == nil {
		t.Fatal("Vega Gull or American Herring Gull is not drawn")
	}
	for _, s := range []*dataset.Species{vega, amhg} {
		p := s.Segments.Bill.Prov
		if !strings.HasSuffix(p.Source, "sheet "+avonet.SheetBirdLife) || p.SourceTaxon != "Larus smithsonianus" {
			t.Errorf("%s is drawn from %q %q, want AVONET1_BirdLife Larus smithsonianus", s.ScientificName, p.Source, p.SourceTaxon)
		}
	}
	if vega.Segments.Bill.Prov.Kind != dataset.NearMatch {
		t.Errorf("Vega Gull's bill is %q, want near-match", vega.Segments.Bill.Prov.Kind)
	}
	for _, m := range []struct {
		name string
		a, b float64
	}{
		{"bill", vega.Segments.Bill.Value, amhg.Segments.Bill.Value},
		{"tail", vega.Segments.Tail.Value, amhg.Segments.Tail.Value},
		{"mass", vega.Mass.Value, amhg.Mass.Value},
		{"wingspan", vega.Wingspan.Value, amhg.Wingspan.Value},
	} {
		if m.a != m.b {
			t.Errorf("Vega Gull's %s is %v and American Herring Gull's %v; one row should give one value", m.name, m.a, m.b)
		}
	}
}

// Verifies: P-060, P-062, CR-041. Criterion: AC-118.
func TestEuropeanHerringGullReadsItsOwnRow(t *testing.T) {
	wb, err := avonet.Open(avoFile)
	if err != nil {
		t.Fatal(err)
	}
	want, broad := wb.BirdLife["Larus argentatus"], wb.EBird["Larus argentatus"]
	if want.Culmen == broad.Culmen || want.Tail == broad.Tail {
		t.Fatalf("BirdLife and eBird Larus argentatus agree (culmen %v/%v, tail %v/%v); "+
			"this test can no longer distinguish them", want.Culmen, broad.Culmen, want.Tail, broad.Tail)
	}

	ds := load(t)
	got := map[string]*dataset.Species{}
	for i := range ds.Species {
		got[ds.Species[i].ScientificName] = &ds.Species[i]
	}
	for _, c := range []struct {
		sci, sheet string
	}{{"Larus argentatus", avonet.SheetBirdLife}, {"Larus smithsonianus", avonet.SheetBirdLife}} {
		s := got[c.sci]
		if s == nil {
			t.Fatalf("%s is not drawn", c.sci)
		}
		for _, m := range []struct {
			name string
			got  dataset.Metric
		}{{"bill", s.Segments.Bill}, {"tail", s.Segments.Tail}, {"mass", s.Mass}} {
			p := m.got.Prov
			if p.Kind != dataset.Measured || !strings.HasSuffix(p.Source, "sheet "+c.sheet) || p.SourceTaxon != c.sci {
				t.Errorf("%s %s is %s from %q %q, want measured from %s %s",
					c.sci, m.name, p.Kind, p.Source, p.SourceTaxon, c.sheet, c.sci)
			}
		}
	}
	if s := got["Larus argentatus"]; s.Segments.Bill.Value != want.Culmen || s.Segments.Tail.Value != want.Tail {
		t.Errorf("European Herring Gull has bill %v, tail %v; BirdLife Larus argentatus has %v, %v",
			s.Segments.Bill.Value, s.Segments.Tail.Value, want.Culmen, want.Tail)
	}

	if err := checkOwnRowSheets(wb, map[string]string{"Larus vegae": avonet.SheetBirdLife}); err == nil {
		t.Error("an own-row sheet that holds no row for the species was accepted")
	}
}

// TestBelowBarRegionsAreCountedNotListed asserts AC-119's second assertion:
// the ingest keeps a below-bar cell out of Regions, so no region or month
// filter sees it, and carries its counts in the Below fields.
//
// Verifies: P-011, CR-042. Criterion: AC-119.
func TestBelowBarRegionsAreCountedNotListed(t *testing.T) {
	table := "total\tUSA.44_1\t10|10|10|10|10|10|10|10|10|10|10|10\n" +
		"total\tUSA.5_1\t10|10|10|10|10|10|10|10|10|10|10|10\n" +
		"Dromaius novaehollandiae\tUSA.44_1~:2|1|3|1|2|1|1|2|2|1|2|1,USA.5_1:1|1|1|1|1|1|1|1|1|1|1|1\n"
	path := t.TempDir() + "/presence.tsv"
	if err := os.WriteFile(path, []byte(table), 0o644); err != nil {
		t.Fatal(err)
	}
	ds := &dataset.Dataset{Species: []dataset.Species{{ScientificName: "Dromaius novaehollandiae"}}}
	n, err := applyRegions(ds, path)
	if err != nil {
		t.Fatal(err)
	}
	s := ds.Species[0]
	if len(s.Regions) != 1 || s.Regions[0] != "USA.5_1" || s.RegionCounts[0] != 12 {
		t.Errorf("regions %v counts %v, want only USA.5_1 with 12", s.Regions, s.RegionCounts)
	}
	if len(s.BelowRegions) != 1 || s.BelowRegions[0] != "USA.44_1" || s.BelowCounts[0] != 19 ||
		len(s.BelowMonths) != 1 || s.BelowMonths[0][2] != 3 {
		t.Errorf("below regions %v counts %v months %v, want USA.44_1 with 19", s.BelowRegions, s.BelowCounts, s.BelowMonths)
	}
	if n != 1 {
		t.Errorf("%d species placed, want 1", n)
	}
}
