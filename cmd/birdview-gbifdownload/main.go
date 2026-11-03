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

// Command birdview-gbifdownload requests a GBIF SQL download from a committed query, waits for it,
// fetches it into fetched/gbif/, and records its key, DOI, license, and hash in
// data/gbif-downloads.tsv (tech.gbifDownloadsRecorded).
//
//	birdview-gbifdownload -query data/gbif-queries/ny-counties.sql -validate
//	GBIF_USER=… GBIF_PASSWORD=… birdview-gbifdownload -query data/gbif-queries/ny-counties.sql
//	birdview-gbifdownload -query data/gbif-queries/ny-counties.sql -key 0001234-261103000000000
//
// The last form fetches and records a download already requested, from GBIF's website or an
// earlier run; it needs no credentials.
package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"birdview/internal/gbif"
)

func main() {
	query := flag.String("query", "", "the committed SQL query, under data/gbif-queries/")
	key := flag.String("key", "", "fetch and record this download instead of requesting one")
	validate := flag.Bool("validate", false, "only ask GBIF whether it would accept the query")
	out := flag.String("out", "fetched/gbif", "where downloads are written; kept out of the repository")
	record := flag.String("record", "data/gbif-downloads.tsv", "the committed table of downloads")
	poll := flag.Duration("poll", time.Minute, "how long to wait between status checks")
	flag.Parse()
	if *query == "" {
		log.Fatal("birdview-gbifdownload: -query is required")
	}
	sqlBytes, err := os.ReadFile(*query)
	if err != nil {
		log.Fatal(err)
	}
	sql := strings.TrimSpace(string(sqlBytes))
	c := gbif.DownloadClient{
		BaseURL: gbif.DefaultBaseURL, HTTP: &http.Client{Timeout: 10 * time.Minute},
		User: os.Getenv("GBIF_USER"), Password: os.Getenv("GBIF_PASSWORD"), Poll: *poll,
	}
	ctx := context.Background()

	// Validated even when only fetching: GBIF's formatting of the query is how a download
	// requested from its website records it.
	formatted, err := c.Validate(ctx, sql)
	if err != nil {
		log.Fatal(err)
	}
	k := *key
	if k == "" {
		log.Printf("GBIF accepts %s", *query)
		if *validate {
			return
		}
		if k, err = c.Submit(ctx, sql); err != nil {
			log.Fatal(err)
		}
		log.Printf("requested download %s; GBIF also emails when it is ready", k)
	}
	d, err := c.Wait(ctx, k)
	if err != nil {
		log.Fatal(err)
	}
	if err := os.MkdirAll(*out, 0o755); err != nil {
		log.Fatal(err)
	}
	path := filepath.Join(*out, k+".zip")
	f, err := os.Create(path)
	if err != nil {
		log.Fatal(err)
	}
	sum, err := c.Fetch(ctx, k, f)
	if cerr := f.Close(); err == nil {
		err = cerr
	}
	if err != nil {
		log.Fatal(err)
	}
	rec, err := gbif.RecordFor(d, filepath.ToSlash(*query), []string{sql, formatted}, sum)
	if err != nil {
		log.Fatal(err)
	}
	if err := writeRecord(*record, rec); err != nil {
		log.Fatal(err)
	}
	log.Printf("wrote %s (%d records), DOI %s, recorded in %s", path, rec.TotalRecords, rec.DOI, *record)
}

// writeRecord adds rec to the table, replacing any earlier row for the same key.
func writeRecord(path string, rec gbif.Record) error {
	header := []string{
		"# GBIF downloads BirdView reads, each requested from a committed query (tech.gbifDownloadsRecorded).",
		"# Written by cmd/birdview-gbifdownload. Cite each by its DOI (gbifterms.CiteDownloadDoi).",
		"#",
		"# " + gbif.RecordHeader,
	}
	var rows []string
	if f, err := os.Open(path); err == nil {
		s := bufio.NewScanner(f)
		for s.Scan() {
			line := s.Text()
			if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, rec.Key+"\t") {
				continue
			}
			rows = append(rows, line)
		}
		f.Close()
		if err := s.Err(); err != nil {
			return err
		}
	} else if !os.IsNotExist(err) {
		return err
	}
	rows = append(rows, rec.Row())
	return os.WriteFile(path, []byte(strings.Join(header, "\n")+"\n"+strings.Join(rows, "\n")+"\n"), 0o644)
}

func init() { log.SetFlags(0); log.SetPrefix(fmt.Sprintf("%s: ", filepath.Base(os.Args[0]))) }
