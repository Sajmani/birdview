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
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
)

// robotsRule is one Allow or Disallow line of a robots.txt group.
type robotsRule struct {
	allow   bool
	pattern string
}

// robotsAllowed decides a path the way RFC 9309 says a crawler must: the
// longest matching rule wins, Allow wins a tie, and no match means allowed.
// In a pattern, * matches any run of characters and a trailing $ anchors the
// end. The path includes its query string, as the RFC requires.
func robotsAllowed(rules []robotsRule, path string) bool {
	best, allowed := -1, true
	for _, r := range rules {
		if r.pattern == "" {
			continue // an empty Disallow allows everything; it matches nothing here
		}
		expr := "^" + strings.ReplaceAll(regexp.QuoteMeta(strings.TrimSuffix(r.pattern, "$")), `\*`, ".*")
		if strings.HasSuffix(r.pattern, "$") {
			expr += "$"
		}
		if !regexp.MustCompile(expr).MatchString(path) {
			continue
		}
		if n := len(r.pattern); n > best || n == best && r.allow {
			best, allowed = n, r.allow
		}
	}
	return allowed
}

// parseRobots returns the rules of the group that applies to every crawler,
// User-agent: *. It fails the test on a line it does not understand, so a
// typo in the policy cannot silently drop a rule.
func parseRobots(t *testing.T, body string) []robotsRule {
	t.Helper()
	var rules []robotsRule
	inStar := false
	for _, line := range strings.Split(body, "\n") {
		line, _, _ = strings.Cut(line, "#")
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		key, val, ok := strings.Cut(line, ":")
		if !ok {
			t.Fatalf("robots.txt line %q has no field", line)
		}
		val = strings.TrimSpace(val)
		switch strings.ToLower(strings.TrimSpace(key)) {
		case "user-agent":
			inStar = val == "*"
		case "allow", "disallow":
			if inStar {
				rules = append(rules, robotsRule{allow: strings.EqualFold(strings.TrimSpace(key), "allow"), pattern: val})
			}
		case "sitemap":
		default:
			t.Fatalf("robots.txt line %q has an unknown field", line)
		}
	}
	return rules
}

// TestRobotsKeepsCrawlersToPages asserts that /robots.txt lets a crawler
// fetch every page the site has without a query string, except the
// comparison, and asks it to fetch nothing with one, nor /now or the boundary
// file. The pages come from the same crawl the accessibility checks use, so a
// page added later is judged without anyone listing it here.
//
// Verifies: tech.crawlersSkipFilteredViews, tech.servesOnlyItsPages. Criterion: acceptance.robotsKeepsCrawlersToPages.
func TestRobotsKeepsCrawlersToPages(t *testing.T) {
	s := newTestServer(t)
	w := httptest.NewRecorder()
	s.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/robots.txt", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("GET /robots.txt = %d, want 200", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); !strings.HasPrefix(ct, "text/plain") {
		t.Errorf("robots.txt is served as %q, want text/plain", ct)
	}
	rules := parseRobots(t, w.Body.String())
	if len(rules) == 0 {
		t.Fatalf("robots.txt has no rules for User-agent: *:\n%s", w.Body.String())
	}

	var free, filtered int
	for _, p := range discoverPages(t, s) {
		if p.url == "/robots.txt" {
			continue
		}
		path, query, _ := strings.Cut(p.url, "?")
		blocked := strings.HasPrefix(path, "/compare")
		want := query == "" && !blocked
		if strings.Contains(p.url, "?") {
			filtered++
		} else if !blocked {
			free++
		}
		if got := robotsAllowed(rules, p.url); got != want {
			t.Errorf("robots.txt allows %s: %v, want %v", p.url, got, want)
		}
	}
	// The crawl must have offered both kinds, or the loop proved nothing.
	if free < 3 || filtered < 3 {
		t.Errorf("crawl found %d query-free and %d filtered pages; too few to judge the policy", free, filtered)
	}
	for _, p := range []string{"/now", "/now?region=ME", s.boundaryURL} {
		if robotsAllowed(rules, p) {
			t.Errorf("robots.txt allows %s, which is no page to index", p)
		}
	}
	for _, p := range []string{"/", "/species/amecro", "/docs", "/docs/using", "/static/styles.css"} {
		if !robotsAllowed(rules, p) {
			t.Errorf("robots.txt disallows %s, which a crawler needs", p)
		}
	}
}

// TestRobotsMatcher pins robotsAllowed to RFC 9309's examples of precedence
// and wildcards, so the policy test above is judged by the rules crawlers
// actually apply rather than by whatever the matcher happens to do.
func TestRobotsMatcher(t *testing.T) {
	rules := []robotsRule{
		{false, "/*?"}, {false, "/compare"}, {true, "/page"}, {false, "/*.php$"}, {false, "/fish"}, {true, "/fish.html"},
	}
	for path, want := range map[string]bool{
		"/":                 true,
		"/?family=Corvidae": false,
		"/species/x?u=1":    false,
		"/compare":          false,
		"/compareX":         false,
		"/page":             true,
		"/index.php":        false,
		"/index.php?x":      false, // "/*?" matches; "/*.php$" does not, since $ anchors the end
		"/index.phpx":       true,
		"/fish":             false,
		"/fish.html":        true, // longer Allow beats shorter Disallow
		"/Fish":             true, // paths are case-sensitive
	} {
		if got := robotsAllowed(rules, path); got != want {
			t.Errorf("robotsAllowed(%q) = %v, want %v", path, got, want)
		}
	}
}
