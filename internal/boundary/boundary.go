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

// Package boundary is the region boundary file the Here button tests a
// reader's position against (tech.regionBoundaries), and the reference lookup
// over it.
//
// The lookup that matters runs in the reader's browser, because the position
// must never leave it (product.locationStaysInBrowser). This one exists so the
// file can be checked in Go against places whose region is known, and so the
// browser's lookup can be held to the same answers.
//
// Coordinates are stored as integers in thousandths of a degree, each ring
// as an absolute first point followed by deltas, which keeps the file small
// enough to load with every index page.
//
// Verifies: tech.regionBoundaries.
package boundary

import (
	"encoding/json"
	"fmt"
)

// Scale is how many stored units make one degree.
const Scale = 1000

// File is the boundary file as committed.
type File struct {
	// Source names where the land boundaries came from.
	Source string `json:"source"`
	// Land regions are tested first, in this order.
	Land []Land `json:"land"`
	// Sea regions are tested only where no land region contains the point.
	Sea []Sea `json:"sea"`
}

// Land is one land region's polygons.
type Land struct {
	// Region is the region's URL slug, what the Here button selects.
	Region string `json:"region"`
	// ID is the region's GADM id, for checking coverage.
	ID string `json:"id"`
	// Rings are the region's polygon rings, outer and inner alike, each as
	// [lon0, lat0, dlon1, dlat1, ...] in thousandths of a degree. A point is
	// inside when it is inside an odd number of them.
	Rings [][]int32 `json:"rings"`
}

// Sea is one sea region's coordinate box, in degrees.
type Sea struct {
	Region string `json:"region"`
	ID     string `json:"id"`
	// Box is [minLat, maxLat, minLon, maxLon].
	Box [4]float64 `json:"box"`
}

// Parse reads a boundary file.
func Parse(b []byte) (*File, error) {
	var f File
	if err := json.Unmarshal(b, &f); err != nil {
		return nil, fmt.Errorf("boundary: %w", err)
	}
	return &f, nil
}

// Points decodes a ring into degrees, as [lon, lat] pairs.
func Points(ring []int32) [][2]float64 {
	out := make([][2]float64, 0, len(ring)/2)
	var x, y int32
	for i := 0; i+1 < len(ring); i += 2 {
		x += ring[i]
		y += ring[i+1]
		out = append(out, [2]float64{float64(x) / Scale, float64(y) / Scale})
	}
	return out
}

// Lookup returns the slug of the region containing a point, or "" if none
// does: land regions first, then sea boxes.
func (f *File) Lookup(lat, lon float64) string {
	for _, l := range f.Land {
		inside := false
		for _, r := range l.Rings {
			if inRing(Points(r), lat, lon) {
				inside = !inside
			}
		}
		if inside {
			return l.Region
		}
	}
	for _, s := range f.Sea {
		if lat >= s.Box[0] && lat <= s.Box[1] && lon >= s.Box[2] && lon <= s.Box[3] {
			return s.Region
		}
	}
	return ""
}

// inRing is the even-odd ray test.
func inRing(pts [][2]float64, lat, lon float64) bool {
	in := false
	for i, j := 0, len(pts)-1; i < len(pts); j, i = i, i+1 {
		xi, yi := pts[i][0], pts[i][1]
		xj, yj := pts[j][0], pts[j][1]
		if (yi > lat) != (yj > lat) && lon < (xj-xi)*(lat-yi)/(yj-yi)+xi {
			in = !in
		}
	}
	return in
}
