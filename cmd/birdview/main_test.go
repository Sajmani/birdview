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
	"bytes"
	"compress/gzip"
	"encoding/xml"
	"fmt"
	"html"
	"io"
	"math"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"
	"testing"

	"birdview/internal/dataset"
	"birdview/internal/month"
	"birdview/internal/region"
	"birdview/internal/status"
	"birdview/internal/taxonomy"
	"birdview/internal/trait"
	"birdview/internal/view"
	"birdview/web"
)

// fixtureHabitat and fixtureNiche give each fixture family a plausible
// AVONET value, spread across the vocabulary so the controls have more than
// one option and the "hidden when empty" rule is exercised rather than
// assumed. Ardeidae gets a second, ours-marked habitat so the origin
// distinction appears in the fixture too.
func fixtureHabitat(fam string) dataset.Assignment {
	switch fam {
	case "Ardeidae", "Recurvirostridae":
		return trait.Published("Wetland")
	case "Alcidae":
		return trait.Published("Marine")
	case "Corvidae":
		return trait.Published("Human Modified")
	default:
		return trait.Published("Forest")
	}
}

func fixtureNiche(fam string) dataset.Assignment {
	switch fam {
	case "Ardeidae", "Alcidae", "Recurvirostridae":
		return trait.Published("Aquatic predator")
	case "Trochilidae":
		return trait.Published("Nectarivore")
	case "Corvidae":
		return trait.Published("Omnivore")
	default:
		return trait.Published("Invertivore")
	}
}

func testDataset() *dataset.Dataset {
	mk := func(id, common, sci, fam string, bill, head, neck, body, tail, span, mass, chord float64, k dataset.Kind) dataset.Species {
		m := func(v float64) dataset.Metric {
			return dataset.Metric{Value: v, Prov: dataset.Provenance{
				Kind: dataset.Measured, Source: "AVONET v7, sheet AVONET2_eBird", SourceTaxon: sci}}
		}
		mod := func(v float64, s string) dataset.Metric {
			return dataset.Metric{Value: v, Prov: dataset.Provenance{Kind: dataset.Modeled, Model: s}}
		}
		sp := dataset.Species{
			// The id IS the eBird code in production (product.englishNameChoice), so the fixture
			// mirrors that rather than keeping the scientific-name slug it
			// used to carry. The slug argument survives only as a readable
			// label in the table below.
			ID: code(common),
			Names: map[dataset.Locale]string{
				dataset.LocaleNA:    common,
				dataset.LocaleWorld: worldName(common),
			},
			SpeciesCode: code(common), ScientificName: sci, Family: fam,
			FamilyCommon: fam + " family", FamilyColor: "#86efac",
			TotalLength: mod(bill+head+neck+body+tail, "sum of the five drawn segments"),
			Wingspan:    dataset.Metric{Value: span, Prov: dataset.Provenance{Kind: k, Source: "Fox et al. 2026", SourceTaxon: "Other species"}},
			Mass:        m(mass),
			WingChord:   m(chord),
			Segments: dataset.Segments{
				Bill: m(bill), Head: mod(head, "cranial vault model"), Neck: mod(neck, "neck ratio for "+fam),
				Body: mod(body, "trunk model"), Tail: m(tail),
			},
			// Habitat and diet (product.habitatAndDietFilters). The fixture carries real AVONET
			// values so the trait controls have something to offer: an
			// option no fixture species holds is correctly hidden, and a
			// fixture with no assignments renders an empty control.
			Habitats: []dataset.Assignment{fixtureHabitat(fam)},
			Niches:   []dataset.Assignment{fixtureNiche(fam)},
		}
		return sp
	}
	// Passeriformes deliberately holds TWO families. With one family per order,
	// "select the whole group" and "select the first family" are the same
	// operation, and a bug in the group logic passes unnoticed.
	orders := map[string]string{
		"Trochilidae": "Apodiformes", "Turdidae": "Passeriformes",
		"Corvidae": "Passeriformes", "Ardeidae": "Pelecaniformes",
		"Recurvirostridae": "Charadriiformes", "Alcidae": "Charadriiformes",
	}
	// The stilt adds a fourth order, so the order-group controls are exercised
	// against something other than the two-family Passeriformes case.
	//
	// The final column is the wing chord, and the crow's wingspan is MODELED on
	// purpose: without one modeled span in the fixture, every crossbar renders
	// solid and the split of product.measuredVersusModeled is never exercised. Every chord is shorter
	// than its own half-span, as it is for all 841 real species.
	sp := []dataset.Species{
		mk("annhum", "Anna's Hummingbird", "Calypte anna", "Trochilidae", 18, 12, 6, 40, 26, 120, 4, 49, dataset.Measured),
		mk("amerob", "American Robin", "Turdus migratorius", "Turdidae", 20, 18, 12, 90, 85, 390, 77, 130, dataset.Measured),
		mk("amecro", "American Crow", "Corvus brachyrhynchos", "Corvidae", 50, 30, 22, 170, 180, 1000, 450, 330, dataset.Modeled),
		mk("blasti", "Black-necked Stilt", "Himantopus mexicanus", "Recurvirostridae", 66.6, 42.1, 40.3, 89.5, 73.6, 750, 176.8, 230, dataset.NearMatch),
		mk("greblu", "Great Blue Heron", "Ardea herodias", "Ardeidae", 140, 40, 200, 300, 180, 1825, 2400, 466.4, dataset.NearMatch),
		// The murre is heavier than the crow while being shorter and
		// narrower-winged, and it is the only species here for which the
		// three metrics disagree. Without it the fixture is monotonic in
		// length, wingspan and mass alike, so every sort produces the same
		// order and a Sort by control that ignored its metric entirely would
		// pass its own test. Real values: Uria lomvia is a dense diver.
		mk("thimur", "Thick-billed Murre", "Uria lomvia", "Alcidae", 32, 40, 28, 200, 46, 726, 964, 206, dataset.Measured),
	}
	// Regions, so the filter of product.regionAndMonthFilter has something to filter. Deliberately
	// uneven: the hummingbird is western only, the murre is a seabird that
	// also comes ashore, and the heron is in no region at all -- which is the
	// case the filter must not silently swallow.
	regionsOf := map[string][]string{
		"annhum": {"USA.5_1", "USA.38_1"},            // California, Oregon
		"amerob": {"USA.5_1", "USA.39_1", "CAN.9_1"}, // CA, Pennsylvania, Ontario
		"amecro": {"USA.39_1", "CAN.9_1"},            // Pennsylvania, Ontario
		"blasti": {"USA.5_1"},                        // California
		"thimur": {"USA.2_1", "SEA.PAC"},             // Alaska, and at sea
	}
	// Where each is the biggest share of what its region records (product.shareOfRecords).
	// The heron is left without one on purpose: it is in no region either,
	// and the tile must render the em-dash line rather than a shorter tile
	// that takes its row out of step.
	commonestOf := map[string]*dataset.Commonest{
		"annhum": {Region: "USA.5_1", Records: 4000, RegionRecords: 1000000},
		"amerob": {Region: "CAN.9_1", Records: 73005, RegionRecords: 1817968},
		"amecro": {Region: "USA.39_1", Records: 1200, RegionRecords: 480000},
		"blasti": {Region: "USA.5_1", Records: 90, RegionRecords: 1000000},
		"thimur": {Region: "USA.2_1", Records: 5000, RegionRecords: 6055121},
	}
	// Records per region, parallel to Regions, and the regions' own totals.
	// Uneven on purpose: the robin is dense in Ontario and thin in
	// California, so narrowing the filter to Ontario must raise its share.
	countsOf := map[string][]int64{
		"annhum": {4000, 900},
		"amerob": {9000, 21000, 73005},
		"amecro": {1200, 40000},
		"blasti": {90},
		"thimur": {5000, 300},
	}
	for i := range sp {
		sp[i].Regions = regionsOf[sp[i].ID]
		sp[i].RegionCounts = countsOf[sp[i].ID]
		sp[i].Commonest = commonestOf[sp[i].ID]
	}
	// One assignment of each non-published origin, so the fixture exercises
	// what BirdView adds rather than only what AVONET publishes. Without
	// these the /added page renders empty and its test proves nothing.
	for i := range sp {
		switch sp[i].ID {
		case "greblu": // heron: AVONET's wetland, plus the tidal flats it also works
			sp[i].Habitats = append(sp[i].Habitats,
				trait.BirdView("Coastal", "forages on tidal flats as readily as on fresh water"))
		case "amecro": // crow: level says mixed, niche names one food
			sp[i].Niches = append(sp[i].Niches,
				trait.Derived("Invertivore", "AVONET's trophic level is Omnivore though its niche names seeds"))
		}
	}

	// One mass is a near-match, as 10 of the 841 real ones are. Without it
	// every mass square renders solid and the dashed branch of product.measuredVersusModeled is never
	// drawn -- the same gap the crow's modeled wingspan fills for the crossbar,
	// and it was found by breaking the code and watching the check not fail.
	for i := range sp {
		if sp[i].ID == "blasti" {
			sp[i].Mass.Prov = dataset.Provenance{Kind: dataset.NearMatch,
				Source: "AVONET v7, sheet AVONET2_eBird", SourceTaxon: "Himantopus himantopus"}
		}
	}

	// iNaturalist taxon IDs (product.inaturalistLink). The stilt is deliberately left unmatched so
	// both branches of the detail view are exercised: a species iNaturalist has
	// not confidently matched must show no link at all.
	inat := map[string]int{
		"annhum": 5305, "amerob": 12727,
		"amecro": 8021, "greblu": 4956,
	}
	for i := range sp {
		sp[i].INatTaxonID = inat[sp[i].ID]
	}

	var fams []dataset.Family
	seenFam := map[string]bool{}
	for i := range sp {
		sp[i].Order = orders[sp[i].Family]
		if seenFam[sp[i].Family] {
			continue
		}
		seenFam[sp[i].Family] = true
		fams = append(fams, dataset.Family{
			Name: sp[i].Family, Common: sp[i].FamilyCommon, Order: sp[i].Order,
			Color: "#86efac", TextColor: "#000000",
			OrderCommon: orders[sp[i].Family] + " (group)",
		})
	}
	// Every bird record each region holds, over all species. The
	// denominator of the share a tile shows: dividing by what the same
	// observers saw in the same place is what cancels out how hard they
	// were looking.
	//
	// SEA.PAC carries one again (decisions.honestSeaDenominator). Its value is far larger than the
	// murre's 300 records there, as a real sea's is -- every bird recorded
	// on that water, not a sum over the few species placed in it, which is
	// the error that made a marine share meaningless.
	totals := map[string]int64{
		"USA.5_1": 1000000, "USA.38_1": 200000, "USA.39_1": 480000,
		"CAN.9_1": 1817968, "USA.2_1": 6055121, "SEA.PAC": 1596171,
	}
	return &dataset.Dataset{Families: fams, Species: sp, RegionTotals: totals,
		Undrawn: undrawnFixture()}
}

// undrawnFixture holds one bird of each verdict the derivation can reject a
// species with (product.undrawnSpeciesAnswer).
//
// All three, because the page says genuinely different things for each, and
// the one that matters is `absent`: it is the verdict whose counts were never
// measured, so it is the only case that can catch a page printing zeros as
// though they were a result. A fixture carrying the two measured verdicts
// would render correctly and prove nothing about the third.
//
// The figures are the real ones, so the prose read while developing is the
// prose a reader gets: Passenger Pigeon really does recur across 89 episodes
// in 52 years, and really does stop on 1902-05-18.
func undrawnFixture() []dataset.Undrawn {
	names := func(na string) map[dataset.Locale]string {
		return map[dataset.Locale]string{
			dataset.LocaleNA: na, dataset.LocaleWorld: worldName(na)}
	}
	return []dataset.Undrawn{{
		ID: "paspig", Names: names("Passenger Pigeon"),
		ScientificName: "Ectopistes migratorius", Status: "historical",
		Region: "CAN.9_1", Episodes: 89, Years: 52, Records: 145,
		LastSeen: "1902-05-18",
	}, {
		ID: "sedwar1", Names: names("Sedge Warbler"),
		ScientificName: "Acrocephalus schoenobaenus", Status: "vagrant",
		Region: "USA.2_1", Episodes: 2, Years: 2, Records: 4,
		LastSeen: "2018-10-04",
	}, {
		// No region, no counts, no date -- which is exactly what the ingest
		// emits for `absent`, the derivation never having run a date query
		// for this bird.
		ID: "lillov1", Names: names("Lilian's Lovebird"),
		ScientificName: "Agapornis lilianae", Status: "absent",
	}}
}

// fixtureCounts reports how many species, families and orders testDataset
// holds. The selector checks state a relationship against these rather than a
// literal, which would go quietly stale the next time the fixture grows a
// species -- and growing it is how those checks get sharper, not weaker.
func fixtureCounts() (species, families, orders int) {
	ds := testDataset()
	seen := map[string]bool{}
	for _, f := range ds.Families {
		if !seen[f.Order] {
			seen[f.Order] = true
			orders++
		}
	}
	return len(ds.Species), len(ds.Families), orders
}

// code builds a plausible eBird species code from a common name, so the
// fixture exercises the same rendering path as real data.
//
// Lower case and SIX characters where the name allows, matching the real
// shape: "American Robin" gives "amerob". Real codes run 4 to 8 characters,
// which is why nothing may assume a width.
func code(common string) string {
	var out []byte
	for _, w := range strings.Fields(common) {
		for len(w) < 3 {
			w += "x"
		}
		out = append(out, strings.ToLower(w[:3])...)
		if len(out) >= 6 {
			break
		}
	}
	for len(out) < 6 {
		out = append(out, 'x')
	}
	return string(out[:6])
}

// worldName gives the fixture a world name that is different from its North
// American one AND is not a superstring of it, so "the page does not contain
// the other locale's name" is a check that can actually fail.
//
// The first version appended " (world)", which contains the NA name whole --
// so the world page trivially contained both and the negative assertion was
// vacuous. Reversing the words avoids it: "American Robin" / "Robin American".
func worldName(common string) string {
	w := strings.Fields(common)
	for i, j := 0, len(w)-1; i < j; i, j = i+1, j-1 {
		w[i], w[j] = w[j], w[i]
	}
	return strings.Join(w, " ")
}

func newTestServer(t *testing.T) *server {
	t.Helper()
	s, err := newServer(testDataset())
	if err != nil {
		t.Fatalf("newServer: %v", err)
	}
	return s
}

func get(t *testing.T, s *server, path string) (int, string) {
	t.Helper()
	// Drop the fragment, because a browser does: it is dereferenced on the
	// client and never sent. httptest.NewRequest takes the target literally,
	// so without this a pick link's "#s-annhum" lands inside RawQuery
	// and the last species value becomes "annhum#s-annhum" --
	// which resolves to nothing and is quietly dropped as a dead id. Following
	// a pick link in a test would then exercise something no browser can do.
	if i := strings.IndexByte(path, '#'); i >= 0 {
		path = path[:i]
	}
	w := httptest.NewRecorder()
	s.ServeHTTP(w, httptest.NewRequest(http.MethodGet, path, nil))
	return w.Code, w.Body.String()
}

// Verifies: tech.twoCommandsOneModule, tech.servesOnLocalhost. Criterion: acceptance.bothCommandsBuild.
func TestRoutesServe(t *testing.T) {
	s := newTestServer(t)
	for _, path := range []string{
		"/", "/static/styles.css", "/species/greblu",
		"/?family=Ardeidae", "/?by=mass&scale=global&units=metric&step=30",
	} {
		if code, body := get(t, s, path); code != http.StatusOK {
			t.Errorf("%s: status %d", path, code)
		} else if len(body) == 0 {
			t.Errorf("%s: empty body", path)
		}
	}
	if code, _ := get(t, s, "/species/nonesuch"); code != http.StatusNotFound {
		t.Errorf("unknown species: status %d, want 404", code)
	}
}

// externalSubresource finds something the browser would FETCH from another
// host: any src attribute, or a stylesheet link.
//
// Deliberately not every external href. A hyperlink the reader may click is not
// a subresource and does not affect self-containment -- and CC BY requires
// links to each source and license, so a rule against all external hrefs is in
// direct conflict with acceptance.attributionComplete. The first version of this check made exactly
// that mistake and failed the moment proper attribution was added.
var externalSubresource = regexp.MustCompile(
	`(?:src\s*=\s*"(?:https?:)?//[^"]+|<link[^>]+href\s*=\s*"(?:https?:)?//[^"]+)`)

// TestNoExternalResources asserts tech.selfContainedServer where it actually bites: a page that
// pulls a stylesheet or script from a CDN is not self-contained, however
// self-contained the binary is. The previous design loaded Tailwind from
// cdn.tailwindcss.com and would have failed this.
//
// What it must NOT forbid is an ordinary hyperlink. tech.selfContainedServer is about what the
// page loads, not where it points, and CC BY obliges BirdView to link out to
// every source and license.
//
// Verifies: tech.selfContainedServer. Criterion: acceptance.noServerNetworkRequest.
func TestNoExternalResources(t *testing.T) {
	s := newTestServer(t)
	for _, path := range []string{"/", "/species/greblu"} {
		_, body := get(t, s, path)
		if m := externalSubresource.FindAllString(body, -1); len(m) > 0 {
			t.Errorf("%s fetches external resources: %v", path, m)
		}
	}
}

// Verifies: product.tileContents. Criterion: acceptance.tileCarriesHeaderAndMetrics.
func TestTileShowsHeaderAndMetrics(t *testing.T) {
	_, body := get(t, newTestServer(t), "/")
	for _, want := range []string{
		"American Robin",     // common name
		"(amerob)",           // eBird species code, in parentheses after it
		"Turdus migratorius", // scientific name
		"Turdidae",           // family badge, scientific name
		"Turdidae family",    // family badge, common name alongside it
		"<svg",               // the diagram
		// The three main metrics, named in full rather than as L/W/M (product.tileContents).
		// Spelled out they are also worth asserting: "M " matched almost
		// anything on the page.
		"Length ", "Wingspan ", "Mass ",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("tile is missing %q", want)
		}
	}

	// The common and scientific names share one header line (product.tileContents), so no
	// block-level element may close between them.
	i := strings.Index(body, "American Robin")
	j := strings.Index(body, "Turdus migratorius")
	if i < 0 || j < 0 || j < i {
		t.Fatal("names not found in the expected order")
	}
	if strings.Contains(body[i:j], "</div>") {
		t.Error("scientific name is not on the same header line as the common name")
	}
}

// diagrams splits a page into its individual <svg> blocks.
func diagrams(t *testing.T, body string) []string {
	t.Helper()
	var out []string
	for _, m := range regexp.MustCompile(`(?s)<svg class="dia".*?</svg>`).FindAllString(body, -1) {
		out = append(out, m)
	}
	if len(out) == 0 {
		t.Fatal("page contains no diagrams")
	}
	return out
}

// TestDiagramStaysInsideItsBox checks the whole scale chain end to end on the
// rendered page: Scales picks a reference, Layout turns it into pixels, and the
// template centers the result. An error anywhere in that chain shows up here as
// a drawing that runs off the top or bottom of its own viewBox.
//
// It runs under per-section scale deliberately. That is where the largest bird
// in a section sets its own reference and fills the box exactly, leaving no
// slack for an arithmetic error to hide in; under global scale almost every
// bird is drawn small and a bad offset merely looks off-center -- visible to a
// human, invisible to an assertion.
//
// It also runs at a wide step, and refuses to run without one. At the default
// step every bird in this fixture lands in a section of its own, so a section's
// minimum and maximum are the same bird and Scales cannot be caught returning
// the wrong one -- the first version of this test passed happily with Scales
// inverted to take the minimum. Grouping is the whole point, so the check that
// grouping happened comes first.
//
// Verifies: product.tileContents, product.scaleToggle. Criteria: acceptance.tileCarriesHeaderAndMetrics, acceptance.sharedScale.
func TestDiagramStaysInsideItsBox(t *testing.T) {
	const step = 100

	var all []*dataset.Species
	ds := testDataset()
	for i := range ds.Species {
		all = append(all, &ds.Species[i])
	}
	var grouped int
	for _, sec := range view.Sections(all, view.ByLength, step) {
		if len(sec.Species) > 1 {
			grouped++
		}
	}
	if grouped == 0 {
		t.Fatalf("at step %d%% no section holds more than one species, so this test cannot "+
			"see a wrong scale reference -- widen the fixture's size range", step)
	}

	_, body := get(t, newTestServer(t), fmt.Sprintf("/?scale=section&step=%d", step))

	coord := regexp.MustCompile(`\b(?:y1|y2|cy)="(-?[0-9.]+)"`)
	for _, d := range diagrams(t, body) {
		vb := mustMatch(t, d, `viewBox="0 0 [0-9.]+ ([0-9.]+)"`, "the diagram viewBox")
		boxH, err := strconv.ParseFloat(vb, 64)
		if err != nil {
			t.Fatalf("unreadable viewBox height %q", vb)
		}
		dy, err := strconv.ParseFloat(mustMatch(t, d, `translate\(0,(-?[0-9.]+)\)`, "the centering transform"), 64)
		if err != nil {
			t.Fatalf("unreadable transform offset: %v", err)
		}
		var lowest float64
		for _, m := range coord.FindAllStringSubmatch(d, -1) {
			v, err := strconv.ParseFloat(m[1], 64)
			if err != nil {
				t.Fatalf("unreadable coordinate %q", m[1])
			}
			if v > lowest {
				lowest = v
			}
		}
		// The template rounds to one decimal, so allow a rounding epsilon.
		name := mustMatch(t, d, `aria-label="scaled outline of ([^"]+)"`, "the diagram label")
		if dy+lowest > boxH+0.2 {
			t.Errorf("%s: drawing reaches %.1f in a %.1f box -- it is clipped at the bottom",
				name, dy+lowest, boxH)
		}
		if dy < -0.2 {
			t.Errorf("%s: drawing starts at %.1f, above the top of the box", name, dy)
		}
	}
}

// TestDiagramIsVerticallyCentered asserts product.tileContents: a bird shorter than the
// drawing area sits in the middle of it, not on the floor.
//
// This matters most under global scale, where a small bird occupies a fraction
// of the tile's height and bottom-alignment reads as a rendering fault rather
// than as the size difference it is.
//
// Verifies: product.tileContents. Criterion: acceptance.tileCarriesHeaderAndMetrics.
func TestDiagramIsVerticallyCentered(t *testing.T) {
	_, body := get(t, newTestServer(t), "/?scale=global")

	if strings.Contains(body, "xMidYMax") {
		t.Error("diagram is bottom-anchored (xMidYMax); it should be centered")
	}
	m := regexp.MustCompile(`translate\(0,([0-9.]+)\)`).FindAllStringSubmatch(body, -1)
	if len(m) == 0 {
		t.Fatal("no diagram transforms found")
	}
	// Under one global scale the smallest bird has the most slack, so at least
	// one diagram must be inset from the top.
	var inset bool
	for _, g := range m {
		if v, _ := strconv.ParseFloat(g[1], 64); v > 0.05 {
			inset = true
		}
	}
	if !inset {
		t.Error("every diagram sits at offset 0; under global scale smaller birds should be inset")
	}
}

// TestDetailShowsEverySegmentWithProvenance asserts product.tileOpensDetail and product.metricProvenanceExplained: the detail
// view lists all five segments as well as the three main metrics, each with its
// provenance explained.
//
// Verifies: product.tileOpensDetail, product.metricProvenanceExplained. Criterion: acceptance.detailShowsProvenance.
func TestDetailShowsEverySegmentWithProvenance(t *testing.T) {
	_, body := get(t, newTestServer(t), "/species/greblu")
	for _, label := range []string{"Total length", "Wingspan", "Mass",
		"Bill", "Head", "Neck", "Body", "Tail"} {
		if !strings.Contains(body, ">"+label+"<") {
			t.Errorf("detail view omits %q", label)
		}
	}
	for _, want := range []string{
		"neck ratio for Ardeidae",   // the family's model is named (product.neckModelPerFamily)
		"measured on Other species", // near-match names the borrowed taxon
		"drawing convention",        // the neck disclosure (product.metricProvenanceExplained)
	} {
		if !strings.Contains(body, want) {
			t.Errorf("detail view omits %q", want)
		}
	}
	// It is the same tile, enlarged: the header must be there too.
	for _, want := range []string{"Great Blue Heron", "Ardea herodias", "<svg"} {
		if !strings.Contains(body, want) {
			t.Errorf("detail view is not the tile enlarged; missing %q", want)
		}
	}
}

// Verifies: product.familyFilter. Criterion: acceptance.familyFilterSelectsAndClears.
func TestFamilyFilter(t *testing.T) {
	s := newTestServer(t)
	_, all := get(t, s, "/")
	if !strings.Contains(all, "American Robin") || !strings.Contains(all, "Great Blue Heron") {
		t.Fatal("unfiltered view should show every species")
	}
	_, one := get(t, s, "/?family=Ardeidae")
	if strings.Contains(one, "American Robin") {
		t.Error("filtering to Ardeidae still shows a Turdidae species")
	}
	if !strings.Contains(one, "Great Blue Heron") {
		t.Error("filtering to Ardeidae dropped the heron")
	}

	// The selector labels each family with both names (product.familyFilter): the scientific
	// name alone makes the user translate before they can choose.
	sel := box(t, all, "families")
	for _, want := range []string{"Ardeidae", "Ardeidae family", "Turdidae", "Trochilidae"} {
		if !strings.Contains(sel, want) {
			t.Errorf("family selector is missing %q", want)
		}
	}
	// And carries NO per-control count. One count sits above the list instead,
	// because a per-chip count is only true while family is the only filter:
	// add a region and "Parulidae 51" is false in Hawaii (product.familyFilter).
	// Look at the chips, not the whole box: the help text above them uses <b>
	// legitimately. A chip count rendered as "<em>common</em><b>51</b>".
	chips := sel[strings.Index(sel, `class="ordergroups"`):]
	if strings.Contains(chips, "</em><b>") {
		t.Error("the family selector still carries per-entry counts; they belong above the list")
	}
	if !strings.Contains(all, `<p class="count">`) {
		t.Error("no count above the list, so the counts were removed and not replaced")
	}

	// Grouped by order, in checklist sequence rather than alphabetically
	// (product.familyFilter). Alphabetical would put Ardeidae first; the fixture's checklist
	// order puts Trochilidae first.
	for _, order := range []string{"Apodiformes", "Passeriformes", "Pelecaniformes"} {
		if !strings.Contains(sel, order) {
			t.Errorf("selector does not show the order %q as a group", order)
		}
	}
	if a, b := strings.Index(sel, "Apodiformes"), strings.Index(sel, "Pelecaniformes"); a > b {
		t.Error("selector is not in checklist sequence; groups appear reordered")
	}
	// Each family chip sits inside its order's group.
	gi := strings.Index(sel, "Passeriformes")
	next := strings.Index(sel[gi:], `class="ogroup"`)
	if next < 0 {
		next = len(sel) - gi
	}
	if !strings.Contains(sel[gi:gi+next], "Turdidae") {
		t.Error("Turdidae is not inside the Passeriformes group")
	}
}

// tileArea returns the part of the index below the help and controls panels.
//
// The help now contains a real labeled diagram of a real species (product.pageExplainsItself), so a
// whole-page search for a bird's name or for a diagram label finds the legend
// as well as the tiles. Every assertion about what is DISPLAYED has to be
// scoped here, or it silently answers a different question.
func tileArea(t *testing.T, body string) string {
	t.Helper()
	i := strings.LastIndex(body, "</details>")
	if i < 0 {
		t.Fatal("no collapsible panels on the page; cannot locate the tile area")
	}
	return body[i:]
}

// footer returns the contents of the first tile's metric footer.
func footer(t *testing.T, body string) string {
	t.Helper()
	i := strings.Index(body, `<div class="ft">`)
	if i < 0 {
		t.Fatal("no tile footer in the page")
	}
	rest := body[i:]
	return rest[:strings.Index(rest, "</div>")]
}

// TestUnitToggle asserts product.imperialByDefault where the user actually reads the numbers.
//
// The earlier version of this check only looked for a centimeter value
// somewhere on the page, and passed while every tile was hardcoded to inches:
// the section headers alone satisfied it. Asserting on the tile footer is the
// difference between checking the requirement and checking that some part of
// the page happens to mention the right unit.
//
// Verifies: product.imperialByDefault. Criterion: acceptance.unitToggleSwitches.
func TestUnitToggle(t *testing.T) {
	s := newTestServer(t)

	// Counting units, rather than looking for a forbidden substring, is what
	// makes this bite. A footer hardcoded to render inches as `3.9"` contains
	// no " in" token at all, so an exclusion list silently misses it; requiring
	// that all three metrics carry the selected system's units does not.
	check := func(path string, wantLen, wantMass []string) {
		t.Helper()
		_, body := get(t, s, path)
		f := footer(t, body)
		if n := countAny(f, wantLen); n != 2 {
			t.Errorf("%s: footer %q has %d of the two length values in %v", path, f, n, wantLen)
		}
		if n := countAny(f, wantMass); n != 1 {
			t.Errorf("%s: footer %q has %d of the one mass value in %v", path, f, n, wantMass)
		}
	}
	check("/", []string{" in"}, []string{" oz", " lb"})
	check("/?units=metric", []string{" cm"}, []string{" g", " kg"})

	_, a := get(t, s, "/")
	_, b := get(t, s, "/?units=metric")
	if footer(t, a) == footer(t, b) {
		t.Error("tile footer is identical in both unit systems; the toggle does nothing")
	}
}

// countAny counts how many of the given tokens appear in s, totaled.
func countAny(s string, tokens []string) int {
	n := 0
	for _, tok := range tokens {
		n += strings.Count(s, tok)
	}
	return n
}

// TestMassUnitsBySystem pins the mass conversions, which are the ones an
// earlier version got backwards by rendering kilograms under imperial.
//
// Expected values are computed from the definition (1 oz = 28.349523125 g),
// not read off a run.
//
// Verifies: product.imperialByDefault. Criterion: acceptance.unitToggleSwitches.
func TestMassUnitsBySystem(t *testing.T) {
	cases := []struct {
		g                float64
		imperial, metric string
	}{
		{4, "0.14 oz", "4 g"},       // hummingbird
		{77, "2.7 oz", "77 g"},      // robin
		{2400, "5.3 lb", "2.40 kg"}, // heron
	}
	for _, c := range cases {
		// One bird is its own scope, which is the old per-value behaviour.
		if got := unitsFor(false, c.g).format(c.g, view.ByMass); got != c.imperial {
			t.Errorf("%.0f g imperial: got %q, want %q", c.g, got, c.imperial)
		}
		if got := unitsFor(true, c.g).format(c.g, view.ByMass); got != c.metric {
			t.Errorf("%.0f g metric: got %q, want %q", c.g, got, c.metric)
		}
	}
}

// TestDefaultScaleIsPerSection asserts product.scaleToggle's default, which decisions.perSectionDefault moved.
//
// Global still makes the cross-family argument better than anything else, and
// this test asserted it for good reasons. What changed is that the diagram
// acquired a vocabulary -- dashed for modeled, chord ticks, labels -- and at
// global scale a median bird is a couple of pixels tall, so the landing page
// was hiding the drawing's own language from every first-time reader.
//
// Verifies: product.scaleToggle. Criterion: acceptance.sharedScale.
func TestDefaultScaleIsPerSection(t *testing.T) {
	if got := parseOptions(httptest.NewRequest(http.MethodGet, "/", nil)).Scale; got != view.PerSection {
		t.Errorf("default scale is %q, want %q", got, view.PerSection)
	}
	// And the toggle must still reach the other mode: global is one click away,
	// which is the premise the default change rests on.
	r := httptest.NewRequest(http.MethodGet, "/?scale=global", nil)
	if got := parseOptions(r).Scale; got != view.Global {
		t.Errorf("?scale=global gave %q, want %q", got, view.Global)
	}
}

// TestDetailDrawsFullScale asserts product.tileOpensDetail: the detail view fills its box with
// the one bird it is about, whatever scale the index is using.
//
// Under the global default a kinglet is a few pixels tall on the index, which
// is the honest comparison. Inheriting that here would leave nothing to look at
// on the page devoted to it.
//
// Verifies: product.tileOpensDetail. Criterion: acceptance.detailDrawsFullScale.
func TestDetailDrawsFullScale(t *testing.T) {
	s := newTestServer(t)
	// The hummingbird is the smallest fixture bird, so it is the one that would
	// be reduced to nothing if the index's scale leaked through.
	for _, path := range []string{
		"/species/annhum",
		"/species/annhum?scale=global",
		"/species/annhum?scale=section",
	} {
		_, body := get(t, s, path)
		m := regexp.MustCompile(`translate\(0,([0-9.]+)\)`).FindStringSubmatch(body)
		if m == nil {
			t.Fatalf("%s: no diagram transform", path)
		}
		// Filling the box vertically means no vertical slack, so no inset.
		if v, _ := strconv.ParseFloat(m[1], 64); v > 0.5 {
			t.Errorf("%s: diagram inset by %.1f px; it should fill the view", path, v)
		}
	}
}

// TestHelpCoversEveryControl asserts product.pageExplainsItself: the page explains its own controls
// and its diagram, rather than assuming the reader can infer them.
//
// The list is deliberately tied to the controls that exist. If a control is
// added and this check is not updated, the new control is undocumented and
// this test is the reminder.
//
// Verifies: product.pageExplainsItself. Criterion: acceptance.pageExplainsControls.
func TestHelpCoversEveryControl(t *testing.T) {
	_, body := get(t, newTestServer(t), "/")
	i := strings.Index(body, `<details class="help"`)
	if i < 0 {
		t.Fatal("no help block on the page")
	}
	help := body[i:]
	help = help[:strings.Index(help, "</details>")]

	// The control explanations now sit WITH the controls rather than in the
	// Presentation panel, so they are checked in their own scope (decisions.perSectionDefault).
	// Scope to the EXPLANATION block, not the whole controls panel. Every
	// control name also appears as its own button label a few lines below, so a
	// panel-wide search is satisfied by the control existing rather than by it
	// being explained -- which is how the first version of this passed with the
	// explanations deleted outright.
	// Each panel documents its own controls (product.pageExplainsItself), so this searches all of
	// them rather than the one that used to hold everything.
	if !strings.Contains(body, `<div class="ctrlhelp">`) {
		t.Fatal("no panel carries an explanation block")
	}
	controls := body
	for _, control := range []string{"Order by", "Scale", "Section step", "Units", "Names",
		"Family", "Region", "Compare birds"} {
		if !strings.Contains(controls, "<dt>"+control+"</dt>") {
			t.Errorf("no panel explains the %q control", control)
		}
	}
	// And both scale modes, which is the one control a reader cannot discover
	// by clicking: it changes what the app is saying, not what it shows.
	for _, mode := range []string{"Global", "Per section"} {
		if !strings.Contains(controls, mode) {
			t.Errorf("the controls panel does not explain %q", mode)
		}
	}
	// Every part of the diagram, since none of it is self-evident. These are now
	// labels on a rendered example rather than swatches in a list, so each is
	// matched as the start of its label text.
	// Matched against the legend's own LABELS, not against any text in the
	// panel. The cube carries a <title> tooltip reading "Mass as water: ...",
	// which satisfied a raw ">Mass " search even with the label blanked --
	// so the check passed against a legend that labelled nothing.
	var legendLabels []string
	for _, m := range regexp.MustCompile(`<text class="dlabel[^"]*"[^>]*>([^<]*)</text>`).
		FindAllStringSubmatch(help, -1) {
		legendLabels = append(legendLabels, strings.TrimSpace(html.UnescapeString(m[1])))
	}
	// Mass included: it is labelled on the legend like every other metric
	// (product.labeledDetailDiagram), and it is the one that went missing before.
	for _, part := range []string{"Bill", "Head", "Neck", "Body", "Tail", "Wingspan", "Wing chord", "Mass"} {
		if !slices.ContainsFunc(legendLabels, func(l string) bool {
			return strings.HasPrefix(l, part+" ")
		}) {
			t.Errorf("diagram legend omits %q; its labels are %v", part, legendLabels)
		}
	}
	// Two sections, in a stated order (product.pageExplainsItself): the thing on screen, then what
	// its strokes mean. How to change it moved down to the controls.
	order := []string{"The stick figure", "Where the numbers come from"}
	at := make([]int, len(order))
	for i, h := range order {
		if at[i] = strings.Index(help, ">"+h+"<"); at[i] < 0 {
			t.Errorf("help has no %q section", h)
		}
	}
	for i := 1; i < len(order); i++ {
		if at[i-1] >= 0 && at[i] >= 0 && at[i] < at[i-1] {
			t.Errorf("help section %q comes before %q; the order is %v",
				order[i], order[i-1], order)
		}
	}

	// The mass cube: a shape with no label of its own, so the panel that
	// shows it has to say what it is. Checked in the Presentation panel and
	// not merely in the diagram key, since this is where a reader is taught
	// to read the drawing (product.pageExplainsItself).
	// Whitespace-normalised: these are phrases, and the template wraps its
	// prose, so a reflow that put "cube" and "root" on separate lines would
	// fail a raw match and look like the explanation had been deleted.
	flat := strings.Join(strings.Fields(help), " ")
	for _, s := range []string{"mass drawn as water", "cube root"} {
		if !strings.Contains(flat, s) {
			t.Errorf("the Presentation panel does not explain the mass cube: no %q", s)
		}
	}

	// What the estimate marker means: the other thing a reader cannot infer.
	for _, concept := range []string{"estimated"} {
		if !strings.Contains(help, concept) {
			t.Errorf("Presentation panel does not explain %q", concept)
		}
	}
	if !strings.Contains(help, "open") {
		t.Error("help should start expanded")
	}
}

// TestChipTextInheritsItsBackgroundsTextColor guards a bug that shipped: chips
// carry a generated background with a matching black-or-white text color, and
// the stylesheet set the family common name to a fixed grey. Grey is unreadable
// on both halves of the palette, so the common name vanished on roughly a
// dozen chips.
//
// The rule is structural: nothing inside a colored chip or badge may set its
// own color. Checking the stylesheet is crude but it catches precisely the
// regression, which a rendering test cannot see.
//
// Verifies: product.generatedBadgeColors, tech.badgeContrast. Criterion: acceptance.badgeTextInheritsColor.
func TestChipTextInheritsItsBackgroundsTextColor(t *testing.T) {
	css, err := web.Static.ReadFile("static/styles.css")
	if err != nil {
		t.Fatal(err)
	}
	for _, rule := range strings.Split(string(css), "}") {
		open := strings.Index(rule, "{")
		if open < 0 {
			continue
		}
		sel, body := rule[:open], rule[open+1:]
		// Only descendants matter. A default color on the bare .chip is fine --
		// the "All" chip has no generated background and needs one. The bug was
		// in the parts nested inside a chip whose background IS generated.
		if !descendantOfColored(sel) {
			continue
		}
		for _, decl := range strings.Split(body, ";") {
			decl = strings.TrimSpace(decl)
			if !strings.HasPrefix(decl, "color:") {
				continue
			}
			v := strings.TrimSpace(strings.TrimPrefix(decl, "color:"))
			if v != "inherit" && v != "currentColor" {
				t.Errorf("%s sets %q; text on a generated background must inherit "+
					"that background's own text color", strings.TrimSpace(sel), decl)
			}
		}
		// Dimming has the same effect as the wrong color: it undercuts the
		// contrast ratio the background was picked for.
		if strings.Contains(body, "opacity:") && strings.Contains(sel, "em") {
			t.Errorf("%s dims its text; hierarchy must come from weight, not opacity, "+
				"or the 7:1 guarantee is lost", strings.TrimSpace(sel))
		}
	}
}

// descendantOfColored reports whether a selector targets something nested
// inside an element whose background is a generated color.
func descendantOfColored(selector string) bool {
	for _, sel := range strings.Split(selector, ",") {
		parts := strings.Fields(strings.TrimSpace(sel))
		for i, p := range parts {
			if i == len(parts)-1 {
				break // the subject itself, not a descendant of it
			}
			if strings.Contains(p, ".chip") || strings.Contains(p, ".badge") {
				return true
			}
		}
	}
	return false
}

// TestOrderToggles asserts product.familyFilter's order control: each order is a real toggle
// over its whole group, labeled with its common name.
//
// Since a fresh page selects everything, the first click on an order REMOVES
// it. This replaces a check that asserted the opposite -- that order labels
// were inert text. When that requirement inverted, the check kept passing,
// because it searched for `<a class="ohdr` and the markup had become
// `<a class="btn ohdr`. A test that survives the reversal of its own
// requirement is worse than none.
//
// Verifies: product.familyFilter. Criterion: acceptance.familyFilterSelectsAndClears.
func TestOrderToggles(t *testing.T) {
	s := newTestServer(t)
	_, body := get(t, s, "/")

	// Each order is a link carrying its common name, not inert text, and starts
	// selected.
	for _, order := range []string{"Apodiformes", "Passeriformes", "Pelecaniformes"} {
		if !strings.Contains(body, order+`<em>`+order+" (group)") {
			t.Errorf("%s toggle does not show its common name", order)
		}
		if !regexp.MustCompile(`<a class="btn ohdr on"[^>]*>` + order).MatchString(body) {
			t.Errorf("%s does not start selected", order)
		}
	}

	// Clicking an order drops EVERY family in it. Passeriformes holds two, so a
	// control that touches only the first fails here.
	href := html.UnescapeString(mustMatch(t, body,
		`<a class="btn ohdr on"[^>]*href="([^"]+)"[^>]*>Passeriformes`, "the Passeriformes toggle"))
	_, dropped := get(t, s, href)
	for _, gone := range []string{"American Robin", "American Crow"} {
		if strings.Contains(tileArea(t, dropped), gone) {
			t.Errorf("dropping Passeriformes left %s visible", gone)
		}
	}
	if !strings.Contains(tileArea(t, dropped), "Great Blue Heron") {
		t.Error("dropping Passeriformes removed another order's species")
	}

	// And the same control puts them back.
	back := html.UnescapeString(mustMatch(t, dropped,
		`<a class="btn ohdr"[^>]*href="([^"]+)"[^>]*>Passeriformes`,
		"the Passeriformes toggle once cleared"))
	_, restored := get(t, s, back)
	for _, want := range []string{"American Robin", "American Crow", "Great Blue Heron"} {
		if !strings.Contains(restored, want) {
			t.Errorf("restoring Passeriformes did not bring back %s", want)
		}
	}
}

// mustMatch returns the first capture group, failing the test rather than
// panicking when the pattern does not match -- a break that changes the markup
// should report what it could not find, not an index-out-of-range.
func mustMatch(t *testing.T, body, pattern, what string) string {
	t.Helper()
	m := regexp.MustCompile(pattern).FindStringSubmatch(body)
	if m == nil {
		t.Fatalf("could not find %s (pattern %s)", what, pattern)
	}
	return m[1]
}

// TestAllToggle asserts the All control round-trips: from the default, where
// everything is selected, it clears the lot; from cleared, it selects them all
// again.
//
// The species count cannot distinguish the endpoints on its own -- "everything
// selected" and a restored selection both show every bird -- so the check reads
// the control states as well.
//
// Verifies: product.familyFilter. Criterion: acceptance.familyFilterSelectsAndClears.
func TestAllToggle(t *testing.T) {
	s := newTestServer(t)

	_, fresh := get(t, s, "/")
	// The Regions box carries an All control too and comes first on the page,
	// so this must say which selector it means (product.regionAndMonthFilter).
	clear := html.UnescapeString(mustMatch(t, grouping(t, box(t, fresh, "families"), "Taxonomic order"),
		`<a class="btn on"[^>]*href="([^"]+)"[^>]*>All`, "the family All control on a fresh page"))

	_, empty := get(t, s, clear)
	taxoEmpty := grouping(t, box(t, empty, "families"), "Taxonomic order")
	if strings.Count(taxoEmpty, `class="btn ohdr on"`) != 0 {
		t.Error("orders still show as selected after clearing")
	}
	if strings.Count(taxoEmpty, `class="chip on"`) != 0 {
		t.Error("families still show as selected after clearing")
	}

	// Back again: All selects every family, and every order lights up.
	restore := html.UnescapeString(mustMatch(t, box(t, empty, "families"),
		`<a class="btn"[^>]*href="([^"]+)"[^>]*>All`, "the family All control on an emptied page"))
	_, all := get(t, s, restore)
	_, _, wantOrders := fixtureCounts()
	taxo := grouping(t, box(t, all, "families"), "Taxonomic order")
	if n := strings.Count(taxo, `class="btn ohdr on"`); n != wantOrders {
		t.Errorf("%d of %d orders selected after re-selecting All", n, wantOrders)
	}
	if !strings.Contains(all, "Great Blue Heron") {
		t.Error("re-selecting All did not restore the species")
	}
}

// TestDefaultSelectsEverything asserts that a fresh page has everything on:
// All, every order and every family (product.familyFilter, product.firstLoadShowsAll).
//
// Verifies: product.familyFilter, product.firstLoadShowsAll. Criterion: acceptance.familyFilterSelectsAndClears.
func TestDefaultSelectsEverything(t *testing.T) {
	s := newTestServer(t)
	_, body := get(t, s, "/")
	fams := box(t, body, "families")
	if !regexp.MustCompile(`<a class="btn on"[^>]*>All`).MatchString(fams) {
		t.Error("All is not selected on a fresh page")
	}
	_, wantFams, wantOrders := fixtureCounts()
	taxo := grouping(t, fams, "Taxonomic order")
	if n := strings.Count(taxo, `class="btn ohdr on"`); n != wantOrders {
		t.Errorf("%d of %d orders selected on a fresh page", n, wantOrders)
	}
	if n := strings.Count(taxo, `class="chip on"`); n != wantFams {
		t.Errorf("%d of %d families selected on a fresh page", n, wantFams)
	}
	// Everywhere is selected on arrival too, and by the same rule (product.regionAndMonthFilter).
	if !regexp.MustCompile(`<a class="btn on"[^>]*>All`).MatchString(box(t, body, "regions")) {
		t.Error("All regions is not selected on a fresh page")
	}
	if !strings.Contains(body, "Great Blue Heron") {
		t.Error("fresh page should show every species")
	}
}

// TestClearingEverythingShowsGuidance asserts the empty state: with nothing
// selected the page shows no birds and says what to do about it, rather than
// reporting that nothing matched -- the user turned them off, they did not
// fail to match.
//
// Verifies: product.familyFilter. Criterion: acceptance.emptyStateGuides.
func TestClearingEverythingShowsGuidance(t *testing.T) {
	s := newTestServer(t)
	_, body := get(t, s, "/")
	// Scoped to the family block. The Species panel now opens with the
	// habitat and diet controls, each with its own All, so an unscoped
	// first-match grabs the wrong one and clears nothing (product.habitatAndDietFilters).
	clear := html.UnescapeString(mustMatch(t, grouping(t, box(t, body, "families"), "Taxonomic order"),
		`<a class="btn on"[^>]*href="([^"]+)"[^>]*>All`, "the family All control on a fresh page"))

	_, empty := get(t, s, clear)
	for _, gone := range []string{"Great Blue Heron", "American Robin", "American Crow"} {
		if strings.Contains(tileArea(t, empty), gone) {
			t.Errorf("%s is still displayed after clearing every selection", gone)
		}
	}
	if !strings.Contains(empty, "Select birds to display in") {
		t.Error("empty state does not tell the user how to get birds back")
	}
	if strings.Contains(empty, "No species match") {
		t.Error("empty state blames a failed match; the user cleared the selection")
	}
	nSpecies, _, _ := fixtureCounts()
	if want := fmt.Sprintf("0 of %d species", nSpecies); !strings.Contains(empty, want) {
		t.Errorf("count line should report nothing shown as %q", want)
	}
	// And All must offer the way back. Scoped to the family block: the
	// habitat and diet controls are untouched by clearing families, so
	// their All buttons are legitimately still on (product.habitatAndDietFilters).
	if regexp.MustCompile(`<a class="btn on"[^>]*>All`).
		MatchString(grouping(t, box(t, empty, "families"), "Taxonomic order")) {
		t.Error("All is still marked selected with nothing selected")
	}
}

// TestScrollRestorationIsInlineAndSelfContained checks the one script BirdView
// ships is embedded rather than fetched, and that it is scoped to the index.
//
// It cannot check that scrolling actually works -- that needs a browser this
// project does not have -- so acceptance.scrollScriptScoped is honest about covering presence and
// self-containment only.
//
// Verifies: tech.selfContainedServer, product.togglesKeepPlace. Criterion: acceptance.scrollScriptScoped.
func TestScrollRestorationIsInlineAndSelfContained(t *testing.T) {
	s := newTestServer(t)
	_, body := get(t, s, "/")

	if !strings.Contains(body, "birdview.scroll") {
		t.Error("index carries no scroll-restoration script")
	}
	// Inline only: an external script would break tech.selfContainedServer the same way a CDN
	// stylesheet would, and acceptance.noServerNetworkRequest only looks at src/href attributes.
	for _, tag := range regexp.MustCompile(`<script[^>]*>`).FindAllString(body, -1) {
		if strings.Contains(tag, "src=") {
			t.Errorf("script is fetched rather than embedded: %s", tag)
		}
	}
	// Scoped to the index, so a detail page does not consume the saved position.
	if !strings.Contains(body, `window.location.pathname !== "/"`) {
		t.Error("script is not scoped to the index; a detail page would eat the saved scroll")
	}
	// The detail view ships no script at all.
	_, detail := get(t, s, "/species/greblu")
	if strings.Contains(detail, "<script") {
		t.Error("detail view carries a script; only the index needs one")
	}
	// And the controls remain addressable for anyone who prefers an anchor.
	if !strings.Contains(body, `id="controls"`) {
		t.Error("controls have no id; #controls should work as a fallback target")
	}
}

// TestCollapsiblePanelsPersist checks the panels can hold their collapsed state
// across a navigation.
//
// The markup says `open`, so without this every reload reopened a panel the
// reader had deliberately collapsed, pushing the page down -- the same class of
// defect as losing the scroll position, and visible in the same moment.
//
// Both panels need an id for the state to be keyed on, and the restore must run
// before the scroll restore, or collapsing a panel shifts everything under it
// after the position has already been set.
//
// Verifies: product.pageExplainsItself, product.collapsibleControls, product.togglesKeepPlace. Criterion: acceptance.scrollScriptScoped.
func TestCollapsiblePanelsPersist(t *testing.T) {
	_, body := get(t, newTestServer(t), "/")

	for _, id := range []string{"presentation", "controls"} {
		if !strings.Contains(body, `id="`+id+`"`) {
			t.Errorf("panel %q has no id; its collapsed state cannot be remembered", id)
		}
	}
	if !strings.Contains(body, "birdview.open.") {
		t.Error("no script remembers the panels' collapsed state")
	}
	// Ordering: panels before scroll.
	if p, s := strings.Index(body, "birdview.open."), strings.Index(body, "birdview.scroll"); p > s {
		t.Error("scroll is restored before the panels; collapsing one would then shift the page")
	}
	// Every panel defaults to CLOSED (product.collapsibleControls): four open panels put the birds
	// below the fold before the reader has chosen anything.
	for _, id := range []string{"presentation", "controls", "regions", "traits", "families"} {
		open := regexp.MustCompile(`<details[^>]*id="` + id + `"[^>]*\bopen\b`)
		if open.MatchString(body) {
			t.Errorf("panel %q defaults to open; product.collapsibleControls says closed", id)
		}
	}
	// A closed panel is a bare word unless it says what it holds, so each
	// summary carries a one-line description.
	for _, id := range []string{"presentation", "controls", "regions", "traits", "families"} {
		sum := mustMatch(t, body, `(?s)<details[^>]*id="`+id+`">\s*<summary>(.*?)</summary>`,
			"the "+id+" summary")
		if !strings.Contains(sum, `class="sumdesc"`) {
			t.Errorf("panel %q has no description in its summary: %q", id, sum)
		}
	}
}

// TestControlOrderAndOptions pins two presentation choices that are easy to
// undo by accident: the default scale is listed first, and the section step
// offers its full range.
//
// Verifies: product.sectionThresholdDefault, product.scaleToggle. Criterion: acceptance.controlsMatchServer.
func TestControlOrderAndOptions(t *testing.T) {
	_, page := get(t, newTestServer(t), "/")

	// Scope to the controls panel. The Presentation panel above it also names both
	// scale modes, in the same order -- searching the whole page matched the
	// prose and passed even with the buttons reversed.
	ci := strings.Index(page, `<details class="controls"`)
	if ci < 0 {
		t.Fatal("controls panel not found")
	}
	body := page[ci:]

	// Per section is the default (product.scaleToggle, decisions.perSectionDefault), so it leads the pair.
	g := strings.Index(body, `"scale" "global"`)
	if g < 0 {
		g = strings.Index(body, `scale=global`)
	}
	p := strings.Index(body, `scale=section`)
	if g < 0 || p < 0 {
		t.Fatalf("scale controls not found in the panel (global=%d section=%d)", g, p)
	}
	if p > g {
		t.Error("Global is listed before Per section; the default leads the pair")
	}

	// Every offered step must be one the server actually accepts, or the
	// control silently falls back to the default.
	for _, step := range []string{"5", "10", "15", "20", "30"} {
		if !strings.Contains(body, `>`+step+`%<`) {
			t.Errorf("section step %s%% is not offered", step)
		}
		r := httptest.NewRequest(http.MethodGet, "/?step="+step, nil)
		if got := parseOptions(r).Step; got != mustFloat(t, step) {
			t.Errorf("step=%s parsed as %v; the control offers a value the server rejects", step, got)
		}
	}
}

func mustFloat(t *testing.T, s string) float64 {
	t.Helper()
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		t.Fatal(err)
	}
	return v
}

// TestAttributionIsComplete checks every CC BY source is credited the way that
// license asks: creator, title, a link to the source, and a link to the license.
//
// The obligation is real and becomes enforceable the moment the app is public.
// An earlier footer named the datasets but linked neither the deposits nor the
// license, which is credit rather than attribution.
//
// Verifies: tech.localNowPublicLater, goals.goodDataCitizen. Criterion: acceptance.attributionComplete.
func TestAttributionIsComplete(t *testing.T) {
	_, body := get(t, newTestServer(t), "/")
	f := body[strings.Index(body, "<footer>"):]

	for _, want := range []string{
		"Tobias",                              // AVONET creator
		"10.6084/m9.figshare.16586228",        // AVONET source
		"Fox",                                 // wingspan creator
		"10.6084/m9.figshare.32089083",        // wingspan source
		"creativecommons.org/licenses/by/4.0", // the license itself
		"AviList",                             // every name on the page comes from it
		"10.2173/avilist.v2025b",              // AviList source
		"GBIF",                                // the occurrence records behind regions and scope
		"10.15468/aomfnb",                     // the eBird Observation Dataset
		"iNaturalist",                         // the other occurrence dataset
	} {
		if !strings.Contains(f, want) {
			t.Errorf("attribution is missing %q", want)
		}
	}
	// The ABA checklist must NOT be credited: decisions.scopeFromRecurrence replaced the one thing
	// it supplied, and crediting a source for nothing is its own kind of
	// false statement -- it tells a reader the page reproduces something it
	// does not, which is what blocked deployment in the first place.
	for _, gone := range []string{"American Birding Association", "ABA Checklist"} {
		if strings.Contains(f, gone) {
			t.Errorf("the footer still credits %q, but nothing on the page comes from it "+
				"since scope was derived (decisions.scopeFromRecurrence)", gone)
		}
	}
	// And the honesty note that the numbers are mostly modeled.
	if !strings.Contains(f, "modeled") {
		t.Error("footer does not disclose that lengths are modeled")
	}
	// The detail view is a shareable URL of its own, so it carries credit too.
	_, detail := get(t, s2(t), "/species/greblu")
	if !strings.Contains(detail, "creativecommons.org/licenses/by/4.0") {
		t.Error("detail view carries no license attribution")
	}
}

func s2(t *testing.T) *server { t.Helper(); return newTestServer(t) }

// TestNoUnearnedPermissionClaim guards against the page asserting a permission
// the project does not hold.
//
// The ABA checklist is all rights reserved and BirdView has not asked for a
// licence yet. A footer reading "used by permission" was written during the
// attribution pass and is simply false; claiming a right you have not been
// granted is worse than claiming none, because it misleads the rights holder's
// own readers.
//
// Delete this test only when a written grant exists and is recorded in
// river/requirements/requirements.md.
//
// Verifies: goals.goodDataCitizen. Criterion: acceptance.attributionComplete.
func TestNoUnearnedPermissionClaim(t *testing.T) {
	for _, page := range []string{"/", "/species/greblu"} {
		_, body := get(t, newTestServer(t), page)
		for _, claim := range []string{"used by permission", "with permission", "licensed from", "under licence", "under license"} {
			if strings.Contains(strings.ToLower(body), claim) {
				t.Errorf("%s claims %q; no such grant exists", page, claim)
			}
		}
	}
}

// neckRow pulls the Neck row's displayed value out of a detail page, so the
// assertion is scoped to that row rather than to the whole document. A bare
// search for a formatted number matches the bill, the tail or a section header
// just as happily, which is how several checks in this suite previously passed
// without testing anything.
func neckRow(t *testing.T, body string) string {
	t.Helper()
	return mustMatch(t, body, `<th>Neck</th>\s*<td class="v">([^<]+)</td>`, "the Neck row")
}

// TestFoldedNeckRendersAsAnArc asserts product.segmentComposition where the user sees it: a heron's
// neck reaches the page as an SVG arc path, every other bird's as a straight
// line, and the detail view still reports the neck at its full modeled length.
//
// That last part is the requirement's real content. The fold is a change to the
// silhouette only -- if it ever reached the numbers, BirdView would be quietly
// reporting a shorter heron than it measured.
//
// Verifies: product.segmentComposition, product.segmentsSumToLength, product.tileOpensDetail. Criterion: acceptance.foldedNeckSameStroke.
func TestFoldedNeckRendersAsAnArc(t *testing.T) {
	s := newTestServer(t)
	_, body := get(t, s, "/")

	var folded, straight int
	for _, d := range diagrams(t, body) {
		name := mustMatch(t, d, `aria-label="scaled outline of ([^"]+)"`, "the diagram label")
		isArc := strings.Contains(d, `<path class="neck"`)
		isLine := strings.Contains(d, `<line class="neck"`)
		if isArc == isLine {
			t.Errorf("%s: neck is drawn as %d path and %d line elements; want exactly one of each kind",
				name, boolToInt(isArc), boolToInt(isLine))
			continue
		}
		// Great Blue Heron is the fixture's only Ardeidae.
		if want := name == "Great Blue Heron"; isArc != want {
			t.Errorf("%s: neck drawn as an arc = %v, want %v", name, isArc, want)
		}
		if isArc {
			folded++
		} else {
			straight++
		}
	}
	if folded == 0 || straight == 0 {
		t.Fatalf("%d folded and %d straight necks rendered; the fixture must exercise both",
			folded, straight)
	}

	// The metric is untouched: the heron's neck is 200 mm in the fixture, and
	// the detail view must say so rather than the 12.7 cm the S spans.
	_, det := get(t, s, "/species/greblu?units=metric")
	if got := neckRow(t, det); got != "20.0 cm" {
		t.Errorf("detail view reports the neck as %q, want the full modeled 20.0 cm -- the "+
			"fold has leaked out of the drawing and into the data", got)
	}
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// TestINatLinkOnlyForAConfidentMatch asserts product.inaturalistLink on the rendered page: a
// species with a mapping gets a link to exactly its taxon page, and one without
// gets nothing -- no placeholder, no search URL, no "not found" row.
//
// The absence is the half worth testing. A search link would never 404, which is
// exactly what makes it the wrong answer: it always looks like it worked. An
// absent link is legible.
//
// Verifies: product.inaturalistLink. Criterion: acceptance.inaturalistLinkConfidentOnly.
func TestINatLinkOnlyForAConfidentMatch(t *testing.T) {
	s := newTestServer(t)

	_, matched := get(t, s, "/species/greblu")
	href := mustMatch(t, matched, `<a href="(https://www\.inaturalist\.org/taxa/[0-9]+)"`, "the iNaturalist link")
	if href != "https://www.inaturalist.org/taxa/4956" {
		t.Errorf("link points at %q, want the heron's taxon page", href)
	}
	if !strings.Contains(matched, "Great Blue Heron on\niNaturalist") &&
		!strings.Contains(matched, "Great Blue Heron on iNaturalist") {
		t.Error("the link does not name the species and the destination")
	}

	// The unmatched species must show no iNaturalist reference of any kind.
	_, unmatched := get(t, s, "/species/blasti")
	if strings.Contains(unmatched, "inaturalist.org") {
		t.Error("a species with no confident match still mentions iNaturalist; it should show nothing")
	}
	if strings.Contains(strings.ToLower(unmatched), "not found on inaturalist") {
		t.Error("a species with no confident match shows a placeholder; it should show nothing")
	}
	// And it is otherwise a normal page, so the check above is not passing
	// because the page failed to render.
	if !strings.Contains(unmatched, "Black-necked Stilt") {
		t.Fatal("the unmatched species page did not render")
	}
}

// TestNoINaturalistImagesAreEmbedded states the intent behind BirdView linking
// rather than copying (product.inaturalistLink, inatterms.ContributorKeepsRights). acceptance.noServerNetworkRequest's subresource scan already
// forbids the mechanism for tech.selfContainedServer reasons; this says why it also matters here.
//
// Photographs on iNaturalist are contributors' work under their own licenses,
// frequently CC BY-NC. Linking engages none of them. Embedding would engage all
// of them, per photograph.
//
// Verifies: product.inaturalistLink. Criterion: acceptance.inaturalistLinkConfidentOnly.
func TestNoINaturalistImagesAreEmbedded(t *testing.T) {
	_, body := get(t, newTestServer(t), "/species/greblu")
	for _, bad := range []string{
		`src="https://static.inaturalist.org`,
		`src="https://inaturalist-open-data`,
		`<img`,
	} {
		if strings.Contains(body, bad) {
			t.Errorf("detail page contains %q; BirdView links to iNaturalist and copies nothing from it", bad)
		}
	}
}

// TestFaviconIsServedLocally asserts the icon is BirdView's own, not a remote
// one (tech.selfContainedServer), and that both pages actually reference it.
//
// No new requirement was needed: tech.selfContainedServer already forbids fetching a subresource
// from another host, and acceptance.noServerNetworkRequest's scan already covers a `<link href>` pointing
// off-host. What that scan cannot see is a favicon that is referenced but not
// served, or served but referenced by nothing -- so this checks the two ends
// match.
//
// Verifies: tech.selfContainedServer. Criterion: acceptance.ownRenderableFavicon.
func TestFaviconIsServedLocally(t *testing.T) {
	s := newTestServer(t)

	code, body := get(t, s, "/static/favicon.svg")
	if code != http.StatusOK {
		t.Fatalf("favicon returned %d; the pages reference an asset that is not served", code)
	}
	// Parse it, rather than searching for "<svg". The first version of this
	// check looked for that substring and passed against a file XML cannot
	// read: the comment contained a double hyphen, which XML forbids, so every
	// browser would have shown no icon at all. A format assertion that does not
	// run the parser is not a format assertion.
	var probe struct {
		XMLName xml.Name
		ViewBox string `xml:"viewBox,attr"`
	}
	if err := xml.Unmarshal([]byte(body), &probe); err != nil {
		t.Fatalf("favicon is not well-formed XML, so no browser will render it: %v", err)
	}
	if probe.XMLName.Local != "svg" {
		t.Errorf("favicon root element is %q, want svg", probe.XMLName.Local)
	}
	if probe.ViewBox == "" {
		t.Error("favicon has no viewBox, so it will not scale to a 16px tab")
	}
	// The palette check that used to sit here searched for three hardcoded
	// literals, one of which (#5ccbfa) appeared nowhere but the favicon -- so
	// it compared the icon against a copy of itself and could not fail. It is
	// now derived from the stylesheet in web.TestFaviconUsesTheAppPalette
	// (product.oneDiagramVocabulary, acceptance.faviconFromPalette), which lives where both assets are embedded.

	for _, path := range []string{"/", "/species/greblu"} {
		_, page := get(t, s, path)
		if !strings.Contains(page, `href="/static/favicon.svg"`) {
			t.Errorf("%s does not reference the favicon", path)
		}
	}

	// A browser that probes /favicon.ico must get a clean 404, not the index
	// page -- the catch-all route makes that worth pinning.
	if code, _ := get(t, s, "/favicon.ico"); code != http.StatusNotFound {
		t.Errorf("/favicon.ico returned %d, want 404", code)
	}
}

// TestCrossbarSplitsAtTheChord asserts product.measuredVersusModeled on the rendered page: a measured
// wingspan draws no dashed segment, a modeled one draws exactly one, flanked by
// two solid segments and marked by two ticks.
//
// Verifies: product.measuredVersusModeled. Criterion: acceptance.diagramDistinguishesModeled.
func TestCrossbarSplitsAtTheChord(t *testing.T) {
	s := newTestServer(t)
	wing := regexp.MustCompile(`<line class="(wing(?: est)?)"`)

	cases := []struct {
		id, name    string
		spanModeled bool
	}{
		{"greblu", "Great Blue Heron", false},
		{"amecro", "American Crow", true},
	}
	for _, c := range cases {
		_, body := get(t, s, "/species/"+c.id)
		var solid, dash int
		for _, m := range wing.FindAllStringSubmatch(body, -1) {
			if strings.Contains(m[1], "est") {
				dash++
			} else {
				solid++
			}
		}
		ticks := strings.Count(body, `class="chordtick"`)
		if c.spanModeled {
			if dash != 1 || solid != 2 {
				t.Errorf("%s: modeled span drew %d solid and %d dashed wing segments, want 2 and 1",
					c.name, solid, dash)
			}
		} else if dash != 0 || solid != 3 {
			// A measured span still splits into three; all three are solid.
			t.Errorf("%s: measured span drew %d solid and %d dashed wing segments, want 3 and 0",
				c.name, solid, dash)
		}
		// Both cases split, so both carry ticks: the split shows where the
		// measurement reaches, which is true whether or not the span is modeled.
		if ticks != 2 {
			t.Errorf("%s: %d chord ticks, want 2", c.name, ticks)
		}
	}
}

// TestBothPagesExplainTheStrokeCode asserts the other half of product.measuredVersusModeled: a visual
// code nobody documents is decoration.
//
// Verifies: product.measuredVersusModeled. Criterion: acceptance.diagramDistinguishesModeled.
func TestBothPagesExplainTheStrokeCode(t *testing.T) {
	s := newTestServer(t)
	for _, path := range []string{"/", "/species/greblu"} {
		_, body := get(t, s, path)
		low := strings.ToLower(body)
		for _, want := range []string{"solid", "dashed", "tick", "chord"} {
			if !strings.Contains(low, want) {
				t.Errorf("%s does not explain %q; the stroke code is undocumented there", path, want)
			}
		}
	}
}

// TestDetailDiagramIsLabeled asserts product.labeledDetailDiagram on the rendered page: five labels,
// each naming its segment and carrying the value the provenance table gives for
// it.
//
// The values are read out of the table rather than recomputed here, so the two
// cannot drift apart. A label that disagreed with the table would be the worst
// outcome of this feature -- two numbers for one segment on one page.
//
// Verifies: product.labeledDetailDiagram. Criterion: acceptance.segmentLabelsDontCollide.
func TestDetailDiagramIsLabeled(t *testing.T) {
	_, body := get(t, newTestServer(t), "/species/greblu?units=metric")

	// Matches a segment label, dashed or not, but deliberately NOT the chord:
	// that one is checked separately below, and folding it in here would let a
	// missing segment hide behind it.
	labels := regexp.MustCompile(`<text class="dlabel(?: est)?"[^>]*>([^<]+)</text>`).FindAllStringSubmatch(body, -1)
	if len(labels) != 5 {
		t.Fatalf("detail diagram carries %d segment labels, want one per segment", len(labels))
	}
	got, marked := map[string]string{}, map[string]bool{}
	for _, m := range labels {
		f := strings.Fields(m[1])
		seg := strings.ToLower(f[0])
		val := strings.Join(f[1:], " ")
		// The estimated marker is emitted in the markup, so strip it before
		// comparing with the table -- and remember that it was there.
		if marked[seg] = strings.HasSuffix(val, " ~"); marked[seg] {
			val = strings.TrimSuffix(val, " ~")
		}
		got[seg] = val
	}
	for _, seg := range []string{"bill", "head", "neck", "body", "tail"} {
		val, ok := got[seg]
		if !ok {
			t.Errorf("no label for the %s", seg)
			continue
		}
		name := strings.ToUpper(seg[:1]) + seg[1:]
		row := mustMatch(t, body,
			`<th>`+name+`</th>\s*<td class="v">([^<]+)</td>`, "the "+seg+" row")
		if val != row {
			t.Errorf("%s: diagram says %q, provenance table says %q", seg, val, row)
		}
		// The marker must track the provenance the table states for that same
		// row, not a constant: asserted on the emitted text, because the class
		// attribute was present and correct while nothing was drawn.
		kind := mustMatch(t, body,
			`<th>`+name+`</th>\s*<td class="v">[^<]+</td>\s*<td class="k"><span class="tag t-([a-z-]+)">`,
			"the "+seg+" provenance kind")
		if want := kind == "modeled"; marked[seg] != want {
			t.Errorf("%s: label %q carries the ~ marker = %v, but the table calls it %q",
				seg, got[seg], marked[seg], kind)
		}
	}

	// The wing chord is labeled on the crossbar, and like every other label its
	// value is read out of the provenance table so the two cannot drift.
	chord := mustMatch(t, body, `<text class="dlabel chord"[^>]*>([^<]+)</text>`, "the wing chord label")
	if !strings.HasPrefix(chord, "Wing chord ") {
		t.Errorf("chord label is %q; it should name what it measures", chord)
	}
	if got, want := strings.TrimSpace(strings.TrimPrefix(chord, "Wing chord")),
		mustMatch(t, body, `<th>Wing chord</th>\s*<td class="v">([^<]+)</td>`, "the wing chord row"); got != want {
		t.Errorf("wing chord: diagram says %q, provenance table says %q", got, want)
	}

	// Leaders tie each label to its segment.
	if n := strings.Count(body, `class="leader"`); n != 5 {
		t.Errorf("%d leader lines, want 5", n)
	}
	// A tile has none: the index is dense on purpose.
	_, index := get(t, newTestServer(t), "/")
	if strings.Contains(tileArea(t, index), `class="dlabel`) {
		t.Error("index tiles carry diagram labels; those belong to the detail view and the legend")
	}
}

// TestWingLabelsClearTheColumn asserts the geometry the wing labels depend on:
// the wingspan label sits just beyond the wingtip, so the bird must be laid out
// narrow enough that it never reaches the segment label column.
//
// Checked across every species rather than one, because the binding case is the
// widest-winged bird and which species that is changes with the data. Before
// the gutter existed a Black-footed Albatross's crossbar reached x=988 in a
// 988-wide box while the column began at 838, so the bar crossed every label on
// that side.
//
// Verifies: product.labeledDetailDiagram. Criterion: acceptance.segmentLabelsDontCollide.
func TestWingLabelsClearTheColumn(t *testing.T) {
	// Wider than any label this renders; if the estimate is ever wrong it is
	// wrong in the safe direction.
	const spanLabelWidth = 150

	full := view.Box{W: tileBox.W * 2.6, H: tileBox.H * 1.9}
	inner := view.Box{W: full.W - labelGutter, H: full.H}
	columnX := full.W - 150

	ds := testDataset()
	var worst float64 = 1e9
	var worstName string
	for i := range ds.Species {
		sp := &ds.Species[i]
		g := view.Layout(sp, sp.TotalLength.Value, sp.Wingspan.Value, inner)
		end := g.CX + g.CrossbarHalfWidth + 8 + spanLabelWidth
		if gap := columnX - end; gap < worst {
			worst, worstName = gap, sp.Name(dataset.LocaleNA)
		}
	}
	if worst < 0 {
		t.Errorf("%s: the wingspan label would reach the segment column by %.1f px; "+
			"widen labelGutter", worstName, -worst)
	}

	// And the gutter must not be so wide that it is doing nothing: the widest
	// bird should still use most of the space it has.
	if labelGutter >= full.W/2 {
		t.Errorf("labelGutter is %d of a %.0f box; the bird has less room than its labels",
			labelGutter, full.W)
	}
}

// TestEveryControlEntryIsWrapped is the markup half of the column-break fix:
// the CSS rule only helps if every heading-and-text pair is actually inside a
// wrapper for it to act on.
//
// Verifies: product.pageExplainsItself. Criterion: acceptance.pageExplainsControls.
func TestEveryControlEntryIsWrapped(t *testing.T) {
	_, body := get(t, newTestServer(t), "/")
	i := strings.Index(body, `<div class="ctrlhelp">`)
	if i < 0 {
		t.Fatal("no control explanation block")
	}
	block := body[i:]
	block = block[:strings.Index(block, "</details>")]

	dts := strings.Count(block, "<dt>")
	wrapped := strings.Count(block, `<div class="ctrl"><dt>`)
	if dts == 0 {
		t.Fatal("no control entries found")
	}
	if wrapped != dts {
		t.Errorf("%d of %d control entries are wrapped for break-inside; an unwrapped one can "+
			"have its heading split from its text across the column break", wrapped, dts)
	}
}

// TestDiagramTextPaintsOverLines asserts the paint order of the annotated
// diagram: every line is emitted before every label, so a leader that crosses a
// label passes behind the words.
//
// SVG has no z-index; document order IS paint order. The wing labels used to
// sit inside the centering group, which put them ahead of the leaders and let
// the neck's leader strike through "Wing chord" and "Wingspan".
//
// Verifies: product.labeledDetailDiagram. Criterion: acceptance.segmentLabelsDontCollide.
func TestDiagramTextPaintsOverLines(t *testing.T) {
	_, body := get(t, newTestServer(t), "/species/greblu")
	svg := mustMatch(t, body, `(?s)(<svg class="dia".*?</svg>)`, "the detail diagram")

	lastLine := strings.LastIndex(svg, "<line")
	firstText := strings.Index(svg, "<text")
	if lastLine < 0 || firstText < 0 {
		t.Fatalf("diagram has %d lines and %d texts; expected both", lastLine, firstText)
	}
	if lastLine > firstText {
		t.Errorf("a <line> is emitted after the first <text> (line at %d, text at %d); "+
			"SVG paints in document order, so that line is drawn over the label",
			lastLine, firstText)
	}

	// And the labels really do exist to be covered, or the check above is
	// satisfied by a diagram with no text at all.
	if n := strings.Count(svg, "<text"); n < 7 {
		t.Errorf("only %d labels in the diagram; expected five segments plus chord and wingspan", n)
	}
}

// TestHeaderNamesThePanelsThatExist guards against the header pointing at a
// panel by a name it no longer has.
//
// The intro read "Read the Instructions for details" for a while after the
// panel was retitled Presentation, so it directed the reader to something that
// was not on the page. Nothing caught it: the panel had a test, the header had
// a test, and neither knew about the other.
//
// Verifies: product.pageExplainsItself. Criterion: acceptance.pageExplainsControls.
func TestHeaderNamesThePanelsThatExist(t *testing.T) {
	_, body := get(t, newTestServer(t), "/")
	intro := mustMatch(t, body, `(?s)<p class="sub">(.*?)</p>`, "the header intro")

	// The title is the text before any markup: a summary now also carries a
	// <span class="sumdesc"> describing what the panel holds (product.collapsibleControls), so
	// anchoring on </summary> matches nothing.
	titles := regexp.MustCompile(`<summary>([^<]+)`).FindAllStringSubmatch(body, -1)
	have := map[string]bool{}
	for _, m := range titles {
		have[strings.TrimSpace(m[1])] = true
	}
	if len(have) == 0 {
		t.Fatal("no collapsible panels found")
	}
	// Any capitalised panel-ish word the intro points the reader at must be a
	// panel that exists.
	for _, word := range regexp.MustCompile(`\b(Instructions|Presentation|Controls)\b`).FindAllString(intro, -1) {
		if !have[word] {
			t.Errorf("the header sends the reader to %q, but the panels are %v", word, keysOf(have))
		}
	}
}

func keysOf(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// TestDetailDiagramLabelsTotalLength asserts product.labeledDetailDiagram's total length label: the one
// annotation that names the whole bird rather than a part of it.
//
// The Great Blue Heron is not an arbitrary fixture. Ardeidae folds its neck
// (decisions.foldedNeckAsS), so its drawn height is 6-9% short of its total length -- which is
// what makes the value assertion below bite. For every unfolded bird the two
// agree exactly, so a version of this test built on any other species would
// pass just as happily against a label derived from the drawing.
//
// Verifies: product.labeledDetailDiagram. Criterion: acceptance.totalLengthLabeled.
func TestDetailDiagramLabelsTotalLength(t *testing.T) {
	_, body := get(t, newTestServer(t), "/species/greblu?units=metric")

	el := mustMatch(t, body, `(<text class="dlabel total[^"]*"[^>]*>[^<]+</text>)`,
		"the total length label")

	// Horizontal: SVG text rotates only via a transform, so its absence is the
	// assertion. Stated explicitly because "horizontal" is otherwise invisible.
	if strings.Contains(el, "transform") || strings.Contains(el, "rotate") {
		t.Errorf("the total length label is rotated; product.labeledDetailDiagram requires it horizontal: %s", el)
	}
	if !strings.Contains(el, `text-anchor="middle"`) {
		t.Errorf("the total length label is not centred: %s", el)
	}

	// Centred on the BIRD's axis, not the box's. The two differ by half the
	// label gutter, so a label centred on the box sits well to the right of the
	// bird it names.
	lt := labelledTile(findSpecies(t, "greblu"), view.Box{W: tileBox.W * 2.6, H: tileBox.H * 1.9}, units{metric: true}, dataset.LocaleNA)
	if x := attrF(t, el, "x"); math.Abs(x-lt.Geom.CX) > 0.05 {
		t.Errorf("total length label at x=%.1f; the bird's axis is %.1f (box centre is %.1f)",
			x, lt.Geom.CX, lt.Box.W/2)
	}
	// Below the tail tip, not above it.
	tail := lt.TopOffset + lt.Geom.YTailTip
	if y := attrF(t, el, "y"); y <= tail {
		t.Errorf("total length label at y=%.1f, tail tip at %.1f; it belongs below the tail", y, tail)
	}

	// It states the bird's length, read from the provenance table -- never the
	// drawn height, which for this folded-neck bird is a different number.
	text := mustMatch(t, el, `>([^<]+)</text>`, "the label text")
	val := strings.TrimSpace(strings.TrimPrefix(text, "Total length"))
	marked := strings.HasSuffix(val, " ~")
	val = strings.TrimSuffix(val, " ~")
	row := mustMatch(t, body, `<th>Total length</th>\s*<td class="v">([^<]+)</td>`,
		"the total length row")
	if val != row {
		t.Errorf("total length: diagram says %q, provenance table says %q", val, row)
	}
	// Guard the guard: if the fixture ever stops folding its neck, the check
	// above goes vacuous and should say so rather than quietly weaken.
	if !lt.Geom.NeckFolded {
		t.Fatal("the fixture no longer folds its neck, so this test can no longer " +
			"distinguish the bird's length from the drawing's height; pick another Ardeidae")
	}

	// The marker tracks the provenance the table states, not a constant.
	kind := mustMatch(t, body,
		`<th>Total length</th>\s*<td class="v">[^<]+</td>\s*<td class="k"><span class="tag t-([a-z-]+)">`,
		"the total length provenance kind")
	if want := kind == "modeled"; marked != want {
		t.Errorf("total length label carries the ~ marker = %v, but the table calls it %q",
			marked, kind)
	}

	// No leader: there is no part to point at.
	if n := strings.Count(body, `class="leader"`); n != 5 {
		t.Errorf("%d leader lines; the total length label should not add one", n)
	}

	// The legend carries it; an index tile does not.
	_, index := get(t, newTestServer(t), "/")
	if !strings.Contains(index, `class="dlabel total`) {
		t.Error("the help legend has no total length label; it renders through the same builder")
	}
	if strings.Contains(tileArea(t, index), `class="dlabel`) {
		t.Error("index tiles carry diagram labels; those belong to the detail view and the legend")
	}
}

// TestTotalLengthLabelStaysInTheBox asserts the room half of acceptance.totalLengthLabeled across every
// species, because the binding case is whichever bird is tallest relative to its
// wingspan and that changes with the data.
//
// Without the reserved strip this fails for essentially the whole dataset: the
// detail view scales each bird to fill the drawing area, so the tail tip lands
// on its bottom edge and the label is drawn outside the viewBox.
//
// Verifies: product.labeledDetailDiagram. Criterion: acceptance.totalLengthLabeled.
func TestTotalLengthLabelStaysInTheBox(t *testing.T) {
	// The 11px text's descender. The baseline may sit this far above the edge
	// and no closer.
	const descender = 3

	ds := testDataset()
	for i := range ds.Species {
		sp := &ds.Species[i]
		lt := labelledTile(sp, view.Box{W: tileBox.W * 2.6, H: tileBox.H * 1.9}, units{metric: true}, dataset.LocaleNA)
		if len(lt.FootLabels) == 0 {
			t.Fatalf("%s: no whole-bird labels at all", sp.Name(dataset.LocaleNA))
		}
		// Two things hang below the drawing and the strip must clear whichever
		// reaches lower: the stack under the tail, and the mass label under
		// the cube. Checking only one leaves the other free to overrun.
		last := lt.FootLabels[len(lt.FootLabels)-1]
		for _, c := range []struct {
			what string
			y    float64
		}{{last.Text, last.Y}, {lt.MassCube.Label, lt.MassCube.LabelY}} {
			if c.y+descender > lt.Box.H {
				t.Errorf("%s: %q baseline at %.1f (+%d descender) overruns the %.1f box",
					sp.Name(dataset.LocaleNA), c.what, c.y, descender, lt.Box.H)
			}
		}
	}
}

// findSpecies returns a fixture species by id, so a test can rebuild the same
// tile the handler built and compare against it rather than against a number
// copied out of a run.
func findSpecies(t *testing.T, id string) *dataset.Species {
	t.Helper()
	ds := testDataset()
	for i := range ds.Species {
		if ds.Species[i].ID == id {
			return &ds.Species[i]
		}
	}
	t.Fatalf("no fixture species %q", id)
	return nil
}

// attrF reads a numeric attribute off a single rendered element.
func attrF(t *testing.T, el, name string) float64 {
	t.Helper()
	m := regexp.MustCompile(name + `="([0-9.-]+)"`).FindStringSubmatch(el)
	if m == nil {
		t.Fatalf("no %s attribute on %s", name, el)
	}
	v, err := strconv.ParseFloat(m[1], 64)
	if err != nil {
		t.Fatalf("%s=%q: %v", name, m[1], err)
	}
	return v
}

// TestEveryOrderGroupWrapsItsFamilies is the markup half of the wrapped-family
// indent: the two-column grid only indents anything if the families are in a
// box of their own, and every group's chips are in it.
//
// A chip left as a direct child of .ogroup becomes a third grid item, which
// lands in a new implicit row at the page margin -- the exact bug the grid was
// added to fix, reintroduced for one order.
//
// Verifies: product.familyFilter. Criterion: acceptance.familyListWraps.
func TestEveryOrderGroupWrapsItsFamilies(t *testing.T) {
	_, body := get(t, newTestServer(t), "/")

	// Split on the group opener rather than matching balanced </div>s: the
	// nesting is exactly what this test exists to check, so a parse that
	// depends on it reports a mangled structure as "no groups" instead of as
	// the defect it is.
	block := body[strings.Index(body, `<div class="ordergroups">`):]
	block = block[:strings.Index(block, `<p class="count">`)]
	parts := strings.Split(block, `<div class="ogroup">`)[1:]
	groups := make([][]string, 0, len(parts))
	for _, p := range parts {
		groups = append(groups, []string{"", p})
	}
	if len(groups) == 0 {
		t.Fatal("no order groups on the page")
	}
	// The fixture spans several orders; if it ever collapses to one this test
	// stops covering the multi-order case and should say so.
	if len(groups) < 2 {
		t.Fatalf("only %d order group in the fixture; this cannot show that EVERY group "+
			"wraps its families", len(groups))
	}

	totalChips := 0
	for _, g := range groups {
		inner := g[1]
		if n := strings.Count(inner, `<div class="ofams">`); n != 1 {
			t.Errorf("an order group has %d family boxes, want exactly 1:\n%.200s", n, inner)
			continue
		}
		// Every chip must sit after the wrapper opens. A chip before it is a
		// direct child of the grid and wraps to the margin.
		open := strings.Index(inner, `<div class="ofams">`)
		if before := strings.Count(inner[:open], `class="chip`); before > 0 {
			t.Errorf("%d family chips sit outside the .ofams box, so they wrap to the "+
				"page margin instead of indenting", before)
		}
		n := strings.Count(inner, `class="chip`)
		if n == 0 {
			t.Error("an order group has no family chips at all")
		}
		totalChips += n
		// The order header stays out of the families' box: it is the first
		// grid column, and inside .ofams it would wrap along with the chips.
		if hdr := strings.Index(inner, `class="btn ohdr`); hdr < 0 || hdr > open {
			t.Error("the order header is missing or sits inside the family box; it must be " +
				"the grid's first column")
		}
	}
	// Month chips are deliberately outside any group: twelve items in a row
	// need no zones, and decisions.monthPresence rejected season groups outright because a
	// season name inverts between hemispheres and means nothing in the
	// tropics. They are subtracted rather than folded in, so this still
	// catches a family or region chip that escapes its box.
	outside := strings.Count(body, `class="chip`) -
		strings.Count(tileArea(t, body), `class="chip`) -
		strings.Count(body, `class="chip mchip`)
	if want := outside; totalChips != want {
		t.Errorf("the groups account for %d chips but the page has %d; some chip is outside "+
			"an order group entirely", totalChips, want)
	}
}

// drawnSpans pulls each compare panel's species name, drawn bill-to-tail height
// and crossbar half-width out of the rendered page, in document order.
func drawnSpans(t *testing.T, body string) []struct {
	Name   string
	Height float64
	Half   float64
} {
	t.Helper()
	var out []struct {
		Name   string
		Height float64
		Half   float64
	}
	for _, m := range regexp.MustCompile(`(?s)<svg class="dia".*?</svg>`).FindAllString(body, -1) {
		name := html.UnescapeString(
			mustMatch(t, m, `aria-label="scaled outline of ([^"]+)"`, "the panel's species"))
		tail := attrF(t, mustMatch(t, m, `(<line class="tail"[^>]*>)`, "the tail line"), "y2")
		bill := attrF(t, mustMatch(t, m, `(<line class="bill"[^>]*>)`, "the bill line"), "y1")
		wing := mustMatch(t, m, `(<line class="wing[^"]*"[^>]*>)`, "a wing line")
		cx := attrF(t, mustMatch(t, m, `(<line class="bill"[^>]*>)`, "the bill line"), "x1")
		out = append(out, struct {
			Name   string
			Height float64
			Half   float64
		}{name, tail - bill, cx - attrF(t, wing, "x1")})
	}
	return out
}

// TestCompareSharesOneScale is the assertion the comparison view exists for:
// every bird is laid out against the largest in the set, not against itself.
//
// This is the failure the code is one argument away from: labelledTileAt with
// the bird's own length produces a page that looks completely correct, with
// every panel neatly filled, and compares nothing. No assertion on the markup
// can see it -- only the drawn geometry can.
//
// Verifies: product.comparisonOneScale. Criterion: acceptance.comparisonSharesScale.
func TestCompareSharesOneScale(t *testing.T) {
	s := newTestServer(t)
	// Unfolded families only. A folded neck draws shorter than the bird is
	// long (decisions.foldedNeckAsS), so its height is deliberately NOT proportional and would
	// make this test wrong rather than strict.
	const q = "/compare?units=metric&species=annhum&species=amerob&species=amecro"
	_, body := get(t, s, q)

	got := drawnSpans(t, body)
	if len(got) != 3 {
		t.Fatalf("%d panels, want 3", len(got))
	}
	for _, p := range got {
		if sp := findSpeciesByName(t, p.Name); view.Layout(sp, sp.TotalLength.Value,
			sp.Wingspan.Value, view.Box{W: 100, H: 100}).NeckFolded {
			t.Fatalf("%s folds its neck, so its drawn height is not proportional to its "+
				"length and this fixture can no longer test the scale", p.Name)
		}
	}

	// Drawn height must track total length by one constant across the set.
	// Derived from the requirement -- one scale for the group -- not from a run.
	var scale float64
	for _, p := range got {
		sp := findSpeciesByName(t, p.Name)
		px := p.Height / sp.TotalLength.Value
		if scale == 0 {
			scale = px
			continue
		}
		if math.Abs(px-scale)/scale > 0.005 {
			t.Errorf("%s is drawn at %.5f px/mm but the first bird at %.5f; the panels do not "+
				"share a scale", p.Name, px, scale)
		}
	}

	// Proportionality alone does not prove the reference is the LARGEST bird:
	// every panel at half scale would pass it too. The largest bird pins it --
	// at ref = max it is drawn exactly as it is on its own detail page, where
	// it is laid out against itself.
	big := findSpecies(t, "amecro")
	_, solo := get(t, s2(t), "/species/amecro?units=metric")
	soloH := drawnSpans(t, solo)[0].Height
	var bigH float64
	for _, p := range got {
		if p.Name == big.Name(dataset.LocaleNA) {
			bigH = p.Height
		}
	}
	if math.Abs(bigH-soloH) > 0.2 {
		t.Errorf("the largest bird is drawn %.1fpx here but %.1fpx on its own page; the shared "+
			"scale is not set by the largest bird in the set", bigH, soloH)
	}
	// And every other bird must be strictly smaller, or the scale is per-panel.
	for _, p := range got {
		if p.Name != big.Name(dataset.LocaleNA) && p.Height >= bigH {
			t.Errorf("%s is drawn %.1fpx, not smaller than the largest bird's %.1fpx",
				p.Name, p.Height, bigH)
		}
	}
}

// findSpeciesByName looks a fixture species up by common name, which is what
// the rendered diagram carries.
func findSpeciesByName(t *testing.T, common string) *dataset.Species {
	t.Helper()
	ds := testDataset()
	for i := range ds.Species {
		if ds.Species[i].Name(dataset.LocaleNA) == common {
			return &ds.Species[i]
		}
	}
	t.Fatalf("no fixture species named %q", common)
	return nil
}

// TestCompareSelectsAndExplainsItself covers the rest of acceptance.comparisonSharesScale: which birds
// appear, in what order, and what the page does and does not carry.
//
// Verifies: product.comparisonOneScale. Criterion: acceptance.comparisonSharesScale.
func TestCompareSelectsAndExplainsItself(t *testing.T) {
	s := newTestServer(t)

	// Named order is the reader's and is preserved.
	_, body := get(t, s, "/compare?species=greblu&species=annhum")
	names := []string{}
	for _, p := range drawnSpans(t, body) {
		names = append(names, p.Name)
	}
	if want := []string{"Great Blue Heron", "Anna's Hummingbird"}; !slices.Equal(names, want) {
		t.Errorf("panels are %v, want %v in the order named", names, want)
	}
	// Reversing the URL reverses the page, or the order is not being honored.
	_, rev := get(t, s, "/compare?species=annhum&species=greblu")
	if first := drawnSpans(t, rev)[0].Name; first != "Anna's Hummingbird" {
		t.Errorf("reversing the URL gave %q first; the page is imposing its own order", first)
	}

	// A bird named twice is drawn once.
	_, dup := get(t, s, "/compare?species=annhum&species=annhum")
	if n := len(drawnSpans(t, dup)); n != 1 {
		t.Errorf("a species named twice drew %d panels, want 1", n)
	}

	// An unrecognised id is dropped and named, not refused (decisions.compareUrlContract). Checked
	// with a mixed set: an all-invalid one cannot tell dropping from refusing.
	code, msg := get(t, s, "/compare?species=annhum&species=no_such_bird")
	if code != http.StatusOK {
		t.Errorf("a set with one dead id gave %d; it should render what it recognises", code)
	}
	if n := len(drawnSpans(t, msg)); n != 1 {
		t.Errorf("%d panels, want the one recognised bird", n)
	}
	if !strings.Contains(msg, "no_such_bird") {
		t.Errorf("the page does not name the dropped id: a rotted link must say what it lost")
	}
	for _, m := range regexp.MustCompile(`href="([^"]+)"`).FindAllStringSubmatch(msg, -1) {
		if strings.Contains(html.UnescapeString(m[1]), "no_such_bird") {
			t.Errorf("link %q still carries the dead id; the set does not self-heal", m[1])
		}
	}

	// Naming none explains itself rather than rendering an empty page.
	_, none := get(t, s, "/compare")
	if len(drawnSpans(t, none)) != 0 {
		t.Error("/compare with no species drew a panel")
	}
	if !strings.Contains(none, "species") || !strings.Contains(none, "compare?species=") {
		t.Error("/compare with no species does not say how to name one")
	}

	// The diagram key is a page-level explanation: once, however many birds.
	for _, q := range []string{
		"/compare?species=annhum",
		"/compare?species=annhum&species=amerob&species=greblu",
	} {
		_, b := get(t, s, q)
		if n := strings.Count(b, `class="note diagramkey"`); n != 1 {
			t.Errorf("%s carries the diagram key %d times, want exactly 1", q, n)
		}
	}

	// No provenance table: product.comparisonOneScale trades it for a link to each detail view.
	_, three := get(t, s, "/compare?species=annhum&species=amerob")
	if strings.Contains(three, `class="prov"`) {
		t.Error("the compare page carries a provenance table; product.comparisonOneScale defers that to the detail view")
	}
	for _, id := range []string{"annhum", "amerob"} {
		if !strings.Contains(three, `href="/species/`+id) {
			t.Errorf("no link to /species/%s; that link is what makes omitting the table honest", id)
		}
	}
	// Marking survives even though explaining does not (product.tileMarksModeled).
	if !strings.Contains(three, " ~</text>") {
		t.Error("no estimated marker on any label; the compare view may omit the explanation " +
			"of provenance but not the marking of it")
	}
}

// TestCompareBackLinkKeepsTheComparison asserts the reversal in decisions.compareUrlContract: the
// species list goes back with the reader.
//
// It used to be dropped, and that was right while the index had no use for it
// -- a parameter that means nothing where it lands is litter, and the index
// copies whatever it is handed into every control link. product.pickFromIndex makes the index
// the picker, so the list is now live state and "back" means "add another
// bird". Dropping it would break the plainest flow the picker has.
//
// The display options must still survive, or back has stopped keeping the
// reader's place (product.togglesKeepPlace).
//
// Verifies: product.comparisonOneScale, product.pickFromIndex, product.togglesKeepPlace. Criterion: acceptance.comparisonSharesScale.
func TestCompareBackLinkKeepsTheComparison(t *testing.T) {
	s := newTestServer(t)
	want := []string{"annhum", "greblu"}
	_, body := get(t, s, "/compare?units=metric&scale=global&species=annhum&species=greblu")

	href := html.UnescapeString(mustMatch(t, body,
		`<a class="backlink" href="([^"]+)"`, "the back link"))
	q := mustQuery(t, href)
	if got := q["species"]; !slices.Equal(got, want) {
		t.Errorf("back link carries species=%v, want %v: returning to the picker must not "+
			"discard the comparison", got, want)
	}
	if q.Get("units") != "metric" || q.Get("scale") != "global" {
		t.Errorf("back link %q dropped the display options", href)
	}

	// And it says where it goes, since the index is now a picker mid-task.
	label := mustMatch(t, body, `class="backlink"[^>]*>(?s)(.*?)</a>`, "the back link text")
	if !strings.Contains(label, "picking") {
		t.Errorf("back link reads %q; it returns to the picker, not to an unfiltered list",
			strings.TrimSpace(label))
	}
}

// tileHrefs returns every index tile's own link.
func tileHrefs(t *testing.T, body string) []string {
	t.Helper()
	var out []string
	for _, m := range regexp.MustCompile(`<a class="tile" href="([^"]+)"`).FindAllStringSubmatch(body, -1) {
		out = append(out, html.UnescapeString(m[1]))
	}
	return out
}

// TestTileLinksCarryViewState asserts the half of product.pickFromIndex that is about what a
// link into the detail view carries.
//
// Before this, a tile link was a bare /species/{id}: opening a bird from a
// filtered, metric index landed on an unfiltered imperial page, and coming
// back landed on the unfiltered 841. The picker makes that fatal rather than
// merely untidy, because the picks would be lost on the way in -- before the
// back link is ever reached.
//
// `family` is the deliberate exception and is asserted as such: it is the only
// unbounded parameter, and carrying it would add ~1.4 MB to one page.
//
// Verifies: product.pickFromIndex, product.tileContents. Criterion: acceptance.pickModeSurvivesNavigation.
func TestTileLinksCarryViewState(t *testing.T) {
	s := newTestServer(t)
	// A filter is active, so `family` is present and would leak if carried.
	const q = "/?units=metric&by=wingspan&scale=global&step=30" +
		"&family=Trochilidae&family=Turdidae&species=amecro"
	_, body := get(t, s, q)

	hrefs := tileHrefs(t, body)
	if len(hrefs) == 0 {
		t.Fatal("no tiles on the page")
	}
	for _, h := range hrefs {
		u, err := url.Parse(h)
		if err != nil {
			t.Fatalf("unparseable tile href %q: %v", h, err)
		}
		if !strings.HasPrefix(u.Path, "/species/") {
			t.Errorf("tile href %q does not point at a detail view", h)
		}
		got := u.Query()
		for k, want := range map[string]string{
			"units": "metric", "by": "wingspan", "scale": "global", "step": "30",
			"species": "amecro",
		} {
			if got.Get(k) != want {
				t.Errorf("tile href %q: %s=%q, want %q", h, k, got.Get(k), want)
			}
		}
		if fam := got["family"]; len(fam) > 0 {
			t.Errorf("tile href %q carries family=%v; it is the one unbounded parameter and "+
				"would add ~1.4 MB across 841 tiles", h, fam)
		}
	}

	// And the state actually survives the hop: follow a tile, read its back
	// link. This is the round trip the picker depends on.
	_, detail := get(t, s, hrefs[0])
	back := html.UnescapeString(mustMatch(t, detail,
		`<a class="backlink" href="([^"]+)"`, "the detail page's back link"))
	bq, err := url.Parse(back)
	if err != nil {
		t.Fatalf("unparseable back link %q: %v", back, err)
	}
	for k, want := range map[string]string{
		"units": "metric", "by": "wingspan", "scale": "global", "step": "30",
		"species": "amecro",
	} {
		if bq.Query().Get(k) != want {
			t.Errorf("back link %q lost %s (got %q, want %q)", back, k, bq.Query().Get(k), want)
		}
	}
}

// pickBar returns just the pick bar's markup, or "" when there is none.
//
// Matched by its own container rather than by searching the page for "picked":
// that word is on every tile control too, so a page-wide search would report a
// bar that is not there. This trap is recorded in acceptance.pickModeSurvivesNavigation.
func pickBarOf(t *testing.T, body string) string {
	t.Helper()
	m := regexp.MustCompile(`(?s)<div class="pickbar">(.*?)</div>\s*</div>`).FindString(body)
	return m
}

func pickControls(body string) []string {
	var out []string
	// href="([^"]*)", not +: a control rendered with an EMPTY href is still a
	// control on the page, and a regex that skipped it would let "shown in
	// browsing mode" pass as "not shown at all".
	for _, m := range regexp.MustCompile(`<a class="pick( on)?" href="([^"]*)"`).
		FindAllStringSubmatch(body, -1) {
		out = append(out, html.UnescapeString(m[2]))
	}
	return out
}

// TestPickModeIsTheSpeciesParameter asserts the mode itself: absent, present
// and empty, present and populated.
//
// Verifies: product.pickFromIndex, product.tileContents. Criterion: acceptance.pickModeSurvivesNavigation.
func TestPickModeIsTheSpeciesParameter(t *testing.T) {
	s := newTestServer(t)

	_, browsing := get(t, s, "/")
	if n := len(pickControls(browsing)); n != 0 {
		t.Errorf("browsing mode shows %d pick controls; the tile a reader sees almost all "+
			"the time should be unchanged", n)
	}
	if pickBarOf(t, browsing) != "" {
		t.Error("browsing mode shows the pick bar")
	}

	// Bare species=: picking, nothing picked. The mode must hold open or
	// unpicking your last bird drops you out mid-task.
	_, empty := get(t, s, "/?species=")
	if n, want := len(pickControls(empty)), len(tileHrefs(t, empty)); n != want {
		t.Errorf("a bare species= gave %d pick controls for %d tiles", n, want)
	}
	if bar := pickBarOf(t, empty); !strings.Contains(bar, "No birds picked yet") {
		t.Errorf("a bare species= does not hold the mode open with an empty bar; got %q", bar)
	}

	// Populated: a control on every tile, and the picked one says so.
	_, picking := get(t, s, "/?species=annhum")
	if n, want := len(pickControls(picking)), len(tileHrefs(t, picking)); n != want {
		t.Errorf("%d pick controls for %d tiles; every tile needs one", n, want)
	}
	if n := strings.Count(picking, `class="pick on"`); n != 1 {
		t.Errorf("%d tiles render as picked, want exactly 1", n)
	}

	// product.tileOpensDetail in both modes: the tile itself still opens the detail view. This
	// is what fails if the control is ever built by rewriting the tile's href.
	for _, body := range []string{browsing, picking} {
		for _, h := range tileHrefs(t, body) {
			if u, _ := url.Parse(h); !strings.HasPrefix(u.Path, "/species/") {
				t.Errorf("a tile links to %q rather than a detail view", h)
			}
		}
	}
}

// TestPickingIsReversibleAndKeepsTheFilter covers adding, removing, and the
// interaction with the family filter.
//
// Verifies: product.pickFromIndex. Criterion: acceptance.pickModeSurvivesNavigation.
func TestPickingIsReversibleAndKeepsTheFilter(t *testing.T) {
	s := newTestServer(t)

	// Add, then remove, returns the original set in the original order.
	start := "/?species=amecro&species=annhum"
	_, body := get(t, s, start)
	var addURL string
	for _, h := range pickControls(body) {
		if q, _ := url.Parse(h); len(q.Query()["species"]) == 3 {
			addURL = h
			break
		}
	}
	if addURL == "" {
		t.Fatal("no pick control adds a third bird")
	}
	_, added := get(t, s, addURL)
	var removeURL string
	for _, h := range pickControls(added) {
		if q, _ := url.Parse(h); len(q.Query()["species"]) == 2 {
			if u, _ := url.Parse(addURL); slices.Equal(q.Query()["species"],
				u.Query()["species"][:2]) {
				removeURL = h
			}
		}
	}
	if removeURL == "" {
		t.Fatal("after adding, no control removes it again")
	}
	back, _ := url.Parse(removeURL)
	if got, want := back.Query()["species"],
		[]string{"amecro", "annhum"}; !slices.Equal(got, want) {
		t.Errorf("add then remove gave %v, want the original %v in the original order", got, want)
	}

	// The filter survives a pick, and the bar names a pick the filter hides.
	_, filtered := get(t, s, "/?family=Trochilidae&species=greblu")
	for _, h := range pickControls(filtered) {
		u, _ := url.Parse(h)
		if got := u.Query()["family"]; !slices.Equal(got, []string{"Trochilidae"}) {
			t.Errorf("a pick link carries family=%v; picking must not drop the filter "+
				"the reader is picking from", got)
		}
	}
	if bar := pickBarOf(t, filtered); !strings.Contains(bar, "Great Blue Heron") {
		t.Error("the bar omits a pick the filter is hiding; a filter that appeared to " +
			"discard picks would be worse than one that refused to run")
	}
}

// TestPicksSurviveTheDetailRoundTrip is the flow the whole picker rests on:
// open a bird to look at it, go back, pick something.
//
// Two picks, not one: with a single pick, "preserved the set" and "rebuilt a
// set of one" are indistinguishable. Recorded as a trap in acceptance.pickModeSurvivesNavigation.
//
// Verifies: product.pickFromIndex. Criterion: acceptance.pickModeSurvivesNavigation.
func TestPicksSurviveTheDetailRoundTrip(t *testing.T) {
	s := newTestServer(t)
	want := []string{"amecro", "annhum"}
	_, body := get(t, s, "/?species=amecro&species=annhum")

	_, detail := get(t, s, tileHrefs(t, body)[0])
	back := html.UnescapeString(mustMatch(t, detail,
		`<a class="backlink" href="([^"]+)"`, "the back link"))
	u, err := url.Parse(back)
	if err != nil {
		t.Fatalf("unparseable back link %q: %v", back, err)
	}
	if got := u.Query()["species"]; !slices.Equal(got, want) {
		t.Errorf("after index -> detail -> back the picks are %v, want %v", got, want)
	}
	_, returned := get(t, s, back)
	if pickBarOf(t, returned) == "" {
		t.Error("returning from a detail view left pick mode")
	}
}

// TestOnlyClearLeavesPickMode asserts that the picks are sticky against every
// control except the one meant to drop them.
//
// Verifies: product.pickFromIndex. Criterion: acceptance.pickModeSurvivesNavigation.
func TestOnlyClearLeavesPickMode(t *testing.T) {
	s := newTestServer(t)
	_, body := get(t, s, "/?species=annhum&units=metric")

	clear := html.UnescapeString(mustMatch(t, body,
		`<a class="btn pickclear" href="([^"]+)"`, "the Clear link"))
	cu, _ := url.Parse(clear)
	if _, present := cu.Query()["species"]; present {
		t.Errorf("Clear leaves a species parameter behind (%q), so it does not leave pick mode",
			clear)
	}

	// Every link EXCEPT the ones whose job is to change the picks must keep
	// them. The pick controls and the bar's chips are exempt by definition --
	// they are how a reader edits the set — so they are gathered by class and
	// excluded, leaving the ordinary controls: units, scale, ordering, step,
	// the family chips and the back-to-index links.
	edits := map[string]bool{clear: true}
	for _, h := range pickControls(body) {
		edits[h] = true
	}
	for _, m := range regexp.MustCompile(`<a class="pickchip" href="([^"]+)"`).
		FindAllStringSubmatch(body, -1) {
		edits[html.UnescapeString(m[1])] = true
	}
	checked := 0
	for _, m := range regexp.MustCompile(`href="(/\?[^"]*)"`).FindAllStringSubmatch(body, -1) {
		h := html.UnescapeString(m[1])
		if edits[h] {
			continue
		}
		checked++
		u, _ := url.Parse(h)
		if got := u.Query()["species"]; !slices.Equal(got, []string{"annhum"}) {
			t.Errorf("control link %q carries species=%v; only Clear may drop the picks", h, got)
		}
	}
	// Guard the guard: if the exclusions ever swallow everything, this test
	// passes while checking nothing.
	if checked == 0 {
		t.Fatal("no ordinary control links were checked; the exclusions have eaten the test")
	}
}

// TestUnrecognisedPicksAreDroppedAndNamed asserts the rot handling of product.pickFromIndex
// and decisions.compareUrlContract.
//
// The set mixes a good id with a dead one on purpose: with only a dead id,
// "dropped it and carried on" and "refused the page" look the same.
//
// Verifies: product.pickFromIndex. Criterion: acceptance.pickModeSurvivesNavigation.
func TestUnrecognisedPicksAreDroppedAndNamed(t *testing.T) {
	s := newTestServer(t)
	code, body := get(t, s, "/?species=annhum&species=grus_canadensis")
	if code != http.StatusOK {
		t.Fatalf("a set with one dead id gave %d; it should render what it recognises", code)
	}

	bar := pickBarOf(t, body)
	if !strings.Contains(bar, "Anna") {
		t.Error("the surviving pick is not shown")
	}
	if !strings.Contains(bar, "grus_canadensis") {
		t.Error("the dropped id is not named; a reader cannot tell what the link lost")
	}

	// And it is gone from every link, so the set repairs itself on the next
	// click. This is the half that silently failed first time: the notice
	// rendered while 970 links still carried the dead id.
	for _, m := range regexp.MustCompile(`href="([^"]+)"`).FindAllStringSubmatch(body, -1) {
		if strings.Contains(html.UnescapeString(m[1]), "grus_canadensis") {
			t.Errorf("link %q still carries the dead id; the set does not self-heal", m[1])
		}
	}
}

// detailAction returns the detail page's pick control: its label and its href.
func detailAction(t *testing.T, body string) (label, href string) {
	t.Helper()
	m := regexp.MustCompile(`<p class="pickaction"><a class="[^"]*" href="([^"]+)">([^<]+)</a>`).
		FindStringSubmatch(body)
	if m == nil {
		t.Fatal("the detail page has no pick control")
	}
	return html.UnescapeString(m[2]), html.UnescapeString(m[1])
}

// TestDetailPagePickStates asserts the three states of product.pickFromIndex, and that the
// back link says which mode it is returning to.
//
// Three and not two: offering "add" for a bird already in the set would be a
// click that silently does nothing, because the comparison view deduplicates.
//
// Verifies: product.pickFromIndex. Criterion: acceptance.pickModeSurvivesNavigation.
func TestDetailPagePickStates(t *testing.T) {
	s := newTestServer(t)

	// Browsing: the entry point, which seeds a new comparison with this bird.
	label, href := detailAction(t, mustGet(t, s, "/species/annhum"))
	if !strings.Contains(label, "Compare with another bird") {
		t.Errorf("browsing-mode action is %q, want the entry point", label)
	}
	if got := mustQuery(t, href)["species"]; !slices.Equal(got, []string{"annhum"}) {
		t.Errorf("the entry point seeds species=%v, want just this bird", got)
	}

	// Pick mode, this bird not in the set: add it, keeping the others.
	body := mustGet(t, s, "/species/annhum?species=amecro")
	label, href = detailAction(t, body)
	if !strings.Contains(label, "Pick for comparison") {
		t.Errorf("pick-mode action is %q, want an add", label)
	}
	if got, want := mustQuery(t, href)["species"],
		[]string{"amecro", "annhum"}; !slices.Equal(got, want) {
		t.Errorf("adding gave species=%v, want %v — the existing picks must survive", got, want)
	}

	// Pick mode, already in the set: remove, not a second add.
	body = mustGet(t, s, "/species/annhum?species=amecro&species=annhum")
	label, href = detailAction(t, body)
	if !strings.Contains(label, "remove") {
		t.Errorf("a bird already picked offers %q; adding it again would silently do nothing, "+
			"because the comparison view deduplicates", label)
	}
	if got, want := mustQuery(t, href)["species"],
		[]string{"amecro"}; !slices.Equal(got, want) {
		t.Errorf("removing gave species=%v, want %v", got, want)
	}

	// The back link names the mode it returns to.
	for _, c := range []struct{ url, want string }{
		{"/species/annhum", "back to all birds"},
		{"/species/annhum?species=amecro", "back to picking birds"},
	} {
		got := mustMatch(t, mustGet(t, s, c.url), `class="backlink"[^>]*>(?s)(.*?)</a>`, "the back link")
		if !strings.Contains(strings.Join(strings.Fields(got), " "), c.want) {
			t.Errorf("%s: back link reads %q, want %q", c.url, strings.TrimSpace(got), c.want)
		}
	}

	// The bar rides along, so Compare is reachable without going back first.
	if pickBarOf(t, mustGet(t, s, "/species/annhum?species=amecro")) == "" {
		t.Error("no pick bar on the detail page in pick mode")
	}
	if pickBarOf(t, mustGet(t, s, "/species/annhum")) != "" {
		t.Error("the detail page shows a pick bar while browsing")
	}
}

func mustGet(t *testing.T, s *server, path string) string {
	t.Helper()
	code, body := get(t, s, path)
	if code != http.StatusOK {
		t.Fatalf("GET %s: %d", path, code)
	}
	return body
}

func mustQuery(t *testing.T, href string) url.Values {
	t.Helper()
	u, err := url.Parse(href)
	if err != nil {
		t.Fatalf("unparseable href %q: %v", href, err)
	}
	return u.Query()
}

// TestCompareBirdsToggle asserts the Controls entry point of product.pickFromIndex: a way into
// pick mode for a reader who has no particular bird in mind yet.
//
// The empty state it enters is only reachable because a bare `species=` holds
// the mode open, so this control and that decision stand or fall together.
//
// Verifies: product.pickFromIndex. Criterion: acceptance.pickModeSurvivesNavigation.
func TestCompareBirdsToggle(t *testing.T) {
	s := newTestServer(t)
	find := func(body string) (class, href string) {
		t.Helper()
		// Found by class, not by its text: the button now states the mode
		// ("On"/"Off") and the control set's label carries the name, so
		// matching on "Compare birds" would find the label instead.
		m := regexp.MustCompile(`(?s)<a class="(btn cmptoggle[^"]*)" href="([^"]+)"[^>]*>\s*([A-Za-z]+)</a>`).
			FindStringSubmatch(body)
		if m == nil {
			t.Fatal("no Compare birds toggle in the controls row")
		}
		// The control needs a label of its own, or it reads as one more
		// Section step option -- which is what it looked like at first.
		if !strings.Contains(body, `<span class="lbl">Compare birds</span>`) {
			t.Error("the toggle has no control-set label, so it reads as part of the " +
				"control before it")
		}
		if want := map[bool]string{true: "On", false: "Off"}[strings.Contains(m[1], " on")]; m[3] != want {
			t.Errorf("the toggle reads %q but its class is %q; the label must state the mode",
				m[3], m[1])
		}
		return m[1], html.UnescapeString(m[2])
	}

	// Off while browsing, and it turns the mode on with nothing picked.
	class, href := find(mustGet(t, s, "/?units=metric"))
	if strings.Contains(class, "on") {
		t.Errorf("the toggle reads as on while browsing (class %q)", class)
	}
	q := mustQuery(t, href)
	if v, present := q["species"]; !present || !slices.Equal(v, []string{""}) {
		t.Errorf("the toggle goes to %q; it should enter pick mode with a bare species=", href)
	}
	if q.Get("units") != "metric" {
		t.Errorf("the toggle dropped the display options: %q", href)
	}
	// Following it really does land in pick mode with an empty bar.
	entered := mustGet(t, s, href)
	if bar := pickBarOf(t, entered); !strings.Contains(bar, "No birds picked yet") {
		t.Errorf("following the toggle did not enter pick mode; bar was %q", bar)
	}

	// On once picking, whether or not anything is picked, and it leaves by
	// exactly the link Clear uses -- same behavior, so the same URL.
	// units= matters: with no other parameter, Clear and a hardcoded "/?"
	// are the same string, and the comparison below cannot tell a shared link
	// from a coincidence.
	for _, u := range []string{"/?species=&units=metric", "/?species=annhum&units=metric"} {
		body := mustGet(t, s, u)
		class, href = find(body)
		if !strings.Contains(class, "on") {
			t.Errorf("%s: the toggle reads as off while picking (class %q)", u, class)
		}
		if _, present := mustQuery(t, href)["species"]; present {
			t.Errorf("%s: the toggle does not leave pick mode: %q", u, href)
		}
		clear := html.UnescapeString(mustMatch(t, body,
			`<a class="btn pickclear" href="([^"]+)"`, "the Clear link"))
		if href != clear {
			t.Errorf("%s: the toggle leaves via %q but Clear uses %q; they should be the "+
				"same link, not two that can drift", u, href, clear)
		}
	}
}

// TestPickLinksAnchorAtTheirTile asserts the fix for the jump that picking used
// to cause on the index.
//
// Every control is a link, so picking reloads a 1.7 MB page with 841 inline
// diagrams. The scroll script runs after that has painted, so the reader saw
// the top of the list and was then moved — a long way, since picking happens
// deep in the list. Navigating to the tile's own fragment makes the browser
// scroll during load instead, and the script stands down when a hash is
// present so the two do not fight.
//
// Verifies: product.togglesKeepPlace, product.pickFromIndex. Criterion: acceptance.pickModeSurvivesNavigation.
func TestPickLinksAnchorAtTheirTile(t *testing.T) {
	s := newTestServer(t)
	body := mustGet(t, s, "/?species=")

	ids := regexp.MustCompile(`<div class="tilewrap" id="([^"]+)"`).FindAllStringSubmatch(body, -1)
	picks := pickControls(body)
	if len(ids) == 0 || len(ids) != len(picks) {
		t.Fatalf("%d tile anchors for %d pick controls", len(ids), len(picks))
	}
	for i, h := range picks {
		want := "#" + ids[i][1]
		if !strings.HasSuffix(h, want) {
			t.Errorf("pick link %q does not end at its own tile %q; the browser cannot "+
				"restore the reading position during load", h, want)
		}
	}

	// The script must stand down when a hash is present, or it drags the
	// reader away from the anchor the browser just scrolled to.
	script := mustMatch(t, body, `(?s)(<script>.*</script>)`, "the page script")
	if !strings.Contains(script, "window.location.hash") {
		t.Error("the scroll script does not check for a hash, so it will override the " +
			"browser's own fragment scroll and reintroduce the jump")
	}
}

// TestEveryReportedMetricIsLabelled asserts product.labeledDetailDiagram's general rule: the diagram
// carries every metric the detail view reports, not a chosen subset.
//
// Written against the provenance table rather than a list of metric names, so
// it maintains itself. Mass was reported in the table and absent from the
// drawing for weeks without anything noticing, because every check named the
// metrics it expected and mass was not among them. Adding a row to the table
// now fails this until the drawing carries it too.
//
// Verifies: product.labeledDetailDiagram. Criterion: acceptance.segmentLabelsDontCollide.
func TestEveryReportedMetricIsLabelled(t *testing.T) {
	for _, units := range []string{"metric", "imperial"} {
		body := mustGet(t, newTestServer(t), "/species/greblu?units="+units)

		// Scoped to the measurements table. The detail page now carries a
		// second provenance table for categories (product.categoryProvenance), whose rows are
		// not metrics and are not drawn in the diagram.
		metrics := body[strings.Index(body, `<table class="prov metrics">`):]
		metrics = metrics[:strings.Index(metrics, "</table>")]
		rows := regexp.MustCompile(
			`<th>([^<]+)</th>\s*<td class="v">([^<]+)</td>`).FindAllStringSubmatch(metrics, -1)
		if len(rows) < 9 {
			t.Fatalf("%s: only %d rows in the provenance table; the fixture cannot show that "+
				"EVERY reported metric is labelled", units, len(rows))
		}

		// Every label the diagram draws, whitespace-normalised and stripped of
		// the estimated marker, which the table does not carry.
		drawn := map[string]bool{}
		for _, m := range regexp.MustCompile(
			`<text class="dlabel[^"]*"[^>]*>([^<]+)</text>`).FindAllStringSubmatch(body, -1) {
			s := strings.Join(strings.Fields(html.UnescapeString(m[1])), " ")
			drawn[strings.TrimSuffix(s, " ~")] = true
		}

		for _, r := range rows {
			want := strings.Join(strings.Fields(
				html.UnescapeString(r[1])+" "+html.UnescapeString(r[2])), " ")
			if !drawn[want] {
				t.Errorf("%s: the table reports %q but the diagram does not label it; "+
					"product.labeledDetailDiagram requires every reported metric to appear in the drawing", units, want)
			}
		}
	}
}

// massCubeOf returns the mass square's attributes from a rendered diagram.
func massCubeOf(t *testing.T, body string) (class string, x, y, side float64, ok bool) {
	t.Helper()
	g := regexp.MustCompile(`(?s)<g class="(masscube[^"]*)">.*?</g>`).FindStringSubmatch(body)
	if g == nil {
		return "", 0, 0, 0, false
	}
	m := regexp.MustCompile(
		`(?s)<rect class="mcface" x="([0-9.]+)" y="([0-9.]+)"\s*width="([0-9.]+)"`).
		FindStringSubmatch(g[0])
	if m == nil {
		return "", 0, 0, 0, false
	}
	m = []string{m[0], g[1], m[1], m[2], m[3]}
	f := func(s string) float64 { v, _ := strconv.ParseFloat(s, 64); return v }
	return m[1], f(m[2]), f(m[3]), f(m[4]), true
}

// TestMassCubeIsWaterToScale asserts the mass square of product.labeledDetailDiagram and decisions.massAsWaterCube.
//
// The square's whole claim is that it is a real volume at the diagram's own
// scale, not a glyph on a scale of its own. So the assertion is arithmetic:
// its side must be the cube root of the mass in grams, converted cm to mm, at
// the diagram's px/mm — derived from the definition of the gram, never read
// off a run.
//
// Verifies: product.labeledDetailDiagram. Criterion: acceptance.massDrawnAsWaterCube.
func TestMassCubeIsWaterToScale(t *testing.T) {
	ds := testDataset()

	for i := range ds.Species {
		sp := &ds.Species[i]
		lt := labelledTile(sp, view.Box{W: tileBox.W * 2.6, H: tileBox.H * 1.9}, units{}, dataset.LocaleNA)

		// One gram of water is one cubic centimetre, so the cube of water
		// weighing massG grams has side cbrt(massG) cm = 10*cbrt(massG) mm.
		want := 10 * math.Cbrt(sp.Mass.Value) * lt.Geom.PxPerMm
		if math.Abs(lt.MassCube.Side-want) > 0.05 {
			t.Errorf("%s: mass square side %.2f px, want %.2f — it must be the cube of water "+
				"that weighs the same, at the diagram's own scale",
				sp.Name(dataset.LocaleNA), lt.MassCube.Side, want)
		}
		// The face's base sits on the floor of the drawing area, at the same
		// place in EVERY diagram. That fixed corner is the whole point of the
		// cube form: the projection grows up and right from it, so two cubes
		// can be compared by eye. A base that moved with the bird would
		// destroy that without changing a single size.
		drawH := lt.Box.H - math.Max(footStrip(len(lt.FootLabels)),
			massLabelGap+footDescender)
		if base, want := lt.MassCube.Y+lt.MassCube.Side, drawH-massCubeMargin; math.Abs(base-want) > 0.01 {
			t.Errorf("%s: the cube's face stands at y=%.2f, want %.2f in every diagram",
				sp.Name(dataset.LocaleNA), base, want)
		}
		if lt.MassCube.X != massCubeMargin {
			t.Errorf("%s: the cube's face starts at x=%.2f, want %d in every diagram",
				sp.Name(dataset.LocaleNA), lt.MassCube.X, massCubeMargin)
		}
		// The projection grows up and right, and must stay in the drawing and
		// clear of the bird.
		if topY := lt.MassCube.Y - lt.MassCube.Depth; topY < 0 {
			t.Errorf("%s: the cube's top reaches y=%.1f, above the drawing", sp.Name(dataset.LocaleNA), topY)
		}
		if right := lt.MassCube.X + lt.MassCube.Side + lt.MassCube.Depth; right >= lt.Geom.CX {
			t.Errorf("%s: the cube reaches x=%.1f, past the bird's axis at %.1f",
				sp.Name(dataset.LocaleNA), right, lt.Geom.CX)
		}
		if lt.MassCube.Depth <= 0 {
			t.Errorf("%s: the cube has no depth, so it is a square", sp.Name(dataset.LocaleNA))
		}
		// Its label must be centred under the face without being clipped --
		// a Scissor-tailed Flycatcher's 29px cube once rendered "ass 1.4 oz".
		if left := lt.MassCube.LabelX - float64(len(lt.MassCube.Label))*massLabelCharWidth/2; left < -0.01 {
			t.Errorf("%s: the mass label %q starts at x=%.1f and is clipped",
				sp.Name(dataset.LocaleNA), lt.MassCube.Label, left)
		}
		// Measured mass draws solid, near-match dashed (product.measuredVersusModeled).
		if want := sp.Mass.Prov.Kind != dataset.Measured; lt.MassCube.Est != want {
			t.Errorf("%s: mass square Est=%v but provenance is %q",
				sp.Name(dataset.LocaleNA), lt.MassCube.Est, sp.Mass.Prov.Kind)
		}
	}
}

// TestMassCubeIsRenderedAndExplained asserts that the square reaches the page
// and that the page says what it is.
//
// An unexplained shape beside a bird reads as part of the bird, so the key
// naming it is not documentation of the feature — it is the feature.
//
// Verifies: product.labeledDetailDiagram, product.pageExplainsItself. Criterion: acceptance.massDrawnAsWaterCube.
func TestMassCubeIsRenderedAndExplained(t *testing.T) {
	s := newTestServer(t)
	for _, u := range []string{
		"/species/greblu",
		"/compare?species=greblu&species=annhum",
	} {
		body := mustGet(t, s, u)
		class, x, _, side, ok := massCubeOf(t, body)
		if !ok {
			t.Errorf("%s: no mass square in the diagram", u)
			continue
		}
		if side <= 0 || x < 0 {
			t.Errorf("%s: mass square has side %.1f at x=%.1f", u, side, x)
		}
		if !strings.Contains(class, "masscube") {
			t.Errorf("%s: mass square class is %q", u, class)
		}
		// Exactly one key per page however many birds, and it names the square.
		if n := strings.Count(body, `class="note diagramkey"`); n != 1 {
			t.Errorf("%s: %d diagram keys, want exactly 1", u, n)
		}
		if !strings.Contains(body, "drawn as water") {
			t.Errorf("%s: the diagram key does not say what the square is, so it reads as "+
				"part of the bird", u)
		}
	}

	// On index tiles too, since decisions.massCubeOnTiles -- but unlabelled. A tile carries no
	// labels at all and names the mass in its footer, so the cube appears
	// with no text of its own.
	_, index := get(t, s, "/")
	tiles := tileArea(t, index)
	cubes := strings.Count(tiles, `<g class="masscube`)
	if n := strings.Count(tiles, `<a class="tile"`); cubes != n {
		t.Errorf("%d mass cubes across %d tiles; every tile should carry one", cubes, n)
	}
	if strings.Contains(tiles, `class="dlabel`) {
		t.Error("an index tile carries a diagram label; the tile has none, and the mass is " +
			"named in its footer instead")
	}
}

// sortControls returns the comparison view's Sort by options: label, href, on.
func sortControls(t *testing.T, body string) []struct {
	Label, Href string
	On          bool
} {
	t.Helper()
	block := regexp.MustCompile(`(?s)<div class="row cmpsort">.*?</div>`).FindString(body)
	var out []struct {
		Label, Href string
		On          bool
	}
	for _, m := range regexp.MustCompile(
		`<a class="(btn[^"]*)" href="([^"]+)">([^<]+)</a>`).FindAllStringSubmatch(block, -1) {
		out = append(out, struct {
			Label, Href string
			On          bool
		}{m[3], html.UnescapeString(m[2]), strings.Contains(m[1], " on")})
	}
	return out
}

// TestCompareSortReordersTheURL asserts the Sort by control of product.comparisonOneScale.
//
// The design decision it pins: sorting rewrites the species order in the URL
// rather than adding a sort parameter, so the page still draws them in the
// order given and a sorted comparison is an ordinary comparison URL. If a
// `sort=` parameter ever appears instead, the panels and the URL can disagree
// and this fails.
//
// Verifies: product.comparisonOneScale. Criterion: acceptance.comparisonSortRewritesUrl.
func TestCompareSortReordersTheURL(t *testing.T) {
	s := newTestServer(t)
	// Deliberately in no metric's order, so no option starts out on.
	// The murre must be in the set: it is the only fixture species for which
	// the three metrics disagree, so without it every sort yields the same
	// order and this test cannot tell one metric from another.
	const start = "/compare?units=metric&species=greblu&species=annhum" +
		"&species=amecro&species=thimur"
	body := mustGet(t, s, start)

	opts := sortControls(t, body)
	if len(opts) != 3 {
		t.Fatalf("%d sort options, want length, wingspan and mass", len(opts))
	}
	ds := testDataset()
	byID := map[string]*dataset.Species{}
	for i := range ds.Species {
		byID[ds.Species[i].ID] = &ds.Species[i]
	}

	for _, o := range opts {
		m := map[string]view.Metric{
			"Length": view.ByLength, "Wingspan": view.ByWingspan, "Mass": view.ByMass,
		}[o.Label]
		if m == "" {
			t.Errorf("unexpected sort option %q", o.Label)
			continue
		}
		u := mustQuery(t, o.Href)
		if u.Get("units") != "metric" {
			t.Errorf("%s: the sort link dropped the display options: %q", o.Label, o.Href)
		}
		// No sort parameter: the order IS the species order (product.comparisonOneScale).
		if u.Get("sort") != "" || u.Get("by") != "" {
			t.Errorf("%s: the sort link carries an ordering parameter (%q); the order must "+
				"live in the species list and nowhere else", o.Label, o.Href)
		}
		// The link names the same birds, ascending by that metric.
		got := u["species"]
		if len(got) != 4 {
			t.Errorf("%s: the sort link names %d species, want 4", o.Label, len(got))
			continue
		}
		for i := 1; i < len(got); i++ {
			prev, cur := byID[got[i-1]], byID[got[i]]
			if prev == nil || cur == nil {
				t.Fatalf("%s: sort link names an unknown species: %v", o.Label, got)
			}
			if m.Value(cur) < m.Value(prev) {
				t.Errorf("%s: %s (%.1f) is ordered before %s (%.1f); the sort is ascending",
					o.Label, cur.Name(dataset.LocaleNA), m.Value(cur), prev.Name(dataset.LocaleNA), m.Value(prev))
			}
		}

		// Following it puts the panels in that order, and lights that option.
		after := mustGet(t, s, o.Href)
		var drawn []string
		for _, p := range drawnSpans(t, after) {
			drawn = append(drawn, findSpeciesByName(t, p.Name).ID)
		}
		if !slices.Equal(drawn, got) {
			t.Errorf("%s: the URL names %v but the page draws %v", o.Label, got, drawn)
		}
		for _, a := range sortControls(t, after) {
			if a.Label == o.Label && !a.On {
				t.Errorf("%s: after sorting by it, the option does not read as on", o.Label)
			}
		}
	}

	// Guard the guard: if every metric yields the same order, this test cannot
	// distinguish them and a control that ignored its metric would pass.
	if slices.Equal(mustQuery(t, opts[0].Href)["species"], mustQuery(t, opts[2].Href)["species"]) {
		t.Fatal("length and mass sort this set identically, so the fixture cannot tell the " +
			"three metrics apart; it needs a species that is heavy for its length")
	}

	// Nothing starts out on, since the fixture order is no metric's order --
	// otherwise "reads as on after sorting" would pass without sorting.
	for _, o := range opts {
		if o.On {
			t.Errorf("%s reads as on before any sort; the fixture order was meant to match "+
				"no metric, so this test cannot tell sorting from doing nothing", o.Label)
		}
	}

	// Below two birds there is nothing to sort.
	for _, u := range []string{"/compare?species=annhum", "/compare?species="} {
		if n := len(sortControls(t, mustGet(t, s, u))); n != 0 {
			t.Errorf("%s offers %d sort options; there is nothing to sort", u, n)
		}
	}
}

// TestNameLocaleTogglesNamesAndNothingElse asserts product.englishNameChoice's central claim: the
// reader picks a vocabulary and the identifiers stay put.
//
// The fixture gives every species a world name that differs from its North
// American one (see worldName), so a toggle that serves the same string both
// ways fails here rather than passing on data that cannot tell them apart.
//
// Verifies: product.englishNameChoice. Criterion: acceptance.nameLocaleToggle.
func TestNameLocaleTogglesNamesAndNothingElse(t *testing.T) {
	s := newTestServer(t)
	sp := findSpecies(t, "amerob")
	na, world := sp.Name(dataset.LocaleNA), sp.Name(dataset.LocaleWorld)
	if na == world {
		t.Fatalf("fixture gives %q the same name in both locales; the check cannot bite", na)
	}

	for _, c := range []struct{ query, want, notWant string }{
		{"", na, world},
		{"?names=na", na, world},
		{"?names=world", world, na},
	} {
		_, body := get(t, s, "/species/"+sp.ID+c.query)
		if !strings.Contains(body, c.want) {
			t.Errorf("/species/%s%s omits %q", sp.ID, c.query, c.want)
		}
		if strings.Contains(body, c.notWant) {
			t.Errorf("/species/%s%s shows %q, the other locale's name", sp.ID, c.query, c.notWant)
		}
		// The identifier must not move with the vocabulary. This is the
		// property an alpha code could not have had: it was derived from the
		// English name, so it would have changed with it.
		if !strings.Contains(body, sp.SpeciesCode) {
			t.Errorf("/species/%s%s does not show the code %q", sp.ID, c.query, sp.SpeciesCode)
		}

	}
}

// TestNameLocaleControlOffersBothAndMarksTheActive checks the control itself,
// not just the effect. A toggle that renders and does nothing looks identical
// to one that works until you follow it.
//
// Verifies: product.englishNameChoice. Criterion: acceptance.nameLocaleToggle.
func TestNameLocaleControlOffersBothAndMarksTheActive(t *testing.T) {
	s := newTestServer(t)
	for _, c := range []struct{ query, onHref string }{
		{"/", "names=na"},
		{"/?names=world", "names=world"},
	} {
		_, body := get(t, s, c.query)
		for _, want := range []string{`"names" "na"`, `"names" "world"`} {
			_ = want // the template renders these through the query helper
		}
		if !strings.Contains(body, "names=na") || !strings.Contains(body, "names=world") {
			t.Errorf("%s: the Names control does not offer both locales", c.query)
		}
		// The active one is marked. Located by finding the on-state button
		// whose href carries this locale.
		re := regexp.MustCompile(`<a class="btn on"[^>]*href="([^"]*)"`)
		var onHrefs []string
		for _, m := range re.FindAllStringSubmatch(body, -1) {
			onHrefs = append(onHrefs, html.UnescapeString(m[1]))
		}
		found := false
		for _, h := range onHrefs {
			if strings.Contains(h, c.onHref) {
				found = true
			}
		}
		if !found {
			t.Errorf("%s: no on-state control carries %q; on-states were %v", c.query, c.onHref, onHrefs)
		}
	}
}

// TestNameFallsBackToWorldWhenNoNorthAmericanName covers the five species with
// no eBird/Clements concept: they must render a bird, not a blank (product.englishNameChoice).
//
// Verifies: product.englishNameChoice. Criterion: acceptance.nameLocaleToggle.
func TestNameFallsBackToWorldWhenNoNorthAmericanName(t *testing.T) {
	sp := dataset.Species{
		ID:    "nonam",
		Names: map[dataset.Locale]string{dataset.LocaleWorld: "Only A World Name"},
	}
	for _, loc := range []dataset.Locale{dataset.LocaleNA, dataset.LocaleWorld} {
		if got := sp.Name(loc); got != "Only A World Name" {
			t.Errorf("locale %q: got %q, want the world name as a fallback", loc, got)
		}
	}
}

// TestEmptyStateNamesTheFiltersThatEmptiedIt asserts the half of product.helpfulEmptyState that
// says WHICH filter emptied the page, not merely that one did.
//
// The two empty states must stay distinct. "Nothing selected" is not a failed
// match -- the reader turned everything off -- and saying "no birds match"
// there blames the data for their click. acceptance.emptyStateGuides already pins that direction;
// this pins the other.
//
// Verifies: product.helpfulEmptyState. Criterion: acceptance.emptyStateGuides.
func TestEmptyStateNamesTheFiltersThatEmptiedIt(t *testing.T) {
	s := newTestServer(t)

	// A selection that matches nothing: a real family, plus a metric window
	// no bird falls in. The fixture has no way to empty a family selection on
	// its own, which is why this needs the second filter to exist at all --
	// see the note below.
	_, body := get(t, s, "/?family=Ardeidae&family=Turdidae")
	if strings.Contains(body, "No birds match") {
		t.Fatal("a selection with species in it reported no matches")
	}

	// Nothing selected: the other message, and emphatically not this one.
	_, none := get(t, s, "/?family=")
	if !strings.Contains(none, "Select birds to display") {
		t.Error("an emptied selection does not say how to get birds back")
	}
	if strings.Contains(none, "No birds match") {
		t.Error("an emptied selection reports a failed match; nothing failed, the reader cleared it")
	}

	// The description itself, which is the new part. Built directly because
	// the fixture cannot yet reach an empty intersection through the UI: one
	// filter cannot contradict itself, and the region filter of product.regionAndMonthFilter does not
	// exist. When it does, this should become a rendering test.
	got := describeFilters(s.fams, options{HasSelection: true,
		Families: map[string]bool{"Ardeidae": true, "Turdidae": true}})
	if len(got) != 1 || got[0].Label != "Family" {
		t.Fatalf("describeFilters gave %+v, want one Family entry", got)
	}
	for _, want := range []string{"Ardeidae", "Turdidae"} {
		if !strings.Contains(got[0].Value, want) {
			t.Errorf("the description %q does not name %q", got[0].Value, want)
		}
	}

	// A filter that is not narrowing anything is not described. Naming it
	// would present the arrival state as a choice the reader made.
	everything := map[string]bool{}
	for _, f := range s.fams {
		everything[f.Name] = true
	}
	if d := describeFilters(s.fams, options{HasSelection: true, Families: everything}); d != nil {
		t.Errorf("every family selected is described as a filter: %+v", d)
	}
	if d := describeFilters(s.fams, options{}); d != nil {
		t.Errorf("no selection at all is described as a filter: %+v", d)
	}
}

// TestEveryControlIsDocumented asserts product.pageExplainsItself's rule: a control is documented
// if and only if it exists.
//
// Written because the requirement used to ENUMERATE the controls it covered,
// and the enumeration had gone stale twice over -- Compare birds and Names
// were both live and both undocumented, while the requirement, the panel and
// the controls each looked complete in isolation. A list cannot notice what it
// omits, so this compares the two sets instead.
//
// Verifies: product.pageExplainsItself. Criterion: acceptance.everyControlDocumented.
func TestEveryControlIsDocumented(t *testing.T) {
	_, body := get(t, newTestServer(t), "/")

	// Locate before slicing. A missing marker makes strings.Index return -1,
	// and body[-1:] panics before any assertion below can run -- so a broken
	// selector would report a crash rather than "the selector is broken",
	// which is the same confusion acceptance.familyListWraps's parser once produced.
	// Every panel documents its own controls (product.pageExplainsItself), so both sets are
	// gathered across the whole page rather than from one box.
	if !strings.Contains(body, `class="ctrlhelp"`) {
		t.Fatal("no panel carries a documentation block; the selector is broken, not the page")
	}
	var help string
	for _, part := range strings.Split(body, `class="ctrlhelp"`)[1:] {
		end := strings.Index(part, "</dl>")
		if end < 0 {
			t.Fatal("a documentation block is never closed; the selector is broken, not the page")
		}
		help += part[:end]
	}
	// An entry marked `ctrl note` explains something that is not a control --
	// what a zone is, what pelagic means. Those are allowed to have no
	// matching label, and are dropped before the sets are compared. Entries
	// not so marked must correspond to a control, which is the half of this
	// check that catches documentation for a control that was removed.
	var docs []string
	for _, part := range strings.Split(help, `<div class="ctrl`)[1:] {
		if strings.HasPrefix(part, " note") {
			continue
		}
		docs = append(docs, part)
	}
	help = strings.Join(docs, "")
	documented := map[string]bool{}
	for _, m := range regexp.MustCompile(`<dt>([^<]+)</dt>`).FindAllStringSubmatch(help, -1) {
		documented[strings.TrimSpace(html.UnescapeString(m[1]))] = true
	}

	// Control labels live in the rows, which sit below each help block.
	rows := body
	labels := map[string]bool{}
	for _, m := range regexp.MustCompile(`<span class="lbl">([^<]+)</span>`).
		FindAllStringSubmatch(rows, -1) {
		labels[strings.TrimSpace(html.UnescapeString(m[1]))] = true
	}

	// Both sets must be non-empty, or two broken selectors compare equal and
	// this passes while checking nothing.
	if len(documented) == 0 {
		t.Fatal("no <dt> terms found; the documentation selector is broken, not the page")
	}
	if len(labels) == 0 {
		t.Fatal("no control labels found; the control selector is broken, not the page")
	}
	if !labels["Scale"] {
		t.Fatalf("the label set does not contain a control known to exist; it is %v", keysOf(labels))
	}

	for l := range labels {
		if !documented[l] {
			t.Errorf("the %q control is documented in no panel (product.pageExplainsItself)", l)
		}
	}
	for d := range documented {
		if !labels[d] {
			t.Errorf("the documentation defines %q, which is not a control on the page", d)
		}
	}
}

// box returns one collapsible panel of the index by its id. The controls were
// one panel and are now three (product.collapsibleControls), so a test that counts chips or order
// headers has to say WHICH selector it means -- the family and region
// selectors share markup and CSS deliberately, so a page-wide count sees both.
func box(t *testing.T, body, id string) string {
	t.Helper()
	open := `<details class="controls" id="` + id + `"`
	i := strings.Index(body, open)
	if i < 0 {
		t.Fatalf("no panel with id %q on the page", id)
	}
	rest := body[i:]
	j := strings.Index(rest, "</details>")
	if j < 0 {
		t.Fatalf("panel %q is never closed", id)
	}
	return rest[:j]
}

// grouping narrows a panel to one of its grouping blocks, by heading.
//
// Both selectors now render every chip twice, once per grouping (product.regionsGroupedTwoWays,
// product.familiesByOrder), so a count taken over a whole panel counts each thing twice and an
// assertion about "the groups" is really an assertion about two groupings at
// once. Scoping is what keeps those counts meaningful.
func grouping(t *testing.T, panel, name string) string {
	t.Helper()
	open := `<div class="ghdr">` + name + `<em>`
	i := strings.Index(panel, open)
	if i < 0 {
		t.Fatalf("no grouping headed %q in this panel", name)
	}
	rest := panel[i+len(open):]
	if k := strings.Index(rest, `<div class="ghdr">`); k >= 0 {
		return rest[:k] // stop at the next grouping
	}
	return rest
}

// TestRegionFilter asserts product.regionAndMonthFilter and its composition with family (product.filtersCompose).
//
// Verifies: product.regionAndMonthFilter, product.filtersCompose. Criterion: acceptance.regionFilterBehaves.
func TestRegionFilter(t *testing.T) {
	s := newTestServer(t)

	// One region shows what is reported there and excludes the rest.
	_, ca := get(t, s, "/?region=CA")
	// Unescaped: the fixture's names carry apostrophes, which the template
	// writes as &#39; and a raw substring search then misses.
	tiles := html.UnescapeString(tileArea(t, ca))
	for _, want := range []string{"Anna's Hummingbird", "American Robin", "Black-necked Stilt"} {
		if !strings.Contains(tiles, want) {
			t.Errorf("California omits %s", want)
		}
	}
	for _, gone := range []string{"American Crow", "Thick-billed Murre"} {
		if strings.Contains(tiles, gone) {
			t.Errorf("California shows %s, which is not reported there", gone)
		}
	}

	// Absent means everywhere; explicitly empty means nowhere. The family
	// filter draws the same distinction, and conflating them is the bug it
	// had first.
	_, all := get(t, s, "/")
	_, none := get(t, s, "/?region=")
	if strings.Count(tileArea(t, all), "tilewrap") == 0 {
		t.Fatal("no species with no region parameter")
	}
	if n := strings.Count(tileArea(t, none), "tilewrap"); n != 0 {
		t.Errorf("an empty region= showed %d species; it must show none", n)
	}

	// Composition: region and family narrow together, and an empty
	// intersection names both filters rather than blaming one.
	_, both := get(t, s, "/?region=CA&family=Corvidae")
	if n := strings.Count(tileArea(t, both), "tilewrap"); n != 0 {
		t.Errorf("California plus Corvidae showed %d species; the crow is not in California", n)
	}
	if !strings.Contains(both, "No birds match") {
		t.Error("an empty intersection does not report a failed match")
	}
	for _, want := range []string{"Region", "California", "Family", "Corvidae"} {
		if !strings.Contains(both, want) {
			t.Errorf("the empty state does not name %q", want)
		}
	}

	// A seabird appears under its ocean as well as the coast it is seen from.
	_, sea := get(t, s, "/?region=NorthPacific")
	if !strings.Contains(html.UnescapeString(tileArea(t, sea)), "Thick-billed Murre") {
		t.Error("the Pacific shows no murre; a pelagic species has nowhere else to be")
	}
	_, ak := get(t, s, "/?region=AK")
	if !strings.Contains(html.UnescapeString(tileArea(t, ak)), "Thick-billed Murre") {
		t.Error("Alaska omits the murre, which is seen from shore there as well as at sea")
	}
}

// TestRegionSelectorIsGroupedWithPelagicLast checks the control, not the
// filtering: both groupings are offered, each groups the regions, and each
// ends with the sea (product.regionAndMonthFilter, product.regionsGroupedTwoWays).
//
// The per-grouping split matters here. Before it, this test read the zone
// toggles across the whole box and checked only the final one — which the
// hand-drawn grouping satisfied on its own, leaving the derived grouping's
// ordering unchecked.
//
// Verifies: product.regionAndMonthFilter, product.regionsGroupedTwoWays. Criterion: acceptance.regionFilterBehaves.
func TestRegionSelectorIsGroupedWithPelagicLast(t *testing.T) {
	_, body := get(t, newTestServer(t), "/")
	sel := box(t, body, "regions")

	// Split the box into its grouping blocks, so each is checked on its own.
	blocks := regexp.MustCompile(`<div class="ghdr">([^<]+)<em>([^<]*)</em>`).
		FindAllStringSubmatchIndex(sel, -1)
	if len(blocks) != 2 {
		t.Fatalf("found %d grouping blocks, want the two of product.regionsGroupedTwoWays", len(blocks))
	}
	for i, b := range blocks {
		name := html.UnescapeString(sel[b[2]:b[3]])
		note := html.UnescapeString(sel[b[4]:b[5]])
		end := len(sel)
		if i+1 < len(blocks) {
			end = blocks[i+1][0]
		}
		part := sel[b[0]:end]

		if note == "" {
			t.Errorf("grouping %q carries no note; a derived grouping and a "+
				"hand-drawn one must not look alike (product.regionsGroupedTwoWays)", name)
		}
		var zones []string
		for _, m := range regexp.MustCompile(`every region in ([^"]+)"`).
			FindAllStringSubmatch(part, -1) {
			zones = append(zones, html.UnescapeString(m[1]))
		}
		if len(zones) < 2 {
			t.Fatalf("grouping %q has %d zone toggles; it is not grouped", name, len(zones))
		}
		if last := zones[len(zones)-1]; last != "Pelagic" {
			t.Errorf("grouping %q ends with zone %q, want Pelagic — "+
				"the sea is not a place anyone lives", name, last)
		}
		// Every region is reachable from every grouping, each being a
		// partition: a region missing from one block is unreachable there.
		for _, want := range []string{`>CA<em>California</em>`, `>PA<em>Pennsylvania</em>`} {
			if !strings.Contains(part, want) {
				t.Errorf("grouping %q is missing a chip matching %q", name, want)
			}
		}
	}

	// The two must actually be the two product.regionsGroupedTwoWays names, not the same one twice,
	// and the hand-drawn one is rendered first: a reader arrives looking for
	// their own state, not for a cluster analysis.
	if !strings.Contains(sel, `<div class="ghdr">Common birds<em>`) ||
		!strings.Contains(sel, `<div class="ghdr">Geographic<em>`) {
		t.Error("the selector does not offer both a measured and a hand-drawn grouping")
	}
	if first := html.UnescapeString(sel[blocks[0][2]:blocks[0][3]]); first != "Geographic" {
		t.Errorf("the first grouping rendered is %q, want Geographic", first)
	}
}

// TestRegionURLsCarrySlugsNotGADMIDs checks what the reader actually gets:
// links built from postal codes and names, with the internal key nowhere in
// sight (decisions.regionUrlIsPostalCode).
//
// Verifies: product.regionAndMonthFilter. Criterion: acceptance.regionUrlRoundTrips.
func TestRegionURLsCarrySlugsNotGADMIDs(t *testing.T) {
	_, body := get(t, newTestServer(t), "/")
	sel := box(t, body, "regions")

	// No GADM id anywhere in a link. Checked against the real table rather
	// than a pattern, so a new id shape cannot slip past.
	for _, id := range append(region.LandIDs(), region.MarineIDs()...) {
		if strings.Contains(sel, "region="+id) {
			t.Errorf("the GADM id %q appears in a URL; it is a database key", id)
		}
	}
	for _, want := range []string{"region=CA", "region=GulfOfMexico", "region=JohnstonAtoll"} {
		if !strings.Contains(html.UnescapeString(sel), want) {
			t.Errorf("no link carries %q", want)
		}
	}

	// A coded region shows code then name; an uncoded one shows the name
	// alone, with no empty <em> for the stylesheet to put a separator before.
	if !strings.Contains(sel, `>CA<em>California</em>`) {
		t.Error("California's chip does not show its postal code beside its name")
	}
	for _, uncoded := range []string{"Johnston Atoll", "Gulf of Mexico", "North Pacific"} {
		if strings.Contains(sel, `<em>`+uncoded+`</em>`) {
			t.Errorf("%q renders inside <em>, which means an empty code precedes it "+
				"and the stylesheet will open the chip with a stray separator", uncoded)
		}
		if !strings.Contains(sel, ">"+uncoded+"</a>") {
			t.Errorf("%q does not render as a bare chip label", uncoded)
		}
	}

	// The ocean gloss survives, in the tooltip rather than the label.
	if !strings.Contains(sel, `title="North Pacific, off the west coast"`) {
		t.Error("the North Pacific chip lost the gloss saying which water it means")
	}
}

// TestTraitAndFamilyPanels checks the two controls after decisions.habitatDietSpeciesLevel: diet and
// habitat in their own box, assigned per species, and the taxonomic family
// list in its own.
//
// They are separate boxes because they answer separate questions. The
// family control sorts birds by what they are; diet and habitat sort them
// by how they live, and the two were briefly crowded into one panel called
// Species, which named neither well.
//
// Verifies: product.habitatAndDietFilters, product.familiesByOrder. Criterion: acceptance.habitatDietFilterPerSpecies.
func TestTraitAndFamilyPanels(t *testing.T) {
	_, body := get(t, newTestServer(t), "/")

	traits := box(t, body, "traits")
	fams := box(t, body, "families")

	blocks := func(panel string) []string {
		var out []string
		for _, m := range regexp.MustCompile(`<div class="ghdr">([^<]+)<em>([^<]*)</em>`).
			FindAllStringSubmatch(panel, -1) {
			out = append(out, html.UnescapeString(m[1]))
			if strings.TrimSpace(m[2]) == "" {
				t.Errorf("block %q carries no note", m[1])
			}
		}
		return out
	}
	if got := blocks(traits); !slices.Equal(got, []string{"Diet", "Habitat"}) {
		t.Errorf("the Diet and Habitat panel has blocks %v, want [Diet Habitat] — "+
			"the order matches the panel's own title", got)
	}
	if got := blocks(fams); !slices.Equal(got, []string{"Taxonomic order"}) {
		t.Errorf("the Families panel has blocks %v, want [Taxonomic order]", got)
	}
	if strings.Contains(body, "Birder groups") {
		t.Error("the retired birder grouping is still rendered")
	}
	// The family chips belong to the family panel and the trait rows to the
	// other; crossing over is how the two became one box before.
	if strings.Contains(traits, `class="ofams"`) {
		t.Error("family chips are in the Diet and Habitat panel")
	}
	if strings.Contains(fams, `class="tdesc"`) {
		t.Error("trait rows are in the Families panel")
	}

	// Each trait option carries AVONET's definition and a count, which is
	// the whole reason these are rows and not chips.
	for _, want := range []string{"Marsh and lake", "Open sea", "Hunts in water", "Drinks nectar"} {
		if !strings.Contains(traits, ">"+want+"</a>") {
			t.Errorf("the trait controls do not offer %q", want)
		}
	}
	rows := regexp.MustCompile(
		`class="tcount">(\d+)</span>\s*<span class="tdesc"><b>([^<]+)</b>: ([^<]*)`).
		FindAllStringSubmatch(traits, -1)
	if len(rows) == 0 {
		t.Fatal("no trait rows rendered at all")
	}
	if n := strings.Count(traits, `btn thdr`); n != len(rows) {
		t.Errorf("%d trait toggles but %d carry a count and a description", n, len(rows))
	}
	for _, m := range rows {
		if m[1] == "0" {
			t.Error("an option with no species is offered; clicking it returns an empty page")
		}
		if strings.TrimSpace(m[2]) == "" {
			t.Error("a trait option does not name the AVONET term its definition belongs to")
		}
		if strings.TrimSpace(m[3]) == "" {
			t.Error("a trait option carries no definition")
		}
	}
	if !strings.Contains(traits, `<b>Invertivore</b>: Species obtaining`) {
		t.Error(`the "Eats insects" option does not lead its definition with Invertivore`)
	}
}

// TestTileCarriesExactlyOneBadge pins product.tileContents's count. The family is the one
// grouping a tile names; the birder group, the order, and any grouping added
// later are navigation and stay in the controls.
//
// A count rather than a list of forbidden badges, because a list invites one
// more entry — and because every tile is one height, so a second badge costs
// a line on all of them (acceptance.tileHeaderNeverWraps).
//
// Verifies: product.tileContents. Criterion: acceptance.familyGroupingComplete.
func TestTileCarriesExactlyOneBadge(t *testing.T) {
	s := newTestServer(t)
	for _, path := range []string{"/", "/species/greblu", "/compare?species=greblu&species=annhum"} {
		code, body := get(t, s, path)
		if code != http.StatusOK {
			t.Fatalf("%s: HTTP %d", path, code)
		}
		badges := strings.Count(body, `class="badge"`)
		if badges == 0 {
			t.Errorf("%s: no family badge at all", path)
			continue
		}
		// One badge per bird shown. On the index that is one per tile; on
		// the detail and compare pages it is one per species panel.
		var want int
		switch path {
		case "/":
			want = strings.Count(body, `<a class="tile" href=`)
		case "/species/greblu":
			want = 1
		default:
			want = 2
		}
		if badges != want {
			t.Errorf("%s: %d birds shown but %d badges; product.tileContents allows the family badge and "+
				"no other, so a grouping has leaked onto the tile", path, want, badges)
		}
	}
}

// TestFiltersAreNamedOnTheCountLine asserts the half of product.helpfulEmptyState that applies
// while the list still has birds in it: what is narrowing the view is stated
// next to the count, not only once nothing matches.
//
// This is the rendering test acceptance.emptyStateGuides recorded as missing. It was unreachable
// when only the family filter existed -- one filter cannot contradict itself
// -- and the region filter of product.regionAndMonthFilter made the empty intersection reachable.
//
// Verifies: product.helpfulEmptyState. Criterion: acceptance.emptyStateGuides.
func TestFiltersAreNamedOnTheCountLine(t *testing.T) {
	s := newTestServer(t)

	countLine := func(body string) string {
		t.Helper()
		return mustMatch(t, body, `<p class="count">(.*?)</p>`, "the species count line")
	}

	// Arrival: nothing is narrowing anything, so nothing is named. Saying
	// "Family: all 6" here would present the default as a choice.
	if got := countLine(mustGet(t, s, "/")); strings.Contains(got, "fsum") {
		t.Errorf("the arrival page names a filter on the count line: %q", got)
	}

	// One filter set, and birds still matching: named anyway.
	full := mustGet(t, s, "/?family=Ardeidae")
	line := countLine(full)
	if !strings.Contains(line, "Ardeidae") {
		t.Errorf("a narrowed list does not name its filter on the count line: %q", line)
	}
	if !strings.Contains(full, "Great Blue Heron") {
		t.Fatal("the fixture no longer matches; this case is meant to have birds in it")
	}

	// Two filters whose intersection is empty: the count line still names
	// both, and the empty message does not repeat them.
	empty := mustGet(t, s, "/?family=Ardeidae&region=AK")
	line = countLine(empty)
	for _, want := range []string{"Ardeidae", "Alaska"} {
		if !strings.Contains(html.UnescapeString(line), want) {
			t.Errorf("the count line omits %q when the intersection is empty: %q", want, line)
		}
	}
	if !strings.Contains(empty, "No birds match") {
		t.Fatal("an empty intersection does not report a failed match")
	}
	if n := strings.Count(empty, `class="fsum"`); n != 2 {
		t.Errorf("the page carries %d filter chips, want 2 — the empty state is "+
			"repeating what the count line already said", n)
	}
}

// TestDetailListsEveryCategoryWithItsOrigin asserts product.categoryProvenance: the detail page
// names every group the bird is sorted into and says who decided.
//
// Verifies: product.categoryProvenance. Criterion: acceptance.categoryOriginShown.
func TestDetailListsEveryCategoryWithItsOrigin(t *testing.T) {
	s := newTestServer(t)
	// The robin, which has regions. The heron is deliberately in none, and
	// is checked below as its own case.
	body := mustGet(t, s, "/species/amerob")

	tbl := mustMatch(t, body, `(?s)<table class="prov cats">(.*?)</table>`, "the category table")
	rows := regexp.MustCompile(
		`(?s)<th>([^<]*)</th>\s*<td class="v">(.*?)</td>.*?t-([a-z]+)">[^<]*</span></td>\s*<td class="d">(.*?)</td>`).
		FindAllStringSubmatch(tbl, -1)
	if len(rows) < 4 {
		t.Fatalf("only %d category rows; want at least family, order, habitat and diet", len(rows))
	}

	labels := map[string]bool{}
	for _, r := range rows {
		label := strings.TrimSpace(html.UnescapeString(r[1]))
		origin, detail := r[3], strings.TrimSpace(stripTags(r[4]))
		labels[strings.Fields(label + " ")[0]] = true

		switch origin {
		case "published":
			if detail == "" {
				t.Errorf("%s: published but names no source", label)
			}
		case "derived", "birdview":
			if detail == "" {
				t.Errorf("%s: %s but states no rule or reason", label, origin)
			}
		default:
			t.Errorf("%s: origin %q — an unmarked category is an opinion presented "+
				"as a finding (product.categoryProvenance)", label, origin)
		}
	}
	for _, want := range []string{"Family", "Order", "Habitat", "Diet", "Regions"} {
		if !labels[want] {
			t.Errorf("the category table omits %q", want)
		}
	}

	// Regions are ONE row carrying every region, not a row apiece: they
	// share a single rule, and repeating it per region would bury the
	// habitat rows where BirdView differs from AVONET.
	if n := strings.Count(tbl, `>Regions`); n != 1 {
		t.Errorf("%d region rows; the regions share one provenance and take one row", n)
	}
	regionRow := mustMatch(t, tbl, `(?s)<th>(Regions[^<]*)</th>\s*<td class="v">(.*?)</td>`,
		"the regions row")
	if len(strings.Fields(regionRow)) < 2 {
		t.Error("the regions row lists no regions")
	}

	// A bird in no region gets no Regions row at all, rather than an empty
	// one claiming it was placed nowhere by a rule. Six species are at sea
	// where no land polygon reaches (product.regionAndMonthFilter), and the fixture heron stands
	// for them.
	heron := mustMatch(t, mustGet(t, s, "/species/greblu"),
		`(?s)<table class="prov cats">(.*?)</table>`, "the heron's category table")
	if strings.Contains(heron, ">Regions") {
		t.Error("a bird in no region still shows a Regions row")
	}
	if !strings.Contains(heron, ">Habitat") {
		t.Error("the heron lost its other categories along with its regions")
	}

	// Every doc a provenance cites must be one this server serves, **and
	// the section it names must exist**. A link to the right document and
	// the wrong anchor resolves with a 200 and lands the reader nowhere
	// near the explanation, which is how the regions row came to cite a
	// document that said nothing about regions.
	for _, m := range regexp.MustCompile(`href="(/docs/[^"#]+)(#[^"]*)?"`).
		FindAllStringSubmatch(body, -1) {
		code, doc := get(t, s, m[1])
		if code != http.StatusOK {
			t.Errorf("provenance cites %s, which the server answers with %d", m[1], code)
			continue
		}
		if m[2] == "" {
			continue
		}
		if anchor := `id="` + strings.TrimPrefix(m[2], "#") + `"`; !strings.Contains(doc, anchor) {
			t.Errorf("provenance cites %s%s, and %s has no such section",
				m[1], m[2], m[1])
		}
	}
}

// stripTags removes markup so a cell's text can be compared.
func stripTags(s string) string {
	return html.UnescapeString(regexp.MustCompile(`<[^>]+>`).ReplaceAllString(s, ""))
}

// TestServedDocumentsHaveNoBrokenLinks walks every document BirdView serves
// and fetches every internal link in it.
//
// The bug it was written for: the guides link to the spec as
// `../river/requirements/sources/product/`, correct in a repository checkout and a 404 in a
// browser. Nothing noticed, because acceptance.categoryOriginShown only checks the `/docs/` links
// on a species page, and these are links between documents.
//
// **Relative links are resolved the way a browser resolves them**, against
// the page they sit on. The first version followed only links beginning
// with "/", so an unrewritten repository link -- the founding bug -- was
// skipped rather than fetched, and the test passed over 53 of them when the
// bundle moved from spec/ to river/.
//
// Verifies: product.categoryProvenance. Criterion: acceptance.categoryOriginShown.
func TestServedDocumentsHaveNoBrokenLinks(t *testing.T) {
	s := newTestServer(t)
	seen := map[string]bool{}
	repoLinks := 0
	var visit func(path string)
	visit = func(path string) {
		if seen[path] {
			return
		}
		seen[path] = true
		code, body := get(t, s, path)
		if code != http.StatusOK {
			t.Errorf("%s answers %d", path, code)
			return
		}
		base, err := url.Parse(path)
		if err != nil {
			t.Fatalf("visiting unparsable path %q: %v", path, err)
		}
		for _, m := range regexp.MustCompile(`href="([^"]*)"`).FindAllStringSubmatch(body, -1) {
			ref, err := url.Parse(html.UnescapeString(m[1]))
			if err != nil {
				continue
			}
			// A link into the published repository must name a file that
			// exists in this checkout, or it is a 404 waiting for publication.
			if rest, ok := strings.CutPrefix(ref.String(), repository+"/"); ok {
				repoLinks++
				parts := strings.SplitN(rest, "/", 3) // blob|tree, main, path
				if len(parts) != 3 {
					t.Errorf("%s links to %s, which names no path in the repository", path, ref)
					continue
				}
				file := strings.SplitN(parts[2], "#", 2)[0]
				if _, err := os.Stat(filepath.Join("..", "..", filepath.FromSlash(file))); err != nil {
					t.Errorf("%s links to %s, but %s does not exist in the repository", path, ref, file)
				}
				continue
			}
			if ref.Scheme != "" || ref.Host != "" {
				continue // external, and not ours to check
			}
			href := base.ResolveReference(ref).Path
			if href == path && ref.Path == "" {
				continue // a fragment within the page
			}
			if href == "" || strings.HasPrefix(href, "/static/") {
				continue
			}
			// Every link must resolve; document links are then followed.
			if c, _ := get(t, s, href); c != http.StatusOK {
				t.Errorf("%s links to %s, which answers %d", path, href, c)
				continue
			}
			if strings.HasPrefix(href, "/docs") {
				visit(href)
			}
		}
	}
	visit("/docs/")
	if len(seen) < 4 {
		t.Errorf("only walked %d documents; the index is not listing the set", len(seen))
	}
	if repoLinks == 0 {
		t.Error("no guide links into the repository, so the existence check proved nothing")
	}
}

// TestVendoredTermsAreNotServed is a licensing check, not a routing one.
//
// river/requirements/sources/ and third_party/ hold other people's terms — the ABA
// checklist's, Wiley's, iNaturalist's — kept for hashing and provenance.
// The iNaturalist pages moved to third_party/ for the Apache 2.0 release,
// so both trees are probed: a check that only knew the old home would go on
// passing after the files it guards had moved somewhere it never looks.
// They are quoted for a record, not ours to republish, so no path may serve
// them (tech.fetchRestrictedInputs). Since decisions.requirementsNotServed nothing under river/ is embedded at all; the
// probes stay, so that serving the requirements again cannot quietly bring
// the vendored terms back with them.
//
// Verifies: tech.fetchRestrictedInputs. Criterion: acceptance.categoryOriginShown.
func TestVendoredTermsAreNotServed(t *testing.T) {
	s := newTestServer(t)
	for _, path := range []string{
		"/spec/sources/aba", "/spec/sources/wiley", "/spec/sources/inat-terms",
		"/third_party", "/third_party/inatterms/terms.html",
		"/third_party/avonet/LICENSE",
		"/spec", "/spec/product", "/spec/check", "/docs/../README", "/spec/../README",
	} {
		if code, _ := get(t, s, path); code == http.StatusOK {
			t.Errorf("%s is served; vendored terms and non-documents must not be", path)
		}
	}
}

// TestDocumentTitlesAreHeadings pins the title rule: a document is listed
// and titled by its first top-level heading, not by its first line.
//
// The case that forced it was the requirements' process.md, which opened
// with an HTML comment and was listed in the index as "<!--". The
// requirements are no longer served (decisions.requirementsNotServed); the rule stands for the guides.
//
// Verifies: product.categoryProvenance. Criterion: acceptance.categoryOriginShown.
func TestDocumentTitlesAreHeadings(t *testing.T) {
	s := newTestServer(t)
	for _, set := range []string{"/docs"} {
		body := mustGet(t, s, set)
		for _, m := range regexp.MustCompile(`<li><a href="[^"]*">([^<]*)</a>`).
			FindAllStringSubmatch(body, -1) {
			title := html.UnescapeString(m[1])
			if strings.ContainsAny(title, "<>") || strings.TrimSpace(title) == "" {
				t.Errorf("%s lists a document titled %q; a title is a heading, "+
					"not whatever is on line one", set, title)
			}
		}
	}
}

// TestAddedListsEveryUnpublishedAssignment checks the one page that answers
// "what has this project decided on its own account".
//
// The detail pages answer it bird by bird, which is a different question.
// Somebody auditing BirdView wants the whole of it at once, and a claim
// that is only discoverable one species at a time is not really disclosed.
//
// Verifies: product.categoryProvenance. Criterion: acceptance.categoryOriginShown.
func TestAddedListsEveryUnpublishedAssignment(t *testing.T) {
	s := newTestServer(t)
	body := mustGet(t, s, "/added")

	rows := regexp.MustCompile(`<tr>\s*<td>(Habitat|Diet)</td>`).FindAllString(body, -1)

	// Count what the dataset actually carries, and insist the page shows
	// all of it: a page that quietly lists a subset is worse than none.
	want := 0
	for i := range s.ds.Species {
		sp := &s.ds.Species[i]
		for _, a := range append(append([]dataset.Assignment{}, sp.Habitats...), sp.Niches...) {
			if a.Origin != dataset.Published {
				want++
			}
		}
	}
	if want == 0 {
		t.Fatal("the fixture has no unpublished assignment; this proves nothing")
	}
	if len(rows) != want {
		t.Errorf("the page lists %d assignments, the dataset has %d that are not published",
			len(rows), want)
	}
	// Published ones are excluded: they are AVONET's and there are 1,678.
	if strings.Contains(body, "<code>published</code>") {
		t.Error("the page lists published assignments, which are not BirdView's to claim")
	}
	// Every row says who and why.
	for _, m := range regexp.MustCompile(
		`<td><code>(\w+)</code></td>\s*<td>([^<]*)</td>`).FindAllStringSubmatch(body, -1) {
		if m[1] != "derived" && m[1] != "birdview" {
			t.Errorf("row origin %q is neither derived nor birdview", m[1])
		}
		if strings.TrimSpace(m[2]) == "" {
			t.Errorf("a %s row gives no reason", m[1])
		}
	}
}

// onlyLink pulls the Only href that follows a given group toggle, scoped to
// one grouping so the same zone name in the other grouping is not matched.
func onlyLink(t *testing.T, scope, afterLabel string) string {
	t.Helper()
	i := strings.Index(scope, ">"+afterLabel)
	if i < 0 {
		t.Fatalf("no group toggle labelled %q in this scope", afterLabel)
	}
	m := regexp.MustCompile(`<a class="btn only" href="([^"]+)"`).FindStringSubmatch(scope[i:])
	if m == nil {
		t.Fatalf("no Only link after %q", afterLabel)
	}
	return html.UnescapeString(m[1])
}

// Only narrows to one group in a click, and must not disturb the sections
// beside it (product.onlyPerGroup).
//
// Watched fail: building OnlyURL from the current selection instead of from
// the group makes the two starting selections disagree; pointing a trait's
// Only at both trait parameters clears the habitat selection on a diet click.
//
// Verifies: product.onlyPerGroup. Criterion: acceptance.onlyNarrowsOneGroup.
func TestOnlyNarrowsToOneGroup(t *testing.T) {
	s := newTestServer(t)
	_, body := get(t, s, "/")

	// A region zone: its regions stay, every other region goes.
	regions := box(t, body, "regions")
	geo := grouping(t, regions, "Geographic")
	zone := mustMatch(t, geo, `<a class="btn ohdr[^"]*" href="[^"]*"\s*\n?\s*title="[^"]*">([^<]+)</a>`,
		"the first zone in the Geographic grouping")
	href := onlyLink(t, geo, zone)

	u, err := url.Parse(href)
	if err != nil {
		t.Fatalf("Only href %q does not parse: %v", href, err)
	}
	q := u.Query()
	if _, ok := q["region"]; !ok {
		t.Fatalf("Only for zone %q sets no region parameter: %q", zone, href)
	}
	// It must not reach into the other filters at all.
	for _, p := range []string{"family", "habitat", "niche"} {
		if _, ok := q[p]; ok {
			t.Errorf("Only on a region set %q=%v; regions and %s are different sections",
				p, q[p], p)
		}
	}

	// Following it leaves exactly that zone's regions on, and nothing else.
	_, after := get(t, s, href)
	ar := box(t, after, "regions")
	onChips := regexp.MustCompile(`<a class="chip rchip on"`).FindAllString(
		grouping(t, ar, "Geographic"), -1)
	zoneChips := regexp.MustCompile(`<a class="chip rchip[^"]*"`).FindAllString(
		onlyZone(t, grouping(t, ar, "Geographic"), zone), -1)
	if len(onChips) != len(zoneChips) {
		t.Errorf("after Only on %q, %d regions are selected page-wide but the zone holds %d",
			zone, len(onChips), len(zoneChips))
	}

	// A trait value: its own axis narrows, the other axis is untouched.
	traits := box(t, body, "traits")
	diet := grouping(t, traits, "Diet")
	val := mustMatch(t, diet, `<a class="btn thdr[^"]*" href="[^"]*"\s*\n?\s*title="[^"]*">([^<]+)</a>`,
		"the first diet value")
	dh := onlyLink(t, diet, val)
	du, err := url.Parse(dh)
	if err != nil {
		t.Fatalf("Only href %q does not parse: %v", dh, err)
	}
	if _, ok := du.Query()["habitat"]; ok {
		t.Errorf("Only on a diet value set habitat=%v; diet and habitat are separate sections "+
			"and narrowing one must leave the other whole", du.Query()["habitat"])
	}

	// Independence from the current selection: the same group yields the same
	// Only target from a different starting state.
	_, narrowed := get(t, s, "/?region=CA")
	g2 := grouping(t, box(t, narrowed, "regions"), "Geographic")
	if h2 := onlyLink(t, g2, zone); h2 != href {
		t.Errorf("Only for %q differs by starting selection:\n  from all: %s\n  from CA:  %s",
			zone, href, h2)
	}
}

// onlyZone narrows a grouping to one zone's block.
func onlyZone(t *testing.T, grouping, zone string) string {
	t.Helper()
	i := strings.Index(grouping, ">"+zone)
	if i < 0 {
		t.Fatalf("no zone %q", zone)
	}
	rest := grouping[i:]
	if k := strings.Index(rest, `<div class="ogroup">`); k >= 0 {
		return rest[:k]
	}
	return rest
}

// The unit buttons must name the units the page actually renders. They did
// not: imperial was labelled "in / g" while every imperial mass on the page
// was in ounces, so the control advertised a combination the site never
// produced.
//
// Derived from the renderer, not from the button: format() is what decides
// the suffix, so asking it is asking the thing the label is about. Reading
// the expectation off the button would make this test agree with itself.
//
// Watched fail: restoring "in / g" reports that the imperial label says g
// where the page renders oz.
//
// Verifies: product.sectionThresholdDefault. Criterion: acceptance.controlsMatchServer.
func TestUnitButtonsNameTheUnitsRendered(t *testing.T) {
	_, body := get(t, newTestServer(t), "/")

	for _, tc := range []struct {
		metric3 bool
		param   string
	}{{false, "imperial"}, {true, "metric"}} {
		// A small bird's mass: format() switches to pounds and kilograms
		// further up, so the base unit the button names is the one a
		// sparrow-sized mass renders in.
		rendered := unitsFor(tc.metric3, 25).format(25, view.ByMass)
		unit := strings.TrimLeft(rendered, "0123456789.,  ~")
		unit = strings.TrimSpace(unit)

		label := mustMatch(t, body,
			`href="[^"]*units=`+tc.param+`[^"]*">([^<]+)</a>`,
			"the "+tc.param+" units button")
		if !strings.Contains(label, unit) {
			t.Errorf("the %s button is labelled %q, but %s masses render as %q (unit %q)",
				tc.param, label, tc.param, rendered, unit)
		}
	}
}

// realDataset loads the compiled dataset, or skips. Tests make no network
// request and the dataset is not committed (tech.datasetNeverCommitted), so a check over the real
// data is a check that runs when the data is there.
func realDataset(t *testing.T) *dataset.Dataset {
	t.Helper()
	ds, err := load(filepath.Join("..", "..", "web", "data", "species.json"))
	if err != nil {
		t.Skipf("no compiled dataset; run birdview-ingest first (%v)", err)
	}
	return ds
}

// The tile cube must never reach the crossbar, over every species at every
// scale and metric the index offers.
//
// acceptance.massDrawnAsWaterCube recorded this as "0 overlaps in 841" in prose. It is a geometric
// invariant over the real dataset, and prose cannot be run: decisions.trunkExponentCalibrated moved
// every length in the project and nothing failed, because nothing checked.
// It was caught by reading. This is that sentence turned into a test.
//
// It asserts a CLEARANCE rather than a boolean overlap, and reports the
// tightest case. A yes/no check turned out to be a poor guard: the cube sits
// bottom-left and the crossbar crosses the upper middle, so the two
// conditions are ANDed and the vertical separation dominates. Floating the
// cube to the top of the tile, sliding it 120px right, tripling the
// projection depth and scaling the whole cube by 1.6 all leave it clear.
// Only the margin moves under those, which is exactly what a regression
// needs to see.
//
// The floor asserted is zero, taken from the requirement -- the cube must
// not reach the crossbar -- and not a figure read off a run. The measured
// minimum is logged instead of asserted, so a change that halves it is
// visible without pinning today's number as a rule.
//
// Watched fail: sizing the side on raw mass instead of its cube root --
// acceptance.massDrawnAsWaterCube's own mutation for the arithmetic -- overlaps on 785 of 839 species
// by length, 730 by wingspan and 790 by mass.
//
// Verifies: product.labeledDetailDiagram. Criterion: acceptance.massDrawnAsWaterCube.
func TestTileCubeNeverReachesTheCrossbar(t *testing.T) {
	ds := realDataset(t)
	all := make([]*dataset.Species, 0, len(ds.Species))
	for i := range ds.Species {
		all = append(all, &ds.Species[i])
	}

	for _, scale := range []view.ScaleMode{view.PerSection, view.Global} {
		for _, metric := range []view.Metric{view.ByLength, view.ByWingspan, view.ByMass} {
			secs := view.Sections(all, metric, view.DefaultStepPercent)
			refs := view.Scales(secs, scale)

			spanRefs := make([]float64, len(secs))
			var globalSpan float64
			for i, sec := range secs {
				for _, sp := range sec.Species {
					if w := sp.Wingspan.Value; w > spanRefs[i] {
						spanRefs[i] = w
					}
				}
				if spanRefs[i] > globalSpan {
					globalSpan = spanRefs[i]
				}
			}
			if scale == view.Global {
				for i := range spanRefs {
					spanRefs[i] = globalSpan
				}
			}

			checked, overlaps := 0, 0
			minClear, tightest, worst := math.Inf(1), "", ""
			for i, sec := range secs {
				for _, sp := range sec.Species {
					g := view.Layout(sp, refs[i], spanRefs[i], tileBox)
					c := newMassCube(sp, tileBox.H, g.PxPerMm, "")
					if c.Side <= 0 {
						continue // no mass, no cube
					}
					checked++
					// The crossbar is a horizontal line at YShoulder running
					// CX +/- CrossbarHalfWidth. The cube is its front face
					// plus the projection to the right. They are clear if the
					// bar passes above the cube, or begins to its right;
					// the clearance is whichever separation is larger.
					vertical := c.Y - g.YShoulder
					horizontal := (g.CX - g.CrossbarHalfWidth) - (c.X + c.Side + c.Depth)
					clear := math.Max(vertical, horizontal)
					if clear < minClear {
						minClear, tightest = clear, sp.ScientificName
					}
					if clear <= 0 {
						overlaps++
						if worst == "" {
							worst = fmt.Sprintf("%s: cube x[%.1f,%.1f] y[%.1f,%.1f], "+
								"crossbar y=%.1f from x=%.1f",
								sp.ScientificName, c.X, c.X+c.Side+c.Depth, c.Y, c.Y+c.Side,
								g.YShoulder, g.CX-g.CrossbarHalfWidth)
						}
					}
				}
			}
			if checked == 0 {
				t.Fatalf("scale=%s metric=%s: no cubes checked; the test proves nothing",
					scale, metric)
			}
			if overlaps > 0 {
				t.Errorf("scale=%s metric=%s: the cube reaches the crossbar on %d of %d species. %s",
					scale, metric, overlaps, checked, worst)
			}
			t.Logf("scale=%-7s metric=%-8s %d species, tightest clearance %.1fpx (%s)",
				scale, metric, checked, minClear, tightest)
		}
	}
}

// The smallest bird in a scope chooses the unit, so a comparison never mixes
// two (product.oneUnitInComparison).
//
// Expected values are computed from the definition -- 1 oz = 28.349523125 g,
// 16 oz = 1 lb -- not read off a run.
//
// Watched fail: choosing the unit from the largest mass instead of the
// smallest renders the 27 g sparrow as "0.1 lb" -- one significant figure
// for a bird, against "0.95 oz".
//
// Verifies: product.oneUnitInComparison. Criterion: acceptance.massInOneUnit.
func TestUnitsForTakesTheSmallestBird(t *testing.T) {
	const sparrow, swan = 27.0, 10886.0 // g: House Sparrow, Trumpeter Swan

	// Mixed scope: the small bird pulls everything into ounces.
	u := unitsFor(false, swan, sparrow)
	if got, want := u.format(sparrow, view.ByMass), "0.95 oz"; got != want {
		t.Errorf("sparrow in a mixed scope: got %q, want %q", got, want)
	}
	if got, want := u.format(swan, view.ByMass), "384.0 oz"; got != want {
		t.Errorf("swan in a mixed scope: got %q, want %q -- the scope holds a bird "+
			"that cannot be written in pounds", got, want)
	}

	// All-large scope keeps pounds: the rule is "the smallest decides", not
	// "always ounces".
	if got, want := unitsFor(false, swan, 5000).format(swan, view.ByMass), "24.0 lb"; got != want {
		t.Errorf("swan in an all-large scope: got %q, want %q", got, want)
	}
	// The same, metric.
	if got, want := unitsFor(true, swan, sparrow).format(swan, view.ByMass), "10886 g"; got != want {
		t.Errorf("swan in a mixed metric scope: got %q, want %q", got, want)
	}
	if got, want := unitsFor(true, swan, 5000).format(swan, view.ByMass), "10.89 kg"; got != want {
		t.Errorf("swan in an all-large metric scope: got %q, want %q", got, want)
	}
	// No recorded mass must not pick the large unit on no evidence.
	if u := unitsFor(false); u.big {
		t.Error("an empty scope chose pounds with nothing to justify it")
	}
	if u := unitsFor(false, 0, 0); u.big {
		t.Error("a scope of unrecorded masses chose pounds with nothing to justify it")
	}
}

// No section and no comparison may render two mass units at once, over the
// real dataset.
//
// This is the assertion that would have caught the original defect: one
// section printing "14.5 oz" beside "1.2 lb" for birds of the same length,
// found in a screenshot and by no check.
//
// Watched fail: reverting the index to a per-value unit makes sections
// render both oz and lb.
//
// Verifies: product.oneUnitInComparison. Criterion: acceptance.massInOneUnit.
func TestNoSectionMixesMassUnits(t *testing.T) {
	ds := realDataset(t)
	srv, err := newServer(ds)
	if err != nil {
		t.Fatalf("newServer: %v", err)
	}
	unit := regexp.MustCompile(`\b\d[\d.,]*\s*(oz|lb|kg|g)\b`)

	for _, q := range []string{
		"/?by=mass&scale=section", "/?by=mass&scale=global",
		"/?by=length&scale=section", "/?by=length&scale=global",
		"/?by=mass&scale=section&units=metric",
	} {
		_, body := get(t, srv, q)
		// Each section is a <section> block; check them one at a time, since
		// two sections may legitimately differ from each other.
		secs := regexp.MustCompile(`(?s)<section\b.*?</section>`).FindAllString(body, -1)
		if len(secs) == 0 {
			t.Fatalf("%s: no sections found; the test proves nothing", q)
		}
		for i, sec := range secs {
			seen := map[string]bool{}
			for _, m := range unit.FindAllStringSubmatch(sec, -1) {
				seen[m[1]] = true
			}
			if (seen["oz"] && seen["lb"]) || (seen["g"] && seen["kg"]) {
				t.Errorf("%s: section %d mixes mass units %v -- a reader has to "+
					"convert to compare two birds in one section", q, i+1, keysOf(seen))
			}
		}
	}
}

// A near-match is a measurement, of the wrong bird. The tile must say that
// rather than call it an estimate.
//
// All three footer cells used to carry title="estimated" for anything not
// measured, which is false for a near-match and contradicted the stroke
// beside it: product.measuredVersusModeled draws near-match SOLID precisely because callipers
// touched a real bird. It also told the reader nothing they could act on --
// the Chihuahuan Meadowlark and the Eastern Meadowlark render identical
// length, wingspan and mass, and the only thing on the index distinguishing
// them was one tilde labelled with the wrong reason.
//
// Watched fail: returning "estimated" from whyNotMeasured for NearMatch
// reports the donor missing.
//
// Verifies: product.provenancePerMetric, product.measuredVersusModeled. Criterion: acceptance.unitToggleSwitches.
func TestNearMatchTooltipNamesTheDonor(t *testing.T) {
	near := dataset.Provenance{Kind: dataset.NearMatch, SourceTaxon: "Sturnella magna"}
	got := whyNotMeasured(near)
	if !strings.Contains(got, "Sturnella magna") {
		t.Errorf("near-match tooltip = %q, want it to name the donor: two tiles showing "+
			"an identical mass are not a coincidence when one borrowed it", got)
	}
	if strings.HasPrefix(got, "estimated") {
		t.Errorf("near-match tooltip = %q, but a near-match is measured, with callipers, "+
			"on a sibling taxon (product.measuredVersusModeled draws it solid for that reason)", got)
	}
	// A modeled value is still an estimate, and still says so.
	mod := dataset.Provenance{Kind: dataset.Modeled, Model: "cranium modeled as 7.5 x mass^(1/3)"}
	if m := whyNotMeasured(mod); !strings.HasPrefix(m, "estimated") {
		t.Errorf("modeled tooltip = %q, want it to lead with estimated", m)
	}
	// A near-match with no donor recorded must not claim one.
	if m := whyNotMeasured(dataset.Provenance{Kind: dataset.NearMatch}); strings.Contains(m, "Sturnella") {
		t.Errorf("a near-match with no source taxon invented one: %q", m)
	}
}

// Every `#fragment` a page cites must resolve to a heading in the document
// it points at.
//
// TestServedDocumentsHaveNoBrokenLinks walks document links but drops the
// fragment first -- "fragments are within the page" -- so it checks that
// /docs/data exists and never that #commonest does. Nothing else looked
// either. The provenance rows on a detail page cite by anchor precisely so
// a reader lands on the rule and not the top of a 300-line document, and a
// citation that silently lands at the top is the failure that makes the
// anchor pointless.
//
// Caught for real: this test was written after adding a row citing
// /docs/data#commonest, with no such heading in the file, and the whole
// suite stayed green.
//
// Watched fail: renaming the "Commonest" heading in docs/data.md to
// anything else reports the dangling citation.
//
// Verifies: product.categoryProvenance. Criterion: acceptance.categoryOriginShown.
func TestCitedAnchorsResolve(t *testing.T) {
	s := newTestServer(t)

	// The ids goldmark actually emitted, per served document.
	ids := map[string]map[string]bool{}
	headingID := regexp.MustCompile(`<h[1-6][^>]*\bid="([^"]+)"`)
	for _, set := range []string{"/docs/"} {
		index := mustGet(t, s, set)
		for _, m := range regexp.MustCompile(`href="(`+set+`[^"#]+)"`).
			FindAllStringSubmatch(index, -1) {
			doc := m[1]
			if ids[doc] != nil {
				continue
			}
			ids[doc] = map[string]bool{}
			for _, h := range headingID.FindAllStringSubmatch(mustGet(t, s, doc), -1) {
				ids[doc][h[1]] = true
			}
		}
	}
	if len(ids) < 4 {
		t.Fatalf("only found %d served documents; the index is not listing the set", len(ids))
	}
	var withIDs int
	for _, set := range ids {
		if len(set) > 0 {
			withIDs++
		}
	}
	if withIDs == 0 {
		t.Fatal("no document emitted a heading id, so every citation would pass vacuously")
	}

	// Every anchored citation anywhere on the site.
	cite := regexp.MustCompile(`href="(/docs/[^"#]+)#([^"]+)"`)
	checked := 0
	for _, page := range []string{"/", "/species/greblu", "/species/amerob", "/added",
		"/docs/", "/docs/data", "/docs/using", "/excluded"} {
		body := mustGet(t, s, page)
		for _, m := range cite.FindAllStringSubmatch(body, -1) {
			doc, frag := m[1], m[2]
			known, ok := ids[doc]
			if !ok {
				t.Errorf("%s cites %s#%s, but %s is not a served document", page, doc, frag, doc)
				continue
			}
			checked++
			if !known[frag] {
				t.Errorf("%s cites %s#%s, and %s has no such heading; the reader lands "+
					"at the top of the document instead of on the rule", page, doc, frag, doc)
			}
		}
	}
	if checked == 0 {
		t.Error("no anchored citations found; this test is checking nothing")
	}
}

// The share on a tile is the share of what the regions ON SCREEN record,
// and it must move when the selection moves.
//
// This replaced a fixed "commonest in Yukon" line, which answered a
// question a filtered index does not pose: narrow to Arizona and be told
// about Yukon. The figure is now a ratio of two sums over the selection,
// which is also why nothing is precomputed -- sums compose, so 69 regions
// do not imply 2^69 subaggregates.
//
// Watched fail: ignoring opt.Regions in shareOf and always summing every
// region makes the one-region and all-region shares identical.
//
// Verifies: product.shareOfRecords. Criterion: acceptance.commonestShareIsShare.
func TestTileShareFollowsTheRegionFilter(t *testing.T) {
	s := newTestServer(t)

	// The fixture robin: 73,005 records in Ontario of 1,817,968 there.
	oneIn := func(path, bird string) string {
		_, body := get(t, s, path)
		i := strings.Index(body, ">"+bird)
		if i < 0 {
			t.Fatalf("%s: %s is not on the page", path, bird)
		}
		seg := body[i:]
		if j := strings.Index(seg, "</div></a>"); j > 0 {
			seg = seg[:j]
		}
		m := regexp.MustCompile(`<b>1 in ([\d,]+)</b>`).FindStringSubmatch(seg)
		if m == nil {
			t.Fatalf("%s: no share on %s's tile", path, bird)
		}
		return m[1]
	}

	all := oneIn("/", "American Robin")
	ontario := oneIn("/?region=ON", "American Robin")
	if all == ontario {
		t.Errorf("the robin's figure is %q both with all regions and in Ontario "+
			"alone; the figure is not following the filter", all)
	}
	// Ontario is where the fixture robin is densest, so narrowing to it must
	// make the bird a LARGER share -- a smaller "1 in N".
	a, o := atoiComma(t, all), atoiComma(t, ontario)
	if o >= a {
		t.Errorf("robin is 1 in %d across everything and 1 in %d in Ontario alone; "+
			"narrowing to where it is densest must raise its share", a, o)
	}
}

// TestTheUnitSaysSpeciesAndNotBird asserts the noun the "1 in N" figure is
// counted in, on the tile and in the tooltip that explains it.
//
// The denominator counts species on checklists and deliberately ignores how
// many individuals each one saw -- twelve sparrows is one, and so is a flock
// of fifty thousand Snow Geese. "bird observations" named the one quantity
// the figure is NOT, so a reader taking 1 in 32 to mean one bird in
// thirty-two misread nothing except the unit, and the unit was the only
// thing that could have told them (product.shareOfRecords).
//
// It reads the tile's own cell rather than searching the page: the word
// appears in the docs and the help text too, so a whole-page search would
// pass with the figure missing from every tile, which is the shape of a
// check that tests nothing.
//
// Watched fail: see acceptance.figureSaysWhatItCounts.
//
// Verifies: product.shareOfRecords. Criterion: acceptance.figureSaysWhatItCounts.
func TestTheUnitSaysSpeciesAndNotBird(t *testing.T) {
	s := newTestServer(t)
	_, body := get(t, s, "/")
	i := strings.Index(body, ">American Robin")
	if i < 0 {
		t.Fatal("the robin is not on the index")
	}
	seg := body[i:]
	if j := strings.Index(seg, "</div></a>"); j > 0 {
		seg = seg[:j]
	}
	// The figure and its unit have to be in the same cell, or this asserts
	// that the words appear somewhere near a bird's name.
	unit := regexp.MustCompile(`<b>1 in [\d,]+</b>\s*([a-z ]*observations)`).FindStringSubmatch(seg)
	if unit == nil {
		t.Fatalf("no \"1 in N <unit> observations\" on the robin's tile: %q", seg)
	}
	if got := strings.TrimSpace(unit[1]); got != "species observations" {
		t.Errorf("the tile counts in %q, want \"species observations\": the denominator "+
			"is species on checklists, and naming it in birds asserts the individual "+
			"count the figure deliberately does not read", got)
	}

	// And the tooltip that spells the figure out agrees with the cell. Two
	// nouns for one quantity is worse than the wrong one used consistently.
	ds := realDataset(t)
	var sp *dataset.Species
	for i := range ds.Species {
		if len(ds.Species[i].Regions) > 0 {
			sp = &ds.Species[i]
			break
		}
	}
	if sp == nil {
		t.Fatal("no species has a region")
	}
	cell := shareOf(sp, ds, options{})
	if cell == nil {
		t.Fatal("no share cell for a species with regions")
	}
	if !strings.Contains(cell.Title, "species observations") {
		t.Errorf("the tooltip reads %q, which does not count in species observations "+
			"as the cell beside it does", cell.Title)
	}
	if strings.Contains(cell.Title, "bird observations") {
		t.Errorf("the tooltip still counts in bird observations: %q", cell.Title)
	}
}

func atoiComma(t *testing.T, s string) int {
	t.Helper()
	n, err := strconv.Atoi(strings.ReplaceAll(s, ",", ""))
	if err != nil {
		t.Fatalf("not a number: %q", s)
	}
	return n
}

// No page BirdView serves may publish a personal address, in any of the
// forms that reach a reader (product.noPersonalAddress).
//
// "birdview at ajmani dot net" was in the header for the project's whole
// life. The obfuscation is the tell: it was written that way because a
// plain mailto on a public page is a harvested mailbox, which is an
// argument against publishing the address rather than for spelling it
// oddly. Feedback belongs on a tracker, and the owner's personal mail is
// not the project's contact point.
//
// The spaced-out form is why this matches on the DOMAIN rather than on an
// `@`: "ajmani dot net" contains no at-sign and a check for one would have
// passed over the thing it exists to catch.
//
// Watched fail: putting the old sentence back in the header reports it on
// every page that renders the header.
//
// Verifies: product.noPersonalAddress. Criterion: acceptance.noContactAddressPublished.
func TestNoPersonalAddressIsPublished(t *testing.T) {
	s := newTestServer(t)
	pages := []string{"/", "/species/greblu", "/species/amerob", "/added",
		"/compare?species=amerob&species=blujay", "/docs/"}
	guides := regexp.MustCompile(`href="(/docs/[^"#]+)"`).FindAllStringSubmatch(mustGet(t, s, "/docs/"), -1)
	if len(guides) < 4 {
		t.Fatalf("the docs index lists only %d guides; it is not listing the set", len(guides))
	}
	for _, m := range guides {
		pages = append(pages, m[1])
	}

	// Both the plain form and the spelled-out obfuscation.
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}`),
		regexp.MustCompile(`(?i)\b[a-z0-9]+\s+(?:at|AT)\s+[a-z0-9]+\s+dot\s+[a-z]{2,}\b`),
		regexp.MustCompile(`(?i)mailto:`),
	}
	for _, p := range pages {
		body := mustGet(t, s, p)
		// The User-Agent constants are a different thing and are not served;
		// if one ever leaks onto a page, this catches it too.
		for _, re := range patterns {
			if m := re.FindString(body); m != "" {
				t.Errorf("%s publishes a contact address %q; feedback goes to the "+
					"issue tracker, and a personal mailbox on a public page is "+
					"harvested within the week", p, m)
			}
		}
	}

	// The requirements are no longer served (decisions.requirementsNotServed), but they are published
	// with the repository, and this check covered them while they were served.
	// Read the same files from the checkout so that coverage is not lost.
	// The same files the server used to serve, now spread through the nested
	// layout. Provenance records and transcriptions under requirements/sources/<pkg>/
	// were never served, and are left out so the scope does not silently grow.
	var docs []string
	for _, pattern := range []string{"*.md", "goals/*.md", "requirements/*.md",
		"requirements/sources/product/*.md", "requirements/sources/tech/*.md"} {
		m, err := filepath.Glob(filepath.Join("..", "..", "river", pattern))
		if err != nil {
			t.Fatal(err)
		}
		docs = append(docs, m...)
	}
	if len(docs) < 15 {
		t.Fatalf("found %d requirement documents under river/; the layout has changed and the check would prove nothing", len(docs))
	}
	for _, path := range docs {
		body, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		for _, re := range patterns {
			if m := re.Find(body); m != nil {
				t.Errorf("%s publishes a contact address %q; the requirements are public "+
					"with the repository", filepath.ToSlash(path), m)
			}
		}
	}
}

// TestUndrawnSpeciesNamesItselfAt404 asserts that a URL for a bird the
// derivation classified and rejected says which bird and why, rather than
// answering `404 page not found`.
//
// What each verdict must carry is different, and the differences are the
// point:
//
//   - `historical` turns entirely on WHEN. Passenger Pigeon satisfies every
//     test for recurrence -- 89 episodes across 52 years -- and the only fact
//     that reconciles that with not drawing it is that the records stop in
//     1902. A page asserting "historical" without the date states a verdict
//     and withholds its evidence.
//   - `vagrant` turns on the bar, which is read from the derivation rather
//     than written into the page, so the two cannot drift apart.
//   - `absent` must print NO counts. The zeros in those columns mean "not
//     measured", and a page rendering "0 episodes across 0 years" would turn
//     an absence of measurement into a measurement of absence.
//
// And the status code is asserted three ways -- 404, not 200, not 410 -- so
// that a change to any one of them is caught. 200 would tell caches an
// address serving no bird is fine; 410 would claim a permanence that a scope
// regenerated from live data does not have (decisions.speciesUrlOutlivesSpecies).
//
// Verifies: product.undrawnSpeciesAnswer, decisions.speciesUrlOutlivesSpecies. Criterion: acceptance.outOfScopeExplained.
func TestUndrawnSpeciesNamesItselfAt404(t *testing.T) {
	s := newTestServer(t)

	for _, tc := range []struct {
		id      string
		want    []string
		absent  []string
		verdict string
	}{{
		id:      "paspig",
		verdict: "historical",
		want: []string{"Passenger Pigeon", "Ectopistes migratorius", "historical",
			// The evidence, and the date that is the whole of the verdict.
			"89 episodes", "52 years", "145 records", "18 May 1902",
			// The region the verdict came from, as a name and not a GADM id.
			"Ontario"},
		absent: []string{"CAN.9_1"},
	}, {
		id:      "sedwar1",
		verdict: "vagrant",
		want: []string{"Sedge Warbler", "Acrocephalus schoenobaenus", "vagrant",
			"2 episodes", "2 years", "4 records", "Alaska",
			// The bar it did not clear, derived from the thresholds rather
			// than spelled out here.
			fmt.Sprintf("%d episodes across %d years", status.MinEpisodes, status.MinYears)},
	}, {
		id:      "lillov1",
		verdict: "absent",
		want:    []string{"Lilian&#39;s Lovebird", "Agapornis lilianae", "absent"},
		// Nothing that reads as a count, in any of the forms the other two
		// verdicts render one.
		absent: []string{"0 episodes", "episodes across", "records.", "Recorded in"},
	}} {
		t.Run(tc.id, func(t *testing.T) {
			code, body := get(t, s, "/species/"+tc.id)
			if code != http.StatusNotFound {
				t.Errorf("status %d, want %d: the address serves no bird, and "+
					"200 would tell every cache otherwise", code, http.StatusNotFound)
			}
			if code == http.StatusGone {
				t.Errorf("status 410 claims the bird is permanently gone; scope is "+
					"re-derived from live data and %s can come back (decisions.speciesUrlOutlivesSpecies)", tc.id)
			}
			if !strings.Contains(body, "BirdView does not draw this bird") {
				t.Fatalf("no verdict on the page; body:\n%s", body)
			}
			for _, w := range tc.want {
				if !strings.Contains(body, w) {
					t.Errorf("page does not say %q", w)
				}
			}
			for _, a := range tc.absent {
				if strings.Contains(body, a) {
					t.Errorf("page says %q, which it has no measurement for", a)
				}
			}
		})
	}

	// An id that matches no species at all keeps the terse answer. The named
	// page is for a bird BirdView knows of and chose not to draw, and
	// inventing one to apologise for would be worse than saying nothing.
	code, body := get(t, s, "/species/nosuchbird")
	if code != http.StatusNotFound {
		t.Errorf("unknown id: status %d, want 404", code)
	}
	if strings.Contains(body, "BirdView does not draw this bird") {
		t.Errorf("unknown id gets the named page; it names no bird, so the page "+
			"would be an explanation of nothing:\n%s", body)
	}
}

// TestNoIDIsBothDrawnAndNot asserts over the real dataset that the drawn and
// rejected sets do not share an id.
//
// Over the real data and not the fixture, because the claim is about 1,411
// ids produced by one AviList join, and a fixture asserts it over whichever
// handful the fixture holds. The two sets are built the same way -- eBird
// code, scientific-name slug if there is none -- so nothing structural stops
// a collision; what prevents it is that AviList gives each species one row.
// That is a property of an input file, which is exactly the kind of thing
// that changes under you.
//
// A collision is not cosmetic. The server reads byID first, so the drawn bird
// would win and the rejected one would be unreachable -- but reverse the
// lookup order, which is an implementation detail nothing pins, and a live
// species page becomes an epitaph.
//
// Verifies: product.undrawnSpeciesAnswer. Criterion: acceptance.outOfScopeExplained.
func TestNoIDIsBothDrawnAndNot(t *testing.T) {
	ds := realDataset(t)
	if len(ds.Undrawn) == 0 {
		t.Fatal("the dataset carries no rejected species, so this check tests nothing; " +
			"re-run birdview-ingest")
	}
	drawn := make(map[string]string, len(ds.Species))
	for _, sp := range ds.Species {
		drawn[sp.ID] = sp.ScientificName
	}
	for _, u := range ds.Undrawn {
		if other, clash := drawn[u.ID]; clash {
			t.Errorf("id %q is both %s (drawn) and %s (not drawn)", u.ID, other, u.ScientificName)
		}
		if u.ID == "" || u.ScientificName == "" || u.Status == "" {
			t.Errorf("%+v: a rejected species needs an id, a name and a verdict "+
				"before it can explain itself", u)
		}
		if u.Name(dataset.LocaleWorld) == "" {
			t.Errorf("%s has no English name, so its page would apologise for a binomial",
				u.ScientificName)
		}
	}
	t.Logf("%d drawn, %d classified and not drawn, no shared id",
		len(ds.Species), len(ds.Undrawn))
}

// TestAbsentSpeciesCarryNoEvidence asserts over the real dataset that the
// ingest withheld the counts for every `absent` species, and kept them for
// every other verdict.
//
// Both halves. The first is the honesty rule of product.undrawnSpeciesAnswer; the second is what
// stops it being satisfied by an ingest that emits no evidence at all, which
// would pass a check written only one way round.
//
// Verifies: product.undrawnSpeciesAnswer. Criterion: acceptance.outOfScopeExplained.
func TestAbsentSpeciesCarryNoEvidence(t *testing.T) {
	ds := realDataset(t)
	seen := map[string]int{}
	for _, u := range ds.Undrawn {
		seen[u.Status]++
		switch u.Status {
		case "absent":
			if u.Measured() || u.Episodes != 0 || u.Years != 0 || u.Records != 0 || u.LastSeen != "" {
				t.Errorf("%s is absent but carries evidence %+v; those columns mean "+
					"NOT MEASURED, and publishing them makes an absence of "+
					"measurement look like a measurement of absence", u.ScientificName, u)
			}
		default:
			if !u.Measured() {
				t.Errorf("%s is %s but names no region, so the page can state the "+
					"verdict and not show it", u.ScientificName, u.Status)
			}
			if u.Episodes == 0 {
				t.Errorf("%s is %s with zero episodes, which no measured verdict is",
					u.ScientificName, u.Status)
			}
		}
	}
	for _, want := range []string{"vagrant", "historical", "absent"} {
		if seen[want] == 0 {
			t.Errorf("no %s species in the dataset, so that branch is unchecked", want)
		}
	}
	t.Logf("rejected: %v", seen)
}

// TestCommonestIsTheLargestLandShare asserts product.shareOfRecords's central claim over the
// real dataset: the region a species page names is the land region where the
// bird really is the biggest share of what gets recorded.
//
// Nothing asserted this before, which is how decisions.seaShareDiffers shipped. Every other
// test of the share renders `commonestOf` straight back out of the fixture,
// so the claim had only ever been checked against a value the test itself
// supplied -- and a value the fixture provides is a value no fixture test
// can falsify. It has to run over the compiled dataset or it proves nothing.
//
// Every region, land and sea (decisions.honestSeaDenominator). A sea's denominator is now every
// bird record on open water inside its box, measured the same way a land
// region's is, so the ranking spans both -- and 45 species peak at sea,
// including the albatross whose "commonest in Maine" started all this.
//
// Verifies: product.shareOfRecords, decisions.seaShareDiffers, decisions.honestSeaDenominator. Criterion: acceptance.commonestIsLargestShare.
func TestCommonestIsTheLargestShare(t *testing.T) {
	ds := realDataset(t)
	var checked, skipped int
	for i := range ds.Species {
		sp := &ds.Species[i]
		c := sp.Commonest
		if c == nil || c.Records == 0 || c.RegionRecords == 0 {
			skipped++
			continue
		}
		checked++
		claimed := float64(c.Records) / float64(c.RegionRecords)
		for j, id := range sp.Regions {
			total, ok := ds.RegionTotals[id]
			if !ok || total == 0 || j >= len(sp.RegionCounts) {
				continue // no denominator, so not in the ranking
			}
			if got := float64(sp.RegionCounts[j]) / float64(total); got > claimed {
				t.Errorf("%s: page says commonest in %s (1 in %d), but %s is 1 in %d",
					sp.ScientificName, c.Region, c.RegionRecords/c.Records,
					id, total/sp.RegionCounts[j])
			}
		}
	}
	if checked < len(ds.Species)/2 {
		t.Fatalf("only %d of %d species carry a commonest region; the check is "+
			"passing over most of the dataset", checked, len(ds.Species))
	}
	t.Logf("%d species checked, %d with no commonest region", checked, skipped)
}

// TestShareRangesOverOneSetOfRegions asserts the structural guard decisions.seaShareDiffers
// installed and decisions.honestSeaDenominator kept: shareIn will not take a numerator from a region
// it has no denominator for.
//
// The numerator and the denominator used to be two separate loops with two
// separate conditions, and that is how they came to range over different
// sets -- sea records over land records, for every pelagic. decisions.honestSeaDenominator gave the
// seas an honest denominator, so the symptom is gone; the guard is not, and
// must not be, because what went wrong was the structure and not the data.
// Asserting the invariant rather than today's region list keeps the mismatch
// unrepresentable.
//
// Also asserts the direction decisions.honestSeaDenominator reversed: every region now HAS a total,
// seas included. Both halves, so a regression in either direction shows.
//
// Verifies: product.shareOfRecords, decisions.seaShareDiffers, decisions.honestSeaDenominator. Criterion: acceptance.commonestIsLargestShare.
func TestShareRangesOverOneSetOfRegions(t *testing.T) {
	ds := realDataset(t)
	for _, id := range append(region.LandIDs(), region.MarineIDs()...) {
		if _, ok := ds.RegionTotals[id]; !ok {
			t.Errorf("%s has no record total, so a reader selecting it is told nothing; "+
				"every region BirdView offers must have a denominator (decisions.honestSeaDenominator)", id)
		}
	}

	// A species with records in both kinds, so a numerator that wrongly
	// skipped one would show.
	var sp *dataset.Species
	for i := range ds.Species {
		var land, sea bool
		for _, id := range ds.Species[i].Regions {
			if rg, ok := region.Get(id); ok && rg.Kind == region.Marine {
				sea = true
			} else {
				land = true
			}
		}
		if land && sea {
			sp = &ds.Species[i]
			break
		}
	}
	if sp == nil {
		t.Fatal("no species has records both on land and at sea; the check has nothing to catch")
	}

	// Everything selected: the numerator is every region that has a total,
	// which is now all of them.
	num, den := shareIn(sp, ds, options{})
	var wantNum int64
	for j, id := range sp.Regions {
		if _, ok := ds.RegionTotals[id]; ok {
			wantNum += sp.RegionCounts[j]
		}
	}
	if num != wantNum {
		t.Errorf("%s: numerator %d, want %d", sp.ScientificName, num, wantNum)
	}
	if den == 0 {
		t.Fatalf("%s: no denominator at all", sp.ScientificName)
	}

	// The invariant itself, exercised directly: a region withheld from the
	// totals contributes to neither side.
	partial := make(map[string]int64, len(ds.RegionTotals))
	for id, n := range ds.RegionTotals {
		partial[id] = n
	}
	var withheld string
	for _, id := range sp.Regions {
		if rg, ok := region.Get(id); ok && rg.Kind == region.Marine {
			withheld = id
			delete(partial, id)
			break
		}
	}
	n2, d2 := shareIn(sp, &dataset.Dataset{RegionTotals: partial,
		RegionMonthlyTotals: ds.RegionMonthlyTotals}, options{})
	if n2 >= num {
		t.Errorf("%s: dropping %s from the totals left the numerator at %d (was %d); "+
			"a region with no denominator must leave the numerator too",
			sp.ScientificName, withheld, n2, num)
	}
	if d2 >= den {
		t.Errorf("%s: dropping %s left the denominator at %d (was %d)",
			sp.ScientificName, withheld, d2, den)
	}
}

// TestShareStatesZeroRatherThanDashing asserts that a bird with no records in
// the regions on screen says so with a number, and that having no denominator
// at all is a different message naming the selection.
//
// The em-dash this replaced lost the half of the answer that carries the
// weight. "0 in 1,225,176,035 bird records across every land region" and "0
// in 900" are different findings, and only the stated denominator tells them
// apart -- a dash says neither, and cannot even be told from a figure that
// failed to render.
//
// Verifies: product.shareOfRecords. Criterion: acceptance.commonestIsLargestShare.
func TestShareStatesZeroRatherThanDashing(t *testing.T) {
	s := newTestServer(t)

	// The fixture heron is in no region at all, so it has nothing in any
	// selection and is the zero case on every view.
	_, body := get(t, s, "/")
	if strings.Contains(body, "&mdash;</span></div>") {
		t.Error("a share cell still renders a bare em-dash; absence is stated as a number")
	}
	if !regexp.MustCompile(`<b>0 in [\d,]+</b>`).MatchString(body) {
		t.Fatalf("no tile reads \"0 in N\"; the fixture heron is in no region and must")
	}

	// The denominator is the point, so it has to be the real one and not a
	// zero or a placeholder.
	m := regexp.MustCompile(`<b>0 in ([\d,]+)</b>`).FindStringSubmatch(body)
	if n, err := strconv.Atoi(strings.ReplaceAll(m[1], ",", "")); err != nil || n <= 0 {
		t.Errorf("zero share quotes a denominator of %q; it must say how hard anyone "+
			"looked before concluding the bird was absent", m[1])
	}

	// Seas now carry an honest denominator (decisions.honestSeaDenominator), so selecting one
	// yields a figure like anywhere else rather than an explanation. The
	// explaining branch survives for a region added ahead of its data, and
	// is exercised by TestShareRangesOverOneSetOfRegions.
	seas := make([]string, 0, 5)
	for _, id := range region.MarineIDs() {
		if rg, ok := region.Get(id); ok {
			seas = append(seas, "region="+rg.Slug())
		}
	}
	_, body = get(t, s, "/?"+strings.Join(seas, "&"))
	if !regexp.MustCompile(`<b>[01] in [\d,]+</b>`).MatchString(body) {
		t.Error("selecting the seas yields no figure; they have had a denominator since decisions.honestSeaDenominator")
	}
}

// TestCommonestRoundTripsThroughTheIndex asserts that the detail page's peak
// region and the index's share are the same arithmetic: select the region the
// detail page names, and the index must print the figure the detail page
// claimed.
//
// This is the check that would have caught the bug the owner reported. The
// two numbers came from different tables with different denominators and
// different membership rules, so `/species/yenalb2` could say "commonest in
// Maine" while `/?region=NorthAtlantic` said something that looked like a
// flat contradiction -- and nothing compared them, because each was tested
// against its own source.
//
// Over the real dataset, and over every species rather than a sample: the
// failure was not in the logic of either page but in the pairing, and a
// pairing is only wrong for the species where the two tables disagree.
//
// Verifies: product.shareOfRecords, decisions.honestSeaDenominator. Criterion: acceptance.peakAndShareOneCalculation.
func TestCommonestRoundTripsThroughTheIndex(t *testing.T) {
	ds := realDataset(t)
	var checked, atSea int
	for i := range ds.Species {
		sp := &ds.Species[i]
		c := sp.Commonest
		if c == nil || c.OneIn() == 0 {
			continue
		}
		checked++
		if rg, ok := region.Get(c.Region); ok && rg.Kind == region.Marine {
			atSea++
		}

		// The peak must be a region the bird is actually listed in. The old
		// figure could name one the reader would not find in the filter:
		// Maine holds 12 albatross records in 14 million, below the
		// presence bar, so the page named a place the index did not offer.
		if !slices.Contains(sp.Regions, c.Region) {
			t.Errorf("%s peaks in %s, which is not among the regions it is listed in",
				sp.ScientificName, c.Region)
			continue
		}

		// Select exactly that region on the index and read the share back.
		cell := shareOf(sp, ds, options{
			HasRegions: true, Regions: map[string]bool{c.Region: true}})
		if cell == nil {
			t.Errorf("%s: the index gives no share for %s, which its own page calls its peak",
				sp.ScientificName, c.Region)
			continue
		}
		want := "1 in " + comma(int64(c.OneIn()))
		if cell.Ratio != want {
			t.Errorf("%s in %s: detail page says %q, index says %q",
				sp.ScientificName, c.Region, want, cell.Ratio)
		}
	}
	if checked < len(ds.Species)/2 {
		t.Fatalf("only %d of %d species carry a peak; the check is passing over the dataset",
			checked, len(ds.Species))
	}
	if atSea == 0 {
		t.Error("no species peaks at sea, so the case decisions.honestSeaDenominator exists for is unexercised")
	}

	// The row's label, pinned. It has been three things -- "Commonest in",
	// then "Commonest land region" while the seas had no honest denominator
	// (decisions.seaShareDiffers), now "Top region" -- and nothing asserted any of them, so
	// each rename was unverified. The superseded forms are rejected as well
	// as the current one required: a stale qualifier is how "Commonest land
	// region" would have survived decisions.honestSeaDenominator giving the seas a denominator.
	// amerob, not greblu: the fixture heron is deliberately in no region and
	// carries no peak, so its page has no such row to label.
	_, page := get(t, newTestServer(t), "/species/amerob")
	if !strings.Contains(page, "Top region") {
		t.Error("the detail page does not label the peak row \"Top region\"")
	}
	for _, stale := range []string{"Commonest in", "Commonest land region", "Commonest region"} {
		if strings.Contains(page, stale) {
			t.Errorf("the detail page still says %q", stale)
		}
	}
	t.Logf("%d species round-tripped, %d of them peaking at sea", checked, atSea)
}

// TestMissingTotalsDoNotClaimTheBirdIsAbsent pins the distinction between the
// two empty answers, which read alike and mean opposite things.
//
//   - "0 in 1,225,176,035" is a finding about the BIRD: it was not recorded
//     in the regions on screen, and the denominator says how hard anyone
//     looked before concluding that.
//   - "No data for these regions yet" is a gap in BIRDVIEW: the region's own
//     record total is missing, so there is nothing to divide by.
//
// The second must never be worded as the first. A region with no total is
// skipped on BOTH sides of the ratio, so the numerator comes back zero too --
// and that zero is an artifact of the skip, not a count. Buller's Shearwater
// has 39,143 records in the North Pacific and would be reported as having
// none if that sea's total went missing, which is what the marine regions
// did for months.
//
// Verifies: product.shareOfRecords, decisions.honestSeaDenominator. Criterion: acceptance.commonestIsLargestShare.
func TestMissingTotalsDoNotClaimTheBirdIsAbsent(t *testing.T) {
	ds := realDataset(t)
	var sp *dataset.Species
	var id string
	var really int64
	for i := range ds.Species {
		for j, r := range ds.Species[i].Regions {
			if ds.Species[i].RegionCounts[j] > 1000 {
				sp, id, really = &ds.Species[i], r, ds.Species[i].RegionCounts[j]
				break
			}
		}
		if sp != nil {
			break
		}
	}
	if sp == nil {
		t.Fatal("no species has a region with over a thousand records")
	}

	// Withhold that region's total, as the seas were withheld.
	cell := shareOf(sp, &dataset.Dataset{RegionTotals: map[string]int64{}}, options{
		HasRegions: true, Regions: map[string]bool{id: true}})
	if cell == nil {
		t.Fatal("a missing total renders nothing at all")
	}
	if cell.Ratio != "" {
		t.Errorf("a missing total produced a ratio of %q, dividing by a denominator "+
			"BirdView does not have", cell.Ratio)
	}
	for _, wrong := range []string{"No records", "no records", "0 in", "not recorded", "absent"} {
		if strings.Contains(cell.Note+" "+cell.Title, wrong) {
			t.Errorf("a missing total is reported as %q, which says %q — but %s has %d "+
				"records in %s. That is a gap in BirdView's data stated as a fact about "+
				"the bird", cell.Note, wrong, sp.ScientificName, really, id)
		}
	}
	if cell.Note == "" {
		t.Error("a missing total explains nothing")
	}
	// And the word the owner found unintuitive is gone from the short label.
	if strings.Contains(cell.Note, "share") {
		t.Errorf("the label still reads %q; \"share\" is jargon on a one-line cell, and "+
			"the tooltips that do use it define it in the same sentence", cell.Note)
	}
}

// TestInterfaceNeverSaysShare asserts the vocabulary rule of product.shareOfRecords: the word
// is the right technical term and belongs in the spec, and a reader meeting
// a one-line cell has nowhere to learn it.
//
// Over the app's own pages AND the guides under /docs, because a guide is
// where a confused reader goes next and finding the jargon there teaches
// nothing. The normative documents under river/ are exempt: they define the
// term, and `presentation.share` is the subject this requirement is filed
// under.
//
// That exemption deliberately does not follow acceptance.noContactAddressPublished's precedent, where
// exempting the spec was found to be the wrong fix. An address published
// anywhere is harm wherever it appears; a defined term inside the document
// that defines it is not the same thing.
//
// Matches the NOUN, not the verb. "Bookmark it, share it", "need not share
// a habitat" and "one shared scale" are ordinary English in a different
// sense and are correct where they stand; the metric always appears as
// "share of" or behind a determiner, which is what this looks for.
//
// Verifies: product.shareOfRecords. Criterion: acceptance.commonestIsLargestShare.
func TestInterfaceNeverSaysShare(t *testing.T) {
	s := newTestServer(t)
	pages := []string{
		"/", "/?region=AK", "/?by=mass&scale=global",
		"/species/greblu", "/species/amerob", "/species/thimur",
		"/compare?species=amerob&species=greblu", "/added", "/docs/",
	}
	for _, m := range regexp.MustCompile(`href="(/docs/[^"#]+)"`).
		FindAllStringSubmatch(mustGet(t, s, "/docs/"), -1) {
		pages = append(pages, m[1])
	}
	if len(pages) < 11 {
		t.Fatalf("only %d pages to check; /docs/ is not listing its guides", len(pages))
	}

	noun := regexp.MustCompile(`(?i)\b(?:a|the|its|what|every|no|this|marine)\s+shares?\b|\bshares?\s+of\b`)
	var checked int
	for _, p := range pages {
		body := mustGet(t, s, p)
		checked++
		for _, m := range noun.FindAllString(body, -1) {
			t.Errorf("%s says %q; the interface explains the number instead of naming it (product.shareOfRecords)",
				p, strings.Join(strings.Fields(m), " "))
		}
	}
	if checked < 11 {
		t.Fatalf("only %d pages fetched", checked)
	}
}

// TestResponsesCompressWhenAsked asserts tech.compressedResponses, which exists for one reason:
// the index inlines an SVG per species and is megabytes of markup, and egress
// is the only axis of a serverless deployment that costs real money.
//
// Verifies: tech.compressedResponses. Criterion: acceptance.responsesCompress.
func TestResponsesCompressWhenAsked(t *testing.T) {
	s := newTestServer(t)
	fetch := func(path string, gzipOK bool) (*http.Response, []byte) {
		t.Helper()
		req := httptest.NewRequest(http.MethodGet, path, nil)
		if gzipOK {
			req.Header.Set("Accept-Encoding", "gzip")
		}
		w := httptest.NewRecorder()
		s.ServeHTTP(w, req)
		res := w.Result()
		body, err := io.ReadAll(res.Body)
		if err != nil {
			t.Fatalf("%s: %v", path, err)
		}
		return res, body
	}

	for _, path := range []string{"/", "/species/greblu", "/static/styles.css", "/docs/data"} {
		plainRes, plain := fetch(path, false)
		gzRes, packed := fetch(path, true)

		// Vary on BOTH, or a cache that stored the uncompressed reply would
		// go on serving it to clients that asked for gzip -- and worse, the
		// reverse, handing a gzip stream to a client that cannot read one.
		for _, res := range []*http.Response{plainRes, gzRes} {
			if v := res.Header.Get("Vary"); !strings.Contains(v, "Accept-Encoding") {
				t.Errorf("%s: Vary is %q, missing Accept-Encoding", path, v)
			}
		}
		if enc := plainRes.Header.Get("Content-Encoding"); enc != "" {
			t.Errorf("%s: compressed for a client that did not ask (%q)", path, enc)
		}
		if enc := gzRes.Header.Get("Content-Encoding"); enc != "gzip" {
			t.Errorf("%s: Content-Encoding %q, want gzip", path, enc)
			continue
		}
		// Content-Length describing the uncompressed body would make the
		// response unreadable; it must be gone, not merely wrong.
		if cl := gzRes.Header.Get("Content-Length"); cl != "" && cl != strconv.Itoa(len(packed)) {
			t.Errorf("%s: Content-Length %s against %d bytes sent", path, cl, len(packed))
		}

		zr, err := gzip.NewReader(bytes.NewReader(packed))
		if err != nil {
			t.Errorf("%s: body is not a gzip stream: %v", path, err)
			continue
		}
		got, err := io.ReadAll(zr)
		if err != nil {
			t.Errorf("%s: truncated gzip stream: %v", path, err)
			continue
		}
		if !bytes.Equal(got, plain) {
			t.Errorf("%s: decompressed body differs from the plain one (%d vs %d bytes)",
				path, len(got), len(plain))
		}
		// Logged, not asserted at a figure read off a run. The floor is loose
		// on purpose: what must not regress is that compression happens at
		// all, and markup this repetitive cannot fail to halve.
		saved := 100 - 100*float64(len(packed))/float64(len(plain))
		if saved < 50 {
			t.Errorf("%s: only %.0f%% saved (%d -> %d); markup this repetitive should "+
				"compress far better, so something is not being compressed",
				path, saved, len(plain), len(packed))
		}
		t.Logf("%-20s %8d -> %7d bytes, %.0f%% saved", path, len(plain), len(packed), saved)
	}

	// A status code set before the body must survive the wrapper, and the
	// body must still be readable: product.undrawnSpeciesAnswer's page is a 404 WITH content, which
	// is exactly the shape a careless wrapper turns into an empty reply.
	res, packed := fetch("/species/paspig", true)
	if res.StatusCode != http.StatusNotFound {
		t.Errorf("undrawn species: status %d through the compressor, want 404", res.StatusCode)
	}
	// The header, asserted BEFORE the body. A handler that calls WriteHeader
	// itself commits the header map before the first Write, so a compressor
	// that labels from Write sends a gzip body with no Content-Encoding and
	// the browser renders binary. Reading the body with gzip.NewReader does
	// not catch that -- the stream is valid, it is simply unannounced -- so
	// checking the body first would pass over exactly this bug. It did.
	if enc := res.Header.Get("Content-Encoding"); enc != "gzip" {
		t.Errorf("the 404 page is compressed but labelled %q; a browser would "+
			"render the gzip stream as binary", enc)
	}
	zr, err := gzip.NewReader(bytes.NewReader(packed))
	if err != nil {
		t.Fatalf("404 body is not a gzip stream: %v", err)
	}
	body, _ := io.ReadAll(zr)
	if !strings.Contains(string(body), "Passenger Pigeon") {
		t.Error("the 404 page lost its body through the compressor")
	}

	// A status that may carry no body must not carry a gzip header either.
	plainRes, _ := fetch("/", false)
	if plainRes.Header.Get("Content-Encoding") != "" {
		t.Error("an unasked-for response is labelled compressed")
	}
}

// TestFrequencyOrdersAndBandsByDecade asserts product.orderByFrequency over the real dataset:
// Frequency is a fourth ordering, commonest first, in tiers of powers of ten.
//
// Verifies: product.orderByFrequency, decisions.frequencyMetric. Criterion: acceptance.frequencyTiers.
func TestFrequencyOrdersAndBandsByDecade(t *testing.T) {
	ds := realDataset(t)
	srv, err := newServer(ds)
	if err != nil {
		t.Fatalf("newServer: %v", err)
	}
	heads := func(query string) []string {
		t.Helper()
		_, body := get(t, srv, query)
		var out []string
		for _, m := range regexp.MustCompile(`(?s)<h2>(.*?)</h2>`).FindAllStringSubmatch(body, -1) {
			txt := regexp.MustCompile(`<[^>]+>`).ReplaceAllString(m[1], " ")
			out = append(out, strings.Join(strings.Fields(html.UnescapeString(txt)), " "))
		}
		return out
	}

	got := heads("/?by=frequency")
	if len(got) < 5 {
		t.Fatalf("only %d sections; frequency spans 7.5 decades and should band into "+
			"about nine: %v", len(got), got)
	}
	// Every tier edge is a power of ten, and the tiers ascend. Checked by
	// reading the headers rather than the internals, because a reader who
	// cannot see the boundary in the header gets no benefit from it.
	edge := regexp.MustCompile(`1 in (\d[\d,]*) – 1 in (\d[\d,]*)`)
	var prev int64
	var banded int
	for _, h := range got {
		m := edge.FindStringSubmatch(h)
		if m == nil {
			continue
		}
		banded++
		lo := mustComma(t, m[1])
		hi := mustComma(t, m[2])
		if !isPowerOfTen(lo) {
			t.Errorf("tier starts at 1 in %s, which is not a power of ten: %q", m[1], h)
		}
		if !isPowerOfTen(hi + 1) {
			t.Errorf("tier ends at 1 in %s, which is not one below a power of ten: %q", m[2], h)
		}
		if lo <= prev {
			t.Errorf("tier starting at %d does not follow one starting at %d", lo, prev)
		}
		prev = lo
	}
	if banded < 5 {
		t.Fatalf("only %d banded tiers parsed from %v", banded, got)
	}

	// Membership, not just labels. The header is written by one function and
	// the banding done by another, so a tier can be labelled "1 in 1,000 -
	// 1 in 9,999" while holding whatever the sectioning rule actually put
	// there. Replacing the decade rule with a proportional one left every
	// label a tidy power of ten and only the tier COUNT gave it away.
	//
	// So: recompute each species' frequency here, independently, and check
	// that the birds in a tier are exactly the birds whose frequency falls
	// in the band the header claims.
	want := map[string]int{}
	for i := range ds.Species {
		sp := &ds.Species[i]
		if sp.Exotic != "" {
			continue // off the index (product.escapedBirdsListed), so in no tier
		}
		num, den := shareIn(sp, ds, options{})
		if num == 0 || den == 0 {
			continue
		}
		n := den / num
		lo := int64(1)
		for lo*10 <= n {
			lo *= 10
		}
		want[fmt.Sprintf("1 in %s – 1 in %s", comma(lo), comma(lo*10-1))]++
	}
	counts := regexp.MustCompile(`(1 in [\d,]+ – 1 in [\d,]+) (\d+) species`)
	var verified int
	for _, h := range got {
		m := counts.FindStringSubmatch(h)
		if m == nil {
			continue
		}
		verified++
		n, _ := strconv.Atoi(m[2])
		if want[m[1]] != n {
			t.Errorf("tier %q holds %d species; %d have a frequency in that band",
				m[1], n, want[m[1]])
		}
	}
	if verified < 5 {
		t.Fatalf("only %d tiers had their membership checked", verified)
	}

	// The birds with no ratio are kept, in ONE final section. They each had
	// a section of their own in the first version: Decade(+Inf) is +Inf and
	// Inf >= Inf, so every one of them started a new band.
	var noRatio int
	for _, h := range got {
		if strings.Contains(h, "No records in the regions selected") {
			noRatio++
		}
	}
	if noRatio != 1 {
		t.Errorf("%d sections for the birds with no records; want exactly one", noRatio)
	}

	// 100x halves the tier count, and the same control under a size metric
	// still means percent.
	if wide := heads("/?by=frequency&step=100"); len(wide) >= len(got) {
		t.Errorf("100x gave %d sections against 10x's %d; it should be coarser",
			len(wide), len(got))
	}

	// THE property the whole design is for: a tier edge is the same number
	// under every filter. A data-anchored band would start at whichever bird
	// is commonest in the selection, and "1 in 1,000 – 9,999" would denote
	// one set nationally and another in Alaska.
	ak := heads("/?by=frequency&region=AK")
	national := map[string]bool{}
	for _, h := range got {
		if m := edge.FindStringSubmatch(h); m != nil {
			national[m[0]] = true
		}
	}
	var checked int
	for _, h := range ak {
		m := edge.FindStringSubmatch(h)
		if m == nil {
			continue
		}
		checked++
		if !national[m[0]] {
			t.Errorf("Alaska has a tier %q that does not appear nationally; tier edges "+
				"must not move with the filter", m[0])
		}
	}
	if checked == 0 {
		t.Fatal("no banded tier under the Alaska filter; the check compared nothing")
	}
	t.Logf("%d tiers nationally, %d under Alaska, all edges shared", banded, checked)
}

func isPowerOfTen(n int64) bool {
	for n > 1 {
		if n%10 != 0 {
			return false
		}
		n /= 10
	}
	return n == 1
}

func mustComma(t *testing.T, s string) int64 {
	t.Helper()
	n, err := strconv.ParseInt(strings.ReplaceAll(s, ",", ""), 10, 64)
	if err != nil {
		t.Fatalf("parsing %q: %v", s, err)
	}
	return n
}

// TestFrequencyTiersAreNotNamed holds product.shareOfRecords's refusal in place: the metric is
// not a rarity tag, and a named tier is that threshold with a label on it.
//
// Cheap to write, and the thing most likely to be added later by someone
// trying to be helpful.
//
// Verifies: product.orderByFrequency, product.shareOfRecords. Criterion: acceptance.frequencyTiers.
func TestFrequencyTiersAreNotNamed(t *testing.T) {
	ds := realDataset(t)
	srv, err := newServer(ds)
	if err != nil {
		t.Fatalf("newServer: %v", err)
	}
	_, body := get(t, srv, "/?by=frequency")
	heads := regexp.MustCompile(`(?s)<h2>(.*?)</h2>`).FindAllString(body, -1)
	if len(heads) == 0 {
		t.Fatal("no section headers")
	}
	joined := strings.ToLower(strings.Join(heads, " "))
	for _, word := range []string{"common", "scarce", "rare", "abundant", "uncommon", "everyday"} {
		if strings.Contains(joined, word) {
			t.Errorf("a tier header says %q; product.shareOfRecords keeps this a number and not a rarity "+
				"tag, and a named band is the threshold it refuses", word)
		}
	}
}

// TestFrequencyDefaultsToGlobalScale asserts decisions.frequencyGlobalScale: the scale default
// depends on the ordering, and BOTH scale values stay selectable under it.
//
// The second half is the one that bites. The parser used to recognise only
// the non-default value, with a comment warning that this silently disables
// the toggle whenever the default moves -- which is what happened when
// decisions.perSectionDefault moved it. A default that varies by metric is the same bug arriving
// by a different road: read only Global, and "Per section" is a dead button
// under Frequency.
//
// Verifies: product.scaleToggle, decisions.frequencyGlobalScale. Criterion: acceptance.frequencyTiers.
func TestFrequencyDefaultsToGlobalScale(t *testing.T) {
	for _, tc := range []struct {
		query string
		want  view.ScaleMode
	}{
		{"/", view.PerSection},
		{"/?by=mass", view.PerSection},
		{"/?by=frequency", view.Global},
		// Both directions explicit, under both orderings.
		{"/?by=frequency&scale=section", view.PerSection},
		{"/?by=frequency&scale=global", view.Global},
		{"/?by=length&scale=global", view.Global},
		{"/?by=length&scale=section", view.PerSection},
	} {
		req := httptest.NewRequest(http.MethodGet, tc.query, nil)
		if got := parseOptions(req).Scale; got != tc.want {
			t.Errorf("%s: scale %q, want %q", tc.query, got, tc.want)
		}
	}

	// And it reaches the page: under Frequency the control must show Global
	// as the active one without the reader having asked.
	_, body := get(t, newTestServer(t), "/?by=frequency")
	global := regexp.MustCompile(`<a class="btn on"[^>]*scale=global`)
	if !global.MatchString(body) {
		section := regexp.MustCompile(`<a class="btn on"[^>]*scale=section`)
		t.Errorf("the scale control does not show Global active under Frequency "+
			"(Per section active: %v)", section.MatchString(body))
	}
}

// TestSharedNamesAreDisclosed asserts product.namesRecordsFiledUnder over the real dataset: where one
// occurrence name covers two species BirdView draws, both say so, and the
// disclosure reaches the index as well as the detail page.
//
// On the index too, because that is where the misreading happens. The bug
// that prompted this was visible as a ranking — a Caribbean mangrove
// resident sitting 51st most-recorded in North America — and a note on a
// page nobody opened would not have caught it.
//
// Verifies: product.namesRecordsFiledUnder, decisions.sharedCountShownOnBoth. Criterion: acceptance.sharedCountDisclosed.
func TestSharedNamesAreDisclosed(t *testing.T) {
	ds := realDataset(t)
	srv, err := newServer(ds)
	if err != nil {
		t.Fatalf("newServer: %v", err)
	}
	if len(taxonomy.SharedNames) == 0 {
		t.Fatal("no shared names declared; this check has nothing to assert")
	}

	for _, sn := range taxonomy.SharedNames {
		for _, sci := range sn.Covers {
			id := srv.idOf[sci]
			if id == "" {
				t.Errorf("%s is declared as sharing %s but is not drawn", sci, sn.Source)
				continue
			}
			other := sn.Covers[0]
			if other == sci {
				other = sn.Covers[1]
			}

			// The detail page names the source binomial and the other bird.
			_, body := get(t, srv, "/species/"+id)
			for _, want := range []string{"Recorded as", sn.Source, "covering two species"} {
				if !strings.Contains(body, want) {
					t.Errorf("/species/%s (%s) does not say %q", id, sci, want)
				}
			}
			if !strings.Contains(body, other) {
				t.Errorf("/species/%s does not name the species it shares with (%s)", id, other)
			}

			// And the index tile carries the seam, with the other bird's
			// common name in reach of a reader who never opens the page.
			_, idx := get(t, srv, "/")
			i := strings.Index(idx, id)
			if i < 0 {
				t.Errorf("%s is not on the index at all", id)
				continue
			}
			// The share cell sits after the tile's inline SVG, which runs to
			// thousands of characters, so seek to the cell rather than
			// guessing a window.
			rest := idx[i:]
			c := strings.Index(rest, `class="share"`)
			if c < 0 {
				t.Errorf("no share cell after %s on the index", id)
				continue
			}
			cell := rest[c:min(c+600, len(rest))]
			if !strings.Contains(cell, `class="seam"`) {
				t.Errorf("the index tile for %s (%s) carries no shared-count marker", id, sci)
			}
			if !strings.Contains(cell, "Shared with") {
				t.Errorf("the marker on %s's tile does not say what it is shared with", id)
			}
		}
	}

	// A withheld species is NOT reported as a naming disagreement. The two
	// are indistinguishable in the data and opposite in meaning: telling a
	// reader eBird calls the Spotted Owl something else would send them
	// after a name that does not exist.
	for sci := range taxonomy.WithheldByEBird {
		common := sci
		id := srv.idOf[sci]
		if id == "" {
			continue // out of scope is fine; several are declared exceptions
		}
		_, body := get(t, srv, "/species/"+id)
		if !strings.Contains(body, "withheld by eBird") {
			t.Errorf("/species/%s (%s) does not say eBird withholds it", id, common)
		}
		if strings.Contains(body, "covering two species") {
			t.Errorf("/species/%s (%s) is withheld, not shared, but is reported as shared",
				id, common)
		}
	}

	// An ordinary species carries neither, or the disclosure means nothing.
	_, plain := get(t, srv, "/species/amerob")
	for _, unwanted := range []string{"Recorded as", "covering two species", "withheld by eBird"} {
		if strings.Contains(plain, unwanted) {
			t.Errorf("American Robin's page says %q; its records are its own", unwanted)
		}
	}
}

// TestOrderByStatesItsDirection asserts product.ascendingOrder's second half: the page says
// which way each ordering runs, and says it for both groups at once.
//
// Frequency ascends a reciprocal — "1 in N" rises as the bird gets scarcer
// — so it runs most-common-first while the three size metrics run
// smallest-first. Three renames were proposed to dissolve that and all
// three were declined (decisions.frequencyMetric): every frequency-polarity name points the
// wrong way, every scarcity-polarity name overclaims, and the one word
// right on both counts is one readers stumble over. What ships instead is
// the direction, captioned under each group.
//
// Both captions are present on every view, whichever metric is selected. A
// caption that changed with the selection would tell a reader what they had
// already chosen; two standing captions tell them the controls behave
// differently before they click, which is the thing worth knowing.
//
// Verifies: product.ascendingOrder, product.orderByFrequency. Criterion: acceptance.frequencyTiers.
func TestOrderByStatesItsDirection(t *testing.T) {
	s := newTestServer(t)
	for _, query := range []string{"/", "/?by=length", "/?by=mass", "/?by=frequency"} {
		_, body := get(t, s, query)
		i := strings.Index(body, `class="lbl">Order by`)
		if i < 0 {
			t.Fatalf("%s: no Order by row; the selector is broken, not the page", query)
		}
		row := body[i:min(i+2000, len(body))]

		groups := strings.Split(row, `class="dirbtns"`)
		if len(groups) != 3 {
			t.Errorf("%s: %d button groups, want 2 — the size metrics and Frequency",
				query, len(groups)-1)
			continue
		}
		// The three size metrics share one caption; Frequency has its own.
		// Asserted by COUNTING buttons per group, so a metric silently
		// moving between them fails rather than passing on the strings.
		if n := strings.Count(groups[1][:strings.Index(groups[1], "</span>")], `class="btn`); n != 3 {
			t.Errorf("%s: first group holds %d buttons, want 3", query, n)
		}
		if n := strings.Count(groups[2][:strings.Index(groups[2], "</span>")], `class="btn`); n != 1 {
			t.Errorf("%s: second group holds %d buttons, want 1", query, n)
		}

		caps := regexp.MustCompile(`class="dirn">([^<]+)<`).FindAllStringSubmatch(row, -1)
		if len(caps) != 2 {
			t.Errorf("%s: %d direction captions, want 2", query, len(caps))
			continue
		}
		if got := strings.TrimSpace(caps[0][1]); got != "smallest first" {
			t.Errorf("%s: size metrics captioned %q", query, got)
		}
		if got := strings.TrimSpace(caps[1][1]); got != "most common first" {
			t.Errorf("%s: Frequency captioned %q", query, got)
		}
	}
}

// TestOrderingRowAlignsOnTheBaseline pins the one declaration that keeps the
// captioned groups from dragging their row out of line.
//
// The groups are two lines tall and everything else on that row is one, so
// under the stylesheet's default `align-items: center` the row's height
// follows the captions and ORDER BY sits level with the rules instead of
// with the buttons. Baseline alignment puts the label, the three size
// buttons and the Frequency button on one line and lets the captions hang
// below it.
//
// Asserted because it is invisible to every other check here: the markup is
// identical either way, the contrast tests do not read layout, and nothing
// renders pixels. A one-line CSS deletion would put it back and no test
// would notice.
//
// Verifies: product.ascendingOrder. Criterion: acceptance.frequencyTiers.
func TestOrderingRowAlignsOnTheBaseline(t *testing.T) {
	b, err := web.Static.ReadFile("static/styles.css")
	if err != nil {
		t.Fatalf("reading the stylesheet: %v", err)
	}
	css := string(b)
	if !regexp.MustCompile(`\.row\.ordering\s*\{[^}]*align-items:\s*baseline`).MatchString(css) {
		t.Error("the ordering row does not align on the baseline; its two-line groups will " +
			"drag ORDER BY down level with the rules")
	}
	// And the row has to carry the class, or the rule applies to nothing.
	_, body := get(t, newTestServer(t), "/")
	if !strings.Contains(body, `class="row ordering"`) {
		t.Error("no row carries the ordering class, so the alignment rule matches nothing")
	}
	// The caption still hangs below its own group rather than floating in
	// the row: it must be inside a .dirgroup, not a sibling of the buttons.
	i := strings.Index(body, `class="dirgroup"`)
	if i < 0 {
		t.Fatal("no direction group on the page")
	}
	group := body[i:min(i+700, len(body))]
	if strings.Index(group, `class="dirbtns"`) > strings.Index(group, `class="dirn"`) {
		t.Error("the caption precedes its buttons; it should hang below them")
	}
}

// TestTheDetailPageSaysEverythingTheTileSays asserts product.tileOpensDetail's superset rule
// on the one field that broke it: the frequency figure.
//
// product.tileOpensDetail has always described the detail view as a larger version of the
// tile "plus" three additions. That was read as permission to add to the
// detail page rather than as a floor under it, so when product.shareOfRecords's figure went
// onto every tile it never reached the detail page at all. A reader saw
// `1 in 23,393 species observations (all regions)` on the Northern Yellow
// Warbler's tile, opened the bird, and found `Top region -- 1 in 2,986
// there`: a different number, in the same unit, for a scope they had not
// asked about.
//
// It runs under two scopes because the figure follows the region filter.
// Carrying the figure but not the selection would be the worse bug of the
// two -- a number that silently changes what it ranges over between one
// page and the next, with nothing on either page revealing the swap.
//
// Watched fail: see acceptance.detailSaysWhatTileSays.
//
// Verifies: product.tileOpensDetail, product.shareOfRecords. Criterion: acceptance.detailSaysWhatTileSays.
func TestTheDetailPageSaysEverythingTheTileSays(t *testing.T) {
	s, err := newServer(realDataset(t))
	if err != nil {
		t.Fatalf("newServer: %v", err)
	}
	norm := func(h string) string {
		return strings.Join(strings.Fields(stripTags(h)), " ")
	}
	tile := regexp.MustCompile(`(?s)id="s-([a-z0-9]+)".*?<div class="share">(.*?)</div>`)
	span := regexp.MustCompile(`(?s)<span[^>]*>(.*?)</span>`)

	seen := map[string]string{}
	// min is per scope: unfiltered is the whole dataset, Alaska a large
	// subset of it. A single floor would have to be the smaller of the two
	// and would stop guarding the bigger case.
	for _, scope := range []struct {
		query, want string
		min         int
	}{
		{"", "(all regions)", 1000},
		{"?region=AK", "in Alaska", 300},
	} {
		_, index := get(t, s, "/"+scope.query)
		tiles := tile.FindAllStringSubmatch(index, -1)
		// A parse that silently matched nothing would make every assertion
		// below vacuous, which is the shape this project keeps paying for.
		if len(tiles) < scope.min {
			t.Fatalf("%q: parsed %d tiles off the index, want at least %d",
				scope.query, len(tiles), scope.min)
		}
		checked := 0
		for i, m := range tiles {
			if i%40 != 0 {
				continue
			}
			id := m[1]
			// Each span in the cell is its own statement -- the figure, then
			// any caveats -- and the detail page may set them apart, so
			// each is looked for on its own rather than as one run of text.
			var statements []string
			for _, sm := range span.FindAllStringSubmatch(m[2], -1) {
				if st := norm(sm[1]); st != "" {
					statements = append(statements, st)
				}
			}
			if len(statements) == 0 {
				t.Errorf("%s: tile carries no figure at all", id)
				continue
			}
			_, page := get(t, s, "/species/"+id+scope.query)
			body := norm(page)
			for _, figure := range statements {
				if !strings.Contains(body, figure) {
					t.Errorf("%s%s: the tile says %q and the detail page does not. "+
						"product.tileOpensDetail makes the detail view a superset of the tile: a reader "+
						"who opens a bird to learn more must not find less",
						id, scope.query, figure)
				}
			}
			// And it is reported under the scope the reader selected, not
			// quietly swapped for a different one. The figure is the first
			// statement; the caveats after it name no scope.
			figure := statements[0]
			if !strings.Contains(figure, scope.want) {
				t.Errorf("%s: tile figure %q does not name the scope %q",
					id, figure, scope.want)
			}
			if i == 0 {
				seen[scope.query] = figure
			}
			checked++
		}
		if checked < 10 {
			t.Fatalf("%q: checked only %d species", scope.query, checked)
		}
	}
	// The two scopes must actually differ, or both halves above could be
	// satisfied by a page ignoring the filter entirely.
	if seen[""] == seen["?region=AK"] {
		t.Errorf("the same bird reads %q unfiltered and in Alaska; the figure "+
			"is not following the region selection on either page", seen[""])
	}
}

// TestEveryNameDisagreementIsOnThePage asserts product.namesRecordsFiledUnder over all four kinds,
// in both directions: every species whose name the occurrence data disputes
// says so, and no species that nobody disputes says anything.
//
// Only two of the four were implemented. A shared name and an eBird
// suppression reached the page; an alias and a lump did not, though product.namesRecordsFiledUnder
// has named the alias case since it was written -- "the page names the
// alias so a reader can find the bird in the source". A visitor who knows
// the Hairy Woodpecker as Dryobates villosus and finds a page headed
// Leuconotopicus villosus could not tell a taxonomic update from a mistake.
//
// The second direction is what stops this passing by putting the row on
// every page. 27 species of 1,083 are in one of these tables, and the other
// 1,056 must stay silent: a disclosure that fires everywhere says nothing
// and trains a reader to skip it.
//
// Watched fail: see acceptance.disputedNamesShown.
//
// Verifies: product.namesRecordsFiledUnder. Criterion: acceptance.disputedNamesShown.
func TestEveryNameDisagreementIsOnThePage(t *testing.T) {
	ds := realDataset(t)
	s, err := newServer(ds)
	if err != nil {
		t.Fatalf("newServer: %v", err)
	}
	// What each affected species' page must name, keyed by AviList binomial.
	want := map[string][]string{}
	for aviList, old := range taxonomy.AviListRenamedFrom {
		want[aviList] = append(want[aviList], old)
	}
	for _, c := range taxonomy.CountedAs {
		want[c.Species] = append(want[c.Species], c.Published)
	}
	for _, sn := range taxonomy.SharedNames {
		for _, covered := range sn.Covers {
			want[covered] = append(want[covered], sn.Source)
		}
	}
	for aviList := range taxonomy.DomesticForms {
		want[aviList] = append(want[aviList], "domestic and feral")
	}
	for aviList := range taxonomy.WithheldByEBird {
		want[aviList] = append(want[aviList], "withheld by eBird")
	}
	// Withheld in particular regions or seasons (decisions.ebirdSilenceNotAbsence): read off the
	// dataset, because whether a listing applies is decided against the
	// records at build time, not by the list alone.
	for i := range ds.Species {
		sp := &ds.Species[i]
		if taxonomy.WithheldByEBird[sp.ScientificName] || len(sp.RegionWithheld) == 0 {
			continue
		}
		want[sp.ScientificName] = append(want[sp.ScientificName], "withheld by eBird in")
	}
	if len(want) < 25 {
		t.Fatalf("only %d species to check; the tables are nearly empty and this "+
			"check asserts almost nothing", len(want))
	}

	id := map[string]string{}
	for i := range ds.Species {
		id[ds.Species[i].ScientificName] = ds.Species[i].ID
	}
	row := regexp.MustCompile(`(?s)<th>Recorded as</th>.*?</tr>(?:\s*<tr class="[^"]*cont">.*?</tr>)*`)

	checked := 0
	for aviList, musts := range want {
		sid, drawn := id[aviList]
		if !drawn {
			// In a table but not drawn -- a withheld Hawaiian endemic that
			// did not make scope. Nothing to assert about a page that does
			// not exist.
			continue
		}
		_, page := get(t, s, "/species/"+sid)
		seam := row.FindString(page)
		if seam == "" {
			t.Errorf("%s (%s): the occurrence data disputes this name and the page is "+
				"silent; product.namesRecordsFiledUnder requires it to say which name the records were filed under",
				aviList, sid)
			continue
		}
		for _, must := range musts {
			if !strings.Contains(stripTags(seam), must) {
				t.Errorf("%s (%s): the disclosure does not mention %q. It reads: %s",
					aviList, sid, must, strings.Join(strings.Fields(stripTags(seam)), " "))
			}
		}
		checked++
	}
	if checked < 20 {
		t.Fatalf("checked only %d species", checked)
	}

	// And the other direction: a bird nobody disputes says nothing. Sampled,
	// because rendering all 1,056 is a second of nothing useful.
	quiet := 0
	for i := range ds.Species {
		sp := &ds.Species[i]
		if _, disputed := want[sp.ScientificName]; disputed || i%37 != 0 {
			continue
		}
		_, page := get(t, s, "/species/"+sp.ID)
		if row.MatchString(page) {
			t.Errorf("%s (%s) is in none of the naming tables and still carries a "+
				"\"Recorded as\" row; a disclosure that fires on every page says nothing",
				sp.ScientificName, sp.ID)
		}
		quiet++
	}
	if quiet < 15 {
		t.Fatalf("checked only %d undisputed species", quiet)
	}
}

// TestResponsesCarryAValidator asserts tech.cacheLifetimeStated: every response says how long
// it may be reused and gives a cache something to check it against.
//
// BirdView sent no Cache-Control, no ETag and no Last-Modified at all,
// which does not mean "do not cache". It means the browser may apply its
// own heuristic and has nothing to revalidate against when it does, so a
// reader who reloads after a rebuild can be served the old page with no
// way to tell and no way to force it. That cost a review round on a page
// that was correct on the server.
//
// Watched fail: see acceptance.honestRevalidation.
//
// Verifies: tech.cacheLifetimeStated. Criterion: acceptance.honestRevalidation.
func TestResponsesCarryAValidator(t *testing.T) {
	s, err := newServer(realDataset(t))
	if err != nil {
		t.Fatalf("newServer: %v", err)
	}
	paths := []string{"/", "/species/yelwar1", "/species/haiwoo", "/?region=AK",
		"/static/styles.css"}

	tags := map[string]string{}
	for _, p := range paths {
		w := httptest.NewRecorder()
		s.ServeHTTP(w, httptest.NewRequest("GET", p, nil))
		if w.Code != 200 {
			t.Fatalf("%s: %d", p, w.Code)
		}
		et := w.Header().Get("ETag")
		if et == "" {
			t.Errorf("%s: no ETag, so a cache has nothing to revalidate against "+
				"and must either refetch everything or guess", p)
		}
		if cc := w.Header().Get("Cache-Control"); cc != "no-cache" {
			t.Errorf("%s: Cache-Control is %q, want \"no-cache\" — store it, but ask "+
				"before reusing it", p, cc)
		}
		for q, seen := range tags {
			if seen == et {
				t.Errorf("%s and %s share the ETag %s; a cache would serve one for "+
					"the other", p, q, et)
			}
		}
		tags[p] = et
	}

	// A validator is only worth sending if it actually saves the body.
	for _, p := range paths {
		r := httptest.NewRequest("GET", p, nil)
		r.Header.Set("If-None-Match", tags[p])
		w := httptest.NewRecorder()
		s.ServeHTTP(w, r)
		if w.Code != http.StatusNotModified {
			t.Errorf("%s: revalidating with its own ETag returned %d, not 304; the "+
				"ETag costs a header and saves nothing", p, w.Code)
		}
		if w.Body.Len() != 0 {
			t.Errorf("%s: 304 carried %d bytes of body", p, w.Body.Len())
		}
	}

	// The ETag is computed from the build rather than from the body, which
	// is sound only if a page is a pure function of what the build covers.
	// Asserted, not assumed: a clock anywhere in the render path would make
	// every cached page wrong in a way nothing else here would catch.
	for _, p := range paths {
		var first string
		for i := 0; i < 3; i++ {
			w := httptest.NewRecorder()
			s.ServeHTTP(w, httptest.NewRequest("GET", p, nil))
			if i == 0 {
				first = w.Body.String()
				continue
			}
			if w.Body.String() != first {
				t.Errorf("%s renders differently on run %d with nothing changed; the "+
					"ETag is a fingerprint of the build and would now be a lie", p, i)
			}
		}
	}

	// And it moves when the build does. Without this the whole scheme could
	// be satisfied by a constant.
	ds2 := realDataset(t)
	ds2.Species[0].Mass.Value += 1
	s2, err := newServer(ds2)
	if err != nil {
		t.Fatalf("newServer: %v", err)
	}
	w := httptest.NewRecorder()
	s2.ServeHTTP(w, httptest.NewRequest("GET", "/", nil))
	if w.Header().Get("ETag") == tags["/"] {
		t.Errorf("the index keeps ETag %s after the dataset changed; every reader "+
			"with a cached copy would keep it", tags["/"])
	}
}

// TestTheNameDisclosureExplainsItselfAndLinksSomewhereUseful asserts the two
// things a reader needs from a `Recorded as` row beyond the bare fact: that
// the difference between the two names is legible, and that the link under
// it lands on the section explaining it.
//
// Both were wrong. Every row linked to "how common a bird is", the section
// about the frequency figure, rather than to the one about name
// disagreements -- the label said "how names are joined" and the anchor
// went elsewhere. And the rename text read "AviList moved this species to
// Pica nuttallii and eBird has not followed, so BirdView looks the records
// up under Pica nuttalli", which for a one-letter difference looks like the
// same name printed twice, and says "moved" about a species whose genus did
// not change.
//
// Watched fail: see acceptance.nameDisclosureLegible.
//
// Verifies: product.namesRecordsFiledUnder. Criterion: acceptance.nameDisclosureLegible.
func TestTheNameDisclosureExplainsItselfAndLinksSomewhereUseful(t *testing.T) {
	s, err := newServer(realDataset(t))
	if err != nil {
		t.Fatalf("newServer: %v", err)
	}
	// The anchor the rows point at must exist in the page they point into.
	_, docs := get(t, s, "/docs/data")
	const anchor = "when-the-sources-disagree-about-a-name"
	if !strings.Contains(docs, `id="`+anchor+`"`) {
		t.Fatalf("/docs/data has no #%s to land on", anchor)
	}

	id := map[string]string{}
	for i := range s.ds.Species {
		id[s.ds.Species[i].ScientificName] = s.ds.Species[i].ID
	}
	row := regexp.MustCompile(`(?s)<th>Recorded as</th>.*?</tr>`)
	checked := 0
	for aviList, old := range taxonomy.AviListRenamedFrom {
		sid, drawn := id[aviList]
		if !drawn {
			continue
		}
		_, page := get(t, s, "/species/"+sid)
		seam := stripTags(row.FindString(page))
		if seam == "" {
			t.Errorf("%s: no disclosure", aviList)
			continue
		}
		if !strings.Contains(page, "/docs/data#"+anchor) {
			t.Errorf("%s: the row links somewhere other than #%s; its own label says "+
				"\"how names are joined\"", aviList, anchor)
		}
		// eBird's name in full, because that is the actionable half: it is
		// what a reader types back into eBird to check. AviList's is the
		// page heading, so the row names only the part that differs, which
		// the genus/epithet assertions below cover.
		if !strings.Contains(seam, old) {
			t.Errorf("%s: the disclosure never gives eBird's name %q in full, so a "+
				"reader cannot go and look the bird up", aviList, old)
		}
		oldGenus, oldEp, _ := strings.Cut(old, " ")
		newGenus, newEp, _ := strings.Cut(aviList, " ")
		switch {
		case oldGenus == newGenus:
			// Nothing moved, so the text must not say it did, and it has to
			// put the two epithets where they can be compared.
			if strings.Contains(seam, "moved") {
				t.Errorf("%s: same genus (%s) and the text says \"moved\": %s",
					aviList, newGenus, strings.Join(strings.Fields(seam), " "))
			}
			if !strings.Contains(seam, "epithet") {
				t.Errorf("%s: differs from %s only in the epithet (%s against %s) and the "+
					"text never says so, which reads as the same name printed twice",
					aviList, old, newEp, oldEp)
			}
		default:
			if !strings.Contains(seam, "genus") {
				t.Errorf("%s: a genus move from %s and the text does not say so",
					aviList, oldGenus)
			}
		}
		// It says eBird, which is the only claim true of all seven.
		if !strings.Contains(seam, "eBird") {
			t.Errorf("%s: the disclosure does not say whose name it is", aviList)
		}
		checked++
	}
	if checked < 5 {
		t.Fatalf("checked only %d renamed species", checked)
	}

	// A boundary disagreement names BOTH databases, and does not call
	// itself a lump unless it is one. Two of the three are splits eBird
	// has not adopted and only Saxicola is a lump in the taxonomic sense,
	// so one sentence covering all three named the wrong event for two.
	for _, c := range taxonomy.CountedAs {
		sid, drawn := id[c.Species]
		if !drawn {
			continue
		}
		_, page := get(t, s, "/species/"+sid)
		seam := strings.Join(strings.Fields(stripTags(row.FindString(page))), " ")
		for _, src := range []string{"eBird", "iNaturalist"} {
			if !strings.Contains(seam, src) {
				t.Errorf("%s: the disclosure does not say what %s holds, so a reader "+
					"cannot tell which database disagrees with which: %s",
					c.Species, src, seam)
			}
		}
		isLump := c.Species == "Saxicola maurus"
		if strings.Contains(seam, "lump") && !isLump {
			t.Errorf("%s: calls itself a lump, but AviList SPLIT this species and eBird "+
				"did not follow — a lump is the opposite event: %s", c.Species, seam)
		}
		if strings.Contains(seam, "does not carry") && c.Published != "Saxicola stejnegeri" {
			t.Errorf("%s: says AviList does not carry %s, which it does — as a separate "+
				"species that is not this one: %s", c.Species, c.Published, seam)
		}
	}
}

// TestMonthScopeWraps asserts the label a figure carries when the months
// are narrowed, and the one case a naive renderer gets wrong.
//
// A run of three or more collapses to a range, and a range is allowed to
// cross the year end: a winter selection reads "Nov-Feb" and not
// "Jan, Feb, Nov, Dec". Winter is the selection a birder is most likely to
// make, so it is the one worth the extra code — and it is unreachable by
// any renderer that walks January to December and stops.
//
// Watched fail: see acceptance.monthsFilterAndAbstain.
//
// Verifies: product.regionAndMonthFilter, decisions.monthPresence. Criterion: acceptance.monthsFilterAndAbstain.
func TestMonthScopeWraps(t *testing.T) {
	mk := func(ms ...int) options {
		o := options{HasMonths: true, Months: map[int]bool{}}
		for _, m := range ms {
			o.Months[m] = true
		}
		return o
	}
	cases := []struct {
		in   options
		want string
	}{
		// The whole year is the default scope and saying so is noise.
		{options{}, ""},
		{mk(0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11), ""},
		{mk(5, 6, 7), "Jun-Aug"},
		{mk(5), "Jun"},
		// Two is not a run: "Jun-Jul" saves one character and reads as a
		// range the reader has to check.
		{mk(5, 6), "Jun, Jul"},
		// The wrap, and the reason this function is not three lines.
		{mk(10, 11, 0, 1), "Nov-Feb"},
		{mk(11, 0), "Dec, Jan"},
		{mk(0, 1, 2, 6, 7, 8), "Jan-Mar, Jul-Sep"},
		{mk(11), "Dec"},
		{mk(0, 2, 4, 6, 8, 10), "Jan, Mar, May, Jul, Sep, Nov"},
	}
	for _, c := range cases {
		if got := monthScope(c.in); got != c.want {
			t.Errorf("monthScope(%v) = %q, want %q", selectedMonths(c.in.Months), got, c.want)
		}
	}
}

func selectedMonths(m map[int]bool) []string {
	var out []string
	for i := 0; i < month.Count; i++ {
		if m[i] {
			out = append(out, month.Abbrev[i])
		}
	}
	return out
}

// TestAllTwelveMonthsIsTheWholeYear asserts the invariant decisions.monthPresence rests on:
// selecting every month must reproduce the figure for no month selected,
// exactly, for every species.
//
// Every occurrence record carries a month — 65,504,204 checked across both
// sources, none missing — so the twelve sum to the year by construction and
// any drift means the monthly table lost records the annual one kept. That
// failure would be nearly invisible: every figure would simply read a
// little scarcer than it should, with nothing on the page to say so.
//
// It compares the rendered ratio and not the internal sums, because the
// server has two paths — one that reads RegionCounts and one that walks
// RegionMonths — and the point is that a reader cannot tell them apart.
//
// Watched fail: see acceptance.monthsFilterAndAbstain.
//
// Verifies: product.regionAndMonthFilter, decisions.monthPresence. Criterion: acceptance.monthsFilterAndAbstain.
func TestAllTwelveMonthsIsTheWholeYear(t *testing.T) {
	ds := realDataset(t)
	every := options{HasMonths: true, Months: map[int]bool{}}
	for i := 0; i < month.Count; i++ {
		every.Months[i] = true
	}

	checked := 0
	for i := range ds.Species {
		sp := &ds.Species[i]
		if len(sp.Regions) == 0 {
			continue
		}
		year := shareOf(sp, ds, options{})
		twelve := shareOf(sp, ds, every)
		if year == nil || twelve == nil {
			t.Fatalf("%s: no cell", sp.ScientificName)
		}
		if year.Ratio != twelve.Ratio {
			t.Errorf("%s: the whole year reads %q and all twelve months read %q; "+
				"the monthly counts do not add up to the annual one",
				sp.ScientificName, year.Ratio, twelve.Ratio)
		}
		checked++
	}
	if checked < 1000 {
		t.Fatalf("checked only %d species", checked)
	}

	// And the sums themselves, over the denominator every tile divides by.
	var annual, monthly int64
	for id, n := range ds.RegionTotals {
		annual += n
		for _, v := range ds.RegionMonthlyTotals[id] {
			monthly += v
		}
	}
	if annual != monthly {
		t.Errorf("region totals sum to %d a year and %d over twelve months, a difference "+
			"of %d; every figure on the site divides by this", annual, monthly, annual-monthly)
	}

	// BOTH halves of the ratio must narrow, and the denominator is the half
	// that fails silently. A month-blind denominator divides June's records
	// by the whole year's birding, so every figure reads about twelve times
	// scarcer with nothing on the page to say why -- and it survives the
	// comparison above, because with all twelve selected the two agree.
	sp := &ds.Species[0]
	for i := range ds.Species {
		if len(ds.Species[i].Regions) > 20 {
			sp = &ds.Species[i]
			break
		}
	}
	var numSum, denSum int64
	for m := 0; m < month.Count; m++ {
		one := options{HasMonths: true, Months: map[int]bool{m: true}}
		num, den := shareIn(sp, ds, one)
		numSum += num
		denSum += den
		var want int64
		for id := range ds.RegionTotals {
			want += ds.RegionMonthlyTotals[id][m]
		}
		if den != want {
			t.Errorf("%s in %s: divided by %d, want %d — that month's birding, not "+
				"another period's", sp.ScientificName, month.Full[m], den, want)
		}
	}
	yearNum, yearDen := shareIn(sp, ds, options{})
	if numSum != yearNum || denSum != yearDen {
		t.Errorf("%s: the twelve single-month figures sum to %d/%d and the year is %d/%d",
			sp.ScientificName, numSum, denSum, yearNum, yearDen)
	}
}

// TestMonthsNarrowTheListAndTheFigureTogether asserts that months behave
// like every other filter — they decide which birds are drawn — and that
// the figure they produce is scoped to the same months.
//
// The owner asked for the first explicitly: "months should filter which
// diagrams are shown, not just record counts, as this is how all the other
// filters work." A control that silently did something else would be the
// worse surprise.
//
// Watched fail: see acceptance.monthsFilterAndAbstain.
//
// Verifies: product.regionAndMonthFilter, product.shareOfRecords, decisions.monthPresence. Criterion: acceptance.monthsFilterAndAbstain.
func TestMonthsNarrowTheListAndTheFigureTogether(t *testing.T) {
	s := newTestServerReal(t)

	count := func(path string) int {
		_, body := get(t, s, path)
		return strings.Count(body, `class="tilewrap"`)
	}
	all := count("/")
	june := count("/?month=Jun")
	jan := count("/?month=Jan")
	if all < 1000 {
		t.Fatalf("the unfiltered index shows %d tiles", all)
	}
	if june >= all {
		t.Errorf("June shows %d of %d species; a month that admits everything is "+
			"not filtering", june, all)
	}
	if june == jan {
		t.Errorf("June and January both show %d species; the filter is not reading "+
			"which month was asked for", june)
	}
	// Every month selected is the same as none, which is what makes All a
	// toggle rather than a third state.
	months := make([]string, 0, month.Count)
	for i := 0; i < month.Count; i++ {
		months = append(months, "month="+month.Abbrev[i])
	}
	if n := count("/?" + strings.Join(months, "&")); n != all {
		t.Errorf("all twelve months show %d species and the whole year shows %d", n, all)
	}

	// The figure narrows with the list, and says which months it covers.
	_, body := get(t, s, "/?region=NY&month=Jun&month=Jul&month=Aug")
	if !strings.Contains(body, "Jun-Aug") {
		t.Error("a three-month selection does not say so on the tiles")
	}
	if strings.Contains(body, "(all regions)") {
		t.Error("the scope still reads (all regions) under a region filter")
	}
}

// newTestServerReal builds a server over the compiled dataset, or skips.
func newTestServerReal(t *testing.T) *server {
	t.Helper()
	s, err := newServer(realDataset(t))
	if err != nil {
		t.Fatalf("newServer: %v", err)
	}
	return s
}

// TestAbstainedTilesSayWhy asserts that a bird the month filter declined to
// judge says so, with the reason, on its tile and on its page.
//
// Millerbird is the case decisions.monthPresence was written around: 28 records in Hawaii,
// all in five months, from a bird that lives on Nihoa the year round. Under
// a June filter it is shown, because hiding it would turn a fact about the
// survey into a statement about the bird -- and the tile has to say why it
// is there. The first wording was "shown in every month", which reads as a
// claim that the bird is present all year: the very misreading the
// abstention exists to prevent. The reason belongs on the tile, not only
// in a tooltip.
//
// Watched fail: see acceptance.monthsFilterAndAbstain.
//
// Verifies: product.regionAndMonthFilter, product.tileOpensDetail, decisions.monthPresence. Criterion: acceptance.monthsFilterAndAbstain.
func TestAbstainedTilesSayWhy(t *testing.T) {
	s := newTestServerReal(t)
	const path = "/?region=HI&month=Jun"
	_, body := get(t, s, path)

	tileOf := func(id string) string {
		i := strings.Index(body, `id="s-`+id+`"`)
		if i < 0 {
			t.Fatalf("%s: %s is not on the page", path, id)
		}
		seg := body[i:]
		if j := strings.Index(seg[1:], `class="tilewrap"`); j > 0 {
			seg = seg[:j]
		}
		return strings.Join(strings.Fields(stripTags(seg)), " ")
	}
	const marker = "too few records"

	// The abstaining bird is shown, and says why.
	if tile := tileOf("miller"); !strings.Contains(tile, marker) {
		t.Errorf("Millerbird is shown under %s without saying why; its tile reads: %s",
			path, tile)
	}
	// A bird with plenty of evidence is judged on it and carries no marker,
	// or the marker would be noise on every tile and say nothing.
	if tile := tileOf("warwhe1"); strings.Contains(tile, marker) {
		t.Errorf("Warbling White-eye, with %s, carries the too-few-records marker: %s",
			"over 140,000 Hawaii records", tile)
	}
	// And the page says it too, because anything a tile shows the detail
	// page must (product.tileOpensDetail).
	_, page := get(t, s, "/species/miller"+strings.TrimPrefix(path, "/"))
	if !strings.Contains(stripTags(page), marker) {
		t.Error("Millerbird's detail page, opened under the same filter, does not say " +
			"the month filter abstained")
	}
}

// TestComparisonOffersAWayOutOfPickMode asserts the second exit from the
// comparison view: one that returns to the index with pick mode off.
//
// The only way back used to be "back to picking birds", which keeps the
// picks -- right for changing them, wrong for the common case of a reader
// who has seen the comparison and wants the index again. They were left to
// find Clear in a bar on a page they had not yet returned to.
//
// It must be the SAME exit as the bar's Clear and not a second one built
// beside it: two links that both mean "leave pick mode" and are computed
// separately will one day disagree about what else they keep.
//
// Watched fail: see acceptance.pickModeSurvivesNavigation.
//
// Verifies: product.pickFromIndex. Criterion: acceptance.pickModeSurvivesNavigation.
func TestComparisonOffersAWayOutOfPickMode(t *testing.T) {
	s := newTestServer(t)
	const cmp = "/compare?species=annhum&species=amecro&region=CA&metric=mass"
	body := mustGet(t, s, cmp)

	back := html.UnescapeString(mustMatch(t, body,
		`<a class="backlink" href="([^"]+)"`, "the back link"))
	stop := html.UnescapeString(mustMatch(t, body,
		`<a class="stoplink" href="([^"]+)"`, "the stop-comparing link"))
	label := mustMatch(t, body, `class="stoplink"[^>]*>(?s)(.*?)</a>`, "its text")
	if !strings.Contains(html.UnescapeString(label), "stop comparing") {
		t.Errorf("the second exit reads %q", label)
	}
	// Below the back link, as asked, not above it or elsewhere on the page.
	if strings.Index(body, `class="stoplink"`) < strings.Index(body, `class="backlink"`) {
		t.Error("the stop link comes before the back link")
	}

	u, err := url.Parse(stop)
	if err != nil {
		t.Fatalf("stop link %q: %v", stop, err)
	}
	q := u.Query()
	if _, picking := q["species"]; picking {
		t.Errorf("stop link %q still carries species, so it lands in pick mode", stop)
	}
	// Everything else the comparison carried comes back with the reader.
	for k, want := range map[string]string{"region": "CA", "metric": "mass"} {
		if got := q.Get(k); got != want {
			t.Errorf("stop link %q drops %s=%s; leaving pick mode must not reset the view",
				stop, k, want)
		}
	}
	if stop == back {
		t.Error("both exits go to the same place, so one of them is not an exit")
	}

	// Following it lands on the index outside pick mode: no pick controls.
	landed := mustGet(t, s, stop)
	if strings.Contains(landed, `class="pick`) {
		t.Errorf("following the stop link from %s lands in pick mode", cmp)
	}

	// And it is the bar's own Clear, so the two cannot drift apart.
	bar := html.UnescapeString(mustMatch(t, mustGet(t, s,
		"/?species=annhum&species=amecro&region=CA&metric=mass"),
		`<a class="btn pickclear" href="([^"]+)"`, "the bar's Clear"))
	if bar != stop {
		t.Errorf("the bar's Clear goes to %q and the comparison's stop link to %q; two "+
			"exits from one mode built separately will disagree about what they keep",
			bar, stop)
	}
}

// TestWithheldMonthsAreNotJudged asserts decisions.ebirdSilenceNotAbsence on the cases that found it:
// a month eBird withholds a species is shown and explained, never read as
// a month without the bird -- and a month eBird did publish is still judged.
//
// Painted Bunting was hidden from Florida in January, the heart of its
// Florida season, because eBird publishes none of its Florida records from
// October to March. Short-eared Owl was hidden from New York in June the
// same way. The controls matter as much: a bird genuinely gone in a month
// must stay gone, or the fix has simply switched the month filter off.
//
// Watched fail: see acceptance.withheldMonthsExplained.
//
// Verifies: product.regionAndMonthFilter, product.namesRecordsFiledUnder, decisions.ebirdSilenceNotAbsence. Criterion: acceptance.withheldMonthsExplained.
func TestWithheldMonthsAreNotJudged(t *testing.T) {
	s := newTestServerReal(t)
	tile := func(path, id string) (shown bool, text string) {
		_, body := get(t, s, path)
		i := strings.Index(body, `id="s-`+id+`"`)
		if i < 0 {
			return false, ""
		}
		seg := body[i:]
		if j := strings.Index(seg[1:], `class="tilewrap"`); j > 0 {
			seg = seg[:j]
		}
		return true, strings.Join(strings.Fields(stripTags(seg)), " ")
	}

	for _, c := range []struct{ path, id, want string }{
		{"/?region=FL&month=Jan", "paibun", "Withheld by eBird in Florida, Jan"},
		{"/?region=NY&month=Jun", "sheowl", "Withheld by eBird in New York, Jun"},
	} {
		shown, text := tile(c.path, c.id)
		if !shown {
			t.Errorf("%s is hidden under %s, in a month eBird withholds it: missing by policy "+
				"is not absent", c.id, c.path)
			continue
		}
		if !strings.Contains(text, c.want) {
			t.Errorf("%s under %s does not say %q; its tile reads: %s", c.id, c.path, c.want, text)
		}
	}

	// Without a month filter the caveat names the whole window, and it wraps
	// the year end rather than listing six months.
	if _, text := tile("/?region=FL", "paibun"); !strings.Contains(text, "Withheld by eBird in Florida, Oct-Mar") {
		t.Errorf("Painted Bunting in Florida, all year, reads: %s", text)
	}

	// The list is not the export. Greater Sage-Grouse is listed across the US
	// from February to May, but Wyoming's records run 302 a month inside the
	// window against 269 outside: eBird published them, and saying otherwise
	// would be false.
	if _, text := tile("/?region=WY&month=Mar", "saggro"); strings.Contains(text, "Withheld by eBird") {
		t.Errorf("Greater Sage-Grouse in Wyoming is marked withheld, but its records are in "+
			"the export: %s", text)
	}

	// Controls: birds genuinely absent in a month stay hidden.
	for _, c := range []struct{ path, id string }{
		{"/?region=NY&month=Jan", "boboli"},
		{"/?region=NY&month=Jun", "snobun"},
	} {
		if shown, _ := tile(c.path, c.id); shown {
			t.Errorf("%s is shown under %s; it is not withheld there, just gone, and the month "+
				"filter must still say so", c.id, c.path)
		}
	}

	// And the detail page says it, opened under the same filter (product.tileOpensDetail).
	_, page := get(t, s, "/species/paibun?region=FL&month=Jan")
	if !strings.Contains(stripTags(page), "withheld by eBird in Florida, Oct-Mar") {
		t.Error("Painted Bunting's page does not list Florida's withheld months")
	}
}

// tileOf returns one tile's own markup on a rendered index, from its wrapper
// to the next tile's, so an assertion about it cannot be satisfied by a
// neighbor.
func tileOf(t *testing.T, page, id string) string {
	t.Helper()
	open := `<div class="tilewrap" id="` + tileAnchor(id) + `"`
	i := strings.Index(page, open)
	if i < 0 {
		t.Fatalf("no tile for %s on the index", id)
	}
	seg := page[i:]
	if j := strings.Index(seg[len(open):], `class="tilewrap"`); j > 0 {
		seg = seg[:len(open)+j]
	}
	return seg
}

// TestSameMeasurementsAreNoted asserts product.borrowedMeasurementsSaySo over the real dataset: every
// species drawn from the same AVONET row as another drawn species says so,
// on its tile and on its page, and no other tile does.
//
// The expected groups are computed here from the provenance, independently
// of sharedRows, and Vega with American Herring Gull is named outright so a
// grouping that found nothing could not pass.
//
// Verifies: product.borrowedMeasurementsSaySo, decisions.vegaGullBorrows. Criterion: acceptance.sharedMeasurementsDisclosed.
func TestSameMeasurementsAreNoted(t *testing.T) {
	ds := realDataset(t)
	srv, err := newServer(ds)
	if err != nil {
		t.Fatalf("newServer: %v", err)
	}
	byRow := map[string][]*dataset.Species{}
	for i := range ds.Species {
		sp := &ds.Species[i]
		if p := sp.Segments.Bill.Prov; p.SourceTaxon != "" {
			byRow[p.Source+"|"+p.SourceTaxon] = append(byRow[p.Source+"|"+p.SourceTaxon], sp)
		}
	}
	gulls := byRow["AVONET v7, sheet AVONET1_BirdLife|Larus smithsonianus"]
	if len(gulls) != 2 {
		t.Fatalf("BirdLife's Larus smithsonianus row draws %d species, want Vega and American Herring Gull", len(gulls))
	}

	_, idx := get(t, srv, "/")
	want := 0
	for _, group := range byRow {
		if len(group) < 2 {
			continue
		}
		for _, sp := range group {
			tile := tileOf(t, idx, sp.ID)
			_, page := get(t, srv, "/species/"+sp.ID)
			row := sp.Segments.Bill.Prov
			for _, other := range group {
				if other == sp {
					continue
				}
				want++
				name := html.EscapeString(other.Name(dataset.LocaleNA))
				if !strings.Contains(tile, "Same measurements as "+name) {
					t.Errorf("the tile for %s does not say it shares measurements with %s", sp.ScientificName, name)
				}
				if !strings.Contains(page, "Same measurements as "+name) {
					t.Errorf("/species/%s does not say it shares measurements with %s", sp.ID, name)
				}
			}
			sheet := strings.TrimPrefix(row.Source, "AVONET v7, ")
			if !strings.Contains(page, row.SourceTaxon+", "+sheet) {
				t.Errorf("/species/%s does not name the shared row (%s, %s)", sp.ID, row.SourceTaxon, sheet)
			}
		}
	}
	if got := strings.Count(idx, "Same measurements as "); got != want {
		t.Errorf("the index carries %d shared-measurement notes, want %d: a tile outside every group has one",
			got, want)
	}
}

// TestSameMeasurementsNeedNoTable asserts acceptance.sharedMeasurementsDisclosed's fourth assertion: making a
// fixture species borrow another's row is enough for both tiles to say so.
//
// Verifies: product.borrowedMeasurementsSaySo. Criterion: acceptance.sharedMeasurementsDisclosed.
func TestSameMeasurementsNeedNoTable(t *testing.T) {
	ds := testDataset()
	if len(ds.Species) < 3 {
		t.Fatal("the fixture needs three species: a donor, a borrower and a bystander")
	}
	donor, borrower, bystander := &ds.Species[0], &ds.Species[1], &ds.Species[2]
	borrowed := donor.Segments.Bill.Prov
	borrowed.Kind = dataset.NearMatch
	borrower.Segments.Bill.Prov = borrowed

	srv, err := newServer(ds)
	if err != nil {
		t.Fatalf("newServer: %v", err)
	}
	_, idx := get(t, srv, "/")
	for _, c := range []struct{ sp, other *dataset.Species }{{donor, borrower}, {borrower, donor}} {
		name := html.EscapeString(c.other.Name(dataset.LocaleNA))
		if !strings.Contains(tileOf(t, idx, c.sp.ID), "Same measurements as "+name) {
			t.Errorf("the tile for %s does not name %s", c.sp.ScientificName, name)
		}
	}
	if strings.Contains(tileOf(t, idx, bystander.ID), "Same measurements as") {
		t.Errorf("%s shares no row, but its tile says it does", bystander.ScientificName)
	}
}

// TestRecordsBelowTheBarCount asserts acceptance.belowBarCountsKept's third assertion over the real
// dataset: with every region selected, the figure counts every record the
// species has, listed or not. Emu is named because it is the case that
// was found, and a species with records that reads zero is the bug.
//
// Verifies: product.regionAndMonthFilter, product.shareOfRecords, decisions.belowBarCounted. Criterion: acceptance.belowBarCountsKept.
func TestRecordsBelowTheBarCount(t *testing.T) {
	ds := realDataset(t)
	var emu *dataset.Species
	below := 0
	for i := range ds.Species {
		sp := &ds.Species[i]
		if sp.ScientificName == "Dromaius novaehollandiae" {
			emu = sp
		}
		var want int64
		for _, n := range sp.RegionCounts {
			want += n
		}
		for _, n := range sp.BelowCounts {
			want += n
		}
		below += len(sp.BelowRegions)
		if num, _ := shareIn(sp, ds, options{}); num != want {
			t.Errorf("%s: the figure counts %d records, want %d (listed and below the bar)",
				sp.ScientificName, num, want)
		}
	}
	if below == 0 {
		t.Fatal("no species has a region below the bar; the table was built without them")
	}
	if emu == nil {
		t.Fatal("Emu is not drawn")
	}
	if num, _ := shareIn(emu, ds, options{}); num == 0 {
		t.Error("Emu's figure is zero; its Texas records are below the bar and must still count")
	}
}

// TestExoticsAreOffTheIndexAndListed asserts acceptance.escapedBirdsListedAndServed's second to fourth
// assertions over the real dataset: no flagged bird reaches the index or its
// family control, /excluded lists every one with its reason and a link, and
// each keeps a detail page that says why it is not on the index.
//
// Verifies: product.escapedBirdsListed, decisions.escapedBirdsOffIndex. Criterion: acceptance.escapedBirdsListedAndServed.
func TestExoticsAreOffTheIndexAndListed(t *testing.T) {
	ds := realDataset(t)
	srv, err := newServer(ds)
	if err != nil {
		t.Fatalf("newServer: %v", err)
	}
	var exotic []*dataset.Species
	famAll, famExotic := map[string]int{}, map[string]int{}
	for i := range ds.Species {
		sp := &ds.Species[i]
		famAll[sp.Family]++
		if sp.Exotic != "" {
			exotic = append(exotic, sp)
			famExotic[sp.Family]++
		}
	}
	if len(exotic) < 30 {
		t.Fatalf("only %d flagged species in the dataset; the check has too little to assert", len(exotic))
	}

	_, idx := get(t, srv, "/")
	if !strings.Contains(idx, `href="/excluded"`) {
		t.Error("the index does not link to /excluded")
	}
	tiles := 0
	for _, sp := range exotic {
		if strings.Contains(idx, `id="`+tileAnchor(sp.ID)+`"`) {
			tiles++
			t.Errorf("%s is flagged exotic but has a tile on the index", sp.ScientificName)
		}
	}
	// The count the page reports leaves them out too.
	shown := len(ds.Species) - len(exotic)
	if want := fmt.Sprintf("%d of %d species", shown, shown); !strings.Contains(stripTags(idx), want) {
		t.Errorf("the index does not report %q; flagged birds are still being counted", want)
	}
	gone := 0
	for fam, n := range famExotic {
		if n == famAll[fam] {
			gone++
			if strings.Contains(idx, "family="+fam) {
				t.Errorf("family %s holds only flagged birds but is still offered", fam)
			}
		}
	}
	if gone == 0 {
		t.Error("no family is made only of flagged birds; the family check asserted nothing")
	}

	code, list := get(t, srv, "/excluded")
	if code != http.StatusOK {
		t.Fatalf("/excluded answers %d", code)
	}
	for _, sp := range exotic {
		if !strings.Contains(list, `href="/species/`+sp.ID+`"`) {
			t.Errorf("/excluded does not link %s", sp.ScientificName)
		}
		if !strings.Contains(list, html.EscapeString(sp.Exotic)) {
			t.Errorf("/excluded does not give %s's reason", sp.ScientificName)
		}
		code, page := get(t, srv, "/species/"+sp.ID)
		if code != http.StatusOK {
			t.Errorf("/species/%s answers %d; a bird off the index keeps its page", sp.ID, code)
			continue
		}
		if !strings.Contains(page, "Not on the index") || !strings.Contains(page, html.EscapeString(sp.Exotic)) ||
			!strings.Contains(page, `href="/excluded"`) {
			t.Errorf("/species/%s does not say it is off the index, why, and where the list is", sp.ID)
		}
	}
	// An ordinary bird says none of it.
	if _, plain := get(t, srv, "/species/amerob"); strings.Contains(plain, "Not on the index") {
		t.Error("American Robin's page says it is not on the index")
	}
}

// TestReadersAreToldItIsAStickFigure pins the reader-facing term. On the
// pages a reader sees, the drawing is a stick figure, and the only
// "diagram" left is the link to the data-pipeline diagram, which product.pipelineDiagram owns.
// Visible text only: class names, template comments and the code keep
// "diagram", by design.
//
// Verifies: product.segmentComposition. Criterion: acceptance.readersToldStickFigure.
func TestReadersAreToldItIsAStickFigure(t *testing.T) {
	s := newTestServer(t)
	strip := []*regexp.Regexp{
		regexp.MustCompile(`(?s)<!--.*?-->`),
		regexp.MustCompile(`(?s)<script.*?</script>`),
		regexp.MustCompile(`(?s)<style.*?</style>`),
		regexp.MustCompile(`(?s)<a href="/docs/pipeline">.*?</a>`),
		regexp.MustCompile(`<[^>]+>`),
	}
	visible := func(page string) string {
		for _, re := range strip {
			page = re.ReplaceAllString(page, " ")
		}
		return html.UnescapeString(page)
	}
	word := regexp.MustCompile(`(?i)\bdiagrams?\b`)
	for _, path := range []string{"/", "/species/amerob", "/compare?species=amerob&species=blujay"} {
		text := visible(mustGet(t, s, path))
		for _, m := range word.FindAllStringIndex(text, -1) {
			lo, hi := max(0, m[0]-40), min(len(text), m[1]+40)
			t.Errorf("%s says %q to readers; the drawing is a stick figure (decisions.stickFigure)",
				path, strings.Join(strings.Fields(text[lo:hi]), " "))
		}
	}
	if !strings.Contains(visible(mustGet(t, s, "/")), "stick figure") {
		t.Error("the index never calls the drawing a stick figure")
	}
	if !strings.Contains(visible(mustGet(t, s, "/docs/using")), "stick figure") {
		t.Error("the using guide never calls the drawing a stick figure")
	}
}
