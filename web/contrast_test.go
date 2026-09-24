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

// linkFloor is WCAG 2.2 SC 1.4.6, the same 7:1 the badges are held to (T-010),
// and it is applied against the page background because that is where every
// plain link in BirdView actually sits. Links are interactive, so they get the
// strict floor; nothing in this palette makes it hard to clear.
const linkFloor = 7.0

// textFloor is SC 1.4.3, AA, applied to every declared text color against both
// surfaces. Secondary text -- scientific names, provenance detail, counts --
// clears AA comfortably but not AAA: --dim is 6.96:1 on the page background and
// 5.71:1 on a panel. That gap is the one place BirdView's contrast claim is
// weaker than its badge claim, and T-023 states it rather than rounding it up.
const textFloor = 4.5

// linkSurface is the surface plain links appear on. Verified: no bare <a> sits
// inside .help or .controls, whose background is --panel.
const linkSurface = "--bg"

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

		// Every text color must clear AA on both surfaces...
		for _, v := range surfaceVars {
			br, bgc, bb := hexRGB(t, vars[v])
			if got := palette.Contrast(lfg, palette.RelativeLuminance(br, bgc, bb)); got < textFloor {
				t.Errorf("%-30s color %s on %s is %.2f:1, below the %.1f:1 text floor",
					sel, fg, v, got, textFloor)
			}
			checked++
		}
		// ...and a link must clear AAA on the surface links live on.
		if isLinkSelector(sel) {
			br, bgc, bb := hexRGB(t, vars[linkSurface])
			if got := palette.Contrast(lfg, palette.RelativeLuminance(br, bgc, bb)); got < linkFloor {
				t.Errorf("%-30s link color %s on %s is %.2f:1, below the %.1f:1 link floor",
					sel, fg, linkSurface, got, linkFloor)
			}
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

// isLinkSelector reports whether a selector targets an anchor. Matching on a
// bare "a" token rather than a substring, so ".seealso" and ".back" do not
// count themselves.
func isLinkSelector(sel string) bool {
	for _, part := range regexp.MustCompile(`[\s,>+~]+`).Split(sel, -1) {
		base, _, _ := strings.Cut(part, ":")
		if base == "a" {
			return true
		}
	}
	return false
}
