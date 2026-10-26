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

// TestNextLandingIsThisMonthByFrequency checks /next with no filter: every
// bird the month filter admits for the clock's month, or the months the
// address names, or every bird with month=all, most often recorded first.
//
// Verifies: product.nextView. Criterion: app.nextViewDrawsOneSet.
func TestNextLandingIsThisMonthByFrequency(t *testing.T) {
	s := newTestServerReal(t)
	fixNextClock(t, time.October)
	old := nextShowAll
	nextShowAll = 5000 // every bird drawn, so every admitted bird can be looked for
	t.Cleanup(func() { nextShowAll = old })
	for _, c := range []struct {
		target string
		months []int
	}{
		{"/next", []int{9}},
		{"/next?month=May", []int{4}},
		{"/next?month=May&month=Jun", []int{4, 5}},
		{"/next?month=all", nil},
	} {
		opt := nextOptions(httptest.NewRequest(http.MethodGet, "/next", nil))
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
	if a, b := nextAt(s, "/next").Count, nextAt(s, "/next?month=all").Count; a >= b {
		t.Errorf("/next shows %d birds and month=all %d; the month does not narrow", a, b)
	}
	for _, c := range nextAt(s, "/next").Controls {
		if c.Label == "Show" {
			t.Error("the count line still has a Show setting")
		}
	}
	if c := nextAt(s, "/next").Controls; len(c) == 0 || c[0].Label != "Order by" || !strings.HasPrefix(c[0].Value, "frequency") {
		t.Errorf("the count line starts with %+v; want Order by frequency", c[0])
	}
}

// TestNextDrawsOneScale checks that every bird /next shows is drawn at the
// same pixels per millimeter, the longest bird's, whatever the order.
//
// Verifies: product.nextView. Criterion: app.nextViewDrawsOneSet.
func TestNextDrawsOneScale(t *testing.T) {
	s := newTestServerReal(t)
	fixNextClock(t, time.October)
	for _, target := range []string{"/next", "/next?by=frequency", "/next?family=Ardeidae&family=Trochilidae", "/next?month=all&by=mass"} {
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
	p := nextAt(s, "/next?by=mass&family=Parulidae")
	for i := 1; i < len(p.Tiles); i++ {
		if p.Tiles[i].S.Mass.Value < p.Tiles[i-1].S.Mass.Value {
			t.Errorf("by mass: %s after %s", p.Tiles[i].S.ID, p.Tiles[i-1].S.ID)
		}
	}
	opt := nextOptions(httptest.NewRequest(http.MethodGet, "/next?month=Oct", nil))
	p = nextAt(s, "/next?by=frequency&month=Oct")
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
	herons := nextAt(s, "/next?family=Ardeidae")
	if len(herons.Tiles) < 5 {
		t.Fatalf("family=Ardeidae shows %d birds, want every heron", len(herons.Tiles))
	}
	fl := nextAt(s, "/next?family=Ardeidae&region=FL")
	me := nextAt(s, "/next?family=Ardeidae&region=ME")
	if len(fl.Tiles) == 0 || len(me.Tiles) >= len(herons.Tiles) {
		t.Errorf("herons: %d everywhere, %d in Florida, %d in Maine; region does not narrow",
			len(herons.Tiles), len(fl.Tiles), len(me.Tiles))
	}
	for _, tl := range fl.Tiles {
		if tl.S.Family != "Ardeidae" {
			t.Errorf("family=Ardeidae&region=FL shows %s, of %s", tl.S.ID, tl.S.Family)
		}
	}
	if got, want := len(nextAt(s, "/next?family=").Tiles), len(nextAt(s, "/next").Tiles); got != want {
		t.Errorf("/next?family= shows %d birds, /next %d; an empty filter must mean all", got, want)
	}
	if got := len(nextAt(s, "/next?region=&month=").Tiles); got == 0 {
		t.Error("/next?region=&month= shows nothing; an empty filter must mean all")
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
	start := "/next?family=Ardeidae"
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
	pinned := follow(t, s, string(m[0].URL)+"&find=sharp-shinned")
	if u, _ := url.Parse(pinned); u.Query().Get("find") != "" || u.Query().Get("family") != "Ardeidae" {
		t.Errorf("pinning went to %s: want the family kept and the box emptied", pinned)
	}
	withHawk := nextAt(s, pinned)
	unpin, on, drawn := pinOf(withHawk, "shshaw")
	if !drawn || !on || withHawk.Count != before.Count+1 {
		t.Fatalf("after pinning, %d birds (was %d), hawk drawn %t pinned %t", withHawk.Count, before.Count, drawn, on)
	}
	if back, _, _ := strings.Cut(string(unpin), "#"); back != start {
		t.Errorf("unpinning the hawk goes to %s, want %s", unpin, start)
	}
	if m := nextAt(s, pinned+"&find=sharp-shinned").Matches; len(m) != 1 || !m[0].Pinned {
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
		p := nextAt(s, "/next?find="+url.QueryEscape(q))
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
	start := "/next?month=all&pin=annhum"
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
			Items []string `json:"items"`
		} `json:"typed"`
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
		if want := html.UnescapeString(fm[1]); r.Found != want {
			t.Errorf("typing %q: the script says %q, the server %q", r.Q, r.Found, want)
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
}

// TestNextToggle checks the count line's choosing rule: from all, choosing
// one selects only it; choosing more adds them; choosing a selected one
// removes it; removing the last returns to all.
//
// Verifies: product.nextView. Criterion: app.nextViewDrawsOneSet.
func TestNextToggle(t *testing.T) {
	q := func(s string) url.Values { v, _ := url.ParseQuery(s); return v }
	for _, c := range []struct{ from, v, want string }{
		{"", "FL", "/next?region=FL"},
		{"region=FL", "ME", "/next?region=FL&region=ME"},
		{"region=FL&region=ME", "FL", "/next?region=ME"},
		{"region=ME", "ME", "/next"},
		{"region=", "ME", "/next?region=ME"},
		{"region=FL&find=hawk&all=1&by=mass", "FL", "/next?by=mass"},
		{"pin=blujay&show=pinned", "FL", "/next?pin=blujay&region=FL"},
	} {
		if got := string(nextToggle(q(c.from), "region", c.v)); got != c.want {
			t.Errorf("toggling %s from %q = %s, want %s", c.v, c.from, got, c.want)
		}
	}
}

// TestNextShowAll checks the cap: a set larger than nextShowAll draws the
// top of its order and a link that shows them all, in the same order, and a
// set within it draws every bird with no link.
//
// Verifies: product.nextView. Criterion: app.nextViewDrawsOneSet.
func TestNextShowAll(t *testing.T) {
	s := newTestServerReal(t)
	fixNextClock(t, time.October)
	p := nextAt(s, "/next")
	if p.Count <= nextShowAll {
		t.Fatalf("/next has %d birds; the check needs more than %d", p.Count, nextShowAll)
	}
	if len(p.Tiles) != nextShowAll || p.MoreURL == "" || p.Hidden != p.Count-nextShowAll {
		t.Errorf("/next draws %d of %d, hides %d, more link %q", len(p.Tiles), p.Count, p.Hidden, p.MoreURL)
	}
	all := nextAt(s, string(p.MoreURL))
	if len(all.Tiles) != p.Count || all.MoreURL != "" {
		t.Errorf("%s draws %d of %d", p.MoreURL, len(all.Tiles), p.Count)
	}
	for i := range p.Tiles {
		if p.Tiles[i].S != all.Tiles[i].S {
			t.Errorf("Show all reorders: %s at %d, was %s", all.Tiles[i].S.ID, i, p.Tiles[i].S.ID)
			break
		}
	}
	_, body := get(t, s, "/next")
	if want := `<span class="shown">First 100 of ` + strconv.Itoa(p.Count) + ` birds</span>
  <a class="showall" href="` + html.EscapeString(string(p.MoreURL)) + `">Show all</a>`; !strings.Contains(body, want) {
		t.Errorf("the count line has no Show all after its count; want %q", want)
	}
	if strings.Count(body, `href="`+html.EscapeString(string(p.MoreURL))+`"`) != 2 {
		t.Error("Show all is not both on the count line and at the end")
	}
	if p := nextAt(s, "/next?family=Ardeidae"); p.MoreURL != "" || len(p.Tiles) != p.Count {
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
	c := choices("/next?by=mass&family=Anatidae")
	if !c["Mass"].On || string(c["Mass"].URL) != "/next?by=mass&family=Anatidae&rev=1" {
		t.Errorf("chosen Mass links to %s, want its reverse", c["Mass"].URL)
	}
	if string(c["Length"].URL) != "/next?by=length&family=Anatidae" {
		t.Errorf("Length links to %s", c["Length"].URL)
	}
	c = choices("/next?by=mass&family=Anatidae&rev=1")
	if string(c["Mass"].URL) != "/next?by=mass&family=Anatidae" || string(c["Length"].URL) != "/next?by=length&family=Anatidae" {
		t.Errorf("reversed: Mass links to %s and Length to %s; want back, and Length unreversed", c["Mass"].URL, c["Length"].URL)
	}
	up, down := nextAt(s, "/next?by=mass&family=Anatidae"), nextAt(s, "/next?by=mass&family=Anatidae&rev=1")
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
	fwd, rev := nextAt(s, "/next?family=Anatidae"), nextAt(s, "/next?family=Anatidae&rev=1")
	if fwd.Tiles[0].S != rev.Tiles[len(rev.Tiles)-1].S {
		t.Errorf("the most recorded duck, %s, is not at the end of the reversed order", fwd.Tiles[0].S.ID)
	}
	if v := nextAt(s, "/next?rev=1").Controls[0].Value; v != "frequency, least often recorded first" {
		t.Errorf("reversed frequency reads %q", v)
	}
}

// TestNextWithoutAMonthIsNotStored checks that /next with no month, which
// answers from the clock, is neither stored nor revalidated, and that with a
// month it is cached as any page is.
//
// Verifies: product.nextView. Criterion: app.nextViewDrawsOneSet.
func TestNextWithoutAMonthIsNotStored(t *testing.T) {
	s := newTestServer(t)
	s.build = "test-build"
	for target, store := range map[string]bool{
		"/next": false, "/next?region=FL": false, "/next?month=": false,
		"/next?month=Oct": true, "/next?month=Oct&region=FL": true,
	} {
		w := httptest.NewRecorder()
		s.ServeHTTP(w, httptest.NewRequest(http.MethodGet, target, nil))
		if w.Code != http.StatusOK {
			t.Errorf("GET %s = %d", target, w.Code)
		}
		cc, etag := w.Header().Get("Cache-Control"), w.Header().Get("ETag")
		if store && (cc == "no-store" || etag == "") {
			t.Errorf("%s: Cache-Control %q, ETag %q; a month makes it cacheable", target, cc, etag)
		}
		if !store && (cc != "no-store" || etag != "") {
			t.Errorf("%s: Cache-Control %q, ETag %q; want no-store and no validator", target, cc, etag)
		}
	}
	_, body := get(t, s, "/")
	if strings.Contains(body, `href="/next`) {
		t.Error("the index links to /next; the prototype is not linked until it is compared")
	}
}

// TestNextControlsOpenOneAtATime checks that every setting on the count line
// is a <details> sharing one name, which is what makes the browser close the
// others when one opens, and that Reset is there and goes to the landing set.
//
// Verifies: product.nextView. Criterion: app.nextViewDrawsOneSet.
func TestNextControlsOpenOneAtATime(t *testing.T) {
	s := newTestServer(t)
	_, body := get(t, s, "/next?region=FL")
	n := strings.Count(body, `<details class="ctl"`)
	named := len(regexp.MustCompile(`<details class="ctl" name="ctl"[ >]`).FindAllString(body, -1))
	if n < 8 || named != n {
		t.Errorf("%d of %d count-line settings share the name ctl; want all, and at least 8", named, n)
	}
	if !strings.Contains(body, `<a class="reset" href="/next"`) || !strings.Contains(body, ">Reset</a>") {
		t.Error("no Reset control linking to /next")
	}
}

// TestNextAddedPastTheFirstIsDrawn checks that a bird pinned from past the
// first nextShowAll is drawn in its place in the order, with gap rows where
// unpinned birds are left out before, between, and after pinned birds, and
// none once Show all is chosen.
//
// Verifies: product.nextView. Criterion: app.nextViewDrawsOneSet.
func TestNextAddedPastTheFirstIsDrawn(t *testing.T) {
	s := newTestServerReal(t)
	fixNextClock(t, time.October)
	before := nextAt(s, "/next")
	all := nextAt(s, "/next?all=1")
	last := all.Tiles[len(all.Tiles)-1].S // the least recorded: far past the first
	if hasTile(before, last.ID) {
		t.Fatalf("%s is drawn already; the check needs a bird past the first %d", last.ID, nextShowAll)
	}
	m := nextAt(s, "/next?find="+url.QueryEscape(last.ScientificName)).Matches
	if len(m) != 1 || m[0].Pinned {
		t.Fatalf("finding %s, a bird not drawn, lists %+v; want it unpinned", last.ScientificName, m)
	}
	after := nextAt(s, follow(t, s, string(m[0].URL)))
	if !hasTile(after, last.ID) {
		t.Errorf("after adding %s from past the first %d, it is not drawn", last.ID, nextShowAll)
	}
	// It was the very last of the hidden birds, so one fewer is hidden, all
	// of them sit in one gap before it, and none sit after it.
	if after.AddedPast != 1 || after.First != nextShowAll || after.Hidden != before.Hidden-1 {
		t.Errorf("after adding: first %d, added past it %d, hidden %d (was %d)", after.First, after.AddedPast, after.Hidden, before.Hidden)
	}
	if len(after.Tiles) != len(before.Tiles)+1 {
		t.Errorf("after adding one bird, %d drawn (was %d)", len(after.Tiles), len(before.Tiles))
	}
	if n := len(after.Items); n != nextShowAll+2 || after.Items[n-2].Hidden != after.Hidden || after.Items[n-1].Tile == nil {
		t.Errorf("pinning the last bird leaves items ending %+v, %+v; want a gap of %d then its tile",
			after.Items[n-2], after.Items[n-1], after.Hidden)
	}

	// Two birds mid-order by length: gaps sit before the shorter, between the
	// two, and after the longer, summing to Hidden, each linking to Show all.
	two := "/next?by=length&pin=fuwduc&pin=amecro"
	p := nextAt(s, two)
	_, body := get(t, s, two)
	tokens := regexp.MustCompile(`id="(s-[a-z0-9]+)"|<p class="gap">([0-9]+) birds? hidden — <a href="([^"]+)">Show all</a></p>`).
		FindAllStringSubmatch(body, -1)
	if len(tokens) != nextShowAll+5 {
		t.Fatalf("%s has %d tiles and gaps, want %d tiles and 3 gaps", two, len(tokens), nextShowAll+2)
	}
	tail := tokens[nextShowAll:]
	if tail[0][2] == "" || tail[1][1] != "s-amecro" || tail[2][2] == "" || tail[3][1] != "s-fuwduc" || tail[4][2] == "" {
		t.Fatalf("after the first %d, got %v; want gap, amecro, gap, fuwduc, gap", nextShowAll, tail)
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
	old := nextShowAll
	nextShowAll = 2
	t.Cleanup(func() { nextShowAll = old })
	_, sbody := get(t, small, "/next?month=all&by=mass&pin=amecro")
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
	start := nextAt(s, "/next?by=mass&all=1")
	if string(start.ShowURL) != "/next?by=mass&show=pinned" {
		t.Errorf("Show only pinned links to %s, want the order kept", start.ShowURL)
	}
	if a, b := nextAt(s, "/next?by=mass&show=all").Count, nextAt(s, "/next?by=mass").Count; a != b {
		t.Errorf("show=all draws from %d birds, the default from %d; all is the default", a, b)
	}
	empty := nextAt(s, string(start.ShowURL))
	if !empty.PinnedOnly || len(empty.Tiles) != 0 || empty.Count != 0 || empty.Title != "Birds you chose" {
		t.Errorf("pinned only: %d drawn of %d, title %q", len(empty.Tiles), empty.Count, empty.Title)
	}
	two := "/next?by=mass&pin=amecro&pin=annhum&show=pinned"
	p := nextAt(s, two)
	if got := tileIDs(p); !slices.Equal(got, []string{"annhum", "amecro"}) {
		t.Errorf("pinned only, with two pinned, draws %v, want annhum and amecro by mass", got)
	}
	if back := follow(t, s, two+"&bird=amecro"); back != "/next?by=mass&pin=annhum&show=pinned" {
		t.Errorf("removing one of two goes to %s", back)
	}
	undo := nextAt(s, string(p.ShowURL))
	if undo.PinnedOnly || !hasTile(undo, "amecro") || !hasTile(undo, "annhum") || undo.Count < 500 {
		t.Errorf("Show all matching birds (%s): pinned only %t, %d birds", p.ShowURL, undo.PinnedOnly, undo.Count)
	}
}

// TestNextHereAndNow checks the Region setting's Here and the Month
// setting's Now. Here's address, filled with a region, is the page with only
// that region chosen, every other setting kept, and the pinned-only view left;
// the page carries here.js and the boundary file to find the region with.
// Now is the first month choice, chosen when no month is named, and links
// to no month named.
//
// Verifies: product.nextView. Criterion: app.nextViewDrawsOneSet.
func TestNextHereAndNow(t *testing.T) {
	s := newTestServer(t)
	fixNextClock(t, time.October)
	p := nextAt(s, "/next?by=mass&region=CA&show=pinned&month=May")
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
	if got := strings.Replace(region.HereURL, hereSlot, "FL", 1); got != "/next?by=mass&month=May&region=FL" {
		t.Errorf("Here, filled with FL, is %s", got)
	}
	if region.Boundaries != s.boundaryURL {
		t.Errorf("Here looks regions up in %q, want %q", region.Boundaries, s.boundaryURL)
	}
	_, body := get(t, s, "/next")
	if !strings.Contains(body, `class="btn here"`) || !strings.Contains(body, `class="herestatus"`) ||
		!strings.Contains(body, string(s.hereJS)[:200]) {
		t.Error("/next has no Here button, status, or here.js")
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
	if now := ch[1]; now.On || string(now.URL) != "/next?by=mass&region=CA" {
		t.Errorf("with May named, Now is %+v; want it off, linking to no month", now)
	}
	if all := ch[0]; all.On || string(all.URL) != "/next?by=mass&month=all&region=CA" {
		t.Errorf("with May named, All is %+v; want it off, linking to month=all", all)
	}
	if ch := nextAt(s, "/next").Controls[1].Groups[0].Choices; !ch[1].On || ch[0].On || ch[11].On {
		t.Errorf("with no month named, All %t, Now %t, Oct %t; want only Now chosen", ch[0].On, ch[1].On, ch[11].On)
	}
	if ch := nextAt(s, "/next?month=all").Controls[1].Groups[0].Choices; !ch[0].On || ch[1].On {
		t.Errorf("with month=all, All %t and Now %t; want only All chosen", ch[0].On, ch[1].On)
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
		"--dump-dom", ts.URL+"/next?by=mass").Output()
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
	if res.URL != "/next?by=mass&month=May" || !slices.Equal(res.Open, []string{"Month"}) {
		t.Errorf("after choosing %s, the page is %s with %v open; want /next?by=mass&month=May with Month open", res.Chosen, res.URL, res.Open)
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
	p := nextAt(s, "/next")
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
	_, body := get(t, s, "/next")
	f := s.fams[0]
	if !strings.Contains(body, `<a class="chip" href="/next?family=`+f.Name+`"`) ||
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
	p := nextAt(s, "/next?by=mass&region=FL&month=May&pin=annhum&pin=amecro&show=pinned&rev=1")
	if string(p.ResetURL) != "/next?pin=annhum&pin=amecro" {
		t.Errorf("Reset links to %s, want the defaults with both pins, in order", p.ResetURL)
	}
	if string(p.ClearPinsURL) != "/next?by=mass&month=May&region=FL&rev=1" {
		t.Errorf("Clear pins links to %s, want every setting kept and no pins", p.ClearPinsURL)
	}
	if p := nextAt(s, "/next?by=mass"); p.ClearPinsURL != "" || string(p.ResetURL) != "/next" {
		t.Errorf("with nothing pinned: Clear pins %q, Reset %q; want no Clear pins and Reset to /next", p.ClearPinsURL, p.ResetURL)
	}
	_, body := get(t, s, "/next?pin=amecro")
	clear, reset := strings.Index(body, ">Clear pins</a>"), strings.Index(body, ">Reset</a>")
	if clear < 0 || reset < 0 || clear > reset {
		t.Errorf("Clear pins at %d, Reset at %d; want Clear pins just before Reset", clear, reset)
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
	p := nextAt(s, "/next?by=mass&month=all&pin=annhum&pin=amecro&pin=nosuchbird")
	var names []string
	for _, e := range p.Pinned {
		names = append(names, e.Anchor)
	}
	if !slices.Equal(names, []string{"s-amecro", "s-annhum"}) {
		t.Fatalf("the pinned list links to %v, want s-amecro then s-annhum (alphabetical, not pin or mass order)", names)
	}
	if string(p.Pinned[0].URL) != "/next?by=mass&month=all&pin=annhum&pin=nosuchbird" {
		t.Errorf("the first entry's pin goes to %s, want the page without that pin", p.Pinned[0].URL)
	}
	// Under IOC names, thimur ("Murre Thick-billed") precedes amerob ("Robin
	// American"), though amerob precedes thimur by id, pin order, and mass.
	wp := nextAt(s, "/next?by=mass&month=all&names=world&pin=amerob&pin=thimur")
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
	_, body := get(t, s, "/next?pin=amecro")
	if !strings.Contains(body, `<div class="pinned" role="region" aria-label="Pinned birds">`) {
		t.Error("with a bird pinned, the page draws no pinned list")
	}
	if !strings.Contains(body, `<a class="pinname" href="#s-amecro"`) || !strings.Contains(body, `id="s-amecro"`) {
		t.Error("the pinned list's name does not link to the bird's tile")
	}
	if p := nextAt(s, "/next"); len(p.Pinned) != 0 {
		t.Errorf("with nothing pinned, the list has %d entries", len(p.Pinned))
	}
	if _, body := get(t, s, "/next"); strings.Contains(body, `class="pinned"`) {
		t.Error("with nothing pinned, the page still draws the pinned list")
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
		"--dump-dom", ts.URL+"/next?month=all").Output()
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
// foot of one within a pixel of the other's. The page is October's birds by
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
			"--dump-dom", ts.URL+"/next?by=length").Output()
		cancel()
		if err != nil {
			t.Fatalf("chrome at %s: %v", size, err)
		}
		m := dataAttr("data-probe").FindSubmatch(out)
		if m == nil {
			t.Fatalf("at %s: no result from the probe; it did not finish", size)
		}
		var res struct {
			Tiles int      `json:"tiles"`
			Notes int      `json:"notes"`
			Off   []string `json:"off"`
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
	}
}
