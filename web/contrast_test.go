package web

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"birdview/internal/palette"
)

// mustReadStyles reads the stylesheet as it is actually shipped -- out of the
// embedded FS, not off disk -- so the check cannot pass against a file the
// binary does not serve.
func mustReadStyles(t *testing.T) []byte {
	t.Helper()
	b, err := Static.ReadFile("static/styles.css")
	if err != nil {
		t.Fatalf("reading the embedded stylesheet: %v", err)
	}
	return b
}

// The two surfaces every unstyled run of text sits on. Parsed from the
// stylesheet rather than hardcoded, so changing the theme cannot silently
// invalidate the check.
var surfaceVars = []string{"--bg", "--panel"}

// textFloor is WCAG 2.2 SC 1.4.6 (Contrast Enhanced), the same 7:1 the badges
// are held to (T-010), applied to every declared text color against both
// surfaces.
//
// It was 4.5:1 for a day. The two-floor scheme existed only because --dim sat
// at 6.96:1 on the page background and 5.71:1 on a panel, and repainting the
// interface's secondary text looked disproportionate to the link bug that
// prompted the check. The owner then noticed the dim text too, so the colors
// moved and one floor covers everything (CR-014).
const textFloor = 7.0

var (
	reVar   = regexp.MustCompile(`(--[a-z-]+):\s*(#[0-9a-fA-F]{3,6})`)
	reRule  = regexp.MustCompile(`(?s)([^{}]+)\{([^{}]*)\}`)
	reColor = regexp.MustCompile(`(?:^|[;\s])color:\s*([^;]+)`)
	reBg    = regexp.MustCompile(`(?:^|[;\s])background(?:-color)?:\s*`)
	reHex   = regexp.MustCompile(`#[0-9a-fA-F]{3,6}`)
)

func hexRGB(t *testing.T, h string) (uint8, uint8, uint8) {
	t.Helper()
	if len(h) == 4 { // #abc -> #aabbcc
		h = fmt.Sprintf("#%c%c%c%c%c%c", h[1], h[1], h[2], h[2], h[3], h[3])
	}
	var r, g, b uint8
	if _, err := fmt.Sscanf(h, "#%02x%02x%02x", &r, &g, &b); err != nil {
		t.Fatalf("unreadable color %q: %v", h, err)
	}
	return r, g, b
}

// TestStylesheetContrast asserts T-023 over the shipped stylesheet.
//
// The bug this exists for: the iNaturalist link was added with no color rule,
// so it fell back to the browser default -- #0000EE on a #0f172a background,
// which is 1.90:1, and #551A8B once visited, which is 1.62:1. Both are
// unreadable, and neither was caught by anything, because every contrast check
// this project had was scoped to family badges.
//
// Two assertions follow from that. Every color the stylesheet declares for text
// on a bare surface must clear the floor, and no link may go uncolored at all --
// a rule that only checks declared colors cannot see the color nobody declared.
//
// Verifies: T-023, `wcag/1.4.3`, `wcag/1.4.6`. Criterion: AC-070.
func TestStylesheetContrast(t *testing.T) {
	css := stripComments(string(mustReadStyles(t)))

	// Resolve the custom properties first.
	vars := map[string]string{}
	for _, m := range reVar.FindAllStringSubmatch(css, -1) {
		vars[m[1]] = m[2]
	}
	for _, v := range surfaceVars {
		if _, ok := vars[v]; !ok {
			t.Fatalf("stylesheet declares no %s; the check cannot know what text sits on", v)
		}
	}

	resolve := func(decl string) string {
		if m := regexp.MustCompile(`var\((--[a-z-]+)\)`).FindStringSubmatch(decl); m != nil {
			return vars[m[1]]
		}
		if h := reHex.FindString(decl); h != "" {
			return h
		}
		return ""
	}

	var checked int
	for _, rule := range reRule.FindAllStringSubmatch(css, -1) {
		sel, body := strings.TrimSpace(rule[1]), rule[2]
		if strings.HasPrefix(sel, "@") || strings.Contains(sel, ":root") {
			continue
		}
		cm := reColor.FindStringSubmatch(body)
		if cm == nil {
			continue
		}
		// A rule that sets its own background is self-contained: badges and
		// chips pair a generated background with a matching text color and are
		// governed by T-010 and AC-002 instead.
		if reBg.MatchString(body) {
			continue
		}
		fg := resolve(cm[1])
		if fg == "" { // inherit, currentColor, a keyword -- nothing to measure
			continue
		}
		fr, fg2, fb := hexRGB(t, fg)
		lfg := palette.RelativeLuminance(fr, fg2, fb)

		// One floor, both surfaces, links and text alike.
		for _, v := range surfaceVars {
			br, bgc, bb := hexRGB(t, vars[v])
			if got := palette.Contrast(lfg, palette.RelativeLuminance(br, bgc, bb)); got < textFloor {
				t.Errorf("%-30s color %s on %s is %.2f:1, below the %.1f:1 floor",
					sel, fg, v, got, textFloor)
			}
			checked++
		}
	}
	if checked == 0 {
		t.Fatal("no color declarations were checked; the parser is broken, not the stylesheet")
	}
	t.Logf("%d color/surface pairs checked", checked)
}

// reFill matches a fill declaration, which is how SVG text is painted. SVG
// elements ignore `color` entirely, so the check above cannot see a diagram
// label at all.
var (
	reFill = regexp.MustCompile(`(?:^|[;\s])fill:\s*([^;]+)`)
	reFont = regexp.MustCompile(`(?:^|[;\s])font-(?:size|weight|family|style):`)
	reTok  = regexp.MustCompile(`\.([A-Za-z][-\w]*)`)
)

// TestSVGTextContrast holds diagram labels to the same floor as every other run
// of text (T-023).
//
// The bug: `.dlabel.chord, .dlabel.span` painted the wing chord and wingspan
// labels #0ea5e9, which is 6.44:1 on the page background and 5.28:1 on a panel
// against a 7:1 floor. It shipped, and TestStylesheetContrast could not see it,
// because that check matches `color:` and SVG text has no `color`.
//
// Which rules style text is DERIVED, not listed: a class token that any rule
// gives a font property to is a text token, and a rule naming one is styling
// text. An explicit allowlist would have to be maintained by whoever just added
// the selector it is missing, which is the failure being fixed rather than a
// fix for it.
//
// Verifies: T-023, `wcag/1.4.6`. Criterion: AC-070.
func TestSVGTextContrast(t *testing.T) {
	css := stripComments(string(mustReadStyles(t)))

	vars := map[string]string{}
	for _, m := range reVar.FindAllStringSubmatch(css, -1) {
		vars[m[1]] = m[2]
	}

	// Pass one: which class tokens does the stylesheet treat as text?
	textTokens := map[string]bool{}
	for _, rule := range reRule.FindAllStringSubmatch(css, -1) {
		if !reFont.MatchString(rule[2]) {
			continue
		}
		for _, m := range reTok.FindAllStringSubmatch(rule[1], -1) {
			textTokens[m[1]] = true
		}
	}
	if len(textTokens) == 0 {
		t.Fatal("no text tokens found; the parser is broken, not the stylesheet")
	}

	// Pass two: every fill on a rule naming one of them is a text color.
	var checked int
	for _, rule := range reRule.FindAllStringSubmatch(css, -1) {
		sel, body := strings.TrimSpace(rule[1]), rule[2]
		if strings.HasPrefix(sel, "@") || strings.Contains(sel, ":root") {
			continue
		}
		fm := reFill.FindStringSubmatch(body)
		if fm == nil {
			continue
		}
		isText := false
		for _, m := range reTok.FindAllStringSubmatch(sel, -1) {
			if textTokens[m[1]] {
				isText = true
				break
			}
		}
		if !isText {
			continue // a shape: WCAG 1.4.11's 3:1, not this floor
		}
		var fg string
		if m := regexp.MustCompile(`var\((--[a-z-]+)\)`).FindStringSubmatch(fm[1]); m != nil {
			fg = vars[m[1]]
		} else if h := reHex.FindString(fm[1]); h != "" {
			fg = h
		}
		if fg == "" || len(fg) == 9 { // none/inherit, or #rrggbbaa -- a tint, not text
			continue
		}
		fr, fg2, fb := hexRGB(t, fg)
		lfg := palette.RelativeLuminance(fr, fg2, fb)
		for _, v := range surfaceVars {
			br, bgc, bb := hexRGB(t, vars[v])
			if got := palette.Contrast(lfg, palette.RelativeLuminance(br, bgc, bb)); got < textFloor {
				t.Errorf("%-30s fill %s on %s is %.2f:1, below the %.1f:1 floor",
					sel, fg, v, got, textFloor)
			}
			checked++
		}
	}
	if checked == 0 {
		t.Fatal("no SVG text fills were checked; the derivation found nothing, so it cannot bite")
	}
	t.Logf("%d fill/surface pairs checked across %d text tokens", checked, len(textTokens))
}

// TestNoLinkUsesTheBrowserDefault is the other half, and the half that would
// have caught the actual bug.
//
// A contrast check over declared colors is blind to a link with no rule, which
// is precisely how an unreadable link shipped. So this asserts the stylesheet
// carries a bare `a` rule and a bare `a:visited` rule -- the catch-alls that
// make "someone forgot" impossible rather than merely unlikely.
//
// Verifies: T-023. Criterion: AC-070.
func TestNoLinkUsesTheBrowserDefault(t *testing.T) {
	css := string(mustReadStyles(t))
	for _, want := range []struct{ sel, why string }{
		{"a {", "an unstyled link falls back to #0000EE, which is 1.9:1 here"},
		{"a:visited {", "an unstyled visited link falls back to #551A8B, which is 1.6:1 here"},
	} {
		if !regexp.MustCompile(`(?m)^` + regexp.QuoteMeta(want.sel)).MatchString(css) {
			t.Errorf("stylesheet has no top-level `%s` rule: %s", want.sel, want.why)
		}
	}
}

// stripComments removes CSS comments before parsing. Without this the selector
// capture swallows the comment preceding a rule, and any rule documented with a
// sentence containing an "a" is misread as a link selector -- which is exactly
// what happened on the first run of this test.
func stripComments(css string) string {
	return regexp.MustCompile(`(?s)/\*.*?\*/`).ReplaceAllString(css, "")
}

// TestStrokeStyleCarriesProvenance asserts P-065 where it is actually defined:
// in the stylesheet, one rule per drawn element.
//
// Verifies: P-065. Criterion: AC-072.
func TestStrokeStyleCarriesProvenance(t *testing.T) {
	css := stripComments(string(mustReadStyles(t)))

	decl := func(sel string) string {
		m := regexp.MustCompile(regexp.QuoteMeta(sel) + `\s*\{([^}]*)\}`).FindStringSubmatch(css)
		if m == nil {
			t.Fatalf("stylesheet has no rule for %q", sel)
		}
		return m[1]
	}
	dashed := func(sel string) bool { return strings.Contains(decl(sel), "stroke-dasharray") }

	// Modeled elements are dashed.
	for _, sel := range []string{".dia .neck", ".dia .body", ".dia .wing.est"} {
		if !dashed(sel) {
			t.Errorf("%s is modeled but carries no stroke-dasharray", sel)
		}
	}
	// The head is hollow, since a dash means nothing on a closed shape.
	if !strings.Contains(decl(".dia .head"), "fill: none") {
		t.Error(".dia .head is modeled but is not drawn hollow")
	}
	// Measured elements are not dashed. The tail is the one that matters: dash
	// used to mean "this is the tail", and anyone restoring that on aesthetic
	// grounds inverts the meaning of the whole scheme while every other
	// assertion here still passes.
	if d := decl(".dia .tail"); !strings.Contains(d, "stroke-dasharray: none") {
		t.Errorf(".dia .tail must explicitly clear its dash: it is a MEASURED length, and a dashed "+
			"tail reverses what dash means everywhere else. Got %q", strings.TrimSpace(d))
	}
	if strings.Contains(decl(".bill"), "stroke-dasharray") {
		t.Error(".bill is measured but is dashed")
	}

	// The neck's pattern differs from the body's, deliberately: matching them
	// makes a 6px neck render as one lozenge, i.e. solid (CR-015).
	pat := func(sel string) string {
		m := regexp.MustCompile(`stroke-dasharray:\s*([^;]+)`).FindStringSubmatch(decl(sel))
		if m == nil {
			return ""
		}
		return strings.TrimSpace(m[1])
	}
	if n, b := pat(".dia .neck"), pat(".dia .body"); n == b {
		t.Errorf("neck and body share the dash pattern %q; the neck needs a finer one, or short "+
			"necks paint as a single dash and read solid", n)
	}
}

// TestControlHelpEntriesAreUnbreakable asserts that a column break cannot land
// between a control's heading and the text under it.
//
// This is not a style preference. The control help is laid out in two columns,
// and "Section step" shipped with its heading at the foot of the first column
// and its explanation at the head of the second, reading as though the text
// belonged to whatever came next.
//
// `break-inside: avoid` on the dd is NOT sufficient and was what was there
// before: it keeps the paragraph whole while still allowing the dt to be
// orphaned above it. The pair has to be one unbreakable box.
//
// Verifies: P-035. Criterion: AC-059.
func TestControlHelpEntriesAreUnbreakable(t *testing.T) {
	css := stripComments(string(mustReadStyles(t)))
	m := regexp.MustCompile(`\.ctrlhelp \.ctrl\s*\{([^}]*)\}`).FindStringSubmatch(css)
	if m == nil {
		t.Fatal("no rule for .ctrlhelp .ctrl; control entries are not grouped into unbreakable units")
	}
	if !strings.Contains(m[1], "break-inside: avoid") {
		t.Errorf(".ctrlhelp .ctrl does not set break-inside: avoid, so a column break can "+
			"separate a control's heading from its text. Got %q", strings.TrimSpace(m[1]))
	}
	// The multi-column context has to exist, or the rule above guards nothing.
	if c := regexp.MustCompile(`\.ctrlhelp\s*\{([^}]*)\}`).FindStringSubmatch(css); c == nil ||
		!strings.Contains(c[1], "columns") {
		t.Error(".ctrlhelp is not a multi-column layout; the unbreakable rule is guarding nothing")
	}
}

// TestOrderGroupsIndentWrappedFamilies asserts that a long family list wraps
// into the column beside its order header rather than back to the page margin.
//
// Passeriformes has 44 families and wraps over seven lines. As siblings of the
// header in one wrapping flex row, every line after the first began at the page
// margin, under the order toggle, so the block read as a wall of chips with no
// relationship to the order heading it belonged to.
//
// Both halves are asserted here and in the markup test, because either alone is
// useless: the families need their own box, and the group needs to be a
// two-column grid for that box to be a column of.
//
// Verifies: P-010. Criterion: AC-075.
func TestOrderGroupsIndentWrappedFamilies(t *testing.T) {
	css := stripComments(string(mustReadStyles(t)))

	g := regexp.MustCompile(`\.ogroup\s*\{([^}]*)\}`).FindStringSubmatch(css)
	if g == nil {
		t.Fatal("no rule for .ogroup")
	}
	if !strings.Contains(g[1], "grid") {
		t.Errorf(".ogroup is not a grid, so the families have no column to wrap inside "+
			"and a wrapped line returns to the page margin. Got %q", strings.TrimSpace(g[1]))
	}
	// Three tracks: the order toggle, the Only button (P-072), then the
	// families. Fewer would push a child onto its own row, and the last track
	// must stay the flexible one or the chips get a fixed width to wrap in.
	cols := regexp.MustCompile(`grid-template-columns:\s*([^;}]+)`).FindStringSubmatch(g[1])
	if cols == nil {
		t.Fatalf(".ogroup sets no grid-template-columns; got %q", strings.TrimSpace(g[1]))
	}
	if fields := strings.Fields(cols[1]); len(fields) != 3 {
		t.Errorf(".ogroup has %d grid columns (%q), want 3: the order header, Only, and the families",
			len(fields), strings.TrimSpace(cols[1]))
	} else if fields[2] != "1fr" {
		t.Errorf(".ogroup's last column is %q, want 1fr: the families take the slack, "+
			"and a fixed last track gives wrapped chips the wrong width", fields[2])
	}

	// The families' own box must wrap, or a long list overflows the column
	// instead of running onto a second line inside it.
	f := regexp.MustCompile(`\.ofams\s*\{([^}]*)\}`).FindStringSubmatch(css)
	if f == nil {
		t.Fatal("no rule for .ofams; the families are not a box that can wrap on their own")
	}
	if !strings.Contains(f[1], "wrap") {
		t.Errorf(".ofams does not wrap, so a long family list cannot break onto a second "+
			"line inside its column. Got %q", strings.TrimSpace(f[1]))
	}
}

// TestComparePanelsWrap asserts the layout half of the comparison view: the
// panels flow onto as many lines as they need.
//
// P-067 requires this rather than treating it as taste. With several birds the
// panels cannot fit on one line at a readable size, and a row that does not
// wrap either overflows the page or squeezes the diagrams until the comparison
// they exist to make is gone.
//
// Verifies: P-067. Criterion: AC-076.
func TestComparePanelsWrap(t *testing.T) {
	css := stripComments(string(mustReadStyles(t)))

	row := regexp.MustCompile(`\.cmprow\s*\{([^}]*)\}`).FindStringSubmatch(css)
	if row == nil {
		t.Fatal("no rule for .cmprow; the comparison panels are not laid out as a row")
	}
	if !strings.Contains(row[1], "wrap") {
		t.Errorf(".cmprow does not wrap, so several birds cannot flow onto further lines. Got %q",
			strings.TrimSpace(row[1]))
	}
	// A wrapping row of items that cannot shrink or share the width is still a
	// single overflowing line, so the panel needs its own flexible basis.
	panel := regexp.MustCompile(`\.cmppanel\s*\{([^}]*)\}`).FindStringSubmatch(css)
	if panel == nil {
		t.Fatal("no rule for .cmppanel; the panels have no width of their own and inherit " +
			"the detail view's fixed one")
	}
	if !strings.Contains(panel[1], "flex:") {
		t.Errorf(".cmppanel sets no flex basis, so the panels cannot share a line. Got %q",
			strings.TrimSpace(panel[1]))
	}
}
