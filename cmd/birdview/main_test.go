package main

import (
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
	mk := func(id, common, sci, fam string, bill, head, neck, body, tail, span, mass float64, k dataset.Kind) dataset.Species {
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
	}
	sp := []dataset.Species{
		mk("calypte_anna", "Anna's Hummingbird", "Calypte anna", "Trochilidae", 18, 12, 6, 40, 26, 120, 4, dataset.Measured),
		mk("turdus_migratorius", "American Robin", "Turdus migratorius", "Turdidae", 20, 18, 12, 90, 85, 390, 77, dataset.Measured),
		mk("corvus_brachyrhynchos", "American Crow", "Corvus brachyrhynchos", "Corvidae", 50, 30, 22, 170, 180, 1000, 450, dataset.Measured),
		mk("ardea_herodias", "Great Blue Heron", "Ardea herodias", "Ardeidae", 140, 40, 200, 300, 180, 1825, 2400, dataset.NearMatch),
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

// externalRef finds a resource the browser would fetch from another host.
var externalRef = regexp.MustCompile(`(?:src|href)\s*=\s*"(?:https?:)?//[^"]+`)

// TestNoExternalResources asserts T-002 where it actually bites: a page that
// pulls a stylesheet or script from a CDN is not self-contained, however
// self-contained the binary is. The previous design loaded Tailwind from
// cdn.tailwindcss.com and would have failed this.
//
// Verifies: T-002. Criterion: AC-051.
func TestNoExternalResources(t *testing.T) {
	s := newTestServer(t)
	for _, path := range []string{"/", "/species/ardea_herodias"} {
		_, body := get(t, s, path)
		if m := externalRef.FindAllString(body, -1); len(m) > 0 {
			t.Errorf("%s references external resources: %v", path, m)
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
	for _, label := range []string{"Total length", "Wingspan", "Mass", "Bill", "Head", "Neck", "Body", "Tail"} {
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
// that selects or clears its whole group, labeled with its common name.
//
// This replaces a check that asserted the opposite -- that order labels were
// inert text. When the requirement inverted, that check kept passing, because
// it looked for `<a class="ohdr` and the markup had become `<a class="btn ohdr`.
// A test that survives the reversal of its own requirement is worse than none.
//
// Verifies: P-010. Criterion: AC-056.
func TestOrderToggles(t *testing.T) {
	s := newTestServer(t)
	_, body := get(t, s, "/")

	// Each order is a link carrying its common name, not inert text.
	for _, order := range []string{"Apodiformes", "Passeriformes", "Pelecaniformes"} {
		re := regexp.MustCompile(`<a class="btn ohdr[^"]*" href="([^"]+)"[^>]*>` + order)
		m := re.FindStringSubmatch(body)
		if m == nil {
			t.Fatalf("%s is not rendered as a toggle", order)
		}
		if !strings.Contains(body, order+`<em>`+order+" (group)") {
			t.Errorf("%s toggle does not show its common name", order)
		}
	}

	// Selecting an order selects EVERY family in it. Passeriformes holds two,
	// so a control that adds only the first fails here.
	re := regexp.MustCompile(`<a class="btn ohdr[^"]*" href="([^"]+)"[^>]*>Passeriformes`)
	href := html.UnescapeString(re.FindStringSubmatch(body)[1])
	_, filtered := get(t, s, href)
	for _, want := range []string{"American Robin", "American Crow"} {
		if !strings.Contains(filtered, want) {
			t.Errorf("selecting Passeriformes did not include %s", want)
		}
	}
	if strings.Contains(filtered, "Great Blue Heron") {
		t.Error("selecting Passeriformes left another order's species visible")
	}

	// With the group fully selected the same control clears it.
	m2 := regexp.MustCompile(`<a class="btn ohdr on" href="([^"]+)"[^>]*>Passeriformes`).
		FindStringSubmatch(filtered)
	if m2 == nil {
		t.Fatal("Passeriformes is not marked selected once both families are filtered in")
	}
	_, cleared := get(t, s, html.UnescapeString(m2[1]))
	if !strings.Contains(cleared, "Great Blue Heron") {
		t.Error("clearing the order did not restore the unfiltered view")
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

// TestAllToggle asserts that "All" behaves like an order toggle over the whole
// dataset: it selects every family, and untoggling clears the selection.
//
// Note that "nothing selected" and "everything selected" both display all 841
// species, so the control cannot be verified by the species count alone -- the
// selection state is what differs, and what has to be checked.
//
// Verifies: P-010. Criterion: AC-056.
func TestAllToggle(t *testing.T) {
	s := newTestServer(t)

	// A fresh view has no selection, so All is off.
	_, fresh := get(t, s, "/")
	if regexp.MustCompile(`<a class="btn on"[^>]*>All`).MatchString(fresh) {
		t.Error("All is marked selected on a view where nothing is selected")
	}
	href := html.UnescapeString(mustMatch(t, fresh, `<a class="btn"[^>]*href="([^"]+)"[^>]*>All`,
		"the All control on a fresh view"))

	// Clicking it selects every family.
	_, all := get(t, s, href)
	for _, fam := range []string{"Trochilidae", "Turdidae", "Corvidae", "Ardeidae"} {
		if !strings.Contains(all, `family=`+fam) && !strings.Contains(all, fam) {
			t.Errorf("selecting All did not include %s", fam)
		}
	}
	if !regexp.MustCompile(`<a class="btn on"[^>]*>All`).MatchString(all) {
		t.Fatal("All is not marked selected after selecting every family")
	}
	// Every order shows as selected too, since all their families are in.
	if n := strings.Count(all, `class="btn ohdr on"`); n != 3 {
		t.Errorf("%d of 3 orders marked selected when everything is selected", n)
	}

	// Untoggling clears the selection outright, rather than leaving every
	// family behind as redundant parameters.
	clear := html.UnescapeString(mustMatch(t, all, `<a class="btn on"[^>]*href="([^"]+)"[^>]*>All`,
		"the All control once everything is selected"))
	if strings.Contains(clear, "family=") {
		t.Errorf("untoggling All produced %q; it should clear the selection", clear)
	}
	_, cleared := get(t, s, clear)
	if strings.Contains(cleared, `class="btn ohdr on"`) {
		t.Error("orders still show as selected after clearing")
	}
	if !strings.Contains(cleared, "Great Blue Heron") {
		t.Error("clearing the selection should still show every species")
	}
}
