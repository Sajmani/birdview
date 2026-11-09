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
// Package shapefile reads the two parts of an ESRI shapefile BirdView uses: polygon geometry and its
// dBASE attribute table. Natural Earth publishes in this form (tech.regionBoundaries,
// tech.seaDistanceTable).
package shapefile

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"strings"
)

// Polygons reads a polygon shapefile (.shp): for each record, its rings as [lon, lat] points. A
// null shape has no rings.
func Polygons(b []byte) ([][][][2]float64, error) {
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

// Records reads a dBASE table (.dbf) into one map per record, values trimmed.
func Records(b []byte) ([]map[string]string, error) {
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
