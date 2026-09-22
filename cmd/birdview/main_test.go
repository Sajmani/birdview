package main

import (
	"fmt"
	"html"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"birdview/internal/dataset"
	"birdview/internal/morphology"
	"birdview/internal/view"
	"birdview/web"
)

func testDataset() *dataset.Dataset {
	mk := func(id, common, sci, fam string, bill, head, neck, body, tail, span, mass, tarsus float64, k dataset.Kind) dataset.Species {
		m := func(v float64) dataset.Metric {
			return dataset.Metric{Value: v, Prov: dataset.Provenance{
				Kind: dataset.Measured, Source: "AVONET v7, sheet AVONET2_eBird", SourceTaxon: sci}}
		}
		mod := func(v float64, s string) dataset.Metric {
			return dataset.Metric{Value: v, Prov: dataset.Provenance{Kind: dataset.Modeled, Model: s}}
		}
		// The leg comes from the real model rather than a literal, so the
		// fixture cannot quietly disagree with the family factors it claims.
		legV, legModel := morphology.LegLength(tarsus, fam)
		sp := dataset.Species{
			ID: id, CommonName: common, BandingCode: code(common), ScientificName: sci, Family: fam,
			FamilyCommon: fam + " family", FamilyColor: "#86efac",
			TotalLength: mod(bill+head+neck+body+tail, "sum of the five diagram segments"),
			Wingspan:    dataset.Metric{Value: span, Prov: dataset.Provenance{Kind: k, Source: "Fox et al. 2026", SourceTaxon: "Other species"}},
			Mass:        m(mass),
			Tarsus:      m(tarsus),
			Leg:         mod(morphology.Round1(legV), legModel),
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
	// Tarsus values are the real AVONET figures, so the modeled legs the fixture
	// carries are the ones the shipped dataset carries. The stilt is here for
	// one reason: its feet fall below its tail tip, the case the renderer has to
	// leave room for and no other fixture bird exercises.
	sp := []dataset.Species{
		mk("calypte_anna", "Anna's Hummingbird", "Calypte anna", "Trochilidae", 18, 12, 6, 40, 26, 120, 4, 4.6, dataset.Measured),
		mk("turdus_migratorius", "American Robin", "Turdus migratorius", "Turdidae", 20, 18, 12, 90, 85, 390, 77, 32.7, dataset.Measured),
		mk("corvus_brachyrhynchos", "American Crow", "Corvus brachyrhynchos", "Corvidae", 50, 30, 22, 170, 180, 1000, 450, 55.0, dataset.Measured),
		mk("himantopus_mexicanus", "Black-necked Stilt", "Himantopus mexicanus", "Recurvirostridae", 66.6, 42.1, 40.3, 89.5, 73.6, 750, 176.8, 106.7, dataset.NearMatch),
		mk("ardea_herodias", "Great Blue Heron", "Ardea herodias", "Ardeidae", 140, 40, 200, 300, 180, 1825, 2400, 167.6, dataset.NearMatch),
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

// TestDiagramFitsItsBoxFeetIncluded asserts that the renderer leaves room for
// legs that reach below the tail tip (P-040).
//
// The diagram used to be centered on its bill-to-tail extent, which was safe
// only while the legs were the measured tarsus. The modeled leg reaches further:
// 11 of the 841 shipped species stand lower on their feet than their tail
// reaches, a Black-necked Stilt by 12% of its own length. Centering on
// bill-to-tail now pushes those feet through the floor of the box.
//
// So this reads the emitted SVG and checks that nothing lands outside the
// viewBox. It runs under per-section scale deliberately: that is where the
// long-legged bird sets its own reference and fills the box exactly, so an
// off-by-the-feet error has nowhere to hide. Under global scale the heron sets
// the scale, the stilt is drawn small, and the same bug leaves it merely
// off-center -- visible to a human, invisible to an assertion.
//
// Verifies: P-040, P-031. Criterion: AC-066.
func TestDiagramFitsItsBoxFeetIncluded(t *testing.T) {
	_, body := get(t, newTestServer(t), "/?scale=section")

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
			t.Errorf("%s: drawing reaches %.1f in a %.1f box -- its feet are clipped",
				name, dy+lowest, boxH)
		}
		if dy < -0.2 {
			t.Errorf("%s: drawing starts at %.1f, above the top of the box", name, dy)
		}
	}
}

// TestLegsDrawBehindTheBody asserts the draw order of P-040: the body stick is
// painted after the legs, so the legs emerge from behind it.
//
// SVG has no z-index -- paint order is document order -- so this is a property
// of where the element sits in the template and nothing else. Drawn the other
// way, two leg strokes cross the body stick and the junction mark, and a small
// bird reads as a scribble rather than as a bird.
//
// Verifies: P-040. Criterion: AC-066.
func TestLegsDrawBehindTheBody(t *testing.T) {
	_, body := get(t, newTestServer(t), "/")
	for _, d := range diagrams(t, body) {
		name := mustMatch(t, d, `aria-label="scaled outline of ([^"]+)"`, "the diagram label")
		leg, trunk := strings.Index(d, `class="leg"`), strings.Index(d, `class="body"`)
		if leg < 0 || trunk < 0 {
			t.Errorf("%s: diagram is missing a leg or body line", name)
			continue
		}
		if leg > trunk {
			t.Errorf("%s: legs are drawn after the body, so they cross it instead of "+
				"passing behind it", name)
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
	for _, label := range []string{"Total length", "Wingspan", "Mass", "Leg", "Tarsus",
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

// TestDefaultScaleIsGlobal asserts P-042's default. The cross-family surprise
// is the product; opening on a per-section scale shows every bird at a
// comfortable size, which is the impression BirdView exists to correct.
//
// Verifies: P-042. Criterion: AC-031.
func TestDefaultScaleIsGlobal(t *testing.T) {
	if got := parseOptions(httptest.NewRequest(http.MethodGet, "/", nil)).Scale; got != view.Global {
		t.Errorf("default scale is %q, want %q", got, view.Global)
	}
	// And the toggle must still reach the other mode.
	r := httptest.NewRequest(http.MethodGet, "/?scale=section", nil)
	if got := parseOptions(r).Scale; got != view.PerSection {
		t.Errorf("?scale=section gave %q, want %q", got, view.PerSection)
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
	// Every part of the diagram, since none of it is self-evident.
	for _, part := range []string{"Bill", "Head", "Neck", "Body", "Tail", "Wingspan"} {
		if !strings.Contains(help, ">"+part+"<") {
			t.Errorf("diagram legend omits %q", part)
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
		if strings.Contains(dropped, gone) {
			t.Errorf("dropping Passeriformes left %s visible", gone)
		}
	}
	if !strings.Contains(dropped, "Great Blue Heron") {
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
		if strings.Contains(empty, gone) {
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

	// Global is the default (P-042), so it leads the pair.
	g := strings.Index(body, `"scale" "global"`)
	if g < 0 {
		g = strings.Index(body, `scale=global`)
	}
	p := strings.Index(body, `scale=section`)
	if g < 0 || p < 0 {
		t.Fatalf("scale controls not found in the panel (global=%d section=%d)", g, p)
	}
	if g > p {
		t.Error("Per section is listed before Global, but Global is the default")
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
