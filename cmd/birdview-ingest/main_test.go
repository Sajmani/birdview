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
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"birdview/internal/avonet"
	"birdview/internal/dataset"
	"birdview/internal/region"
	"birdview/internal/status"
	"birdview/internal/taxonomy"
	"birdview/internal/trait"
)

const (
	statusFile        = "../../data/status.tsv"
	avoFile           = "../../third_party/avonet/AVONET Supplementary dataset 1.xlsx"
	foxFile           = "../../third_party/foxwingspan/wingspan_fox2026_AppendixS1.xlsx"
	avlFile           = "../../third_party/avilist/AviList-v2025b-10Jun2026-extended.xlsx"
	regFile           = "../../data/region-presence.tsv"
	countyDir         = "../../data/county-presence"
	habitatReviewFile = "../../data/habitat-review.tsv"
	neckReviewFile    = "../../data/neck-review.tsv"
	bblFile           = "../../third_party/bbl/speclist.html"
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
	// (decisions.scopeFromRecurrence), so this no longer skips on a clean checkout. That is the
	// point of the change: the build depends on nothing it may not
	// redistribute.
	if err := loadHabitatReview(habitatReviewFile); err != nil {
		t.Fatalf("loadHabitatReview: %v", err)
	}
	if err := loadNeckReview(neckReviewFile); err != nil {
		t.Fatalf("loadNeckReview: %v", err)
	}
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
	if _, err := applyCounties(ds, countyDir); err != nil {
		t.Fatalf("applyCounties: %v", err)
	}
	if _, err := applyBanding(ds, bblFile); err != nil {
		t.Fatalf("applyBanding: %v", err)
	}
	built = ds
	return ds
}

// Scope is now every species the status table says recurs, plus the declared
// exceptions, less what cannot be named or measured (product.scopeIsRecurringAvifauna, decisions.scopeFromRecurrence).
//
// The expected count is derived from those inputs rather than from a run:
// count the in-scope rows, add the declared, subtract the names AviList has
// no concept for. Writing the total as a literal would make this test agree
// with whatever the code last produced.
//
// Verifies: product.scopeIsRecurringAvifauna. Criterion: acceptance.scopeMatchesStatusTable.
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
	// subtracted: Vireo swainsoni left scope under decisions.countEbirdFirst, and subtracting
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
// scope filter and then discarded (product.scopeIsRecurringAvifauna).
//
// This inspects field NAMES rather than searching the whole record for "code",
// which an earlier version did. That was too broad: it would now reject the
// four-letter banding code, which is a legitimate part of a species' identity
// and which product.namesFromAvilist requires. The requirement is about rarity specifically.
//
// Verifies: product.scopeIsRecurringAvifauna. Criterion: acceptance.noRarityCodes.
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

// TestBandingCodesArePresent checks product.namesFromAvilist's newest requirement: the tile shows
// the four-letter code, so the ingest must carry it.
//
// Verifies: product.namesFromAvilist. Criterion: acceptance.namesNeverDefaulted.
func TestSpeciesCodesArePresent(t *testing.T) {
	ds := load(t)
	var missing []string
	for _, s := range ds.Species {
		if s.SpeciesCode == "" {
			missing = append(missing, s.ScientificName)
		}
	}
	// None is left without a code (product.ebirdLink). AviList gives none to exactly three, and
	// they are named rather than counted: each is a taxon AviList recognizes
	// and Clements has no 1:1 counterpart for, so taxonomy.EBirdCodes supplies
	// eBird's code. A fourth appearing means the join moved, not that the
	// world changed, and it shows here as a species with no code at all.
	if len(missing) > 0 {
		t.Errorf("%d species have no eBird code, and no declared one: %v", len(missing), missing)
	}
	gaps := map[string]string{
		"Oxyura jamaicensis": "rudduc", "Pterodroma cervicalis": "whnpet", "Tyto furcata": "brnowl",
	}
	for i := range ds.Species {
		sp := &ds.Species[i]
		want, gap := gaps[sp.ScientificName]
		if !gap {
			continue
		}
		if sp.SpeciesCode != want || !slices.Contains(sp.CodeAliases, strings.ToLower(strings.ReplaceAll(sp.ScientificName, " ", "_"))) {
			t.Errorf("%s: code %q, aliases %v; AviList gives it none, so it takes %q from taxonomy.EBirdCodes and keeps its old address as an alias",
				sp.ScientificName, sp.SpeciesCode, sp.CodeAliases, want)
		}
	}
}

// TestSpeciesCodesAreUniqueAndURLSafe guards the thing the code now carries
// that an alpha code never did: it IS the URL (product.englishNameChoice).
//
// Verifies: product.englishNameChoice. Criterion: acceptance.stableSpeciesId.
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

// Verifies: product.namesFromAvilist. Criterion: acceptance.namesNeverDefaulted.
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

// Verifies: product.segmentsSumToLength. Criterion: acceptance.segmentsSumExactly.
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

// Verifies: product.provenancePerMetric. Criterion: acceptance.provenanceWellFormed.
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

// Verifies: product.lengthAndWingspanSources, product.provenancePerMetric. Criterion: acceptance.provenanceMatchesSources.
func TestProvenanceMatchesTheSourcesUsed(t *testing.T) {
	ds := load(t)
	var measuredSpan, nearSpan, modeledSpan int
	for i := range ds.Species {
		s := &ds.Species[i]
		if s.TotalLength.Prov.Kind != dataset.Modeled {
			t.Errorf("%s: length is %q; no measured length source exists (decisions.ricklefsProhibited)",
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
	// river/requirements/imported/figshare.com/foxwingspan/PROVENANCE.md. decisions.scopeFromRecurrence replaced that scope, so
	// the number is a fact about a species list that no longer exists, and
	// restating it as whatever this run produces would be reading the
	// expectation off the code under test.
	//
	// What survives the change is the PROPERTY, which is what product.lengthAndWingspanSources actually
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

// Verifies: product.neckModelPerFamily, product.provenancePerMetric. Criterion: acceptance.neckLengthSerialized.
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

// Verifies: product.lengthAndWingspanSources, product.segmentsSumToLength. Criterion: acceptance.totalIsSumOfSegments.
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

// Verifies: product.joinOnBinomial, product.unmatchedSpeciesReported. Criterion: acceptance.joinCoversEverySpecies.
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
			// The species left the dataset with decisions.avilistReplacesAba, so there is nothing
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

// Verifies: product.lengthAndWingspanSources, product.provenancePerMetric. Criterion: acceptance.whiteTernFromBirdlife.
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
		case "Gygis candida", "Larus vegae": // decisions.whiteTernBorrows, decisions.vegaGullBorrows
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
// Verifies: product.provenancePerMetric. Criterion: acceptance.craniumVaries.
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
// Verifies: product.shareOfRecords, decisions.seaShareDiffers. Criterion: acceptance.commonestShareIsShare, acceptance.commonestIsLargestShare.
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
	// This briefly excluded the seas, under decisions.seaShareDiffers, because a sea's figure
	// summed only the species already judged pelagic -- a denominator built
	// out of its own numerators, which is not the quantity a land total is.
	// decisions.honestSeaDenominator fixed the number rather than hiding it: a sea's total is now
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
	// This is the shape of the decisions.seaShareDiffers bug, and it is the only part a test
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
				"birds we happen to draw (decisions.honestSeaDenominator)", id, total, summed[id])
		}
	}
}

// Verifies: product.borrowedMeasurementsSaySo, decisions.vegaGullBorrows. Criterion: acceptance.sharedMeasurementsDisclosed.
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

// Verifies: product.lengthAndWingspanSources, product.provenancePerMetric, decisions.herringGullEuropeanRow. Criterion: acceptance.herringGullFromEuropeanRow.
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

// TestBelowBarRegionsAreCountedNotListed asserts acceptance.belowBarCountsKept's second assertion:
// the ingest keeps a below-bar cell out of Regions, so no region or month
// filter sees it, and carries its counts in the Below fields.
//
// Verifies: product.regionAndMonthFilter, decisions.belowBarCounted. Criterion: acceptance.belowBarCountsKept.
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

// Verifies: product.statusIsPublished, product.escapedBirdsListed, decisions.escapedBirdsOffIndex. Criterion: acceptance.escapedBirdsListedAndServed.
func TestExoticsAreFlaggedFromTheList(t *testing.T) {
	ds := load(t)
	if len(status.Exotic) < 30 {
		t.Fatalf("status.Exotic lists %d species; the check has too little to assert", len(status.Exotic))
	}
	flagged := 0
	for i := range ds.Species {
		s := &ds.Species[i]
		want := status.Exotic[s.ScientificName]
		if s.Exotic != want {
			t.Errorf("%s carries exotic %q, want %q", s.ScientificName, s.Exotic, want)
		}
		if s.Exotic != "" {
			flagged++
		}
	}
	if flagged != len(status.Exotic) {
		t.Errorf("%d species flagged, but status.Exotic lists %d", flagged, len(status.Exotic))
	}
	out := []dataset.Species{{ScientificName: "Ara macao"}}
	if err := flagExotic(out, map[string]string{"Ara macao": "x", "Nonesuch nonesuch": "y"}); err == nil {
		t.Error("a listed species that is not drawn was accepted")
	}
}

// TestBandingCodes checks the four-letter codes the ingest joins from the
// USGS Bird Banding Laboratory's list: known birds carry the Laboratory's
// code, every code is four capital letters and names one species, and most
// of the dataset is covered (the birds left out are those the Laboratory
// does not band with federal bands: introduced birds, most gamebirds).
//
// Verifies: product.bandingCodes. Criterion: app.bandingCodesShown.
func TestBandingCodes(t *testing.T) {
	ds := load(t)
	by := map[string]*dataset.Species{}
	seen := map[string]string{}
	n := 0
	for i := range ds.Species {
		sp := &ds.Species[i]
		by[sp.ID] = sp
		if sp.BandingCode == "" {
			continue
		}
		n++
		if len(sp.BandingCode) != 4 || strings.ToUpper(sp.BandingCode) != sp.BandingCode {
			t.Errorf("%s: banding code %q is not four capital letters", sp.ID, sp.BandingCode)
		}
		if other, dup := seen[sp.BandingCode]; dup {
			t.Errorf("%s and %s share the banding code %s", other, sp.ID, sp.BandingCode)
		}
		seen[sp.BandingCode] = sp.ID
	}
	for id, code := range map[string]string{"amerob": "AMRO", "blujay": "BLJA", "amecro": "AMCR", "norcar": "NOCA", "baleag": "BAEA"} {
		if sp := by[id]; sp == nil || sp.BandingCode != code {
			t.Errorf("%s has banding code %q, want %s", id, sp.BandingCode, code)
		}
	}
	if n < 850 {
		t.Errorf("only %d of %d species have a banding code", n, len(ds.Species))
	}
}

// TestBandingJoin checks the join's rules on a small list: by scientific
// name; where two rows share one, the row named as the species is; with no
// scientific match, by the North American English name when one row has it;
// and a code two species claim goes to neither.
//
// Verifies: product.bandingCodes. Criterion: app.bandingCodesShown.
func TestBandingJoin(t *testing.T) {
	row := func(num, code, common, sci string) string {
		return "<tr><td>" + num + "</td><td>" + code + "</td><td>" + common + "</td><td>2</td><td>x</td><td>" +
			sci + "</td><td></td><td></td><td>1</td></tr>"
	}
	page := row("1", "AMRO", "American Robin", "Turdus migratorius") +
		row("2", "CAGN", "California Gnatcatcher", "Polioptila californica") +
		row("3", "BACG", "Baja Gnatcatcher", "Polioptila californica") +
		row("4", "OLDN", "Renamed Bird", "Oldus genus") +
		row("5", "TWIN", "Twin One", "Twinus one") +
		row("6", "TWIN", "Twin Two", "Twinus two")
	f := t.TempDir() + "/list.html"
	if err := os.WriteFile(f, []byte(page), 0o644); err != nil {
		t.Fatal(err)
	}
	mk := func(id, na, sci string) dataset.Species {
		return dataset.Species{ID: id, ScientificName: sci, Names: map[dataset.Locale]string{dataset.LocaleNA: na}}
	}
	ds := &dataset.Dataset{Species: []dataset.Species{
		mk("amerob", "American Robin", "Turdus migratorius"),
		mk("calgna", "California Gnatcatcher", "Polioptila californica"),
		mk("renbir", "Renamed Bird", "Newus genus"),
		mk("twin1", "Twin One", "Twinus one"),
		mk("twin2", "Twin Two", "Twinus two"),
	}}
	n, err := applyBanding(ds, f)
	if err != nil {
		t.Fatal(err)
	}
	got := map[string]string{}
	for _, sp := range ds.Species {
		got[sp.ID] = sp.BandingCode
	}
	want := map[string]string{"amerob": "AMRO", "calgna": "CAGN", "renbir": "OLDN", "twin1": "", "twin2": ""}
	for id, code := range want {
		if got[id] != code {
			t.Errorf("%s got %q, want %q", id, got[id], code)
		}
	}
	if n != 3 {
		t.Errorf("applyBanding counted %d, want 3", n)
	}
}

// TestEverySpeciesHasAnEBirdCode asserts product.ebirdLink's "every species has a code": no
// drawn species is left without one, each species' own code is the simplest of its codes and
// its id, and the rest, with any address it had before, are aliases no other species claims.
//
// Verifies: product.ebirdLink. Criterion: app.everySpeciesHasAnEBirdCode.
func TestEverySpeciesHasAnEBirdCode(t *testing.T) {
	ds := load(t)
	bySci := map[string]*dataset.Species{}
	var missing []string
	for i := range ds.Species {
		sp := &ds.Species[i]
		bySci[sp.ScientificName] = sp
		if sp.SpeciesCode == "" {
			missing = append(missing, sp.ScientificName)
		}
	}
	if len(missing) > 0 {
		t.Errorf("%d species have no eBird code, so no eBird link: %v", len(missing), missing)
	}
	// The species' own code is the simplest, and its id; every other code,
	// and the address it had before, is an alias (decisions.ebirdCodeAliases).
	for _, c := range []struct {
		sci, code string
		aliases   []string
	}{
		{"Oxyura jamaicensis", "rudduc", []string{"oxyura_jamaicensis"}},
		{"Pterodroma cervicalis", "whnpet", []string{"pterodroma_cervicalis"}},
		{"Tyto furcata", "brnowl", []string{"tyto_furcata"}},
		{"Apus pacificus", "fotswi", []string{"fotswi1"}},
		{"Molothrus aeneus", "brocow", []string{"brocow1"}},
		{"Saxicola maurus", "sibsto1", []string{"y00427"}},
		{"Cistothorus palustris", "marwre", []string{"eamwre1"}},
		{"Tringa semipalmata", "willet1", []string{"willet2"}},
		{"Anas platyrhynchos", "mallar3", nil},
	} {
		sp, ok := bySci[c.sci]
		if !ok {
			t.Errorf("%s is not in the dataset", c.sci)
			continue
		}
		if sp.SpeciesCode != c.code || sp.ID != c.code {
			t.Errorf("%s has code %q and id %q, want both %q", c.sci, sp.SpeciesCode, sp.ID, c.code)
		}
		if !slices.Equal(sp.CodeAliases, c.aliases) {
			t.Errorf("%s answers also to %v, want %v", c.sci, sp.CodeAliases, c.aliases)
		}
	}
	// Every code names one species, drawn or not.
	owner := map[string]string{}
	claim := func(code, sci string) {
		if prev, dup := owner[code]; dup && prev != sci {
			t.Errorf("code %q answers for both %s and %s", code, prev, sci)
		}
		owner[code] = sci
	}
	for i := range ds.Species {
		for _, c := range append([]string{ds.Species[i].ID}, ds.Species[i].CodeAliases...) {
			claim(c, ds.Species[i].ScientificName)
		}
	}
	for i := range ds.Undrawn {
		for _, c := range append([]string{ds.Undrawn[i].ID}, ds.Undrawn[i].CodeAliases...) {
			claim(c, ds.Undrawn[i].ScientificName)
		}
	}
	seen := map[string]string{}
	for i := range ds.Species {
		sp := &ds.Species[i]
		if prev, dup := seen[sp.SpeciesCode]; dup && sp.SpeciesCode != "" {
			t.Errorf("eBird code %q is on both %s and %s", sp.SpeciesCode, prev, sp.ScientificName)
		}
		seen[sp.SpeciesCode] = sp.ScientificName
	}
}

// TestCountiesJoinedAsRegions asserts the ingest carries a county table into the dataset
// (decisions.countyListingRule): a listed county joins the species' regions with its counts and months, one
// under the bar joins the Below fields, the county's totals join the regions', each county is named
// and placed in its state, and the peak share still names a state.
//
// Verifies: product.regionAndMonthFilter. Criterion: acceptance.countyListingApplied.
func TestCountiesJoinedAsRegions(t *testing.T) {
	dir := t.TempDir()
	table := "#region\tUSA.33.24_1\tKings\n#region\tUSA.33.32_1\tNew York\n" +
		"total\tUSA.33.24_1\t100|100|100|100|100|100|100|100|100|100|100|100\n" +
		"total\tUSA.33.32_1\t50|50|50|50|50|50|50|50|50|50|50|50\n" +
		"Turdus migratorius\tUSA.33.24_1:5|5|5|5|5|5|5|5|5|5|5|5,USA.33.32_1~:1|0|0|0|0|0|0|0|0|0|0|1\n"
	if err := os.WriteFile(filepath.Join(dir, "USA.33_1.tsv"), []byte(table), 0o644); err != nil {
		t.Fatal(err)
	}
	peak := &dataset.Commonest{Region: "USA.33_1", Records: 900, RegionRecords: 60000}
	ds := &dataset.Dataset{
		RegionTotals: map[string]int64{"USA.33_1": 60000},
		Species: []dataset.Species{{ScientificName: "Turdus migratorius",
			Regions: []string{"USA.33_1"}, RegionCounts: []int64{900},
			RegionMonths: [][12]int64{{75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75, 75}}, Commonest: peak}},
	}
	n, err := applyCounties(ds, dir)
	if err != nil {
		t.Fatal(err)
	}
	s := ds.Species[0]
	if n != 2 || len(ds.Counties) != 2 {
		t.Fatalf("joined %d counties, carried %+v", n, ds.Counties)
	}
	for _, c := range ds.Counties {
		if c.State != "USA.33_1" || (c.ID == "USA.33.24_1" && c.Name != "Kings") {
			t.Errorf("carried the county %+v", c)
		}
	}
	if len(s.Regions) != 2 || s.Regions[1] != "USA.33.24_1" || s.RegionCounts[1] != 60 || s.RegionMonths[1][0] != 5 {
		t.Errorf("regions %v counts %v; want Kings added with 60", s.Regions, s.RegionCounts)
	}
	if len(s.BelowRegions) != 1 || s.BelowRegions[0] != "USA.33.32_1" || s.BelowCounts[0] != 2 {
		t.Errorf("below regions %v counts %v; want New York County with 2", s.BelowRegions, s.BelowCounts)
	}
	if ds.RegionTotals["USA.33.24_1"] != 1200 || ds.RegionMonthlyTotals["USA.33.32_1"][0] != 50 {
		t.Errorf("county totals %v, monthly %v", ds.RegionTotals, ds.RegionMonthlyTotals)
	}
	if s.Commonest != peak {
		t.Errorf("the peak share moved to %+v; it names a state, whose counties are inside it", s.Commonest)
	}
}

// TestEveryDrawnBirdHasAHabitatReview holds data/habitat-review.tsv complete against the dataset
// (decisions.habitatReviewedPerSpecies): every species BirdView draws has a row, no row names a species it does not
// draw, each row's AVONET habitat is AVONET's today, and exactly the reviewed rows' additions reach
// the dataset. A new species, or an AVONET update, fails here, naming what to review.
//
// Verifies: product.habitatAndDietFilters. Criterion: acceptance.habitatReviewComplete.
func TestEveryDrawnBirdHasAHabitatReview(t *testing.T) {
	ds := load(t)
	f, err := os.Open(habitatReviewFile)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	rev, err := trait.ParseHabitatReview(f)
	if err != nil {
		t.Fatal(err)
	}
	drawn := map[string]bool{}
	for _, sp := range ds.Species {
		drawn[sp.ScientificName] = true
		r, ok := rev[sp.ScientificName]
		if !ok {
			t.Errorf("%s is drawn and has no row in data/habitat-review.tsv; add one, pending until reviewed", sp.ScientificName)
			continue
		}
		var avonet string
		var ours []string
		for _, a := range sp.Habitats {
			if a.Origin == dataset.Published {
				avonet = a.Value
			} else {
				ours = append(ours, a.Value)
			}
		}
		if r.AVONET != avonet {
			t.Errorf("%s: AVONET's habitat is %q, but the review was of %q; review the row again", sp.ScientificName, avonet, r.AVONET)
		}
		var want []string
		if r.Status == "reviewed" {
			for _, a := range r.Additions {
				if a.Habitat != avonet {
					want = append(want, a.Habitat)
				}
			}
		}
		if strings.Join(ours, ",") != strings.Join(want, ",") {
			t.Errorf("%s (%s): the dataset adds %v, the review %v", sp.ScientificName, r.Status, ours, want)
		}
	}
	for sci := range rev {
		if !drawn[sci] {
			t.Errorf("data/habitat-review.tsv has a row for %s, which BirdView does not draw", sci)
		}
	}
}

// TestNeckReviewAppliesToDrawnTaxa asserts decisions.proportionsReviewedPerTaxon in the build: every
// taxon in data/neck-review.tsv is a genus or family BirdView draws, and its rows reach the
// species, the scaup's neck naming its genus row and the grebe's head its family factor.
//
// Verifies: product.neckModelPerFamily, product.craniumFromMass. Criterion: acceptance.proportionsReviewTable.
func TestNeckReviewAppliesToDrawnTaxa(t *testing.T) {
	ds := load(t)
	drawn := map[string]bool{}
	var scaup, grebe *dataset.Species
	for i := range ds.Species {
		sp := &ds.Species[i]
		drawn[sp.Family] = true
		drawn[strings.Fields(sp.ScientificName)[0]] = true
		switch sp.ScientificName {
		case "Aythya affinis":
			scaup = sp
		case "Podiceps auritus":
			grebe = sp
		}
	}
	for taxon := range neckReview {
		if !drawn[taxon] {
			t.Errorf("data/neck-review.tsv names %s, which is no genus or family BirdView draws", taxon)
		}
	}
	if scaup == nil || grebe == nil {
		t.Fatal("the dataset has no Lesser Scaup or Horned Grebe")
	}
	if m := scaup.Segments.Neck.Prov.Model; !strings.Contains(m, "genus Aythya") {
		t.Errorf("Lesser Scaup's neck model reads %q; want the Aythya row", m)
	}
	if m := grebe.Segments.Head.Prov.Model; !strings.Contains(m, "Podicipedidae") {
		t.Errorf("Horned Grebe's head model reads %q; want the Podicipedidae factor", m)
	}
}
