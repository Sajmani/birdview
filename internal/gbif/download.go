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
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// DownloadClient requests GBIF SQL downloads, waits for them, and fetches them
// (tech.gbifDownloadsRecorded). User and Password are the GBIF account's, read by the caller from
// the environment; they are sent only to submit, never to validate or fetch.
type DownloadClient struct {
	BaseURL        string
	HTTP           *http.Client
	User, Password string
	// Poll is how long Wait sleeps between status checks. A download takes
	// minutes to hours, so checking often only loads GBIF.
	Poll time.Duration
}

// Download is the part of GBIF's download record BirdView keeps.
type Download struct {
	Key          string `json:"key"`
	DOI          string `json:"doi"`
	Status       string `json:"status"`
	Created      string `json:"created"`
	License      string `json:"license"`
	TotalRecords int64  `json:"totalRecords"`
	Request      struct {
		SQL string `json:"sql"`
	} `json:"request"`
}

func (c DownloadClient) body(sql string) []byte {
	b, _ := json.Marshal(map[string]string{"format": "SQL_TSV_ZIP", "sql": sql})
	return b
}

func (c DownloadClient) post(ctx context.Context, path string, sql string, auth bool) (int, []byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+path, bytes.NewReader(c.body(sql)))
	if err != nil {
		return 0, nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", UserAgent) // tech.identifiedClient
	if auth {
		req.SetBasicAuth(c.User, c.Password)
	}
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	return resp.StatusCode, b, err
}

// Validate asks GBIF whether it would accept the query, without credentials, and returns the
// query in GBIF's own formatting, which is the SQL a download requested from GBIF's website carries.
func (c DownloadClient) Validate(ctx context.Context, sql string) (string, error) {
	code, b, err := c.post(ctx, "/occurrence/download/request/validate", sql, false)
	if err != nil {
		return "", err
	}
	if code != http.StatusCreated && code != http.StatusOK {
		return "", fmt.Errorf("GBIF rejects the query (HTTP %d): %s", code, strings.TrimSpace(string(b)))
	}
	var v struct {
		SQL string `json:"sql"`
	}
	if err := json.Unmarshal(b, &v); err != nil {
		return "", fmt.Errorf("reading GBIF's validation: %w", err)
	}
	return v.SQL, nil
}

// Submit requests the download and returns its key.
func (c DownloadClient) Submit(ctx context.Context, sql string) (string, error) {
	if c.User == "" || c.Password == "" {
		return "", fmt.Errorf("no GBIF credentials; set GBIF_USER and GBIF_PASSWORD")
	}
	code, b, err := c.post(ctx, "/occurrence/download/request", sql, true)
	if err != nil {
		return "", err
	}
	if code != http.StatusCreated {
		return "", fmt.Errorf("GBIF refused the download (HTTP %d): %s", code, strings.TrimSpace(string(b)))
	}
	return strings.TrimSpace(string(b)), nil
}

// Status reads a download's record. It needs no credentials.
func (c DownloadClient) Status(ctx context.Context, key string) (Download, error) {
	var d Download
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.BaseURL+"/occurrence/download/"+key, nil)
	if err != nil {
		return d, err
	}
	req.Header.Set("User-Agent", UserAgent)
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return d, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return d, fmt.Errorf("download %s: HTTP %d", key, resp.StatusCode)
	}
	return d, json.NewDecoder(resp.Body).Decode(&d)
}

// Wait polls until the download succeeds, and fails if GBIF reports it failed,
// was cancelled, or was killed.
func (c DownloadClient) Wait(ctx context.Context, key string) (Download, error) {
	for {
		d, err := c.Status(ctx, key)
		if err != nil {
			return d, err
		}
		switch d.Status {
		case "SUCCEEDED":
			return d, nil
		case "FAILED", "CANCELLED", "KILLED", "FILE_ERASED":
			return d, fmt.Errorf("download %s is %s", key, d.Status)
		}
		select {
		case <-ctx.Done():
			return d, ctx.Err()
		case <-time.After(c.Poll):
		}
	}
}

// Fetch writes the download's file to w and returns its SHA-256, hex-encoded.
func (c DownloadClient) Fetch(ctx context.Context, key string, w io.Writer) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.BaseURL+"/occurrence/download/request/"+key+".zip", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", UserAgent)
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("fetching %s: HTTP %d", key, resp.StatusCode)
	}
	h := sha256.New()
	if _, err := io.Copy(io.MultiWriter(w, h), resp.Body); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// Record is one row of data/gbif-downloads.tsv: what BirdView cites a download by, and what checks
// a file against it (tech.gbifDownloadsRecorded, gbifterms.CiteDownloadDoi).
type Record struct {
	Key, DOI, Created, License string
	TotalRecords               int64
	Query, SHA256              string
}

// RecordHeader names the columns of data/gbif-downloads.tsv.
const RecordHeader = "key\tdoi\tcreated\tlicense\trecords\tquery\tsha256"

// RecordFor builds a download's record, refusing one that has not succeeded or that was requested
// with SQL other than the committed query, so no file is ever cited under the wrong query. The
// query may be given in several forms: as committed, and as GBIF's validator formats it.
func RecordFor(d Download, queryFile string, querySQL []string, sha string) (Record, error) {
	if d.Status != "SUCCEEDED" {
		return Record{}, fmt.Errorf("download %s is %s, not SUCCEEDED", d.Key, d.Status)
	}
	matched := false
	for _, q := range querySQL {
		matched = matched || (q != "" && strings.TrimSpace(d.Request.SQL) == strings.TrimSpace(q))
	}
	if !matched {
		return Record{}, fmt.Errorf("download %s was requested with different SQL than %s", d.Key, queryFile)
	}
	if d.DOI == "" {
		return Record{}, fmt.Errorf("download %s has no DOI to cite", d.Key)
	}
	created := d.Created
	if len(created) >= 10 {
		created = created[:10]
	}
	return Record{Key: d.Key, DOI: d.DOI, Created: created, License: d.License,
		TotalRecords: d.TotalRecords, Query: queryFile, SHA256: sha}, nil
}

// Row is the record as a line of data/gbif-downloads.tsv.
func (r Record) Row() string {
	return strings.Join([]string{r.Key, r.DOI, r.Created, r.License,
		fmt.Sprint(r.TotalRecords), r.Query, r.SHA256}, "\t")
}
