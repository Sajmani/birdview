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

package inat

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// DefaultBaseURL is the public API root. Overridable so tests run against an
// httptest server and never touch the live service (tech.md bindings).
const DefaultBaseURL = "https://api.inaturalist.org/v1"

// MinInterval is the pacing iNaturalist asks for: "about 1 per second"
// (inatapi.RequestRate, tech.pacedInaturalistMapping). A full rebuild is 841 requests, so roughly fifteen
// minutes -- which is exactly why this runs in its own tool and not in the
// ingest.
const MinInterval = time.Second

// UserAgent identifies the client, as inatapi.UserAgent and tech.identifiedClient both require.
// A contact address is included because a service operator who needs to ask a
// client to stop should not have to guess who to ask. No project URL: BirdView
// is not published anywhere yet, and a plausible-looking dead link is worse
// than none.
const UserAgent = "birdview-inatmap/1.0 (contact: birdview@ajmani.net)"

// Client is a paced, identified, read-only iNaturalist API client. It sends no
// credential: BirdView has no account and needs none, which satisfies
// inatapi.JwtAuthorization, R6, and R7 by construction.
type Client struct {
	BaseURL   string
	UserAgent string
	Interval  time.Duration
	HTTP      *http.Client

	last time.Time
}

// NewClient returns a client with the defaults the adopted source asks for.
func NewClient() *Client {
	return &Client{
		BaseURL:   DefaultBaseURL,
		UserAgent: UserAgent,
		Interval:  MinInterval,
		HTTP:      &http.Client{Timeout: 30 * time.Second},
	}
}

// taxaResponse is the sliver of the taxa endpoint BirdView reads. Everything
// else in the response -- and it is a lot -- is discarded unparsed.
type taxaResponse struct {
	Results []struct {
		ID                  int    `json:"id"`
		Name                string `json:"name"`
		Rank                string `json:"rank"`
		PreferredCommonName string `json:"preferred_common_name"`
		Names               []struct {
			Name   string `json:"name"`
			Locale string `json:"locale"`
		} `json:"names"`
	} `json:"results"`
}

// Lookup resolves one species. It returns the entry and true only when
// iNaturalist offers a species-rank taxon that Accept is satisfied with;
// otherwise it returns false, and the caller emits no link (product.inaturalistLink).
//
// A miss is not an error. Most of the time it will mean iNaturalist has not yet
// followed a split that ABA has, which is a fact about the world rather than a
// failure of this code.
func (c *Client) Lookup(ctx context.Context, sciName, commonName string) (Entry, bool, error) {
	c.pace()

	// autocomplete rather than the plain search: the latter fails outright on
	// short repeated binomials -- "Alle alle" returns a mustard plant and a
	// domestic cat -- while autocomplete finds the Dovekie. all_names asks for
	// every vernacular, not just the preferred one (product.inaturalistLink).
	u := fmt.Sprintf("%s/taxa/autocomplete?rank=species&all_names=true&q=%s",
		c.BaseURL, url.QueryEscape(sciName))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return Entry{}, false, err
	}
	req.Header.Set("User-Agent", c.UserAgent)
	req.Header.Set("Accept", "application/json")

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return Entry{}, false, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return Entry{}, false, fmt.Errorf("iNaturalist %s: %s", sciName, resp.Status)
	}

	var body taxaResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return Entry{}, false, fmt.Errorf("iNaturalist %s: %w", sciName, err)
	}
	for _, r := range body.Results {
		if r.Rank != "species" {
			continue
		}
		commons := []string{r.PreferredCommonName}
		for _, n := range r.Names {
			if strings.HasPrefix(n.Locale, "en") {
				commons = append(commons, n.Name)
			}
		}
		if Accept(sciName, commonName, r.Name, commons) {
			// Record the preferred name: it is what a reader will see on the
			// page they land on, even where the match was made on another.
			return Entry{ScientificName: sciName, TaxonID: r.ID, CommonName: r.PreferredCommonName}, true, nil
		}
	}
	return Entry{}, false, nil
}

// pace waits out the remainder of Interval since the last request.
func (c *Client) pace() {
	if !c.last.IsZero() {
		if wait := c.Interval - time.Since(c.last); wait > 0 {
			time.Sleep(wait)
		}
	}
	c.last = time.Now()
}
