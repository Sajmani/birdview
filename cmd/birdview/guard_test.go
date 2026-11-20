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
	"sync"
	"testing"
	"time"
)

// TestCrawlersEnforcedOnFilteredViews checks that filtered home views carry
// X-Robots-Tag: noindex, nofollow, that links to query-string views carry
// rel="nofollow", that self-identified crawlers are rejected on disallowed
// URLs while allowed on query-free pages, and that &reg-corrupted queries are
// rejected.
//
// Verifies: app.crawlersEnforcedOnFilteredViews.
func TestCrawlersEnforcedOnFilteredViews(t *testing.T) {
	s := newTestServer(t)

	// 1. X-Robots-Tag on filtered views, absent on query-free pages.
	for _, path := range []string{"/?by=length", "/?show=all", "/?pin=amecro"} {
		w := httptest.NewRecorder()
		s.ServeHTTP(w, httptest.NewRequest(http.MethodGet, path, nil))
		if w.Code != http.StatusOK {
			t.Fatalf("GET %s = %d, want 200", path, w.Code)
		}
		if got := w.Header().Get("X-Robots-Tag"); got != "noindex, nofollow" {
			t.Errorf("GET %s: X-Robots-Tag = %q, want %q", path, got, "noindex, nofollow")
		}
	}
	for _, path := range []string{"/", "/species/amecro", "/docs/using", "/added", "/excluded"} {
		w := httptest.NewRecorder()
		s.ServeHTTP(w, httptest.NewRequest(http.MethodGet, path, nil))
		if w.Code != http.StatusOK {
			t.Fatalf("GET %s = %d, want 200", path, w.Code)
		}
		if got := w.Header().Get("X-Robots-Tag"); got != "" {
			t.Errorf("GET %s: unexpected X-Robots-Tag %q on query-free page", path, got)
		}
	}

	// 2. Links to query-string views on the home page carry rel="nofollow";
	// links to species pages and guides do not.
	_, home := get(t, s, "/")
	anchorRe := regexp.MustCompile(`<a\b[^>]*>`)
	hrefRe := regexp.MustCompile(`\bhref="([^"]*)"`)
	relRe := regexp.MustCompile(`\brel="([^"]*)"`)
	var queryLinks, pageLinks int
	for _, tag := range anchorRe.FindAllString(home, -1) {
		hm := hrefRe.FindStringSubmatch(tag)
		if hm == nil {
			continue
		}
		href := hm[1]
		var rel string
		if rm := relRe.FindStringSubmatch(tag); rm != nil {
			rel = rm[1]
		}
		hasNofollow := strings.Contains(" "+rel+" ", " nofollow ")
		switch {
		case strings.HasPrefix(href, "/?"):
			queryLinks++
			if !hasNofollow {
				t.Errorf("query-string link %s lacks rel=\"nofollow\"", tag)
			}
		case strings.HasPrefix(href, "/species/") || strings.HasPrefix(href, "/docs/"):
			pageLinks++
			if hasNofollow {
				t.Errorf("indexable link %s unexpectedly carries rel=\"nofollow\"", tag)
			}
		}
	}
	if queryLinks < 10 || pageLinks < 3 {
		t.Fatalf("parsed %d query links and %d indexable links on /; too few to check", queryLinks, pageLinks)
	}

	// 3. Self-identified crawlers get 403 on disallowed URLs and 200 on
	// allowed query-free pages.
	crawlers := []string{
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/145.0.0.0 Safari/537.36 (compatible; meta-externalagent/1.1 (+https://developers.facebook.com/docs/sharing/webmasters/crawler))",
		"Mozilla/5.0 (compatible; GPTBot/1.0; +https://openai.com/gptbot)",
		"Bytespider",
		"Scrapy/2.11",
	}
	disallowed := []string{"/?by=length", "/?show=all&pin=amecro", s.boundaryURL}
	allowed := []string{"/", "/species/amecro", "/docs/using", "/robots.txt"}
	for _, ua := range crawlers {
		for _, path := range disallowed {
			req := httptest.NewRequest(http.MethodGet, path, nil)
			req.Header.Set("User-Agent", ua)
			w := httptest.NewRecorder()
			s.ServeHTTP(w, req)
			if w.Code != http.StatusForbidden {
				t.Errorf("GET %s with crawler UA %q = %d, want 403", path, ua, w.Code)
			}
			if cc := w.Header().Get("Cache-Control"); cc != "no-store" {
				t.Errorf("GET %s 403 Cache-Control = %q, want no-store", path, cc)
			}
			if etag := w.Header().Get("ETag"); etag != "" {
				t.Errorf("GET %s 403 set unexpected ETag %q", path, etag)
			}
		}
		for _, path := range allowed {
			req := httptest.NewRequest(http.MethodGet, path, nil)
			req.Header.Set("User-Agent", ua)
			w := httptest.NewRecorder()
			s.ServeHTTP(w, req)
			if w.Code != http.StatusOK {
				t.Errorf("GET %s with crawler UA %q = %d, want 200", path, ua, w.Code)
			}
		}
	}

	// Normal browsers, monitoring, and birdview-smoke are not blocked.
	legit := []struct {
		ua   string
		path string
		want int
	}{
		{"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/154.0.0.0 Safari/537.36", "/?by=length", http.StatusOK},
		{"GoogleStackdriverMonitoring-UptimeChecks(https://cloud.google.com/monitoring)", "/species/amecro", http.StatusOK},
		{"birdview-smoke/1.0", "/?month=now", http.StatusSeeOther},
		{"birdview-smoke/1.0", s.boundaryURL, http.StatusOK},
		{"facebookexternalhit/1.1", "/?by=length&pin=amecro", http.StatusOK},
	}
	for _, tc := range legit {
		req := httptest.NewRequest(http.MethodGet, tc.path, nil)
		req.Header.Set("User-Agent", tc.ua)
		w := httptest.NewRecorder()
		s.ServeHTTP(w, req)
		if w.Code != tc.want {
			t.Errorf("GET %s with UA %q = %d, want %d", tc.path, tc.ua, w.Code, tc.want)
		}
	}

	// 4. Malformed &reg-corrupted queries return 400 with Cache-Control: no-store.
	for _, bad := range []string{
		"/?family=Procellariidae&pin=bulshe%C2%AEion%3DPalmyraAtoll",
		"/?pin=amgplo&pin=pipplo%c2%aeion%3DSD",
	} {
		w := httptest.NewRecorder()
		s.ServeHTTP(w, httptest.NewRequest(http.MethodGet, bad, nil))
		if w.Code != http.StatusBadRequest {
			t.Errorf("GET %s = %d, want 400", bad, w.Code)
		}
		if cc := w.Header().Get("Cache-Control"); cc != "no-store" {
			t.Errorf("GET %s 400 Cache-Control = %q, want no-store", bad, cc)
		}
		if etag := w.Header().Get("ETag"); etag != "" {
			t.Errorf("GET %s 400 set unexpected ETag %q", bad, etag)
		}
	}
}

// TestFilteredViewsShedLoad checks that concurrent large and filtered home
// renders shed excess load immediately with 429 without starving query-free
// pages, redirects, or revalidations.
//
// Verifies: app.filteredViewsShedLoad.
func TestFilteredViewsShedLoad(t *testing.T) {
	s := newTestServer(t)
	s.gate = newFilteredGate()
	s.gate.filtered.rate = 0
	s.gate.large.rate = 0

	// Prime an ETag for a filtered view so we can check 304 revalidation
	// stays fast even while render slots are full.
	prime := httptest.NewRecorder()
	s.ServeHTTP(prime, httptest.NewRequest(http.MethodGet, "/?by=mass", nil))
	if prime.Code != http.StatusOK || prime.Header().Get("ETag") == "" {
		t.Fatalf("priming /?by=mass: code=%d etag=%q", prime.Code, prime.Header().Get("ETag"))
	}
	massETag := prime.Header().Get("ETag")

	// Hold specific in-flight renders via onRender without blocking probes
	// that should be rejected (so an unpatched server fails loudly with 200
	// rather than deadlocking).
	entered := make(chan struct{}, 4)
	release := make(chan struct{})
	var mu sync.Mutex
	holdNext := 1
	s.onRender = func(name string) {
		mu.Lock()
		h := holdNext > 0 && name == "next.html"
		if h {
			holdNext--
		}
		mu.Unlock()
		if h {
			entered <- struct{}{}
			<-release
		}
	}

	var wg sync.WaitGroup
	wg.Add(1)
	first := httptest.NewRecorder()
	go func() {
		defer wg.Done()
		s.ServeHTTP(first, httptest.NewRequest(http.MethodGet, "/?show=all", nil))
	}()
	<-entered

	// A second concurrent show=all must fail fast with 429, Retry-After, no-store, no ETag.
	second := httptest.NewRecorder()
	s.ServeHTTP(second, httptest.NewRequest(http.MethodGet, "/?show=all&by=length", nil))
	if second.Code != http.StatusTooManyRequests {
		t.Errorf("concurrent /?show=all = %d, want 429", second.Code)
	}
	if second.Header().Get("Retry-After") == "" {
		t.Errorf("429 response missing Retry-After header")
	}
	if cc := second.Header().Get("Cache-Control"); cc != "no-store" {
		t.Errorf("429 Cache-Control = %q, want no-store", cc)
	}
	if etag := second.Header().Get("ETag"); etag != "" {
		t.Errorf("429 response set unexpected ETag %q", etag)
	}

	// Fill the remaining filtered slot with a normal filtered view so all
	// filtered slots are occupied.
	mu.Lock()
	holdNext = 1
	mu.Unlock()
	wg.Add(1)
	filtered2 := httptest.NewRecorder()
	go func() {
		defer wg.Done()
		s.ServeHTTP(filtered2, httptest.NewRequest(http.MethodGet, "/?by=length", nil))
	}()
	<-entered

	// Another filtered view now fails fast with 429.
	filtered3 := httptest.NewRecorder()
	s.ServeHTTP(filtered3, httptest.NewRequest(http.MethodGet, "/?by=bill", nil))
	if filtered3.Code != http.StatusTooManyRequests {
		t.Errorf("3rd concurrent filtered view = %d, want 429", filtered3.Code)
	}

	// Meanwhile, query-free pages, redirects, and 304 revalidations still succeed.
	for _, tc := range []struct {
		path string
		want int
	}{
		{"/", http.StatusOK},
		{"/species/amecro", http.StatusOK},
		{"/docs/using", http.StatusOK},
		{"/robots.txt", http.StatusOK},
		{"/?month=now", http.StatusSeeOther},
	} {
		w := httptest.NewRecorder()
		s.ServeHTTP(w, httptest.NewRequest(http.MethodGet, tc.path, nil))
		if w.Code != tc.want {
			t.Errorf("while filtered slots full: GET %s = %d, want %d", tc.path, w.Code, tc.want)
		}
	}
	revalReq := httptest.NewRequest(http.MethodGet, "/?by=mass", nil)
	revalReq.Header.Set("If-None-Match", massETag)
	reval := httptest.NewRecorder()
	s.ServeHTTP(reval, revalReq)
	if reval.Code != http.StatusNotModified {
		t.Errorf("while filtered slots full: revalidated /?by=mass = %d, want 304", reval.Code)
	}

	// Release held renders and verify filtered views succeed again.
	close(release)
	wg.Wait()
	if first.Code != http.StatusOK || filtered2.Code != http.StatusOK {
		t.Errorf("held renders returned %d and %d, want 200", first.Code, filtered2.Code)
	}
	after := httptest.NewRecorder()
	s.ServeHTTP(after, httptest.NewRequest(http.MethodGet, "/?show=all", nil))
	if after.Code != http.StatusOK {
		t.Errorf("after release: GET /?show=all = %d, want 200", after.Code)
	}

	// Rate limit: production gate with a controlled clock bounds both
	// large views (show=all, burst 2, 0.5/s) and normal filtered views
	// (burst 8, 4/s), without touching query-free pages.
	rs := newTestServer(t)
	now := time.Date(2026, 11, 20, 12, 0, 0, 0, time.UTC)
	rs.gate = newFilteredGate()
	rs.gate.now = func() time.Time { return now }

	// Large burst is 2: first two succeed, third returns 429.
	for i := range 2 {
		w := httptest.NewRecorder()
		rs.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/?show=all", nil))
		if w.Code != http.StatusOK {
			t.Fatalf("large burst request %d = %d, want 200", i+1, w.Code)
		}
	}
	overLarge := httptest.NewRecorder()
	rs.ServeHTTP(overLarge, httptest.NewRequest(http.MethodGet, "/?show=all", nil))
	if overLarge.Code != http.StatusTooManyRequests {
		t.Errorf("3rd /?show=all past burst = %d, want 429", overLarge.Code)
	}
	if overLarge.Header().Get("Retry-After") == "" {
		t.Errorf("rate-limited 429 missing Retry-After")
	}
	if cc := overLarge.Header().Get("Cache-Control"); cc != "no-store" {
		t.Errorf("rate-limited 429 Cache-Control = %q, want no-store", cc)
	}

	// Numeric show > nextCap also hits the large bucket.
	overCap := httptest.NewRecorder()
	rs.ServeHTTP(overCap, httptest.NewRequest(http.MethodGet, "/?show=500", nil))
	if overCap.Code != http.StatusTooManyRequests {
		t.Errorf("/?show=500 while large bucket empty = %d, want 429", overCap.Code)
	}

	// Advancing 2s refills 1 large token (0.5/s), so show=all succeeds again.
	now = now.Add(2 * time.Second)
	refilled := httptest.NewRecorder()
	rs.ServeHTTP(refilled, httptest.NewRequest(http.MethodGet, "/?show=all", nil))
	if refilled.Code != http.StatusOK {
		t.Errorf("/?show=all after 2s refill = %d, want 200", refilled.Code)
	}

	// Exhaust the filtered bucket (burst 8); subsequent filtered views get 429,
	// while query-free / still returns 200, and advancing 250ms refills 1 token.
	rs.gate = newFilteredGate()
	rs.gate.now = func() time.Time { return now }
	for i := range 8 {
		w := httptest.NewRecorder()
		rs.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/?by=length", nil))
		if w.Code != http.StatusOK {
			t.Fatalf("filtered burst request %d = %d, want 200", i+1, w.Code)
		}
	}
	overFiltered := httptest.NewRecorder()
	rs.ServeHTTP(overFiltered, httptest.NewRequest(http.MethodGet, "/?by=length", nil))
	if overFiltered.Code != http.StatusTooManyRequests {
		t.Errorf("9th /?by=length past burst = %d, want 429", overFiltered.Code)
	}
	home := httptest.NewRecorder()
	rs.ServeHTTP(home, httptest.NewRequest(http.MethodGet, "/", nil))
	if home.Code != http.StatusOK {
		t.Errorf("GET / while filtered bucket empty = %d, want 200", home.Code)
	}
	now = now.Add(250 * time.Millisecond)
	okAgain := httptest.NewRecorder()
	rs.ServeHTTP(okAgain, httptest.NewRequest(http.MethodGet, "/?by=length", nil))
	if okAgain.Code != http.StatusOK {
		t.Errorf("/?by=length after 250ms refill = %d, want 200", okAgain.Code)
	}
}
