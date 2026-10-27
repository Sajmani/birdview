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

// Package smoke checks a running BirdView deployment from outside, the way a
// reader reaches it (tech.smokeCheckAfterDeploy).
//
// It is meant to run after every deploy, against the public site, so it is
// careful with egress: the index, the one large page, is fetched once and
// gzipped; the boundary file is only asked about with HEAD; everything else
// is a page of a few kilobytes or a redirect with no body.
package smoke

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

// UserAgent names the checker in the site's request logs.
const UserAgent = "birdview-smoke/1.0"

// Result is the outcome of one check.
type Result struct {
	Name string
	Err  error // nil when the check passed
}

// Run checks the deployment at base, an origin such as https://birdview.site,
// and, for each URL in redirectFrom, that it redirects permanently to the same
// path and query on base's host over HTTPS (tech.oneCanonicalHost). client may
// be nil.
func Run(base string, redirectFrom []string, client *http.Client) []Result {
	c := &checker{base: strings.TrimSuffix(base, "/")}
	if client == nil {
		client = &http.Client{Timeout: 30 * time.Second}
	}
	// Redirects are checked, not followed: following one would hide it.
	cl := *client
	cl.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	c.client = &cl

	index := c.index()
	c.revalidated()
	c.page("a species page", firstMatch(speciesLink, index, "/species/amecro"), "text/html")
	c.page("a guide", "/docs/using", "text/html")
	c.robots()
	c.unknownSpecies()
	c.boundaryFile(firstMatch(boundaryLink, index, ""))
	for _, from := range redirectFrom {
		c.redirect(from)
	}
	return c.results
}

type checker struct {
	base    string
	client  *http.Client
	results []Result
}

var (
	speciesLink  = regexp.MustCompile(`href="(/species/[a-z0-9]+)"`)
	boundaryLink = regexp.MustCompile(`"(/boundaries/[0-9a-f]+\.json)"`)
)

func firstMatch(re *regexp.Regexp, body, fallback string) string {
	if m := re.FindStringSubmatch(body); m != nil {
		return m[1]
	}
	return fallback
}

func (c *checker) record(name string, err error) {
	c.results = append(c.results, Result{Name: name, Err: err})
}

func (c *checker) get(method, path string, header http.Header) (*http.Response, []byte, error) {
	req, err := http.NewRequest(method, c.base+path, nil)
	if err != nil {
		return nil, nil, err
	}
	for k, v := range header {
		req.Header[k] = v
	}
	req.Header.Set("User-Agent", UserAgent)
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	return resp, body, err
}

// index fetches the home page once, gzipped, checks how it was served, and
// returns its body for the checks that follow links from it. With no month
// named it answers from the clock, so it is never stored (product.nextView).
func (c *checker) index() string {
	// The transport asks for gzip and decompresses, reporting it in
	// Uncompressed; asking explicitly would turn that off.
	resp, body, err := c.get(http.MethodGet, "/", nil)
	if err != nil {
		c.record("the home page", err)
		return ""
	}
	var problems []string
	if resp.StatusCode != http.StatusOK {
		problems = append(problems, fmt.Sprintf("status %d, want 200", resp.StatusCode))
	}
	if !resp.Uncompressed {
		problems = append(problems, "not gzipped (tech.compressedResponses)")
	}
	if !strings.Contains(resp.Header.Get("Vary"), "Accept-Encoding") {
		problems = append(problems, "no Vary: Accept-Encoding")
	}
	if resp.Header.Get("Cache-Control") != "no-store" {
		problems = append(problems, fmt.Sprintf("Cache-Control %q, want no-store: with no month it answers from the clock", resp.Header.Get("Cache-Control")))
	}
	c.record("the home page", joined(problems))
	return string(body)
}

// revalidated fetches the home page for every month, which is a pure
// function of its address, and checks that it is revalidated rather than
// stored.
func (c *checker) revalidated() {
	const path, name = "/?month=all", "the home page for every month"
	resp, _, err := c.get(http.MethodGet, path, nil)
	if err != nil {
		c.record(name, err)
		return
	}
	var problems []string
	if resp.StatusCode != http.StatusOK {
		problems = append(problems, fmt.Sprintf("status %d, want 200", resp.StatusCode))
	}
	if resp.Header.Get("Cache-Control") != "no-cache" {
		problems = append(problems, fmt.Sprintf("Cache-Control %q, want no-cache (tech.cacheLifetimeStated)", resp.Header.Get("Cache-Control")))
	}
	etag := resp.Header.Get("ETag")
	if etag == "" {
		problems = append(problems, "no ETag")
	} else if again, _, err := c.get(http.MethodGet, path, http.Header{"If-None-Match": {etag}}); err != nil {
		problems = append(problems, "revalidating: "+err.Error())
	} else if again.StatusCode != http.StatusNotModified {
		problems = append(problems, fmt.Sprintf("revalidating with its ETag gave %d, want 304", again.StatusCode))
	}
	c.record(name, joined(problems))
}

func (c *checker) page(name, path, ctype string) {
	resp, _, err := c.get(http.MethodGet, path, nil)
	switch {
	case err != nil:
	case resp.StatusCode != http.StatusOK:
		err = fmt.Errorf("GET %s: status %d, want 200", path, resp.StatusCode)
	case !strings.HasPrefix(resp.Header.Get("Content-Type"), ctype):
		err = fmt.Errorf("GET %s: Content-Type %q, want %s", path, resp.Header.Get("Content-Type"), ctype)
	}
	c.record(name+" ("+path+")", err)
}

func (c *checker) robots() {
	resp, body, err := c.get(http.MethodGet, "/robots.txt", nil)
	switch {
	case err != nil:
	case resp.StatusCode != http.StatusOK:
		err = fmt.Errorf("status %d, want 200", resp.StatusCode)
	case !strings.Contains(string(body), "Disallow: /*?"):
		err = fmt.Errorf("the policy no longer asks crawlers to skip filtered views (tech.crawlersSkipFilteredViews)")
	}
	c.record("robots.txt", err)
}

func (c *checker) unknownSpecies() {
	resp, _, err := c.get(http.MethodGet, "/species/no-such-bird-smoke-check", nil)
	if err == nil && resp.StatusCode != http.StatusNotFound {
		err = fmt.Errorf("status %d, want 404", resp.StatusCode)
	}
	c.record("an unknown species is not found", err)
}

// boundaryFile asks about the boundary file with HEAD: it is the one large
// static file, and only its headers are in question.
func (c *checker) boundaryFile(path string) {
	if path == "" {
		c.record("the boundary file", fmt.Errorf("the home page links no /boundaries/ file"))
		return
	}
	resp, _, err := c.get(http.MethodHead, path, nil)
	switch {
	case err != nil:
	case resp.StatusCode != http.StatusOK:
		err = fmt.Errorf("HEAD %s: status %d, want 200", path, resp.StatusCode)
	case !strings.Contains(resp.Header.Get("Cache-Control"), "immutable"):
		err = fmt.Errorf("HEAD %s: Cache-Control %q, want immutable", path, resp.Header.Get("Cache-Control"))
	}
	c.record("the boundary file", err)
}

// redirect checks that from sends a reader to the same path and query on the
// base's host, over HTTPS, permanently.
func (c *checker) redirect(from string) {
	name := "redirect from " + from
	b, err := url.Parse(c.base)
	if err != nil {
		c.record(name, err)
		return
	}
	const path = "/docs/using?smoke=1"
	want := "https://" + b.Host + path
	req, err := http.NewRequest(http.MethodGet, strings.TrimSuffix(from, "/")+path, nil)
	if err != nil {
		c.record(name, err)
		return
	}
	req.Header.Set("User-Agent", UserAgent)
	resp, err := c.client.Do(req)
	if err == nil {
		resp.Body.Close()
		if loc := resp.Header.Get("Location"); resp.StatusCode != http.StatusMovedPermanently || loc != want {
			err = fmt.Errorf("status %d to %q, want 301 to %q", resp.StatusCode, loc, want)
		}
	}
	c.record(name, err)
}

func joined(problems []string) error {
	if len(problems) == 0 {
		return nil
	}
	return fmt.Errorf("%s", strings.Join(problems, "; "))
}
