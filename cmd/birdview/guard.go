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
	"net/url"
	"strings"
	"sync"
	"time"
)

// disallowedCrawlerTokens are User-Agent substrings of crawlers and scrapers
// that do not belong on URLs disallowed by /robots.txt (query strings and
// /boundaries/), checked case-insensitively (tech.crawlersSkipFilteredViews).
// Link-preview fetchers (such as facebookexternalhit), uptime checks, and
// birdview-smoke are deliberately excluded.
var disallowedCrawlerTokens = []string{
	"meta-externalagent",
	"bytespider",
	"gptbot",
	"claudebot",
	"ccbot",
	"amazonbot",
	"petalbot",
	"ahrefsbot",
	"semrushbot",
	"dotbot",
	"mj12bot",
	"blexbot",
	"dataforseobot",
	"yandexbot",
	"baiduspider",
	"sogou",
	"oai-searchbot",
	"perplexitybot",
	"applebot-extended",
	"scrapy",
	"crawler",
	"spider",
	"scraper",
}

func isDisallowedCrawler(ua string) bool {
	if ua == "" {
		return false
	}
	lower := strings.ToLower(ua)
	for _, tok := range disallowedCrawlerTokens {
		if strings.Contains(lower, tok) {
			return true
		}
	}
	return false
}

// isRobotsDisallowed reports whether r asks for a URL /robots.txt disallows
// (any URL with a query string, or /boundaries/).
func isRobotsDisallowed(r *http.Request) bool {
	return r.URL.RawQuery != "" || strings.HasPrefix(r.URL.Path, "/boundaries/")
}

// hasMalformedQuery reports a query string corrupted by naive HTML-entity
// decoding (e.g. &region= turned into ®ion= / %C2%AEion= by a scraper).
func hasMalformedQuery(raw string) bool {
	if raw == "" {
		return false
	}
	if strings.Contains(raw, "®") || strings.Contains(strings.ToLower(raw), "%c2%ae") {
		return true
	}
	if decoded, err := url.QueryUnescape(raw); err == nil && strings.Contains(decoded, "®") {
		return true
	}
	return false
}

// isLargeNextView reports whether q asks the home page to draw more than the
// default cap (show=all or a numeric show > nextCap).
func isLargeNextView(q url.Values) bool {
	mode, capN, ok := nextShowOf(q)
	if !ok {
		return false
	}
	return mode == "all" || (mode == "first" && capN > nextCap)
}

// enforceCrawlers enforces /robots.txt on disallowed URLs for self-identified
// crawlers, rejects &reg-corrupted queries, and stamps X-Robots-Tag on
// filtered home views (tech.crawlersSkipFilteredViews). Runs before
// revalidated so rejections carry Cache-Control: no-store and no ETag.
func (s *server) enforceCrawlers(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if isRobotsDisallowed(r) && isDisallowedCrawler(r.Header.Get("User-Agent")) {
			w.Header().Set("Cache-Control", "no-store")
			http.Error(w, "Disallowed by /robots.txt.", http.StatusForbidden)
			return
		}
		if r.URL.Path == "/" && hasMalformedQuery(r.URL.RawQuery) {
			w.Header().Set("Cache-Control", "no-store")
			http.Error(w, "Malformed query string.", http.StatusBadRequest)
			return
		}
		if r.URL.Path == "/" && r.URL.RawQuery != "" {
			w.Header().Set("X-Robots-Tag", "noindex, nofollow")
		}
		h.ServeHTTP(w, r)
	})
}

// tokenBucket holds concurrency and token-bucket rate state for one class of
// rendered views. rate == 0 disables rate limiting while keeping concurrency
// bounds active.
type tokenBucket struct {
	maxIn    int
	inFlight int
	rate     float64
	burst    float64
	tokens   float64
	last     time.Time
}

func (b *tokenBucket) refill(now time.Time) {
	if b.rate <= 0 {
		return
	}
	if !b.last.IsZero() {
		if dt := now.Sub(b.last).Seconds(); dt > 0 {
			b.tokens += dt * b.rate
			if b.tokens > b.burst {
				b.tokens = b.burst
			}
		}
	}
	b.last = now
}

func (b *tokenBucket) available() bool {
	if b.maxIn > 0 && b.inFlight >= b.maxIn {
		return false
	}
	if b.rate > 0 && b.tokens < 1 {
		return false
	}
	return true
}

func (b *tokenBucket) take() {
	if b.rate > 0 {
		b.tokens--
	}
	b.inFlight++
}

// filteredGate bounds concurrent renders and rate for filtered home views,
// with a tighter bucket for large views (show=all or show > nextCap), so
// crawlers cannot saturate the server or starve query-free pages
// (tech.filteredViewsBounded).
type filteredGate struct {
	mu       sync.Mutex
	filtered tokenBucket
	large    tokenBucket
	now      func() time.Time
}

func newFilteredGate() *filteredGate {
	return &filteredGate{
		// Per instance (under --max-instances=2): at most 2 concurrent
		// filtered renders (4/s sustained, burst 8), and at most 1 concurrent
		// large render (0.5/s sustained, burst 2).
		filtered: tokenBucket{maxIn: 2, rate: 4, burst: 8, tokens: 8},
		large:    tokenBucket{maxIn: 1, rate: 0.5, burst: 2, tokens: 2},
		now:      time.Now,
	}
}

// acquire reserves concurrency and rate capacity atomically across the
// filtered bucket (and the large bucket when large is true). On refusal it
// returns nil without consuming any token; on success the caller must invoke
// the returned release function once rendering finishes.
func (g *filteredGate) acquire(large bool) func() {
	if g == nil {
		return func() {}
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	now := g.now()
	g.filtered.refill(now)
	if large {
		g.large.refill(now)
	}
	if !g.filtered.available() || (large && !g.large.available()) {
		return nil
	}
	g.filtered.take()
	if large {
		g.large.take()
	}
	var once sync.Once
	return func() {
		once.Do(func() {
			g.mu.Lock()
			g.filtered.inFlight--
			if large {
				g.large.inFlight--
			}
			g.mu.Unlock()
		})
	}
}
