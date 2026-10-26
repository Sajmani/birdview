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
	"math"
	"net/http"
	"net/http/httptest"
	"net/url"
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

// TestNextLandingIsCommonestOfEachFamily checks /next with no filter against
// the rule, worked out here from the counts: of the birds at 1 in 1,000 or
// more of the month's records across every region, the most often recorded
// in each family, one per family. The month is the clock's, or the
// address's when it names one.
//
// Verifies: product.nextView. Criterion: app.nextViewDrawsOneSet.
func TestNextLandingIsCommonestOfEachFamily(t *testing.T) {
	s := newTestServerReal(t)
	for _, c := range []struct {
		clock  time.Month
		target string
		month  int
	}{
		{time.October, "/next", 9},
		{time.October, "/next?month=May", 4},
	} {
		fixNextClock(t, c.clock)
		p := nextAt(s, c.target)
		share := map[string]float64{}
		best := map[string]string{}
		for i := range s.ds.Species {
			sp := &s.ds.Species[i]
			var num, den int64
			for j, id := range append(append([]string{}, sp.Regions...), sp.BelowRegions...) {
				var months [][12]int64
				k := j
				if j < len(sp.Regions) {
					months = sp.RegionMonths
				} else {
					months, k = sp.BelowMonths, j-len(sp.Regions)
				}
				if _, ok := s.ds.RegionTotals[id]; ok && k < len(months) {
					num += months[k][c.month]
				}
			}
			for id := range s.ds.RegionTotals {
				den += s.ds.RegionMonthlyTotals[id][c.month]
			}
			if den == 0 || float64(num)/float64(den) < 1e-3 {
				continue
			}
			f := float64(num) / float64(den)
			if b, ok := best[sp.Family]; !ok || f > share[b] {
				best[sp.Family] = sp.ID
			}
			share[sp.ID] = f
		}
		got := map[string]bool{}
		fams := map[string]bool{}
		for _, tl := range p.Tiles {
			if fams[tl.S.Family] {
				t.Errorf("%s: two birds of %s", c.target, tl.S.Family)
			}
			fams[tl.S.Family] = true
			got[tl.S.ID] = true
		}
		for fam, id := range best {
			if !got[id] {
				t.Errorf("%s: %s is the commonest of %s at %.4f but is not shown", c.target, id, fam, share[id])
			}
		}
		if len(p.Tiles) != len(best) {
			t.Errorf("%s: %d birds, want %d, one per family", c.target, len(p.Tiles), len(best))
		}
		if len(best) < 20 {
			t.Errorf("%s: only %d families qualify; the check proves little", c.target, len(best))
		}
	}
}

// TestNextDrawsOneScale checks that every bird /next shows is drawn at the
// same pixels per millimeter, the longest bird's, whatever the order.
//
// Verifies: product.nextView. Criterion: app.nextViewDrawsOneSet.
func TestNextDrawsOneScale(t *testing.T) {
	s := newTestServerReal(t)
	fixNextClock(t, time.October)
	for _, target := range []string{"/next", "/next?by=frequency", "/next?family=Ardeidae&family=Trochilidae", "/next?set=every&by=mass"} {
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
	p := nextAt(s, "/next?by=mass")
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

// TestNextAddsAndRemovesByName checks the add box: a name finds the bird,
// its link adds it to the set and drops the box, and the added bird's link
// removes it again and keeps everything else.
//
// Verifies: product.nextView. Criterion: app.nextViewDrawsOneSet.
func TestNextAddsAndRemovesByName(t *testing.T) {
	s := newTestServerReal(t)
	fixNextClock(t, time.October)
	p := nextAt(s, "/next?family=Ardeidae&find=sharp-shinned")
	var add string
	for _, m := range p.Matches {
		if m.Name == "Sharp-shinned Hawk" {
			add = string(m.AddURL)
		}
	}
	if add == "" {
		t.Fatalf("find=sharp-shinned lists %+v, without an add link for Sharp-shinned Hawk", p.Matches)
	}
	u, _ := url.Parse(add)
	if u.Path != "/next" || u.Query().Get("find") != "" || u.Query().Get("family") != "Ardeidae" {
		t.Errorf("add link %s: want /next keeping family and dropping find", add)
	}
	p = nextAt(s, add)
	var remove string
	for _, tl := range p.Tiles {
		if tl.S.Name(dataset.LocaleNA) == "Sharp-shinned Hawk" {
			remove = string(tl.RemoveURL)
		} else if tl.RemoveURL != "" {
			t.Errorf("%s was not added but has a remove link", tl.S.ID)
		}
	}
	if remove == "" {
		t.Fatalf("after %s, Sharp-shinned Hawk is not shown with a remove link", add)
	}
	after := nextAt(s, remove)
	for _, tl := range after.Tiles {
		if tl.S.Family != "Ardeidae" {
			t.Errorf("after removing, %s is still shown", tl.S.ID)
		}
	}
	if len(after.Tiles) != len(p.Tiles)-1 {
		t.Errorf("removing one bird left %d of %d", len(after.Tiles), len(p.Tiles))
	}
	shown := nextAt(s, "/next?family=Ardeidae&find=great+blue+heron")
	if len(shown.Matches) == 0 || !shown.Matches[0].Shown || shown.Matches[0].AddURL != "" {
		t.Errorf("a bird already shown is offered to add again: %+v", shown.Matches)
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

// TestNextShowAll checks the threshold: a set larger than nextShowAll shows
// its first birds and a link that shows them all, and that address does.
//
// Verifies: product.nextView. Criterion: app.nextViewDrawsOneSet.
func TestNextShowAll(t *testing.T) {
	s := newTestServerReal(t)
	fixNextClock(t, time.October)
	old := nextShowAll
	nextShowAll = 200 // below the dataset's 1,041, so the cap is exercised
	t.Cleanup(func() { nextShowAll = old })
	p := nextAt(s, "/next?set=every")
	if p.Count <= nextShowAll {
		t.Fatalf("set=every has %d birds; the check needs more than %d", p.Count, nextShowAll)
	}
	if len(p.Tiles) != nextShowAll || p.MoreURL == "" || p.Hidden != p.Count-nextShowAll {
		t.Errorf("set=every draws %d of %d, hides %d, more link %q", len(p.Tiles), p.Count, p.Hidden, p.MoreURL)
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
	if p := nextAt(s, "/next"); p.MoreURL != "" {
		t.Errorf("the landing set of %d offers Show all", p.Count)
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
