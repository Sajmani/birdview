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
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"birdview/internal/boundary"
	"birdview/internal/region"
	"birdview/internal/shapefile"
)

// Pinned county/census-division boundary archives. See
// river/requirements/imported/census.gov/cartographic-boundaries/PROVENANCE.md
// and river/requirements/imported/statcan.gc.ca/census-divisions/PROVENANCE.md.
const (
	usCountyURL    = "https://www2.census.gov/geo/tiger/GENZ2021/shp/cb_2021_us_county_5m.zip"
	usCountySHA256 = "27444a787c0fb9c38f665293e99f7d527b7a5b4cef7a87aa3360d467be54ce38"
	usCountyStem   = "cb_2021_us_county_5m"
	usCountySource = "U.S. Census Bureau 2021 1:5,000,000 Cartographic Boundary Files (counties), public domain"

	canCountySHA256 = "1b638cea73fbf8349ba908d86be678272b2abcc74240fb0dfd9daf7a28457ad0"
	canCountyStem   = "lcd_000b21a_e"
	canCountySource = "Adapted from Statistics Canada, 2021 Census – Boundary files (Census divisions), 2021. This does not constitute an endorsement by Statistics Canada of this product."
)

// usCountyOverrides maps Census 2021 units whose names or boundaries changed
// after GADM 4.1's county list (keyed by "STUSPS|NAMELSAD") to their GADM 4.1
// level-2 ID. Multiple post-split Census polygons may map to one GADM unit.
var usCountyOverrides = map[string]string{
	"SD|Oglala Lakota County":              "USA.42.56_1", // GADM: Shannon (renamed 2015)
	"AK|Kusilvak Census Area":              "USA.2.24_1",  // GADM: Wade Hampton (renamed 2015)
	"AK|Chugach Census Area":               "USA.2.23_1",  // GADM: Valdez-Cordova (split 2019)
	"AK|Copper River Census Area":          "USA.2.23_1",  // GADM: Valdez-Cordova (split 2019)
	"AK|Skagway Municipality":              "USA.2.21_1",  // GADM: Skagway-Yakutat-Angoon (split)
	"AK|Yakutat City and Borough":          "USA.2.21_1",  // GADM: Skagway-Yakutat-Angoon (split)
	"AK|Hoonah-Angoon Census Area":         "USA.2.21_1",  // GADM: Skagway-Yakutat-Angoon (split)
	"AK|Wrangell City and Borough":         "USA.2.25_1",  // GADM: Wrangell-Petersburg (split)
	"AK|Petersburg Borough":                "USA.2.25_1",  // GADM: Wrangell-Petersburg (split)
	"AK|Prince of Wales-Hyder Census Area": "USA.2.19_1",  // GADM: Prince of Wales-Outer Ketchi
}

// statcanProvince maps Statistics Canada 2-digit PRUID to BirdView's GADM
// province/territory ID.
var statcanProvince = map[string]string{
	"10": "CAN.5_1",  // Newfoundland and Labrador
	"11": "CAN.10_1", // Prince Edward Island
	"12": "CAN.7_1",  // Nova Scotia
	"13": "CAN.4_1",  // New Brunswick
	"24": "CAN.11_1", // Quebec
	"35": "CAN.9_1",  // Ontario
	"46": "CAN.3_1",  // Manitoba
	"47": "CAN.12_1", // Saskatchewan
	"48": "CAN.1_1",  // Alberta
	"59": "CAN.2_1",  // British Columbia
	"60": "CAN.13_1", // Yukon
	"61": "CAN.6_1",  // Northwest Territories
	"62": "CAN.8_1",  // Nunavut
}

// canCountyOverrides maps Statistics Canada 2021 CDUIDs whose names or
// boundaries differ from GADM 4.1's census division list to their GADM 4.1 ID.
var canCountyOverrides = map[string]string{
	// Québec renames/reorganizations since GADM's baseline
	"2401": "CAN.11.65_1", // Les Îles-de-la-Madeleine
	"2408": "CAN.11.75_1", // La Matanie -> Matane
	"2423": "CAN.11.21_1", // Québec -> Communauté-Urbaine-de-Québec
	"2425": "CAN.11.23_1", // Lévis -> Desjardins (merged with CAN.11.62_1)
	"2431": "CAN.11.29_1", // Les Appalaches -> L'Amiante
	"2436": "CAN.11.51_1", // Shawinigan -> Le Centre-de-la-Mauricie
	"2440": "CAN.11.7_1",  // Les Sources -> Asbestos
	"2443": "CAN.11.43_1", // Sherbrooke -> La Région-Sherbrookoise
	"2453": "CAN.11.50_1", // Pierre-De Saurel -> Le Bas-Richelieu
	"2458": "CAN.11.15_1", // Longueuil -> Champlain
	"2459": "CAN.11.48_1", // Marguerite-D'Youville -> Lajemmerais
	"2466": "CAN.11.20_1", // Montréal -> Communauté-Urbaine-de-Montréal
	"2481": "CAN.11.19_1", // Gatineau -> Communauté-Urbaine-de-l'Outaouai
	"2489": "CAN.11.98_1", // La Vallée-de-l'Or -> Vallée-de-l'Or
	"2490": "CAN.11.58_1", // La Tuque -> Le Haut-Saint-Maurice
	"2494": "CAN.11.53_1", // Le Saguenay-et-son-Fjord -> Le Fjord-du-Saguenay
	"2498": "CAN.11.79_1", // Minganie--Le Golfe-du-Saint-Laurent -> Minganie--Basse-Côte-Nord
	// British Columbia split of Comox-Strathcona (2008)
	"5924": "CAN.2.9_1", // Strathcona -> Comox-Strathcona
	"5926": "CAN.2.9_1", // Comox Valley -> Comox-Strathcona
	// Ontario bilingual name
	"3553": "CAN.9.11_1", // Greater Sudbury / Grand Sudbury -> Greater Sudbury
	// Nunavut Inuit names
	"6204": "CAN.8.1_1", // Qikiqtaaluk -> Baffin
	"6205": "CAN.8.2_1", // Kivalliq -> Keewatin
	// Northwest Territories 6 regions -> GADM's 2 historical regions
	"6101": "CAN.6.2_1", // Region 1 -> Inuvik
	"6102": "CAN.6.2_1", // Region 2 -> Inuvik
	"6103": "CAN.6.1_1", // Region 3 -> Fort Smith
	"6104": "CAN.6.1_1", // Region 4 -> Fort Smith
	"6105": "CAN.6.1_1", // Region 5 -> Fort Smith
	"6106": "CAN.6.1_1", // Region 6 -> Fort Smith
}

// countyBoundaryExempt lists GADM 4.1 level-2 units in data/county-presence/
// that have no distinct land polygon in 2021 Census or Statistics Canada
// cartographic boundaries: Great Lakes water-only units and cities merged into
// their surrounding county/division before 2021. Positions on Great Lakes
// water fall back to the state or province.
var countyBoundaryExempt = map[string]string{
	// US Great Lakes water units
	"USA.14.51_1": "Lake Michigan (IL water unit)",
	"USA.15.46_1": "Lake Michigan (IN water unit)",
	"USA.23.44_1": "Lake Huron (MI water unit)",
	"USA.23.45_1": "Lake Michigan (MI water unit)",
	"USA.23.46_1": "Lake St. Clair (MI water unit)",
	"USA.23.47_1": "Lake Superior (MI water unit)",
	"USA.24.40_1": "Lake Superior (MN water unit)",
	"USA.33.25_1": "Lake Ontario (NY water unit)",
	"USA.36.44_1": "Lake Erie (OH water unit)",
	"USA.50.34_1": "Lake Michigan (WI water unit)",
	"USA.50.35_1": "Lake Superior (WI water unit)",
	// US obsolete Virginia independent cities (reverted to towns in their counties)
	"USA.47.12_1": "Bedford City (reverted to Bedford County in 2013)",
	"USA.47.29_1": "Clifton Forge City (reverted to Alleghany County in 2001)",
	// Canadian Great Lakes water units
	"CAN.9.21_1": "Lake Erie (ON water unit)",
	"CAN.9.22_1": "Lake Huron (ON water unit)",
	"CAN.9.23_1": "Lake Ontario (ON water unit)",
	"CAN.9.24_1": "Lake Superior (ON water unit)",
	// Québec merged MRC (merged with CAN.11.23_1 Desjardins into Lévis CDUID 2425)
	"CAN.11.62_1": "Les Chutes-de-la-Chaudière (merged into Lévis, mapped to CAN.11.23_1)",
}

// runCounties builds per-state county boundary files into outDir.
func runCounties(usArchive, canArchive, presenceDir, outDir string, tolerance int32) error {
	if canArchive == "" {
		return fmt.Errorf("-can-archive is required with -counties (download lcd_000b21a_e.zip from Statistics Canada into fetched/)")
	}
	usZip, err := loadChecked(usArchive, usCountyURL, usCountySHA256)
	if err != nil {
		return fmt.Errorf("US county archive: %w", err)
	}
	canZip, err := loadChecked(canArchive, "", canCountySHA256)
	if err != nil {
		return fmt.Errorf("Canada census-division archive: %w", err)
	}
	presence, err := readCountyPresenceHeaders(presenceDir)
	if err != nil {
		return err
	}
	files, err := buildCounties(usZip, canZip, presence, tolerance)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return err
	}
	states := make([]string, 0, len(files))
	for st := range files {
		states = append(states, st)
	}
	sort.Strings(states)
	var maxGZ, totalGZ int
	var maxState string
	for _, st := range states {
		b, err := encode(files[st])
		if err != nil {
			return fmt.Errorf("encoding %s: %w", st, err)
		}
		gz, err := gzipReproducible(b)
		if err != nil {
			return fmt.Errorf("compressing %s: %w", st, err)
		}
		path := filepath.Join(outDir, st+".json.gz")
		if err := os.WriteFile(path, gz, 0o644); err != nil {
			return err
		}
		totalGZ += len(gz)
		if len(gz) > maxGZ {
			maxGZ = len(gz)
			maxState = st
		}
	}
	log.Printf("wrote %d state county-boundary files to %s: total %d bytes compressed, largest %s (%d bytes)",
		len(states), outDir, totalGZ, maxState, maxGZ)
	return nil
}

// loadChecked reads an archive from path (or downloads url if path is empty)
// and verifies its SHA-256 against wantSHA.
func loadChecked(path, url, wantSHA string) ([]byte, error) {
	var b []byte
	var err error
	if path != "" {
		b, err = os.ReadFile(path)
	} else if url != "" {
		b, err = downloadURL(url)
	} else {
		return nil, fmt.Errorf("archive path is required")
	}
	if err != nil {
		return nil, err
	}
	sum := sha256.Sum256(b)
	if got := hex.EncodeToString(sum[:]); got != wantSHA {
		return nil, fmt.Errorf("archive SHA-256 is %s, want the pinned %s; a new release is a refresh, not a drop-in", got, wantSHA)
	}
	return b, nil
}

func downloadURL(url string) ([]byte, error) {
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

type gadmCounty struct {
	id   string
	name string
}

// readCountyPresenceHeaders reads every *.tsv.gz under dir and returns each
// state's GADM level-2 counties from its "#region\t<id>\t<name>" lines.
func readCountyPresenceHeaders(dir string) (map[string][]gadmCounty, error) {
	matches, err := filepath.Glob(filepath.Join(dir, "*.tsv.gz"))
	if err != nil {
		return nil, err
	}
	if len(matches) == 0 {
		return nil, fmt.Errorf("no *.tsv.gz files in %s", dir)
	}
	out := make(map[string][]gadmCounty, len(matches))
	for _, path := range matches {
		st := strings.TrimSuffix(filepath.Base(path), ".tsv.gz")
		if r, ok := region.Get(st); !ok || r.Kind != region.Land {
			return nil, fmt.Errorf("%s: %q is not a BirdView land region", path, st)
		}
		f, err := os.Open(path)
		if err != nil {
			return nil, err
		}
		gz, err := gzip.NewReader(f)
		if err != nil {
			f.Close()
			return nil, fmt.Errorf("%s: %w", path, err)
		}
		sc := bufio.NewScanner(gz)
		sc.Buffer(make([]byte, 64*1024), 4*1024*1024)
		for sc.Scan() {
			line := sc.Text()
			if !strings.HasPrefix(line, "#") {
				break // header finished
			}
			if !strings.HasPrefix(line, "#region\t") {
				continue
			}
			parts := strings.Split(line, "\t")
			if len(parts) < 3 {
				gz.Close()
				f.Close()
				return nil, fmt.Errorf("%s: malformed #region line %q", path, line)
			}
			out[st] = append(out[st], gadmCounty{id: parts[1], name: parts[2]})
		}
		err = sc.Err()
		gz.Close()
		f.Close()
		if err != nil {
			return nil, fmt.Errorf("%s: %w", path, err)
		}
	}
	return out, nil
}

// buildCounties turns the US and Canadian county archives into one boundary.File
// per state or province.
func buildCounties(usZip, canZip []byte, presence map[string][]gadmCounty, tolerance int32) (map[string]*boundary.File, error) {
	files := make(map[string]*boundary.File, len(presence))
	if err := addUSCounties(files, usZip, presence, tolerance); err != nil {
		return nil, err
	}
	if err := addCANCounties(files, canZip, presence, tolerance); err != nil {
		return nil, err
	}
	// Verify every non-exempt county in presence received at least one ring.
	var missing []string
	for st, counties := range presence {
		f, ok := files[st]
		if !ok {
			missing = append(missing, fmt.Sprintf("entire state %s", st))
			continue
		}
		have := make(map[string]bool, len(f.Land))
		for _, l := range f.Land {
			if len(l.Rings) > 0 {
				have[l.ID] = true
			}
		}
		for _, c := range counties {
			if _, exempt := countyBoundaryExempt[c.id]; exempt {
				continue
			}
			if !have[c.id] {
				missing = append(missing, fmt.Sprintf("%s (%s %s)", c.id, st, c.name))
			}
		}
	}
	if len(missing) > 0 {
		sort.Strings(missing)
		return nil, fmt.Errorf("%d counties in county-presence have no boundary: %s", len(missing), strings.Join(missing, ", "))
	}
	return files, nil
}

func readShapefileZip(zipBytes []byte, stem string) ([][][][2]float64, []map[string]string, error) {
	zr, err := zip.NewReader(bytes.NewReader(zipBytes), int64(len(zipBytes)))
	if err != nil {
		return nil, nil, err
	}
	read := func(name string) ([]byte, error) {
		f, err := zr.Open(name)
		if err != nil {
			return nil, err
		}
		defer f.Close()
		return io.ReadAll(f)
	}
	shp, err := read(stem + ".shp")
	if err != nil {
		return nil, nil, err
	}
	dbf, err := read(stem + ".dbf")
	if err != nil {
		return nil, nil, err
	}
	shapes, err := shapefile.Polygons(shp)
	if err != nil {
		return nil, nil, err
	}
	records, err := shapefile.Records(dbf)
	if err != nil {
		return nil, nil, err
	}
	if len(shapes) != len(records) {
		return nil, nil, fmt.Errorf("%s: %d shapes but %d attribute records", stem, len(shapes), len(records))
	}
	return shapes, records, nil
}

type stateMatcher struct {
	byID      map[string]string
	byNorm    map[string][]string
	byCompact map[string][]string
}

func newStateMatcher(counties []gadmCounty) *stateMatcher {
	m := &stateMatcher{
		byID:      make(map[string]string, len(counties)),
		byNorm:    make(map[string][]string, len(counties)),
		byCompact: make(map[string][]string, len(counties)),
	}
	for _, c := range counties {
		m.byID[c.id] = c.name
		n := normPlaceName(c.name)
		m.byNorm[n] = append(m.byNorm[n], c.id)
		cmp := strings.ReplaceAll(n, " ", "")
		m.byCompact[cmp] = append(m.byCompact[cmp], c.id)
	}
	return m
}

func (m *stateMatcher) match(candidates ...string) (string, error) {
	for _, raw := range candidates {
		if raw == "" {
			continue
		}
		n := normPlaceName(raw)
		if ids := m.byNorm[n]; len(ids) == 1 {
			return ids[0], nil
		} else if len(ids) > 1 {
			return "", fmt.Errorf("ambiguous match for %q (%q): %v", raw, n, ids)
		}
		cmp := strings.ReplaceAll(n, " ", "")
		if ids := m.byCompact[cmp]; len(ids) == 1 {
			return ids[0], nil
		} else if len(ids) > 1 {
			return "", fmt.Errorf("ambiguous compact match for %q (%q): %v", raw, cmp, ids)
		}
	}
	return "", nil
}

func addUSCounties(files map[string]*boundary.File, zipBytes []byte, presence map[string][]gadmCounty, tolerance int32) error {
	shapes, records, err := readShapefileZip(zipBytes, usCountyStem)
	if err != nil {
		return err
	}
	usByPostal := map[string]string{}
	for _, id := range region.LandIDs() {
		if !strings.HasPrefix(id, "USA.") {
			continue
		}
		r, _ := region.Get(id)
		if r.Code != "" {
			usByPostal[r.Code] = id
		}
	}
	matchers := map[string]*stateMatcher{}
	byState := map[string]map[string]*boundary.Land{}
	for st, counties := range presence {
		if strings.HasPrefix(st, "USA.") {
			matchers[st] = newStateMatcher(counties)
			byState[st] = map[string]*boundary.Land{}
		}
	}

	for i, rec := range records {
		stusps := rec["STUSPS"]
		stID, ok := usByPostal[stusps]
		if !ok {
			continue // PR, VI, GU, MP, AS: outside BirdView's land regions
		}
		name := decodeLatin1(rec["NAME"])
		namelsad := decodeLatin1(rec["NAMELSAD"])
		m := matchers[stID]

		var id string
		if ov, ok := usCountyOverrides[stusps+"|"+namelsad]; ok {
			if _, exists := m.byID[ov]; !exists {
				return fmt.Errorf("US override %s|%s -> %s not in %s", stusps, namelsad, ov, stID)
			}
			id = ov
		} else {
			// Try NAMELSAD first so e.g. Virginia's "Fairfax city" matches
			// GADM's "Fairfax City" rather than "Fairfax" county.
			var err error
			id, err = m.match(namelsad, name)
			if err != nil {
				return fmt.Errorf("US %s %s (%s): %w", stusps, namelsad, rec["GEOID"], err)
			}
			if id == "" {
				return fmt.Errorf("unmatched US county: %s %q (NAMELSAD=%q GEOID=%s)", stusps, name, namelsad, rec["GEOID"])
			}
		}

		l := byState[stID][id]
		if l == nil {
			l = &boundary.Land{Region: id, ID: id}
			byState[stID][id] = l
		}
		for _, ring := range shapes[i] {
			if enc := encodeCountyRing(ring, tolerance); enc != nil {
				l.Rings = append(l.Rings, enc)
			}
		}
		if len(l.Rings) == 0 {
			for _, ring := range shapes[i] {
				if enc := encodeRing(ring, tolerance); enc != nil {
					l.Rings = append(l.Rings, enc)
				}
			}
		}
	}

	for stID, lands := range byState {
		f := &boundary.File{Source: usCountySource}
		for _, l := range lands {
			f.Land = append(f.Land, *l)
		}
		sort.Slice(f.Land, func(i, j int) bool { return f.Land[i].Region < f.Land[j].Region })
		files[stID] = f
	}
	return nil
}

func addCANCounties(files map[string]*boundary.File, zipBytes []byte, presence map[string][]gadmCounty, tolerance int32) error {
	shapes, records, err := readShapefileZip(zipBytes, canCountyStem)
	if err != nil {
		return err
	}
	matchers := map[string]*stateMatcher{}
	byState := map[string]map[string]*boundary.Land{}
	for st, counties := range presence {
		if strings.HasPrefix(st, "CAN.") {
			matchers[st] = newStateMatcher(counties)
			byState[st] = map[string]*boundary.Land{}
		}
	}

	for i, rec := range records {
		pruid := rec["PRUID"]
		stID, ok := statcanProvince[pruid]
		if !ok {
			return fmt.Errorf("unknown Statistics Canada PRUID %q (CDUID %s)", pruid, rec["CDUID"])
		}
		cduid := rec["CDUID"]
		cdname := decodeLatin1(rec["CDNAME"])
		m := matchers[stID]

		var id string
		if ov, ok := canCountyOverrides[cduid]; ok {
			if _, exists := m.byID[ov]; !exists {
				return fmt.Errorf("CAN override %s (%s) -> %s not in %s", cduid, cdname, ov, stID)
			}
			id = ov
		} else {
			var err error
			id, err = m.match(cdname)
			if err != nil {
				return fmt.Errorf("CAN %s %s (%s): %w", stID, cdname, cduid, err)
			}
			if id == "" {
				return fmt.Errorf("unmatched CAN census division: %s CDUID=%s CDNAME=%q", stID, cduid, cdname)
			}
		}

		l := byState[stID][id]
		if l == nil {
			l = &boundary.Land{Region: id, ID: id}
			byState[stID][id] = l
		}
		for _, ring := range shapes[i] {
			deg := make([][2]float64, len(ring))
			for k, p := range ring {
				lon, lat := lambertToLonLat(p[0], p[1])
				deg[k] = [2]float64{lon, lat}
			}
			if enc := encodeCountyRing(deg, tolerance); enc != nil {
				l.Rings = append(l.Rings, enc)
			}
		}
		if len(l.Rings) == 0 {
			for _, ring := range shapes[i] {
				deg := make([][2]float64, len(ring))
				for k, p := range ring {
					lon, lat := lambertToLonLat(p[0], p[1])
					deg[k] = [2]float64{lon, lat}
				}
				if enc := encodeRing(deg, tolerance); enc != nil {
					l.Rings = append(l.Rings, enc)
				}
			}
		}
	}

	for stID, lands := range byState {
		f := &boundary.File{Source: canCountySource}
		for _, l := range lands {
			f.Land = append(f.Land, *l)
		}
		sort.Slice(f.Land, func(i, j int) bool { return f.Land[i].Region < f.Land[j].Region })
		files[stID] = f
	}
	return nil
}

// encodeCountyRing is encodeRing for county and census-division polygons: rings
// that collapse below a triangle under simplification (tiny uninhabited rocky
// islets and skerries in Statistics Canada's 1:50,000 shoreline) are dropped
// rather than restored unsimplified.
func encodeCountyRing(ring [][2]float64, tolerance int32) []int32 {
	var pts [][2]int32
	for _, p := range ring {
		q := [2]int32{int32(math.Round(p[0] * boundary.Scale)), int32(math.Round(p[1] * boundary.Scale))}
		if len(pts) == 0 || pts[len(pts)-1] != q {
			pts = append(pts, q)
		}
	}
	if len(pts) > 1 && pts[0] == pts[len(pts)-1] {
		pts = pts[:len(pts)-1]
	}
	if len(pts) < 3 {
		return nil
	}
	var lat float64
	for _, p := range pts {
		lat += float64(p[1])
	}
	squash := math.Cos(lat / float64(len(pts)) / boundary.Scale * math.Pi / 180)
	simple := simplify(pts, tolerance, squash)
	if len(simple) < 3 {
		return nil
	}
	out := make([]int32, 0, 2*len(simple))
	var px, py int32
	for _, p := range simple {
		out = append(out, p[0]-px, p[1]-py)
		px, py = p[0], p[1]
	}
	return out
}

// gzipReproducible compresses b with a deterministic gzip header (zero ModTime,
// empty Name) so regenerating produces byte-identical files.
func gzipReproducible(b []byte) ([]byte, error) {
	var buf bytes.Buffer
	w, err := gzip.NewWriterLevel(&buf, gzip.BestCompression)
	if err != nil {
		return nil, err
	}
	w.Header.ModTime = time.Time{}
	w.Header.OS = 255
	if _, err := w.Write(b); err != nil {
		return nil, err
	}
	if err := w.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// decodeLatin1 converts a dBASE string (ISO-8859-1 / Windows-1252) to UTF-8 if
// it is not already valid UTF-8.
func decodeLatin1(s string) string {
	if utf8.ValidString(s) {
		return s
	}
	var b strings.Builder
	b.Grow(len(s) * 2)
	for i := 0; i < len(s); i++ {
		b.WriteRune(rune(s[i]))
	}
	return b.String()
}

// normPlaceName normalizes a US county or Canadian census-division name for
// matching against GADM 4.1: folds diacritics, lowercases, turns punctuation
// into spaces, and standardizes Saint/Sainte to St/Ste.
func normPlaceName(s string) string {
	var b strings.Builder
	lastSpace := true
	for _, r := range strings.ToLower(s) {
		r = foldLatin(r)
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
			lastSpace = false
		} else if !lastSpace {
			b.WriteByte(' ')
			lastSpace = true
		}
	}
	words := strings.Fields(b.String())
	for i, w := range words {
		switch w {
		case "saint":
			words[i] = "st"
		case "sainte":
			words[i] = "ste"
		}
	}
	return strings.Join(words, " ")
}

func foldLatin(r rune) rune {
	switch r {
	case 'à', 'á', 'â', 'ã', 'ä', 'å':
		return 'a'
	case 'ç':
		return 'c'
	case 'è', 'é', 'ê', 'ë':
		return 'e'
	case 'ì', 'í', 'î', 'ï':
		return 'i'
	case 'ñ':
		return 'n'
	case 'ò', 'ó', 'ô', 'õ', 'ö':
		return 'o'
	case 'ù', 'ú', 'û', 'ü':
		return 'u'
	case 'ý', 'ÿ':
		return 'y'
	}
	return r
}

// NAD83_Statistics_Canada_Lambert (EPSG:3347) projection constants from
// lcd_000b21a_e.prj.
const (
	lccA    = 6378137.0
	lccInvF = 298.2572221008916
	lccPhi1 = 49.0 * math.Pi / 180.0
	lccPhi2 = 77.0 * math.Pi / 180.0
	lccPhi0 = 63.390675 * math.Pi / 180.0
	lccLam0 = -91.86666666666666 * math.Pi / 180.0
	lccFE   = 6200000.0
	lccFN   = 3000000.0
)

var (
	lccE    float64
	lccN    float64
	lccF    float64
	lccRho0 float64
)

func init() {
	f := 1.0 / lccInvF
	lccE = math.Sqrt(2*f - f*f)
	m := func(phi float64) float64 {
		s := math.Sin(phi)
		return math.Cos(phi) / math.Sqrt(1-lccE*lccE*s*s)
	}
	t := func(phi float64) float64 {
		es := lccE * math.Sin(phi)
		return math.Tan(math.Pi/4-phi/2) / math.Pow((1-es)/(1+es), lccE/2)
	}
	m1, m2 := m(lccPhi1), m(lccPhi2)
	t1, t2, t0 := t(lccPhi1), t(lccPhi2), t(lccPhi0)
	lccN = (math.Log(m1) - math.Log(m2)) / (math.Log(t1) - math.Log(t2))
	lccF = m1 / (lccN * math.Pow(t1, lccN))
	lccRho0 = lccA * lccF * math.Pow(t0, lccN)
}

// lambertToLonLat inverse-projects a point (x, y) in meters in
// NAD83_Statistics_Canada_Lambert (EPSG:3347) to WGS84/NAD83 (lon, lat) degrees.
func lambertToLonLat(x, y float64) (lon, lat float64) {
	dx := x - lccFE
	dy := lccRho0 - (y - lccFN)
	rho := math.Hypot(dx, dy)
	if lccN < 0 {
		rho = -rho
		dx = -dx
		dy = -dy
	}
	theta := math.Atan2(dx, dy)
	lam := lccLam0 + theta/lccN
	t := math.Pow(rho/(lccA*lccF), 1/lccN)
	phi := math.Pi/2 - 2*math.Atan(t)
	for range 8 {
		es := lccE * math.Sin(phi)
		next := math.Pi/2 - 2*math.Atan(t*math.Pow((1-es)/(1+es), lccE/2))
		if math.Abs(next-phi) < 1e-13 {
			phi = next
			break
		}
		phi = next
	}
	return lam * 180 / math.Pi, phi * 180 / math.Pi
}
