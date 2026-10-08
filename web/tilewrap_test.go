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

// A tile header that wraps makes its tile taller than its neighbours, and
// .tiles is a flex row, so one tall tile leaves the whole row ragged with the
// diagrams at different offsets. Reading down a section is what the index is
// for (P-031).
//
// This is structural: whether text actually wraps depends on font metrics and
// the viewport, which needs a browser harness this project does not have. What
// is checkable is that the rules making wrapping impossible are present, and
// that the markup gives them something to act on -- which is the shape the bug
// actually had. A declaration was missing, not mismeasured.
//
// Verifies: P-031. Criterion: AC-082.

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
	reCnname   = regexp.MustCompile(`(?s)<span class="cnname">(.*?)</span>`)
)

// TestBandingCodeIsOutsideTheTruncatingElement is the half no stylesheet check
// can see. `.cnname` truncates with an ellipsis, so a banding code nested
// inside it would be the first thing eaten -- and the markup looks correct
// until a name is long enough to prove otherwise.
//
// Verifies: P-031. Criterion: AC-082.
func TestBandingCodeIsOutsideTheTruncatingElement(t *testing.T) {
	for _, name := range []string{"static/index.html", "static/detail.html", "static/compare.html"} {
		b, err := Static.ReadFile(name)
		if err != nil {
			t.Fatalf("reading %s: %v", name, err)
		}
		heads := reNameline.FindAllStringSubmatch(string(b), -1)
		if len(heads) == 0 {
			t.Errorf("%s has no .nameline header; the check cannot bite", name)
			continue
		}
		for _, h := range heads {
			inner := reCnname.FindStringSubmatch(h[1])
			if inner == nil {
				t.Errorf("%s: header has no .cnname, so the common name has nothing to truncate in", name)
				continue
			}
			if strings.Contains(inner[1], `class="code"`) {
				t.Errorf("%s: the banding code is inside .cnname, so the ellipsis eats it first", name)
			}
			if !strings.Contains(h[1], `class="code"`) {
				t.Errorf("%s: header emits no banding code at all", name)
			}
		}
	}
}
