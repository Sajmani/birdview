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
	"net/http"
	"net/http/httptest"
	"os/exec"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"birdview/internal/boundary"
	"birdview/internal/dataset"
	"birdview/internal/month"
)

// hereCountyTestServer builds a server with -counties enabled and counties in
// NY (Kings), CA (Los Angeles), and QC (Montréal) so Here's second stage can
// resolve and render them.
func hereCountyTestServer(t *testing.T) *server {
	t.Helper()
	countiesOn(t)
	ds, cs := countyDataset()
	extra := []dataset.County{
		{ID: "USA.33.24_1", Name: "Kings", State: "USA.33_1", Total: 200000, Monthly: [12]int64{4: 200000}},
		{ID: "CAN.11.20_1", Name: "Communauté-Urbaine-de-Montréal", State: "CAN.11_1", Total: 150000, Monthly: [12]int64{4: 150000}},
	}
	cs = testCounties(ds, append(cs.Counties, extra...), map[string]map[string][12]int64{
		"USA.5.19_1":  {"annhum": {4: 3000}},
		"USA.33.24_1": {"amecro": {4: 600}},
		"CAN.11.20_1": {"amecro": {4: 600}},
	})
	s, err := newServer(ds, cs)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

// TestCountyBoundaryRoute checks the per-state county boundary route: 200 with
// immutable Cache-Control under -counties, 404 without -counties or for a wrong
// hash or unknown state, and 403 for disallowed crawlers.
//
// Verifies: app.herePinpointsCounty.
func TestCountyBoundaryRoute(t *testing.T) {
	s := hereCountyTestServer(t)
	body := mustGet(t, s, "/?month=all")
	here := mustMatch(t, body, `(<button type="button" class="btn here[^"]*"[^>]*>)`, "the Here button")
	tmpl := mustMatch(t, here, `data-county-boundaries="([^"]+)"`, "the county boundary URL template")
	if !strings.HasSuffix(tmpl, "/"+stateSlot+".json") {
		t.Fatalf("data-county-boundaries = %q, want suffix /%s.json", tmpl, stateSlot)
	}

	nyURL := strings.Replace(tmpl, stateSlot, "USA.33_1", 1)
	w := httptest.NewRecorder()
	s.ServeHTTP(w, httptest.NewRequest(http.MethodGet, nyURL, nil))
	if w.Code != http.StatusOK {
		t.Fatalf("GET %s = %d, want 200", nyURL, w.Code)
	}
	if ct := w.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("GET %s Content-Type = %q, want application/json", nyURL, ct)
	}
	if cc := w.Header().Get("Cache-Control"); cc != "public, max-age=31536000, immutable" {
		t.Errorf("GET %s Cache-Control = %q, want immutable", nyURL, cc)
	}
	f, err := boundary.Parse(w.Body.Bytes())
	if err != nil {
		t.Fatalf("GET %s returned invalid boundary JSON: %v", nyURL, err)
	}
	if got := f.Lookup(40.6501, -73.9496); got != "USA.33.24_1" {
		t.Errorf("NY county boundaries resolve Brooklyn to %q, want USA.33.24_1", got)
	}

	// Wrong hash, unknown state, and nested path must 404.
	for _, bad := range []string{
		"/boundaries/0000000000000000/USA.33_1.json",
		strings.Replace(tmpl, stateSlot, "ZZZ.99_1", 1),
		strings.Replace(tmpl, stateSlot, "sub/USA.33_1", 1),
	} {
		if code, _ := get(t, s, bad); code != http.StatusNotFound {
			t.Errorf("GET %s = %d, want 404", bad, code)
		}
	}

	// Disallowed crawlers get 403 under /boundaries/.
	req := httptest.NewRequest(http.MethodGet, nyURL, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; AhrefsBot/7.0)")
	cw := httptest.NewRecorder()
	s.ServeHTTP(cw, req)
	if cw.Code != http.StatusForbidden {
		t.Errorf("crawler GET %s = %d, want 403", nyURL, cw.Code)
	}

	// Without -counties, no data-county-boundaries attribute is emitted and the
	// county boundary path answers 404.
	plain := newTestServer(t)
	plainBody := mustGet(t, plain, "/?month=all")
	plainHere := mustMatch(t, plainBody, `(<button type="button" class="btn here[^"]*"[^>]*>)`, "the Here button")
	if strings.Contains(plainHere, "data-county-boundaries") {
		t.Errorf("without -counties, Here button carries data-county-boundaries: %s", plainHere)
	}
	if code, _ := get(t, plain, nyURL); code != http.StatusNotFound {
		t.Errorf("without -counties, GET %s = %d, want 404", nyURL, code)
	}
}

// TestFooterCreditsCountyBoundaries checks that the home page footer credits
// the U.S. Census Bureau and carries Statistics Canada's verbatim Value-added
// Product attribution notice.
//
// Verifies: app.herePinpointsCounty.
func TestFooterCreditsCountyBoundaries(t *testing.T) {
	s := newTestServer(t)
	body := mustGet(t, s, "/")
	footer := mustMatch(t, body, `(?s)<footer id="credits">(.*?)</footer>`, "the credits footer")
	for _, want := range []string{
		"U.S. Census Bureau",
		"2021 Cartographic Boundary Files",
		"Adapted from Statistics Canada, 2021 Census – Boundary files (Census divisions), 2021.",
		"This does not constitute an endorsement by Statistics Canada of this product.",
	} {
		if !strings.Contains(footer, want) {
			t.Errorf("credits footer does not contain %q", want)
		}
	}
}

// TestHerePinpointsCountyInTheBrowser drives Here and #here-and-now in headless
// Chrome with -counties enabled (and without it), checking that:
//   - pressing Here in Kings NY, Los Angeles CA, and Montréal QC fetches only
//     that single state's county boundary file and navigates to the county;
//   - #here-and-now pinpoints the county as well;
//   - a point inside a state's border but outside county polygons (Lake Ontario
//     NY) falls back cleanly to the state;
//   - without -counties, Here stops at the state and fetches no county file;
//   - exact coordinates never appear in any request URL.
//
// Verifies: app.herePinpointsCounty.
func TestHerePinpointsCountyInTheBrowser(t *testing.T) {
	chrome := findChrome()
	if chrome == "" {
		t.Skip("no Chrome or Chromium found (set BIRDVIEW_CHROME); county Here was not tested in a browser")
	}
	countySrv := hereCountyTestServer(t)
	plainSrv := newTestServer(t)
	thisMonth := month.Abbrev[time.Now().Month()-1]
	const start = "/?by=mass&month=May"

	cases := []struct {
		name           string
		srv            *server
		frag           string
		wantLand       string
		wantCountyFile string // "" means no county boundary file may be fetched
		leaks          []string
	}{
		{
			name:           "Here in Kings NY",
			srv:            countySrv,
			frag:           "#probe=1&click=40.6501,-73.9496",
			wantLand:       "/?by=mass&month=May&region=USA.33.24_1",
			wantCountyFile: countySrv.countyBoundaryPrefix + "USA.33_1.json",
			leaks:          []string{"40.65", "73.94", "here", "lat"},
		},
		{
			name:           "Here in Los Angeles CA",
			srv:            countySrv,
			frag:           "#probe=1&click=34.0522,-118.2437",
			wantLand:       "/?by=mass&month=May&region=USA.5.19_1",
			wantCountyFile: countySrv.countyBoundaryPrefix + "USA.5_1.json",
			leaks:          []string{"34.05", "118.24", "here", "lat"},
		},
		{
			name:           "Here in Montreal QC",
			srv:            countySrv,
			frag:           "#probe=1&click=45.5017,-73.5673",
			wantLand:       "/?by=mass&month=May&region=CAN.11.20_1",
			wantCountyFile: countySrv.countyBoundaryPrefix + "CAN.11_1.json",
			leaks:          []string{"45.50", "73.56", "here", "lat"},
		},
		{
			name:           "here-and-now in Kings NY",
			srv:            countySrv,
			frag:           "#here-and-now&probe=1&geo=40.6501,-73.9496",
			wantLand:       "/?by=mass&month=" + thisMonth + "&region=USA.33.24_1",
			wantCountyFile: countySrv.countyBoundaryPrefix + "USA.33_1.json",
			leaks:          []string{"40.65", "73.94", "here", "lat"},
		},
		{
			name:           "Lake Ontario NY falls back to state",
			srv:            countySrv,
			frag:           "#probe=1&click=43.50,-77.50",
			wantLand:       "/?by=mass&month=May&region=USA.33_1",
			wantCountyFile: countySrv.countyBoundaryPrefix + "USA.33_1.json",
			leaks:          []string{"43.50", "77.50", "here", "lat"},
		},
		{
			name:           "without -counties stops at state",
			srv:            plainSrv,
			frag:           "#probe=1&click=40.6501,-73.9496",
			wantLand:       "/?by=mass&month=May&region=USA.33_1",
			wantCountyFile: "",
			leaks:          []string{"40.65", "73.94", "here", "lat"},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			var mu sync.Mutex
			var requests []string
			ts := httptest.NewServer(instrument(c.srv, hereProbe, func(uri string) {
				mu.Lock()
				requests = append(requests, uri)
				mu.Unlock()
			}))
			defer ts.Close()

			ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
			defer cancel()
			cmd := exec.CommandContext(ctx, chrome, "--headless", "--disable-gpu",
				"--no-first-run", "--no-default-browser-check", "--disable-background-networking",
				"--disable-component-update", "--disable-sync", "--disable-extensions",
				"--user-data-dir="+profileDir(t), "--virtual-time-budget=3000",
				"--dump-dom", ts.URL+start+c.frag)
			if err := cmd.Start(); err != nil {
				t.Fatalf("chrome: %v", err)
			}
			deadline := time.Now().Add(20 * time.Second)
			for time.Now().Before(deadline) {
				mu.Lock()
				arrived := slices.Contains(requests, c.wantLand)
				mu.Unlock()
				if arrived {
					time.Sleep(500 * time.Millisecond)
					break
				}
				time.Sleep(100 * time.Millisecond)
			}
			cmd.Process.Kill()
			cmd.Wait()

			mu.Lock()
			defer mu.Unlock()
			for _, q := range requests {
				for _, leak := range c.leaks {
					if strings.Contains(strings.ToLower(q), leak) {
						t.Errorf("request %q leaked %q", q, leak)
					}
				}
			}

			var pages, countyFetches []string
			for _, q := range requests {
				switch {
				case strings.HasPrefix(q, "/static/"), q == c.srv.boundaryURL:
				case strings.HasPrefix(q, "/boundaries/"):
					countyFetches = append(countyFetches, q)
				default:
					pages = append(pages, q)
				}
			}
			if len(pages) < 2 || pages[1] != c.wantLand {
				t.Errorf("navigated to pages %v, want %s after %s (all requests: %v)", pages, c.wantLand, start, requests)
			}
			if c.wantCountyFile == "" {
				if len(countyFetches) != 0 {
					t.Errorf("fetched county boundaries %v, want none without -counties", countyFetches)
				}
			} else {
				if len(countyFetches) != 1 || countyFetches[0] != c.wantCountyFile {
					t.Errorf("fetched county boundaries %v, want only [%s]", countyFetches, c.wantCountyFile)
				}
			}
		})
	}
}
