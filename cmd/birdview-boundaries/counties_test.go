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
	"io"
	"io/fs"
	"math"
	"testing"

	"birdview/data"
	"birdview/internal/boundary"
)

// forwardLambert projects (lon, lat) degrees to (x, y) meters in
// NAD83_Statistics_Canada_Lambert (EPSG:3347), for testing lambertToLonLat.
func forwardLambert(lonDeg, latDeg float64) (x, y float64) {
	phi := latDeg * math.Pi / 180.0
	lam := lonDeg * math.Pi / 180.0
	es := lccE * math.Sin(phi)
	t := math.Tan(math.Pi/4-phi/2) / math.Pow((1-es)/(1+es), lccE/2)
	rho := lccA * lccF * math.Pow(t, lccN)
	theta := lccN * (lam - lccLam0)
	return lccFE + rho*math.Sin(theta), lccFN + lccRho0 - rho*math.Cos(theta)
}

func TestLambertInverseKnownCities(t *testing.T) {
	cities := []struct {
		name     string
		lon, lat float64
	}{
		{"Toronto", -79.3832, 43.6532},
		{"Montréal", -73.5673, 45.5017},
		{"Vancouver", -123.1207, 49.2827},
		{"Halifax", -63.5752, 44.6488},
		{"Ottawa", -75.6972, 45.4215},
		{"Whitehorse", -135.0568, 60.7212},
		{"Iqaluit", -68.5170, 63.7467},
	}
	for _, c := range cities {
		x, y := forwardLambert(c.lon, c.lat)
		gotLon, gotLat := lambertToLonLat(x, y)
		if math.Abs(gotLon-c.lon) > 1e-9 || math.Abs(gotLat-c.lat) > 1e-9 {
			t.Errorf("%s: lambertToLonLat(%.2f, %.2f) = (%.10f, %.10f), want (%.10f, %.10f)",
				c.name, x, y, gotLon, gotLat, c.lon, c.lat)
		}
	}
}

// TestCountyBoundariesCoverageAndKnownPoints verifies every committed per-state
// county boundary file: all 64 US states/DC and Canadian provinces/territories
// are present, each fits under the 200 KB compressed budget, every non-exempt
// county in data/county-presence/ has rings, and known US and Canadian cities
// resolve to their exact GADM county or census-division IDs.
//
// Verifies: app.herePinpointsCounty.
func TestCountyBoundariesCoverageAndKnownPoints(t *testing.T) {
	presence, err := readCountyPresenceHeaders("../../data/county-presence")
	if err != nil {
		t.Fatal(err)
	}
	if len(presence) != 64 {
		t.Fatalf("data/county-presence has %d states/provinces, want 64", len(presence))
	}

	const budget = 200 << 10
	parsed := make(map[string]*boundary.File, len(presence))
	var totalGZ int
	for st, counties := range presence {
		gzBytes, err := fs.ReadFile(data.CountyBoundaries, "county-boundaries/"+st+".json.gz")
		if err != nil {
			t.Errorf("missing county boundaries for %s: %v", st, err)
			continue
		}
		totalGZ += len(gzBytes)
		if len(gzBytes) > budget {
			t.Errorf("%s county boundaries are %d bytes compressed, over the %d-byte budget", st, len(gzBytes), budget)
		}
		gr, err := gzip.NewReader(bytes.NewReader(gzBytes))
		if err != nil {
			t.Errorf("%s: gzip header: %v", st, err)
			continue
		}
		raw, err := io.ReadAll(gr)
		gr.Close()
		if err != nil {
			t.Errorf("%s: gunzip: %v", st, err)
			continue
		}
		f, err := boundary.Parse(raw)
		if err != nil {
			t.Errorf("%s: parse: %v", st, err)
			continue
		}
		parsed[st] = f

		want := make(map[string]string, len(counties))
		for _, c := range counties {
			want[c.id] = c.name
		}
		have := make(map[string]bool, len(f.Land))
		for _, l := range f.Land {
			if _, ok := want[l.ID]; !ok {
				t.Errorf("%s: boundary %s is not in data/county-presence/%s.tsv.gz", st, l.ID, st)
			}
			if l.Region != l.ID {
				t.Errorf("%s: boundary %s has region %q, want %q", st, l.ID, l.Region, l.ID)
			}
			if len(l.Rings) == 0 {
				t.Errorf("%s: boundary %s has no rings", st, l.ID)
			}
			have[l.ID] = true
		}
		for _, c := range counties {
			if _, exempt := countyBoundaryExempt[c.id]; exempt {
				continue
			}
			if !have[c.id] {
				t.Errorf("%s: county %s (%s) has no boundary", st, c.id, c.name)
			}
		}
	}

	// Stage 1 -> stage 2 resolution on known US and Canadian cities.
	stage1 := committed(t)
	known := []struct {
		place      string
		lat, lon   float64
		wantState  string
		wantCounty string
	}{
		{"Brooklyn/Kings NY", 40.6501, -73.9496, "USA.33_1", "USA.33.24_1"},
		{"Los Angeles CA", 34.0522, -118.2437, "USA.5_1", "USA.5.19_1"},
		{"San Antonio/Bexar TX", 29.4241, -98.4936, "USA.44_1", "USA.44.15_1"},
		{"Chicago/Cook IL", 41.8781, -87.6298, "USA.14_1", "USA.14.16_1"},
		{"Toronto ON", 43.6532, -79.3832, "CAN.9_1", "CAN.9.50_1"},
		{"Montréal QC", 45.5017, -73.5673, "CAN.11_1", "CAN.11.20_1"},
		{"Vancouver BC", 49.2827, -123.1207, "CAN.2_1", "CAN.2.14_1"},
		{"Halifax NS", 44.7500, -63.6500, "CAN.7_1", "CAN.7.8_1"},
		// Inside NY's Natural Earth polygon on Lake Ontario, outside Census land polygons.
		{"Lake Ontario NY (state fallback)", 43.50, -77.50, "USA.33_1", ""},
	}
	for _, k := range known {
		if gotState := stage1.Lookup(k.lat, k.lon); gotState != k.wantState {
			t.Errorf("%s (%.4f, %.4f): stage 1 = %q, want %q", k.place, k.lat, k.lon, gotState, k.wantState)
			continue
		}
		f := parsed[k.wantState]
		if f == nil {
			continue
		}
		if gotCounty := f.Lookup(k.lat, k.lon); gotCounty != k.wantCounty {
			t.Errorf("%s (%.4f, %.4f): stage 2 = %q, want %q", k.place, k.lat, k.lon, gotCounty, k.wantCounty)
		}
	}
	t.Logf("%d states/provinces verified, %d bytes total compressed", len(parsed), totalGZ)
}
