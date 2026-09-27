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
