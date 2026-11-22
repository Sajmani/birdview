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
	"net/http/httptest"
	"net/url"
	"os/exec"
	"regexp"
	"strings"
	"testing"
	"time"

	"birdview/internal/dataset"
)

// placeTestServer builds a server with -counties enabled and a rich set of
// counties (15 Washington counties across states, plus Kings CA/NY, Atlantic
// NJ, Los Angeles CA, San Diego CA, and Philadelphia PA) to exercise ranking,
// top-12 truncation, diacritics, and state/county replacement.
func placeTestServer(t *testing.T) *server {
	t.Helper()
	countiesOn(t)
	ds, cs := countyDataset()
	extra := []dataset.County{
		{ID: "USA.5.37_1", Name: "San Diego", State: "USA.5_1", Total: 300000, Monthly: [12]int64{0: 300000}},
		{ID: "USA.5.16_1", Name: "Kings", State: "USA.5_1", Total: 100000, Monthly: [12]int64{0: 100000}},
		{ID: "USA.33.24_1", Name: "Kings", State: "USA.33_1", Total: 200000, Monthly: [12]int64{0: 200000}},
		{ID: "USA.31.1_1", Name: "Atlantic", State: "USA.31_1", Total: 150000, Monthly: [12]int64{0: 150000}},
		{ID: "USA.39.51_1", Name: "Philadelphia", State: "USA.39_1", Total: 200000, Monthly: [12]int64{0: 200000}},
	}
	// 15 Washington counties across states so "washington" exceeds nextPlaceLimit (12).
	washStates := []string{
		"USA.1_1", "USA.4_1", "USA.6_1", "USA.10_1", "USA.11_1",
		"USA.13_1", "USA.14_1", "USA.15_1", "USA.16_1", "USA.17_1",
		"USA.20_1", "USA.21_1", "USA.33_1", "USA.34_1", "USA.39_1",
	}
	for i, st := range washStates {
		extra = append(extra, dataset.County{
			ID:      fmt.Sprintf("%s.%d_1", strings.TrimSuffix(st, "_1"), 90+i),
			Name:    "Washington",
			State:   st,
			Total:   100000,
			Monthly: [12]int64{0: 100000},
		})
	}
	cs = testCounties(ds, append(cs.Counties, extra...), map[string]map[string][12]int64{
		"USA.5.19_1":  {"annhum": {0: 3000}},
		"USA.39.51_1": {"amecro": {0: 600}},
	})
	s, err := newServer(ds, cs)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

// TestPlaceSearchRankingAndDiacritics asserts app.placeSearchFindsAndReplaces's
// matching and ranking rules: diacritics fold ("quebec" -> "Québec"), states,
// provinces, islands, and seas rank ahead of counties ("washington" puts
// Washington state #1, "atlantic" puts North Atlantic ahead of Atlantic, NJ),
// postal codes and parent states disambiguate ("washington ny", "washington ar"
// without matching North Carolina, "kings ca", "ny"), and broad queries are
// never hidden.
//
// Verifies: app.placeSearchFindsAndReplaces
func TestPlaceSearchRankingAndDiacritics(t *testing.T) {
	s := placeTestServer(t)
	entries := s.placeEntries()

	hits := placeMatches(entries, "washington")
	if len(hits) != 16 {
		t.Fatalf("washington matched %d places, want 16 (state + 15 counties)", len(hits))
	}
	if hits[0].Label != "Washington" || hits[0].County {
		t.Errorf("washington #1 = %+v, want Washington state first", hits[0])
	}

	if got := placeMatches(entries, "washington ny"); len(got) != 1 || got[0].Label != "Washington, NY" {
		t.Errorf("washington ny = %v, want only Washington, NY", got)
	}
	if got := placeMatches(entries, "washington ar"); len(got) != 1 || got[0].Label != "Washington, AR" {
		t.Errorf("washington ar = %v, want only Washington, AR (not North Carolina)", got)
	}
	if got := placeMatches(entries, "quebec"); len(got) == 0 || got[0].Label != "Québec" {
		t.Errorf("quebec = %v, want Québec first", got)
	}
	if got := placeMatches(entries, "kings ca"); len(got) != 1 || got[0].Label != "Kings, CA" {
		t.Errorf("kings ca = %v, want only Kings, CA", got)
	}
	if got := placeMatches(entries, "ny"); len(got) == 0 || got[0].Label != "New York" || got[0].County {
		t.Errorf("ny #1 = %v, want New York state first", got)
	}
	if got := placeMatches(entries, "atlantic"); len(got) < 2 || got[0].Label != "North Atlantic" || got[1].Label != "Atlantic, NJ" {
		t.Errorf("atlantic = %v, want North Atlantic before Atlantic, NJ", got)
	}
}

// TestPlaceToggleReplacesParentAndChild asserts app.placeSearchFindsAndReplaces's
// selection and replacement rules: no separate County setting exists; choosing a
// county while its parent state is selected replaces the state; choosing a state
// while any of its counties are selected replaces those counties; counties in
// different states compose; clicking a selected place removes it, and removing
// the last returns to All.
//
// Verifies: app.placeSearchFindsAndReplaces
func TestPlaceToggleReplacesParentAndChild(t *testing.T) {
	s := placeTestServer(t)

	// No separate County control even when one state with counties is selected.
	for _, target := range []string{"/", "/?region=USA.5_1", "/?region=USA.5.19_1"} {
		for _, c := range nextAt(s, target).Controls {
			if c.Label == "County" || c.Label == "Region" {
				t.Errorf("%s has legacy %q setting; Place replaces both", target, c.Label)
			}
		}
	}

	// From California, picking Los Angeles replaces California (dropping any place=).
	if got := string(placeToggle(url.Values{"region": {"USA.5_1"}, "place": {"los"}}, "USA.5.19_1")); got != "/?region=USA.5.19_1" {
		t.Errorf("CA -> LA = %q, want /?region=USA.5.19_1", got)
	}
	// From Los Angeles, picking San Diego adds it.
	if got := string(placeToggle(url.Values{"region": {"USA.5.19_1"}}, "USA.5.37_1")); got != "/?region=USA.5.19_1&region=USA.5.37_1" {
		t.Errorf("LA -> +SD = %q, want both counties", got)
	}
	// From Los Angeles, picking Philadelphia (PA) keeps both cross-state counties.
	if got := string(placeToggle(url.Values{"region": {"USA.5.19_1"}}, "USA.39.51_1")); got != "/?region=USA.5.19_1&region=USA.39.51_1" {
		t.Errorf("LA -> +Philadelphia = %q, want both cross-state counties", got)
	}
	// From Los Angeles + San Diego + Philadelphia, picking California replaces LA and SD while keeping Philadelphia.
	if got := string(placeToggle(url.Values{"region": {"USA.5.19_1", "USA.5.37_1", "USA.39.51_1"}}, "USA.5_1")); got != "/?region=USA.39.51_1&region=USA.5_1" {
		t.Errorf("LA+SD+PHL -> +CA = %q, want /?region=USA.39.51_1&region=USA.5_1", got)
	}
	// Clicking the last selected place unselects it and returns to All.
	if got := string(placeToggle(url.Values{"region": {"USA.5.19_1"}}, "USA.5.19_1")); got != "/" {
		t.Errorf("LA -> unselect = %q, want / (All)", got)
	}
}

// TestPlaceSearchNoJS asserts app.placeSearchFindsAndReplaces without JavaScript:
// submitting ?place=washington opens the Place popover, shows the top 12 matches
// with Washington first and a truncation note naming all 16, keeps the browse
// disclosure collapsed, and lists selected places as removable chips.
//
// Verifies: app.placeSearchFindsAndReplaces
func TestPlaceSearchNoJS(t *testing.T) {
	s := placeTestServer(t)
	_, body := get(t, s, "/?region=USA.5_1&place=washington")
	if !strings.Contains(body, `<details class="ctl" name="ctl" data-ctl="Place" open>`) {
		t.Fatal("submitting place= did not open the Place popover")
	}
	pop := box(t, body, "Place")
	if !strings.Contains(pop, `<details class="browse">`) {
		t.Error("Place popover lacks a collapsed browse disclosure")
	}
	if !strings.Contains(pop, `Showing 12 of 16 matches`) {
		t.Errorf("Place popover missing top-12 truncation note:\n%s", pop)
	}
	if !strings.Contains(pop, `title="Remove California"`) {
		t.Error("Place popover does not show currently selected California as a removable chip")
	}
	// /places.json is not gated by the filtered-view rate limiter (tech.filteredViewsBounded).
	_, pj := get(t, s, "/places.json")
	if !strings.Contains(pj, `"USA.5.19_1"`) || !strings.Contains(pj, `"Québec"`) {
		t.Error("/places.json missing expected county or Québec entry")
	}
}

//go:embed testdata/next_place_probe.js
var nextPlaceProbe string

// TestPlaceSearchInTheBrowser types into the Place search box in headless Chrome
// and asserts that next-place.js produces the exact same status message, close
// link, ranked matches, and toggle URLs as the server for the same place= query,
// and that opening the Browse disclosure stays within a 360px phone viewport.
//
// Verifies: app.placeSearchFindsAndReplaces
func TestPlaceSearchInTheBrowser(t *testing.T) {
	chrome := findChrome()
	if chrome == "" {
		t.Skip("no Chrome or Chromium found (set BIRDVIEW_CHROME); Place search was not tested in a browser")
	}
	s := placeTestServer(t)
	qs := []string{"washington", "washington ny", "washington ar", "quebec", "kings ca", "atlantic", "los angeles", "zzzz"}
	js, _ := json.Marshal(qs)
	ts := httptest.NewServer(instrument(s, nextPlaceProbe, nil))
	defer ts.Close()
	// Start with California selected so state->county replacement (Los Angeles)
	// and cross-state addition (Washington, NY) are both exercised in JS toggleURL.
	start := "/?month=May&region=USA.5_1"
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	out, err := exec.CommandContext(ctx, chrome, "--headless", "--disable-gpu",
		"--window-size=360,800",
		"--no-first-run", "--no-default-browser-check", "--disable-background-networking",
		"--disable-component-update", "--disable-sync", "--disable-extensions",
		"--user-data-dir="+t.TempDir(), "--virtual-time-budget=5000",
		"--dump-dom", ts.URL+start+"#queries="+url.QueryEscape(string(js))).Output()
	if err != nil {
		t.Fatalf("chrome: %v", err)
	}
	m := dataAttr("data-probe").FindSubmatch(out)
	if m == nil {
		t.Fatal("no result from the place probe; it did not run")
	}
	var res struct {
		Typed []struct {
			Q     string `json:"q"`
			Found string `json:"found"`
			Close string `json:"close"`
			Items []struct {
				Cls  string `json:"cls"`
				Text string `json:"text"`
				Href string `json:"href"`
			} `json:"items"`
		} `json:"typed"`
		BrowseInitiallyOpen *bool   `json:"browseInitiallyOpen"`
		PopLeft             float64 `json:"popLeft"`
		PopRight            float64 `json:"popRight"`
		Viewport            float64 `json:"viewport"`
	}
	if err := json.Unmarshal([]byte(html.UnescapeString(string(m[1]))), &res); err != nil {
		t.Fatalf("reading the place probe: %v", err)
	}
	if len(res.Typed) != len(qs) {
		t.Fatalf("the probe typed %d of %d queries", len(res.Typed), len(qs))
	}
	if res.BrowseInitiallyOpen == nil || *res.BrowseInitiallyOpen {
		t.Errorf("browse disclosure initially open = %v, want false", res.BrowseInitiallyOpen)
	}
	if res.PopLeft < 0 || res.PopRight > res.Viewport+1 {
		t.Errorf("opened Place popover with Browse spans [%.1f, %.1f] on %.0fpx viewport", res.PopLeft, res.PopRight, res.Viewport)
	}

	foundRe := regexp.MustCompile(`<div class="placefinds"[^>]*>(?s:(.*?))</div>\s*<details class="browse">`)
	pRe := regexp.MustCompile(`<p class="found"><a class="findclose" href="([^"]*)"[^>]*>&times;</a>([^<]*)</p>`)
	aRe := regexp.MustCompile(`<a class="(btn(?: on)?)" href="([^"]*)"[^>]*>([^<]*)</a>`)
	for _, r := range res.Typed {
		_, body := get(t, s, start+"&place="+url.QueryEscape(r.Q))
		fm := foundRe.FindStringSubmatch(body)
		if fm == nil {
			t.Fatalf("server page for place=%q has no .placefinds block", r.Q)
		}
		pm := pRe.FindStringSubmatch(fm[1])
		if pm == nil {
			t.Fatalf("server page for place=%q has no .found status line", r.Q)
		}
		if want := html.UnescapeString(pm[2]); r.Found != want {
			t.Errorf("typing %q: script status = %q, server = %q", r.Q, r.Found, want)
		}
		if want := html.UnescapeString(pm[1]); r.Close != want {
			t.Errorf("typing %q: script close = %q, server = %q", r.Q, r.Close, want)
		}
		var wantItems []string
		for _, am := range aRe.FindAllStringSubmatch(fm[1], -1) {
			wantItems = append(wantItems, am[1]+"|"+html.UnescapeString(am[3])+"|"+html.UnescapeString(am[2]))
		}
		var gotItems []string
		for _, it := range r.Items {
			gotItems = append(gotItems, it.Cls+"|"+it.Text+"|"+it.Href)
		}
		if strings.Join(gotItems, "\n") != strings.Join(wantItems, "\n") {
			t.Errorf("typing %q: script items differ from server:\ngot:\n%s\nwant:\n%s",
				r.Q, strings.Join(gotItems, "\n"), strings.Join(wantItems, "\n"))
		}
	}
}
