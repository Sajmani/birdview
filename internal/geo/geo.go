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

// Package geo answers two questions of a point against coastlines: is it on land, and how far is it
// from the nearest shore (tech.seaDistanceTable).
package geo

import (
	"math"
	"sort"
	"sync"
)

type point struct{ x, y float64 } // longitude, latitude
type edge struct{ a, b point }

// bucket is the size, in degrees, of the grid edges are indexed by.
const bucket = 0.25

type cell struct{ i, j int }

// Shore holds every edge of the land polygons that reach west of a longitude. Its answers hold for
// points west of that longitude: a polygon lying wholly east of it is left out, and a ray from such
// a point crosses a left-out polygon an even number of times, so the land test is unaffected.
type Shore struct {
	edges []edge
	index map[cell][]int

	mu    sync.Mutex
	lines map[float64][]float64 // latitude -> sorted longitudes where edges cross it
}

// NewShore indexes the polygons (each a list of rings of [lon, lat] points) that reach west of the
// longitude west.
func NewShore(polygons [][][][2]float64, west float64) *Shore {
	s := &Shore{index: map[cell][]int{}, lines: map[float64][]float64{}}
	for _, poly := range polygons {
		minX := math.Inf(1)
		for _, r := range poly {
			for _, p := range r {
				minX = math.Min(minX, p[0])
			}
		}
		if minX > west {
			continue
		}
		for _, r := range poly {
			for k := 0; k+1 < len(r); k++ {
				e := edge{point{r[k][0], r[k][1]}, point{r[k+1][0], r[k+1][1]}}
				n := len(s.edges)
				s.edges = append(s.edges, e)
				i0, i1 := floorDiv(math.Min(e.a.x, e.b.x)), floorDiv(math.Max(e.a.x, e.b.x))
				j0, j1 := floorDiv(math.Min(e.a.y, e.b.y)), floorDiv(math.Max(e.a.y, e.b.y))
				for i := i0; i <= i1; i++ {
					for j := j0; j <= j1; j++ {
						s.index[cell{i, j}] = append(s.index[cell{i, j}], n)
					}
				}
			}
		}
	}
	return s
}

func floorDiv(v float64) int { return int(math.Floor(v / bucket)) }

// OnLand reports whether the point lies inside land, by the even-odd rule along its latitude: an odd
// number of edge crossings east of it means land. A lake, drawn as a hole, counts as water.
func (s *Shore) OnLand(lat, lon float64) bool {
	xs := s.crossings(lat)
	return (len(xs)-sort.SearchFloat64s(xs, lon))%2 == 1
}

func (s *Shore) crossings(lat float64) []float64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	if xs, ok := s.lines[lat]; ok {
		return xs
	}
	j := floorDiv(lat)
	seen := map[int]bool{}
	var xs []float64
	for c, ids := range s.index {
		if c.j != j {
			continue
		}
		for _, n := range ids {
			if seen[n] {
				continue
			}
			seen[n] = true
			e := s.edges[n]
			if (e.a.y > lat) != (e.b.y > lat) {
				xs = append(xs, e.a.x+(lat-e.a.y)*(e.b.x-e.a.x)/(e.b.y-e.a.y))
			}
		}
	}
	sort.Float64s(xs)
	s.lines[lat] = xs
	return xs
}

// DistanceKm is the distance from the point to the nearest shore, in kilometers, measured on a local
// equirectangular projection (accurate to well under 1% at the distances banded), and capKm when no
// shore is that close.
func (s *Shore) DistanceKm(lat, lon, capKm float64) float64 {
	kx := 111.32 * math.Cos(lat*math.Pi/180)
	const ky = 110.57
	ri := int(math.Ceil(capKm/kx/bucket)) + 1
	rj := int(math.Ceil(capKm/ky/bucket)) + 1
	ci, cj := floorDiv(lon), floorDiv(lat)
	best := capKm
	for i := ci - ri; i <= ci+ri; i++ {
		for j := cj - rj; j <= cj+rj; j++ {
			for _, n := range s.index[cell{i, j}] {
				e := s.edges[n]
				ax, ay := (e.a.x-lon)*kx, (e.a.y-lat)*ky
				bx, by := (e.b.x-lon)*kx, (e.b.y-lat)*ky
				dx, dy := bx-ax, by-ay
				t := 0.0
				if l := dx*dx + dy*dy; l > 0 {
					t = math.Max(0, math.Min(1, -(ax*dx+ay*dy)/l))
				}
				if d := math.Hypot(ax+t*dx, ay+t*dy); d < best {
					best = d
				}
			}
		}
	}
	return best
}
