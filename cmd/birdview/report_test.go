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
	"html"
	"net/url"
	"regexp"
	"strings"
	"testing"
)

// TestReportDataIssueLinks asserts product.reportDataIssue: with a tracker set, the species page
// has a Report a data issue link under each of its two tables, opening a new issue whose title
// names the bird and the table and whose body holds the page's address, that table's rows, and the
// data's versions; with no tracker, no link.
//
// Verifies: product.reportDataIssue, product.noPersonalAddress. Criterion: acceptance.reportDataIssueLinks.
func TestReportDataIssueLinks(t *testing.T) {
	s := newTestServer(t)
	s.feedback = "https://github.com/google/birdview/issues"
	s.ds.SourceVersions = map[string]string{"avonet": "AVONET v7", "avilist": "AviList v2025b"}
	s.revision = "0123456789ab"
	_, body := get(t, s, "/species/amecro?units=metric")
	links := regexp.MustCompile(`<a class="report" href="([^"]*)"[^>]*>Report a data issue</a>`).FindAllStringSubmatch(body, -1)
	if len(links) != 2 {
		t.Fatalf("the species page has %d Report a data issue links, want 2", len(links))
	}
	for i, table := range []struct {
		heading, title string
		rows           []string
	}{
		{"Measurements and sources", "measurements", []string{"Wingspan: 100.0 cm (modeled)", "Mass: 450 g (measured)"}},
		{"Categories and sources", "categories", []string{"Habitat: ", "Diet: ", "Family: ", "(published)"}},
	} {
		at := strings.Index(body, links[i][0])
		if h := strings.LastIndex(body[:at], "<h2>"); h < 0 || !strings.HasPrefix(body[h:], "<h2>"+table.heading+"</h2>") {
			t.Errorf("link %d is not under %s", i, table.heading)
		}
		u, err := url.Parse(html.UnescapeString(links[i][1]))
		if err != nil {
			t.Fatal(err)
		}
		if u.Scheme+"://"+u.Host+u.Path != "https://github.com/google/birdview/issues/new" {
			t.Errorf("%s's link opens %s, want a new issue", table.heading, u)
		}
		title, b := u.Query().Get("title"), u.Query().Get("body")
		if !strings.Contains(title, "American Crow") || !strings.Contains(title, "Corvus brachyrhynchos") || !strings.Contains(title, table.title) {
			t.Errorf("%s's issue title is %q; want the bird's names and %q", table.heading, title, table.title)
		}
		for _, want := range append([]string{"/species/amecro?units=metric", "avonet: AVONET v7", "avilist: AviList v2025b", "BirdView server: 0123456789ab"}, table.rows...) {
			if !strings.Contains(b, want) {
				t.Errorf("%s's issue body lacks %q:\n%s", table.heading, want, b)
			}
		}
	}
	for _, l := range links {
		if len(l[1]) > 8000 {
			t.Errorf("a report link is %d bytes long; GitHub refuses addresses much past 8,000", len(l[1]))
		}
	}
	s.feedback = ""
	if _, body := get(t, s, "/species/amecro"); strings.Contains(body, "Report a data issue") {
		t.Error("with no tracker set, the species page still offers Report a data issue")
	}
}

// TestHomeReportLink asserts product.reportDataIssue's home page link: with a tracker set, Found
// something wrong? Report it opens a new issue whose body holds the page's address, its heading (the
// settings in view), the data's versions, and the server's revision, the title left to the reader;
// with no tracker, no link.
//
// Verifies: product.reportDataIssue, tech.imageRecordsRevision. Criterion: acceptance.reportDataIssueLinks.
func TestHomeReportLink(t *testing.T) {
	s := newTestServer(t)
	s.feedback = "https://github.com/google/birdview/issues"
	s.ds.SourceVersions = map[string]string{"avonet": "AVONET v7"}
	s.revision = "0123456789ab"
	_, body := get(t, s, "/?month=May&freq=1000")
	m := regexp.MustCompile(`<a href="([^"]*)">Report it</a>`).FindStringSubmatch(body)
	if m == nil {
		t.Fatal("the home page has no Report it link")
	}
	u, err := url.Parse(html.UnescapeString(m[1]))
	if err != nil {
		t.Fatal(err)
	}
	if u.Scheme+"://"+u.Host+u.Path != "https://github.com/google/birdview/issues/new" || u.Query().Has("title") {
		t.Errorf("Report it opens %s; want a new issue with the title left to the reader", u)
	}
	b := u.Query().Get("body")
	for _, want := range []string{"/?month=May&freq=1000", "Showing birds in May", "At least 1 in 1,000",
		"avonet: AVONET v7", "BirdView server: 0123456789ab"} {
		if !strings.Contains(b, want) {
			t.Errorf("the home page's issue body lacks %q:\n%s", want, b)
		}
	}
	s.feedback = ""
	if _, body := get(t, s, "/"); strings.Contains(body, "Report it") {
		t.Error("with no tracker set, the home page still offers Report it")
	}
}
