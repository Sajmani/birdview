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
	"archive/zip"
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"birdview/internal/gbif"
	"birdview/internal/region"
)

// minCountyRecords is the floor under a county listing: one per million of a county's records is
// less than one record in every New York county, so the bar alone would list a bird on one
// sighting (decisions.countyListingRule).
const minCountyRecords = 5

// countyRow is one row of a county download: records of one name, in one county, in one month,
// from one dataset (data/gbif-queries/ny-counties.sql).
type countyRow struct {
	Dataset, GID string
	Month        int
	Name         string
	N            int64
}

// readCountyRows reads a county download as GBIF writes it: a zip holding one tab-separated
// table with a header (tech.gbifDownloadsRecorded).
func readCountyRows(path string) ([]countyRow, error) {
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
	return parseCountyRows(f)
}

func parseCountyRows(r io.Reader) ([]countyRow, error) {
	s := bufio.NewScanner(r)
	s.Buffer(make([]byte, 1<<20), 1<<20)
	col := map[string]int{}
	var rows []countyRow
	for line := 0; s.Scan(); line++ {
		f := strings.Split(s.Text(), "\t")
		if line == 0 {
			for i, name := range f {
				col[name] = i
			}
			for _, want := range []string{"datasetkey", "level2gid", "month", "v_scientificname", "n"} {
				if _, ok := col[want]; !ok {
					return nil, fmt.Errorf("the download has no %s column; it is not the county query's", want)
				}
			}
			continue
		}
		m, err1 := strconv.Atoi(f[col["month"]])
		n, err2 := strconv.ParseInt(f[col["n"]], 10, 64)
		if err1 != nil || err2 != nil || m < 1 || m > 12 {
			return nil, fmt.Errorf("line %d: month %q, count %q", line+1, f[col["month"]], f[col["n"]])
		}
		rows = append(rows, countyRow{Dataset: f[col["datasetkey"]], GID: f[col["level2gid"]],
			Month: m, Name: f[col["v_scientificname"]], N: n})
	}
	return rows, s.Err()
}

// countyPresence applies decisions.countyListingRule: each county is measured as a state is (measure:
// eBird first, iNaturalist only where eBird withholds, one per million of its records), and then a
// species stays listed only if its state lists it and it has at least minCountyRecords records
// there. Everything else with records is written under the bar, so it still counts.
func countyPresence(rows []countyRow, stateListed map[string]bool, species []string) (
	map[string][]region.Reported, map[string]int64, map[string][12]int64, error) {
	raw := map[string]map[string]map[int]map[string]int64{}
	for _, r := range rows {
		if raw[r.GID] == nil {
			raw[r.GID] = map[string]map[int]map[string]int64{}
		}
		byMonth := raw[r.GID][r.Dataset]
		if byMonth == nil {
			byMonth = map[int]map[string]int64{}
			for m := 0; m <= 12; m++ {
				byMonth[m] = map[string]int64{}
			}
			raw[r.GID][r.Dataset] = byMonth
		}
		byMonth[r.Month][r.Name] += r.N
		byMonth[0][r.Name] += r.N // the year, which measure checks the months against
	}
	aviName := make(map[string]string, len(species))
	for _, sp := range species {
		aviName[strings.ToLower(sp)] = sp
	}
	regionsOf := map[string][]region.Reported{}
	totals := map[string]int64{}
	monthly := map[string][12]int64{}
	gids := make([]string, 0, len(raw))
	for g := range raw {
		gids = append(gids, g)
	}
	sort.Strings(gids)
	for _, gid := range gids {
		m, err := measure(raw[gid], gid, aviName)
		if err != nil {
			return nil, nil, nil, err
		}
		totals[gid], monthly[gid] = m.total, m.monthlyTotal
		for _, sp := range species {
			lc := strings.ToLower(sp)
			n := m.counts[lc]
			if n == 0 {
				continue
			}
			listed := m.present[lc] && stateListed[sp] && n >= minCountyRecords
			rep := region.Reported{ID: gid, Records: n, Below: !listed}
			if mo := m.monthly[lc]; mo != nil {
				rep.Months = *mo
			}
			if listed {
				rep.Withheld = m.withheld[lc]
			}
			regionsOf[sp] = append(regionsOf[sp], rep)
		}
	}
	return regionsOf, totals, monthly, nil
}

// stateList reads which species the presence table lists in a state, under the bar or not.
func stateList(presencePath, stateGID string) (map[string]bool, error) {
	f, err := os.Open(presencePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	table, _, _, err := region.LoadPresence(f)
	if err != nil {
		return nil, err
	}
	out := map[string]bool{}
	for sp, rs := range table {
		for _, r := range rs {
			if r.ID == stateGID && !r.Below {
				out[sp] = true
			}
		}
	}
	return out, nil
}

// writeCounties builds one state's county table from a recorded download and writes it in the
// presence table's format, named by GBIF's GADM endpoint (decisions.countyListingRule, decisions.regionCodesFromData).
func writeCounties(zipPath, stateGID, presencePath, outPath string, species []string) error {
	rows, err := readCountyRows(zipPath)
	if err != nil {
		return err
	}
	listed, err := stateList(presencePath, stateGID)
	if err != nil {
		return err
	}
	if len(listed) == 0 {
		return fmt.Errorf("%s lists no species in %s; the county rule needs the state's list", presencePath, stateGID)
	}
	regionsOf, totals, monthly, err := countyPresence(rows, listed, species)
	if err != nil {
		return err
	}
	subs, err := gbif.NewClient().Subdivisions(context.Background(), stateGID)
	if err != nil {
		return err
	}
	names := map[string]string{}
	for _, s := range subs {
		names[s.ID] = s.Name
	}
	for gid := range totals {
		if names[gid] == "" {
			return fmt.Errorf("the download has records in %s, which GBIF does not list under %s", gid, stateGID)
		}
	}
	for gid := range names {
		if _, ok := totals[gid]; !ok {
			delete(names, gid) // a unit with no bird records at all names no region
		}
	}
	return write(outPath, species, regionsOf, names, totals, monthly, fmt.Sprintf(
		"# A county lists a species its state lists, at one per million of the county's records\n"+
			"# and at least %d records (decisions.countyListingRule). From GBIF download %s,\n"+
			"# whose DOI is recorded in data/gbif-downloads.tsv (tech.gbifDownloadsRecorded).\n",
		minCountyRecords, strings.TrimSuffix(filepath.Base(zipPath), ".zip")))
}
