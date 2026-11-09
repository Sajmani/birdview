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

// Command birdview-sea writes data/sea-distance.tsv: for each species BirdView draws, how many of its
// eBird records on water lie under 10 km, 10 to 30 km, and 30 km or more from shore
// (tech.seaDistanceTable, decisions.openOceanSplit). It reads the offshore-cells GBIF download and
// Natural Earth's land and minor islands, pinned by hash.
package main

import (
	"archive/zip"
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"birdview/internal/gbif"
	"birdview/internal/geo"
	"birdview/internal/region"
	"birdview/internal/shapefile"
)

// The pinned sources. See river/requirements/imports.md#natural-earth.
var archives = []struct{ url, sha256, stem string }{
	{"https://naciscdn.org/naturalearth/10m/physical/ne_10m_land.zip",
		"e547d749445eaa0964aba76738090ec88f5e63c4585122170f98c67a7ea922dc", "ne_10m_land"},
	{"https://naciscdn.org/naturalearth/10m/physical/ne_10m_minor_islands.zip",
		"47c0b3ce26df7b4dbabdbb251023918cb2f5c83462a426288f1f253da7284db2", "ne_10m_minor_islands"},
}

const userAgent = "birdview-sea/1.0 (contact: birdview@ajmani.net)" // tech.identifiedClient

// west is the longitude west of which polygons are read: every sea box lies west of it.
const west = -30

// Bands, in kilometers from shore (decisions.openOceanSplit).
const (
	nearKm = 10
	farKm  = 30
)

func main() {
	log.SetFlags(0)
	log.SetPrefix("birdview-sea: ")
	download := flag.String("download", "fetched/gbif/0025653-261023074918325.zip", "the offshore-cells GBIF download (data/gbif-queries/offshore-cells.sql)")
	downloads := flag.String("downloads", "data/gbif-downloads.tsv", "the record of GBIF downloads, for the download's DOI")
	taxa := flag.String("taxa", "data/gbif-taxa.tsv", "GBIF species key to AviList name")
	data := flag.String("data", "web/data/species.json", "the built dataset, for the species BirdView draws")
	cache := flag.String("cache", filepath.Join(os.TempDir(), "birdview-sea"), "where the Natural Earth archives are kept once fetched")
	out := flag.String("out", "data/sea-distance.tsv", "where to write the table")
	flag.Parse()

	var polygons [][][][2]float64
	for _, a := range archives {
		p, err := polygonsOf(a.url, a.sha256, a.stem, *cache)
		if err != nil {
			log.Fatal(err)
		}
		polygons = append(polygons, p...)
	}
	shore := geo.NewShore(polygons, west)

	key := strings.TrimSuffix(filepath.Base(*download), ".zip")
	doi, err := doiOf(*downloads, key)
	if err != nil {
		log.Fatal(err)
	}
	names, err := drawnNames(*data, *taxa)
	if err != nil {
		log.Fatal(err)
	}
	rows, err := readCells(*download)
	if err != nil {
		log.Fatal(err)
	}
	bands, unmatched := measure(rows, shore, names)
	f, err := os.Create(*out)
	if err != nil {
		log.Fatal(err)
	}
	if err := write(f, bands, key, doi); err != nil {
		log.Fatal(err)
	}
	if err := f.Close(); err != nil {
		log.Fatal(err)
	}
	log.Printf("wrote %s: %d species; %d eBird records on water under names BirdView does not draw", *out, len(bands), unmatched)
}

// cellRow is one row of the download: records of one name in one 0.1° cell and month.
type cellRow struct {
	dataset      string
	lat10, lon10 int
	name, key    string
	n            int64
}

func readCells(path string) ([]cellRow, error) {
	zr, err := zip.OpenReader(path)
	if err != nil {
		return nil, err
	}
	defer zr.Close()
	if len(zr.File) != 1 {
		return nil, fmt.Errorf("%s holds %d files, want the one table", path, len(zr.File))
	}
	f, err := zr.File[0].Open()
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return parseCells(f)
}

func parseCells(r io.Reader) ([]cellRow, error) {
	s := bufio.NewScanner(r)
	s.Buffer(make([]byte, 1<<20), 1<<20)
	col := map[string]int{}
	var rows []cellRow
	for line := 0; s.Scan(); line++ {
		f := strings.Split(s.Text(), "\t")
		if line == 0 {
			for i, name := range f {
				col[name] = i
			}
			for _, want := range []string{"datasetkey", "lat10", "lon10", "v_scientificname", "specieskey", "n"} {
				if _, ok := col[want]; !ok {
					return nil, fmt.Errorf("the download has no %s column; it is not the offshore-cells query's", want)
				}
			}
			continue
		}
		lat, err1 := strconv.Atoi(f[col["lat10"]])
		lon, err2 := strconv.Atoi(f[col["lon10"]])
		n, err3 := strconv.ParseInt(f[col["n"]], 10, 64)
		if err1 != nil || err2 != nil || err3 != nil {
			return nil, fmt.Errorf("line %d: lat10 %q, lon10 %q, n %q", line+1, f[col["lat10"]], f[col["lon10"]], f[col["n"]])
		}
		rows = append(rows, cellRow{dataset: f[col["datasetkey"]], lat10: lat, lon10: lon,
			name: f[col["v_scientificname"]], key: f[col["specieskey"]], n: n})
	}
	return rows, s.Err()
}

// names finds the AviList name of a drawn species from a GBIF species key, else from the verbatim
// name.
type names struct {
	byKey  map[string]string
	byName map[string]string
}

func (n names) of(r cellRow) (string, bool) {
	if sci, ok := n.byKey[r.key]; ok {
		return sci, true
	}
	sci, ok := n.byName[strings.ToLower(r.name)]
	return sci, ok
}

// measure bands every eBird row inside a sea box by its cell center's distance to shore, per drawn
// species, dropping cells whose center lies on land. It returns the bands by AviList name and the
// records under names no drawn species has.
func measure(rows []cellRow, shore *geo.Shore, n names) (map[string][3]int64, int64) {
	bands := map[string][3]int64{}
	var unmatched int64
	for _, r := range rows {
		if r.dataset != gbif.EBirdDatasetKey {
			continue
		}
		lat, lon := (float64(r.lat10)+0.5)/10, (float64(r.lon10)+0.5)/10
		if !inSeaBox(lat, lon) {
			continue
		}
		sci, ok := n.of(r)
		if !ok {
			unmatched += r.n
			continue
		}
		if shore.OnLand(lat, lon) {
			continue
		}
		b := bands[sci]
		switch d := shore.DistanceKm(lat, lon, 100); {
		case d < nearKm:
			b[0] += r.n
		case d < farKm:
			b[1] += r.n
		default:
			b[2] += r.n
		}
		bands[sci] = b
	}
	return bands, unmatched
}

func inSeaBox(lat, lon float64) bool {
	for _, id := range region.MarineIDs() {
		if minLat, maxLat, minLon, maxLon, _ := region.MarineBox(id); lat >= minLat && lat <= maxLat && lon >= minLon && lon <= maxLon {
			return true
		}
	}
	return false
}

func write(w io.Writer, bands map[string][3]int64, key, doi string) error {
	bw := bufio.NewWriter(w)
	fmt.Fprintf(bw, "# Where each drawn species' eBird records on water lie relative to shore (tech.seaDistanceTable).\n")
	fmt.Fprintf(bw, "# Written by cmd/birdview-sea from GBIF download %s (doi:%s) and Natural Earth 5.1.1\n", key, doi)
	fmt.Fprintf(bw, "# 1:10m land and minor islands. eBird records only, in BirdView's sea boxes; a 0.1-degree cell\n")
	fmt.Fprintf(bw, "# whose center lies on land is dropped, and the rest are banded by their center's distance to shore.\n")
	fmt.Fprintf(bw, "#\n")
	fmt.Fprintf(bw, "scientific name\tunder %d km\t%d to %d km\t%d km or more\n", nearKm, nearKm, farKm, farKm)
	sci := make([]string, 0, len(bands))
	for s := range bands {
		sci = append(sci, s)
	}
	sort.Strings(sci)
	for _, s := range sci {
		b := bands[s]
		fmt.Fprintf(bw, "%s\t%d\t%d\t%d\n", s, b[0], b[1], b[2])
	}
	return bw.Flush()
}

// drawnNames reads the species BirdView draws and the GBIF bridge to their AviList names.
func drawnNames(dataPath, taxaPath string) (names, error) {
	b, err := os.ReadFile(dataPath)
	if err != nil {
		return names{}, fmt.Errorf("%v; run birdview-ingest first", err)
	}
	var ds struct {
		Species []struct {
			ScientificName string `json:"scientificName"`
		} `json:"species"`
	}
	if err := json.Unmarshal(b, &ds); err != nil {
		return names{}, err
	}
	n := names{byKey: map[string]string{}, byName: map[string]string{}}
	for _, sp := range ds.Species {
		n.byName[strings.ToLower(sp.ScientificName)] = sp.ScientificName
	}
	tf, err := os.Open(taxaPath)
	if err != nil {
		return names{}, err
	}
	defer tf.Close()
	s := bufio.NewScanner(tf)
	for s.Scan() {
		if strings.HasPrefix(s.Text(), "#") {
			continue
		}
		f := strings.Split(s.Text(), "\t")
		if len(f) < 3 {
			continue
		}
		if sci, ok := n.byName[strings.ToLower(f[2])]; ok {
			n.byKey[f[1]] = sci
		}
	}
	return n, s.Err()
}

// doiOf finds a download's DOI in the record of downloads (tech.gbifDownloadsRecorded).
func doiOf(path, key string) (string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	for _, line := range strings.Split(string(b), "\n") {
		if f := strings.Split(line, "\t"); len(f) > 1 && f[0] == key {
			return f[1], nil
		}
	}
	return "", fmt.Errorf("download %s is not recorded in %s", key, path)
}

// polygonsOf reads one Natural Earth archive, from the cache or the network, refusing any but the
// pinned one.
func polygonsOf(url, want, stem, cache string) ([][][][2]float64, error) {
	path := filepath.Join(cache, stem+".zip")
	b, err := os.ReadFile(path)
	if err != nil {
		if b, err = fetch(url); err != nil {
			return nil, err
		}
		if err := os.MkdirAll(cache, 0o755); err == nil {
			_ = os.WriteFile(path, b, 0o644)
		}
	}
	sum := sha256.Sum256(b)
	if got := hex.EncodeToString(sum[:]); got != want {
		return nil, fmt.Errorf("%s has SHA-256 %s, want the pinned %s; a new release is a refresh, not a drop-in", stem, got, want)
	}
	zr, err := zip.NewReader(bytes.NewReader(b), int64(len(b)))
	if err != nil {
		return nil, err
	}
	f, err := zr.Open(stem + ".shp")
	if err != nil {
		return nil, err
	}
	defer f.Close()
	shp, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}
	return shapefile.Polygons(shp)
}

func fetch(url string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GET %s: %s", url, resp.Status)
	}
	return io.ReadAll(resp.Body)
}
