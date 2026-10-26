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

// nextShowAll is how many birds /next draws before it offers to show the
// rest (product.nextView): the most that paint in about a second on a
// phone-class CPU over 5G. Measured 2026-10-26 with Lighthouse's mobile
// throttling (4x CPU, 100 Mbps, 20 ms): about 0.6 ms a bird, 1,041 birds in
// 0.73 s (1.05 s at 6x), 229 KiB transferred. A variable so the checks can
// lower it below the dataset's size.
var nextShowAll = 1200

// nextCommonShare is the landing rule's floor: a bird is a candidate when it
// is at least 1 in 1,000 of the records in the month and regions selected.
const nextCommonShare = 1e-3

// nextFindLimit is how many name matches the add box lists.
const nextFindLimit = 12

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

// nextMatch is one bird the add box found.
type nextMatch struct {
	Name, Scientific string
	AddURL           template.URL
	Shown            bool
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
	// Find is the add box's text; Matches what it found, and FindMore how
	// many more it found than are listed.
	Find     string
	Matches  []nextMatch
	FindMore int
	Fields   []hiddenField
	ResetURL template.URL
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
	return opt
}

// nextCommon reports whether the address asks for the commonest bird of each
// family rather than every bird. It does by default unless a family,
// habitat, or diet is chosen, which ask for every bird of that kind.
func nextCommon(q url.Values, opt options) bool {
	switch q.Get("set") {
	case "common":
		return true
	case "every":
		return false
	}
	return !opt.HasSelection && !opt.HasHabitats && !opt.HasNiches
}

func (s *server) next(w http.ResponseWriter, r *http.Request) {
	s.render(w, "next.html", s.nextPage(r))
}

// nextPage builds /next for the address r carries.
func (s *server) nextPage(r *http.Request) nextData {
	q := r.URL.Query()
	opt := nextOptions(r)
	common := nextCommon(q, opt)

	// The commonest-of-each-family set is of one month: the server's own
	// unless the address names some.
	nowMonth := int(nextClock().Month()) - 1
	defaultMonth := common && !opt.HasMonths
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
	if common {
		best := map[string]*dataset.Species{}
		bestShare := map[string]float64{}
		var order []string
		for _, sp := range base {
			f := share(sp)
			if f < nextCommonShare {
				continue
			}
			if _, ok := best[sp.Family]; !ok {
				order = append(order, sp.Family)
			}
			if f > bestShare[sp.Family] {
				best[sp.Family], bestShare[sp.Family] = sp, f
			}
		}
		base = base[:0:0]
		for _, fam := range order {
			base = append(base, best[fam])
		}
	}

	// Birds added by name join the set, once each.
	in := map[string]bool{}
	for _, sp := range base {
		in[sp.ID] = true
	}
	added := map[string]bool{}
	for _, id := range nextValues(q, "include") {
		sp, ok := s.byID[id]
		if !ok || added[id] {
			continue
		}
		added[id] = true
		if !in[id] {
			in[id] = true
			base = append(base, sp)
		}
	}

	// Order. Sizes smallest first; frequency most often recorded first.
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
			return key[base[i]] < key[base[j]]
		}
		return base[i].ScientificName < base[j].ScientificName
	})

	page := nextData{Feedback: s.feedback, Count: len(base), Total: len(s.ds.Species),
		ResetURL: "/next"}
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
			t.RemoveURL = nextRemove(q, sp.ID)
		}
		page.Tiles = append(page.Tiles, t)
	}

	// The add box.
	page.Find = strings.TrimSpace(q.Get("find"))
	if normalizeSearch(page.Find) != "" {
		for i := range s.ds.Species {
			sp := &s.ds.Species[i]
			if !searchMatches(s.search[sp.ID], page.Find) {
				continue
			}
			if len(page.Matches) == nextFindLimit {
				page.FindMore++
				continue
			}
			m := nextMatch{Name: sp.Name(opt.Locale), Scientific: sp.ScientificName, Shown: in[sp.ID]}
			if !m.Shown {
				aq := cloneValues(q)
				aq.Add("include", sp.ID)
				aq.Del("find")
				m.AddURL = nextURL(aq)
			}
			page.Matches = append(page.Matches, m)
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

	page.Title = s.nextTitle(opt, common, defaultMonth)
	page.Controls = s.nextControls(q, opt, common, defaultMonth, nowMonth)
	return page
}

// nextRemove is the address without one added bird.
func nextRemove(q url.Values, id string) template.URL {
	q = cloneValues(q)
	keep := q["include"]
	q.Del("include")
	for _, v := range keep {
		if v != id && v != "" {
			q.Add("include", v)
		}
	}
	return nextURL(q)
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
func (s *server) nextTitle(opt options, common, defaultMonth bool) string {
	var b strings.Builder
	if common {
		b.WriteString("The most often recorded bird of each family")
	} else {
		b.WriteString("Every bird")
	}
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
func (s *server) nextControls(q url.Values, opt options, common, defaultMonth bool, nowMonth int) []nextControl {
	var out []nextControl

	set := nextControl{Label: "Show"}
	set.Value = "every bird"
	if common {
		set.Value = "commonest of each family"
	}
	set.Groups = []nextChoiceGroup{{Choices: []nextChoice{
		{Label: "Commonest of each family", Title: "The most often recorded bird of each family, of those at 1 in 1,000 or more of the records",
			On: common, URL: nextSet(q, "set", "common", "all", "find")},
		{Label: "Every bird", On: !common, URL: nextSet(q, "set", "every", "all", "find")},
	}}}
	out = append(out, set)

	// Months. The default month is shown as chosen, but choosing from it
	// works as choosing from all: it selects only the month chosen.
	mc := nextControl{Label: "Month", Value: "all"}
	var ms []string
	var mch []nextChoice
	for m := 0; m < month.Count; m++ {
		on := opt.HasMonths && opt.Months[m]
		if on {
			ms = append(ms, month.Full[m])
		}
		mch = append(mch, nextChoice{Label: month.Abbrev[m], Title: month.Full[m], On: on,
			URL: nextToggle(q, "month", month.Abbrev[m])})
	}
	if len(ms) > 0 {
		mc.Value = nextList(ms)
	}
	if defaultMonth {
		mc.Value = month.Full[nowMonth] + " (now)"
		mc.Note = "Choose a month to see its birds instead of this month's."
	}
	if len(nextValues(q, "month")) > 0 {
		mc.AllURL = nextSet(q, "month", "", "all", "find")
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

	oc := nextControl{Label: "Order by"}
	var och []nextChoice
	for _, m := range []struct {
		m     view.Metric
		label string
	}{
		{view.ByLength, "Length"}, {view.ByWingspan, "Wingspan"}, {view.ByMass, "Mass"},
		{view.ByBill, "Bill"}, {view.ByTail, "Tail"}, {view.ByFrequency, "Frequency"},
	} {
		v := string(m.m)
		if m.m == view.ByLength {
			v = ""
		}
		if opt.Metric == m.m {
			oc.Value = strings.ToLower(m.label)
		}
		och = append(och, nextChoice{Label: m.label, On: opt.Metric == m.m, URL: nextSet(q, "by", v)})
	}
	oc.Note = "Sizes run smallest first; frequency runs most often recorded first."
	oc.Groups = []nextChoiceGroup{{Choices: och}}
	out = append(out, oc)

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
