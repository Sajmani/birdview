package gbif

import (
	"context"
	"encoding/json"
	"fmt"
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
	HTTP    *http.Client

	mu   sync.Mutex
	last time.Time
}

// NewClient returns a client with the courtesy defaults above.
func NewClient() *Client {
	return &Client{
		BaseURL:     DefaultBaseURL,
		UserAgent:   UserAgent,
		Interval:    MinInterval,
		Concurrency: DefaultConcurrency,
		Backoff:     DefaultBackoff,
		HTTP:        &http.Client{Timeout: 120 * time.Second},
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
// gadmGid is a level-1 region: "USA.5_1" is California.
func (c *Client) NamesIn(ctx context.Context, datasetKey, gadmGid string) ([]RegionRecord, error) {
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

		var body facetResponse
		if err := c.get(ctx, "/occurrence/search?"+q.Encode(), &body); err != nil {
			return nil, fmt.Errorf("faceting names in %s at offset %d: %w", gadmGid, offset, err)
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

// get performs one paced, identified request and decodes the JSON body.
func (c *Client) get(ctx context.Context, path string, into any) error {
	var last error
	for attempt := 0; attempt < maxAttempts; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(c.Backoff << (attempt - 1)):
			}
		}
		c.pace()

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.BaseURL+path, nil)
		if err != nil {
			return err
		}
		req.Header.Set("User-Agent", c.UserAgent)
		req.Header.Set("Accept", "application/json")

		resp, err := c.HTTP.Do(req)
		if err != nil {
			last = err
			continue
		}
		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			last = fmt.Errorf("GET %s: %s", path, resp.Status)
			// 4xx other than 429 will not improve on a retry.
			if resp.StatusCode < 500 && resp.StatusCode != http.StatusTooManyRequests {
				return last
			}
			continue
		}
		err = json.NewDecoder(resp.Body).Decode(into)
		resp.Body.Close()
		if err != nil {
			return fmt.Errorf("GET %s: %w", path, err)
		}
		return nil
	}
	return last
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
	if !c.last.IsZero() {
		if wait := c.Interval - time.Since(c.last); wait > 0 {
			time.Sleep(wait)
		}
	}
	c.last = time.Now()
}
