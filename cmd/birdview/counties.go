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
	"fmt"
	"os"
	"sort"

	"birdview/internal/dataset"
	"birdview/internal/region"
)

// countyView is the county data the server holds apart from the species (decisions.countiesServedApart),
// joined to the dataset by scientific name. A nil view, which every request carries without
// -counties, has no county in it.
type countyView struct {
	c   *dataset.Counties
	idx map[string]int // scientific name to index in c.Species
	// listed counts, per species and state, the counties that list the species; perState counts
	// each state's counties. A bird's page reads them as "CA 12/58".
	listed   map[string]map[string]int
	perState map[string]int
}

func newCountyView(c *dataset.Counties) *countyView {
	v := &countyView{c: c, idx: make(map[string]int, len(c.Species)),
		listed: map[string]map[string]int{}, perState: map[string]int{}}
	for i, sci := range c.Species {
		v.idx[sci] = i
	}
	for k, ct := range c.Counties {
		v.perState[ct.State]++
		for _, e := range c.Entries[k] {
			if e.Below() {
				continue
			}
			sci := c.Species[e.Species]
			if v.listed[sci] == nil {
				v.listed[sci] = map[string]int{}
			}
			v.listed[sci][ct.State]++
		}
	}
	return v
}

// entry is sp's records in county id, if id is a county and sp has records there.
func (v *countyView) entry(sp *dataset.Species, id string) (dataset.CountyEntry, bool) {
	if v == nil {
		return dataset.CountyEntry{}, false
	}
	i, ok := v.idx[sp.ScientificName]
	if !ok {
		return dataset.CountyEntry{}, false
	}
	return v.c.Entry(id, i)
}

// county is county id, with its totals, if it is one.
func (v *countyView) county(id string) (dataset.County, bool) {
	if v == nil {
		return dataset.County{}, false
	}
	return v.c.County(id)
}

// countiesIn lists, for a bird's page, each state with counties and how many of them list sp out
// of how many it has, as "CA 12/58", by the state's code (decisions.countiesServedApart). Nil without
// county data.
func (v *countyView) countiesIn(sp *dataset.Species) map[string]string {
	if v == nil {
		return nil
	}
	out := map[string]string{}
	for st, n := range v.perState {
		out[st] = fmt.Sprintf("%d/%d", v.listed[sp.ScientificName][st], n)
	}
	return out
}

// register makes every county a region by its GADM code, under its state.
func (v *countyView) register() error {
	ids := make([]int, len(v.c.Counties))
	for i := range ids {
		ids[i] = i
	}
	sort.Slice(ids, func(a, b int) bool { return v.c.Counties[ids[a]].ID < v.c.Counties[ids[b]].ID })
	for _, i := range ids {
		ct := v.c.Counties[i]
		if err := region.AddCounty(ct.ID, ct.Name, ct.State); err != nil {
			return err
		}
	}
	return nil
}

// loadCounties reads the counties the ingest left beside the dataset.
func loadCounties(path string) (*dataset.Counties, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	c, err := dataset.ReadCounties(f)
	if err != nil {
		return nil, fmt.Errorf("decoding %s: %w", path, err)
	}
	return c, nil
}
