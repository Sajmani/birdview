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

package boundary

// KnownPoints are places whose region is a matter of geography, not of what
// any lookup returns. Tampa is the trap: it lies inside the Gulf box, and land
// must win. The Go lookup is checked against them in cmd/birdview-boundaries,
// and the browser's in cmd/birdview, so the two are held to the same answers.
var KnownPoints = []struct {
	Place    string
	Lat, Lon float64
	Want     string
}{
	{"Sacramento", 38.5816, -121.4944, "CA"},
	{"Toronto", 43.6532, -79.3832, "ON"},
	{"Honolulu", 21.3069, -157.8583, "HI"},
	{"Anchorage", 61.2181, -149.9003, "AK"},
	{"the White House", 38.8977, -77.0365, "DC"},
	{"Tampa", 27.9506, -82.4572, "FL"},
	{"Sand Island, Midway", 28.2085, -177.3780, "MidwayAtoll"},
	{"the Gulf off Louisiana", 28.0, -90.0, "GulfOfMexico"},
	{"the open North Atlantic", 40.0, -60.0, "NorthAtlantic"},
	{"Mexico City", 19.4326, -99.1332, ""},
	{"London", 51.5074, -0.1278, ""},
}
