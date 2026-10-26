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
	"net/http/httptest"
	"net/url"
	"os/exec"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"
)

// searchCases are queries and the fixture birds each must leave showing:
// a name, a word shared by two names, the end of one word and the start of
// the next, a scientific name, a code in capitals, a name typed without its
// punctuation, a family; and three that must match nothing: words out of
// order, a query running from one field into the next, and nonsense.
var searchCases = []struct {
	q    string
	want []string
}{
	{"crow", []string{"amecro"}},
	{"american", []string{"amecro", "amerob"}},
	{"n robin", []string{"amerob"}},
	{"corvus brachy", []string{"amecro"}},
	{"AMECRO", []string{"amecro"}},
	{"black necked", []string{"blasti"}},
	{"ardeidae", []string{"greblu"}},
	{"heron great", nil},
	{"heron ardea", nil},
	{"zzz", nil},
}

// goMatches is the fixture birds the Go matcher keeps for q, sorted.
func goMatches(s *server, q string) []string {
	var out []string
	for i := range s.ds.Species {
		if id := s.ds.Species[i].ID; searchMatches(s.search[id], q) {
			out = append(out, id)
		}
	}
	sort.Strings(out)
	return out
}

// TestSearchMatchesNamesCodesAndTaxa checks the matching rule: every word
// must occur, ignoring case and punctuation, in a bird's names, code,
// scientific name, family, or order.
//
// Verifies: product.searchBox. Criterion: app.searchMatchesInGoAndTheBrowser.
func TestSearchMatchesNamesCodesAndTaxa(t *testing.T) {
	s := newTestServer(t)
	for _, c := range searchCases {
		if got := goMatches(s, c.q); !slices.Equal(got, c.want) {
			t.Errorf("search %q keeps %v, want %v", c.q, got, c.want)
		}
	}
	if got := normalizeSearch("  Chuck-will's-widow  "); got != "chuck will s widow" {
		t.Errorf("normalizeSearch = %q", got)
	}
	// Families and orders are searched by their scientific names only: the
	// English descriptions list other birds. The fixture's are "<Family>
	// family" and "<Order> (group)", so neither word may be searchable.
	for id, text := range s.search {
		for _, w := range []string{"family", "group"} {
			if strings.Contains(text, w) {
				t.Errorf("%s's search text %q includes a family's or order's English description", id, text)
			}
		}
	}
	if got := goMatches(s, "passeriformes"); len(got) == 0 {
		t.Error("no bird matches its order's scientific name")
	}
}

// TestSearchFromTheAddress loads the index with q, as a shared link or a
// browser without JavaScript does, and checks the page shows only the
// matches, counts only them, names the search, keeps it in its controls and
// drops it from the links that return to a fixed view.
//
// Verifies: product.searchBox. Criterion: app.searchFromTheAddress.
func TestSearchFromTheAddress(t *testing.T) {
	s := newTestServer(t)
	body := mustGet(t, s, "/?by=mass&q=american")
	shown := regexp.MustCompile(`<div class="tilewrap" id="s-(\w+)" data-search="[^"]*">`).FindAllStringSubmatch(body, -1)
	var ids []string
	for _, m := range shown {
		ids = append(ids, m[1])
	}
	sort.Strings(ids)
	if want := []string{"amecro", "amerob"}; !slices.Equal(ids, want) {
		t.Errorf("?q=american shows %v, want %v", ids, want)
	}
	if n := strings.Count(body, `data-search="`); n != len(s.ds.Species) {
		t.Errorf("%d tiles carry search text, want every one of %d, hidden or not", n, len(s.ds.Species))
	}
	if !strings.Contains(body, ">2 of 6 species<") {
		t.Error("the count line does not count only the matches")
	}
	if !regexp.MustCompile(`<span class="fsum fsearch"><b>Search</b> <span class="fsq">american</span>`).MatchString(body) {
		t.Error("the count line does not name the search")
	}
	if strings.Contains(body, `<p class="empty searchempty">`) {
		t.Error("the page says nothing matches when two birds do")
	}
	// Sections the search empties are hidden; their counts are the matches.
	for _, m := range regexp.MustCompile(`(?s)<section id="tier-\d+"( hidden)?>.*?<span class="n">(\d+) species</span>`).FindAllStringSubmatch(body, -1) {
		if (m[1] == "") == (m[2] == "0") {
			t.Errorf("a section counting %s species has hidden=%t", m[2], m[1] != "")
		}
	}
	// The form keeps every other setting for Enter without JavaScript.
	if !strings.Contains(body, `<input type="hidden" name="by" value="mass">`) {
		t.Error("the search form drops the other settings")
	}
	if !strings.Contains(body, `name="q" value="american"`) {
		t.Error("the box does not show the search it was opened with")
	}
	// Controls keep q; the welcome line, reset, and Clear do not.
	welcome := mustMatch(t, body, `(?s)<p class="welcome">(.*?)</p>`, "the welcome line")
	for _, h := range regexp.MustCompile(`href="([^"]*)"`).FindAllStringSubmatch(welcome, -1) {
		if strings.Contains(html.UnescapeString(h[1]), "q=") {
			t.Errorf("welcome link %s carries the search; it returns to a fixed view", h[1])
		}
	}
	controls := mustMatch(t, body, `(?s)<div class="row ordering">(.*?)</div>\s*</details>`, "the controls")
	if !strings.Contains(controls, "q=american") {
		t.Error("the controls drop the search")
	}
	if got := mustGet(t, s, "/?family=Ardeidae&region=AK&q=heron"); !regexp.MustCompile(`No birds match\. <a href="[^"]*"`).MatchString(got) {
		t.Fatal("Ardeidae in Alaska no longer empties the page; pick another")
	} else if clear := mustMatch(t, got, `No birds match\. <a href="([^"]*)"`, "Clear the filters"); strings.Contains(html.UnescapeString(clear), "q=") {
		t.Errorf("Clear the filters keeps the search: %s", clear)
	}

	none := mustGet(t, s, "/?q=zzz")
	if !strings.Contains(none, `<p class="empty searchempty">`) || !strings.Contains(none, ">0 of 6 species<") {
		t.Error("a search that matches nothing does not say so")
	}
}

//go:embed testdata/search_probe.js
var searchProbe string

// TestSearchInTheBrowser types into the box in headless Chrome and checks that
// the script keeps exactly the birds the Go matcher does, counts them, names
// the search, and puts it into the next link followed -- a control, a tile,
// Here -- but not into the welcome line's links or the guides.
//
// Verifies: product.searchBox. Criterion: app.searchMatchesInGoAndTheBrowser.
func TestSearchInTheBrowser(t *testing.T) {
	chrome := findChrome()
	if chrome == "" {
		t.Skip("no Chrome or Chromium found (set BIRDVIEW_CHROME); the search box was not tested in a browser")
	}
	s := newTestServer(t)
	var qs []string
	for _, c := range searchCases {
		qs = append(qs, c.q)
	}
	qs = append(qs, "crow") // last, so the link checks run with a search in the box
	js, _ := json.Marshal(qs)
	ts := httptest.NewServer(instrument(s, searchProbe, nil))
	defer ts.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	out, err := exec.CommandContext(ctx, chrome, "--headless", "--disable-gpu",
		"--no-first-run", "--no-default-browser-check", "--disable-background-networking",
		"--disable-component-update", "--disable-sync", "--disable-extensions",
		"--user-data-dir="+t.TempDir(), "--virtual-time-budget=5000",
		"--dump-dom", ts.URL+"/?by=mass#queries="+url.QueryEscape(string(js))).Output()
	if err != nil {
		t.Fatalf("chrome: %v", err)
	}
	m := dataAttr("data-probe").FindSubmatch(out)
	if m == nil {
		t.Fatal("no result from the probe; it did not run")
	}
	var res struct {
		Found bool `json:"found"`
		Typed []struct {
			Q        string            `json:"q"`
			IDs      []string          `json:"ids"`
			Sections []string          `json:"sections"`
			Bars     map[string]string `json:"bars"`
			Count    string            `json:"count"`
			Named    string            `json:"named"`
			Empty    bool              `json:"empty"`
		} `json:"typed"`
		Links map[string]*string `json:"links"`
	}
	if err := json.Unmarshal([]byte(html.UnescapeString(string(m[1]))), &res); err != nil {
		t.Fatalf("reading the probe: %v", err)
	}
	if !res.Found || len(res.Typed) != len(qs) {
		t.Fatalf("the probe found the box %t and typed %d of %d queries", res.Found, len(res.Typed), len(qs))
	}
	for _, r := range res.Typed {
		var got []string
		for _, id := range r.IDs {
			got = append(got, strings.TrimPrefix(id, "s-"))
		}
		sort.Strings(got)
		want := goMatches(s, r.Q)
		if !slices.Equal(got, want) {
			t.Errorf("typing %q leaves %v; the Go matcher keeps %v", r.Q, got, want)
		}
		if wantCount := strings.Join([]string{strconv.Itoa(len(want)), "of", strconv.Itoa(len(s.ds.Species)), "species"}, " "); r.Count != wantCount {
			t.Errorf("typing %q: the count line reads %q, want %q", r.Q, r.Count, wantCount)
		}
		if r.Named != r.Q {
			t.Errorf("typing %q: the count line names %q", r.Q, r.Named)
		}
		if r.Empty != (len(want) == 0) {
			t.Errorf("typing %q: the nothing-matches note is shown=%t with %d matches", r.Q, r.Empty, len(want))
		}
		shownSecs := 0
		for _, sec := range r.Sections {
			if strings.HasSuffix(sec, ":0 species") {
				t.Errorf("typing %q leaves an empty section showing: %s", r.Q, sec)
			}
			shownSecs++
		}
		for id, h := range r.Bars {
			if h == "0%" && slices.ContainsFunc(r.Sections, func(s string) bool { return strings.HasPrefix(s, id+":") }) {
				t.Errorf("typing %q: %s shows birds but its bar is 0%%", r.Q, id)
			}
		}
		if len(want) > 0 && shownSecs == 0 {
			t.Errorf("typing %q hides every section", r.Q)
		}
	}
	link := func(name string) string {
		if p := res.Links[name]; p != nil {
			return html.UnescapeString(*p)
		}
		t.Errorf("the probe found no %s link", name)
		return ""
	}
	for _, name := range []string{"control", "tile", "here"} {
		if l := link(name); !strings.Contains(l, "q=crow") {
			t.Errorf("after typing, the %s link goes to %s, without the search", name, l)
		}
	}
	for _, name := range []string{"welcome", "reset", "docs"} {
		if l := link(name); strings.Contains(l, "q=") {
			t.Errorf("the %s link goes to %s, carrying the search", name, l)
		}
	}
}
