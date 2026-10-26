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

// The checks for app.pagesMeetTextContrast. The requirement holds every web
// page to WCAG 1.4.6 without listing the pages, so these find the pages
// themselves: they crawl the server, and fail if a route or a template went
// unvisited. Each page found is then either inside app.stylesheetContract,
// whose guarantee is checked once in web/contrast_test.go, or outside it and
// rendered in full. A sample of every kind of page is rendered regardless, so
// a contract that is wrong cannot hide behind its own checks.

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"html"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
	"time"
)

// webPage is one response the checks found on the server.
type webPage struct {
	url, route, ctype string
	// template is the page template that rendered it, empty for a response
	// that did not come from one.
	template string
	body     string
}

func (p webPage) is(media string) bool { return strings.HasPrefix(p.ctype, media) }

// kind is what makes two pages the same for sampling: the route, and the
// template or, failing one, the content type.
func (p webPage) kind() string {
	if p.template != "" {
		return p.route + " " + p.template
	}
	mt, _, _ := strings.Cut(p.ctype, ";")
	return p.route + " " + mt
}

// isWebPage reports whether a response is one tech.webPages names.
func isWebPage(ctype string) bool {
	mt, _, _ := strings.Cut(ctype, ";")
	switch strings.TrimSpace(mt) {
	case "text/html", "image/svg+xml", "text/plain":
		return true
	}
	return false
}

// routeOf is the pattern that serves a path, by the ServeMux's own rule: an
// exact pattern matches itself, one ending in a slash matches the subtree,
// and the longest match wins.
func routeOf(s *server, path string) string {
	best := ""
	for _, rt := range s.routes {
		match := path == rt.pattern ||
			strings.HasSuffix(rt.pattern, "/") && strings.HasPrefix(path, rt.pattern)
		if match && len(rt.pattern) > len(best) {
			best = rt.pattern
		}
	}
	return best
}

var pageLink = regexp.MustCompile(`(?:href|src)="(/[^"]*)"`)

// discoverPages crawls the server from a handful of seeds, following every
// same-origin link and source, and returns every web page it reached. The
// seeds are the pages nothing links to: pick mode, an undrawn bird, a page
// that does not exist. A route is visited at most perRoute times, since every
// control on the index links to another variant of the index.
func discoverPages(t *testing.T, s *server) []webPage {
	t.Helper()
	const perRoute = 12
	var rendered string
	s.onRender = func(name string) { rendered = name }
	defer func() { s.onRender = nil }()

	queue := []string{"/classic", "/classic?species=", "/classic?species=amecro&species=annhum&family=Corvidae",
		"/docs", "/added", "/excluded", "/compare?species=amecro&species=annhum", "/",
		"/species/" + undrawnFixture()[0].ID, "/no-such-page", "/robots.txt"}
	seen := map[string]bool{}
	visits := map[string]int{}
	var pages []webPage
	for len(queue) > 0 {
		u := queue[0]
		queue = queue[1:]
		if seen[u] {
			continue
		}
		seen[u] = true
		path, _, _ := strings.Cut(u, "?")
		rt := routeOf(s, path)
		if visits[rt] >= perRoute {
			continue
		}
		rendered = ""
		w := httptest.NewRecorder()
		s.ServeHTTP(w, httptest.NewRequest(http.MethodGet, u, nil))
		ctype := w.Header().Get("Content-Type")
		if w.Code/100 == 3 || !isWebPage(ctype) {
			continue // a redirect is followed, not read; a stylesheet is not a page
		}
		visits[rt]++
		p := webPage{url: u, route: rt, ctype: ctype, template: rendered, body: w.Body.String()}
		pages = append(pages, p)
		if !p.is("text/html") {
			continue
		}
		for _, m := range pageLink.FindAllStringSubmatch(p.body, -1) {
			l := html.UnescapeString(m[1])
			l, _, _ = strings.Cut(l, "#")
			if !strings.HasPrefix(l, "//") && !seen[l] {
				queue = append(queue, l)
			}
		}
	}
	return pages
}

// noPages are the routes that serve no web page, and why. Each is probed to
// confirm it, so an exemption cannot hide a page added there later.
var noPages = map[string]string{
	"/now":         "a redirect to the classic view for the current month",
	"/boundaries/": "the region boundary file, JSON",
	"/birds.json":  "the names the home page's find box searches, JSON",
}

// TestEveryKindOfPageIsDiscovered checks the discovery the other page checks
// rely on: every route the server registers, and every page template it can
// render, produced at least one page. A route or template added later that no
// seed or link reaches fails here, rather than going unchecked.
//
// Verifies: tech.webPages. Criterion: app.everyPageIsFound.
func TestEveryKindOfPageIsDiscovered(t *testing.T) {
	s := newTestServer(t)
	pages := discoverPages(t, s)
	routes := map[string]bool{}
	templates := map[string]bool{}
	media := map[string]bool{}
	for _, p := range pages {
		routes[p.route] = true
		templates[p.template] = true
		mt, _, _ := strings.Cut(p.ctype, ";")
		media[mt] = true
	}
	for _, rt := range s.routes {
		if why, exempt := noPages[rt.pattern]; exempt {
			// Exempt only while it really serves no page.
			w := httptest.NewRecorder()
			probe := rt.pattern
			if probe == "/boundaries/" {
				probe = s.boundaryURL
			}
			s.ServeHTTP(w, httptest.NewRequest(http.MethodGet, probe, nil))
			if w.Code/100 != 3 && isWebPage(w.Header().Get("Content-Type")) {
				t.Errorf("%s is exempt as %q but served a %s page", rt.pattern, why, w.Header().Get("Content-Type"))
			}
			continue
		}
		if !routes[rt.pattern] {
			t.Errorf("no page discovered under route %s; add a seed so its pages are checked", rt.pattern)
		}
	}
	names, err := fs.Glob(assetFS, "static/*.html")
	if err != nil || len(names) == 0 {
		t.Fatalf("listing templates: %v (%d found)", err, len(names))
	}
	for _, n := range names {
		if !templates[filepath.Base(n)] {
			t.Errorf("template %s rendered no discovered page; add a seed so its pages are checked",
				filepath.Base(n))
		}
	}
	for _, mt := range []string{"text/html", "image/svg+xml", "text/plain"} {
		if !media[mt] {
			t.Errorf("no %s page discovered; tech.webPages names it, so the crawl has lost one", mt)
		}
	}
	t.Logf("%d pages across %d routes and %d templates", len(pages), len(routes), len(names))
}

var (
	styleAttr   = regexp.MustCompile(`\sstyle="([^"]*)"`)
	colorAttr   = regexp.MustCompile(`\s(fill|stroke|color|bgcolor|stop-color)="([^"]*)"`)
	colorProp   = regexp.MustCompile(`^(color|background|background-color|border|border-color|fill|stroke|outline|outline-color|box-shadow|text-shadow)$`)
	stylesheet  = `<link rel="stylesheet" href="/static/styles.css">`
	noColorAttr = map[string]bool{"none": true, "currentcolor": true}
)

// TestPagesKeepToTheStylesheetContract checks app.stylesheetContract's
// assumption on every page rendered from a template: it links the stylesheet,
// has no style element, and sets a color inline only as a badge's generated
// pair. The contract is closed, so there is no list of exempt pages. An SVG
// document served on its own is outside the contract by its scope, and
// TestRenderedTextContrast renders every one of those in full.
//
// Verifies: tech.textContrastFloor, tech.badgeContrast. Criterion: app.pagesKeepToTheStylesheet.
func TestPagesKeepToTheStylesheetContract(t *testing.T) {
	s := newTestServer(t)
	palette := map[string]bool{}
	for _, f := range s.fams {
		palette[strings.ToLower(f.Color)] = true
		palette[strings.ToLower(f.TextColor)] = true
	}
	if len(palette) == 0 {
		t.Fatal("no palette colors found; the fixture or the server has changed and this check proves nothing")
	}
	checked := 0
	for _, p := range discoverPages(t, s) {
		if p.template == "" {
			continue
		}
		checked++
		if !strings.Contains(p.body, stylesheet) {
			t.Errorf("%s (%s) does not link the stylesheet", p.url, p.template)
		}
		if strings.Contains(p.body, "<style") {
			t.Errorf("%s (%s) has a <style> element, a color source outside the contract", p.url, p.template)
		}
		for _, m := range styleAttr.FindAllStringSubmatch(p.body, -1) {
			for _, decl := range strings.Split(html.UnescapeString(m[1]), ";") {
				prop, val, ok := strings.Cut(decl, ":")
				prop, val = strings.TrimSpace(strings.ToLower(prop)), strings.TrimSpace(strings.ToLower(val))
				if ok && colorProp.MatchString(prop) && !palette[val] {
					t.Errorf("%s (%s) sets %s: %s inline, which is not a generated badge color",
						p.url, p.template, prop, val)
				}
			}
		}
		for _, m := range colorAttr.FindAllStringSubmatch(p.body, -1) {
			if !noColorAttr[strings.ToLower(m[2])] {
				t.Errorf("%s (%s) sets %s=%q as an attribute, a color source outside the contract",
					p.url, p.template, m[1], m[2])
			}
		}
	}
	if checked == 0 {
		t.Fatal("no templated pages discovered; this check proves nothing")
	}
}

//go:embed testdata/contrast.js
var contrastScript string

// contrastResult is what contrast.js reports for one page.
type contrastResult struct {
	Checked    int `json:"checked"`
	Unmeasured int `json:"unmeasured"`
	Failures   []struct {
		Text       string  `json:"text"`
		Element    string  `json:"element"`
		Color      string  `json:"color"`
		Background string  `json:"background"`
		Ratio      float64 `json:"ratio"`
	} `json:"failures"`
}

// findChrome returns a Chrome or Chromium binary, or "" if there is none.
// BIRDVIEW_CHROME overrides the search.
func findChrome() string {
	if c := os.Getenv("BIRDVIEW_CHROME"); c != "" {
		return c
	}
	for _, c := range []string{"/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
		"/Applications/Chromium.app/Contents/MacOS/Chromium"} {
		if _, err := os.Stat(c); err == nil {
			return c
		}
	}
	for _, c := range []string{"google-chrome", "chromium", "chromium-browser", "chrome"} {
		if p, err := exec.LookPath(c); err == nil {
			return p
		}
	}
	return ""
}

var dataContrast = regexp.MustCompile(`data-contrast="([^"]*)"`)

// TestRenderedTextContrast renders pages in headless Chrome and measures the
// contrast of every piece of visible text against what is actually drawn
// behind it. It renders a sample of two pages of every kind, and every page
// outside app.stylesheetContract in full. A plain-text page is not rendered:
// it sets no colors and is drawn in the reader's own (wcag.ContrastRatio).
//
// This is the end-to-end check on the contract. It catches what the contract's
// own checks cannot -- a color the stylesheet check does not read, an override,
// a background nobody declared -- and it is what found that the pipeline diagram
// served on its own was outside every check.
//
// It needs Chrome, and skips without it; on the owner's machine it gates, as
// river/bindings.md records.
//
// Verifies: wcag.ContrastEnhanced. Criterion: app.renderedTextContrast.
func TestRenderedTextContrast(t *testing.T) {
	chrome := findChrome()
	if chrome == "" {
		t.Skip("no Chrome or Chromium found (set BIRDVIEW_CHROME); the rendered contrast check did not run")
	}
	s := newTestServer(t)
	var sample []webPage
	perKind := map[string]int{}
	for _, p := range discoverPages(t, s) {
		switch {
		case p.is("text/plain"):
		case p.template == "":
			sample = append(sample, p) // outside the contract: every one
		case perKind[p.kind()] < 2:
			perKind[p.kind()]++
			sample = append(sample, p)
		}
	}

	ts := httptest.NewServer(instrument(s, contrastScript, nil))
	t.Cleanup(ts.Close) // not defer: the parallel subtests run after this returns
	for _, p := range sample {
		t.Run(p.url, func(t *testing.T) {
			t.Parallel()
			ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
			defer cancel()
			out, err := exec.CommandContext(ctx, chrome, "--headless", "--disable-gpu",
				"--no-first-run", "--no-default-browser-check", "--disable-background-networking",
				"--disable-component-update", "--disable-sync", "--disable-extensions",
				"--user-data-dir="+t.TempDir(), "--window-size=1280,1600",
				"--virtual-time-budget=10000", "--dump-dom", ts.URL+p.url).Output()
			if err != nil {
				t.Fatalf("chrome: %v", err)
			}
			m := dataContrast.FindSubmatch(out)
			if m == nil {
				t.Fatalf("no result from the measuring script; it did not run on this page")
			}
			var res contrastResult
			if err := json.Unmarshal([]byte(html.UnescapeString(string(m[1]))), &res); err != nil {
				t.Fatalf("reading the result: %v", err)
			}
			if res.Checked == 0 {
				if p.is("text/html") {
					t.Fatalf("no text measured on %s; the check would pass over anything", p.url)
				}
				t.Logf("no text on %s, so nothing for 1.4.6 to measure", p.url)
				return
			}
			sort.Slice(res.Failures, func(i, j int) bool { return res.Failures[i].Ratio < res.Failures[j].Ratio })
			for i, f := range res.Failures {
				if i == 8 {
					t.Errorf("... and %d more", len(res.Failures)-i)
					break
				}
				t.Errorf("%.2f:1 below 7:1 — %q in <%s>, %s on %s", f.Ratio, f.Text, f.Element, f.Color, f.Background)
			}
			t.Logf("%d text runs measured, %d could not be brought into view", res.Checked, res.Unmeasured)
		})
	}
}

// instrument serves s with a test script added to every HTML and SVG page,
// and without the Content-Security-Policy that would refuse to run it. If
// logRequest is set it is told the URI of every request the browser makes,
// other than for the injected script itself.
func instrument(s *server, script string, logRequest func(string)) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/__contrast.js" {
			w.Header().Set("Content-Type", "text/javascript")
			fmt.Fprint(w, script)
			return
		}
		if logRequest != nil {
			logRequest(r.URL.RequestURI())
		}
		// Uncompressed, or the script would be spliced into gzip bytes.
		r.Header.Del("Accept-Encoding")
		rec := httptest.NewRecorder()
		s.ServeHTTP(rec, r)
		body := rec.Body.String()
		ctype := rec.Header().Get("Content-Type")
		switch {
		case strings.HasPrefix(ctype, "text/html"):
			body = injectBefore(body, "</body>", `<script src="/__contrast.js"></script>`)
		case strings.HasPrefix(ctype, "image/svg+xml"):
			body = injectBefore(body, "</svg>", `<script href="/__contrast.js"/>`)
		}
		for k, v := range rec.Header() {
			if k != "Content-Security-Policy" && k != "Content-Length" {
				w.Header()[k] = v
			}
		}
		w.WriteHeader(rec.Code)
		fmt.Fprint(w, body)
	})
}

// injectBefore inserts tag before the last occurrence of end, or appends it.
func injectBefore(body, end, tag string) string {
	if i := strings.LastIndex(body, end); i >= 0 {
		return body[:i] + tag + body[i:]
	}
	return body + tag
}
