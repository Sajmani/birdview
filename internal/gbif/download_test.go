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
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

// TestDownloadSubmitsWaitsAndRecords drives a SQL download against a fake GBIF: the query is
// validated without credentials, submitted with them, polled until it succeeds, fetched, and
// recorded with its key, DOI, license, and the SHA-256 of what was fetched; a download whose SQL is
// not the committed query is refused rather than recorded under it.
//
// Verifies: tech.gbifDownloadsRecorded, gbifterms.CiteDownloadDoi.
func TestDownloadSubmitsWaitsAndRecords(t *testing.T) {
	const query = `SELECT datasetkey, "month", COUNT(*) AS n FROM occurrence GROUP BY datasetkey, "month"`
	// GBIF's validator returns the query in its own formatting, which is what its website submits.
	const formatted = "SELECT\n  datasetkey,\n  \"month\",\n  COUNT(*) n\nFROM\n  occurrence\nGROUP BY\n  occurrence.datasetkey,\n  occurrence.\"month\""
	zipBody := []byte("PK fake zip")
	var polls atomic.Int32
	var sawAuthOnValidate, sawAuthOnSubmit bool
	mux := http.NewServeMux()
	mux.HandleFunc("/occurrence/download/request/validate", func(w http.ResponseWriter, r *http.Request) {
		_, _, sawAuthOnValidate = r.BasicAuth()
		var body map[string]any
		json.NewDecoder(r.Body).Decode(&body)
		if body["format"] != "SQL_TSV_ZIP" || body["sql"] != query {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]any{"sql": formatted, "format": "SQL_TSV_ZIP"})
	})
	mux.HandleFunc("/occurrence/download/request", func(w http.ResponseWriter, r *http.Request) {
		u, p, ok := r.BasicAuth()
		sawAuthOnSubmit = ok && u == "birder" && p == "secret"
		if !sawAuthOnSubmit {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusCreated)
		io.WriteString(w, "0001234-261103000000000")
	})
	mux.HandleFunc("/occurrence/download/0001234-261103000000000", func(w http.ResponseWriter, r *http.Request) {
		status := "RUNNING"
		if polls.Add(1) >= 3 {
			status = "SUCCEEDED"
		}
		json.NewEncoder(w).Encode(map[string]any{
			"key": "0001234-261103000000000", "doi": "10.15468/dl.abc123", "status": status,
			"created":      "2026-11-03T22:00:00.000+00:00",
			"license":      "http://creativecommons.org/licenses/by-nc/4.0/legalcode",
			"totalRecords": 42, "request": map[string]any{"sql": query, "format": "SQL_TSV_ZIP"},
		})
	})
	mux.HandleFunc("/occurrence/download/request/0001234-261103000000000.zip", func(w http.ResponseWriter, r *http.Request) {
		w.Write(zipBody)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	c := DownloadClient{BaseURL: srv.URL, HTTP: srv.Client(), User: "birder", Password: "secret", Poll: time.Millisecond}
	ctx := context.Background()
	norm, err := c.Validate(ctx, query)
	if err != nil || norm != formatted {
		t.Fatalf("Validate = %q, %v; want GBIF's formatting of the query", norm, err)
	}
	if sawAuthOnValidate {
		t.Error("validation sent credentials; it needs none, and they should travel only when required")
	}
	if _, err := c.Validate(ctx, "SELECT nonsense"); err == nil {
		t.Error("a query the validator rejects passed Validate")
	}
	key, err := c.Submit(ctx, query)
	if err != nil || key != "0001234-261103000000000" {
		t.Fatalf("Submit = %q, %v", key, err)
	}
	d, err := c.Wait(ctx, key)
	if err != nil || d.Status != "SUCCEEDED" || polls.Load() < 3 {
		t.Fatalf("Wait = %+v, %v after %d polls", d, err, polls.Load())
	}
	var file bytes.Buffer
	sum, err := c.Fetch(ctx, key, &file)
	if err != nil || !bytes.Equal(file.Bytes(), zipBody) {
		t.Fatalf("Fetch = %v; got %q", err, file.Bytes())
	}
	want := sha256.Sum256(zipBody)
	if sum != hex.EncodeToString(want[:]) {
		t.Errorf("Fetch reported SHA-256 %s, want %x", sum, want)
	}

	rec, err := RecordFor(d, "data/gbif-queries/test.sql", []string{query, formatted}, sum)
	if err != nil {
		t.Fatalf("RecordFor: %v", err)
	}
	row := rec.Row()
	for _, f := range []string{"0001234-261103000000000", "10.15468/dl.abc123", "2026-11-03", "data/gbif-queries/test.sql", sum, "by-nc/4.0"} {
		if !strings.Contains(row, f) {
			t.Errorf("the record %q lacks %q", row, f)
		}
	}
	if strings.Contains(row, "secret") {
		t.Error("the record carries the password")
	}
	// Requested from the website, the download carries GBIF's formatting, and is still the query.
	viaWebsite := d
	viaWebsite.Request.SQL = formatted
	if _, err := RecordFor(viaWebsite, "data/gbif-queries/test.sql", []string{query, formatted}, sum); err != nil {
		t.Errorf("a download requested with GBIF's formatting of the query was refused: %v", err)
	}
	if _, err := RecordFor(d, "data/gbif-queries/other.sql", []string{"SELECT something else"}, sum); err == nil {
		t.Error("a download was recorded under a query it was not requested from")
	}
	unfinished := d
	unfinished.Status = "RUNNING"
	if _, err := RecordFor(unfinished, "data/gbif-queries/test.sql", []string{query}, sum); err == nil {
		t.Error("an unfinished download was recorded")
	}
}
