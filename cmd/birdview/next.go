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
	"slices"
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

// The home page (product.nextView): one set of birds, at one scale, in tiers.
// Built as a prototype at /next, which is why its names start with next.

// nextClock is the clock the home page reads for its default month. A variable so
// the checks can fix the month.
var nextClock = time.Now

// nextCap is how many birds the home page draws by default, the top of the order,
// before it offers to show the rest (product.nextView): a reader's first
// screens, not the whole set (Sameer, 2026-10-26). The address can change it
// (cap=50), or lift it (cap=none); the default is left out of addresses. Speed does not force it: measured
// 2026-10-26 with Lighthouse's mobile throttling (4x CPU, 100 Mbps, 20 ms),
// all 1,041 birds paint in 0.73 s (1.05 s at 6x), 229 KiB transferred. A
// variable so the checks can change it.
var nextCap = 100

// nextFindLimit is the most name matches the find box lists; above it, the
// box gives only their number (product.nextView). A variable so the checks
// can lower it below the fixture's size.
var nextFindLimit = 10

// nextChoice is one option inside a count-line control.
type nextChoice struct {
	Label string
	Title string
	On    bool
	URL   template.URL
	// Sep draws a rule before the choice, setting it apart from those before.
	Sep bool
	// Color and TextColor draw the choice as the family's badge, with its
	// description beside its name, as on the tiles.
	Color, TextColor string
	// Now marks Month's Now, whose month=now here.js replaces with the
	// browser's month.
	Now bool
}

// nextChoiceGroup is a labeled run of choices, such as the regions of one zone.
type nextChoiceGroup struct {
	Name    string
	Choices []nextChoice
	// OneLine keeps the choices on one line, scrolling sideways where the
	// screen is too narrow for them, as the twelve months are.
	OneLine bool
	// Lead is the choice that heads the group, set off by a rule: a habitat group's shortcut, or
	// All.
	Lead *nextChoice
	// NoRule draws no rule beside the lead: Habitat's All, which heads the column of group names.
	NoRule bool
	// Also is a second choice drawn beside the lead, before the rule: Month's Now beside its All.
	Also *nextChoice
}

// nextControl is one setting on the count line: its name, what it is set to,
// and the choices that open in place.
type nextControl struct {
	Label  string
	Value  string
	Note   string
	Groups []nextChoiceGroup
	// AllURL returns this setting to all; empty for a setting that cannot be all or that offers
	// All among its choices. AllOn marks it when the setting is all.
	AllURL template.URL
	AllOn  bool
	// Leads draws each group's Lead in a column, so the rule after it and the choices start
	// at the same place on every line; Habitat only.
	Leads bool
	// One marks a setting that takes exactly one choice, drawn as one joined row with All first
	// (product.chooseOneSettings).
	One bool
	// HereURL is this address with the region a placeholder for here.js to
	// fill, and Boundaries the file it looks the region up in; Region only.
	HereURL, Boundaries string
}

// nextPinned is one pinned bird in the list under the find box.
type nextPinned struct {
	Name, Scientific string
	Anchor           string
	URL              template.URL
}

// nextItem is one entry in the home page's grid: a bird tile, a gap row
// saying how many unpinned birds were omitted there, or a tier's heading.
type nextItem struct {
	Tile   *tile
	Hidden int
	Tier   *nextTier
	// Group is the Pinned group's heading: how many birds it holds (product.nextView).
	Group int
}

// nextMatch is one bird the find box found, with its pin.
type nextMatch struct {
	ID, Name, Scientific, Banding string
	Pinned                        bool
	URL                           template.URL
	// DetailURL is the bird's species page, so one bird's measurements need no pin
	// (product.nextView).
	DetailURL template.URL
}

type nextData struct {
	// Feedback opens a new issue about this page, filled in; empty with no tracker set
	// (product.reportDataIssue).
	Feedback template.URL
	Title    string
	Count    int
	Total    int
	Controls []nextControl
	Tiles    []tile
	// GroupTiles are the Pinned group's: the pinned birds again, in the
	// page's order, at the head of the page (product.nextView).
	GroupTiles []tile
	Items      []nextItem
	// Tiers are the tiers of the matching birds, in the order shown, and
	// Histogram one bar per tier (product.nextView).
	Tiers     []nextTier
	Histogram []nextBar
	// Show is which birds are drawn, the one setting show= carries: "first"
	// (the first Cap of the order, and pins past them; show=<n>, 100 the
	// default), "all" (show=all), or "pinned" (only the pinned birds,
	// show=pinned). ShowFields are the address's other settings, which the
	// Show first box's form sends with it, and ShowAllURL and ShowPinnedURL
	// are the other two choices. Matching is how many birds the settings
	// match, pins included, whatever is shown.
	Show          string
	Cap           string
	ShowFields    []hiddenField
	ShowAllURL    template.URL
	ShowPinnedURL template.URL
	ShowFirstURL  template.URL
	Matching      int
	// OrderLabel names the chosen order ("Length", "Bill length"), which the
	// sentence under the histogram leads with ("Length distribution of …").
	OrderLabel string
	// MoreURL shows every bird in the set (show=all); empty when all are
	// shown. First
	// is how many the filters chose that are drawn, Hidden how many are not,
	// and AddedPast how many pinned birds are drawn from past the first.
	MoreURL   template.URL
	Hidden    int
	First     int
	AddedPast int
	// PinnedOnly is set when only pinned birds are shown (show=pinned).
	PinnedOnly bool
	// HereNowURL is the welcome line's "the birds around you now": this
	// address with #here-and-now, which here.js
	// answers by going to HereNowTemplate with the reader's region and the
	// browser's month filled in. The template replaces the region and the
	// month and leaves Pinned only, since the birds around the reader are
	// not pins; every other parameter is kept.
	HereNowURL      template.URL
	HereNowTemplate string
	// Pinned lists the pinned birds in alphabetical order, each with its
	// pin and a link to its tile.
	Pinned []nextPinned
	// Find is the find box's text and Found how many birds it matches;
	// Matches lists them when there are nextFindLimit or fewer.
	Find    string
	Found   int
	Matches []nextMatch
	Fields  []hiddenField
	// ToggleURL is this address without find or bird, ready for more:
	// find=<text>&bird=<id> appended pins or unpins the bird and keeps the
	// list open on the text. CloseURL closes the list: this address without
	// find. PinnedIDs are the pinned birds, space-separated, and BirdsURL the
	// names the script searches.
	ToggleURL string
	CloseURL  template.URL
	PinnedIDs string
	BirdsURL  string
	// DetailSuffix is what follows /species/<id> in a species page's address from here: the
	// units and names, or nothing. The find script builds the list's links with it.
	DetailSuffix string
	FindLimit    int
	FindScript   template.JS
	HereScript   template.JS
	// ResetURL returns every setting to its default and keeps the pins;
	// ClearPinsURL drops the pins and keeps every setting, empty when
	// nothing is pinned.
	ResetURL     template.URL
	ClearPinsURL template.URL

	// included are the pinned birds.
	included map[string]bool
}

// nextURL is the home page with q.
func nextURL(q url.Values) template.URL {
	if len(q) == 0 {
		return "/"
	}
	return template.URL("/?" + q.Encode())
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
// returns to all. Every other parameter is kept: Show, the pins, and the
// find box's text included.
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
	return nextURL(q)
}

// leavePinnedOnly drops show=pinned and keeps any other show: for the two
// transitions that mean to leave the pinned birds, Clear pins (nothing is
// pinned after it) and the birds around you now (which are not pins).
func leavePinnedOnly(q url.Values) {
	if q.Get("show") == "pinned" {
		q.Del("show")
	}
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
		if d == "show" {
			leavePinnedOnly(q) // only show=pinned; a count or all is kept
			continue
		}
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
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	q := r.URL.Query()
	// A bird pinned by another of its codes is addressed by its own
	// (product.ebirdLink), so the page and its links name each bird one way.
	// A region named by a value BirdView no longer writes is addressed by its
	// current code (decisions.regionCodesFromData, tech.stableUrls).
	pins := s.canonicalPins(q)
	regions := canonicalRegions(q)
	if (pins || regions || canonicalFreq(q) || canonicalHabitats(q)) && !q.Has("bird") {
		http.Redirect(w, r, string(nextURL(q)), http.StatusMovedPermanently)
		return
	}
	if q.Has("bird") {
		http.Redirect(w, r, s.nextToggleBird(r), http.StatusSeeOther)
		return
	}
	// The First box sends show= whatever was typed: the default, a blank, or
	// something not a count all mean the default, which addresses leave out.
	if q.Has("show") {
		if _, _, ok := nextShowOf(q); !ok {
			q.Del("show")
			http.Redirect(w, r, string(nextURL(q)), http.StatusSeeOther)
			return
		}
	}
	// Now links to month=now, which here.js rewrites to the browser's month.
	// Without the script it lands here and goes to the server's month; the
	// answer reads the clock, so it is never stored (revalidated).
	if slices.Contains(q["month"], "now") {
		q.Set("month", month.Abbrev[nextClock().Month()-1])
		http.Redirect(w, r, string(nextURL(q)), http.StatusSeeOther)
		return
	}
	s.render(w, "next.html", s.nextPage(r))
}

// canonicalPins rewrites, in place, any pin= value that is one of a species'
// other codes to its id, and reports whether it rewrote one.
func (s *server) canonicalPins(q url.Values) bool {
	changed := false
	for i, v := range q["pin"] {
		if id, ok := s.aliasOf[v]; ok {
			q["pin"][i], changed = id, true
		}
	}
	return changed
}

// canonicalFreq rewrites, in place, common=, the reporting frequency's name until 2026-11-08, to
// freq=, unless freq= is already given, and reports whether it rewrote it
// (decisions.reportingFrequencyParam).
func canonicalFreq(q url.Values) bool {
	if !q.Has("common") {
		return false
	}
	if !q.Has("freq") {
		q["freq"] = q["common"]
	}
	q.Del("common")
	return true
}

// canonicalHabitats rewrites, in place, habitat=Marine, which the filter wrote before 2026-11-09, to
// its two refinements, each once, and reports whether it rewrote it (decisions.openOceanSplit).
func canonicalHabitats(q url.Values) bool {
	if !slices.Contains(q["habitat"], "Marine") {
		return false
	}
	var out []string
	for _, v := range q["habitat"] {
		vs := []string{v}
		if v == "Marine" {
			vs = []string{"Marine-inshore", "Marine-offshore"}
		}
		for _, w := range vs {
			if !slices.Contains(out, w) {
				out = append(out, w)
			}
		}
	}
	q["habitat"] = out
	return true
}

// canonicalRegions rewrites, in place, any region= value BirdView no longer writes to the region's
// current one, and reports whether it rewrote one.
func canonicalRegions(q url.Values) bool {
	changed := false
	for i, v := range q["region"] {
		if r, ok := region.Retired(v); ok {
			q["region"][i], changed = r.Slug(), true
		}
	}
	return changed
}

// nextShowOf reads show: "pinned", "all", or "first" with the count to draw
// before the gap rows. ok is false when show is absent or reads as the
// default, the first nextCap, so the address can leave it out.
func nextShowOf(q url.Values) (mode string, n int, ok bool) {
	v := strings.TrimSpace(q.Get("show"))
	if v == "pinned" || v == "all" {
		return v, nextCap, true
	}
	if c, err := strconv.Atoi(v); err == nil && c > 0 && c != nextCap {
		return "first", c, true
	}
	return "first", nextCap, false
}

// nextToggleBird is where bird=<id> goes: the same address with that bird
// pinned, or unpinned if it was, the find box's text kept. The find box's
// script links here, so the rule stays in Go.
func (s *server) nextToggleBird(r *http.Request) string {
	q := r.URL.Query()
	s.canonicalPins(q)
	id := q.Get("bird")
	if own, ok := s.aliasOf[id]; ok {
		id = own // another of its codes (product.ebirdLink)
	}
	q.Del("bird")
	if _, ok := s.byID[id]; !ok {
		q.Del("find")
		return string(nextURL(q))
	}
	return string(nextPin(q, id))
}

// nextTiers breaks the ordered birds into tiers (view.Sections and
// view.SectionsBy), taken in ascending order and reversed with the order. Within a tier the birds keep the order's tie-break.
func nextTiers(ordered []*dataset.Species, opt options, rev bool, share func(*dataset.Species) float64) []view.Section {
	asc := append([]*dataset.Species(nil), ordered...)
	if rev {
		slices.Reverse(asc)
	}
	var secs []view.Section
	if opt.Metric == view.ByFrequency {
		oneIn := func(sp *dataset.Species) float64 {
			if f := share(sp); f > 0 {
				return 1 / f
			}
			return math.Inf(1) // no records: one last tier
		}
		secs = view.SectionsBy(asc, oneIn, func(min float64) float64 { return view.Decade(min, opt.Decades) })
	} else {
		secs = view.Sections(asc, opt.Metric, opt.Step)
	}
	if rev {
		slices.Reverse(secs)
		for i := range secs {
			secs[i].Species = slices.Clone(secs[i].Species)
			slices.Reverse(secs[i].Species)
		}
	}
	return secs
}

// nextTier is one tier's heading: its number in the order shown, the range
// it covers in the page's units (or a note where there is none), how many
// of the matching birds it holds, and whether any of them is drawn.
type nextTier struct {
	Index    int
	Label    string
	Min, Max string
	Note     string
	Count    int
	Drawn    bool
}

// nextTierAnchor is a tier heading's id.
func nextTierAnchor(i int) string { return "tier-" + strconv.Itoa(i) }

// nextBar is one bar of the histogram: a tier, or an empty step of the
// scale between two tiers, with where it links: the tier's heading, on this
// page when it is drawn, or with every bird shown when it is not.
type nextBar struct {
	histBar
	URL template.URL
}

// nextDetailLink is a bird's species page, carrying only what the page
// draws itself with: the units and the names. Its Back returns by the
// browser's history, so the home page comes back as it was without its
// settings riding along (product.nextView).
func nextDetailLink(q url.Values, id string) template.URL {
	keep := url.Values{}
	for _, k := range []string{"names", "units"} {
		if v := q.Get(k); v != "" {
			keep.Set(k, v)
		}
	}
	if len(keep) == 0 {
		return template.URL("/species/" + id)
	}
	return template.URL("/species/" + id + "?" + keep.Encode())
}

// nextPin is the address with bird id pinned, or unpinned if it was. The
// find box keeps its text, so its list stays open for more pins until it is
// closed. It stays where it is on the page: the pinned list under the find
// box names the bird and links to its tile.
func nextPin(q url.Values, id string) template.URL {
	q = cloneValues(q)
	pins := nextValues(q, "pin")
	q.Del("pin")
	was := false
	for _, v := range pins {
		if v == id {
			was = true
			continue
		}
		q.Add("pin", v)
	}
	if was {
		return nextURL(q)
	}
	q.Add("pin", id)
	return nextURL(q)
}

// nextPage builds the home page for the address r carries.
func (s *server) nextPage(r *http.Request) nextData {
	q := r.URL.Query()
	opt := nextOptions(r)
	opt.Counties = s.counties
	// Every month unless the address names some (month=now has already
	// been redirected to a month by next).

	// show=pinned draws only the pinned birds: a comparison of the reader's
	// own picks. Otherwise the birds the settings match are drawn, the first
	// of them or all (show=all).
	mode, capN, _ := nextShowOf(q)
	pinnedOnly := mode == "pinned"
	matched, _ := s.filterSpecies(opt)
	var base []*dataset.Species
	if !pinnedOnly {
		base = matched
	}
	share := func(sp *dataset.Species) float64 {
		num, den := shareIn(sp, s.ds, opt)
		if den == 0 {
			return 0
		}
		return float64(num) / float64(den)
	}
	// Pinned birds join the set, once each.
	in := map[string]bool{}
	for _, sp := range base {
		in[sp.ID] = true
	}
	added := map[string]bool{}
	for _, id := range nextValues(q, "pin") {
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

	// Tiers (product.nextView, product.sectionBreakRule):
	// a size tier spans one step of the scale from its smallest bird, a
	// frequency tier a decade of "1 in N", and the birds with no records are
	// a tier of their own. Reversed, the tiers run the other way too.
	secs := nextTiers(base, opt, rev, share)
	base = base[:0]
	tierOf := map[*dataset.Species]int{}
	for i, sec := range secs {
		for _, sp := range sec.Species {
			base = append(base, sp)
			tierOf[sp] = i
		}
	}

	page := nextData{Count: len(base), Total: len(s.ds.Species),
		included: added, FindLimit: nextFindLimit,
		FindScript: s.nextFindJS}
	tq := cloneValues(q)
	tq.Del("find")
	tq.Del("bird")
	page.ToggleURL = "/?"
	if len(tq) > 0 {
		page.ToggleURL = "/?" + tq.Encode() + "&"
	}
	page.CloseURL = nextURL(tq)
	page.DetailSuffix = strings.TrimPrefix(string(nextDetailLink(q, "")), "/species/")
	page.BirdsURL = "/birds.json"
	if opt.Locale == dataset.LocaleWorld {
		page.BirdsURL += "?names=" + string(dataset.LocaleWorld)
	}
	// The first cap birds of the order, and every pinned bird past
	// them, with a gap row wherever unpinned birds are left out (including
	// after the last pinned bird).
	type pick struct {
		sp     *dataset.Species
		before int
	}
	var picks []pick
	after := 0
	page.Show, page.Cap = mode, strconv.Itoa(capN)
	for _, k := range sortedKeys(q) {
		if k == "show" || k == "bird" {
			continue
		}
		for _, v := range q[k] {
			page.ShowFields = append(page.ShowFields, hiddenField{k, v})
		}
	}
	page.ShowAllURL = nextSet(q, "show", "all")
	page.ShowPinnedURL = nextSet(q, "show", "pinned")
	page.ShowFirstURL = nextSet(q, "show", "")
	page.Matching = len(matched)
	inMatched := map[string]bool{}
	for _, sp := range matched {
		inMatched[sp.ID] = true
	}
	for id := range added {
		if !inMatched[id] {
			page.Matching++
		}
	}
	if mode != "first" {
		for _, sp := range base {
			picks = append(picks, pick{sp: sp})
		}
	} else {
		gap := 0
		for i, sp := range base {
			switch {
			case i < capN:
				picks = append(picks, pick{sp: sp})
			case added[sp.ID]:
				picks = append(picks, pick{sp: sp, before: gap})
				gap = 0
				page.AddedPast++
			default:
				gap++
				page.Hidden++
			}
		}
		after = gap
		if page.Hidden > 0 {
			page.MoreURL = page.ShowAllURL
			page.First = capN
		}
	}
	pinned := make([]string, 0, len(added))
	for id := range added {
		pinned = append(pinned, id)
	}
	sort.Strings(pinned)
	page.PinnedIDs = strings.Join(pinned, " ")
	keep := url.Values{}
	for _, id := range nextValues(q, "pin") {
		if added[id] && !slices.Contains(keep["pin"], id) {
			keep.Add("pin", id)
		}
	}
	page.ResetURL = nextURL(keep)
	for _, id := range keep["pin"] {
		sp := s.byID[id]
		page.Pinned = append(page.Pinned, nextPinned{Name: sp.Name(opt.Locale), Scientific: sp.ScientificName,
			Anchor: pinnedAnchor(id), URL: nextPin(q, id)})
	}
	sort.Slice(page.Pinned, func(i, j int) bool {
		li, lj := strings.ToLower(page.Pinned[i].Name), strings.ToLower(page.Pinned[j].Name)
		if li != lj {
			return li < lj
		}
		return page.Pinned[i].Name < page.Pinned[j].Name
	})
	if len(added) > 0 {
		// Showing only the pinned birds means nothing once none are.
		page.ClearPinsURL = nextSet(q, "pin", "", "show")
	}
	page.PinnedOnly = pinnedOnly

	// One scale: every bird shown is drawn to the longest and widest of them.
	shown := make([]*dataset.Species, 0, len(picks))
	masses := make([]float64, 0, len(picks))
	for _, pk := range picks {
		shown = append(shown, pk.sp)
		masses = append(masses, pk.sp.Mass.Value)
	}
	refLen, refSpan := view.Reference(shown)
	u := unitsFor(opt.Metric3, masses...)
	// The tiers' headings, in the page's units, with one precision for
	// all so that neighbors read apart (product.sectionHeadersStateRange).
	metricLabel := map[view.Metric]string{view.ByFrequency: "Frequency", view.ByLength: "Length",
		view.ByWingspan: "Wingspan", view.ByMass: "Mass", view.ByBill: "Bill", view.ByTail: "Tail"}[opt.Metric]
	page.Tiers = make([]nextTier, len(secs))
	spans := make([]tierSpan, len(secs))
	for i, sec := range secs {
		page.Tiers[i] = nextTier{Index: i + 1, Label: metricLabel, Count: len(sec.Species)}
		if opt.Metric == view.ByFrequency {
			page.Tiers[i].Min, page.Tiers[i].Max, page.Tiers[i].Note = frequencyBand(sec.Min, sec.Max, opt.Decades)
		}
		spans[i] = tierSpan{sec.Min, sec.Max, u}
	}
	if opt.Metric != view.ByFrequency {
		// Neighbors are neighbors on the scale, whichever way it is shown.
		asc := slices.Clone(spans)
		if rev {
			slices.Reverse(asc)
		}
		extra := headingDigits(asc, opt.Metric)
		for i, sp := range spans {
			page.Tiers[i].Min = sp.u.formatDigits(sp.lo, opt.Metric, extra)
			page.Tiers[i].Max = sp.u.formatDigits(sp.hi, opt.Metric, extra)
		}
	}
	for _, pk := range picks {
		page.Tiers[tierOf[pk.sp]].Drawn = true
	}
	// One bar per tier, with an empty bar for each step of the scale no
	// bird falls in, built in ascending order and then turned to the order
	// shown.
	views := make([]sectionView, len(secs))
	raw := slices.Clone(secs)
	for i, t := range page.Tiers {
		views[i] = sectionView{Index: t.Index, Min: t.Min, Max: t.Max, Note: t.Note, Count: t.Count}
	}
	if rev {
		slices.Reverse(views)
		slices.Reverse(raw)
	}
	bars := histogram(metricLabel, views, raw, opt.Metric, opt.Step)
	if rev {
		slices.Reverse(bars)
	}
	for _, b := range bars {
		nb := nextBar{histBar: b}
		if !b.Empty {
			nb.URL = template.URL("#" + nextTierAnchor(b.Index))
			if !page.Tiers[b.Index-1].Drawn {
				nb.URL = page.ShowAllURL + nb.URL
			}
		}
		page.Histogram = append(page.Histogram, nb)
	}

	// The Pinned group leads the page: every pinned bird, in the page's
	// order and untiered, before the first tier; under Pinned only it is the
	// page (product.nextView, decisions.pinnedGroupLeads). Its tiles have their own anchors, so a bird
	// drawn twice is still addressed once in each place.
	newPageTile := func(sp *dataset.Species, anchor string) tile {
		t := newTile(sp, view.Layout(sp, refLen, refSpan, tileBox), tileBox, u, opt.Locale)
		t.DetailURL, t.Anchor = nextDetailLink(q, sp.ID), anchor
		t.Commonest, t.SameAs = s.shareWithSeams(sp, opt), s.sameMeasurements(sp, opt.Locale)
		t.PinURL, t.Pinned = nextPin(q, sp.ID), added[sp.ID]
		return t
	}
	for _, pk := range picks {
		if added[pk.sp.ID] {
			page.GroupTiles = append(page.GroupTiles, newPageTile(pk.sp, pinnedAnchor(pk.sp.ID)))
		}
	}
	if len(page.GroupTiles) > 0 {
		page.Items = append(page.Items, nextItem{Group: len(page.GroupTiles)})
		for i := range page.GroupTiles {
			page.Items = append(page.Items, nextItem{Tile: &page.GroupTiles[i]})
		}
	}

	page.Tiles = make([]tile, len(picks))
	lastTier := -1
	for i, pk := range picks {
		page.Tiles[i] = newPageTile(pk.sp, tileAnchor(pk.sp.ID))
		if pinnedOnly {
			continue // the group is the page
		}
		if pk.before > 0 {
			page.Items = append(page.Items, nextItem{Hidden: pk.before})
		}
		if t := tierOf[pk.sp]; t != lastTier {
			lastTier = t
			page.Items = append(page.Items, nextItem{Tier: &page.Tiers[t]})
		}
		page.Items = append(page.Items, nextItem{Tile: &page.Tiles[i]})
	}
	if after > 0 && !pinnedOnly {
		page.Items = append(page.Items, nextItem{Hidden: after})
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
					Scientific: sp.ScientificName, Banding: sp.BandingCode, Pinned: added[sp.ID],
					URL:       template.URL(page.ToggleURL + "find=" + url.QueryEscape(page.Find) + "&bird=" + url.QueryEscape(sp.ID)),
					DetailURL: nextDetailLink(q, sp.ID)})
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

	page.Title = s.nextTitle(opt, pinnedOnly)
	page.Feedback = s.issueURL(r, "", "The page as shown", []string{page.Title,
		strconv.Itoa(page.Count) + " of " + strconv.Itoa(page.Total) + " birds match"})
	page.Controls = s.nextControls(q, opt)
	page.OrderLabel = nextOrderLabels[opt.Metric]
	if opt.Metric == view.ByBill || opt.Metric == view.ByTail {
		page.OrderLabel += " length" // "Bill length distribution", not "Bill distribution"
	}
	for i := range page.Controls {
		if page.Controls[i].Label == "Region" {
			hq := cloneValues(q)
			hq.Del("region")
			hq.Set("region", hereSlot)
			page.Controls[i].HereURL = strings.Replace(string(nextURL(hq)), url.QueryEscape(hereSlot), hereSlot, 1)
			page.Controls[i].Boundaries = s.boundaryURL
		}
	}
	hn := cloneValues(q)
	hn.Del("bird")
	page.HereNowURL = nextURL(hn) + "#here-and-now"
	hn.Del("region")
	leavePinnedOnly(hn)
	hn.Set("region", hereSlot)
	hn.Set("month", nowSlot)
	tmpl := string(nextURL(hn))
	tmpl = strings.Replace(tmpl, url.QueryEscape(hereSlot), hereSlot, 1)
	page.HereNowTemplate = strings.Replace(tmpl, url.QueryEscape(nowSlot), nowSlot, 1)
	page.HereScript = s.hereJS
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

// nextTitle says in words which birds are drawn: "Showing pinned birds", or
// "Showing birds in" the regions and months chosen, followed by any chosen
// family, habitat, and diet, each under the label of its setting.
func (s *server) nextTitle(opt options, pinnedOnly bool) string {
	if pinnedOnly {
		return "Showing pinned birds"
	}
	var b strings.Builder
	preset := presetOf(opt)
	if preset != nil {
		// A kind of bird is named in place of the settings it fills in (product.birdPresets).
		b.WriteString("Showing " + strings.ToLower(preset.Label))
	} else {
		b.WriteString("Showing birds")
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
	} else {
		b.WriteString(" in every month")
	}
	if opt.HasSelection && preset == nil {
		var fs []string
		for _, f := range s.fams {
			if opt.Families[f.Name] {
				fs = append(fs, f.Name)
			}
		}
		b.WriteString(" · Family: " + nextList(fs))
	}
	if opt.HasHabitats && preset == nil {
		b.WriteString(" · Habitat: " + nextList(habitatNames(opt.Habitats)))
	}
	if opt.HasNiches && preset == nil {
		b.WriteString(" · Diet: " + nextList(groupNames(opt.Niches, trait.DietGroups, trait.Niches)))
	}
	if opt.Common > 0 {
		b.WriteString(" · At least 1 in " + comma(int64(opt.Common)))
	}
	return b.String()
}

// nextList joins names with commas and "or", or counts them when there are
// too many to read.
func nextList(names []string) string {
	// Semicolons between items that hold commas themselves, as counties do ("Kings, NY"), so the
	// list's own separators stay visible (product.shareOfRecords).
	sep := ", "
	for _, n := range names {
		if strings.Contains(n, ",") {
			sep = "; "
		}
	}
	switch n := len(names); {
	case n == 0:
		return ""
	case n == 1:
		return names[0]
	case n == 2:
		return names[0] + " or " + names[1]
	case n <= 4:
		return strings.Join(names[:n-1], sep) + sep + "or " + names[n-1]
	default:
		return strings.Join(names[:3], sep) + sep + "or " + strconv.Itoa(n-3) + " more"
	}
}

func (s *server) regionNames(sel map[string]bool) []string {
	var out []string
	for _, z := range region.Zones {
		for _, rg := range region.InZone(z) {
			if sel[rg.ID] {
				out = append(out, rg.Name)
			}
			// A chosen county follows its state, named with the state's code.
			for _, c := range region.CountiesOf(rg.ID) {
				if sel[c.ID] {
					out = append(out, c.Label())
				}
			}
		}
	}
	return out
}

// countyControl is the County setting (product.nextView, decisions.countyListingRule): present when the regions chosen are one
// state with counties, or counties of one state. From the state, a county replaces it; among
// counties, a choice toggles; removing the last, or All, returns to the state.
func countyControl(q url.Values, opt options) (nextControl, bool) {
	if !opt.HasRegions || len(opt.Regions) == 0 {
		return nextControl{}, false
	}
	state := ""
	chosen := map[string]bool{}
	for id := range opt.Regions {
		rg, ok := region.Get(id)
		if !ok {
			return nextControl{}, false
		}
		switch {
		case rg.Kind == region.County && (state == "" || state == rg.Parent):
			state = rg.Parent
			chosen[id] = true
		case rg.Kind == region.Land && len(opt.Regions) == 1:
			state = id
		default:
			return nextControl{}, false
		}
	}
	counties := region.CountiesOf(state)
	if len(counties) == 0 {
		return nextControl{}, false
	}
	// The chosen counties in the order the address names them, so a choice adds at the end.
	var current []string
	for _, v := range nextValues(q, "region") {
		if rg, ok := region.BySlug(v); ok && chosen[rg.ID] {
			current = append(current, rg.Slug())
		}
	}
	st, _ := region.Get(state)
	c := nextControl{Label: "County", Value: "all"}
	var names []string
	grp := nextChoiceGroup{Name: "Counties in " + st.Name}
	for _, rg := range counties {
		on := chosen[rg.ID]
		if on {
			names = append(names, rg.Name)
		}
		var next []string
		for _, v := range current {
			if v != rg.Slug() {
				next = append(next, v)
			}
		}
		if !on {
			next = append(next, rg.Slug())
		}
		if len(next) == 0 {
			next = []string{st.Slug()}
		}
		nq := cloneValues(q)
		nq["region"] = next
		grp.Choices = append(grp.Choices, nextChoice{Label: rg.Name, Title: rg.Label(), On: on, URL: nextURL(nq)})
	}
	c.Groups = []nextChoiceGroup{grp}
	c.AllURL, c.AllOn = nextSet(q, "region", st.Slug()), len(names) == 0
	if len(names) > 0 {
		c.Value = nextList(names)
	}
	return c, true
}

// nextControls builds the count line: one control per setting.
func (s *server) nextControls(q url.Values, opt options) []nextControl {
	var out []nextControl
	out = append(out, birdsControl(q, opt), nextOrderControl(q, opt))

	// Months. With none named, every month is chosen; month=all, which
	// older addresses carry, is every month too.
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
			ms = append(ms, month.Abbrev[m])
		}
		mch = append(mch, nextChoice{Label: month.Abbrev[m], Title: month.Full[m], On: on,
			URL: nextToggle(mq, "month", month.Abbrev[m])})
	}
	if len(ms) > 0 {
		mc.Value = nextList(ms)
	}
	// All and Now lead the line, then a rule, then the twelve months, which wrap under themselves
	// on a narrow screen, as Habitat's and Diet's choices do (product.nextView).
	allMonths := nextChoice{Label: "All", Title: "Every month", On: !opt.HasMonths,
		URL: nextSet(q, "month", "")}
	// Now selects only the current month. Its address names month=now, which
	// here.js rewrites to the browser's month and the server, without the
	// script, redirects to its own; the page itself never reads the clock.
	now := nextChoice{Label: "Now", Title: "This month only", URL: nextSet(q, "month", "now"), Now: true}
	mc.Leads = true
	mc.Groups = []nextChoiceGroup{{Lead: &allMonths, Also: &now, Choices: mch}}
	out = append(out, mc)

	rc := nextControl{Label: "Region", Value: "all", AllURL: nextSet(q, "region", ""), AllOn: !opt.HasRegions}
	if opt.HasRegions {
		rc.Value = nextList(s.regionNames(opt.Regions))
	}
	for _, z := range region.Zones {
		grp := nextChoiceGroup{Name: z}
		for _, rg := range region.InZone(z) {
			grp.Choices = append(grp.Choices, nextChoice{Label: rg.Name, Title: rg.Code,
				On: opt.HasRegions && opt.Regions[rg.ID], URL: nextToggle(q, "region", rg.Slug())})
		}
		rc.Groups = append(rc.Groups, grp)
	}
	out = append(out, rc)
	if cc, ok := countyControl(q, opt); ok {
		out = append(out, cc)
	}
	// Reporting frequency: a threshold on the share each tile states, in the same steps as the
	// frequency tiers (product.reportingFrequencyFilter).
	hc := nextControl{Label: "Reporting frequency", Value: "all", One: true}
	hcAll := nextChoice{Label: "All", Title: "Every bird the other settings choose", On: opt.Common == 0, URL: nextSet(q, "freq", "")}
	hcs := []nextChoice{hcAll}
	for _, n := range []int{100, 1000, 10000} {
		v := strconv.Itoa(n)
		on := opt.Common == n
		if on {
			hc.Value = "1 in " + comma(int64(n))
		}
		hcs = append(hcs, nextChoice{Label: "1 in " + comma(int64(n)),
			Title: "Birds at least 1 in " + comma(int64(n)) + " of the observations in the regions and months chosen",
			On:    on, URL: nextSet(q, "freq", v)})
	}
	hc.Groups = []nextChoiceGroup{{Choices: hcs, OneLine: true}}
	out = append(out, hc)

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
			URL: nextToggle(q, "family", f.Name), Color: f.Color, TextColor: f.TextColor})
	}
	fc.AllURL, fc.AllOn = nextSet(q, "family", ""), !opt.HasSelection
	if opt.HasSelection {
		fc.Value = nextList(fnames)
	}
	out = append(out, fc)

	out = append(out, habitatCtl(q, opt), dietCtl(q, opt))

	uc := nextControl{Label: "Units", Value: "imperial", One: true}
	if opt.Metric3 {
		uc.Value = "metric"
	}
	uc.Groups = []nextChoiceGroup{{Choices: []nextChoice{
		{Label: "Imperial", Title: "Inches and ounces", On: !opt.Metric3, URL: nextSet(q, "units", "")},
		{Label: "Metric", Title: "Centimeters and grams", On: opt.Metric3, URL: nextSet(q, "units", "metric")},
	}}}
	out = append(out, uc)

	nc := nextControl{Label: "Names", Value: "North Amer", One: true}
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

// habitatCtl is the Habitat setting: a line per group of AVONET's habitats, the group's shortcut
// first and then each of its habitats, chosen on its own (decisions.habitatGroupShortcuts). A group
// of one habitat needs no shortcut, so the habitat itself leads its line.
func habitatCtl(q url.Values, opt options) nextControl {
	c := groupedCtl(q, "Habitat", "habitat", trait.HabitatGroups, trait.HabitatTerms, opt.Habitats, opt.HasHabitats)
	if opt.HasHabitats {
		c.Value = nextList(habitatNames(opt.Habitats))
	}
	return c
}

// dietCtl is the Diet setting, grouped by AVONET's trophic levels as Habitat is by its groups
// (product.habitatAndDietFilters).
func dietCtl(q url.Values, opt options) nextControl {
	c := groupedCtl(q, "Diet", "niche", trait.DietGroups, trait.Niches, opt.Niches, opt.HasNiches)
	if opt.HasNiches {
		c.Value = nextList(groupNames(opt.Niches, trait.DietGroups, trait.Niches))
	}
	return c
}

// groupedCtl draws a setting as All heading a column of group names, aligned with them and with no
// rule beside it, then a line per group: its shortcut, a rule, and its terms, each chosen on its own.
func groupedCtl(q url.Values, label, param string, groups []trait.TermGroup, vocab []trait.Term,
	sel map[string]bool, has bool) nextControl {
	c := nextControl{Label: label, Value: "all", Leads: true, Groups: []nextChoiceGroup{{NoRule: true,
		Lead: &nextChoice{Label: "All", On: !has, URL: nextSet(q, param, "")}}}}
	for _, g := range groups {
		all := has
		var ls []string
		for _, t := range g.Terms {
			all = all && sel[t]
			ls = append(ls, trait.LabelOf(vocab, t))
		}
		var ch []nextChoice
		for i, t := range g.Terms {
			ch = append(ch, nextChoice{Label: ls[i], Title: trait.DescriptionOf(vocab, t),
				On: has && sel[t], URL: nextToggle(q, param, t)})
		}
		grp := nextChoiceGroup{Lead: &nextChoice{Label: g.Label, Title: "All of " + nextList(ls), On: all,
			URL: groupToggle(q, param, g.Terms, all)}, Choices: ch}
		if len(ch) == 1 {
			grp = nextChoiceGroup{Lead: &ch[0]}
		}
		c.Groups = append(c.Groups, grp)
	}
	return c
}

// groupToggle is the address with a group's terms all removed when all are selected, and otherwise
// all added; from all, that selects only the group's.
func groupToggle(q url.Values, param string, terms []string, all bool) template.URL {
	q = cloneValues(q)
	var next []string
	for _, v := range nextValues(q, param) {
		if !slices.Contains(terms, v) {
			next = append(next, v)
		}
	}
	if !all {
		next = append(next, terms...)
	}
	q.Del(param)
	for _, v := range next {
		q.Add(param, v)
	}
	return nextURL(q)
}

// groupNames names the selected terms: a group by its name when all its terms are selected, and
// otherwise each selected term by its label.
func groupNames(sel map[string]bool, groups []trait.TermGroup, vocab []trait.Term) []string {
	var out []string
	for _, g := range groups {
		var on []string
		for _, t := range g.Terms {
			if sel[t] {
				on = append(on, trait.LabelOf(vocab, t))
			}
		}
		if len(on) == len(g.Terms) && len(on) > 1 {
			out = append(out, g.Label)
		} else {
			out = append(out, on...)
		}
	}
	return out
}

// habitatNames is groupNames for habitats, naming both sea refinements together "Open sea", AVONET's
// Marine, unless their whole group is chosen (decisions.openOceanSplit).
func habitatNames(sel map[string]bool) []string {
	names := groupNames(sel, trait.HabitatGroups, trait.HabitatTerms)
	in, off := trait.LabelOf(trait.HabitatTerms, "Marine-inshore"), trait.LabelOf(trait.HabitatTerms, "Marine-offshore")
	if !sel["Marine-inshore"] || !sel["Marine-offshore"] || !slices.Contains(names, in) {
		return names
	}
	var out []string
	for _, n := range names {
		switch n {
		case in:
			out = append(out, "Open sea")
		case off:
		default:
			out = append(out, n)
		}
	}
	return out
}

// nextBirds serves the names the home page's find box searches as the reader types
// (product.nextView): for every bird it draws, its id, English name in
// the locale asked for, scientific name, and search text, as searchTexts
// builds it. The matching stays the Go rule; the script only looks for the
// typed text, normalized, in the text given here.
func (s *server) nextBirds(w http.ResponseWriter, r *http.Request) {
	loc := dataset.LocaleNA
	if dataset.Locale(r.URL.Query().Get("names")) == dataset.LocaleWorld {
		loc = dataset.LocaleWorld
	}
	out := make([][5]string, 0, len(s.ds.Species))
	for i := range s.ds.Species {
		sp := &s.ds.Species[i]
		out = append(out, [5]string{sp.ID, sp.Name(loc), sp.ScientificName, s.search[sp.ID], sp.BandingCode})
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(out)
}

// pinnedAnchor is the fragment id of a pinned bird's tile in the Pinned group, apart from
// its tile in the tiers, which keeps tileAnchor (product.nextView).
func pinnedAnchor(id string) string { return "p-" + id }

// nextOrders are the orders Order by offers, in its order, with their names.
var nextOrders = []struct {
	m     view.Metric
	label string
}{
	{view.ByFrequency, "Frequency"}, {view.ByLength, "Length"}, {view.ByWingspan, "Wingspan"},
	{view.ByMass, "Mass"}, {view.ByBill, "Bill"}, {view.ByTail, "Tail"},
}

// nextOrderLabels names each order, as Order by does.
var nextOrderLabels = func() map[view.Metric]string {
	out := map[view.Metric]string{}
	for _, o := range nextOrders {
		out[o.m] = o.label
	}
	return out
}()

// nextOrderControl is the Order by setting: frequency, the default, and the
// five sizes.
func nextOrderControl(q url.Values, opt options) nextControl {
	oc := nextControl{Label: "Order by", One: true}
	rev := nextReversed(q)
	var och []nextChoice
	for _, m := range nextOrders {
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
			oc.Value = strings.ToLower(m.label) + " " + nextArrow(m.m, rev)
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
