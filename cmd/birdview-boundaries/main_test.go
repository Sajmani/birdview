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
	"bytes"
	"compress/gzip"
	"testing"

	"birdview/data"
	"birdview/internal/boundary"
	"birdview/internal/region"
)

func committed(t *testing.T) *boundary.File {
	t.Helper()
	f, err := boundary.Parse(data.Boundaries)
	if err != nil {
		t.Fatal(err)
	}
	return f
}

// TestKnownPoints checks the committed boundary file against geography.
//
// Verifies: tech.regionBoundaries. Criterion: app.boundariesResolveKnownPoints.
func TestKnownPoints(t *testing.T) {
	f := committed(t)
	for _, p := range boundary.KnownPoints {
		if got := f.Lookup(p.Lat, p.Lon); got != p.Want {
			t.Errorf("%s (%.4f, %.4f) resolves to %q, want %q", p.Place, p.Lat, p.Lon, got, p.Want)
		}
	}
}

// TestCoverage checks that every land region has a boundary and every boundary
// a region, that the sea boxes are internal/region's own, and that the file
// stays within the size every index page pays to load it.
//
// Verifies: tech.regionBoundaries. Criterion: app.boundariesCoverEveryRegion.
func TestCoverage(t *testing.T) {
	f := committed(t)
	have := map[string]boundary.Land{}
	for _, l := range f.Land {
		r, ok := region.Get(l.ID)
		switch {
		case !ok || r.Kind != region.Land:
			t.Errorf("boundary %s (%s) is not a land region", l.ID, l.Region)
		case r.Slug() != l.Region:
			t.Errorf("boundary %s is labeled %q; the region's slug is %q", l.ID, l.Region, r.Slug())
		}
		if len(l.Rings) == 0 {
			t.Errorf("boundary %s has no rings", l.ID)
		}
		have[l.ID] = l
	}
	for _, id := range region.LandIDs() {
		if _, ok := have[id]; !ok {
			r, _ := region.Get(id)
			t.Errorf("land region %s (%s) has no boundary; Here could never select it", id, r.Name)
		}
	}
	ids := region.MarineIDs()
	if len(f.Sea) != len(ids) {
		t.Fatalf("%d sea boxes in the file, %d in internal/region", len(f.Sea), len(ids))
	}
	for i, s := range f.Sea {
		minLat, maxLat, minLon, maxLon, _ := region.MarineBox(ids[i])
		r, _ := region.Get(ids[i])
		if s.ID != ids[i] || s.Region != r.Slug() || s.Box != [4]float64{minLat, maxLat, minLon, maxLon} {
			t.Errorf("sea entry %d is %+v; internal/region has %s %q %v", i, s, ids[i], r.Slug(),
				[4]float64{minLat, maxLat, minLon, maxLon})
		}
	}
	var gz bytes.Buffer
	w := gzip.NewWriter(&gz) // the default level, as the server sends it
	w.Write(data.Boundaries)
	w.Close()
	const budget = 200 << 10
	if gz.Len() > budget {
		t.Errorf("boundary file is %d bytes compressed, over the %d-byte budget", gz.Len(), budget)
	}
	t.Logf("%d land regions, %d sea boxes, %d bytes, %d compressed", len(f.Land), len(f.Sea), len(data.Boundaries), gz.Len())
}
