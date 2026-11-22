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
	"net/http"
	"net/url"
	"sort"
	"strings"

	"birdview/internal/region"
)

// Place search and the unified Place setting (product.nextView,
// decisions.placeControlUnifiesRegionsAndCounties). The server serves each
// place's normalized search fields in /places.json, and next-place.js applies
// the exact same matching, ranking, and state/county replacement rules as
// placeMatches and placeToggle here. TestPlaceSearchInTheBrowser holds the
// script to the server.

// nextPlaceLimit is the most matches the Place box lists at once. Broad queries
// such as "washington" (33 matches, including the state) always show the top
// nextPlaceLimit rather than hiding everything above a cutoff. A variable so
// tests can inspect or adjust it.
var nextPlaceLimit = 12

// foldLatin folds common accented Latin letters to their ASCII base so
// "quebec" matches "Québec" and "montreal" matches "Montréal". next-place.js
// uses Unicode NFD stripping for the same effect.
func foldLatin(r rune) rune {
	switch r {
	case 'à', 'á', 'â', 'ã', 'ä', 'å':
		return 'a'
	case 'ç':
		return 'c'
	case 'è', 'é', 'ê', 'ë':
		return 'e'
	case 'ì', 'í', 'î', 'ï':
		return 'i'
	case 'ñ':
		return 'n'
	case 'ò', 'ó', 'ô', 'õ', 'ö':
		return 'o'
	case 'ù', 'ú', 'û', 'ü':
		return 'u'
	case 'ý', 'ÿ':
		return 'y'
	}
	return r
}

// normalizePlace folds diacritics, lowercases s, and turns every run of
// non-alphanumeric characters into one space. next-place.js does the same.
func normalizePlace(s string) string {
	var b strings.Builder
	space := true
	for _, r := range strings.ToLower(s) {
		r = foldLatin(r)
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
			space = false
		} else if !space {
			b.WriteByte(' ')
			space = true
		}
	}
	return strings.TrimSpace(b.String())
}

// placeEntry is one searchable place: a state, province, territory, outlying
// island, sea, or (under -counties) county or census division.
type placeEntry struct {
	Slug      string
	Label     string
	Parent    string // parent state/province slug for counties; empty otherwise
	County    bool
	CodeNorm  string
	NameNorm  string
	LabelNorm string
	FullNorm  string
}

// placeEntries lists every filterable place in deterministic order: states,
// provinces, islands, and seas in zone order, followed by counties when
// -counties is enabled.
func (s *server) placeEntries() []placeEntry {
	var out []placeEntry
	for _, z := range region.Zones {
		for _, rg := range region.InZone(z) {
			nameNorm := normalizePlace(rg.Name)
			codeNorm := normalizePlace(rg.Code)
			full := nameNorm
			if codeNorm != "" {
				full += " " + codeNorm
			}
			out = append(out, placeEntry{
				Slug:      rg.Slug(),
				Label:     rg.Name,
				CodeNorm:  codeNorm,
				NameNorm:  nameNorm,
				LabelNorm: nameNorm,
				FullNorm:  full,
			})
		}
	}
	if s.counties != nil {
		for _, z := range region.Zones {
			for _, st := range region.InZone(z) {
				stName := normalizePlace(st.Name)
				stCode := normalizePlace(st.Code)
				for _, ct := range region.CountiesOf(st.ID) {
					nameNorm := normalizePlace(ct.Name)
					label := ct.Label()
					labelNorm := normalizePlace(label)
					full := nameNorm
					if stName != "" {
						full += " " + stName
					}
					if stCode != "" {
						full += " " + stCode
					}
					out = append(out, placeEntry{
						Slug:      ct.Slug(),
						Label:     label,
						Parent:    st.Slug(),
						County:    true,
						NameNorm:  nameNorm,
						LabelNorm: labelNorm,
						FullNorm:  full,
					})
				}
			}
		}
	}
	return out
}

// placeRank scores how well e matches the normalized query q and its
// whitespace-split tokens, returning (rank, true) if e matches. Lower rank is
// better: 0 exact, 1 prefix, 2 all-token word-prefix, 3 substring.
func placeRank(e placeEntry, q string, toks []string) (int, bool) {
	if e.NameNorm == q || (e.CodeNorm != "" && e.CodeNorm == q) || e.LabelNorm == q {
		return 0, true
	}
	if strings.HasPrefix(e.NameNorm, q) || strings.HasPrefix(e.LabelNorm, q) {
		return 1, true
	}
	padded := " " + e.FullNorm
	allWords := true
	for _, t := range toks {
		if !strings.Contains(padded, " "+t) {
			allWords = false
			break
		}
	}
	if allWords {
		return 2, true
	}
	if strings.Contains(e.FullNorm, q) {
		return 3, true
	}
	return 0, false
}

// placeMatches returns every entry matching query, ranked with states,
// provinces, islands, and seas ahead of counties, then by match tier (exact,
// prefix, word-prefix, substring), then alphabetically by label.
func placeMatches(entries []placeEntry, query string) []placeEntry {
	q := normalizePlace(query)
	if q == "" {
		return nil
	}
	toks := strings.Fields(q)
	type scored struct {
		entry placeEntry
		rank  int
	}
	var hits []scored
	for _, e := range entries {
		if r, ok := placeRank(e, q, toks); ok {
			hits = append(hits, scored{e, r})
		}
	}
	sort.Slice(hits, func(i, j int) bool {
		if hits[i].entry.County != hits[j].entry.County {
			return !hits[i].entry.County
		}
		if hits[i].rank != hits[j].rank {
			return hits[i].rank < hits[j].rank
		}
		return hits[i].entry.Label < hits[j].entry.Label
	})
	out := make([]placeEntry, len(hits))
	for i, h := range hits {
		out[i] = h.entry
	}
	return out
}

// placeToggle returns the home page address with target toggled in region=,
// applying state/county replacement (decisions.placeControlUnifiesRegionsAndCounties):
// choosing a county whose parent state is selected replaces the state with the
// county; choosing a state when any of its counties are selected replaces those
// counties with the state; choosing a selected place removes it, and removing
// the last returns to All. Any pending place= search query is dropped.
func placeToggle(q url.Values, target string) template.URL {
	nq := cloneValues(q)
	nq.Del("place")
	targetRg, targetOK := region.BySlug(target)
	cur := nextValues(nq, "region")
	var next []string
	found := false
	for _, v := range cur {
		if v == target {
			found = true
			continue
		}
		next = append(next, v)
	}
	if !found {
		filtered := next[:0]
		for _, v := range next {
			if targetOK {
				if rg, ok := region.BySlug(v); ok {
					if targetRg.Kind == region.County && rg.ID == targetRg.Parent {
						continue
					}
					if targetRg.Kind == region.Land && rg.Kind == region.County && rg.Parent == targetRg.ID {
						continue
					}
				}
			}
			filtered = append(filtered, v)
		}
		next = append(filtered, target)
	}
	nq.Del("region")
	for _, v := range next {
		nq.Add("region", v)
	}
	return nextURL(nq)
}

// nextPlaces serves the places the home page's Place box searches as the
// reader types (product.nextView, decisions.placeControlUnifiesRegionsAndCounties):
// [slug, label, parentSlug, codeNorm, nameNorm, labelNorm, fullNorm].
func (s *server) nextPlaces(w http.ResponseWriter, r *http.Request) {
	entries := s.placeEntries()
	out := make([][7]string, 0, len(entries))
	for _, e := range entries {
		out = append(out, [7]string{e.Slug, e.Label, e.Parent, e.CodeNorm, e.NameNorm, e.LabelNorm, e.FullNorm})
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(out)
}

// placeSearch holds the Place popover's search box state and server-rendered
// matches (for no-JavaScript Enter submission).
type placeSearch struct {
	Query        string
	Placeholder  string
	Found        int
	Limit        int
	Matches      []nextChoice
	Fields       []hiddenField
	CloseURL     template.URL
	PlacesURL    string
	BaseQuery    string
	SelectedData string
}

// placeControl builds the unified Place setting (product.nextView,
// decisions.placeControlUnifiesRegionsAndCounties).
func (s *server) placeControl(q url.Values, opt options) nextControl {
	rc := nextControl{
		Label:  "Place",
		Value:  "all",
		AllURL: nextSet(q, "region", "", "place"),
		AllOn:  !opt.HasRegions,
		Browse: true,
	}
	if opt.HasRegions {
		rc.Value = nextList(s.regionNames(opt.Regions))
	}
	// Currently selected places in address order, each a removable chip, and
	// serialized as slug:parent pairs for next-place.js's toggle URLs.
	var selData []string
	seenSel := map[string]bool{}
	for _, v := range nextValues(q, "region") {
		rg, ok := region.BySlug(v)
		if !ok || seenSel[rg.ID] {
			continue
		}
		seenSel[rg.ID] = true
		slug := rg.Slug()
		parent := ""
		if rg.Kind == region.County {
			if st, ok := region.Get(rg.Parent); ok {
				parent = st.Slug()
			}
		}
		selData = append(selData, slug+":"+parent)
		rc.Selected = append(rc.Selected, nextChoice{
			Label: rg.Label(),
			Title: "Remove " + rg.Label(),
			On:    true,
			URL:   placeToggle(q, slug),
		})
	}

	// Collapsed browse list: states, provinces, islands, and seas by zone.
	for _, z := range region.Zones {
		grp := nextChoiceGroup{Name: z}
		for _, rg := range region.InZone(z) {
			grp.Choices = append(grp.Choices, nextChoice{
				Label: rg.Name,
				Title: rg.Code,
				On:    opt.HasRegions && opt.Regions[rg.ID],
				URL:   placeToggle(q, rg.Slug()),
			})
		}
		rc.Groups = append(rc.Groups, grp)
	}

	placeholder := "Find a state, province, or sea"
	if s.counties != nil {
		placeholder = "Find a state, province, county, or sea"
	}
	baseQ := cloneValues(q)
	baseQ.Del("region")
	baseQ.Del("place")
	baseQ.Del("bird")

	closeQ := cloneValues(q)
	closeQ.Del("place")

	ps := &placeSearch{
		Query:        strings.TrimSpace(q.Get("place")),
		Placeholder:  placeholder,
		Limit:        nextPlaceLimit,
		CloseURL:     nextURL(closeQ),
		PlacesURL:    "/places.json",
		BaseQuery:    baseQ.Encode(),
		SelectedData: strings.Join(selData, " "),
	}
	for _, k := range sortedKeys(q) {
		if k == "place" || k == "bird" {
			continue
		}
		for _, v := range q[k] {
			ps.Fields = append(ps.Fields, hiddenField{k, v})
		}
	}
	if ps.Query != "" {
		rc.Open = true
		hits := placeMatches(s.placeEntries(), ps.Query)
		ps.Found = len(hits)
		shown := hits
		if len(shown) > nextPlaceLimit {
			shown = shown[:nextPlaceLimit]
		}
		for _, e := range shown {
			on := false
			if rg, ok := region.BySlug(e.Slug); ok && opt.HasRegions && opt.Regions[rg.ID] {
				on = true
			}
			ps.Matches = append(ps.Matches, nextChoice{
				Label: e.Label,
				On:    on,
				URL:   placeToggle(q, e.Slug),
			})
		}
	}
	rc.Place = ps
	return rc
}
