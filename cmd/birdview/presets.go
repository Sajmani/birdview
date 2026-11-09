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
	"net/url"
	"strings"
)

// birdPreset is one choice of the Birds setting: a kind of bird, as the Family, Habitat, and Diet
// settings it fills in (product.birdPresets, decisions.birdPresets).
type birdPreset struct {
	Label                      string
	Families, Habitats, Niches []string
}

// birdPresets are the Birds setting's choices, in the order it offers them.
var birdPresets = []birdPreset{
	{Label: "Shorebirds",
		Families: []string{"Recurvirostridae", "Haematopodidae", "Charadriidae", "Jacanidae", "Scolopacidae"},
		Habitats: []string{"Coastal", "Wetland"}},
	{Label: "Raptors",
		Families: []string{"Accipitridae", "Pandionidae", "Falconidae", "Cathartidae"}},
	{Label: "Wading birds",
		Families: []string{"Ardeidae", "Threskiornithidae", "Ciconiidae", "Phoenicopteridae", "Aramidae", "Gruidae"}},
}

// presetKeys are the settings a choice fills in.
var presetKeys = []string{"family", "habitat", "niche"}

// query is the choice's settings as address parameters.
func (p birdPreset) query() url.Values {
	q := url.Values{}
	for _, f := range p.Families {
		q.Add("family", f)
	}
	for _, h := range p.Habitats {
		q.Add("habitat", h)
	}
	for _, n := range p.Niches {
		q.Add("niche", n)
	}
	return q
}

// matches reports whether the settings are exactly the choice's: the same families, habitats, and
// diets, none missing and none added.
func (p birdPreset) matches(opt options) bool {
	same := func(has bool, sel map[string]bool, want []string) bool {
		if len(want) == 0 {
			return !has
		}
		if !has || len(sel) != len(want) {
			return false
		}
		for _, v := range want {
			if !sel[v] {
				return false
			}
		}
		return true
	}
	return same(opt.HasSelection, opt.Families, p.Families) && same(opt.HasHabitats, opt.Habitats, p.Habitats) &&
		same(opt.HasNiches, opt.Niches, p.Niches)
}

// presetOf is the choice the settings are exactly, or nil.
func presetOf(opt options) *birdPreset {
	for i := range birdPresets {
		if birdPresets[i].matches(opt) {
			return &birdPresets[i]
		}
	}
	return nil
}

// birdsControl is the Birds setting: All, or a kind of bird, each filling in Family, Habitat, and
// Diet and keeping every other setting (product.birdPresets).
func birdsControl(q url.Values, opt options) nextControl {
	cleared := cloneValues(q)
	for _, k := range presetKeys {
		cleared.Del(k)
	}
	c := nextControl{Label: "Birds", Value: "all", AllURL: nextURL(cleared),
		AllOn: !opt.HasSelection && !opt.HasHabitats && !opt.HasNiches}
	cur := presetOf(opt)
	if !c.AllOn {
		c.Value = "custom"
		if cur != nil {
			c.Value = strings.ToLower(cur.Label)
		}
	}
	var ch []nextChoice
	for i := range birdPresets {
		p := &birdPresets[i]
		to := cloneValues(cleared)
		if p != cur {
			for k, vs := range p.query() {
				to[k] = vs
			}
		}
		ch = append(ch, nextChoice{Label: p.Label, Title: presetTitle(p), On: p == cur, URL: nextURL(to)})
	}
	c.Groups = []nextChoiceGroup{{Choices: ch}}
	return c
}

// presetTitle says what a choice fills in, for its tooltip.
func presetTitle(p *birdPreset) string {
	t := "Family: " + strings.Join(p.Families, ", ")
	if len(p.Habitats) > 0 {
		t += " · Habitat: " + nextList(habitatNames(sliceSet(p.Habitats)))
	}
	return t
}

func sliceSet(vs []string) map[string]bool {
	m := map[string]bool{}
	for _, v := range vs {
		m[v] = true
	}
	return m
}
