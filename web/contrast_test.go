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
// are held to (tech.badgeContrast), applied to every declared text color against both
// surfaces.
//
// It was 4.5:1 for a day. The two-floor scheme existed only because --dim sat
// at 6.96:1 on the page background and 5.71:1 on a panel, and repainting the
// interface's secondary text looked disproportionate to the link bug that
// prompted the check. The owner then noticed the dim text too, so the colors
// moved and one floor covers everything (decisions.oneContrastFloor).
const textFloor = 7.0

var (
	reVar   = regexp.MustCompile(`(--[a-z-]+):\s*(#[0-9a-fA-F]{3,6})`)
	reRule  = regexp.MustCompile(`(?s)([^{}]+)\{([^{}]*)\}`)
	reColor = regexp.MustCompile(`(?:^|[;\s])color:\s*([^;]+)`)
	reBg    = regexp.MustCompile(`(?:^|[;\s])background(?:-color)?:\s*([^;]+)`)
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

// TestStylesheetContrast asserts tech.textContrastFloor over the shipped stylesheet.
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
// Verifies: tech.textContrastFloor, wcag.ContrastEnhanced. Criterion: acceptance.noTextBelowContrastFloor.
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
		fg := resolve(cm[1])
		if fg == "" { // inherit, currentColor, a keyword -- nothing to measure
			continue
		}
		fr, fg2, fb := hexRGB(t, fg)
		lfg := palette.RelativeLuminance(fr, fg2, fb)

		// A rule that sets its own background is measured against that
		// background, not the page's. This used to skip such rules outright, on
		// the grounds that badges pair a generated background with a generated
		// text color (tech.badgeContrast). Badges do, but the selected button and
		// the pick control pair a hand-written color with --accent, and nothing
		// read those pairs: 2.42:1 on .btn.on passed here, and only the rendered
		// check in cmd/birdview/pages_test.go caught it. A background this file
		// cannot resolve -- none, a gradient -- is still skipped.
		if bm := reBg.FindStringSubmatch(body); bm != nil {
			if bg := resolve(bm[1]); bg != "" {
				br, bgc, bb := hexRGB(t, bg)
				if got := palette.Contrast(lfg, palette.RelativeLuminance(br, bgc, bb)); got < textFloor {
					t.Errorf("%-30s color %s on its own background %s is %.2f:1, below the %.1f:1 floor",
						sel, fg, bg, got, textFloor)
				}
				checked++
			}
			continue
		}

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
// of text (tech.textContrastFloor).
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
// Verifies: tech.textContrastFloor, wcag.ContrastEnhanced. Criterion: acceptance.noTextBelowContrastFloor.
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
// Verifies: tech.textContrastFloor. Criterion: acceptance.noTextBelowContrastFloor.
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

// TestStrokeStyleCarriesProvenance asserts product.measuredVersusModeled where it is actually defined:
// in the stylesheet, one rule per drawn element.
//
// Verifies: product.measuredVersusModeled. Criterion: acceptance.diagramDistinguishesModeled.
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
	// makes a 6px neck render as one lozenge, i.e. solid (decisions.drawingCarriesProvenance).
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
