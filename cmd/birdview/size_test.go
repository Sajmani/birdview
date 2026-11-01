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
	"regexp"
	"testing"
	"time"
)

//go:embed testdata/size.js
var sizeScript string

var dataSize = regexp.MustCompile(`data-size="([^"]*)"`)

// TestRenderedTextSize renders two HTML pages of every kind in headless Chrome, at a laptop's
// width and framed at a phone's 360px, and fails on any visible text drawn below 13px and on
// running text below 15px, or 16px on the phone (tech.textSizeFloor). Sizes are measured as drawn:
// SVG text at its font size times the drawing's scale on screen, which is how a species page's
// diagram labels came to be 3.5px while declared at 11.
//
// It needs Chrome, and skips without it, as TestRenderedTextContrast does.
//
// Verifies: tech.textSizeFloor. Criterion: app.textMeetsSizeFloor.
func TestRenderedTextSize(t *testing.T) {
	chrome := findChrome()
	if chrome == "" {
		t.Skip("no Chrome or Chromium found (set BIRDVIEW_CHROME); text sizes were not measured")
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
	if len(sample) < 5 {
		t.Fatalf("only %d HTML pages to measure; the check needs every kind", len(sample))
	}
	inner := instrument(s, sizeScript, nil)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/__phone" {
			// A frame phoneWidth wide, as in TestPagesFitAPhone.
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			fmt.Fprintf(w, `<!doctype html><html><body style="margin:0">`+
				`<iframe id="f" src="%s" style="border:0;width:%dpx;height:1200px"></iframe>`+
				`<script>setInterval(function () {`+
				`var d = document.getElementById('f').contentDocument;`+
				`var v = d && d.documentElement.getAttribute('data-size');`+
				`if (v) document.documentElement.setAttribute('data-size', v);`+
				`}, 100);</script></body></html>`, html.EscapeString(r.URL.Query().Get("u")), phoneWidth)
			return
		}
		inner.ServeHTTP(w, r)
	}))
	t.Cleanup(ts.Close)
	for _, p := range sample {
		for _, size := range []string{"laptop", "phone"} {
			t.Run(size+p.url, func(t *testing.T) {
				t.Parallel()
				u, body := ts.URL+p.url, 15.0
				if size == "phone" {
					u, body = ts.URL+"/__phone?u="+url.QueryEscape(p.url), 16
				}
				ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
				defer cancel()
				out, err := exec.CommandContext(ctx, chrome, "--headless", "--disable-gpu",
					"--no-first-run", "--no-default-browser-check", "--disable-background-networking",
					"--disable-component-update", "--disable-sync", "--disable-extensions",
					"--user-data-dir="+t.TempDir(), "--window-size=1280,1600",
					"--virtual-time-budget=10000", "--dump-dom", u).Output()
				if err != nil {
					t.Fatalf("chrome: %v", err)
				}
				m := dataSize.FindSubmatch(out)
				if m == nil {
					t.Fatal("no result from the measuring script; it did not run on this page")
				}
				var res struct {
					Width    int     `json:"width"`
					Body     float64 `json:"body"`
					Checked  int     `json:"checked"`
					Failures []struct {
						Text, Element string
						Px            float64
					} `json:"failures"`
				}
				if err := json.Unmarshal([]byte(html.UnescapeString(string(m[1]))), &res); err != nil {
					t.Fatalf("reading the result: %v", err)
				}
				if res.Checked == 0 {
					t.Fatal("no text measured; the check would pass over anything")
				}
				if res.Body < body {
					t.Errorf("running text is %.1fpx at %dpx wide; want %.0fpx", res.Body, res.Width, body)
				}
				for i, f := range res.Failures {
					if i == 10 {
						t.Errorf("... and %d more", len(res.Failures)-i)
						break
					}
					t.Errorf("%.1fpx below 13px at %dpx wide: %q in <%s>", f.Px, res.Width, f.Text, f.Element)
				}
			})
		}
	}
}
