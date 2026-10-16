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

package region

import (
	"os"
	"testing"
)

// presenceFile is the committed table the app itself filters on. The zones
// are derived from it (decisions.regionsTwoWays), so the check reads the same file rather than
// a fixture: a fixture would let the shipped grouping and the shipped data
// drift apart, which is the one failure this criterion exists to catch.
const presenceFile = "../../data/region-presence.tsv"

// regionSpecies inverts the presence table into region -> species set.
func regionSpecies(t *testing.T) map[string]map[string]bool {
	t.Helper()
	f, err := os.Open(presenceFile)
	if err != nil {
		t.Fatalf("opening the presence table: %v", err)
	}
	defer f.Close()
	p, _, _, err := LoadPresence(f)
	if err != nil {
		t.Fatalf("reading the presence table: %v", err)
	}
	out := map[string]map[string]bool{}
	for sp, reported := range p {
		for _, r := range reported {
			if out[r.ID] == nil {
				out[r.ID] = map[string]bool{}
			}
			out[r.ID][sp] = true
		}
	}
	return out
}

// jaccard is the overlap measure the clustering used: shared species over
// species in either. Two regions with no records at all are 0, not 1.
func jaccard(a, b map[string]bool) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	shared := 0
	for s := range a {
		if b[s] {
			shared++
		}
	}
	return float64(shared) / float64(len(a)+len(b)-shared)
}

// meanWithin is the mean pairwise similarity inside a set of regions, and
// meanOutside the mean from those regions to every region not in the set.
func meanWithin(sp map[string]map[string]bool, mem []string) (float64, int) {
	var sum float64
	var n int
	for i := 0; i < len(mem); i++ {
		for j := i + 1; j < len(mem); j++ {
			sum += jaccard(sp[mem[i]], sp[mem[j]])
			n++
		}
	}
	if n == 0 {
		return 0, 0
	}
	return sum / float64(n), n
}

func meanOutside(sp map[string]map[string]bool, mem []string) float64 {
	in := map[string]bool{}
	for _, id := range mem {
		in[id] = true
	}
	var sum float64
	var n int
	for _, id := range mem {
		for other := range sp {
			if in[other] {
				continue
			}
			sum += jaccard(sp[id], sp[other])
			n++
		}
	}
	if n == 0 {
		return 0
	}
	return sum / float64(n)
}

// landZones lists a grouping's zones with their land members, dropping the
// marine ones: the ocean boxes are BirdView's own invention, so clustering
// them would be measuring a decision rather than a bird.
func landZones(grouping string) map[string][]string {
	out := map[string][]string{}
	for _, z := range ZonesOf(grouping) {
		if z == "Pelagic" {
			continue
		}
		for _, r := range InZone(grouping, z) {
			if r.Kind == Land {
				out[z] = append(out[z], r.ID)
			}
		}
	}
	return out
}

// meanTo is a region's mean similarity to a set of other regions, excluding
// itself.
func meanTo(sp map[string]map[string]bool, id string, to []string) (float64, bool) {
	var sum float64
	var n int
	for _, other := range to {
		if other == id {
			continue
		}
		sum += jaccard(sp[id], sp[other])
		n++
	}
	if n == 0 {
		return 0, false
	}
	return sum / float64(n), true
}

// TestEveryRegionResemblesItsOwnZone is the per-region form of the claim the
// "Common birds" grouping makes: a region must look more like its own
// zone-mates than like regions in general.
//
// **Per region, not per zone.** A per-zone average hides an intruder: moving
// Hawaii into the Great Lakes leaves that zone's average still beating its
// surroundings, and the first version of this check passed with Hawaii in
// Michigan's company. Each region has to earn its place individually.
//
// The bar is zero, not a number lifted from a run. "Resembles its zone-mates
// more than it resembles regions at large" is what being in a group *means*;
// any positive threshold would be a record of what this data happened to do
// (process.md).
//
// It deliberately does not re-run the clustering, which would compare the
// algorithm against itself and ratify whatever it produced.
//
// Verifies: product.regionsGroupedTwoWays.
func TestEveryRegionResemblesItsOwnZone(t *testing.T) {
	sp := regionSpecies(t)

	// Every id in the table must exist in the data, or a typo would silently
	// produce an empty zone that the rest of this check then skips.
	for _, id := range LandIDs() {
		if len(sp[id]) == 0 {
			t.Errorf("%s has no species in the presence table", id)
		}
	}

	everywhere := make([]string, 0, len(sp))
	for id := range sp {
		everywhere = append(everywhere, id)
	}

	for z, mem := range landZones(ByBirds) {
		for _, id := range mem {
			own, ok := meanTo(sp, id, mem)
			if !ok {
				continue // a zone of one: nothing to resemble
			}
			general, _ := meanTo(sp, id, everywhere)
			if own <= general {
				r, _ := Get(id)
				t.Errorf("%s is in zone %q but resembles it no more than it resembles "+
					"regions at large (%.3f vs %.3f) — it has not earned its place",
					r.Name, z, own, general)
			}
		}
	}
}

// worstOwnZoneRank is how far down a region's own zone sits when its zones
// are ranked by how much that region resembles them. 1 means every region is
// in the zone it most resembles; the number returned is the worst case.
func worstOwnZoneRank(sp map[string]map[string]bool, grouping string) (int, string) {
	zones := landZones(grouping)
	worst, who := 0, ""
	for z, mem := range zones {
		for _, id := range mem {
			own, ok := meanTo(sp, id, mem)
			if !ok {
				continue
			}
			rank := 1
			for z2, mem2 := range zones {
				if z2 == z {
					continue
				}
				if other, ok := meanTo(sp, id, mem2); ok && other > own {
					rank++
				}
			}
			if rank > worst {
				r, _ := Get(id)
				worst, who = rank, r.Name
			}
		}
	}
	return worst, who
}

// TestDerivedZonesBeatTheHandDrawnOnes is the comparison decisions.regionsTwoWays rests on. If
// the measured grouping were no better than the one somebody drew from
// memory, deriving it would buy nothing and the honest move would be to ship
// the hand-drawn one alone.
//
// Both measures are comparative, so neither carries a fitted threshold: the
// hand-drawn grouping is the baseline, and the derived one has to beat it.
//
// The rank measure is the one with teeth. Mean cohesion barely moves when a
// single region is misfiled, but a misfiled region drags its new zone-mates
// down everyone's ranking, so the worst rank jumps.
//
// **Neither measure demands rank 1 everywhere, because that is not true of
// this data.** Nine regions sit closer to a neighbouring zone than to their
// own, by margins as small as 0.001; reassigning them iteratively does not
// converge but cycles, five regions moving every sweep. The mainland is a
// smooth gradient and any cut is a choice of where to slice (decisions.regionsTwoWays).
//
// Verifies: product.regionsGroupedTwoWays.
func TestDerivedZonesBeatTheHandDrawnOnes(t *testing.T) {
	sp := regionSpecies(t)

	mean := func(grouping string) float64 {
		var sum float64
		var n int
		for _, mem := range landZones(grouping) {
			if len(mem) < 2 {
				continue
			}
			w, pairs := meanWithin(sp, mem)
			sum += w * float64(pairs)
			n += pairs
		}
		if n == 0 {
			return 0
		}
		return sum / float64(n)
	}

	if birds, geo := mean(ByBirds), mean(ByGeography); birds <= geo {
		t.Errorf("the derived grouping is no more cohesive than the hand-drawn one "+
			"(%.4f vs %.4f); decisions.regionsTwoWays claims it is, and if it is not then deriving "+
			"it bought nothing", birds, geo)
	}

	birdRank, birdWho := worstOwnZoneRank(sp, ByBirds)
	geoRank, geoWho := worstOwnZoneRank(sp, ByGeography)
	if birdRank >= geoRank {
		t.Errorf("the worst-placed region in the derived grouping ranks its own zone "+
			"%d (%s), no better than the hand-drawn grouping's %d (%s) — "+
			"a region is filed somewhere the records do not put it",
			birdRank, birdWho, geoRank, geoWho)
	}
}
