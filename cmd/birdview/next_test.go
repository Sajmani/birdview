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
	"html/template"
	"math"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os/exec"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"birdview/internal/dataset"
	"birdview/internal/month"
)

// fixNextClock sets /next's clock to a day in m for the rest of the test.
func fixNextClock(t *testing.T, m time.Month) {
	t.Helper()
	old := nextClock
	nextClock = func() time.Time { return time.Date(2026, m, 15, 12, 0, 0, 0, time.UTC) }
	t.Cleanup(func() { nextClock = old })
}

func nextAt(s *server, target string) nextData {
	return s.nextPage(httptest.NewRequest(http.MethodGet, target, nil))
}

func tileIDs(p nextData) []string {
	var out []string
	for _, t := range p.Tiles {
		out = append(out, t.S.ID)
	}
	return out
}

// TestNextLandingIsThisMonthByFrequency checks / with no filter: every
// bird, of every month, or the months the address names, most often
// recorded first.
//
// Verifies: product.nextView. Criterion: app.nextViewDrawsOneSet.
func TestNextLandingIsThisMonthByFrequency(t *testing.T) {
	s := newTestServerReal(t)
	fixNextClock(t, time.October)
	old := nextCap
	nextCap = 5000 // every bird drawn, so every admitted bird can be looked for
	t.Cleanup(func() { nextCap = old })
	for _, c := range []struct {
		target string
		months []int
	}{
		{"/", nil},
		{"/?month=Oct", []int{9}},
		{"/?month=May", []int{4}},
		{"/?month=May&month=Jun", []int{4, 5}},
		{"/?month=all", nil},
	} {
		opt := nextOptions(httptest.NewRequest(http.MethodGet, "/", nil))
		opt.HasMonths = c.months != nil
		opt.Months = map[int]bool{}
		for _, m := range c.months {
			opt.Months[m] = true
		}
		want, _ := s.filterSpecies(opt)
		p := nextAt(s, c.target)
		if p.Count != len(want) {
			t.Errorf("%s: %d birds, the month filter admits %d", c.target, p.Count, len(want))
		}
		got := map[string]bool{}
		for _, id := range tileIDs(p) {
			got[id] = true
		}
		for _, sp := range want {
			if !got[sp.ID] {
				t.Errorf("%s: %s is admitted but not shown", c.target, sp.ID)
				break
			}
		}
		prev := math.Inf(1)
		for _, tl := range p.Tiles {
			num, den := shareIn(tl.S, s.ds, opt)
			f := 0.0
			if den > 0 {
				f = float64(num) / float64(den)
			}
			if f > prev+1e-15 {
				t.Errorf("%s: %s (%.6f) follows a bird at %.6f; want most often recorded first", c.target, tl.S.ID, f, prev)
				break
			}
			prev = f
		}
	}
	if a, b := nextAt(s, "/?month=Oct").Count, nextAt(s, "/").Count; a >= b {
		t.Errorf("month=Oct shows %d birds and / %d; the month does not narrow", a, b)
	}
	if a, b := nextAt(s, "/").Count, nextAt(s, "/?month=all").Count; a != b {
		t.Errorf("/ shows %d birds and month=all %d; want every month by default", a, b)
	}
	for _, c := range nextAt(s, "/").Controls {
		if c.Label == "Show" {
			t.Error("the count line still has a Show setting")
		}
	}
	if c := nextAt(s, "/").Controls; len(c) == 0 || c[0].Label != "Order by" || !strings.HasPrefix(c[0].Value, "frequency") {
		t.Errorf("the count line starts with %+v; want Order by frequency", c[0])
	}
}

// TestNextDrawsOneScale checks that every bird / shows is drawn at the
// same pixels per millimeter, the longest bird's, whatever the order.
//
// Verifies: product.nextView. Criterion: app.nextViewDrawsOneSet.
func TestNextDrawsOneScale(t *testing.T) {
	s := newTestServerReal(t)
	fixNextClock(t, time.October)
	for _, target := range []string{"/", "/?by=frequency", "/?family=Ardeidae&family=Trochilidae", "/?month=all&by=mass"} {
		p := nextAt(s, target)
		if len(p.Tiles) < 2 {
			t.Fatalf("%s: %d birds; too few to compare", target, len(p.Tiles))
		}
		var longest *dataset.Species
		for _, tl := range p.Tiles {
			if longest == nil || tl.S.TotalLength.Value > longest.TotalLength.Value {
				longest = tl.S
			}
		}
		want := p.Tiles[0].Geom.PxPerMm
		for _, tl := range p.Tiles {
			if math.Abs(tl.Geom.PxPerMm-want) > 1e-9 {
				t.Errorf("%s: %s is drawn at %.5f px/mm, %s at %.5f", target, tl.S.ID, tl.Geom.PxPerMm, p.Tiles[0].S.ID, want)
				break
			}
		}
		for _, tl := range p.Tiles {
			if tl.S == longest && tl.Geom.TotalHeight() > tileBox.H+1e-6 {
				t.Errorf("%s: the longest bird, %s, overflows its tile", target, longest.ID)
			}
		}
	}
}

// TestNextOrders checks the order: sizes smallest first, frequency most
// often recorded first.
//
// Verifies: product.nextView. Criterion: app.nextViewDrawsOneSet.
func TestNextOrders(t *testing.T) {
	s := newTestServerReal(t)
	fixNextClock(t, time.October)
	p := nextAt(s, "/?by=mass&family=Parulidae")
	for i := 1; i < len(p.Tiles); i++ {
		if p.Tiles[i].S.Mass.Value < p.Tiles[i-1].S.Mass.Value {
			t.Errorf("by mass: %s after %s", p.Tiles[i].S.ID, p.Tiles[i-1].S.ID)
		}
	}
	opt := nextOptions(httptest.NewRequest(http.MethodGet, "/?month=Oct", nil))
	p = nextAt(s, "/?by=frequency&month=Oct")
	prev := math.Inf(1)
	for _, tl := range p.Tiles {
		num, den := shareIn(tl.S, s.ds, opt)
		f := float64(num) / float64(den)
		if f > prev {
			t.Errorf("by frequency: %s (%.5f) after a bird at %.5f", tl.S.ID, f, prev)
		}
		prev = f
	}
}

// TestNextFiltersCombineAndEmptyIsAll checks that filters narrow together,
// that a filter or a family asks for every bird of that kind, and that an
// emptied filter means all rather than none.
//
// Verifies: product.nextView. Criterion: app.nextViewDrawsOneSet.
func TestNextFiltersCombineAndEmptyIsAll(t *testing.T) {
	s := newTestServerReal(t)
	fixNextClock(t, time.October)
	herons := nextAt(s, "/?family=Ardeidae")
	if len(herons.Tiles) < 5 {
		t.Fatalf("family=Ardeidae shows %d birds, want every heron", len(herons.Tiles))
	}
	fl := nextAt(s, "/?family=Ardeidae&region=FL")
	me := nextAt(s, "/?family=Ardeidae&region=ME")
	if len(fl.Tiles) == 0 || len(me.Tiles) >= len(herons.Tiles) {
		t.Errorf("herons: %d everywhere, %d in Florida, %d in Maine; region does not narrow",
			len(herons.Tiles), len(fl.Tiles), len(me.Tiles))
	}
	for _, tl := range fl.Tiles {
		if tl.S.Family != "Ardeidae" {
			t.Errorf("family=Ardeidae&region=FL shows %s, of %s", tl.S.ID, tl.S.Family)
		}
	}
	if got, want := len(nextAt(s, "/?family=").Tiles), len(nextAt(s, "/").Tiles); got != want {
		t.Errorf("/?family= shows %d birds, / %d; an empty filter must mean all", got, want)
	}
	if got := len(nextAt(s, "/?region=&month=").Tiles); got == 0 {
		t.Error("/?region=&month= shows nothing; an empty filter must mean all")
	}
}

// follow requests target from s and returns where its redirect leads,
// fragment dropped, or target itself when it does not redirect.
func follow(t *testing.T, s *server, target string) string {
	t.Helper()
	w := httptest.NewRecorder()
	s.ServeHTTP(w, httptest.NewRequest(http.MethodGet, target, nil))
	if w.Code/100 != 3 {
		return target
	}
	loc, _, _ := strings.Cut(w.Header().Get("Location"), "#")
	return loc
}

func hasTile(p nextData, id string) bool {
	for _, tl := range p.Tiles {
		if tl.S.ID == id {
			return true
		}
	}
	return false
}

// at is nextAt for an address that may carry a fragment, which a browser
// keeps to itself.
func at(s *server, target string) nextData {
	target, _, _ = strings.Cut(target, "#")
	return nextAt(s, target)
}

func pinOf(p nextData, id string) (template.URL, bool, bool) {
	for _, tl := range p.Tiles {
		if tl.S.ID == id {
			return tl.PinURL, tl.Pinned, true
		}
	}
	return "", false, false
}

// TestNextPins checks the pins: every drawn bird has one; pinning a bird
// the settings do not choose draws it, pinned, with the settings kept and
// the find box emptied; pinning a bird already drawn marks it pinned; and
// unpinning either returns the address it came from. Nothing excludes a bird.
//
// Verifies: product.nextView. Criterion: app.nextViewDrawsOneSet.
func TestNextPins(t *testing.T) {
	s := newTestServerReal(t)
	fixNextClock(t, time.October)
	start := "/?family=Ardeidae"
	before := nextAt(s, start)
	for _, tl := range before.Tiles {
		if tl.PinURL == "" || tl.Pinned {
			t.Fatalf("%s: pin %q, pinned %t; every drawn bird has an unpinned pin", tl.S.ID, tl.PinURL, tl.Pinned)
		}
	}
	m := nextAt(s, start+"&find=sharp-shinned").Matches
	if len(m) != 1 || m[0].ID != "shshaw" || m[0].Pinned {
		t.Fatalf("find=sharp-shinned lists %+v, want Sharp-shinned Hawk unpinned", m)
	}
	pinned := follow(t, s, string(m[0].URL))
	if u, _ := url.Parse(pinned); u.Query().Get("find") != "sharp-shinned" || len(u.Query()["find"]) != 1 || u.Query().Get("family") != "Ardeidae" {
		t.Errorf("pinning went to %s: want the family kept and the find box's text kept once", pinned)
	}
	if again := nextAt(s, pinned); len(again.Matches) != 1 || !again.Matches[0].Pinned || string(again.CloseURL) != "/?family=Ardeidae&pin=shshaw" {
		t.Errorf("after pinning from the list, it lists %+v and its × goes to %s; want the hawk pinned, and × to the page without find",
			again.Matches, again.CloseURL)
	}
	withHawk := nextAt(s, pinned)
	unpin, on, drawn := pinOf(withHawk, "shshaw")
	if !drawn || !on || withHawk.Count != before.Count+1 {
		t.Fatalf("after pinning, %d birds (was %d), hawk drawn %t pinned %t", withHawk.Count, before.Count, drawn, on)
	}
	if back := string(unpin); back != start+"&find=sharp-shinned" {
		t.Errorf("unpinning the hawk goes to %s, want %s with the text kept", unpin, start+"&find=sharp-shinned")
	}
	if m := nextAt(s, pinned).Matches; len(m) != 1 || !m[0].Pinned {
		t.Errorf("a pinned bird found again is %+v; want it shown pinned", m)
	}

	// A bird already drawn is pinned in place, and unpinned back.
	heron := before.Tiles[0]
	p := at(s, string(heron.PinURL))
	unpin, on, drawn = pinOf(p, heron.S.ID)
	if !drawn || !on || p.Count != before.Count {
		t.Errorf("pinning %s: drawn %t, pinned %t, %d birds (was %d)", heron.S.ID, drawn, on, p.Count, before.Count)
	}
	if back, _, _ := strings.Cut(string(unpin), "#"); back != start {
		t.Errorf("unpinning %s goes to %s, want %s", heron.S.ID, unpin, start)
	}
	if strings.Contains(string(heron.PinURL), "exclude") {
		t.Errorf("a pin links to %s; nothing excludes a bird", heron.PinURL)
	}
}

// TestNextFindCountsThenLists checks the find box without JavaScript: it
// says how many birds match, and lists them only when there are
// nextFindLimit or fewer.
//
// Verifies: product.nextView. Criterion: app.nextViewDrawsOneSet.
func TestNextFindCountsThenLists(t *testing.T) {
	s := newTestServerReal(t)
	fixNextClock(t, time.October)
	for _, q := range []string{"hawk", "sharp", "warbler", "zzzz", "d hawk"} {
		want := 0
		for i := range s.ds.Species {
			if searchMatches(s.search[s.ds.Species[i].ID], q) {
				want++
			}
		}
		p := nextAt(s, "/?find="+url.QueryEscape(q))
		if p.Found != want {
			t.Errorf("find=%q: %d found, the matcher finds %d", q, p.Found, want)
		}
		if listed := len(p.Matches); (want <= nextFindLimit && listed != want) || (want > nextFindLimit && listed != 0) {
			t.Errorf("find=%q: %d listed of %d found; list all at %d or fewer, none above", q, listed, want, nextFindLimit)
		}
	}
}

//go:embed testdata/next_find_probe.js
var nextFindProbe string

// TestNextFindInTheBrowser types into /next's find box in headless Chrome
// and checks that, for each query, the script says what the server's page
// says for the same find=, and lists the same birds with the same links.
//
// Verifies: product.nextView. Criterion: app.nextViewDrawsOneSet.
func TestNextFindInTheBrowser(t *testing.T) {
	chrome := findChrome()
	if chrome == "" {
		t.Skip("no Chrome or Chromium found (set BIRDVIEW_CHROME); the find box was not tested in a browser")
	}
	s := newTestServer(t)
	old := nextFindLimit
	nextFindLimit = 2 // the fixture is small; "a" must be over the limit
	t.Cleanup(func() { nextFindLimit = old })
	qs := []string{"crow", "AMECRO", "a", "hummingbird", "zzzz", "Corvus brachy"}
	js, _ := json.Marshal(qs)
	ts := httptest.NewServer(instrument(s, nextFindProbe, nil))
	defer ts.Close()
	start := "/?month=all&pin=annhum"
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	out, err := exec.CommandContext(ctx, chrome, "--headless", "--disable-gpu",
		"--no-first-run", "--no-default-browser-check", "--disable-background-networking",
		"--disable-component-update", "--disable-sync", "--disable-extensions",
		"--user-data-dir="+t.TempDir(), "--virtual-time-budget=5000",
		"--dump-dom", ts.URL+start+"#queries="+url.QueryEscape(string(js))).Output()
	if err != nil {
		t.Fatalf("chrome: %v", err)
	}
	m := dataAttr("data-probe").FindSubmatch(out)
	if m == nil {
		t.Fatal("no result from the probe; it did not run")
	}
	var res struct {
		Typed []struct {
			Q     string   `json:"q"`
			Found string   `json:"found"`
			Close string   `json:"close"`
			Items []string `json:"items"`
			Texts []string `json:"texts"`
		} `json:"typed"`
		TilePin string `json:"tilePin"`
	}
	if err := json.Unmarshal([]byte(html.UnescapeString(string(m[1]))), &res); err != nil {
		t.Fatalf("reading the probe: %v", err)
	}
	if len(res.Typed) != len(qs) {
		t.Fatalf("the probe typed %d of %d queries", len(res.Typed), len(qs))
	}
	lists := 0
	for _, r := range res.Typed {
		w := httptest.NewRecorder()
		s.ServeHTTP(w, httptest.NewRequest(http.MethodGet, start+"&find="+url.QueryEscape(r.Q), nil))
		body := w.Body.String()
		fm := regexp.MustCompile(`<p class="found">([^<]*(?:<[^>]*>[^<]*)*?)</p>`).FindStringSubmatch(body)
		if fm == nil {
			t.Fatalf("the server's page for find=%q has no count", r.Q)
		}
		closer := regexp.MustCompile(`^<a class="findclose" href="([^"]*)"[^>]*>&times;</a>`)
		cm := closer.FindStringSubmatch(fm[1])
		if cm == nil {
			t.Fatalf("the server's list for find=%q has no × first", r.Q)
		}
		if want := html.UnescapeString(closer.ReplaceAllString(fm[1], "")); r.Found != want {
			t.Errorf("typing %q: the script says %q, the server %q", r.Q, r.Found, want)
		}
		if want := html.UnescapeString(cm[1]); r.Close != want || want != start {
			t.Errorf("typing %q: the script's × goes to %q, the server's to %q; want %s", r.Q, r.Close, want, start)
		}
		var want []string
		finds := body[strings.Index(body, `<div class="finds"`):]
		finds = finds[:strings.Index(finds, "</div>")]
		for _, l := range regexp.MustCompile(`<li><a class="(pin(?: on)?)" href="([^"]*)"`).FindAllStringSubmatch(finds, -1) {
			want = append(want, l[1]+" "+html.UnescapeString(l[2]))
		}
		if !slices.Equal(r.Items, want) {
			t.Errorf("typing %q: the script lists %v, the server %v", r.Q, r.Items, want)
		}
		var texts []string
		for _, li := range regexp.MustCompile(`(?s)<li>(.*?)</li>`).FindAllStringSubmatch(finds, -1) {
			texts = append(texts, strings.Join(strings.Fields(html.UnescapeString(regexp.MustCompile(`<[^>]+>`).ReplaceAllString(li[1], ""))), " "))
		}
		if !slices.Equal(r.Texts, texts) {
			t.Errorf("typing %q: the script's list reads %q, the server's %q", r.Q, r.Texts, texts)
		}
		if len(want) > 0 {
			lists++
		}
	}
	if !strings.Contains(res.Typed[2].Found, "Type more") {
		t.Errorf("typing \"a\" found %q; the check needs a query over the limit", res.Typed[2].Found)
	}
	if lists < 3 {
		t.Errorf("only %d queries listed birds; the check proves little", lists)
	}
	if u, err := url.Parse(res.TilePin); err != nil || u.Query().Get("find") != "crow" || u.Query().Get("month") != "all" {
		t.Errorf("a tile's pin clicked with \"crow\" in the box goes to %q; want the text kept with the settings", res.TilePin)
	}
}

// TestNextToggle checks the count line's choosing rule: from all, choosing
// one selects only it; choosing more adds them; choosing a selected one
// removes it; removing the last returns to all.
//
// Verifies: product.nextView. Criterion: app.nextViewDrawsOneSet.
func TestNextToggle(t *testing.T) {
	q := func(s string) url.Values { v, _ := url.ParseQuery(s); return v }
	for _, c := range []struct{ from, v, want string }{
		{"", "FL", "/?region=FL"},
		{"region=FL", "ME", "/?region=FL&region=ME"},
		{"region=FL&region=ME", "FL", "/?region=ME"},
		{"region=ME", "ME", "/"},
		{"region=", "ME", "/?region=ME"},
		{"region=FL&find=hawk&show=50&by=mass", "FL", "/?by=mass&find=hawk&show=50"},
		{"pin=blujay&show=pinned", "FL", "/?pin=blujay&region=FL&show=pinned"},
	} {
		if got := string(nextToggle(q(c.from), "region", c.v)); got != c.want {
			t.Errorf("toggling %s from %q = %s, want %s", c.v, c.from, got, c.want)
		}
	}
}

// TestNextShowAll checks the Show setting, the one show= parameter: by
// default the first nextCap of the order and a link that shows them all
// (show=all), in the same order; a set within the cap draws every bird with
// no link; show=<n> draws n and is kept across a filter but not Reset; the
// Show setting's First choice is a box in a form that sends every other
// setting, beside All with the count the settings match and Pinned only
// with the count pinned, the chosen one marked; an address naming the
// default, a blank, or nonsense redirects to one without show; and the count
// line is two rows, what is drawn with Show, Clear pins, and Reset, then the
// settings.
//
// Verifies: product.nextView. Criterion: app.nextViewDrawsOneSet.
func TestNextShowAll(t *testing.T) {
	s := newTestServerReal(t)
	fixNextClock(t, time.October)
	p := nextAt(s, "/")
	if p.Count <= nextCap {
		t.Fatalf("/ has %d birds; the check needs more than %d", p.Count, nextCap)
	}
	if len(p.Tiles) != nextCap || string(p.MoreURL) != "/?show=all" || p.Hidden != p.Count-nextCap ||
		p.Show != "first" || p.Cap != "100" {
		t.Errorf("/ draws %d of %d, hides %d, show %q %q, more link %q",
			len(p.Tiles), p.Count, p.Hidden, p.Show, p.Cap, p.MoreURL)
	}
	all := nextAt(s, string(p.MoreURL))
	if len(all.Tiles) != p.Count || all.MoreURL != "" || all.Show != "all" {
		t.Errorf("%s draws %d of %d, show %q", p.MoreURL, len(all.Tiles), p.Count, all.Show)
	}
	for i := range p.Tiles {
		if p.Tiles[i].S != all.Tiles[i].S {
			t.Errorf("Show all reorders: %s at %d, was %s", all.Tiles[i].S.ID, i, p.Tiles[i].S.ID)
			break
		}
	}

	// show=50: fifty drawn, the box says so, kept by a filter, not Reset.
	c := nextAt(s, "/?by=mass&show=50")
	if len(c.Tiles) != 50 || c.First != 50 || c.Hidden != c.Count-50 || c.Cap != "50" ||
		string(c.MoreURL) != "/?by=mass&show=all" {
		t.Errorf("show=50: %d drawn, first %d, hidden %d of %d, box %q, more %s",
			len(c.Tiles), c.First, c.Hidden, c.Count, c.Cap, c.MoreURL)
	}
	if got := string(nextToggle(url.Values{"show": {"all"}, "by": {"mass"}}, "region", "FL")); got != "/?by=mass&region=FL&show=all" {
		t.Errorf("a filter from show=all goes to %s; how many are shown is kept", got)
	}
	if string(c.ShowPinnedURL) != "/?by=mass&show=pinned" || string(c.ShowAllURL) != "/?by=mass&show=all" || string(c.ResetURL) != "/" {
		t.Errorf("show=50: Pinned only %s, All %s, Reset %s", c.ShowPinnedURL, c.ShowAllURL, c.ResetURL)
	}

	// The Show choices, in view under the heading: First [box], All (N),
	// Pinned (k), and Clear pins; First is a form with the other settings, the
	// chosen one marked.
	_, body := get(t, s, "/?by=mass&show=50&find=jay&pin=amecro&pin=annhum")
	row := regexp.MustCompile(`(?s)<div class="showrow">(.*?)\n</div>`).FindStringSubmatch(body)
	if row == nil {
		t.Fatal("the page has no Show row")
	}
	form := regexp.MustCompile(`(?s)<form class="showfirst btn on" action="/" method="get">(.*?)</form>`).FindStringSubmatch(row[1])
	if form == nil {
		t.Fatal("the Show row has no First form, marked chosen")
	}
	var sent []string
	for _, m := range regexp.MustCompile(`<input type="hidden" name="([^"]+)" value="([^"]*)">`).FindAllStringSubmatch(form[1], -1) {
		sent = append(sent, m[1]+"="+m[2])
	}
	if !slices.Equal(sent, []string{"by=mass", "find=jay", "pin=amecro", "pin=annhum"}) {
		t.Errorf("the First form sends %v; want every parameter but show", sent)
	}
	if !strings.Contains(form[1], ">First</button>\n    <input class=\"cap\" type=\"number\" name=\"show\"") || !strings.Contains(form[1], `value="50"`) {
		t.Errorf("the First form does not read First [50]:\n%s", form[1])
	}
	pp := nextAt(s, "/?by=mass&show=50&find=jay&pin=amecro&pin=annhum")
	allBtn, pinBtn := ">All ("+strconv.Itoa(pp.Matching)+")</a>", ">Pinned (2)</a>"
	if f, i, j, k := strings.Index(row[1], "<form"), strings.Index(row[1], allBtn), strings.Index(row[1], pinBtn), strings.Index(row[1], ">Clear pins</a>"); f < 0 || i < f || j < i || k < j {
		t.Errorf("the Show row does not read First, All (%d), Pinned (2), Clear pins, in that order:\n%s", pp.Matching, row[1])
	}
	if strings.Contains(body, `data-ctl="Show"`) || strings.Contains(body, `class="btn compare"`) {
		t.Error("Show is still a closed setting, or the pinned list still has its own Compare")
	}
	// The order of the page, top to bottom.
	var at []int
	marks := []string{"</header>", `<h2 class="settitle">`, `<div class="showrow">`, ">Clear pins</a>", `<div class="pinned"`,
		`<form class="search"`, `<div class="count">`, `data-ctl="Order by"`, `data-ctl="Names"`, ">Reset</a>",
		`<nav class="hist"`, `<span class="shown">Mass distribution of the first 50 of`, `<div class="tiles">`}
	for _, m := range marks {
		at = append(at, strings.Index(body, m))
	}
	for i := range marks {
		if at[i] < 0 || i > 0 && at[i] < at[i-1] {
			t.Fatalf("the page does not run in order: %q at %d, after %q at %d", marks[i], at[i], marks[max(i-1, 0)], at[max(i-1, 0)])
		}
	}
	if n := strings.Count(body, `<div class="countrow">`); n != 1 {
		t.Errorf("the count line has %d rows, want 1", n)
	}
	if strings.Contains(body, ">Show all</a>\n") && strings.Index(body, ">Show all</a>") < strings.Index(body, `<div class="tiles">`) {
		t.Error("the count line still has its own Show all; it is a Show choice now")
	}
	// The sentence under the histogram leads with the chosen order.
	for target, want := range map[string]string{
		"/?show=all": "Frequency distribution of all %d birds",
		"/?by=length&show=pinned&pin=amecro&pin=annhum": "Length distribution of %d pinned birds",
		"/?by=tail&rev=1": "Tail length distribution of the first 100 of %d birds",
	} {
		p := nextAt(s, target)
		if w := fmt.Sprintf(want, p.Count); !strings.Contains(mustGet(t, s, target), `<span class="shown">`+w+`</span>`) {
			t.Errorf("%s does not say %q", target, w)
		}
	}

	for from, to := range map[string]string{
		"/?show=100&by=mass": "/?by=mass", "/?by=mass&show=": "/?by=mass",
		"/?show=zero": "/", "/?show=0": "/", "/?show=-3": "/",
	} {
		w := httptest.NewRecorder()
		s.ServeHTTP(w, httptest.NewRequest(http.MethodGet, from, nil))
		if w.Code != http.StatusSeeOther || w.Header().Get("Location") != to {
			t.Errorf("GET %s = %d to %q, want 303 to %s", from, w.Code, w.Header().Get("Location"), to)
		}
	}
	for _, ok := range []string{"/?show=50", "/?show=all", "/?show=pinned"} {
		w := httptest.NewRecorder()
		s.ServeHTTP(w, httptest.NewRequest(http.MethodGet, ok, nil))
		if w.Code != http.StatusOK {
			t.Errorf("GET %s = %d, want 200", ok, w.Code)
		}
	}
	if p := nextAt(s, "/?family=Ardeidae"); p.MoreURL != "" || len(p.Tiles) != p.Count {
		t.Errorf("the herons, %d birds, draw %d and offer Show all %q", p.Count, len(p.Tiles), p.MoreURL)
	}
}

// TestNextOrderReverses checks the Order by setting: the chosen order links
// to its reverse and back, the others link to themselves in their usual
// direction, and a reversed order runs the other way.
//
// Verifies: product.nextView. Criterion: app.nextViewDrawsOneSet.
func TestNextOrderReverses(t *testing.T) {
	s := newTestServerReal(t)
	fixNextClock(t, time.October)
	choices := func(target string) map[string]nextChoice {
		out := map[string]nextChoice{}
		for _, c := range nextAt(s, target).Controls[0].Groups[0].Choices {
			out[strings.Fields(c.Label)[0]] = c
		}
		return out
	}
	c := choices("/?by=mass&family=Anatidae")
	if !c["Mass"].On || string(c["Mass"].URL) != "/?by=mass&family=Anatidae&rev=1" {
		t.Errorf("chosen Mass links to %s, want its reverse", c["Mass"].URL)
	}
	if string(c["Length"].URL) != "/?by=length&family=Anatidae" {
		t.Errorf("Length links to %s", c["Length"].URL)
	}
	c = choices("/?by=mass&family=Anatidae&rev=1")
	if string(c["Mass"].URL) != "/?by=mass&family=Anatidae" || string(c["Length"].URL) != "/?by=length&family=Anatidae" {
		t.Errorf("reversed: Mass links to %s and Length to %s; want back, and Length unreversed", c["Mass"].URL, c["Length"].URL)
	}
	up, down := nextAt(s, "/?by=mass&family=Anatidae"), nextAt(s, "/?by=mass&family=Anatidae&rev=1")
	if len(up.Tiles) < 5 || len(up.Tiles) != len(down.Tiles) {
		t.Fatalf("%d and %d ducks", len(up.Tiles), len(down.Tiles))
	}
	for i := 1; i < len(down.Tiles); i++ {
		if down.Tiles[i].S.Mass.Value > down.Tiles[i-1].S.Mass.Value {
			t.Errorf("reversed mass: %s after %s", down.Tiles[i].S.ID, down.Tiles[i-1].S.ID)
		}
	}
	if up.Tiles[0].S == down.Tiles[0].S {
		t.Errorf("%s leads both directions", up.Tiles[0].S.ID)
	}
	fwd, rev := nextAt(s, "/?family=Anatidae"), nextAt(s, "/?family=Anatidae&rev=1")
	if fwd.Tiles[0].S != rev.Tiles[len(rev.Tiles)-1].S {
		t.Errorf("the most recorded duck, %s, is not at the end of the reversed order", fwd.Tiles[0].S.ID)
	}
	if v := nextAt(s, "/?rev=1").Controls[0].Value; v != "frequency \u2191" {
		t.Errorf("reversed frequency reads %q", v)
	}
}

// TestNextWithoutAMonthIsNotStored checks that every home page, the default
// for every month included, is cached as any page is, and that month=now,
// Now without JavaScript, redirects to the clock's month with every other
// setting kept, and is neither stored nor revalidated.
//
// Verifies: product.nextView. Criterion: app.nextViewDrawsOneSet.
func TestNextWithoutAMonthIsNotStored(t *testing.T) {
	s := newTestServer(t)
	s.build = "test-build"
	fixNextClock(t, time.October)
	for _, target := range []string{"/", "/?region=FL", "/?month=", "/?month=Oct", "/?month=Oct&region=FL"} {
		w := httptest.NewRecorder()
		s.ServeHTTP(w, httptest.NewRequest(http.MethodGet, target, nil))
		if w.Code != http.StatusOK {
			t.Errorf("GET %s = %d", target, w.Code)
		}
		if cc, etag := w.Header().Get("Cache-Control"), w.Header().Get("ETag"); cc == "no-store" || etag == "" {
			t.Errorf("%s: Cache-Control %q, ETag %q; want it cacheable", target, cc, etag)
		}
	}
	for target, want := range map[string]string{
		"/?month=now":                      "/?month=Oct",
		"/?by=mass&month=now&region=FL":    "/?by=mass&month=Oct&region=FL",
		"/?month=May&month=now&pin=amecro": "/?month=Oct&pin=amecro",
	} {
		w := httptest.NewRecorder()
		s.ServeHTTP(w, httptest.NewRequest(http.MethodGet, target, nil))
		if w.Code != http.StatusSeeOther || w.Header().Get("Location") != want {
			t.Errorf("GET %s = %d to %q; want 303 to %s", target, w.Code, w.Header().Get("Location"), want)
		}
		if cc, etag := w.Header().Get("Cache-Control"), w.Header().Get("ETag"); cc != "no-store" || etag != "" {
			t.Errorf("%s: Cache-Control %q, ETag %q; want no-store and no validator", target, cc, etag)
		}
	}
}

// TestNextControlsOpenOneAtATime checks that every setting on the count line
// is a <details> sharing one name, which is what makes the browser close the
// others when one opens, and that Reset is there and goes to the landing set.
//
// Verifies: product.nextView. Criterion: app.nextViewDrawsOneSet.
func TestNextControlsOpenOneAtATime(t *testing.T) {
	s := newTestServer(t)
	_, body := get(t, s, "/?region=FL")
	n := strings.Count(body, `<details class="ctl"`)
	named := len(regexp.MustCompile(`<details class="ctl" name="ctl"[ >]`).FindAllString(body, -1))
	if n < 8 || named != n {
		t.Errorf("%d of %d count-line settings share the name ctl; want all, and at least 8", named, n)
	}
	if !strings.Contains(body, `<a class="reset" href="/"`) || !strings.Contains(body, ">Reset</a>") {
		t.Error("no Reset control linking to /")
	}
}

// TestNextAddedPastTheFirstIsDrawn checks that a bird pinned from past the
// first nextCap is drawn in its place in the order, with gap rows where
// unpinned birds are left out before, between, and after pinned birds, and
// none once Show all is chosen.
//
// Verifies: product.nextView. Criterion: app.nextViewDrawsOneSet.
func TestNextAddedPastTheFirstIsDrawn(t *testing.T) {
	s := newTestServerReal(t)
	fixNextClock(t, time.October)
	before := nextAt(s, "/")
	all := nextAt(s, "/?show=all")
	last := all.Tiles[len(all.Tiles)-1].S // the least recorded: far past the first
	if hasTile(before, last.ID) {
		t.Fatalf("%s is drawn already; the check needs a bird past the first %d", last.ID, nextCap)
	}
	m := nextAt(s, "/?find="+url.QueryEscape(last.ScientificName)).Matches
	if len(m) != 1 || m[0].Pinned {
		t.Fatalf("finding %s, a bird not drawn, lists %+v; want it unpinned", last.ScientificName, m)
	}
	after := nextAt(s, follow(t, s, string(m[0].URL)))
	if !hasTile(after, last.ID) {
		t.Errorf("after adding %s from past the first %d, it is not drawn", last.ID, nextCap)
	}
	// It was the very last of the hidden birds, so one fewer is hidden, all
	// of them sit in one gap before it, and none sit after it.
	if after.AddedPast != 1 || after.First != nextCap || after.Hidden != before.Hidden-1 {
		t.Errorf("after adding: first %d, added past it %d, hidden %d (was %d)", after.First, after.AddedPast, after.Hidden, before.Hidden)
	}
	if _, body := get(t, s, follow(t, s, string(m[0].URL))); !strings.Contains(body, " of "+strconv.Itoa(after.Count)+" birds, plus 1 pin</span>") {
		t.Errorf("the count line does not say \"plus 1 pin\" after %d birds", after.Count)
	}
	if _, body := get(t, s, "/?show=2&month=all&pin="+last.ID+"&pin="+all.Tiles[len(all.Tiles)-2].S.ID); !strings.Contains(body, " birds, plus 2 pins</span>") {
		t.Error("with two pins past the cap, the count line does not say \"plus 2 pins\"")
	}
	if len(after.Tiles) != len(before.Tiles)+1 {
		t.Errorf("after adding one bird, %d drawn (was %d)", len(after.Tiles), len(before.Tiles))
	}
	// The items end with one gap holding every hidden bird, then (its own
	// tier's heading, when it starts a tier, and) its tile.
	n := len(after.Items)
	gapAt := n - 2
	if after.Items[gapAt].Tier != nil {
		gapAt--
	}
	if after.Items[n-1].Tile == nil || after.Items[gapAt].Hidden != after.Hidden {
		t.Errorf("pinning the last bird leaves items ending %+v; want a gap of %d then its tile",
			after.Items[n-3:], after.Hidden)
	}

	// Two birds mid-order by length: gaps sit before the shorter, between the
	// two, and after the longer, summing to Hidden, each linking to Show all.
	two := "/?by=length&pin=fuwduc&pin=amecro"
	p := nextAt(s, two)
	_, body := get(t, s, two)
	tokens := regexp.MustCompile(`id="(s-[a-z0-9]+)"|<p class="gap">([0-9]+) birds? hidden — <a href="([^"]+)">Show all</a></p>`).
		FindAllStringSubmatch(body, -1)
	if len(tokens) != nextCap+5 {
		t.Fatalf("%s has %d tiles and gaps, want %d tiles and 3 gaps", two, len(tokens), nextCap+2)
	}
	tail := tokens[nextCap:]
	if tail[0][2] == "" || tail[1][1] != "s-amecro" || tail[2][2] == "" || tail[3][1] != "s-fuwduc" || tail[4][2] == "" {
		t.Fatalf("after the first %d, got %v; want gap, amecro, gap, fuwduc, gap", nextCap, tail)
	}
	sum := 0
	for _, g := range []int{0, 2, 4} {
		n, _ := strconv.Atoi(tail[g][2])
		if n <= 0 || html.UnescapeString(tail[g][3]) != string(p.MoreURL) {
			t.Errorf("gap %d: %q birds, link %q (want %s)", g, tail[g][2], tail[g][3], p.MoreURL)
		}
		sum += n
	}
	if sum != p.Hidden {
		t.Errorf("the three gaps hide %d birds, Hidden says %d", sum, p.Hidden)
	}
	if _, allBody := get(t, s, string(p.MoreURL)); strings.Contains(allBody, `class="gap"`) {
		t.Error("Show all still draws a gap row")
	}

	// A gap of one bird reads "1 bird hidden", and a gap after the last pin
	// reads "birds hidden".
	small := newTestServer(t)
	_, sbody := get(t, small, "/?month=all&by=mass&pin=amecro&show=2")
	if !strings.Contains(sbody, `<p class="gap">1 bird hidden — `) || !strings.Contains(sbody, `<p class="gap">2 birds hidden — `) {
		t.Errorf("with 1 hidden before amecro and 2 after, the gaps do not say so:\n%s", sbody)
	}
}

// TestNextShowPinned checks show=pinned: it empties the set the settings
// choose, so only pinned birds are drawn; pinning and unpinning then work on
// those alone; Show all matching birds brings the settings' birds back with
// the pinned ones; show=all is the default; and choosing a filter leaves the
// pinned-only view.
//
// Verifies: product.nextView. Criterion: app.nextViewDrawsOneSet.
func TestNextShowPinned(t *testing.T) {
	s := newTestServerReal(t)
	fixNextClock(t, time.October)
	start := nextAt(s, "/?by=mass&show=all")
	if string(start.ShowPinnedURL) != "/?by=mass&show=pinned" {
		t.Errorf("Pinned only links to %s, want the order kept", start.ShowPinnedURL)
	}
	if a, b := nextAt(s, "/?by=mass&show=all").Count, nextAt(s, "/?by=mass").Count; a != b {
		t.Errorf("show=all draws from %d birds, the default from %d; they match the same birds", a, b)
	}
	empty := nextAt(s, string(start.ShowPinnedURL))
	if empty.Matching != start.Matching || empty.Matching < 500 {
		t.Errorf("under Pinned only, All offers %d birds; the settings match %d", empty.Matching, start.Matching)
	}
	if !empty.PinnedOnly || len(empty.Tiles) != 0 || empty.Count != 0 || empty.Title != "Showing pinned birds" {
		t.Errorf("pinned only: %d drawn of %d, title %q", len(empty.Tiles), empty.Count, empty.Title)
	}
	two := "/?by=mass&pin=amecro&pin=annhum&show=pinned"
	p := nextAt(s, two)
	if got := tileIDs(p); !slices.Equal(got, []string{"annhum", "amecro"}) {
		t.Errorf("pinned only, with two pinned, draws %v, want annhum and amecro by mass", got)
	}
	if back := follow(t, s, two+"&bird=amecro"); back != "/?by=mass&pin=annhum&show=pinned" {
		t.Errorf("removing one of two goes to %s", back)
	}
	undo := nextAt(s, string(p.ShowFirstURL))
	if undo.PinnedOnly || !hasTile(undo, "amecro") || !hasTile(undo, "annhum") || undo.Count < 500 {
		t.Errorf("Show all matching birds (%s): pinned only %t, %d birds", p.ShowFirstURL, undo.PinnedOnly, undo.Count)
	}
}

// TestNextHereAndNow checks the Region setting's Here and the Month
// setting's Now. Here's address, filled with a region, is the page with only
// that region chosen and every other parameter kept, Pinned only included;
// the page carries here.js and the boundary file to find the region with.
// Now is the second month choice, after All, which is chosen when no month
// is named; Now links to month=now and is never itself chosen.
//
// Verifies: product.nextView. Criterion: app.nextViewDrawsOneSet.
func TestNextHereAndNow(t *testing.T) {
	s := newTestServer(t)
	fixNextClock(t, time.October)
	p := nextAt(s, "/?by=mass&region=CA&show=pinned&month=May")
	var region, mon *nextControl
	for i := range p.Controls {
		switch p.Controls[i].Label {
		case "Region":
			region = &p.Controls[i]
		case "Month":
			mon = &p.Controls[i]
		}
	}
	if region == nil || mon == nil {
		t.Fatal("no Region or Month setting")
	}
	if got := strings.Replace(region.HereURL, hereSlot, "FL", 1); got != "/?by=mass&month=May&region=FL&show=pinned" {
		t.Errorf("Here, filled with FL, is %s", got)
	}
	if region.Boundaries != s.boundaryURL {
		t.Errorf("Here looks regions up in %q, want %q", region.Boundaries, s.boundaryURL)
	}
	_, body := get(t, s, "/")
	if !strings.Contains(body, `class="btn here"`) || !strings.Contains(body, `class="herestatus"`) ||
		!strings.Contains(body, string(s.hereJS)[:200]) {
		t.Error("/ has no Here button, status, or here.js")
	}
	// One line: All, Now, a rule, and the twelve months.
	if len(mon.Groups) != 1 || !mon.Groups[0].OneLine || mon.AllURL != "" {
		t.Fatalf("the Month setting has %d groups, one line %t, All link %q; want its choices on one line",
			len(mon.Groups), len(mon.Groups) > 0 && mon.Groups[0].OneLine, mon.AllURL)
	}
	ch := mon.Groups[0].Choices
	var labels []string
	for i, c := range ch {
		labels = append(labels, c.Label)
		if c.Sep != (i == 2) {
			t.Errorf("month choice %s has a rule before it: %t; want one only between Now and Jan", c.Label, c.Sep)
		}
	}
	if want := append([]string{"All", "Now"}, month.Abbrev[:]...); !slices.Equal(labels, want) {
		t.Errorf("the month choices are %v, want %v", labels, want)
	}
	if now := ch[1]; now.On || !now.Now || string(now.URL) != "/?by=mass&month=now&region=CA&show=pinned" {
		t.Errorf("with May named, Now is %+v; want it off and marked, linking to month=now", now)
	}
	if all := ch[0]; all.On || string(all.URL) != "/?by=mass&region=CA&show=pinned" {
		t.Errorf("with May named, All is %+v; want it off, linking to no month", all)
	}
	for _, target := range []string{"/", "/?month=all"} {
		if ch := nextAt(s, target).Controls[1].Groups[0].Choices; !ch[0].On || ch[1].On || ch[11].On {
			t.Errorf("%s: All %t, Now %t, Oct %t; want only All chosen", target, ch[0].On, ch[1].On, ch[11].On)
		}
	}
	// Now is a way to choose the current month, not a state: with October
	// chosen, October is on and Now is not, and the summary is the month.
	op := nextAt(s, "/?month=Oct")
	if ch := op.Controls[1].Groups[0].Choices; ch[1].On || !ch[11].On || op.Controls[1].Value != "Oct" {
		t.Errorf("with month=Oct: Now %t, Oct %t, summary %q; want only Oct, reading Oct", ch[1].On, ch[11].On, op.Controls[1].Value)
	}
	if _, body := get(t, s, "/"); strings.Count(body, " data-now>") != 1 || strings.Contains(body, "(now)") {
		t.Error(`/ does not mark exactly one link data-now for here.js, or still says "(now)"`)
	}
}

//go:embed testdata/next_stay_probe.js
var nextStayProbe string

// TestNextSettingStaysOpen checks, in headless Chrome, that a setting stays
// open across the page its choice loads: the probe opens Month and chooses
// May, and on the page that loads, Month is open and nothing else is; the
// same page loaded again, after no choice, opens nothing. The pages load in
// frames, which share the session's storage, since headless Chrome does not
// finish a dump across navigations.
//
// Verifies: product.nextView. Criterion: app.nextViewDrawsOneSet.
func TestNextSettingStaysOpen(t *testing.T) {
	chrome := findChrome()
	if chrome == "" {
		t.Skip("no Chrome or Chromium found (set BIRDVIEW_CHROME); staying open was not tested in a browser")
	}
	s := newTestServer(t)
	ts := httptest.NewServer(instrument(s, nextStayProbe, nil))
	defer ts.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	out, err := exec.CommandContext(ctx, chrome, "--headless", "--disable-gpu",
		"--no-first-run", "--no-default-browser-check", "--disable-background-networking",
		"--disable-component-update", "--disable-sync", "--disable-extensions",
		"--user-data-dir="+t.TempDir(), "--virtual-time-budget=8000",
		"--dump-dom", ts.URL+"/?by=mass").Output()
	if err != nil {
		t.Fatalf("chrome: %v", err)
	}
	m := dataAttr("data-probe").FindSubmatch(out)
	if m == nil {
		t.Fatal("no result from the probe; it did not finish")
	}
	var res struct {
		Chosen  string   `json:"chosen"`
		Open    []string `json:"open"`
		URL     string   `json:"url"`
		Reload  []string `json:"reload"`
		Visited int      `json:"visited"`
	}
	if err := json.Unmarshal([]byte(html.UnescapeString(string(m[1]))), &res); err != nil {
		t.Fatalf("reading the probe: %v", err)
	}
	if res.URL != "/?by=mass&month=May" || !slices.Equal(res.Open, []string{"Month"}) {
		t.Errorf("after choosing %s, the page is %s with %v open; want /?by=mass&month=May with Month open", res.Chosen, res.URL, res.Open)
	}
	if len(res.Reload) != 0 {
		t.Errorf("a plain reload opened %v; want nothing", res.Reload)
	}
}

// TestNextFamilyBadges checks that the Family setting draws each family as
// its badge does on the tiles: its color, its name, and its description.
//
// Verifies: product.nextView. Criterion: app.nextViewDrawsOneSet.
func TestNextFamilyBadges(t *testing.T) {
	s := newTestServer(t)
	var fam *nextControl
	p := nextAt(s, "/")
	for i := range p.Controls {
		if p.Controls[i].Label == "Family" {
			fam = &p.Controls[i]
		}
	}
	if fam == nil {
		t.Fatal("no Family setting")
	}
	n := 0
	for _, g := range fam.Groups {
		for _, c := range g.Choices {
			n++
			var want familyOption
			for _, f := range s.fams {
				if f.Name == c.Label {
					want = f
				}
			}
			if c.Color != want.Color || c.TextColor != want.TextColor || c.Title != want.Common || want.Common == "" {
				t.Errorf("family %s: color %q on %q, description %q; want %q on %q, %q",
					c.Label, c.TextColor, c.Color, c.Title, want.TextColor, want.Color, want.Common)
			}
		}
	}
	if n != len(s.fams) {
		t.Errorf("the Family setting offers %d families of %d", n, len(s.fams))
	}
	_, body := get(t, s, "/")
	f := s.fams[0]
	if !strings.Contains(body, `<a class="chip" href="/?family=`+f.Name+`"`) ||
		!strings.Contains(body, f.Name+`<em>`+html.EscapeString(f.Common)+`</em></a>`) {
		t.Errorf("%s is not drawn as a badge with its description", f.Name)
	}
}

// TestNextResetKeepsPins checks Reset and Clear pins: Reset returns every
// setting to its default and keeps the pins, in their order; Clear pins
// drops the pins and keeps every setting, and appears only when something
// is pinned.
//
// Verifies: product.nextView. Criterion: app.nextViewDrawsOneSet.
func TestNextResetKeepsPins(t *testing.T) {
	s := newTestServer(t)
	p := nextAt(s, "/?by=mass&region=FL&month=May&pin=annhum&pin=amecro&show=pinned&rev=1")
	if string(p.ResetURL) != "/?pin=annhum&pin=amecro" {
		t.Errorf("Reset links to %s, want the defaults with both pins, in order", p.ResetURL)
	}
	if string(p.ClearPinsURL) != "/?by=mass&month=May&region=FL&rev=1" {
		t.Errorf("Clear pins links to %s, want every setting kept and no pins", p.ClearPinsURL)
	}
	if p := nextAt(s, "/?by=mass"); p.ClearPinsURL != "" || string(p.ResetURL) != "/" {
		t.Errorf("with nothing pinned: Clear pins %q, Reset %q; want no Clear pins and Reset to /", p.ClearPinsURL, p.ResetURL)
	}
	_, body := get(t, s, "/?pin=amecro")
	clear, names, reset := strings.Index(body, ">Clear pins</a>"), strings.Index(body, `data-ctl="Names"`), strings.Index(body, ">Reset</a>")
	if clear < 0 || reset < 0 || clear > names || reset < names {
		t.Errorf("Clear pins at %d, Names at %d, Reset at %d; want Clear pins with Show, before the settings, and Reset after them", clear, names, reset)
	}
}

// TestNextPinnedList checks the list of pinned birds under the find box: one
// entry per pinned bird, in alphabetical order by name, its name linking to
// its tile and its pin unpinning it; and that no pin link jumps the page to a
// tile.
//
// Verifies: product.nextView. Criterion: app.nextViewDrawsOneSet.
func TestNextPinnedList(t *testing.T) {
	s := newTestServer(t)
	p := nextAt(s, "/?by=mass&month=all&pin=annhum&pin=amecro&pin=nosuchbird")
	var names []string
	for _, e := range p.Pinned {
		names = append(names, e.Anchor)
	}
	if !slices.Equal(names, []string{"s-amecro", "s-annhum"}) {
		t.Fatalf("the pinned list links to %v, want s-amecro then s-annhum (alphabetical, not pin or mass order)", names)
	}
	if string(p.Pinned[0].URL) != "/?by=mass&month=all&pin=annhum&pin=nosuchbird" {
		t.Errorf("the first entry's pin goes to %s, want the page without that pin", p.Pinned[0].URL)
	}
	// Under IOC names, thimur ("Murre Thick-billed") precedes amerob ("Robin
	// American"), though amerob precedes thimur by id, pin order, and mass.
	wp := nextAt(s, "/?by=mass&month=all&names=world&pin=amerob&pin=thimur")
	if len(wp.Pinned) != 2 || wp.Pinned[0].Anchor != "s-thimur" || wp.Pinned[1].Anchor != "s-amerob" {
		t.Errorf("under IOC names, the pinned list is %+v, want thimur before amerob", wp.Pinned)
	}
	unpinned := 0
	for _, tl := range p.Tiles {
		if !tl.Pinned {
			unpinned++
		}
		if strings.Contains(string(tl.PinURL), "#") {
			t.Errorf("%s's pin goes to %s, which jumps the page", tl.S.ID, tl.PinURL)
		}
	}
	if unpinned == 0 {
		t.Error("no unpinned bird is drawn, so no pin that pins was checked")
	}
	_, body := get(t, s, "/?pin=amecro")
	if !strings.Contains(body, `<div class="pinned" role="region" aria-label="Pinned birds">`) {
		t.Error("with a bird pinned, the page draws no pinned list")
	}
	if !strings.Contains(body, `<a class="pinname" href="#s-amecro"`) || !strings.Contains(body, `id="s-amecro"`) {
		t.Error("the pinned list's name does not link to the bird's tile")
	}
	if p := nextAt(s, "/"); len(p.Pinned) != 0 {
		t.Errorf("with nothing pinned, the list has %d entries", len(p.Pinned))
	}
	if _, body := get(t, s, "/"); strings.Contains(body, `class="pinned"`) {
		t.Error("with nothing pinned, the page still draws the pinned list")
	}
}

// TestNextHeading checks the heading above the Show choices: what is drawn,
// in words, naming every chosen setting.
//
// Verifies: product.nextView. Criterion: app.nextViewDrawsOneSet.
func TestNextHeading(t *testing.T) {
	s := newTestServer(t)
	for target, want := range map[string]string{
		"/":                     "Showing birds in every month",
		"/?show=pinned":         "Showing pinned birds",
		"/?region=NY&month=Oct": "Showing birds in New York in October",
		"/?family=Corvidae&habitat=Forest&niche=Granivore&niche=Frugivore&month=May&month=Jun": "Showing birds in May or June · Family: Corvidae · Habitat: Forest · Diet: Eats seeds or Eats fruit",
	} {
		if got := nextAt(s, target).Title; got != want {
			t.Errorf("%s: the heading reads %q, want %q", target, got, want)
		}
	}
	if !strings.Contains(mustGet(t, s, "/?region=NY"), `<h2 class="settitle">Showing birds in New York in every month</h2>`) {
		t.Error("the page does not draw the heading")
	}
}

//go:embed testdata/next_pin_probe.js
var nextPinProbe string

// TestNextPinKeepsPlace checks, in headless Chrome, that pinning a bird
// leaves the page where it was: the probe scrolls down, clicks a tile's pin
// with the navigation cancelled, and loads its address in a frame, which
// shares the session's storage; the frame opens scrolled to the same place.
//
// Verifies: product.nextView. Criterion: app.nextViewDrawsOneSet.
func TestNextPinKeepsPlace(t *testing.T) {
	chrome := findChrome()
	if chrome == "" {
		t.Skip("no Chrome or Chromium found (set BIRDVIEW_CHROME); keeping place was not tested in a browser")
	}
	s := newTestServer(t)
	ts := httptest.NewServer(instrument(s, nextPinProbe, nil))
	defer ts.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	out, err := exec.CommandContext(ctx, chrome, "--headless", "--disable-gpu",
		"--no-first-run", "--no-default-browser-check", "--disable-background-networking",
		"--disable-component-update", "--disable-sync", "--disable-extensions",
		"--user-data-dir="+t.TempDir(), "--window-size=800,600", "--virtual-time-budget=8000",
		"--dump-dom", ts.URL+"/?month=all").Output()
	if err != nil {
		t.Fatalf("chrome: %v", err)
	}
	m := dataAttr("data-probe").FindSubmatch(out)
	if m == nil {
		t.Fatal("no result from the probe; it did not finish")
	}
	var res struct {
		Href   string `json:"href"`
		Before int    `json:"before"`
		After  int    `json:"after"`
	}
	if err := json.Unmarshal([]byte(html.UnescapeString(string(m[1]))), &res); err != nil {
		t.Fatalf("reading the probe: %v", err)
	}
	if res.Before < 100 || res.After != res.Before || strings.Contains(res.Href, "#") {
		t.Errorf("pinned at scroll %d by %s; the new page opened at %d", res.Before, res.Href, res.After)
	}
}

//go:embed testdata/next_pin_align_probe.js
var nextPinAlignProbe string

// TestNextPinAlignsWithCube checks, in headless Chrome at a laptop's width
// and a phone's, that every tile's pin stands level with its mass cube: the
// foot of one within a pixel of the other's; and that no tile is wider than
// its box or draws its bird off center, as long names once made them (#4). The page is October's birds by
// length, whose first tiles include ones with a note under the footer and
// ones stretched to a taller neighbor in their row, the two cases that once
// put the pin high.
//
// Verifies: product.nextView. Criterion: app.nextViewDrawsOneSet.
func TestNextPinAlignsWithCube(t *testing.T) {
	chrome := findChrome()
	if chrome == "" {
		t.Skip("no Chrome or Chromium found (set BIRDVIEW_CHROME); pin alignment was not tested in a browser")
	}
	s := newTestServerReal(t)
	fixNextClock(t, time.October)
	ts := httptest.NewServer(instrument(s, nextPinAlignProbe, nil))
	defer ts.Close()
	for _, size := range []string{"1280,800", "390,844"} {
		ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
		out, err := exec.CommandContext(ctx, chrome, "--headless", "--disable-gpu",
			"--no-first-run", "--no-default-browser-check", "--disable-background-networking",
			"--disable-component-update", "--disable-sync", "--disable-extensions",
			"--user-data-dir="+t.TempDir(), "--window-size="+size, "--virtual-time-budget=8000",
			"--dump-dom", ts.URL+"/?by=length").Output()
		cancel()
		if err != nil {
			t.Fatalf("chrome at %s: %v", size, err)
		}
		m := dataAttr("data-probe").FindSubmatch(out)
		if m == nil {
			t.Fatalf("at %s: no result from the probe; it did not finish", size)
		}
		var res struct {
			Tiles      int      `json:"tiles"`
			Notes      int      `json:"notes"`
			Off        []string `json:"off"`
			Uncentered []string `json:"uncentered"`
		}
		if err := json.Unmarshal([]byte(html.UnescapeString(string(m[1]))), &res); err != nil {
			t.Fatalf("reading the probe: %v", err)
		}
		if res.Tiles < 50 || res.Notes == 0 {
			t.Fatalf("at %s: %d tiles measured, %d with a note; the check needs both", size, res.Tiles, res.Notes)
		}
		if len(res.Off) > 0 {
			t.Errorf("at %s, %d of %d pins are not level with the mass cube: %v", size, len(res.Off), res.Tiles, res.Off)
		}
		if len(res.Uncentered) > 0 {
			t.Errorf("at %s, %d of %d tiles are wider than their box or draw off center (#4): %v", size, len(res.Uncentered), res.Tiles, res.Uncentered)
		}
	}
}

//go:embed testdata/next_first_probe.js
var nextFirstProbe string

// TestNextFirstCommitsOnClick checks, in headless Chrome, that the Show
// setting's First choice is sent by clicking it, as the other choices are,
// not only by Enter: clicking the First button or the choice's edge sends the
// typed count, and clicking into the box does not. Without the script, First
// is still a submit button.
//
// Verifies: product.nextView. Criterion: app.nextViewDrawsOneSet.
func TestNextFirstCommitsOnClick(t *testing.T) {
	s := newTestServer(t)
	if body := mustGet(t, s, "/"); !strings.Contains(body, `<button type="submit" title="Show the first this many birds">First</button>`) {
		t.Error("First is not a submit button")
	}
	chrome := findChrome()
	if chrome == "" {
		t.Skip("no Chrome or Chromium found (set BIRDVIEW_CHROME); First was not clicked in a browser")
	}
	ts := httptest.NewServer(instrument(s, nextFirstProbe, nil))
	defer ts.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	out, err := exec.CommandContext(ctx, chrome, "--headless", "--disable-gpu",
		"--no-first-run", "--no-default-browser-check", "--disable-background-networking",
		"--disable-component-update", "--disable-sync", "--disable-extensions",
		"--user-data-dir="+t.TempDir(), "--window-size=1280,800", "--virtual-time-budget=5000",
		"--dump-dom", ts.URL+"/?by=mass").Output()
	if err != nil {
		t.Fatalf("chrome: %v", err)
	}
	m := dataAttr("data-probe").FindSubmatch(out)
	if m == nil {
		t.Fatal("no result from the probe; it did not finish")
	}
	var res struct {
		Box, Button, Edge int
		EdgeHit           bool
		Sent              []string
	}
	if err := json.Unmarshal([]byte(html.UnescapeString(string(m[1]))), &res); err != nil {
		t.Fatalf("reading the probe: %v", err)
	}
	if !res.EdgeHit {
		t.Fatal("the probe's edge click landed on something other than the First choice")
	}
	if res.Box != 0 || res.Button != 1 || res.Edge != 1 || !slices.Equal(res.Sent, []string{"7", "7"}) {
		t.Errorf("clicks sent: box %d, First %d, edge %d, counts %v; want only First and the edge, each sending 7",
			res.Box, res.Button, res.Edge, res.Sent)
	}
}

// TestNextWelcomeLine checks the home page's header: one line saying what
// BirdView draws, with Click a bird and How to read a drawing, and no key
// line ("Drawn to scale", "Solid is measured"); then New here? Try, whose
// four comparisons each pin six real birds and show only them, by length,
// whose herons and ducks draw only those two families in every month, and
// whose birds around you now is /#here-and-now with its status naming the
// Region setting.
//
// Verifies: product.nextView. Criterion: app.nextViewDrawsOneSet.
func TestNextWelcomeLine(t *testing.T) {
	s := newTestServerReal(t)
	fixNextClock(t, time.October)
	body := mustGet(t, s, "/")
	sub := regexp.MustCompile(`(?s)<p class="sub">(.*?)</p>`).FindStringSubmatch(body)
	if sub == nil || !strings.Contains(sub[1], "Click a bird for its numbers and sources.") ||
		!strings.Contains(sub[1], `birds as <a href="/docs/using">stick figures</a> to scale`) {
		t.Errorf("the header's line does not link stick figures to the guide and say Click a bird: %v", sub)
	}
	for _, gone := range []string{`class="key"`, "Drawn to scale", "Solid is measured", "How to read a drawing"} {
		if strings.Contains(body, gone) {
			t.Errorf("the home page still says %q", gone)
		}
	}
	w := regexp.MustCompile(`(?s)<p class="welcome"><b>New here\?</b> Try(.*?)</p>`).FindStringSubmatch(body)
	if w == nil {
		t.Fatal("the home page has no New here? Try line")
	}
	if !strings.Contains(w[1], `<a href="/#here-and-now">the birds around you now</a>`) ||
		!strings.Contains(w[1], `data-where="Region"`) || !strings.Contains(w[1], `data-template="/?month=__MONTH__&amp;region=__REGION__"`) ||
		!strings.Contains(w[1], `<a href="/docs/using#some-things-to-try">more ideas</a>`) {
		t.Errorf("the welcome line lacks the birds around you now, its status, or more ideas:\n%s", w[1])
	}
	// From an address with settings changed, the link keeps them, and its
	// template replaces the region and sets the month to Now, the browser's.
	if p := nextAt(s, "/?by=mass&month=May&region=FL&pin=amecro&show=pinned&find=crow"); string(p.HereNowURL) != "/?by=mass&find=crow&month=May&pin=amecro&region=FL&show=pinned#here-and-now" ||
		p.HereNowTemplate != "/?by=mass&find=crow&month=__MONTH__&pin=amecro&region=__REGION__" {
		t.Errorf("from a changed address, the birds around you now is %s with template %s; want every setting kept, then the region replaced and the month Now",
			p.HereNowURL, p.HereNowTemplate)
	}
	links := regexp.MustCompile(`<a href="([^"]+)">([^<]+)</a>`).FindAllStringSubmatch(w[1], -1)
	sizes := 0
	for _, l := range links {
		href, text := html.UnescapeString(l[1]), l[2]
		switch {
		case strings.HasSuffix(text, " birds"):
			sizes++
			u, _ := url.Parse(href)
			pins := u.Query()["pin"]
			p := nextAt(s, href)
			if len(pins) != 6 || u.Query().Get("show") != "pinned" || u.Query().Get("by") != "length" ||
				len(p.Pinned) != 6 || len(p.Tiles) != 6 {
				t.Errorf("%s (%s): %d pins, %d known, %d drawn; want six real birds shown alone by length",
					text, href, len(pins), len(p.Pinned), len(p.Tiles))
			}
		case text == "herons and ducks":
			p := nextAt(s, href)
			if len(p.Tiles) < 20 {
				t.Errorf("herons and ducks draw %d birds", len(p.Tiles))
			}
			for _, tl := range p.Tiles {
				if f := tl.S.Family; f != "Ardeidae" && f != "Anatidae" {
					t.Errorf("herons and ducks draw %s, of %s", tl.S.ID, f)
				}
			}
			if !strings.Contains(href, "month=all") {
				t.Errorf("herons and ducks is %s; want every month", href)
			}
		}
	}
	if sizes != 4 {
		t.Errorf("the welcome line compares %d sizes of birds, want 4 (small, medium, large, huge)", sizes)
	}
}

// TestNextKeepsEveryParameter audits every transition on the home page: from
// an address setting every parameter, each link changes only the setting it
// belongs to and keeps every other parameter, show=pinned and the find box's
// text included. A setting's choice may change its own key (Order by, its
// key and rev); a pin, the pins; Show, show. The transitions meant to drop
// something drop exactly that: Reset keeps only the pins, Clear pins drops
// them and Pinned only, ×  drops the find box's text, and the birds around
// you now replaces the region, sets the month to Now, and leaves Pinned only.
// A tile's species page carries only the units and names.
//
// Verifies: product.nextView. Criterion: app.nextViewDrawsOneSet.
func TestNextKeepsEveryParameter(t *testing.T) {
	s := newTestServerReal(t)
	fixNextClock(t, time.October)
	const every = "/?by=mass&family=Corvidae&find=crow&habitat=Forest&month=May&names=world&niche=Omnivore&pin=amecro&pin=blujay&region=NY&rev=1&units=metric"
	// From Pinned only and from a count, so that no broken link can pass
	// for one that changes the same keys on purpose.
	for _, c := range []struct {
		from               string
		clearPins, hereNow []string
	}{
		{every + "&show=pinned", []string{"pin", "show"}, []string{"month", "region", "show"}},
		{every + "&show=50", []string{"pin"}, []string{"month", "region"}},
	} {
		from, clearPins, hereNow := c.from, c.clearPins, c.hereNow
		orig, _ := url.ParseQuery(strings.TrimPrefix(from, "/?"))
		p := nextAt(s, from)
		if len(p.Tiles) == 0 {
			t.Fatal("the audit page draws nothing; its links prove little")
		}
		body := mustGet(t, s, from)
		diff := func(q url.Values) []string {
			var changed []string
			keys := map[string]bool{}
			for k := range orig {
				keys[k] = true
			}
			for k := range q {
				keys[k] = true
			}
			for k := range keys {
				if !slices.Equal(orig[k], q[k]) {
					changed = append(changed, k)
				}
			}
			slices.Sort(changed)
			return changed
		}
		special := map[string][]string{
			string(p.ResetURL):     {"by", "family", "find", "habitat", "month", "names", "niche", "region", "rev", "show", "units"},
			string(p.ClearPinsURL): clearPins,
			string(p.CloseURL):     {"find"},
		}
		// The welcome line's examples are fixed starting points, not
		// transitions; its birds around you now is checked here.
		welcome := regexp.MustCompile(`(?s)<p class="welcome">.*?</p>`).FindString(body)
		if strings.Contains(welcome, html.EscapeString(string(p.HereNowURL))) == false {
			t.Errorf("the welcome line does not link to %s", p.HereNowURL)
		}
		if u, _ := url.Parse(string(p.HereNowURL)); u.Fragment != "here-and-now" || len(diff(u.Query())) != 0 {
			t.Errorf("the birds around you now links to %s; want this address with #here-and-now", p.HereNowURL)
		}
		checked := 0
		for _, m := range regexp.MustCompile(`href="(/(?:\?[^"]*)?)"`).FindAllStringSubmatch(strings.Replace(body, welcome, "", 1), -1) {
			href, frag, _ := strings.Cut(html.UnescapeString(m[1]), "#")
			m = []string{m[0], href, frag}
			u, _ := url.Parse(href)
			q := u.Query()
			changed := diff(q)
			checked++
			if want, ok := special[href]; ok {
				if !slices.Equal(changed, want) {
					t.Errorf("%s changes %v, want exactly %v", href, changed, want)
				}
				continue
			}
			// The find box's text is no setting's: only × and Reset drop it.
			if slices.Contains(changed, "find") {
				t.Errorf("%s drops the find box's text; only × and Reset may", href)
				continue
			}
			switch {
			case len(changed) == 0:
			case len(changed) == 1:
			case len(changed) == 2 && changed[0] == "by" && changed[1] == "rev":
			default:
				t.Errorf("%s changes %v; a control may change only its own setting", href, changed)
			}
		}
		if checked < 100 {
			t.Errorf("only %d links audited", checked)
		}
		// The forms send every parameter but their own.
		for _, f := range regexp.MustCompile(`(?s)<form class="([^"]+)" action="/" method="get">(.*?)</form>`).FindAllStringSubmatch(body, -1) {
			q := url.Values{}
			for _, h := range regexp.MustCompile(`<input type="hidden" name="([^"]+)" value="([^"]*)">`).FindAllStringSubmatch(f[2], -1) {
				q.Add(h[1], html.UnescapeString(h[2]))
			}
			own := "show"
			if strings.HasPrefix(f[1], "search") {
				own = "find"
			}
			q[own] = orig[own]
			if c := diff(q); len(c) != 0 {
				t.Errorf("the %s form loses %v", f[1], c)
			}
		}
		// The templates the browser fills in.
		here := p.Controls[2].HereURL
		if p.Controls[2].Label != "Region" {
			t.Fatalf("control 2 is %s, not Region", p.Controls[2].Label)
		}
		hq, _ := url.ParseQuery(strings.TrimPrefix(strings.Replace(here, hereSlot, "CA", 1), "/?"))
		if c := diff(hq); !slices.Equal(c, []string{"region"}) {
			t.Errorf("Here changes %v, want only the region", c)
		}
		nq, _ := url.ParseQuery(strings.TrimPrefix(strings.Replace(p.HereNowTemplate, hereSlot, "CA", 1), "/?"))
		if c := diff(nq); !slices.Equal(c, hereNow) {
			t.Errorf("the birds around you now changes %v, want the region, the month, and leaving Pinned only", c)
		}
		// A species page carries only its units and names; Back returns by
		// history (TestSpeciesBackUsesHistory).
		tl := p.Tiles[0]
		if want := "/species/" + tl.S.ID + "?names=world&units=metric"; string(tl.DetailURL) != want {
			t.Errorf("%s's species page is %s, want %s", tl.S.ID, tl.DetailURL, want)
		}
	}
	// The reported bug: a month from Pinned only stays Pinned only.
	if got := string(nextToggle(url.Values{"by": {"length"}, "pin": {"blujay"}, "show": {"pinned"}}, "month", "Nov")); got != "/?by=length&month=Nov&pin=blujay&show=pinned" {
		t.Errorf("choosing Nov from Pinned only goes to %s", got)
	}
}

// TestNextShortValues checks what each setting's summary says, kept short:
// the order as its name and an arrow, months by their three letters, all
// for every region, imperial or metric, and North Amer or world.
//
// Verifies: product.nextView. Criterion: app.nextViewDrawsOneSet.
func TestNextShortValues(t *testing.T) {
	s := newTestServer(t)
	fixNextClock(t, time.October)
	values := func(target string) map[string]string {
		out := map[string]string{}
		for _, c := range nextAt(s, target).Controls {
			out[c.Label] = c.Value
		}
		return out
	}
	for target, want := range map[string]map[string]string{
		"/":           {"Order by": "frequency \u2193", "Month": "all", "Region": "all", "Units": "imperial", "Names": "North Amer"},
		"/?month=Oct": {"Month": "Oct"},
		"/?by=length&month=May&month=Jun&units=metric&names=world": {"Order by": "length \u2191", "Month": "May or Jun", "Units": "metric", "Names": "world"},
		"/?by=mass&rev=1": {"Order by": "mass \u2193"},
	} {
		got := values(target)
		for k, v := range want {
			if got[k] != v {
				t.Errorf("%s: %s reads %q, want %q", target, k, got[k], v)
			}
		}
	}
	// Units' choices carry the names its summary reads.
	for _, c := range nextAt(s, "/").Controls {
		if c.Label != "Units" {
			continue
		}
		var labels []string
		for _, ch := range c.Groups[0].Choices {
			labels = append(labels, ch.Label)
		}
		if !slices.Equal(labels, []string{"Imperial", "Metric"}) {
			t.Errorf("Units offers %v, want Imperial and Metric, as its summary reads", labels)
		}
	}
}

//go:embed testdata/detail_back_probe.js
var detailBackProbe string

// TestSpeciesBackUsesHistory checks Back on a species page in headless
// Chrome: reached from the home page, it goes back by the browser's history,
// so the home page returns as it was; opened from a shared link, it follows
// its link to the home page in the species page's units and names. Species
// links from the home page carry only those two settings.
//
// Verifies: product.nextView. Criterion: app.nextViewDrawsOneSet.
func TestSpeciesBackUsesHistory(t *testing.T) {
	s := newTestServer(t)
	p := nextAt(s, "/?by=mass&month=all&units=metric&pin=amecro&region=FL&show=all")
	if len(p.Tiles) == 0 || string(p.Tiles[0].DetailURL) != "/species/"+p.Tiles[0].S.ID+"?units=metric" {
		t.Fatalf("the first tile links to %q; want its species page with only the units", p.Tiles[0].DetailURL)
	}
	chrome := findChrome()
	if chrome == "" {
		t.Skip("no Chrome or Chromium found (set BIRDVIEW_CHROME); Back was not tested in a browser")
	}
	ts := httptest.NewServer(instrument(s, detailBackProbe, nil))
	defer ts.Close()
	for _, c := range []struct {
		name, start string
		back        int
		followed    bool
	}{
		{"from the home page", "/?by=mass&month=all&units=metric", 1, false},
		{"from a shared link", "/species/amecro?units=metric", 0, true},
	} {
		ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
		out, err := exec.CommandContext(ctx, chrome, "--headless", "--disable-gpu",
			"--no-first-run", "--no-default-browser-check", "--disable-background-networking",
			"--disable-component-update", "--disable-sync", "--disable-extensions",
			"--user-data-dir="+t.TempDir(), "--virtual-time-budget=5000",
			"--dump-dom", ts.URL+c.start).Output()
		cancel()
		if err != nil {
			t.Fatalf("%s: chrome: %v", c.name, err)
		}
		m := dataAttr("data-probe").FindSubmatch(out)
		if m == nil {
			t.Fatalf("%s: no result from the probe", c.name)
		}
		var res struct {
			Back     int    `json:"back"`
			Followed bool   `json:"followed"`
			Href     string `json:"href"`
			Text     string `json:"text"`
		}
		if err := json.Unmarshal([]byte(html.UnescapeString(string(m[1]))), &res); err != nil {
			t.Fatalf("%s: reading the probe: %v", c.name, err)
		}
		if res.Back != c.back || res.Followed != c.followed || res.Href != "/?units=metric" || res.Text != "\u2190 back" {
			t.Errorf("%s: Back (%q, to %s) went back %d times and followed its link %t; want %d and %t",
				c.name, res.Text, res.Href, res.Back, res.Followed, c.back, c.followed)
		}
	}
}

// TestNextTiers checks the home page's tiers and histogram: the matching
// birds break into tiers as the classic view's sections do, each bird
// within its tier's range and the tiers in the order shown; a heading comes
// before the first drawn bird of each tier drawn, and only those; the
// histogram has a bar per tier, linking to its heading on the page when
// drawn and with every bird shown when not; reversed, the tiers run the
// other way; and frequency tiers are decades of "1 in N".
//
// Verifies: product.nextView. Criterion: app.nextViewDrawsOneSet.
func TestNextTiers(t *testing.T) {
	s := newTestServerReal(t)
	fixNextClock(t, time.October)
	for _, c := range []struct {
		target string
		rev    bool
	}{{"/?by=length", false}, {"/?by=length&rev=1", true}, {"/?by=mass&show=all", false}} {
		p := nextAt(s, c.target)
		if len(p.Tiers) < 5 {
			t.Fatalf("%s: %d tiers", c.target, len(p.Tiers))
		}
		total := 0
		for i, tr := range p.Tiers {
			total += tr.Count
			if tr.Index != i+1 || tr.Min == "" || tr.Max == "" {
				t.Errorf("%s: tier %d reads %+v", c.target, i+1, tr)
			}
		}
		if total != p.Count {
			t.Errorf("%s: the tiers hold %d birds, the set %d", c.target, total, p.Count)
		}
		// Headings: one before the first drawn bird of each drawn tier, in
		// order, and each tile's bird within the tier above it.
		var heads []int
		var cur *nextTier
		var prev float64
		for _, it := range p.Items {
			switch {
			case it.Tier != nil:
				cur = it.Tier
				heads = append(heads, it.Tier.Index)
			case it.Tile != nil:
				if cur == nil {
					t.Fatalf("%s: a tile comes before any heading", c.target)
				}
				v := it.Tile.S.TotalLength.Value
				if strings.Contains(c.target, "mass") {
					v = it.Tile.S.Mass.Value
				}
				if prev != 0 && (v < prev) != c.rev && v != prev {
					t.Errorf("%s: %s out of order", c.target, it.Tile.S.ID)
				}
				prev = v
			}
		}
		drawn := 0
		for _, tr := range p.Tiers {
			if tr.Drawn {
				drawn++
			}
		}
		if len(heads) != drawn || !slices.IsSorted(heads) {
			t.Errorf("%s: headings %v for %d drawn tiers", c.target, heads, drawn)
		}
		// The histogram: a bar per tier, empty steps between.
		bars := 0
		for _, b := range p.Histogram {
			if b.Empty {
				continue
			}
			bars++
			want := "#tier-" + strconv.Itoa(b.Index)
			if !p.Tiers[b.Index-1].Drawn {
				want = string(p.ShowAllURL) + want
			}
			if string(b.URL) != want {
				t.Errorf("%s: tier %d's bar links to %s, want %s", c.target, b.Index, b.URL, want)
			}
		}
		if bars != len(p.Tiers) {
			t.Errorf("%s: %d bars for %d tiers", c.target, bars, len(p.Tiers))
		}
		if strings.Contains(c.target, "show=all") && drawn != len(p.Tiers) {
			t.Errorf("%s: %d of %d tiers drawn with every bird shown", c.target, drawn, len(p.Tiers))
		}
	}
	// Reversed, the first tier is the largest birds.
	up, down := nextAt(s, "/?by=length"), nextAt(s, "/?by=length&rev=1")
	if up.Tiers[0].Min != down.Tiers[len(down.Tiers)-1].Min || up.Tiers[len(up.Tiers)-1].Max != down.Tiers[0].Max {
		t.Errorf("reversed tiers do not mirror: first %+v, last %+v", down.Tiers[0], down.Tiers[len(down.Tiers)-1])
	}
	// Frequency: decades of "1 in N", most recorded first.
	f := nextAt(s, "/")
	if !strings.HasPrefix(f.Tiers[0].Min, "1 in ") || f.Tiers[0].Label != "Frequency" {
		t.Errorf("the first frequency tier reads %+v", f.Tiers[0])
	}
	// Each length tier holds the birds up to 10% longer than its first, and
	// the next tier starts past that (product.sectionThresholdDefault).
	all := nextAt(s, "/?by=length&show=all")
	var first, prevFirst float64
	cur := -1
	for _, it := range all.Items {
		switch {
		case it.Tier != nil:
			cur, prevFirst, first = it.Tier.Index, first, 0
		case it.Tile != nil:
			v := it.Tile.S.TotalLength.Value
			if first == 0 {
				first = v
				if prevFirst != 0 && v <= prevFirst*1.10 {
					t.Errorf("tier %d starts at %.1f mm, within 10%% of tier %d's first, %.1f mm", cur, v, cur-1, prevFirst)
				}
			} else if v > first*1.10+1e-9 {
				t.Errorf("tier %d holds %s at %.1f mm, more than 10%% past its first, %.1f mm", cur, it.Tile.S.ID, v, first)
			}
		}
	}
	body := mustGet(t, s, "/?by=length")
	drawnTiers := 0
	for _, tr := range nextAt(s, "/?by=length").Tiers {
		if tr.Drawn {
			drawnTiers++
		}
	}
	if !strings.Contains(body, `<header id="top">`) || strings.Count(body, `<a class="totop" href="#top">back to top</a>`) != drawnTiers {
		t.Errorf("want a back to top link, to the header, beside each of the %d drawn tiers", drawnTiers)
	}
	if !strings.Contains(body, `<h3 class="tierhd" id="tier-1"><span class="tier">Length Tier 1</span>`) ||
		!strings.Contains(body, `<nav class="hist" aria-label="Birds in each tier">`) {
		t.Error("the page draws no tier heading or histogram")
	}
}

// TestBandingCodesOnThePages checks where the banding codes show: beside
// the name with the eBird code on every tile, on the species page, and in
// the find box's list, which finds a bird by its code in any case. Addresses
// stay the eBird code, which every bird has: /species/AMRO is no page.
//
// Verifies: product.bandingCodes. Criterion: app.bandingCodesShown.
func TestBandingCodesOnThePages(t *testing.T) {
	s := newTestServerReal(t)
	fixNextClock(t, time.October)
	codes := `<span class="cnname">American Robin</span><span class="code banding" title="Banding code, from the USGS Bird Banding Laboratory">AMRO</span><span class="code" title="eBird code">(amerob)</span>`
	if !strings.Contains(mustGet(t, s, "/?pin=amerob"), codes) {
		t.Error("American Robin's tile does not show AMRO and (amerob) beside its name")
	}
	if !strings.Contains(mustGet(t, s, "/species/amerob"), codes) {
		t.Error("the species page does not show AMRO and (amerob) beside the name")
	}
	for _, q := range []string{"AMRO", "amro"} {
		p := nextAt(s, "/?find="+q)
		if len(p.Matches) != 1 || p.Matches[0].ID != "amerob" || p.Matches[0].Banding != "AMRO" {
			t.Errorf("find=%s lists %+v; want American Robin alone, with AMRO", q, p.Matches)
		}
	}
	if !strings.Contains(mustGet(t, s, "/?find=amro"), `American Robin <span class="code">AMRO</span> <span class="code">(amerob)</span> <i>Turdus migratorius</i>`) {
		t.Error("the find list does not show AMRO and (amerob) beside the name")
	}
	for _, path := range []string{"/species/AMRO", "/species/amro"} {
		w := httptest.NewRecorder()
		s.ServeHTTP(w, httptest.NewRequest(http.MethodGet, path, nil))
		if w.Code != http.StatusNotFound {
			t.Errorf("GET %s = %d, want 404: addresses use the eBird code", path, w.Code)
		}
	}
}

//go:embed testdata/detail_pin_probe.js
var detailPinProbe string

// TestSpeciesPinAsksTheHomePage checks the species page's Pin in headless
// Chrome: opened from the home page, it shows whether the bird is pinned
// there, and clicking it goes back and leaves the bird for the home page,
// which, loaded again, asks for its own address with that bird toggled.
// Without the script it is a link to the home page with the bird pinned, and
// a bird left off the home page has none.
//
// Verifies: product.nextView. Criterion: app.nextViewDrawsOneSet.
func TestSpeciesPinAsksTheHomePage(t *testing.T) {
	s := newTestServer(t)
	body := mustGet(t, s, "/species/amecro")
	if !strings.Contains(body, `<a class="btn pinbtn" href="/?pin=amecro" data-pin="amecro" aria-pressed="false">`) {
		t.Error("the species page has no Pin linking to the home page with the bird pinned")
	}
	real := newTestServerReal(t)
	exotic := ""
	for i := range real.all.Species {
		if real.all.Species[i].Exotic != "" {
			exotic = real.all.Species[i].ID
			break
		}
	}
	if exotic == "" {
		t.Fatal("the dataset has no bird left off the home page; the check needs one")
	}
	if strings.Contains(mustGet(t, real, "/species/"+exotic), `data-pin=`) {
		t.Errorf("%s, left off the home page, has a Pin", exotic)
	}
	chrome := findChrome()
	if chrome == "" {
		t.Skip("no Chrome or Chromium found (set BIRDVIEW_CHROME); the species page's Pin was not tested in a browser")
	}
	var mu sync.Mutex
	var requests []string
	ts := httptest.NewServer(instrument(s, detailPinProbe, func(uri string) {
		mu.Lock()
		requests = append(requests, uri)
		mu.Unlock()
	}))
	defer ts.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	out, err := exec.CommandContext(ctx, chrome, "--headless", "--disable-gpu",
		"--no-first-run", "--no-default-browser-check", "--disable-background-networking",
		"--disable-component-update", "--disable-sync", "--disable-extensions",
		"--user-data-dir="+t.TempDir(), "--virtual-time-budget=8000",
		"--dump-dom", ts.URL+"/?month=all&pin=annhum").Output()
	if err != nil {
		t.Fatalf("chrome: %v", err)
	}
	m := dataAttr("data-probe").FindSubmatch(out)
	if m == nil {
		t.Fatal("no result from the probe")
	}
	type frame struct {
		Label, Pressed, Toggle string
		Back                   int
	}
	var res struct{ SpeciesA, SpeciesB frame }
	if err := json.Unmarshal([]byte(html.UnescapeString(string(m[1]))), &res); err != nil {
		t.Fatalf("reading the probe: %v", err)
	}
	if a := res.SpeciesA; a.Label != "Pin" || a.Pressed != "false" || a.Back != 1 || a.Toggle != "amecro" {
		t.Errorf("amecro's Pin: %+v; want Pin, not pressed, going back once and leaving amecro to toggle", a)
	}
	if b := res.SpeciesB; b.Label != "Unpin" || b.Pressed != "true" {
		t.Errorf("annhum's Pin: %+v; want Unpin, pressed, since the home page pins it", b)
	}
	mu.Lock()
	defer mu.Unlock()
	if !slices.Contains(requests, "/?month=all&pin=annhum&bird=amecro") {
		t.Errorf("the home page, loaded again, did not ask to toggle amecro; requests %v", requests)
	}
}
