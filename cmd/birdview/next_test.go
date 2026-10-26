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
	"math"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os/exec"
	"regexp"
	"slices"
	"strings"
	"testing"
	"time"

	"birdview/internal/dataset"
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

// TestNextAddsAndRemovesByName checks the find box's links: a bird not in
// the set is offered to add, and adding it keeps the filters and empties
// the box; a bird in the set, added or not, is offered to remove, and
// removing it takes out that bird and nothing else; undoing either returns
// the address it came from.
//
// Verifies: product.nextView. Criterion: app.nextViewDrawsOneSet.
func TestNextAddsAndRemovesByName(t *testing.T) {
	s := newTestServerReal(t)
	fixNextClock(t, time.October)
	start := "/next?family=Ardeidae"
	before := nextAt(s, start)
	p := nextAt(s, start+"&find=sharp-shinned")
	if len(p.Matches) != 1 || p.Matches[0].ID != "shshaw" || p.Matches[0].Shown {
		t.Fatalf("find=sharp-shinned lists %+v, want Sharp-shinned Hawk to add", p.Matches)
	}
	added := follow(t, s, string(p.Matches[0].URL))
	u, _ := url.Parse(added)
	if u.Query().Get("find") != "" || u.Query().Get("family") != "Ardeidae" {
		t.Errorf("adding went to %s: want the family kept and the box emptied", added)
	}
	withHawk := nextAt(s, added)
	if !hasTile(withHawk, "shshaw") || withHawk.Count != before.Count+1 {
		t.Fatalf("after adding, %d birds (was %d), hawk shown %t", withHawk.Count, before.Count, hasTile(withHawk, "shshaw"))
	}
	for _, tl := range withHawk.Tiles {
		if (tl.RemoveURL != "") != (tl.S.ID == "shshaw") {
			t.Errorf("%s: remove link %q; only the added bird has one on its tile", tl.S.ID, tl.RemoveURL)
		}
	}
	m := nextAt(s, added+"&find=sharp-shinned").Matches
	if len(m) != 1 || !m[0].Shown {
		t.Fatalf("an added bird found again is %+v; want it offered to remove", m)
	}
	if back := follow(t, s, string(m[0].URL)); back != start {
		t.Errorf("removing the added bird went to %s, want %s", back, start)
	}

	// A bird of the set, not added by name, is removed too, and put back.
	m = nextAt(s, start+"&find=great+blue+heron").Matches
	if len(m) != 1 || !m[0].Shown {
		t.Fatalf("find=great blue heron lists %+v; want Great Blue Heron to remove", m)
	}
	gone := follow(t, s, string(m[0].URL))
	without := nextAt(s, gone)
	if hasTile(without, m[0].ID) || without.Count != before.Count-1 {
		t.Errorf("after removing, %d birds (was %d), heron shown %t", without.Count, before.Count, hasTile(without, m[0].ID))
	}
	m2 := nextAt(s, gone+"&find=great+blue+heron").Matches
	if len(m2) != 1 || m2[0].Shown {
		t.Fatalf("a removed bird found again is %+v; want it offered to add", m2)
	}
	if back := follow(t, s, string(m2[0].URL)); back != start {
		t.Errorf("putting the heron back went to %s, want %s", back, start)
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
	start := "/next?month=all&exclude=annhum"
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
		for _, l := range regexp.MustCompile(`<li><a class="(btn(?: rm)?)" href="([^"]*)">`).FindAllStringSubmatch(body, -1) {
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
	named := strings.Count(body, `<details class="ctl" name="ctl">`)
	if n < 8 || named != n {
		t.Errorf("%d of %d count-line settings share the name ctl; want all, and at least 8", named, n)
	}
	if !strings.Contains(body, `<a class="reset" href="/next"`) || !strings.Contains(body, ">Reset</a>") {
		t.Error("no Reset control linking to /next")
	}
}
