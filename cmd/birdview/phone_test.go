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

package main

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"html"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os/exec"
	"testing"
	"time"
)

//go:embed testdata/fit.js
var fitScript string

// phoneWidth is the narrowest window product.fitsAPhone holds pages to, in CSS
// pixels: a small Android phone held upright.
const phoneWidth = 360

// fitResult is what fit.js reports for one page.
type fitResult struct {
	Width     int `json:"width"`
	PageWidth int `json:"pageWidth"`
	Checked   int `json:"checked"`
	Failures  []struct {
		Kind     string `json:"kind"`
		Element  string `json:"element"`
		Box      string `json:"box"`
		Text     string `json:"text"`
		Right    int    `json:"right"`
		BoxRight int    `json:"boxRight"`
	} `json:"failures"`
}

// TestPagesFitAPhone renders two HTML pages of every kind in headless Chrome
// in a window phoneWidth pixels wide, with every <details> open, and fails on
// anything that reaches past the window's right edge, so that the page would
// scroll sideways, and on any text drawn outside the box it sits in.
//
// It needs Chrome, and skips without it, as TestRenderedTextContrast does.
//
// Verifies: product.fitsAPhone. Criterion: app.pagesFitAPhone.
func TestPagesFitAPhone(t *testing.T) {
	chrome := findChrome()
	if chrome == "" {
		t.Skip("no Chrome or Chromium found (set BIRDVIEW_CHROME); the phone-width check did not run")
	}
	s := newTestServer(t)
	var sample []webPage
	perKind := map[string]int{}
	for _, p := range discoverPages(t, s) {
		if p.is("text/html") && perKind[p.kind()] < 2 {
			perKind[p.kind()]++
			sample = append(sample, p)
		}
	}
	if len(sample) == 0 {
		t.Fatal("no HTML pages discovered; this check proves nothing")
	}

	inner := instrument(s, fitScript, nil)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/__phone" {
			// Chrome will not open a window narrower than 500 pixels, so the
			// page is drawn in a frame phoneWidth wide, which is a viewport of
			// its own, and fit.js's result is copied out of it.
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			fmt.Fprintf(w, `<!doctype html><html><body style="margin:0">`+
				`<iframe id="f" src="%s" style="border:0;width:%dpx;height:800px"></iframe>`+
				`<script>setInterval(function () {`+
				`var d = document.getElementById('f').contentDocument;`+
				`var v = d && d.documentElement.getAttribute('data-contrast');`+
				`if (v) document.documentElement.setAttribute('data-contrast', v);`+
				`}, 100);</script></body></html>`, html.EscapeString(r.URL.Query().Get("u")), phoneWidth)
			return
		}
		inner.ServeHTTP(w, r)
	}))
	t.Cleanup(ts.Close) // not defer: the parallel subtests run after this returns
	for _, p := range sample {
		t.Run(p.url, func(t *testing.T) {
			t.Parallel()
			ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
			defer cancel()
			out, err := exec.CommandContext(ctx, chrome, "--headless", "--disable-gpu",
				"--no-first-run", "--no-default-browser-check", "--disable-background-networking",
				"--disable-component-update", "--disable-sync", "--disable-extensions",
				"--hide-scrollbars", "--user-data-dir="+t.TempDir(), "--window-size=800,900",
				"--virtual-time-budget=10000", "--dump-dom",
				ts.URL+"/__phone?u="+url.QueryEscape(p.url)).Output()
			if err != nil {
				t.Fatalf("chrome: %v", err)
			}
			m := dataContrast.FindSubmatch(out)
			if m == nil {
				t.Fatalf("no result from the measuring script; it did not run on this page")
			}
			var res fitResult
			if err := json.Unmarshal([]byte(html.UnescapeString(string(m[1]))), &res); err != nil {
				t.Fatalf("reading the result: %v", err)
			}
			if res.Width != phoneWidth {
				t.Fatalf("the page was laid out %d pixels wide, not %d; the check would prove nothing", res.Width, phoneWidth)
			}
			t.Logf("%d runs of text in a box measured", res.Checked)
			for i, f := range res.Failures {
				if i == 10 {
					t.Errorf("... and %d more", len(res.Failures)-i)
					break
				}
				switch f.Kind {
				case "wide":
					t.Errorf("<%s> reaches %dpx, past the %dpx window: %q", f.Element, f.Right, res.Width, f.Text)
				default:
					t.Errorf("%q in <%s> reaches %dpx, outside its box <%s>, which ends at %dpx", f.Text, f.Element, f.Right, f.Box, f.BoxRight)
				}
			}
			if res.PageWidth > res.Width {
				t.Errorf("the page is %dpx wide in a %dpx window, so it scrolls sideways", res.PageWidth, res.Width)
			}
		})
	}
}
