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
	"io/fs"
	"math"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"os/exec"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"birdview/data"
	"birdview/docs"
	"birdview/internal/boundary"
	"birdview/internal/dataset"
	"birdview/internal/month"
	"birdview/internal/view"
)

// TestBoundaryFileIsServedByHash checks the region boundary file the home
// page's Here button loads: named by its hash, cached for a year, and nothing
// else served under its directory.
//
// Verifies: tech.regionBoundaries, tech.cacheLifetimeStated.
func TestBoundaryFileIsServedByHash(t *testing.T) {
	s := newTestServer(t)
	body := mustGet(t, s, "/?month=all")
	here := mustMatch(t, body, `(<button type="button" class="btn here"[^>]*>)`, "the Here button")
	bURL := mustMatch(t, here, `data-boundaries="([^"]+)"`, "the boundary file's URL")
	w := httptest.NewRecorder()
	s.ServeHTTP(w, httptest.NewRequest(http.MethodGet, bURL, nil))
	if w.Code != http.StatusOK || w.Body.String() != string(data.Boundaries) {
		t.Errorf("GET %s = %d, not the boundary file", bURL, w.Code)
	}
	if cc := w.Header().Get("Cache-Control"); cc != "public, max-age=31536000, immutable" {
		t.Errorf("the boundary file is served with Cache-Control %q", cc)
	}
	if code, _ := get(t, s, "/boundaries/0000000000000000.json"); code != http.StatusNotFound {
		t.Errorf("a wrong hash under /boundaries/ answered %d, want 404", code)
	}
}

//go:embed testdata/here_probe.js
var hereProbe string

var dataAttr = func(name string) *regexp.Regexp { return regexp.MustCompile(name + `="([^"]*)"`) }

// TestHereInTheBrowser presses Here in headless Chrome with the browser's
// location stubbed, and records every request the browser makes.
//
// The lookup must give the Go lookup's answer at every known point; a press
// must land on the URL choosing the region by hand gives; a refusal, a failure,
// and a point outside every region must each say so and go nowhere; and no
// request may carry the position, or show that Here was pressed.
//
// Verifies: product.hereSelectsLocalRegion, product.locationStaysInBrowser.
// Criteria: app.hereLookupAgreesInTheBrowser, app.hereSendsNothingExtra.
func TestHereInTheBrowser(t *testing.T) {
	chrome := findChrome()
	if chrome == "" {
		t.Skip("no Chrome or Chromium found (set BIRDVIEW_CHROME); Here was not tested in a browser")
	}
	s := newTestServer(t)
	// A named month, so the page does not depend on the clock.
	const start = "/?by=mass&month=May"
	pressedCA := string(nextToggle(url.Values{"by": {"mass"}, "month": {"May"}}, "region", "CA"))

	var points [][2]float64
	for _, p := range boundary.KnownPoints {
		points = append(points, [2]float64{p.Lat, p.Lon})
	}
	pts, _ := json.Marshal(points)

	cases := []struct {
		name, click string
		lands       string // the page the browser ends on
		status      string // what the page says, if it stays
	}{
		{name: "lookups"},
		{name: "in California", click: "38.5816,-121.4944", lands: pressedCA},
		{name: "refused", click: "denied", lands: start, status: "refused"},
		{name: "unavailable", click: "unavailable", lands: start, status: "could not be found"},
		{name: "outside", click: "51.5074,-0.1278", lands: start, status: "outside every region"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			var mu sync.Mutex
			var requests []string
			ts := httptest.NewServer(instrument(s, hereProbe, func(uri string) {
				mu.Lock()
				requests = append(requests, uri)
				mu.Unlock()
			}))
			defer ts.Close()

			frag := "#probe=1&points=" + url.QueryEscape(string(pts))
			if c.click != "" {
				frag += "&click=" + c.click
			}
			ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
			defer cancel()
			cmd := exec.CommandContext(ctx, chrome, "--headless", "--disable-gpu",
				"--no-first-run", "--no-default-browser-check", "--disable-background-networking",
				"--disable-component-update", "--disable-sync", "--disable-extensions",
				"--user-data-dir="+profileDir(t), "--virtual-time-budget=3000",
				"--dump-dom", ts.URL+start+frag)
			var dom string
			if c.lands == pressedCA {
				// A press that navigates leaves headless Chrome waiting forever
				// under these flags, which keep it off the network. What the
				// server received is the evidence that matters, so wait for
				// the page request and then stop the browser.
				if err := cmd.Start(); err != nil {
					t.Fatalf("chrome: %v", err)
				}
				deadline := time.Now().Add(20 * time.Second)
				for time.Now().Before(deadline) {
					mu.Lock()
					arrived := slices.Contains(requests, pressedCA)
					mu.Unlock()
					if arrived {
						// Long enough for anything else a press sends to arrive.
						time.Sleep(500 * time.Millisecond)
						break
					}
					time.Sleep(100 * time.Millisecond)
				}
				cmd.Process.Kill()
				cmd.Wait()
			} else {
				out, err := cmd.Output()
				if err != nil {
					t.Fatalf("chrome: %v", err)
				}
				dom = string(out)
			}
			mu.Lock()
			defer mu.Unlock()

			// Loaded with the page: before any page request after the first,
			// so before anything a press could have caused.
			loaded := -1
			for i, q := range requests {
				if q == s.boundaryURL {
					loaded = i
					break
				}
			}
			if loaded < 0 {
				t.Errorf("the boundary file was never requested: %v", requests)
			}
			for i, q := range requests[1:] {
				if i+1 < loaded && !strings.HasPrefix(q, "/static/") {
					t.Errorf("page request %s came before the boundary file: %v", q, requests)
				}
			}
			// Nothing the browser sent carries the position or mentions Here.
			for _, q := range requests {
				for _, leak := range []string{"38.58", "121.49", "51.50", "0.127", "here", "lat"} {
					if strings.Contains(strings.ToLower(q), leak) {
						t.Errorf("request %q contains %q", q, leak)
					}
				}
			}

			if m := dataAttr("data-url").FindStringSubmatch(dom); c.lands != "" && c.lands != pressedCA && (m == nil || html.UnescapeString(m[1]) != c.lands) {
				got := "nothing"
				if m != nil {
					got = html.UnescapeString(m[1])
				}
				t.Errorf("the browser ended on %s, want %s; requests %v", got, c.lands, requests)
			}
			if c.lands == pressedCA {
				// One page request beyond the first, and it is the hand-chosen URL.
				var pages []string
				for _, q := range requests {
					if !strings.HasPrefix(q, "/boundaries/") && !strings.HasPrefix(q, "/static/") {
						pages = append(pages, q)
					}
				}
				if len(pages) < 2 || pages[1] != pressedCA {
					t.Errorf("pressing Here requested pages %v; want %s next after %s", pages, pressedCA, start)
				}
				for _, p := range pages[2:] {
					if p != pressedCA {
						t.Errorf("pressing Here also requested %s", p)
					}
				}
				return
			}

			m := dataAttr("data-probe").FindStringSubmatch(dom)
			if m == nil {
				t.Fatalf("the probe recorded nothing; requests %v", requests)
			}
			var probe struct {
				Enabled bool     `json:"enabled"`
				Lookups []string `json:"lookups"`
				Status  string   `json:"status"`
				Error   string   `json:"error"`
			}
			if err := json.Unmarshal([]byte(html.UnescapeString(m[1])), &probe); err != nil {
				t.Fatal(err)
			}
			if probe.Error != "" {
				t.Fatalf("the boundary file did not load in the browser: %s", probe.Error)
			}
			if !probe.Enabled {
				t.Error("Here is still inactive with JavaScript running")
			}
			for i, p := range boundary.KnownPoints {
				if i >= len(probe.Lookups) || probe.Lookups[i] != p.Want {
					got := "nothing"
					if i < len(probe.Lookups) {
						got = fmt.Sprintf("%q", probe.Lookups[i])
					}
					t.Errorf("in the browser %s resolves to %s; Go and geography say %q", p.Place, got, p.Want)
				}
			}
			if c.status != "" && !strings.Contains(probe.Status, c.status) {
				t.Errorf("after pressing Here the page says %q, want it to say %q", probe.Status, c.status)
			}
		})
	}
}

// TestWelcomeLine checks the home page's welcome line: every link in it is
// also in the guide's "Some things to try", and the link to that list lands
// on its heading.
//
// Verifies: product.nextView.
func TestWelcomeLine(t *testing.T) {
	s := newTestServer(t)
	body := mustGet(t, s, "/")
	welcome := mustMatch(t, body, `(?s)<p class="welcome">(.*?)</p>`, "the welcome line")
	guide, err := docs.Files.ReadFile("using.md")
	if err != nil {
		t.Fatal(err)
	}
	g := string(guide)
	start := strings.Index(g, "## Some things to try")
	if start < 0 {
		t.Fatal("the guide has no Some things to try")
	}
	list := g[start:]
	if end := strings.Index(list[3:], "\n## "); end >= 0 {
		list = list[:end+3]
	}
	links := regexp.MustCompile(`<a href="([^"]+)">([^<]+)</a>`).FindAllStringSubmatch(welcome, -1)
	if len(links) < 3 {
		t.Fatalf("the welcome line has %d links", len(links))
	}
	for _, l := range links {
		href := html.UnescapeString(l[1])
		if href == "/docs/using#some-things-to-try" {
			if code, page := get(t, s, "/docs/using"); code != http.StatusOK || !strings.Contains(page, `id="some-things-to-try"`) {
				t.Errorf("the welcome's link to the full list lands on no heading")
			}
			continue
		}
		// The birds around you now is this address with #here-and-now; the
		// guide links the bare home page's.
		if strings.HasSuffix(href, "#here-and-now") {
			href = "/#here-and-now"
		}
		if !strings.Contains(list, "]("+href+")") {
			t.Errorf("the welcome links to %s, which the guide's Some things to try does not", href)
		}
	}
}

// TestHereAndNowLink follows the home page's #here-and-now link in headless
// Chrome, with the browser's location stood in for: found, it chooses the
// region and sets the month to Now, keeping the rest; refused, it stays and
// says so in the welcome line, naming the Region setting.
//
// Verifies: product.hereSelectsLocalRegion, product.locationStaysInBrowser.
// Criterion: app.hereLookupAgreesInTheBrowser.
func TestHereAndNowLink(t *testing.T) {
	chrome := findChrome()
	if chrome == "" {
		t.Skip("no Chrome or Chromium found (set BIRDVIEW_CHROME); the #here-and-now link was not tested")
	}
	s := newTestServer(t)
	frag := func(geo string) string {
		if strings.HasPrefix(geo, "follow=1&") {
			return "#probe=1&" + geo
		}
		return "#here-and-now&probe=1&geo=" + geo
	}
	// On the home page, the link chooses the region and sets the month to
	// Now, the browser's month, keeping every other setting.
	thisMonth := month.Abbrev[time.Now().Month()-1]
	inCA := "/?by=mass&month=" + thisMonth + "&region=CA"
	for _, c := range []struct{ name, start, geo, lands, says string }{
		{"home in California", "/?by=mass&month=May", "38.5816,-121.4944", inCA, ""},
		{"home refused", "/", "denied", "/", "refused, so the page is unchanged. Choose yours under Region."},
		// Clicked on the page, not arrived by: the link only adds the
		// fragment, so the page must answer its change.
		{"home clicked in California", "/?by=mass&month=May", "follow=1&geo=38.5816,-121.4944", inCA, ""},
		{"home clicked refused", "/?region=FL", "follow=1&geo=denied", "/?region=FL", "refused, so the page is unchanged. Choose yours under Region."},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			var mu sync.Mutex
			var requests []string
			ts := httptest.NewServer(instrument(s, hereProbe, func(uri string) {
				mu.Lock()
				requests = append(requests, uri)
				mu.Unlock()
			}))
			defer ts.Close()
			cmd := exec.Command(chrome, "--headless", "--disable-gpu", "--no-first-run",
				"--no-default-browser-check", "--disable-background-networking",
				"--disable-component-update", "--disable-sync", "--disable-extensions",
				"--user-data-dir="+t.TempDir(), "--virtual-time-budget=3000", "--dump-dom",
				ts.URL+c.start+frag(c.geo))
			var dom string
			if c.lands != c.start {
				// Navigating leaves headless Chrome waiting; read the server's log.
				if err := cmd.Start(); err != nil {
					t.Fatal(err)
				}
				for deadline := time.Now().Add(20 * time.Second); time.Now().Before(deadline); time.Sleep(100 * time.Millisecond) {
					mu.Lock()
					arrived := slices.Contains(requests, c.lands)
					mu.Unlock()
					if arrived {
						time.Sleep(500 * time.Millisecond)
						break
					}
				}
				cmd.Process.Kill()
				cmd.Wait()
			} else {
				out, err := cmd.Output()
				if err != nil {
					t.Fatalf("chrome: %v", err)
				}
				dom = string(out)
			}
			mu.Lock()
			defer mu.Unlock()
			for _, q := range requests {
				for _, leak := range []string{"38.58", "121.49", "here", "lat"} {
					if strings.Contains(strings.ToLower(q), leak) {
						t.Errorf("request %q contains %q", q, leak)
					}
				}
			}
			var pages []string
			for _, q := range requests {
				if !strings.HasPrefix(q, "/boundaries/") && !strings.HasPrefix(q, "/static/") {
					pages = append(pages, q)
				}
			}
			if c.lands != c.start {
				if len(pages) < 2 || pages[1] != c.lands {
					t.Errorf("arriving by #here-and-now requested pages %v; want %s next", pages, c.lands)
				}
				return
			}
			if len(pages) != 1 {
				t.Errorf("a refused #here-and-now still requested pages %v", pages)
			}
			m := dataAttr("data-probe").FindStringSubmatch(dom)
			if m == nil {
				t.Fatalf("the probe recorded nothing; requests %v", requests)
			}
			var probe struct {
				Welcome string `json:"welcome"`
			}
			json.Unmarshal([]byte(html.UnescapeString(m[1])), &probe)
			if !strings.Contains(probe.Welcome, c.says) {
				t.Errorf("the welcome line says %q, want it to say %q", probe.Welcome, c.says)
			}
		})
	}
}

// TestNowIsTheBrowsersMonth loads the home page in headless Chrome and reads
// Month's Now after here.js has run: it links to the browser's own month,
// with every other setting kept, rather than to month=now, which only a
// reader without the script follows (TestNextWithoutAMonthIsNotStored).
//
// Verifies: product.nextView. Criterion: app.nextViewDrawsOneSet.
func TestNowIsTheBrowsersMonth(t *testing.T) {
	chrome := findChrome()
	if chrome == "" {
		t.Skip("no Chrome or Chromium found (set BIRDVIEW_CHROME); Now was not tested in a browser")
	}
	ts := httptest.NewServer(newTestServer(t))
	defer ts.Close()
	out, err := exec.Command(chrome, "--headless", "--disable-gpu", "--no-first-run",
		"--no-default-browser-check", "--disable-background-networking",
		"--disable-component-update", "--disable-sync", "--disable-extensions",
		"--user-data-dir="+t.TempDir(), "--virtual-time-budget=3000", "--dump-dom",
		ts.URL+"/?by=mass&month=May&region=CA").Output()
	if err != nil {
		t.Fatalf("chrome: %v", err)
	}
	m := regexp.MustCompile(`<a [^>]*href="([^"]*)"[^>]*data-now[^>]*>Now</a>`).FindStringSubmatch(string(out))
	if m == nil {
		t.Fatal("the page has no Now link marked data-now")
	}
	if want := "/?by=mass&month=" + month.Abbrev[time.Now().Month()-1] + "&region=CA"; html.UnescapeString(m[1]) != want {
		t.Errorf("after here.js, Now links to %s; want %s", html.UnescapeString(m[1]), want)
	}
}

// TestGuidesReflow checks that a served guide flows as its paragraphs, not as
// its source lines. The renderer once turned every newline into <br>, so a
// guide wrapped at 100 columns in its source broke mid-sentence on the page,
// unlike the same file on a code host.
func TestGuidesReflow(t *testing.T) {
	s := newTestServer(t)
	for _, p := range []string{"/docs/using", "/docs/data", "/docs/building", "/docs/pipeline"} {
		_, page := get(t, s, p)
		if n := strings.Count(page, "<br"); n > 0 {
			t.Errorf("%s has %d line breaks; a wrapped source line should not end a line on the page", p, n)
		}
	}
}

// TestBillAndTailOrderings checks the two orderings by measured parts: the
// home page sorts ascending by bill or tail length and says so in Order by.
//
// Verifies: product.oneSelectedMetric, product.ascendingOrder.
// Criterion: acceptance.billAndTailOrderings.
func TestBillAndTailOrderings(t *testing.T) {
	s := newTestServer(t)
	for _, c := range []struct {
		by    string
		label string
		value func(*dataset.Species) float64
	}{
		{"bill", "Bill", func(sp *dataset.Species) float64 { return sp.Segments.Bill.Value }},
		{"tail", "Tail", func(sp *dataset.Species) float64 { return sp.Segments.Tail.Value }},
	} {
		body := mustGet(t, s, "/?by="+c.by+"&month=all&show=all")
		if !strings.Contains(body, "<summary><b>Order by</b> "+strings.ToLower(c.label)+" \u2191</summary>") {
			t.Errorf("Order by does not say %s, smallest first", c.label)
		}
		// Ascending by the part, read off the tiles in page order.
		var last float64
		for _, id := range regexp.MustCompile(`<a class="tile[^"]*" href="/species/([a-z0-9]+)`).FindAllStringSubmatch(body, -1) {
			v := c.value(s.byID[id[1]])
			if v < last {
				t.Errorf("by %s, %s (%.1f) follows a bird of %.1f", c.by, id[1], v, last)
			}
			last = v
		}
		if last == 0 {
			t.Fatalf("no tiles read by %s", c.by)
		}
	}
}

// TestTierHeadersNameTheOrdering checks every tier heading on the home page
// names the ordering it belongs to, as Order by labels it: "Tail Tier 1"
// under Tail.
//
// Verifies: product.sectionHeadersStateRange. Criterion: acceptance.tierHeadingsNameTheOrdering.
func TestTierHeadersNameTheOrdering(t *testing.T) {
	s := newTestServer(t)
	for by, label := range map[string]string{"": "Frequency", "length": "Length", "wingspan": "Wingspan",
		"mass": "Mass", "bill": "Bill", "tail": "Tail", "frequency": "Frequency"} {
		path := "/?month=all&show=all"
		if by != "" {
			path += "&by=" + by
		}
		tiers := regexp.MustCompile(`<span class="tier">([^<]+)</span>`).FindAllStringSubmatch(mustGet(t, s, path), -1)
		if len(tiers) == 0 {
			t.Fatalf("%s has no tier headings", path)
		}
		for i, m := range tiers {
			if want := fmt.Sprintf("%s Tier %d", label, i+1); m[1] != want {
				t.Errorf("%s heading %d reads %q, want %q", path, i+1, m[1], want)
			}
		}
	}
}

// TestHeadingDigits checks the precision chosen for a page's headings: the
// fewest extra decimals that keep every stated value apart, within a heading
// and between neighbors, so a tier of 5.0 to 6.0 mm bills followed by one
// starting at 6.05 mm never reads as sharing an end, though each tier's own
// ends already differ at two decimals of an inch; a page of wide tiers with
// wide gaps gets none.
//
// Verifies: product.sectionHeadersStateRange. Criterion: acceptance.tierHeadingDigits.
func TestHeadingDigits(t *testing.T) {
	u := unitsFor(false, 10)
	narrow := []tierSpan{{5.0, 6.0, u}, {6.05, 7.0, u}, {8.0, 9.0, u}}
	extra := headingDigits(narrow, view.ByBill)
	var stated []string
	for _, sp := range narrow {
		stated = append(stated, u.formatDigits(sp.lo, view.ByBill, extra), u.formatDigits(sp.hi, view.ByBill, extra))
	}
	for i := 1; i < len(stated); i++ {
		if stated[i] == stated[i-1] {
			t.Errorf("with %d extra decimals the headings read %v; %q appears twice running", extra, stated, stated[i])
		}
	}
	if wide := headingDigits([]tierSpan{{100, 112, u}, {140, 155, u}}, view.ByLength); wide != 0 {
		t.Errorf("wide tiers got %d extra decimals, want none", wide)
	}
}

// TestTierRangesHoldTheirBirds checks that a tier heading states only its own
// birds, smallest to largest, and that the page's headings read as a strictly
// rising sequence in one precision, so a gap between tiers shows as a gap.
// A heading once ran on to where the next tier began, which made four herons
// and an ibis look as if they spanned the five inches up to the pelicans.
//
// Verifies: product.sectionHeadersStateRange. Criterion: acceptance.tierHeadingsHoldTheirBirds.
func TestTierRangesHoldTheirBirds(t *testing.T) {
	s := newTestServer(t)
	heading := regexp.MustCompile(`<span class="tier">[^<]+</span>\s*([^<&]+?)(?: &ndash; ([^<]+?))?\s*<span class="n">`)
	toMM := func(v string) float64 {
		num, unit, _ := strings.Cut(strings.TrimSpace(v), " ")
		f, _ := strconv.ParseFloat(num, 64)
		return map[string]float64{"in": 25.4, "cm": 10}[unit] * f
	}
	decimals := func(v string) int {
		num, _, _ := strings.Cut(strings.TrimSpace(v), " ")
		_, frac, _ := strings.Cut(num, ".")
		return len(frac)
	}
	var drawn []*dataset.Species
	for i := range s.ds.Species {
		drawn = append(drawn, &s.ds.Species[i])
	}
	for _, c := range []struct {
		q string
		m view.Metric
	}{{"by=length", view.ByLength}, {"by=bill", view.ByBill}, {"by=tail", view.ByTail},
		{"by=wingspan", view.ByWingspan}, {"by=length&units=metric", view.ByLength}} {
		hs := heading.FindAllStringSubmatch(mustGet(t, s, "/?month=all&show=all&"+c.q), -1)
		secs := view.Sections(drawn, c.m, view.DefaultStepPercent)
		if len(hs) != len(secs) || len(hs) < 2 {
			t.Fatalf("%s: %d headings for %d tiers", c.q, len(hs), len(secs))
		}
		var prev float64
		for i, h := range hs {
			lo, hi := h[1], h[2]
			if hi == "" {
				hi = lo // one bird: one value
			}
			d := decimals(lo)
			tol := 25.4 * 0.5 * math.Pow(10, -float64(d)) // half the last digit, in inches at worst
			if math.Abs(toMM(lo)-secs[i].Min) > tol || math.Abs(toMM(hi)-secs[i].Max) > tol {
				t.Errorf("%s tier %d reads %s – %s; its birds run %.1f to %.1f mm", c.q, i+1, lo, hi, secs[i].Min, secs[i].Max)
			}
			if decimals(hi) != d || decimals(lo) != decimals(hs[0][1]) {
				t.Errorf("%s tier %d reads %s – %s; every heading should carry tier 1's decimals", c.q, i+1, lo, hi)
			}
			if i > 0 && toMM(lo) <= prev {
				t.Errorf("%s tier %d begins at %s, not above where tier %d ended", c.q, i+1, lo, i)
			}
			prev = toMM(hi)
		}
	}
}

// TestTierCountFollowsTheRange checks a tier heading's "N birds" sits right
// after the range in the count line's style, not pushed to the far edge in
// its own, where it read as part of something else or was not seen at all.
//
// Verifies: product.sectionHeadersStateRange. Criterion: acceptance.tierCountFollowsTheRange.
func TestTierCountFollowsTheRange(t *testing.T) {
	css, err := fs.ReadFile(assetFS, "static/styles.css")
	if err != nil {
		t.Fatal(err)
	}
	rule := mustMatch(t, string(css), `(?m)^\.tierhd \.n \{([^}]*)\}`, "the tier count's rule")
	if strings.Contains(rule, "margin-left: auto") {
		t.Errorf("the tier count is pushed to the far edge: %s", rule)
	}
	count := mustMatch(t, string(css), `(?m)^\.sub, \.count, footer \{([^}]*)\}`, "the count line's rule")
	for _, decl := range []string{"color: var(--dim)", "font-size: 12.5px"} {
		if !strings.Contains(count, decl) || !strings.Contains(rule, decl) {
			t.Errorf("the tier count and the count line differ on %q: %q vs %q", decl, rule, count)
		}
	}
	body := mustGet(t, newTestServer(t), "/?month=all&by=length")
	if !regexp.MustCompile(`&ndash; [^<]+</?[^>]*>?\s*<span class="n">\d+ birds?</span>`).MatchString(body) &&
		!regexp.MustCompile(`in\s*<span class="n">\d+ birds?</span>`).MatchString(body) {
		t.Error("no tier heading has its species count right after the range")
	}
}

// TestTierHistogram checks the home page's histogram: one bar per
// tier, in order, each as tall as its tier's count against the largest, each
// labeled with its tier and linking to it, under every ordering.
//
// Verifies: product.tierHistogram. Criterion: app.tierHistogramMatchesTiers.
func TestTierHistogram(t *testing.T) {
	s := newTestServer(t)
	bar := regexp.MustCompile(`<a href="#tier-(\d+)" style="height:(\d+)%" aria-label="([^"]+)"`)
	for _, q := range []string{"by=length", "by=bill", "by=mass", "by=frequency", "by=length&family=Ardeidae"} {
		body := mustGet(t, s, "/?month=all&show=all&"+q)
		bars := bar.FindAllStringSubmatch(body, -1)
		secs := regexp.MustCompile(`(?s)<h3 class="tierhd" id="tier-(\d+)">.*?<span class="n">(\d+) birds?</span>`).FindAllStringSubmatch(body, -1)
		if len(secs) == 0 {
			t.Fatalf("/?%s: no tier headings with ids", q)
		}
		// One tier needs no histogram, and the home page draws none.
		if len(secs) == 1 && len(bars) == 0 {
			continue
		}
		if len(bars) != len(secs) {
			t.Fatalf("/?%s: %d bars for %d tiers", q, len(bars), len(secs))
		}
		// A stretch with no birds is one empty bucket per whole 15% step it
		// spans, between the bars on either side, under the size orderings.
		if !strings.Contains(q, "frequency") {
			m := view.ByLength
			for _, k := range []view.Metric{view.ByBill, view.ByMass} {
				if strings.Contains(q, string(k)) {
					m = k
				}
			}
			var drawn []*dataset.Species
			for i := range s.ds.Species {
				if !strings.Contains(q, "family=") || s.ds.Species[i].Family == "Ardeidae" {
					drawn = append(drawn, &s.ds.Species[i])
				}
			}
			raw := view.Sections(drawn, m, view.DefaultStepPercent)
			hist := mustMatch(t, body, `(?s)<nav class="hist"[^>]*>(.*?)</nav>`, "the histogram")
			runs := regexp.MustCompile(`<a href="#tier-\d+"|<span class="gap"></span>`).FindAllString(hist, -1)
			var gaps []int
			n := 0
			for _, r := range runs {
				if strings.HasPrefix(r, "<span") {
					n++
				} else {
					gaps = append(gaps, n)
					n = 0
				}
			}
			for k := 1; k < len(raw) && k < len(gaps); k++ {
				step := 1 + view.DefaultStepPercent/100
				want := int(math.Floor(math.Log(raw[k].Min/(raw[k-1].Min*step)) / math.Log(step)))
				if want < 0 {
					want = 0
				}
				if gaps[k] != want {
					t.Errorf("/?%s: %d empty buckets before tier %d, want %d for the stretch from %.1f to %.1f",
						q, gaps[k], k+1, want, raw[k-1].Min*step, raw[k].Min)
				}
			}
		}
		largest := 0
		for _, sec := range secs {
			n, _ := strconv.Atoi(sec[2])
			largest = max(largest, n)
		}
		for i, b := range bars {
			n, _ := strconv.Atoi(secs[i][2])
			if b[1] != secs[i][1] || b[1] != strconv.Itoa(i+1) {
				t.Errorf("/?%s: bar %d links to tier %s; tiers are in order", q, i+1, b[1])
			}
			if h, _ := strconv.Atoi(b[2]); h != int(math.Round(100*float64(n)/float64(largest))) {
				t.Errorf("/?%s: bar %d is %d%% tall for %d of the largest tier's %d species", q, i+1, h, n, largest)
			}
			if label := html.UnescapeString(b[3]); !strings.Contains(label, fmt.Sprintf("Tier %d", i+1)) ||
				!strings.HasSuffix(label, fmt.Sprintf("%d species", n)) {
				t.Errorf("/?%s: bar %d is labeled %q", q, i+1, label)
			}
		}
	}
}

// profileDir is a Chrome profile directory for a browser the test kills.
// A killed browser's helper processes can go on writing the profile for a
// moment, which fails t.TempDir's cleanup, so this one retries the removal
// for a few seconds and gives up quietly.
func profileDir(t *testing.T) string {
	t.Helper()
	dir, err := os.MkdirTemp("", "birdview-chrome-")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		for range 50 {
			if os.RemoveAll(dir) == nil {
				return
			}
			time.Sleep(100 * time.Millisecond)
		}
	})
	return dir
}
