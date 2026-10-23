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
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
)

// countLine is the index's summary of every setting, as text. Every stable
// parameter names itself there, so it is where a parameter that stopped being
// read would show.
func countLine(t *testing.T, s *server, path string) string {
	t.Helper()
	m := regexp.MustCompile(`(?s)<p class="count">(.*?)</p>`).FindStringSubmatch(mustGet(t, s, path))
	if m == nil {
		t.Fatalf("GET %s: no count line", path)
	}
	return html.UnescapeString(regexp.MustCompile(`<[^>]+>`).ReplaceAllString(m[1], " "))
}

// TestStableUrlsKeepWorking asserts the public interface tech.stableUrls
// names: every stable path answers, and every stable query parameter is still
// read, with a value the pages' own links write. A parameter counts as read
// when the index's count line changes because of it; the whole page would
// change anyway, since every link on it carries the query.
//
// Verifies: tech.stableUrls. Criterion: acceptance.stableUrlsKeepWorking.
func TestStableUrlsKeepWorking(t *testing.T) {
	s := newTestServer(t)
	for _, p := range []string{
		"/", "/species/amecro", "/compare?species=amecro&species=annhum",
		"/docs", "/docs/using", "/docs/data", "/docs/pipeline", "/docs/building",
		"/added", "/excluded",
	} {
		if code, _ := get(t, s, p); code != http.StatusOK {
			t.Errorf("GET %s = %d, want 200; tech.stableUrls promises it", p, code)
		}
	}
	w := httptest.NewRecorder()
	s.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/now?region=NY", nil))
	if loc := w.Header().Get("Location"); w.Code != http.StatusSeeOther ||
		!strings.HasPrefix(loc, "/?") || !strings.Contains(loc, "month=") || !strings.Contains(loc, "region=NY") {
		t.Errorf("GET /now?region=NY = %d to %q, want a 303 to the index for this month in NY", w.Code, loc)
	}

	base := countLine(t, s, "/")
	for param, value := range map[string]string{
		"by": "mass", "scale": "global", "step": "30", "units": "metric", "names": "world",
		"species": "amecro", "family": "Corvidae", "region": "NY", "month": "Oct",
		"habitat": "Marine", "niche": "Frugivore",
	} {
		if got := countLine(t, s, "/?"+param+"="+value); got == base {
			t.Errorf("?%s=%s leaves the index's settings as they were; the parameter is no longer read", param, value)
		}
	}
}

// TestCanonicalHostRedirects asserts tech.oneCanonicalHost: given a canonical
// host, a request for any other host is sent there permanently, path and
// query intact; the canonical host itself is served, whatever its case or
// port; and with no canonical host, every host is served.
//
// Verifies: tech.oneCanonicalHost. Criterion: acceptance.canonicalHostRedirects.
func TestCanonicalHostRedirects(t *testing.T) {
	s := newTestServer(t)
	request := func(host, target string) *httptest.ResponseRecorder {
		r := httptest.NewRequest(http.MethodGet, target, nil)
		r.Host = host
		w := httptest.NewRecorder()
		s.ServeHTTP(w, r)
		return w
	}

	for _, host := range []string{"www.birdview.site", "birdview.site", "localhost:8080"} {
		if w := request(host, "/species/amecro"); w.Code != http.StatusOK {
			t.Errorf("with no canonical host, %s answered %d, want 200", host, w.Code)
		}
	}

	s.canonical = "birdview.site"
	for _, host := range []string{"www.birdview.site", "birdview-570018079643.us-central1.run.app", "localhost:8080"} {
		w := request(host, "/species/amecro?units=metric&family=Corvidae")
		want := "https://birdview.site/species/amecro?units=metric&family=Corvidae"
		if w.Code != http.StatusMovedPermanently || w.Header().Get("Location") != want {
			t.Errorf("%s answered %d to %q, want 301 to %q", host, w.Code, w.Header().Get("Location"), want)
		}
	}
	for _, host := range []string{"birdview.site", "BirdView.Site", "birdview.site:443"} {
		if w := request(host, "/species/amecro"); w.Code != http.StatusOK {
			t.Errorf("the canonical host, asked for as %s, answered %d, want 200", host, w.Code)
		}
	}
}
