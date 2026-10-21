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

// Command birdview-boundaries writes data/region-boundaries.json, the region
// boundaries the Here button tests a reader's position against
// (tech.regionBoundaries).
//
// It reads Natural Earth's 1:10m states and provinces, pinned by hash and
// refused under any other, keeps the units that are BirdView regions,
// simplifies each ring and stores it in thousandths of a degree. The sea
// regions are BirdView's own boxes, copied from internal/region so the two
// cannot drift.
//
//	birdview-boundaries [-archive ne_10m_admin_1_states_provinces.zip] [-out data/region-boundaries.json]
//
// Without -archive it downloads the archive.
package main

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"sort"
	"strings"

	"birdview/internal/boundary"
	"birdview/internal/region"
)

// The pinned source. See river/requirements/imported/naturalearthdata.com/naturalearth/PROVENANCE.md.
const (
	archiveURL    = "https://naciscdn.org/naturalearth/10m/cultural/ne_10m_admin_1_states_provinces.zip"
	archiveSHA256 = "efc59726337323058f9446210adc96673179cd344e053666ee3d28cb58ba2b05"
	stem          = "ne_10m_admin_1_states_provinces"
	userAgent     = "birdview-boundaries/1.0 (contact: birdview@ajmani.net)" // tech.identifiedClient
)

// atolls maps Natural Earth's ISO 3166-2 codes for the US Minor Outlying
// Islands to BirdView's GADM ids. They have no postal code to match on.
var atolls = map[string]string{
	"UM-67": "UMI.4_1", // Johnston Atoll
	"UM-71": "UMI.6_1", // Midway (Natural Earth: Midway Islands)
	"UM-76": "UMI.7_1", // Navassa Island
	"UM-95": "UMI.8_1", // Palmyra Atoll
	"UM-79": "UMI.9_1", // Wake Island (Natural Earth: Wake Atoll)
}

func main() {
	archive := flag.String("archive", "", "the Natural Earth archive; downloaded if empty")
	out := flag.String("out", "data/region-boundaries.json", "where to write the boundary file")
	tolerance := flag.Int("tolerance", 4, "simplification tolerance, in thousandths of a degree on the ground (4 is about 450 m)")
	flag.Parse()

	zipBytes, err := load(*archive)
	if err != nil {
		log.Fatal(err)
	}
	f, err := build(zipBytes, int32(*tolerance))
	if err != nil {
		log.Fatal(err)
	}
	b, err := encode(f)
	if err != nil {
		log.Fatal(err)
	}
	if err := os.WriteFile(*out, b, 0o644); err != nil {
		log.Fatal(err)
	}
	log.Printf("wrote %s: %d land regions, %d sea regions, %d bytes", *out, len(f.Land), len(f.Sea), len(b))
}

// load reads the archive from disk or the network and refuses it unless it is
// the pinned one.
func load(path string) ([]byte, error) {
	var b []byte
	var err error
	if path != "" {
		b, err = os.ReadFile(path)
	} else {
		b, err = download()
	}
	if err != nil {
		return nil, err
	}
	sum := sha256.Sum256(b)
	if got := hex.EncodeToString(sum[:]); got != archiveSHA256 {
		return nil, fmt.Errorf("archive SHA-256 is %s, want the pinned %s; a new release is a refresh, not a drop-in", got, archiveSHA256)
	}
	return b, nil
}

func download() ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, archiveURL, nil)
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
		return nil, fmt.Errorf("GET %s: %s", archiveURL, resp.Status)
	}
	return io.ReadAll(resp.Body)
}

// build turns the archive into a boundary file.
func build(zipBytes []byte, tolerance int32) (*boundary.File, error) {
	zr, err := zip.NewReader(bytes.NewReader(zipBytes), int64(len(zipBytes)))
	if err != nil {
		return nil, err
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
		return nil, err
	}
	dbf, err := read(stem + ".dbf")
	if err != nil {
		return nil, err
	}
	shapes, err := readShapes(shp)
	if err != nil {
		return nil, err
	}
	records, err := readRecords(dbf)
	if err != nil {
		return nil, err
	}
	if len(shapes) != len(records) {
		return nil, fmt.Errorf("%d shapes but %d attribute records", len(shapes), len(records))
	}

	// BirdView's land regions, by the key Natural Earth can be matched on.
	byKey := map[string]region.Region{}
	for _, id := range region.LandIDs() {
		r, _ := region.Get(id)
		country, _, _ := strings.Cut(id, ".")
		if r.Code != "" {
			byKey[country+" "+r.Code] = r
		}
	}
	for iso, id := range atolls {
		r, ok := region.Get(id)
		if !ok {
			return nil, fmt.Errorf("atoll %s maps to %s, which is not a region", iso, id)
		}
		byKey["UMI "+iso] = r
	}

	lands := map[string]*boundary.Land{}
	for i, rec := range records {
		var key string
		switch rec["adm0_a3"] {
		case "USA", "CAN":
			key = rec["adm0_a3"] + " " + rec["postal"]
		case "UMI":
			key = "UMI " + rec["iso_3166_2"]
		default:
			continue
		}
		r, ok := byKey[key]
		if !ok {
			continue // Jarvis, Baker, Howland: not BirdView regions
		}
		l := lands[r.ID]
		if l == nil {
			l = &boundary.Land{Region: r.Slug(), ID: r.ID}
			lands[r.ID] = l
		}
		for _, ring := range shapes[i] {
			if enc := encodeRing(ring, tolerance); enc != nil {
				l.Rings = append(l.Rings, enc)
			}
		}
	}
	f := &boundary.File{Source: "Natural Earth 1:10m admin-1 states and provinces, version 5.1.1, public domain"}
	for _, l := range lands {
		f.Land = append(f.Land, *l)
	}
	sort.Slice(f.Land, func(i, j int) bool { return f.Land[i].Region < f.Land[j].Region })
	for _, id := range region.MarineIDs() {
		r, _ := region.Get(id)
		minLat, maxLat, minLon, maxLon, _ := region.MarineBox(id)
		f.Sea = append(f.Sea, boundary.Sea{Region: r.Slug(), ID: id, Box: [4]float64{minLat, maxLat, minLon, maxLon}})
	}
	return f, nil
}

// encode writes the file with one region per line, so a regeneration diffs
// region by region rather than as one enormous line.
func encode(f *boundary.File) ([]byte, error) {
	var b bytes.Buffer
	src, _ := json.Marshal(f.Source)
	fmt.Fprintf(&b, "{\"source\":%s,\n\"land\":[\n", src)
	for i, l := range f.Land {
		j, err := json.Marshal(l)
		if err != nil {
			return nil, err
		}
		b.Write(j)
		if i < len(f.Land)-1 {
			b.WriteByte(',')
		}
		b.WriteByte('\n')
	}
	b.WriteString("],\n\"sea\":[\n")
	for i, s := range f.Sea {
		j, err := json.Marshal(s)
		if err != nil {
			return nil, err
		}
		b.Write(j)
		if i < len(f.Sea)-1 {
			b.WriteByte(',')
		}
		b.WriteByte('\n')
	}
	b.WriteString("]}\n")
	return b.Bytes(), nil
}

// encodeRing quantizes a ring to thousandths of a degree, simplifies it, and
// delta-encodes it. A ring that collapses below a triangle is dropped: at
// this resolution it is an islet under about a hundred meters across.
func encodeRing(ring [][2]float64, tolerance int32) []int32 {
	var pts [][2]int32
	for _, p := range ring {
		q := [2]int32{int32(math.Round(p[0] * boundary.Scale)), int32(math.Round(p[1] * boundary.Scale))}
		if len(pts) == 0 || pts[len(pts)-1] != q {
			pts = append(pts, q)
		}
	}
	if len(pts) > 1 && pts[0] == pts[len(pts)-1] {
		pts = pts[:len(pts)-1] // the closing point is implied
	}
	if len(pts) < 3 {
		return nil
	}
	// Distances are taken on the ground, not in raw degrees: a degree of
	// longitude at 70°N is a third of one at the equator, and simplifying in
	// degrees kept three times the east-west detail in the Arctic.
	var lat float64
	for _, p := range pts {
		lat += float64(p[1])
	}
	squash := math.Cos(lat / float64(len(pts)) / boundary.Scale * math.Pi / 180)
	simple := simplify(pts, tolerance, squash)
	if len(simple) < 3 {
		simple = pts // small enough that simplifying would erase it
	}
	out := make([]int32, 0, 2*len(simple))
	var px, py int32
	for _, p := range simple {
		out = append(out, p[0]-px, p[1]-py)
		px, py = p[0], p[1]
	}
	return out
}

// simplify is Douglas-Peucker over a closed ring, anchored at its first point
// and the point farthest from it. Longitude differences are multiplied by
// squash, the cosine of the ring's latitude, so the tolerance is the same
// distance on the ground everywhere.
func simplify(pts [][2]int32, tolerance int32, squash float64) [][2]int32 {
	far := 0
	var best float64
	for i, p := range pts {
		if d := dist2(p, pts[0], squash); d > best {
			best, far = d, i
		}
	}
	if far == 0 {
		return pts[:1]
	}
	keep := make([]bool, len(pts))
	keep[0], keep[far] = true, true
	closed := append(append([][2]int32{}, pts...), pts[0])
	dp(closed, 0, far, float64(tolerance), squash, keep)
	dp(closed, far, len(pts), float64(tolerance), squash, keep)
	var out [][2]int32
	for i, k := range keep {
		if k {
			out = append(out, pts[i])
		}
	}
	return out
}

func dp(pts [][2]int32, a, b int, tol, squash float64, keep []bool) {
	if b-a < 2 {
		return
	}
	idx, max := -1, tol
	for i := a + 1; i < b; i++ {
		if d := segDist(pts[i], pts[a], pts[b], squash); d > max {
			idx, max = i, d
		}
	}
	if idx < 0 {
		return
	}
	keep[idx%len(keep)] = true
	dp(pts, a, idx, tol, squash, keep)
	dp(pts, idx, b, tol, squash, keep)
}

func dist2(p, q [2]int32, squash float64) float64 {
	dx, dy := float64(p[0]-q[0])*squash, float64(p[1]-q[1])
	return dx*dx + dy*dy
}

// segDist is the distance from p to the segment ab, longitude squashed.
func segDist(p, a, b [2]int32, squash float64) float64 {
	ax, ay := float64(a[0])*squash, float64(a[1])
	bx, by := float64(b[0])*squash, float64(b[1])
	px, py := float64(p[0])*squash, float64(p[1])
	dx, dy := bx-ax, by-ay
	if dx == 0 && dy == 0 {
		return math.Hypot(px-ax, py-ay)
	}
	t := math.Max(0, math.Min(1, ((px-ax)*dx+(py-ay)*dy)/(dx*dx+dy*dy)))
	return math.Hypot(px-(ax+t*dx), py-(ay+t*dy))
}

// readShapes reads a polygon shapefile: for each record, its rings as
// [lon, lat] points. A null shape has no rings.
func readShapes(b []byte) ([][][][2]float64, error) {
	if len(b) < 100 {
		return nil, fmt.Errorf("shapefile too short")
	}
	if t := binary.LittleEndian.Uint32(b[32:]); t != 5 {
		return nil, fmt.Errorf("shape type %d, want 5 (polygon)", t)
	}
	var out [][][][2]float64
	for off := 100; off+8 <= len(b); {
		n := int(binary.BigEndian.Uint32(b[off+4:])) * 2
		c := b[off+8 : off+8+n]
		off += 8 + n
		if binary.LittleEndian.Uint32(c) == 0 {
			out = append(out, nil)
			continue
		}
		parts := int(binary.LittleEndian.Uint32(c[36:]))
		points := int(binary.LittleEndian.Uint32(c[40:]))
		starts := make([]int, parts+1)
		for i := 0; i < parts; i++ {
			starts[i] = int(binary.LittleEndian.Uint32(c[44+4*i:]))
		}
		starts[parts] = points
		base := 44 + 4*parts
		var rings [][][2]float64
		for i := 0; i < parts; i++ {
			var ring [][2]float64
			for k := starts[i]; k < starts[i+1]; k++ {
				x := math.Float64frombits(binary.LittleEndian.Uint64(c[base+16*k:]))
				y := math.Float64frombits(binary.LittleEndian.Uint64(c[base+16*k+8:]))
				ring = append(ring, [2]float64{x, y})
			}
			rings = append(rings, ring)
		}
		out = append(out, rings)
	}
	return out, nil
}

// readRecords reads a dBASE table into one map per record, values trimmed.
func readRecords(b []byte) ([]map[string]string, error) {
	if len(b) < 32 {
		return nil, fmt.Errorf("dbf too short")
	}
	n := int(binary.LittleEndian.Uint32(b[4:]))
	hlen := int(binary.LittleEndian.Uint16(b[8:]))
	rlen := int(binary.LittleEndian.Uint16(b[10:]))
	type field struct {
		name string
		len  int
	}
	var fields []field
	for off := 32; off < hlen && b[off] != 0x0D; off += 32 {
		name, _, _ := bytes.Cut(b[off:off+11], []byte{0}) // NUL-terminated, padding after
		fields = append(fields, field{string(name), int(b[off+16])})
	}
	out := make([]map[string]string, 0, n)
	for i := 0; i < n; i++ {
		r := b[hlen+i*rlen : hlen+(i+1)*rlen]
		rec := map[string]string{}
		pos := 1
		for _, f := range fields {
			rec[f.name] = strings.Trim(string(r[pos:pos+f.len]), " \x00") // padded with either
			pos += f.len
		}
		out = append(out, rec)
	}
	return out, nil
}
