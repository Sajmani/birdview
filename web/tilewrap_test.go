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
	"regexp"
	"strings"
	"testing"
)

// A tile header that wraps makes its tile taller than its neighbors, and
// .tiles is a flex row, so one tall tile leaves the whole row ragged with the
// diagrams at different offsets. Reading down a section is what the index is
// for (product.tileContents).
//
// This is structural: whether text actually wraps depends on font metrics and
// the viewport, which needs a browser harness this project does not have. What
// is checkable is that the rules making wrapping impossible are present, and
// that the markup gives them something to act on -- which is the shape the bug
// actually had. A declaration was missing, not mismeasured.
//
// Verifies: product.tileContents. Criterion: acceptance.tileHeaderNeverWraps.

// rule returns the declaration block of a top-level selector, exactly as
// written, so an assertion can look for a property in it.
func rule(t *testing.T, sel string) string {
	t.Helper()
	css := stripComments(string(mustReadStyles(t)))
	re := regexp.MustCompile(`(?m)^` + regexp.QuoteMeta(sel) + `\s*\{([^}]*)\}`)
	m := re.FindStringSubmatch(css)
	if m == nil {
		t.Fatalf("no top-level %q rule in the stylesheet", sel)
	}
	return m[1]
}

func TestTileHeaderCannotWrap(t *testing.T) {
	for _, c := range []struct{ sel, prop, why string }{
		{".nameline", "flex-wrap: nowrap",
			"the header line could wrap the banding code onto a second row, which is the shipped bug"},
		{".cnname", "white-space: nowrap",
			"a long common name could break across lines inside its own element"},
		{".cnname", "text-overflow: ellipsis",
			"the name would be clipped mid-glyph with no sign it was cut"},
		{".cn .code", "flex: none",
			"the banding code would shrink or truncate; it is the one part a reader cannot reconstruct"},
		{".idline", "flex-wrap: nowrap",
			"the line under a tile's rule could wrap its codes onto a second row"},
		{".idline .code", "flex: none",
			"a code under the rule would shrink or truncate while the scientific name kept its width"},
		{".ft", "flex-wrap: nowrap",
			"the measurements could wrap to a second line and push the drawing down"},
		{".badge", "white-space: nowrap",
			"the family badge could wrap, which makes the tile taller just as surely"},
		{".badge", "text-overflow: ellipsis",
			"nowrap without it would overflow the tile instead of wrapping, which is worse"},
		{".badge", "max-width",
			"without a bound the badge has nothing to truncate against"},
	} {
		if body := rule(t, c.sel); !strings.Contains(body, c.prop) {
			t.Errorf("%s does not declare %q: %s", c.sel, c.prop, c.why)
		}
	}
}

var (
	reNameline = regexp.MustCompile(`(?s)<div class="nameline">(.*?)</div>`)
	reIdline   = regexp.MustCompile(`(?s)<div class="idline">(.*?)</div>`)
	reCnname   = regexp.MustCompile(`(?s)<span class="cnname">(.*?)</span>`)
	reSn       = regexp.MustCompile(`(?s)<span class="sn">(.*?)</span>`)
)

// TestBandingCodeIsOutsideTheTruncatingElement is the half no stylesheet check
// can see. The names truncate with an ellipsis, so a code nested inside one
// would be the first thing eaten -- and the markup looks correct until a name
// is long enough to prove otherwise. The species page keeps the codes beside
// the common name; a tile carries them under its rule, after the scientific
// name (product.tileContents, decisions.tileNumbersFirst).
//
// Verifies: product.tileContents. Criterion: acceptance.tileHeaderNeverWraps.
func TestBandingCodeIsOutsideTheTruncatingElement(t *testing.T) {
	for _, c := range []struct {
		file, what string
		line, name *regexp.Regexp
	}{
		{"static/detail.html", ".nameline", reNameline, reCnname},
		{"static/shared.html", ".idline", reIdline, reSn},
	} {
		b, err := Static.ReadFile(c.file)
		if err != nil {
			t.Fatalf("reading %s: %v", c.file, err)
		}
		lines := c.line.FindAllStringSubmatch(string(b), -1)
		if len(lines) == 0 {
			t.Errorf("%s has no %s line; the check cannot bite", c.file, c.what)
			continue
		}
		for _, h := range lines {
			inner := c.name.FindStringSubmatch(h[1])
			if inner == nil {
				t.Errorf("%s: its %s line has no truncating name element", c.file, c.what)
				continue
			}
			if strings.Contains(inner[1], `class="code`) {
				t.Errorf("%s: a code is inside the truncating name on its %s line, so the ellipsis eats it first", c.file, c.what)
			}
			if !strings.Contains(h[1], `class="code`) {
				t.Errorf("%s: its %s line emits no code at all", c.file, c.what)
			}
		}
	}
}
