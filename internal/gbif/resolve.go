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

package gbif

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

// DefaultBaseURL is the public API root. Overridable so tests run against an
// httptest server and never touch the live service (tech.md bindings).
const DefaultBaseURL = "https://api.gbif.org/v1"

// EBirdDatasetKey is the EOD - eBird Observation Dataset on GBIF,
// doi:10.15468/aomfnb, CC BY 4.0.
const EBirdDatasetKey = "4fa7b334-ce0d-4e88-aaae-2e0c138d049e"

// AvesTaxonKey is the backbone key for class Aves.
const AvesTaxonKey = 212

// UserAgent identifies the client, as T-006 requires, with a contact address
// so an operator who needs to ask it to stop knows who to ask.
const UserAgent = "birdview-gbifmap/1.0 (contact: birdview@ajmani.net)"

// MinInterval is the floor between requests, counted across all workers.
//
// This is BirdView's own choice, not a published GBIF limit. GBIF's terms page
// is behind a Cloudflare challenge that answers an agent with HTTP 403, and
// process.md forbids supplying from memory a number that could not be
// retrieved -- a plausible rate limit is indistinguishable from a real one. So
// the number below is a courtesy BirdView sets for itself: a full run is about
// 2,800 requests, which at this pace takes around five minutes.
const MinInterval = 100 * time.Millisecond

// DefaultConcurrency is how many requests may be in flight at once. Modest on
// purpose; it hides round-trip latency without raising the request rate, which
// MinInterval bounds independently.
const DefaultConcurrency = 4

// facetPage is how many facet values to ask for at a time.
//
// The tool pages until an EMPTY page arrives, advancing by however many values
// actually came back rather than by what it asked for. Stopping at the first
// short page would be one request cheaper and wrong: a server free to cap a
// facet below the requested limit would then truncate the species list
// silently, and a truncated list looks exactly like a dataset with fewer birds
// in it.
const facetPage = 1000

// maxAttempts is how many times a request is tried before giving up. GBIF
// answers a billion-record facet query slowly and occasionally sheds load.
const maxAttempts = 4

// DefaultBackoff is the first retry delay; each further attempt doubles it.
const DefaultBackoff = time.Second

// DefaultRequestTimeout bounds a single attempt. It is deliberately generous:
// a facet over a billion records is slow, and the thing this guards against
// is not slowness but a connection that will never answer at all.
const DefaultRequestTimeout = 120 * time.Second

// Client is a paced, identified, read-only GBIF API client. It sends no
// credential: GBIF's open endpoints need none.
type Client struct {
	BaseURL     string
	UserAgent   string
	Interval    time.Duration
	Concurrency int
	// Backoff is the delay before the second attempt at a failed request,
	// doubling thereafter. A field rather than a constant so a test can
	// exercise the retry path without waiting fifteen seconds for it.
	Backoff time.Duration
	// RequestTimeout bounds one attempt, through the request's context.
	// Zero means DefaultRequestTimeout.
	RequestTimeout time.Duration
	HTTP           *http.Client

	mu   sync.Mutex
	last time.Time
	// gap is the live interval, between Interval and MaxInterval. It widens
	// when GBIF throttles and narrows again on a long clean streak.
	gap time.Duration
	// CleanRun overrides DefaultCleanRun. A field for the same reason Backoff
	// is one: at the default this test waits twenty seconds.
	CleanRun int
	// clean counts consecutive successes since the last 429.
	clean int
}

// MaxInterval caps how far throttling can widen the gap. Past this the run is
// not worth continuing unattended.
const MaxInterval = 4 * time.Second

// slowDown widens the pacing after a 429.
//
// The number was guessed twice and measured once. At 250ms across two workers
// GBIF throttled hard enough to drop a run from 32 species a minute to 4,
// almost all of it spent in backoff; at 500ms and at 1s, twenty and twelve
// consecutive requests went through untouched. Rather than pick one and hope
// it holds -- GBIF accepted 250ms earlier the same morning -- the client
// treats a 429 as the authority it is and widens until the throttling stops.
func (c *Client) slowDown() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.gap == 0 {
		c.gap = c.Interval
	}
	c.clean = 0
	if c.gap < MaxInterval {
		c.gap = min(c.gap*3/2+100*time.Millisecond, MaxInterval)
		log.Printf("gbif: throttled; pacing widened to %v", c.gap)
	}
}

// DefaultCleanRun is how many consecutive successes earn a step back toward
// the configured interval.
//
// Widening without a way back is a ratchet: one burst of throttling early in
// a long run leaves it crawling for hours after the server has recovered. The
// first run to widen sat at the 4s cap for the rest of its life, which would
// have turned a one-hour job into eight. Narrowing is deliberately slower
// than widening -- 200 clean requests to win back a step that one 429 takes
// away -- so the response to real throttling is still immediate.
const DefaultCleanRun = 200

// speedUp narrows the pacing after a long run without throttling.
func (c *Client) speedUp() {
	c.mu.Lock()
	defer c.mu.Unlock()
	run := c.CleanRun
	if run <= 0 {
		run = DefaultCleanRun
	}
	c.clean++
	if c.clean < run || c.gap <= c.Interval {
		return
	}
	c.clean = 0
	c.gap = max(c.gap*2/3, c.Interval)
	log.Printf("gbif: %d clean requests; pacing narrowed to %v", run, c.gap)
}

// NewClient returns a client with the courtesy defaults above.
func NewClient() *Client {
	return &Client{
		BaseURL:        DefaultBaseURL,
		UserAgent:      UserAgent,
		Interval:       MinInterval,
		Concurrency:    DefaultConcurrency,
		Backoff:        DefaultBackoff,
		RequestTimeout: DefaultRequestTimeout,
		HTTP:           &http.Client{Timeout: DefaultRequestTimeout},
	}
}

// SpeciesCount is one species key and the occurrences behind it.
type SpeciesCount struct {
	Key     int
	Records int64
}

type facetResponse struct {
	Count  int64 `json:"count"`
	Facets []struct {
		Field  string `json:"field"`
		Counts []struct {
			Name  string `json:"name"`
			Count int64  `json:"count"`
		} `json:"counts"`
	} `json:"facets"`
}

// SpeciesKeys lists every species the dataset reports from the given
// countries, with its occurrence count, most-reported first.
//
// It returns the total record count alongside, because the sum of the facet is
// smaller than the total and the difference is worth seeing: records whose
// name GBIF could not resolve to a species sit above species rank and appear
// in no facet at all.
func (c *Client) SpeciesKeys(ctx context.Context, datasetKey string, countries []string) ([]SpeciesCount, int64, error) {
	var out []SpeciesCount
	var total int64
	seen := map[int]bool{}
	for offset := 0; ; {
		q := url.Values{}
		q.Set("datasetKey", datasetKey)
		for _, country := range countries {
			q.Add("country", country)
		}
		q.Set("taxonKey", strconv.Itoa(AvesTaxonKey))
		q.Set("limit", "0")
		q.Set("facet", "speciesKey")
		q.Set("facetLimit", strconv.Itoa(facetPage))
		q.Set("facetOffset", strconv.Itoa(offset))

		var body facetResponse
		if err := c.get(ctx, "/occurrence/search?"+q.Encode(), &body); err != nil {
			return nil, 0, fmt.Errorf("faceting species at offset %d: %w", offset, err)
		}
		total = body.Count
		if len(body.Facets) == 0 {
			break
		}
		counts := body.Facets[0].Counts
		if len(counts) == 0 {
			break
		}
		for _, f := range counts {
			key, err := strconv.Atoi(f.Name)
			if err != nil {
				return nil, 0, fmt.Errorf("facet returned species key %q: %w", f.Name, err)
			}
			// A key arriving twice means facetOffset is not being honored, and
			// the loop would otherwise page forever over the same block.
			if seen[key] {
				return nil, 0, fmt.Errorf(
					"facet returned species key %d twice at offset %d; facetOffset is being ignored",
					key, offset)
			}
			seen[key] = true
			out = append(out, SpeciesCount{Key: key, Records: f.Count})
		}
		offset += len(counts)
	}
	return out, total, nil
}

// RegionRecord is one verbatim name and its occurrence count in a region.
type RegionRecord struct {
	// Name is the scientific name AS THE PUBLISHER WROTE IT, lowercased by
	// the facet. Deliberately not GBIF's matched name: the backbone silently
	// drops what it cannot resolve, which is how Cooper's Hawk came to have
	// zero records (sources.md, `gbif`).
	Name    string
	Records int64
}

// NamesIn lists every bird name the dataset reports from one GADM region,
// with its occurrence count.
//
// Faceting on the VERBATIM name rather than on speciesKey is the whole point.
// A speciesKey facet answers "what does GBIF's backbone think is here", which
// omits ~0.4% of records outright; this answers "what did the publisher say
// was here", which is the question scope actually turns on.
//
// gadmGid is a level-1 region: "USA.5_1" is California. A month of 1 to 12
// restricts the count to that month across all years; 0 counts the year.
//
// One month at a time, and twelve calls, because GBIF facets one field per
// query: asking for `facet=verbatimScientificName&facet=month` returns the
// two MARGINAL distributions, not the cross-product, so it cannot say which
// names those monthly records belong to. That is the whole cost of
// CR-036 — the land pass goes from about 140 requests to about 1,660.
func (c *Client) NamesIn(ctx context.Context, datasetKey, gadmGid string, month int) ([]RegionRecord, error) {
	if month < 0 || month > 12 {
		return nil, fmt.Errorf("month %d is not 0 (the whole year) or 1 to 12", month)
	}
	var out []RegionRecord
	seen := map[string]bool{}
	for offset := 0; ; {
		q := url.Values{}
		q.Set("datasetKey", datasetKey)
		q.Set("gadmGid", gadmGid)
		q.Set("taxonKey", strconv.Itoa(AvesTaxonKey))
		q.Set("limit", "0")
		q.Set("facet", "verbatimScientificName")
		q.Set("facetLimit", strconv.Itoa(facetPage))
		q.Set("facetOffset", strconv.Itoa(offset))
		if month > 0 {
			q.Set("month", strconv.Itoa(month))
		}

		var body facetResponse
		if err := c.get(ctx, "/occurrence/search?"+q.Encode(), &body); err != nil {
			return nil, fmt.Errorf("faceting names in %s month %d at offset %d: %w",
				gadmGid, month, offset, err)
		}
		if len(body.Facets) == 0 || len(body.Facets[0].Counts) == 0 {
			break
		}
		for _, f := range body.Facets[0].Counts {
			n := strings.ToLower(strings.TrimSpace(f.Name))
			if n == "" {
				continue
			}
			// As in SpeciesKeys: a repeat means facetOffset is being ignored
			// and the loop would page forever over one block.
			if seen[n] {
				return nil, fmt.Errorf(
					"facet returned %q twice at offset %d in %s; facetOffset is being ignored",
					n, offset, gadmGid)
			}
			seen[n] = true
			out = append(out, RegionRecord{Name: n, Records: f.Count})
		}
		offset += len(body.Facets[0].Counts)
	}
	return out, nil
}

// CountFor returns how many records the dataset holds for one verbatim name
// in the given countries. Used to measure what share of a species' records
// fall outside every land region, which is what makes a bird pelagic.
func (c *Client) CountFor(ctx context.Context, datasetKey, name string, countries []string) (int64, error) {
	q := url.Values{}
	q.Set("datasetKey", datasetKey)
	for _, country := range countries {
		q.Add("country", country)
	}
	q.Set("verbatimScientificName", name)
	q.Set("limit", "0")
	var body facetResponse
	if err := c.get(ctx, "/occurrence/search?"+q.Encode(), &body); err != nil {
		return 0, fmt.Errorf("counting %q: %w", name, err)
	}
	return body.Count, nil
}

// DatesIn returns the distinct calendar dates on which one verbatim name was
// recorded in one GADM region, with the records on each, plus the region's
// total for that name and whether the facet was truncated.
//
// The eventDate facet is what makes a recurrence rule measurable at all. A
// page of occurrence records is a contiguous slice of GBIF's own ordering
// rather than a sample -- all 300 American Robin records for Pennsylvania
// fell in January 2024 -- so dates read off one are not a distribution. A
// facet covers every record behind the query.
//
// Truncation is reported rather than paged away. A species with more distinct
// dates than a facet page is established by inspection, so the extra pages
// would buy nothing; what matters is that the caller never reads a floor as a
// count (internal/status.Shape.Truncated).
//
// Verifies: P-001, T-006.
func (c *Client) DatesIn(ctx context.Context, datasetKey, gadmGid, name string) (
	dates map[string]int64, total int64, truncated bool, err error) {
	q := url.Values{}
	q.Set("datasetKey", datasetKey)
	q.Set("gadmGid", gadmGid)
	q.Set("verbatimScientificName", name)
	q.Set("limit", "0")
	q.Set("facet", "eventDate")
	q.Set("facetLimit", strconv.Itoa(facetPage))

	var body facetResponse
	if err := c.get(ctx, "/occurrence/search?"+q.Encode(), &body); err != nil {
		return nil, 0, false, fmt.Errorf("faceting dates for %q in %s: %w", name, gadmGid, err)
	}
	dates = map[string]int64{}
	if len(body.Facets) > 0 {
		for _, f := range body.Facets[0].Counts {
			// GBIF answers the eventDate facet with an interval field, whose
			// values are plain ISO dates where the record carried one and a
			// range where it did not. Only a single day is usable.
			if len(f.Name) == 10 {
				dates[f.Name] = f.Count
			}
		}
		truncated = len(body.Facets[0].Counts) >= facetPage
	}
	return dates, body.Count, truncated, nil
}

// CountInBox returns how many records for one verbatim name fall inside a
// coordinate box. Used to decide which ocean a pelagic species belongs to.
func (c *Client) CountInBox(ctx context.Context, datasetKey, name string,
	countries []string, minLat, maxLat, minLon, maxLon float64) (int64, error) {
	q := url.Values{}
	q.Set("datasetKey", datasetKey)
	for _, country := range countries {
		q.Add("country", country)
	}
	q.Set("verbatimScientificName", name)
	q.Set("decimalLatitude", fmt.Sprintf("%g,%g", minLat, maxLat))
	q.Set("decimalLongitude", fmt.Sprintf("%g,%g", minLon, maxLon))
	q.Set("limit", "0")
	var body facetResponse
	if err := c.get(ctx, "/occurrence/search?"+q.Encode(), &body); err != nil {
		return 0, fmt.Errorf("counting %q in box: %w", name, err)
	}
	return body.Count, nil
}

// CountOffLandInBox returns how many records inside a coordinate box fall on
// no land at all, and how many the box holds in total.
//
// This is what makes a marine region mean something. GADM covers land, so a
// record on a beach carries a country and a record on the water carries
// none; faceting the box by gadmLevel0Gid and subtracting gives exactly the
// off-land half, with the coastline itself as the boundary. No distance
// threshold is involved, and none would help: measured east of Cape Hatteras,
// composition does not taper with distance, it flips between the band that
// holds coastline (4.9% off land, led by Northern Cardinal) and the band that
// does not (100% off land, led by Wilson's Storm-Petrel). See CR-031.
//
// One request for both numbers: a facet response carries the unfiltered count
// beside the facet rows, so this costs exactly what the plain count it
// replaces did.
//
// An empty name counts every bird instead of one species, which is how a
// sea's own denominator is measured.
//
// What it does NOT exclude is water inside a bay: an estuary has no GADM
// country either, so Chesapeake Bay reads 8.8% off land. The result is
// therefore "on water", not "at sea", and overstates the open ocean by
// roughly 3%. GBIF would accept a WKT geometry that carves the coast off
// properly, which is the refinement available if that 3% ever matters.
// MonthlyOffLandInBox is CountOffLandInBox split by month, in two requests
// rather than twenty-four.
//
// GBIF ORs repeated gadmGid values, so one month facet over the box gives
// the total per month and a second restricted to `gadmGid=USA|CAN` gives the
// part on land; the difference is water. Verified 2026-10-07.
//
// It differs from CountOffLandInBox by counting as on-land only what is
// inside the USA or CAN polygons, where that function sums every
// gadmLevel0Gid the box returns. Measured over the Gulf box, the two
// disagree by **11 records in 108 million** — records inside a third
// country's polygon — which is 0.00001%, against the ~3% by which "on
// water" already overstates "at sea". Named explicitly rather than left for
// someone to discover as a drift between two tables.
//
// Verifies: P-011, CR-036.
func (c *Client) MonthlyOffLandInBox(ctx context.Context, datasetKey, name string,
	countries []string, minLat, maxLat, minLon, maxLon float64) (offLand, total [12]int64, err error) {

	ask := func(onLandOnly bool) ([12]int64, error) {
		var out [12]int64
		q := url.Values{}
		q.Set("datasetKey", datasetKey)
		for _, country := range countries {
			q.Add("country", country)
		}
		if name != "" {
			q.Set("verbatimScientificName", name)
		} else {
			q.Set("taxonKey", strconv.Itoa(AvesTaxonKey))
		}
		q.Set("decimalLatitude", fmt.Sprintf("%g,%g", minLat, maxLat))
		q.Set("decimalLongitude", fmt.Sprintf("%g,%g", minLon, maxLon))
		if onLandOnly {
			q.Add("gadmGid", "USA")
			q.Add("gadmGid", "CAN")
		}
		q.Set("limit", "0")
		q.Set("facet", "month")
		q.Set("facetLimit", "20")
		var body facetResponse
		if err := c.get(ctx, "/occurrence/search?"+q.Encode(), &body); err != nil {
			return out, fmt.Errorf("faceting months for %q in box: %w", name, err)
		}
		if len(body.Facets) == 0 {
			return out, nil
		}
		for _, f := range body.Facets[0].Counts {
			m, err := strconv.Atoi(f.Name)
			if err != nil || m < 1 || m > 12 {
				// A record with no month cannot be placed. None exist in
				// either source -- 65,504,204 checked -- so this is a
				// guard against that changing, not a path taken.
				return out, fmt.Errorf("month facet returned %q, which is not 1 to 12", f.Name)
			}
			out[m-1] = f.Count
		}
		return out, nil
	}

	total, err = ask(false)
	if err != nil {
		return offLand, total, err
	}
	onLand, err := ask(true)
	if err != nil {
		return offLand, total, err
	}
	for i := range offLand {
		offLand[i] = total[i] - onLand[i]
		if offLand[i] < 0 {
			// Negative water is the loud failure mode CountOffLandInBox was
			// given for faceting the wrong field, and it is kept here for
			// the same reason: it means the two queries are not nested.
			return offLand, total, fmt.Errorf(
				"month %d in box holds %d records of which %d are on land, which is more "+
					"than all of them; the two queries are not asking about the same set",
				i+1, total[i], onLand[i])
		}
	}
	return offLand, total, nil
}

func (c *Client) CountOffLandInBox(ctx context.Context, datasetKey, name string,
	countries []string, minLat, maxLat, minLon, maxLon float64) (offLand, total int64, err error) {

	q := url.Values{}
	q.Set("datasetKey", datasetKey)
	for _, country := range countries {
		q.Add("country", country)
	}
	if name != "" {
		q.Set("verbatimScientificName", name)
	} else {
		q.Set("taxonKey", strconv.Itoa(AvesTaxonKey))
	}
	q.Set("decimalLatitude", fmt.Sprintf("%g,%g", minLat, maxLat))
	q.Set("decimalLongitude", fmt.Sprintf("%g,%g", minLon, maxLon))
	q.Set("limit", "0")
	q.Set("facet", "gadmLevel0Gid")
	// Level 0 is the country, and one row per country is all there is to
	// page: faceting plain gadmGid returns every level at once and the rows
	// then sum to more than the box holds, because a record in Virginia is
	// counted under USA, USA.47_1 and USA.47_13 alike. That mistake reads as
	// a NEGATIVE off-land count, which is at least loud.
	q.Set("facetLimit", "100")

	var body facetResponse
	if err := c.get(ctx, "/occurrence/search?"+q.Encode(), &body); err != nil {
		return 0, 0, fmt.Errorf("counting %q off land in box: %w", name, err)
	}
	total = body.Count
	var onLand int64
	if len(body.Facets) > 0 {
		for _, f := range body.Facets[0].Counts {
			onLand += f.Count
		}
	}
	if onLand > total {
		return 0, 0, fmt.Errorf(
			"counting %q off land in box: %d records on land exceeds the %d in the box; "+
				"the facet is returning more than one GADM level", name, onLand, total)
	}
	return total - onLand, total, nil
}

// RegionName returns the GADM level-1 name for a region id, by reading it off
// a single occurrence. Empty when the dataset reports nothing there.
func (c *Client) RegionName(ctx context.Context, datasetKey, gadmGid string) (string, error) {
	q := url.Values{}
	q.Set("datasetKey", datasetKey)
	q.Set("gadmGid", gadmGid)
	q.Set("limit", "1")
	var body struct {
		Results []struct {
			Gadm struct {
				Level1 struct {
					Name string `json:"name"`
				} `json:"level1"`
			} `json:"gadm"`
		} `json:"results"`
	}
	if err := c.get(ctx, "/occurrence/search?"+q.Encode(), &body); err != nil {
		return "", fmt.Errorf("naming %s: %w", gadmGid, err)
	}
	if len(body.Results) == 0 {
		return "", nil
	}
	return body.Results[0].Gadm.Level1.Name, nil
}

type speciesResponse struct {
	Key           int    `json:"key"`
	CanonicalName string `json:"canonicalName"`
	Rank          string `json:"rank"`
	Family        string `json:"family"`
}

type synonymsResponse struct {
	Results []struct {
		CanonicalName string `json:"canonicalName"`
		Rank          string `json:"rank"`
	} `json:"results"`
}

// Taxon fetches one backbone species and the synonyms filed under it.
//
// Only species-rank synonyms are kept. A subspecies-rank synonym names a
// population rather than a species, and matching one against an AviList
// species would assert a correspondence nobody published.
func (c *Client) Taxon(ctx context.Context, key int, records int64) (Taxon, error) {
	var sp speciesResponse
	if err := c.get(ctx, "/species/"+strconv.Itoa(key), &sp); err != nil {
		return Taxon{}, fmt.Errorf("species %d: %w", key, err)
	}
	if sp.CanonicalName == "" {
		return Taxon{}, fmt.Errorf("species %d: no canonical name in the backbone record", key)
	}

	var syn synonymsResponse
	if err := c.get(ctx, "/species/"+strconv.Itoa(key)+"/synonyms?limit=1000", &syn); err != nil {
		return Taxon{}, fmt.Errorf("synonyms of %d (%s): %w", key, sp.CanonicalName, err)
	}
	t := Taxon{Key: key, Name: sp.CanonicalName, Family: sp.Family, Records: records}
	for _, r := range syn.Results {
		if r.Rank == "SPECIES" && r.CanonicalName != "" {
			t.Synonyms = appendUnique(t.Synonyms, r.CanonicalName)
		}
	}
	return t, nil
}

// Taxa fetches every key, with at most Concurrency requests in flight. Results
// come back in the order the keys were given, so a run is reproducible.
//
// One failure fails the batch. A transport error is not evidence that a taxon
// has no AviList counterpart, and writing it into the table as one would bake
// a network hiccup into the repository.
func (c *Client) Taxa(ctx context.Context, keys []SpeciesCount) ([]Taxon, error) {
	out := make([]Taxon, len(keys))
	errs := make([]error, len(keys))

	n := c.Concurrency
	if n < 1 {
		n = 1
	}
	sem := make(chan struct{}, n)
	var wg sync.WaitGroup
	for i, k := range keys {
		if ctx.Err() != nil {
			break
		}
		wg.Add(1)
		go func(i int, k SpeciesCount) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			out[i], errs[i] = c.Taxon(ctx, k.Key, k.Records)
		}(i, k)
	}
	wg.Wait()
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	for _, err := range errs {
		if err != nil {
			return nil, err
		}
	}
	return out, nil
}

// retryable reports whether a status is worth another attempt.
//
// 5xx and 429 are the obvious ones. **404 is here on evidence**: GBIF's
// occurrence search returns a spurious one under load. A full status run
// failed on `mimus polyglottos` in Florida -- Northern Mockingbird, 1,370,398
// records -- which answered 404 mid-run and 200 when the identical query was
// repeated a minute later. Two of 1,429 species failed that way in one run.
//
// The cost of being wrong in this direction is bounded and small: a genuine
// 404, which a malformed name does produce, now takes four attempts and about
// 35 seconds before it is reported. The cost of the other direction is an
// abandoned run, which is what happened.
func retryable(code int) bool {
	return code >= 500 || code == http.StatusTooManyRequests || code == http.StatusNotFound
}

// get performs one paced, identified request and decodes the JSON body.
//
// Each attempt carries its own context deadline, which is belt and braces
// over http.Client.Timeout and is here because the braces failed. A run left
// going overnight hung for nine hours in a TLS read on a socket whose peer
// had gone away while the laptop slept: one goroutine stuck in the netpoller
// with 955 queued behind it on the worker semaphore, and the client timeout
// never fired. A context deadline cancels the read through a separate path.
func (c *Client) get(ctx context.Context, path string, into any) error {
	var last error
	for attempt := 0; attempt < maxAttempts; attempt++ {
		if attempt > 0 {
			// Said out loud. A retry costs seconds of backoff, and a run that
			// spends them silently looks merely slow: the first run after 404
			// became retryable dropped from 32 species a minute to 4, and
			// nothing in the log said why.
			log.Printf("gbif: retry %d/%d after %v", attempt, maxAttempts-1, last)
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(c.Backoff << (attempt - 1)):
			}
		}
		c.pace()

		if err := c.attempt(ctx, path, into, &last); err != nil {
			return err
		} else if last == nil {
			return nil
		}
	}
	return last
}

// attempt makes one request. It returns a non-nil error only when the caller
// must stop retrying; a retryable failure is reported through last.
func (c *Client) attempt(ctx context.Context, path string, into any, last *error) error {
	timeout := c.RequestTimeout
	if timeout <= 0 {
		timeout = DefaultRequestTimeout
	}
	rctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(rctx, http.MethodGet, c.BaseURL+path, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", c.UserAgent)
	req.Header.Set("Accept", "application/json")

	resp, err := c.HTTP.Do(req)
	if err != nil {
		*last = err
		return nil
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		*last = fmt.Errorf("GET %s: %s", path, resp.Status)
		if resp.StatusCode == http.StatusTooManyRequests {
			c.slowDown()
		}
		if !retryable(resp.StatusCode) {
			return *last
		}
		return nil
	}
	if err := json.NewDecoder(resp.Body).Decode(into); err != nil {
		return fmt.Errorf("GET %s: %w", path, err)
	}
	c.speedUp()
	*last = nil
	return nil
}

// pace holds every worker to one request per Interval.
//
// The sleep happens under the mutex on purpose: that is what serializes the
// workers into a single bounded stream, so raising Concurrency hides latency
// without raising the rate GBIF sees.
func (c *Client) pace() {
	if c.Interval <= 0 {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.gap == 0 {
		c.gap = c.Interval
	}
	if !c.last.IsZero() {
		if wait := c.gap - time.Since(c.last); wait > 0 {
			time.Sleep(wait)
		}
	}
	c.last = time.Now()
}
