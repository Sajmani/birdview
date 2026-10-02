package web

import (
	"regexp"
	"strings"
	"testing"
)

// The favicon is the one surface that cannot reference the stylesheet, so it
// restates the diagram's palette by hand (P-048). Index, detail and comparison
// share one `diagram` template and one stylesheet and cannot diverge at all.
//
// Verifies: P-048. Criterion: AC-081.

var (
	reHexAny   = regexp.MustCompile(`#[0-9a-fA-F]{3,8}`)
	rePaint    = regexp.MustCompile(`(?:^|[;\s])(?:stroke|fill|background(?:-color)?|border-color):\s*([^;]+)`)
	reClassAtt = regexp.MustCompile(`class="([^"]*)"`)
	reAction   = regexp.MustCompile(`{{[^}]*}}`)
	reDiagTpl  = regexp.MustCompile(`(?s){{define "diagram"}}(.*?)\n{{end}}`)
	reVarRef   = regexp.MustCompile(`var\((--[a-z-]+)\)`)
	reXMLComm  = regexp.MustCompile(`(?s)<!--.*?-->`)
)

// diagramParts is the set of class tokens the diagram template actually emits,
// read from the template rather than listed here. A part added to the drawing
// is covered automatically; a list would quietly stop covering it.
func diagramParts(t *testing.T) map[string]bool {
	t.Helper()
	b, err := Static.ReadFile("static/index.html")
	if err != nil {
		t.Fatalf("reading the embedded index template: %v", err)
	}
	m := reDiagTpl.FindStringSubmatch(string(b))
	if m == nil {
		t.Fatal(`no {{define "diagram"}} block; the part set cannot be derived`)
	}
	parts := map[string]bool{}
	for _, c := range reClassAtt.FindAllStringSubmatch(m[1], -1) {
		for _, tok := range strings.Fields(reAction.ReplaceAllString(c[1], " ")) {
			parts[tok] = true
		}
	}
	return parts
}

// diagramPalette is every colour the stylesheet paints a diagram part with,
// plus the surfaces a drawing sits on. This is the set the favicon may draw
// from -- deliberately narrower than "every colour in the stylesheet", so that
// borrowing an interface colour such as --accent counts as drift.
func diagramPalette(t *testing.T) map[string]bool {
	t.Helper()
	css := stripComments(string(mustReadStyles(t)))
	parts := diagramParts(t)

	vars := map[string]string{}
	for _, m := range reVar.FindAllStringSubmatch(css, -1) {
		vars[m[1]] = m[2]
	}
	out := map[string]bool{}
	for _, v := range surfaceVars {
		out[strings.ToLower(vars[v])] = true
	}
	for _, rule := range reRule.FindAllStringSubmatch(css, -1) {
		hit := false
		for _, m := range reTok.FindAllStringSubmatch(rule[1], -1) {
			if parts[m[1]] {
				hit = true
				break
			}
		}
		if !hit {
			continue
		}
		for _, p := range rePaint.FindAllStringSubmatch(rule[2], -1) {
			decl := p[1]
			if m := reVarRef.FindStringSubmatch(decl); m != nil {
				decl = vars[m[1]]
			}
			for _, h := range reHexAny.FindAllString(decl, -1) {
				out[strings.ToLower(h)] = true
			}
		}
	}
	delete(out, "")
	return out
}

// partColor reads one diagram part's paint straight out of the stylesheet, so
// the expected value is never written down twice.
func partColor(t *testing.T, part string) string {
	t.Helper()
	css := stripComments(string(mustReadStyles(t)))
	m := regexp.MustCompile(`(?m)^\.` + part + `\s*\{([^}]*)\}`).FindStringSubmatch(css)
	if m == nil {
		t.Fatalf("no top-level .%s rule in the stylesheet; the palette cannot be derived", part)
	}
	h := reHexAny.FindString(m[1])
	if h == "" {
		t.Fatalf(".%s declares no colour literal; the palette cannot be derived", part)
	}
	return strings.ToLower(h)
}

func TestFaviconUsesTheAppPalette(t *testing.T) {
	icon, err := Static.ReadFile("static/favicon.svg")
	if err != nil {
		t.Fatalf("reading the embedded favicon: %v", err)
	}
	// The icon's comments name colours -- including the superseded one, on
	// purpose -- and a comment is not a usage. stripComments handles CSS
	// /* */; this file is XML.
	body := reXMLComm.ReplaceAllString(string(icon), "")

	palette := diagramPalette(t)
	if len(palette) < 2 {
		t.Fatalf("the stylesheet yielded only %d diagram colours; the comparison cannot bite", len(palette))
	}

	used := map[string]bool{}
	for _, h := range reHexAny.FindAllString(body, -1) {
		used[strings.ToLower(h)] = true
	}
	if len(used) == 0 {
		t.Fatal("the favicon yielded no colours; the comparison cannot bite")
	}

	// No colour of its own: not an invented shade, and not one borrowed from
	// the interface either. The crossbar was painted --accent for days.
	for h := range used {
		if !palette[h] {
			t.Errorf("favicon uses %s, which paints no part of the diagram — "+
				"the icon has a colour of its own instead of the app's (P-048)", h)
		}
	}

	// The parts it does draw take the app's colours for those parts, read from
	// the stylesheet so that changing one there and not here goes red.
	for _, part := range []string{"wing", "bill"} {
		if want := partColor(t, part); !used[want] {
			t.Errorf("favicon does not use the %s colour %s; the icon is behind the app (P-048)",
				part, want)
		}
	}
}
