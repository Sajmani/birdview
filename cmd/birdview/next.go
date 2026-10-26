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
	"encoding/json"
	"html/template"
	"math"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"birdview/internal/dataset"
	"birdview/internal/month"
	"birdview/internal/region"
	"birdview/internal/trait"
	"birdview/internal/view"
)

// The /next prototype (product.nextView): one set of birds, at one scale,
// with no sections or tiers, beside the index until simulated readers have
// compared the two.

// nextClock is the clock /next reads for its default month. A variable so
// the checks can fix the month.
var nextClock = time.Now

// nextShowAll is how many birds /next draws, the top of the order, before
// it offers to show the rest (product.nextView): a reader's first screens,
// not the whole set (Sameer, 2026-10-26). Speed does not force it: measured
// 2026-10-26 with Lighthouse's mobile throttling (4x CPU, 100 Mbps, 20 ms),
// all 1,041 birds paint in 0.73 s (1.05 s at 6x), 229 KiB transferred. A
// variable so the checks can change it.
var nextShowAll = 100

// nextFindLimit is the most name matches the find box lists; above it, the
// box gives only their number (product.nextView). A variable so the checks
// can lower it below the fixture's size.
var nextFindLimit = 8

// nextChoice is one option inside a count-line control.
type nextChoice struct {
	Label string
	Title string
	On    bool
	URL   template.URL
}

// nextChoiceGroup is a labeled run of choices, such as the regions of one zone.
type nextChoiceGroup struct {
	Name    string
	Choices []nextChoice
}

// nextControl is one setting on the count line: its name, what it is set to,
// and the choices that open in place.
type nextControl struct {
	Label  string
	Value  string
	Note   string
	Groups []nextChoiceGroup
	// AllURL returns this setting to all; empty when it already is.
	AllURL template.URL
}

// nextMatch is one bird the find box found, with the link that adds it to
// the set or removes it.
type nextMatch struct {
	ID, Name, Scientific string
	Shown                bool
	URL                  template.URL
}

type nextData struct {
	Feedback string
	Title    string
	Count    int
	Total    int
	Controls []nextControl
	Tiles    []tile
	// MoreURL shows every bird in the set; empty when all are shown.
	MoreURL template.URL
	Hidden  int
	// Find is the find box's text and Found how many birds it matches;
	// Matches lists them when there are nextFindLimit or fewer.
	Find    string
	Found   int
	Matches []nextMatch
	Fields  []hiddenField
	// ToggleURL is this address with bird= last and empty: a bird's id
	// appended adds it or removes it. ShownIDs are the birds in the set,
	// space-separated, and BirdsURL the names the script searches.
	ToggleURL  string
	ShownIDs   string
	BirdsURL   string
	FindLimit  int
	FindScript template.JS
	ResetURL   template.URL

	// in and base are the birds in the set, and in it before any added or
	// removed by name; for the bird= toggle.
	in, base map[string]bool
	included map[string]bool
}

// nextURL is /next with q.
func nextURL(q url.Values) template.URL {
	if len(q) == 0 {
		return "/next"
	}
	return template.URL("/next?" + q.Encode())
}

// nextValues is the non-empty values of key, in order.
func nextValues(q url.Values, key string) []string {
	var out []string
	for _, v := range q[key] {
		if v != "" {
			out = append(out, v)
		}
	}
	return out
}

// nextToggle is the address with value v of key toggled under
// product.nextView's rule: nothing selected is all, so choosing from all
// selects only v; choosing a selected value removes it, and removing the last
// returns to all. Changing which birds are shown also drops the find box and
// the Show all, which belonged to the set before.
func nextToggle(q url.Values, key, v string) template.URL {
	q = cloneValues(q)
	cur := nextValues(q, key)
	var next []string
	found := false
	for _, c := range cur {
		if c == v {
			found = true
			continue
		}
		next = append(next, c)
	}
	if !found {
		next = append(next, v)
	}
	q.Del(key)
	for _, c := range next {
		q.Add(key, c)
	}
	q.Del("find")
	q.Del("all")
	return nextURL(q)
}

// nextSet is the address with key set to v, or removed when v is empty.
func nextSet(q url.Values, key, v string, drop ...string) template.URL {
	q = cloneValues(q)
	if v == "" {
		q.Del(key)
	} else {
		q.Set(key, v)
	}
	for _, d := range drop {
		q.Del(d)
	}
	return nextURL(q)
}

func cloneValues(q url.Values) url.Values {
	out := make(url.Values, len(q))
	for k, v := range q {
		out[k] = append([]string(nil), v...)
	}
	return out
}

// nextOptions reads the address as product.nextView does: an empty filter
// is all, never nothing, so no address can empty the page by filtering.
func nextOptions(r *http.Request) options {
	opt := parseOptions(r)
	if len(opt.Families) == 0 {
		opt.HasSelection = false
	}
	if len(opt.Regions) == 0 {
		opt.HasRegions = false
	}
	if len(opt.Months) == 0 {
		opt.HasMonths = false
	}
	if len(opt.Habitats) == 0 {
		opt.HasHabitats = false
	}
	if len(opt.Niches) == 0 {
		opt.HasNiches = false
	}
	// Most often recorded first unless the address names another order.
	if r.URL.Query().Get("by") == "" {
		opt.Metric = view.ByFrequency
	}
	return opt
}

func (s *server) next(w http.ResponseWriter, r *http.Request) {
	if q := r.URL.Query(); q.Has("bird") {
		http.Redirect(w, r, s.nextToggleBird(r), http.StatusSeeOther)
		return
	}
	s.render(w, "next.html", s.nextPage(r))
}

// nextToggleBird is where bird=<id> goes: the same address with that bird
// added to the set, or removed from it if it was there, and the find box
// emptied. An added bird is listed with include=; a bird of the set removed
// is listed with exclude=. Undoing either deletes the entry rather than
// adding its opposite, so the address stays short.
func (s *server) nextToggleBird(r *http.Request) string {
	q := r.URL.Query()
	id := q.Get("bird")
	q.Del("bird")
	q.Del("find")
	if _, ok := s.byID[id]; !ok {
		return string(nextURL(q))
	}
	r2 := r.Clone(r.Context())
	r2.URL.RawQuery = q.Encode()
	page := s.nextPage(r2)
	without := func(key string) {
		keep := nextValues(q, key)
		q.Del(key)
		for _, v := range keep {
			if v != id {
				q.Add(key, v)
			}
		}
	}
	anchor := ""
	switch {
	case page.in[id] && page.included[id]:
		without("include")
	case page.in[id]:
		q.Add("exclude", id)
	default:
		without("exclude")
		if !page.base[id] {
			q.Add("include", id)
		}
		anchor = "#" + tileAnchor(id)
	}
	return string(nextURL(q)) + anchor
}

// nextPage builds /next for the address r carries.
func (s *server) nextPage(r *http.Request) nextData {
	q := r.URL.Query()
	opt := nextOptions(r)
	// The birds of this month, the server's own, unless the address names
	// months; month=all is every month.
	nowMonth := int(nextClock().Month()) - 1
	defaultMonth := len(nextValues(q, "month")) == 0
	if defaultMonth {
		opt.HasMonths = true
		opt.Months = map[int]bool{nowMonth: true}
	}

	base, _ := s.filterSpecies(opt)
	share := func(sp *dataset.Species) float64 {
		num, den := shareIn(sp, s.ds, opt)
		if den == 0 {
			return 0
		}
		return float64(num) / float64(den)
	}
	// Birds removed by name leave the set; birds added by name join it,
	// once each. A bird both added and removed is removed.
	baseIn := map[string]bool{}
	for _, sp := range base {
		baseIn[sp.ID] = true
	}
	removed := map[string]bool{}
	for _, id := range nextValues(q, "exclude") {
		removed[id] = true
	}
	in := map[string]bool{}
	kept := base[:0:0]
	for _, sp := range base {
		if !removed[sp.ID] {
			kept = append(kept, sp)
			in[sp.ID] = true
		}
	}
	base = kept
	added := map[string]bool{}
	for _, id := range nextValues(q, "include") {
		sp, ok := s.byID[id]
		if !ok || added[id] || removed[id] {
			continue
		}
		added[id] = true
		if !in[id] {
			in[id] = true
			base = append(base, sp)
		}
	}

	// Order. Sizes smallest first; frequency most often recorded first;
	// rev=1 reverses either.
	rev := nextReversed(q)
	key := map[*dataset.Species]float64{}
	for _, sp := range base {
		if opt.Metric == view.ByFrequency {
			key[sp] = -share(sp)
		} else {
			key[sp] = opt.Metric.Value(sp)
		}
	}
	sort.SliceStable(base, func(i, j int) bool {
		if key[base[i]] != key[base[j]] {
			return (key[base[i]] < key[base[j]]) != rev
		}
		return base[i].ScientificName < base[j].ScientificName
	})

	page := nextData{Feedback: s.feedback, Count: len(base), Total: len(s.ds.Species),
		ResetURL: "/next", in: in, base: baseIn, included: added, FindLimit: nextFindLimit,
		FindScript: s.nextFindJS}
	tq := cloneValues(q)
	tq.Del("find")
	tq.Del("bird")
	page.ToggleURL = "/next?bird="
	if len(tq) > 0 {
		page.ToggleURL = "/next?" + tq.Encode() + "&bird="
	}
	page.BirdsURL = "/next/birds.json"
	if opt.Locale == dataset.LocaleWorld {
		page.BirdsURL += "?names=" + string(dataset.LocaleWorld)
	}
	ids := make([]string, 0, len(base))
	for _, sp := range base {
		ids = append(ids, sp.ID)
	}
	sort.Strings(ids)
	page.ShownIDs = strings.Join(ids, " ")
	toggle := func(id string) template.URL { return template.URL(page.ToggleURL + url.QueryEscape(id)) }
	shown := base
	if len(shown) > nextShowAll && q.Get("all") != "1" {
		shown = shown[:nextShowAll]
		page.Hidden = len(base) - nextShowAll
		page.MoreURL = nextSet(q, "all", "1")
	}

	// One scale: every bird shown is drawn to the longest and widest of them.
	var refLen, refSpan float64
	masses := make([]float64, 0, len(shown))
	for _, sp := range shown {
		refLen = math.Max(refLen, sp.TotalLength.Value)
		refSpan = math.Max(refSpan, sp.Wingspan.Value)
		masses = append(masses, sp.Mass.Value)
	}
	u := unitsFor(opt.Metric3, masses...)
	for _, sp := range shown {
		g := view.Layout(sp, refLen, refSpan, tileBox)
		t := tile{
			S: sp, Name: sp.Name(opt.Locale), Geom: g, Box: tileBox,
			TopOffset:   (tileBox.H - g.TotalHeight()) / 2,
			DetailURL:   detailLink(r, sp.ID),
			Anchor:      tileAnchor(sp.ID),
			MassCube:    newMassCube(sp, tileBox.H, g.PxPerMm, ""),
			WingModeled: sp.Wingspan.Prov.Kind == dataset.Modeled,
			Length:      u.format(sp.TotalLength.Value, view.ByLength),
			Wing:        u.format(sp.Wingspan.Value, view.ByWingspan),
			MassStr:     u.format(sp.Mass.Value, view.ByMass),
			Commonest:   s.shareWithSeams(sp, opt),
			SameAs:      s.sameMeasurements(sp, opt.Locale),
		}
		if added[sp.ID] {
			t.RemoveURL = toggle(sp.ID)
		}
		page.Tiles = append(page.Tiles, t)
	}

	// The find box: how many birds match, and which once few enough.
	page.Find = strings.TrimSpace(q.Get("find"))
	if normalizeSearch(page.Find) != "" {
		var found []*dataset.Species
		for i := range s.ds.Species {
			if sp := &s.ds.Species[i]; searchMatches(s.search[sp.ID], page.Find) {
				found = append(found, sp)
			}
		}
		page.Found = len(found)
		if len(found) <= nextFindLimit {
			for _, sp := range found {
				page.Matches = append(page.Matches, nextMatch{ID: sp.ID, Name: sp.Name(opt.Locale),
					Scientific: sp.ScientificName, Shown: in[sp.ID], URL: toggle(sp.ID)})
			}
		}
	}
	for _, k := range sortedKeys(q) {
		if k == "find" {
			continue
		}
		for _, v := range q[k] {
			page.Fields = append(page.Fields, hiddenField{k, v})
		}
	}

	page.Title = s.nextTitle(opt)
	page.Controls = s.nextControls(q, opt, defaultMonth, nowMonth)
	return page
}

func sortedKeys(q url.Values) []string {
	keys := make([]string, 0, len(q))
	for k := range q {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// nextTitle says in words which birds these are.
func (s *server) nextTitle(opt options) string {
	var b strings.Builder
	b.WriteString("Birds")
	if opt.HasRegions {
		b.WriteString(" in " + nextList(s.regionNames(opt.Regions)))
	}
	if opt.HasMonths {
		var ms []string
		for m := 0; m < month.Count; m++ {
			if opt.Months[m] {
				ms = append(ms, month.Full[m])
			}
		}
		b.WriteString(" in " + nextList(ms))
	} else {
		b.WriteString(" in every month")
	}
	return b.String()
}

// nextList joins names with commas and "or", or counts them when there are
// too many to read.
func nextList(names []string) string {
	switch n := len(names); {
	case n == 0:
		return ""
	case n == 1:
		return names[0]
	case n == 2:
		return names[0] + " or " + names[1]
	case n <= 4:
		return strings.Join(names[:n-1], ", ") + ", or " + names[n-1]
	default:
		return strings.Join(names[:3], ", ") + ", or " + strconv.Itoa(n-3) + " more"
	}
}

func (s *server) regionNames(sel map[string]bool) []string {
	var out []string
	for _, g := range region.Groupings[:1] {
		for _, z := range g.Zones {
			for _, rg := range region.InZone(g.Name, z) {
				if sel[rg.ID] {
					out = append(out, rg.Name)
				}
			}
		}
	}
	return out
}

// nextControls builds the count line: one control per setting.
func (s *server) nextControls(q url.Values, opt options, defaultMonth bool, nowMonth int) []nextControl {
	var out []nextControl
	out = append(out, nextOrderControl(q, opt))

	// Months. With none named, this month is chosen, and choosing another
	// month selects only that one. "All" is month=all, so that it can be
	// told apart from naming none.
	mc := nextControl{Label: "Month", Value: "all"}
	mq := cloneValues(q)
	mq.Del("month")
	for _, v := range nextValues(q, "month") {
		if v != "all" {
			mq.Add("month", v)
		}
	}
	var ms []string
	var mch []nextChoice
	for m := 0; m < month.Count; m++ {
		on := opt.HasMonths && opt.Months[m]
		if on {
			ms = append(ms, month.Full[m])
		}
		base := mq
		if defaultMonth {
			base = q
		}
		mch = append(mch, nextChoice{Label: month.Abbrev[m], Title: month.Full[m], On: on && !defaultMonth,
			URL: nextToggle(base, "month", month.Abbrev[m])})
	}
	if len(ms) > 0 {
		mc.Value = nextList(ms)
	}
	if defaultMonth {
		mc.Value = month.Full[nowMonth] + " (now)"
		mch[nowMonth].On = true
	}
	if opt.HasMonths {
		mc.AllURL = nextSet(q, "month", "all", "all", "find")
	}
	mc.Groups = []nextChoiceGroup{{Choices: mch}}
	out = append(out, mc)

	rc := nextControl{Label: "Region", Value: "everywhere"}
	if opt.HasRegions {
		rc.Value = nextList(s.regionNames(opt.Regions))
		rc.AllURL = nextSet(q, "region", "", "all", "find")
	}
	g := region.Groupings[0]
	for _, z := range g.Zones {
		grp := nextChoiceGroup{Name: z}
		for _, rg := range region.InZone(g.Name, z) {
			grp.Choices = append(grp.Choices, nextChoice{Label: rg.Name, Title: rg.Code,
				On: opt.HasRegions && opt.Regions[rg.ID], URL: nextToggle(q, "region", rg.Slug())})
		}
		rc.Groups = append(rc.Groups, grp)
	}
	out = append(out, rc)

	fc := nextControl{Label: "Family", Value: "all"}
	var fnames []string
	var order string
	for _, f := range s.fams {
		if f.Order != order || len(fc.Groups) == 0 {
			order = f.Order
			name := f.Order
			if f.OrderCommon != "" {
				name += " (" + f.OrderCommon + ")"
			}
			fc.Groups = append(fc.Groups, nextChoiceGroup{Name: name})
		}
		on := opt.HasSelection && opt.Families[f.Name]
		if on {
			fnames = append(fnames, f.Name)
		}
		last := &fc.Groups[len(fc.Groups)-1]
		last.Choices = append(last.Choices, nextChoice{Label: f.Name, Title: f.Common, On: on,
			URL: nextToggle(q, "family", f.Name)})
	}
	if opt.HasSelection {
		fc.Value = nextList(fnames)
		fc.AllURL = nextSet(q, "family", "", "all", "find")
	}
	out = append(out, fc)

	traitCtl := func(label, param string, terms []trait.Term, sel map[string]bool, has bool) nextControl {
		c := nextControl{Label: label, Value: "all"}
		var names []string
		var ch []nextChoice
		for _, t := range terms {
			on := has && sel[t.Value]
			if on {
				names = append(names, t.Label)
			}
			ch = append(ch, nextChoice{Label: t.Label, Title: t.Description, On: on,
				URL: nextToggle(q, param, t.Value)})
		}
		if has {
			c.Value = nextList(names)
			c.AllURL = nextSet(q, param, "", "all", "find")
		}
		c.Groups = []nextChoiceGroup{{Choices: ch}}
		return c
	}
	out = append(out,
		traitCtl("Habitat", "habitat", trait.Habitats, opt.Habitats, opt.HasHabitats),
		traitCtl("Diet", "niche", trait.Niches, opt.Niches, opt.HasNiches))

	uc := nextControl{Label: "Units", Value: "inches and ounces"}
	if opt.Metric3 {
		uc.Value = "centimeters and grams"
	}
	uc.Groups = []nextChoiceGroup{{Choices: []nextChoice{
		{Label: "Inches and ounces", On: !opt.Metric3, URL: nextSet(q, "units", "")},
		{Label: "Centimeters and grams", On: opt.Metric3, URL: nextSet(q, "units", "metric")},
	}}}
	out = append(out, uc)

	nc := nextControl{Label: "Names", Value: "North American"}
	world := opt.Locale == dataset.LocaleWorld
	if world {
		nc.Value = "world"
	}
	nc.Groups = []nextChoiceGroup{{Choices: []nextChoice{
		{Label: "North American", Title: "As eBird has them", On: !world, URL: nextSet(q, "names", "")},
		{Label: "World", Title: "As the IOC and AviList have them", On: world, URL: nextSet(q, "names", string(dataset.LocaleWorld))},
	}}}
	out = append(out, nc)
	return out
}

// nextBirds serves the names /next's find box searches as the reader types
// (product.nextView): for every bird of the index, its id, English name in
// the locale asked for, scientific name, and search text, as searchTexts
// builds it. The matching stays the Go rule; the script only looks for the
// typed text, normalized, in the text given here.
func (s *server) nextBirds(w http.ResponseWriter, r *http.Request) {
	loc := dataset.LocaleNA
	if dataset.Locale(r.URL.Query().Get("names")) == dataset.LocaleWorld {
		loc = dataset.LocaleWorld
	}
	out := make([][4]string, 0, len(s.ds.Species))
	for i := range s.ds.Species {
		sp := &s.ds.Species[i]
		out = append(out, [4]string{sp.ID, sp.Name(loc), sp.ScientificName, s.search[sp.ID]})
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(out)
}

// nextOrderControl is the Order by setting: frequency, the default, and the
// five sizes.
func nextOrderControl(q url.Values, opt options) nextControl {
	oc := nextControl{Label: "Order by"}
	rev := nextReversed(q)
	var och []nextChoice
	for _, m := range []struct {
		m     view.Metric
		label string
	}{
		{view.ByFrequency, "Frequency"}, {view.ByLength, "Length"}, {view.ByWingspan, "Wingspan"},
		{view.ByMass, "Mass"}, {view.ByBill, "Bill"}, {view.ByTail, "Tail"},
	} {
		v := string(m.m)
		if m.m == view.ByFrequency {
			v = ""
		}
		c := nextChoice{Label: m.label, URL: nextSet(q, "by", v, "rev"),
			Title: nextDirection(m.m, false)}
		if opt.Metric == m.m {
			// The chosen order is a toggle of its direction.
			c.On = true
			c.Label += " " + nextArrow(m.m, rev)
			c.Title = nextDirection(m.m, rev) + "; choose again for " + nextDirection(m.m, !rev)
			flip := "1"
			if rev {
				flip = ""
			}
			c.URL = nextSet(q, "rev", flip)
			oc.Value = strings.ToLower(m.label) + ", " + nextDirection(m.m, rev)
		}
		och = append(och, c)
	}
	oc.Note = "Choose the order again to reverse it."
	oc.Groups = []nextChoiceGroup{{Choices: och}}
	return oc
}

// nextReversed reports whether the address reverses the order.
func nextReversed(q url.Values) bool { return q.Get("rev") == "1" }

// nextDirection says which birds come first.
func nextDirection(m view.Metric, rev bool) string {
	if m == view.ByFrequency {
		if rev {
			return "least often recorded first"
		}
		return "most often recorded first"
	}
	if rev {
		return "largest first"
	}
	return "smallest first"
}

// nextArrow says which way the chosen order runs: down when the numbers fall
// down the page (largest or most often recorded first), up when they rise.
func nextArrow(m view.Metric, rev bool) string {
	if (m == view.ByFrequency) != rev {
		return "\u2193"
	}
	return "\u2191"
}
