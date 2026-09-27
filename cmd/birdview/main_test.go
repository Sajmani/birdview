package main

import (
	"encoding/xml"
	"fmt"
	"html"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"birdview/internal/dataset"
	"birdview/internal/view"
	"birdview/web"
)

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
			ID: id, CommonName: common, BandingCode: code(common), ScientificName: sci, Family: fam,
			FamilyCommon: fam + " family", FamilyColor: "#86efac",
			TotalLength: mod(bill+head+neck+body+tail, "sum of the five diagram segments"),
			Wingspan:    dataset.Metric{Value: span, Prov: dataset.Provenance{Kind: k, Source: "Fox et al. 2026", SourceTaxon: "Other species"}},
			Mass:        m(mass),
			WingChord:   m(chord),
			Segments: dataset.Segments{
				Bill: m(bill), Head: mod(head, "cranial vault model"), Neck: mod(neck, "neck ratio for "+fam),
				Body: mod(body, "trunk model"), Tail: m(tail),
			},
		}
		return sp
	}
	// Passeriformes deliberately holds TWO families. With one family per order,
	// "select the whole group" and "select the first family" are the same
	// operation, and a bug in the group logic passes unnoticed.
	orders := map[string]string{
		"Trochilidae": "Apodiformes", "Turdidae": "Passeriformes",
		"Corvidae": "Passeriformes", "Ardeidae": "Pelecaniformes",
		"Recurvirostridae": "Charadriiformes",
	}
	// The stilt adds a fourth order, so the order-group controls are exercised
	// against something other than the two-family Passeriformes case.
	//
	// The final column is the wing chord, and the crow's wingspan is MODELED on
	// purpose: without one modeled span in the fixture, every crossbar renders
	// solid and the split of P-065 is never exercised. Every chord is shorter
	// than its own half-span, as it is for all 841 real species.
	sp := []dataset.Species{
		mk("calypte_anna", "Anna's Hummingbird", "Calypte anna", "Trochilidae", 18, 12, 6, 40, 26, 120, 4, 49, dataset.Measured),
		mk("turdus_migratorius", "American Robin", "Turdus migratorius", "Turdidae", 20, 18, 12, 90, 85, 390, 77, 130, dataset.Measured),
		mk("corvus_brachyrhynchos", "American Crow", "Corvus brachyrhynchos", "Corvidae", 50, 30, 22, 170, 180, 1000, 450, 330, dataset.Modeled),
		mk("himantopus_mexicanus", "Black-necked Stilt", "Himantopus mexicanus", "Recurvirostridae", 66.6, 42.1, 40.3, 89.5, 73.6, 750, 176.8, 230, dataset.NearMatch),
		mk("ardea_herodias", "Great Blue Heron", "Ardea herodias", "Ardeidae", 140, 40, 200, 300, 180, 1825, 2400, 466.4, dataset.NearMatch),
	}
	// iNaturalist taxon IDs (P-064). The stilt is deliberately left unmatched so
	// both branches of the detail view are exercised: a species iNaturalist has
	// not confidently matched must show no link at all.
	inat := map[string]int{
		"calypte_anna": 5305, "turdus_migratorius": 12727,
		"corvus_brachyrhynchos": 8021, "ardea_herodias": 4956,
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
	return &dataset.Dataset{Families: fams, Species: sp}
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

// code builds a plausible four-letter alpha code from a common name, so the
// fixture exercises the same rendering path as real data.
func code(common string) string {
	var out []byte
	for _, w := range strings.Fields(common) {
		out = append(out, strings.ToUpper(w[:1])...)
	}
	for len(out) < 4 {
		out = append(out, 'X')
	}
	return string(out[:4])
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
	w := httptest.NewRecorder()
	s.ServeHTTP(w, httptest.NewRequest(http.MethodGet, path, nil))
	return w.Code, w.Body.String()
}

// Verifies: T-001, T-003. Criterion: AC-054.
func TestRoutesServe(t *testing.T) {
	s := newTestServer(t)
	for _, path := range []string{
		"/", "/static/styles.css", "/species/ardea_herodias",
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
// direct conflict with AC-064. The first version of this check made exactly
// that mistake and failed the moment proper attribution was added.
var externalSubresource = regexp.MustCompile(
	`(?:src\s*=\s*"(?:https?:)?//[^"]+|<link[^>]+href\s*=\s*"(?:https?:)?//[^"]+)`)

// TestNoExternalResources asserts T-002 where it actually bites: a page that
// pulls a stylesheet or script from a CDN is not self-contained, however
// self-contained the binary is. The previous design loaded Tailwind from
// cdn.tailwindcss.com and would have failed this.
//
// What it must NOT forbid is an ordinary hyperlink. T-002 is about what the
// page loads, not where it points, and CC BY obliges BirdView to link out to
// every source and license.
//
// Verifies: T-002. Criterion: AC-051.
func TestNoExternalResources(t *testing.T) {
	s := newTestServer(t)
	for _, path := range []string{"/", "/species/ardea_herodias"} {
		_, body := get(t, s, path)
		if m := externalSubresource.FindAllString(body, -1); len(m) > 0 {
			t.Errorf("%s fetches external resources: %v", path, m)
		}
	}
}

// Verifies: P-031. Criterion: AC-055.
func TestTileShowsHeaderAndMetrics(t *testing.T) {
	_, body := get(t, newTestServer(t), "/")
	for _, want := range []string{
		"American Robin",     // common name
		"(ARXX)",             // banding code, in parentheses after it
		"Turdus migratorius", // scientific name
		"Turdidae",           // family badge, scientific name
		"Turdidae family",    // family badge, common name alongside it
		"<svg",               // the diagram
		"L ", "W ", "M ",     // the three main metrics
	} {
		if !strings.Contains(body, want) {
			t.Errorf("tile is missing %q", want)
		}
	}

	// The common and scientific names share one header line (P-031), so no
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
// Verifies: P-031, P-042. Criteria: AC-055, AC-031.
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

// TestDiagramIsVerticallyCentered asserts P-031: a bird shorter than the
// drawing area sits in the middle of it, not on the floor.
//
// This matters most under global scale, where a small bird occupies a fraction
// of the tile's height and bottom-alignment reads as a rendering fault rather
// than as the size difference it is.
//
// Verifies: P-031. Criterion: AC-055.
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

// TestDetailShowsEverySegmentWithProvenance asserts P-034 and P-063: the detail
// view lists all five segments as well as the three main metrics, each with its
// provenance explained.
//
// Verifies: P-034, P-063. Criterion: AC-043.
func TestDetailShowsEverySegmentWithProvenance(t *testing.T) {
	_, body := get(t, newTestServer(t), "/species/ardea_herodias")
	for _, label := range []string{"Total length", "Wingspan", "Mass",
		"Bill", "Head", "Neck", "Body", "Tail"} {
		if !strings.Contains(body, ">"+label+"<") {
			t.Errorf("detail view omits %q", label)
		}
	}
	for _, want := range []string{
		"neck ratio for Ardeidae",   // the family's model is named (P-045)
		"measured on Other species", // near-match names the borrowed taxon
		"drawing convention",        // the neck disclosure (P-063)
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

// Verifies: P-010. Criterion: AC-056.
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

	// The selector labels each family with both names and a count (P-010): the
	// scientific name alone makes the user translate before they can choose.
	sel := all[strings.Index(all, `class="ordergroups"`):]
	sel = sel[:strings.Index(sel, `<p class="count"`)]
	for _, want := range []string{"Ardeidae", "Ardeidae family", "Turdidae", "Trochilidae"} {
		if !strings.Contains(sel, want) {
			t.Errorf("family selector is missing %q", want)
		}
	}
	if !strings.Contains(sel, "<b>1</b>") {
		t.Error("family selector shows no species counts")
	}

	// Grouped by order, in checklist sequence rather than alphabetically
	// (P-010). Alphabetical would put Ardeidae first; the fixture's checklist
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
// The help now contains a real labeled diagram of a real species (P-035), so a
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

// TestUnitToggle asserts P-032 where the user actually reads the numbers.
//
// The earlier version of this check only looked for a centimeter value
// somewhere on the page, and passed while every tile was hardcoded to inches:
// the section headers alone satisfied it. Asserting on the tile footer is the
// difference between checking the requirement and checking that some part of
// the page happens to mention the right unit.
//
// Verifies: P-032. Criterion: AC-057.
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
// Verifies: P-032. Criterion: AC-057.
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
		if got := format(c.g, view.ByMass, false); got != c.imperial {
			t.Errorf("%.0f g imperial: got %q, want %q", c.g, got, c.imperial)
		}
		if got := format(c.g, view.ByMass, true); got != c.metric {
			t.Errorf("%.0f g metric: got %q, want %q", c.g, got, c.metric)
		}
	}
}

// TestDefaultScaleIsPerSection asserts P-042's default, which CR-017 moved.
//
// Global still makes the cross-family argument better than anything else, and
// this test asserted it for good reasons. What changed is that the diagram
// acquired a vocabulary -- dashed for modeled, chord ticks, labels -- and at
// global scale a median bird is a couple of pixels tall, so the landing page
// was hiding the drawing's own language from every first-time reader.
//
// Verifies: P-042. Criterion: AC-031.
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

// TestDetailDrawsFullScale asserts P-034: the detail view fills its box with
// the one bird it is about, whatever scale the index is using.
//
// Under the global default a kinglet is a few pixels tall on the index, which
// is the honest comparison. Inheriting that here would leave nothing to look at
// on the page devoted to it.
//
// Verifies: P-034. Criterion: AC-058.
func TestDetailDrawsFullScale(t *testing.T) {
	s := newTestServer(t)
	// The hummingbird is the smallest fixture bird, so it is the one that would
	// be reduced to nothing if the index's scale leaked through.
	for _, path := range []string{
		"/species/calypte_anna",
		"/species/calypte_anna?scale=global",
		"/species/calypte_anna?scale=section",
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

// TestHelpCoversEveryControl asserts P-035: the page explains its own controls
// and its diagram, rather than assuming the reader can infer them.
//
// The list is deliberately tied to the controls that exist. If a control is
// added and this check is not updated, the new control is undocumented and
// this test is the reminder.
//
// Verifies: P-035. Criterion: AC-059.
func TestHelpCoversEveryControl(t *testing.T) {
	_, body := get(t, newTestServer(t), "/")
	i := strings.Index(body, `<details class="help"`)
	if i < 0 {
		t.Fatal("no help block on the page")
	}
	help := body[i:]
	help = help[:strings.Index(help, "</details>")]

	// Every control offered by the controls bar.
	for _, control := range []string{"Order by", "Scale", "Section step", "Units", "Family"} {
		if !strings.Contains(help, control) {
			t.Errorf("help does not mention the %q control", control)
		}
	}
	// Every part of the diagram, since none of it is self-evident. These are now
	// labels on a rendered example rather than swatches in a list, so each is
	// matched as the start of its label text.
	for _, part := range []string{"Bill", "Head", "Neck", "Body", "Tail", "Wingspan", "Wing chord"} {
		if !strings.Contains(help, ">"+part+" ") {
			t.Errorf("diagram legend omits %q", part)
		}
	}
	// The three sections run in a stated order (P-035): the thing on screen,
	// then what its strokes mean, then how to change it.
	order := []string{"The diagram", "Where the numbers come from", "The controls"}
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

	// The two things a reader cannot discover by clicking: what the estimate
	// marker means, and that both scale modes exist.
	for _, concept := range []string{"estimated", "Global", "Per section"} {
		if !strings.Contains(help, concept) {
			t.Errorf("help does not explain %q", concept)
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
// Verifies: P-005, T-010. Criterion: AC-060.
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

// TestOrderToggles asserts P-010's order control: each order is a real toggle
// over its whole group, labeled with its common name.
//
// Since a fresh page selects everything, the first click on an order REMOVES
// it. This replaces a check that asserted the opposite -- that order labels
// were inert text. When that requirement inverted, the check kept passing,
// because it searched for `<a class="ohdr` and the markup had become
// `<a class="btn ohdr`. A test that survives the reversal of its own
// requirement is worse than none.
//
// Verifies: P-010. Criterion: AC-056.
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
// Verifies: P-010. Criterion: AC-056.
func TestAllToggle(t *testing.T) {
	s := newTestServer(t)

	_, fresh := get(t, s, "/")
	clear := html.UnescapeString(mustMatch(t, fresh,
		`<a class="btn on"[^>]*href="([^"]+)"[^>]*>All`, "the All control on a fresh page"))

	_, empty := get(t, s, clear)
	if strings.Count(empty, `class="btn ohdr on"`) != 0 {
		t.Error("orders still show as selected after clearing")
	}
	if strings.Count(empty, `class="chip on"`) != 0 {
		t.Error("families still show as selected after clearing")
	}

	// Back again: All selects every family, and every order lights up.
	restore := html.UnescapeString(mustMatch(t, empty,
		`<a class="btn"[^>]*href="([^"]+)"[^>]*>All`, "the All control on an emptied page"))
	_, all := get(t, s, restore)
	_, _, wantOrders := fixtureCounts()
	if n := strings.Count(all, `class="btn ohdr on"`); n != wantOrders {
		t.Errorf("%d of %d orders selected after re-selecting All", n, wantOrders)
	}
	if !strings.Contains(all, "Great Blue Heron") {
		t.Error("re-selecting All did not restore the species")
	}
}

// TestDefaultSelectsEverything asserts that a fresh page has everything on:
// All, every order and every family (P-010, P-033).
//
// Verifies: P-010, P-033. Criterion: AC-056.
func TestDefaultSelectsEverything(t *testing.T) {
	s := newTestServer(t)
	_, body := get(t, s, "/")
	if !regexp.MustCompile(`<a class="btn on"[^>]*>All`).MatchString(body) {
		t.Error("All is not selected on a fresh page")
	}
	_, wantFams, wantOrders := fixtureCounts()
	if n := strings.Count(body, `class="btn ohdr on"`); n != wantOrders {
		t.Errorf("%d of %d orders selected on a fresh page", n, wantOrders)
	}
	if n := strings.Count(body, `class="chip on"`); n != wantFams {
		t.Errorf("%d of %d families selected on a fresh page", n, wantFams)
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
// Verifies: P-010. Criterion: AC-061.
func TestClearingEverythingShowsGuidance(t *testing.T) {
	s := newTestServer(t)
	_, body := get(t, s, "/")
	clear := html.UnescapeString(mustMatch(t, body,
		`<a class="btn on"[^>]*href="([^"]+)"[^>]*>All`, "the All control on a fresh page"))

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
	// And All must offer the way back.
	if regexp.MustCompile(`<a class="btn on"[^>]*>All`).MatchString(empty) {
		t.Error("All is still marked selected with nothing selected")
	}
}

// TestScrollRestorationIsInlineAndSelfContained checks the one script BirdView
// ships is embedded rather than fetched, and that it is scoped to the index.
//
// It cannot check that scrolling actually works -- that needs a browser this
// project does not have -- so AC-062 is honest about covering presence and
// self-containment only.
//
// Verifies: T-002, P-038. Criterion: AC-062.
func TestScrollRestorationIsInlineAndSelfContained(t *testing.T) {
	s := newTestServer(t)
	_, body := get(t, s, "/")

	if !strings.Contains(body, "birdview.scroll") {
		t.Error("index carries no scroll-restoration script")
	}
	// Inline only: an external script would break T-002 the same way a CDN
	// stylesheet would, and AC-051 only looks at src/href attributes.
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
	_, detail := get(t, s, "/species/ardea_herodias")
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
// Verifies: P-035, P-036, P-038. Criterion: AC-062.
func TestCollapsiblePanelsPersist(t *testing.T) {
	_, body := get(t, newTestServer(t), "/")

	for _, id := range []string{"instructions", "controls"} {
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
	// Both panels still default to open for a first-time reader.
	for _, want := range []string{`class="help" id="instructions" open`, `class="controls" id="controls" open`} {
		if !strings.Contains(body, want) {
			t.Errorf("panel does not default to open: expected %q", want)
		}
	}
}

// TestControlOrderAndOptions pins two presentation choices that are easy to
// undo by accident: the default scale is listed first, and the section step
// offers its full range.
//
// Verifies: P-023, P-042. Criterion: AC-063.
func TestControlOrderAndOptions(t *testing.T) {
	_, page := get(t, newTestServer(t), "/")

	// Scope to the controls panel. The instructions above it also name both
	// scale modes, in the same order -- searching the whole page matched the
	// prose and passed even with the buttons reversed.
	ci := strings.Index(page, `<details class="controls"`)
	if ci < 0 {
		t.Fatal("controls panel not found")
	}
	body := page[ci:]

	// Per section is the default (P-042, CR-017), so it leads the pair.
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
// Verifies: T-011, G-004. Criterion: AC-064.
func TestAttributionIsComplete(t *testing.T) {
	_, body := get(t, newTestServer(t), "/")
	f := body[strings.Index(body, "<footer>"):]

	for _, want := range []string{
		"Tobias",                              // AVONET creator
		"10.6084/m9.figshare.16586228",        // AVONET source
		"Fox",                                 // wingspan creator
		"10.6084/m9.figshare.32089083",        // wingspan source
		"creativecommons.org/licenses/by/4.0", // the license itself
		"American Birding Association",        // the non-CC source still needs credit
		"ABA Checklist",
	} {
		if !strings.Contains(f, want) {
			t.Errorf("attribution is missing %q", want)
		}
	}
	// And the honesty note that the numbers are mostly modeled.
	if !strings.Contains(f, "modeled") {
		t.Error("footer does not disclose that lengths are modeled")
	}
	// The detail view is a shareable URL of its own, so it carries credit too.
	_, detail := get(t, s2(t), "/species/ardea_herodias")
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
// spec/sources.md.
//
// Verifies: G-004. Criterion: AC-064.
func TestNoUnearnedPermissionClaim(t *testing.T) {
	for _, page := range []string{"/", "/species/ardea_herodias"} {
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

// TestFoldedNeckRendersAsAnArc asserts P-040 where the user sees it: a heron's
// neck reaches the page as an SVG arc path, every other bird's as a straight
// line, and the detail view still reports the neck at its full modeled length.
//
// That last part is the requirement's real content. The fold is a change to the
// silhouette only -- if it ever reached the numbers, BirdView would be quietly
// reporting a shorter heron than it measured.
//
// Verifies: P-040, P-041, P-034. Criterion: AC-067.
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
	_, det := get(t, s, "/species/ardea_herodias?units=metric")
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

// TestINatLinkOnlyForAConfidentMatch asserts P-064 on the rendered page: a
// species with a mapping gets a link to exactly its taxon page, and one without
// gets nothing -- no placeholder, no search URL, no "not found" row.
//
// The absence is the half worth testing. A search link would never 404, which is
// exactly what makes it the wrong answer: it always looks like it worked. An
// absent link is legible.
//
// Verifies: P-064. Criterion: AC-068.
func TestINatLinkOnlyForAConfidentMatch(t *testing.T) {
	s := newTestServer(t)

	_, matched := get(t, s, "/species/ardea_herodias")
	href := mustMatch(t, matched, `<a href="(https://www\.inaturalist\.org/taxa/[0-9]+)"`, "the iNaturalist link")
	if href != "https://www.inaturalist.org/taxa/4956" {
		t.Errorf("link points at %q, want the heron's taxon page", href)
	}
	if !strings.Contains(matched, "Great Blue Heron on\niNaturalist") &&
		!strings.Contains(matched, "Great Blue Heron on iNaturalist") {
		t.Error("the link does not name the species and the destination")
	}

	// The unmatched species must show no iNaturalist reference of any kind.
	_, unmatched := get(t, s, "/species/himantopus_mexicanus")
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
// rather than copying (P-064, inat-terms/R1). AC-051's subresource scan already
// forbids the mechanism for T-002 reasons; this says why it also matters here.
//
// Photographs on iNaturalist are contributors' work under their own licenses,
// frequently CC BY-NC. Linking engages none of them. Embedding would engage all
// of them, per photograph.
//
// Verifies: P-064. Criterion: AC-068.
func TestNoINaturalistImagesAreEmbedded(t *testing.T) {
	_, body := get(t, newTestServer(t), "/species/ardea_herodias")
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
// one (T-002), and that both pages actually reference it.
//
// No new requirement was needed: T-002 already forbids fetching a subresource
// from another host, and AC-051's scan already covers a `<link href>` pointing
// off-host. What that scan cannot see is a favicon that is referenced but not
// served, or served but referenced by nothing -- so this checks the two ends
// match.
//
// Verifies: T-002. Criterion: AC-071.
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
	// It is drawn with the app's own colors, so a theme change that leaves the
	// icon behind is visible here rather than only in a browser tab.
	for _, want := range []string{"#0f172a", "#5ccbfa", "#cbd5e1"} {
		if !strings.Contains(body, want) {
			t.Errorf("favicon does not use %s; it should be drawn in the diagram's palette", want)
		}
	}

	for _, path := range []string{"/", "/species/ardea_herodias"} {
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

// TestCrossbarSplitsAtTheChord asserts P-065 on the rendered page: a measured
// wingspan draws no dashed segment, a modeled one draws exactly one, flanked by
// two solid segments and marked by two ticks.
//
// Verifies: P-065. Criterion: AC-072.
func TestCrossbarSplitsAtTheChord(t *testing.T) {
	s := newTestServer(t)
	wing := regexp.MustCompile(`<line class="(wing(?: est)?)"`)

	cases := []struct {
		id, name    string
		spanModeled bool
	}{
		{"ardea_herodias", "Great Blue Heron", false},
		{"corvus_brachyrhynchos", "American Crow", true},
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

// TestBothPagesExplainTheStrokeCode asserts the other half of P-065: a visual
// code nobody documents is decoration.
//
// Verifies: P-065. Criterion: AC-072.
func TestBothPagesExplainTheStrokeCode(t *testing.T) {
	s := newTestServer(t)
	for _, path := range []string{"/", "/species/ardea_herodias"} {
		_, body := get(t, s, path)
		low := strings.ToLower(body)
		for _, want := range []string{"solid", "dashed", "tick", "chord"} {
			if !strings.Contains(low, want) {
				t.Errorf("%s does not explain %q; the stroke code is undocumented there", path, want)
			}
		}
	}
}

// TestDetailDiagramIsLabeled asserts P-066 on the rendered page: five labels,
// each naming its segment and carrying the value the provenance table gives for
// it.
//
// The values are read out of the table rather than recomputed here, so the two
// cannot drift apart. A label that disagreed with the table would be the worst
// outcome of this feature -- two numbers for one segment on one page.
//
// Verifies: P-066. Criterion: AC-073.
func TestDetailDiagramIsLabeled(t *testing.T) {
	_, body := get(t, newTestServer(t), "/species/ardea_herodias?units=metric")

	// Matches a segment label, dashed or not, but deliberately NOT the chord:
	// that one is checked separately below, and folding it in here would let a
	// missing segment hide behind it.
	labels := regexp.MustCompile(`<text class="dlabel(?: est)?"[^>]*>([^<]+)</text>`).FindAllStringSubmatch(body, -1)
	if len(labels) != 5 {
		t.Fatalf("detail diagram carries %d segment labels, want one per segment", len(labels))
	}
	got := map[string]string{}
	for _, m := range labels {
		f := strings.Fields(m[1])
		got[strings.ToLower(f[0])] = strings.Join(f[1:], " ")
	}
	for _, seg := range []string{"bill", "head", "neck", "body", "tail"} {
		val, ok := got[seg]
		if !ok {
			t.Errorf("no label for the %s", seg)
			continue
		}
		row := mustMatch(t, body,
			`<th>`+strings.ToUpper(seg[:1])+seg[1:]+`</th>\s*<td class="v">([^<]+)</td>`,
			"the "+seg+" row")
		if val != row {
			t.Errorf("%s: diagram says %q, provenance table says %q", seg, val, row)
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
// Verifies: P-066. Criterion: AC-073.
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
			worst, worstName = gap, sp.CommonName
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
