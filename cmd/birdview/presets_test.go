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
	"testing"

	"birdview/internal/trait"
)

// TestBirdPresets asserts product.birdPresets: Birds is the first setting, defaulting to All; a
// choice fills in Family, Habitat, and Diet and keeps every other setting; the address names no
// choice; Birds names a choice whose settings match exactly, and the heading says so; a changed
// setting reads custom; choosing it again, or All, clears what it filled in.
//
// Verifies: product.birdPresets. Criterion: acceptance.birdPresets.
func TestBirdPresets(t *testing.T) {
	s := newTestServer(t)
	birds := func(target string) nextControl {
		t.Helper()
		p := nextAt(s, target)
		if len(p.Controls) == 0 || p.Controls[0].Label != "Birds" {
			t.Fatalf("%s: the first setting is not Birds", target)
		}
		return p.Controls[0]
	}
	choice := func(c nextControl, label string) nextChoice {
		t.Helper()
		for _, g := range c.Groups {
			for _, ch := range g.Choices {
				if ch.Label == label {
					return ch
				}
			}
		}
		t.Fatalf("Birds offers no %s", label)
		return nextChoice{}
	}
	if c := birds("/"); c.Value != "all" || !c.AllOn {
		t.Errorf("by default Birds reads %q (All marked %v); want all", c.Value, c.AllOn)
	}
	from := "/?region=USA.33_1&month=Aug&freq=10000&by=length&family=Corvidae&niche=Omnivore"
	sh := choice(birds(from), "Shorebirds")
	u, _ := url.Parse(string(sh.URL))
	q := u.Query()
	for k, want := range map[string]string{"region": "USA.33_1", "month": "Aug", "freq": "10000", "by": "length"} {
		if q.Get(k) != want {
			t.Errorf("Shorebirds from %s drops %s=%s: %s", from, k, want, sh.URL)
		}
	}
	if got := strings.Join(q["family"], ","); got != "Recurvirostridae,Haematopodidae,Charadriidae,Jacanidae,Scolopacidae" {
		t.Errorf("Shorebirds sets family to %s", got)
	}
	if got := strings.Join(q["habitat"], ","); got != "Coastal,Wetland" || q.Has("niche") || q.Has("preset") || q.Has("birds") {
		t.Errorf("Shorebirds sets %s; want habitat Coastal and Wetland, no niche, and no name for the choice", u.RawQuery)
	}
	c := birds(string(sh.URL))
	if c.Value != "shorebirds" || !choice(c, "Shorebirds").On || c.AllOn {
		t.Errorf("at Shorebirds' own address Birds reads %q, Shorebirds marked %v", c.Value, choice(c, "Shorebirds").On)
	}
	if p := nextAt(s, string(sh.URL)); !strings.HasPrefix(p.Title, "Showing shorebirds") || strings.Contains(p.Title, "Family:") {
		t.Errorf("Shorebirds is headed %q; want it named, without the settings it filled in", p.Title)
	}
	if again, _ := url.Parse(string(choice(c, "Shorebirds").URL)); again.Query().Has("family") || again.Query().Has("habitat") || again.Query().Get("region") != "USA.33_1" {
		t.Errorf("choosing Shorebirds again goes to %s; want Family and Habitat cleared, the rest kept", again)
	}
	if all, _ := url.Parse(string(c.AllURL)); all.Query().Has("family") || all.Query().Has("habitat") || all.Query().Get("month") != "Aug" {
		t.Errorf("All goes to %s; want Family and Habitat cleared, the rest kept", all)
	}
	if c := birds(string(sh.URL) + "&habitat=Forest"); c.Value != "custom" || choice(c, "Shorebirds").On {
		t.Errorf("with a habitat added Birds reads %q; want custom", c.Value)
	}
	wading := nextAt(s, string(choice(birds("/?show=all&month=all"), "Wading birds").URL))
	if !hasTile(wading, "greblu") || hasTile(wading, "blasti") {
		t.Error("Wading birds must draw the heron and not the stilt")
	}
	if shore := nextAt(s, "/?show=all&month=all&"+birdPresets[0].query().Encode()); !hasTile(shore, "blasti") || hasTile(shore, "greblu") {
		t.Error("Shorebirds must draw the stilt and not the heron")
	}
}

// TestEveryPresetDrawsBirds asserts decisions.birdPresets against the built dataset: every family a
// choice names is drawn, every habitat is in the vocabulary, and every choice draws birds, so a
// renamed family or a changed habitat cannot leave a choice empty.
//
// Verifies: product.birdPresets. Criterion: acceptance.birdPresets.
func TestEveryPresetDrawsBirds(t *testing.T) {
	ds := realDataset(t)
	fams := map[string]bool{}
	for _, sp := range ds.Species {
		fams[sp.Family] = true
	}
	s := newTestServerReal(t)
	for _, p := range birdPresets {
		for _, f := range p.Families {
			if !fams[f] {
				t.Errorf("%s names %s, which no drawn bird belongs to", p.Label, f)
			}
		}
		for _, h := range p.Habitats {
			if trait.LabelOf(trait.HabitatTerms, h) == h {
				t.Errorf("%s names the habitat %s, which is not in the vocabulary", p.Label, h)
			}
		}
		page := nextAt(s, "/?show=all&month=all&"+p.query().Encode())
		if len(page.Tiles) == 0 {
			t.Errorf("%s draws no birds", p.Label)
		}
		t.Logf("%s: %d birds", p.Label, len(page.Tiles))
	}
}
