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

package month

import "testing"

// nyEffort is New York's birding by month, measured 2026-10-07 from eBird
// via GBIF: 62,467,434 records, and wildly uneven. May holds 20.9% of it
// and December 4.8%, which is the whole reason Present divides by the
// month's own total instead of by a twelfth.
var nyEffort = [Count]int64{
	3373241, 3060904, 4497655, 7808429, 13055694, 5871939,
	4560123, 4185318, 4872460, 4809992, 3310774, 2996905,
}

// pct turns a species' measured month profile -- the share of its OWN
// records falling in each month -- into counts against a total, which is
// the form Present takes.
func pct(total int64, shares [Count]float64) [Count]int64 {
	var out [Count]int64
	for i, s := range shares {
		out[i] = int64(float64(total) * s / 100)
	}
	return out
}

// TestPresentRecoversKnownSeasons asserts decisions.monthPresence's rule against ten species
// whose comings and goings are not in dispute.
//
// The measure is effort-corrected on purpose. A raw share of a species' own
// records would call every bird a May bird, because May is when people are
// out: an American Robin takes 22.1% of its New York records in May and
// 2.6% in December, and is present in both.
//
// The Pacific Golden-Plover is the case worth keeping. It leaves Hawaii in
// late April for Alaska and returns in August, and the rule recovers
// exactly that from record counts, with nothing about the species, its
// migration or its breeding range encoded anywhere.
//
// Watched fail: see acceptance.monthsFilterAndAbstain.
//
// Verifies: product.regionAndMonthFilter, decisions.monthPresence. Criterion: acceptance.monthsFilterAndAbstain.
func TestPresentRecoversKnownSeasons(t *testing.T) {
	// Measured shares of each species' own New York records, 2026-10-07.
	cases := []struct {
		name    string
		records int64
		shares  [Count]float64
		want    string // one character per month: + present, - absent
	}{
		{"Northern Cardinal, resident", 1819703,
			[Count]float64{6.6, 6.9, 9.1, 13.9, 18.0, 8.0, 6.7, 6.3, 7.0, 6.6, 5.2, 5.8},
			"++++++++++++"},
		{"American Robin, partial migrant", 1908205,
			[Count]float64{2.6, 2.8, 8.6, 16.8, 22.1, 11.7, 9.1, 6.1, 6.3, 7.0, 4.2, 2.6},
			"++++++++++++"},
		{"Dark-eyed Junco, winter visitor", 815420,
			[Count]float64{12.5, 12.1, 13.2, 15.4, 5.3, 3.1, 2.4, 1.2, 1.6, 9.7, 12.0, 11.6},
			"+++++++-++++"},
		{"Blackpoll Warbler, long-distance migrant", 96119,
			[Count]float64{0.0, 0.0, 0.0, 0.2, 57.7, 4.7, 1.1, 0.6, 20.7, 14.7, 0.3, 0.0},
			"----++--++--"},
		{"Bobolink, breeder", 91769,
			[Count]float64{0.0, 0.0, 0.0, 0.3, 39.1, 26.8, 10.4, 13.0, 9.2, 1.3, 0.0, 0.0},
			"----+++++---"},
		{"Semipalmated Sandpiper, passage shorebird", 73531,
			[Count]float64{0.0, 0.0, 0.0, 0.1, 16.4, 7.0, 15.6, 37.1, 20.1, 3.5, 0.2, 0.0},
			"----++++++--"},
		{"Snow Bunting, winter visitor", 36348,
			[Count]float64{25.9, 22.4, 8.6, 0.2, 0.0, 0.0, 0.0, 0.0, 0.0, 3.8, 22.2, 16.9},
			"+++------+++"},
	}
	for _, c := range cases {
		in, ok := Present(pct(c.records, c.shares), nyEffort)
		if !ok {
			t.Errorf("%s: abstained on %d records, well over the %d floor",
				c.name, c.records, MinRecords)
			continue
		}
		var got string
		for _, v := range in {
			if v {
				got += "+"
			} else {
				got += "-"
			}
		}
		if got != c.want {
			t.Errorf("%s:\n  got  %s\n  want %s\n  (Jan..Dec, + present)", c.name, got, c.want)
		}
	}
}

// TestPresentAbstainsBelowTheFloor asserts the other half of decisions.monthPresence: a
// species with too little evidence gets no verdict at all.
//
// Millerbird is the case. It holds 28 records in Hawaii, every one in
// March, April, August, September or October, and lives on Nihoa the year
// round — so those months record when anyone last got to Nihoa. Returning
// "absent in June" would state something about the bird out of a fact
// about the survey.
//
// The distinction that matters is between `ok == false` and twelve trues:
// the first says nothing is known, the second says the bird is there all
// year. A caller that conflated them would show the same page for a
// well-measured resident and a bird nobody has visited.
//
// Verifies: product.regionAndMonthFilter, decisions.monthPresence. Criterion: acceptance.monthsFilterAndAbstain.
func TestPresentAbstainsBelowTheFloor(t *testing.T) {
	hi := [Count]int64{270765, 278287, 290821, 208088, 155439, 157946,
		150425, 152932, 165468, 230652, 198060, 245694}

	// Millerbird: 28 records, in five months.
	var mb [Count]int64
	mb[2], mb[3], mb[7], mb[8], mb[9] = 1, 1, 3, 15, 8
	if _, ok := Present(mb, hi); ok {
		t.Errorf("judged a species on 28 records; below %d those months record "+
			"when anyone visited, not when the bird was there", MinRecords)
	}

	// One record over the floor is judged, so the floor is a floor and not
	// a shrug at anything small.
	// Comfortably over, so that dropping one month below still clears it.
	var over [Count]int64
	for i := range over {
		over[i] = MinRecords
	}
	in, ok := Present(over, hi)
	if !ok {
		t.Fatalf("abstained at %d records, which is over the %d floor",
			MinRecords*Count, MinRecords)
	}
	for i, v := range in {
		if !v {
			t.Errorf("a species spread evenly across the year reads absent in %s",
				Full[i])
		}
	}

	// A region nobody birded in a given month tells us nothing about any
	// species, so the species is not what goes missing.
	var noEffort [Count]int64
	copy(noEffort[:], hi[:])
	noEffort[5] = 0
	sp := over
	sp[5] = 0
	in, ok = Present(sp, noEffort)
	if !ok {
		t.Fatal("abstained on a species over the floor")
	}
	if !in[5] {
		t.Error("called a bird absent in a month with no birding at all; " +
			"the records that are missing are everyone's, not this bird's")
	}
}

// TestIndexRoundTrips asserts the URL form, which a shared link depends on.
func TestIndexRoundTrips(t *testing.T) {
	for i, a := range Abbrev {
		got, ok := Index(a)
		if !ok || got != i {
			t.Errorf("Index(%q) = %d, %v; want %d, true", a, got, ok, i)
		}
		// Case-insensitive, because a hand-typed URL will not match case.
		if got, ok := Index("jUn"); !ok || got != 5 {
			t.Errorf("Index is case-sensitive: %d, %v", got, ok)
		}
	}
	for _, bad := range []string{"", "Smarch", "13", "J"} {
		if _, ok := Index(bad); ok {
			t.Errorf("Index(%q) resolved; an unknown month must be dropped, "+
				"not guessed at", bad)
		}
	}
}
