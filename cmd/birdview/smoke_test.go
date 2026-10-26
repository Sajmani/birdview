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
	"net/url"
	"strings"
	"testing"

	"birdview/internal/smoke"
)

// failures returns the names of the checks that failed, with their errors.
func failures(results []smoke.Result) map[string]string {
	out := map[string]string{}
	for _, r := range results {
		if r.Err != nil {
			out[r.Name] = r.Err.Error()
		}
	}
	return out
}

// TestSmokeChecksTheServer runs the post-deploy smoke checks against this
// server over real HTTP, with a second server redirecting to it as a
// canonical host would. They must all pass, and each kind of breakage a
// deploy could bring must fail the check written for it.
//
// Verifies: tech.smokeCheckAfterDeploy. Criterion: acceptance.smokeChecksDeployment.
func TestSmokeChecksTheServer(t *testing.T) {
	site := httptest.NewServer(newTestServer(t))
	defer site.Close()
	u, _ := url.Parse(site.URL)
	other := newTestServer(t)
	other.canonical = u.Host
	mirror := httptest.NewServer(other)
	defer mirror.Close()

	results := smoke.Run(site.URL, []string{mirror.URL}, site.Client())
	if len(results) < 8 {
		t.Fatalf("smoke ran %d checks, want at least 8: %+v", len(results), results)
	}
	if f := failures(results); len(f) > 0 {
		t.Fatalf("smoke checks failed against a healthy server: %v", f)
	}

	// Each breakage fails the check written for it. The handler wraps the
	// healthy server and changes one thing.
	healthy := newTestServer(t)
	for _, c := range []struct {
		name   string
		check  string
		break_ func(http.ResponseWriter, *http.Request) bool
	}{
		{"robots.txt gone", "robots.txt", func(w http.ResponseWriter, r *http.Request) bool {
			if r.URL.Path == "/robots.txt" {
				http.NotFound(w, r)
				return true
			}
			return false
		}},
		{"home page not compressed", "the home page", func(w http.ResponseWriter, r *http.Request) bool {
			if r.URL.Path == "/" {
				r.Header.Del("Accept-Encoding")
			}
			return false
		}},
		{"home page cached", "the home page", func(w http.ResponseWriter, r *http.Request) bool {
			if r.URL.Path == "/" {
				r.URL.RawQuery = "month=Oct" // a named month is cacheable
			}
			return false
		}},
		{"classic view not revalidated", "the classic view", func(w http.ResponseWriter, r *http.Request) bool {
			if r.URL.Path == "/classic" {
				r.Header.Del("If-None-Match")
			}
			return false
		}},
		{"unknown species served", "an unknown species is not found", func(w http.ResponseWriter, r *http.Request) bool {
			if strings.HasPrefix(r.URL.Path, "/species/no-such") {
				w.Write([]byte("ok"))
				return true
			}
			return false
		}},
		{"guide gone", "a guide (/docs/using)", func(w http.ResponseWriter, r *http.Request) bool {
			if r.URL.Path == "/docs/using" {
				http.NotFound(w, r)
				return true
			}
			return false
		}},
	} {
		broken := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !c.break_(w, r) {
				healthy.ServeHTTP(w, r)
			}
		}))
		f := failures(smoke.Run(broken.URL, nil, broken.Client()))
		broken.Close()
		if _, ok := f[c.check]; !ok {
			t.Errorf("with %s, check %q passed; failures were %v", c.name, c.check, f)
		}
	}

	// A host that serves the site instead of redirecting fails the redirect check.
	notRedirecting := httptest.NewServer(newTestServer(t))
	defer notRedirecting.Close()
	f := failures(smoke.Run(site.URL, []string{notRedirecting.URL}, site.Client()))
	if _, ok := f["redirect from "+notRedirecting.URL]; !ok {
		t.Errorf("a host that does not redirect passed the redirect check; failures were %v", f)
	}
}
