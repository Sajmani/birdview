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

package geo

import (
	"math"
	"testing"
)

// square is a unit square of land at the origin with a lake in its middle, as Natural Earth draws a
// lake: a second ring of the same polygon.
func square() [][][][2]float64 {
	return [][][][2]float64{{
		{{0, 0}, {1, 0}, {1, 1}, {0, 1}, {0, 0}},
		{{0.4, 0.4}, {0.6, 0.4}, {0.6, 0.6}, {0.4, 0.6}, {0.4, 0.4}},
	}}
}

// TestShoreLandAndDistance asserts the geometry tech.seaDistanceTable relies on: a point inside land
// is on land, a point in a lake or the sea is not, and the distance to the nearest shore is the
// local great-circle distance to the nearest edge, capped.
//
// Verifies: tech.seaDistanceTable. Criterion: dataset.seaDistanceTable.
func TestShoreLandAndDistance(t *testing.T) {
	s := NewShore(square(), 10)
	for _, c := range []struct {
		lat, lon float64
		land     bool
	}{
		{0.2, 0.2, true}, {0.5, 0.5, false}, {0.5, 1.5, false}, {-0.5, 0.5, false}, {0.95, 0.05, true},
	} {
		if got := s.OnLand(c.lat, c.lon); got != c.land {
			t.Errorf("OnLand(%g, %g) = %v, want %v", c.lat, c.lon, got, c.land)
		}
	}
	want := 0.1 * 111.32 * math.Cos(0.5*math.Pi/180)
	if d := s.DistanceKm(0.5, 1.1, 100); math.Abs(d-want) > 0.05 {
		t.Errorf("0.1° east of the shore is %.2f km from it, want %.2f", d, want)
	}
	if d := s.DistanceKm(0.5, 0.5, 100); math.Abs(d-0.1*111.32) > 0.5 {
		t.Errorf("the lake's center is %.2f km from its shore, want about 11", d)
	}
	if d := s.DistanceKm(0.5, 5, 100); d != 100 {
		t.Errorf("a point 4° out is %.1f km from shore; past the cap it must read the cap, 100", d)
	}
	if s := NewShore(square(), -1); s.OnLand(0.2, 0.2) || s.DistanceKm(0.5, 1.1, 100) != 100 {
		t.Error("a polygon wholly east of the west limit must be left out")
	}
}
