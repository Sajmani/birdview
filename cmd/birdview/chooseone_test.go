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
	"regexp"
	"strings"
	"testing"
)

// TestChooseOneSettings asserts product.chooseOneSettings: each setting that takes one choice draws
// every choice, All first where it has one, in one joined row, with exactly the chosen link marked
// aria-current, and no words saying to choose one; a setting that takes one or more draws no row.
//
// Verifies: product.chooseOneSettings. Criterion: app.chooseOneSettings.
func TestChooseOneSettings(t *testing.T) {
	s := newTestServer(t)
	body := mustGet(t, s, "/?by=length&freq=1000&units=metric")
	pop := func(label string) string {
		t.Helper()
		i := strings.Index(body, `data-ctl="`+label+`"`)
		if i < 0 {
			t.Fatalf("no %s setting", label)
		}
		p := body[i:]
		return p[:strings.Index(p, "</details>")]
	}
	link := regexp.MustCompile(`<a class="btn[^"]*"[^>]*>`)
	for label, chosen := range map[string]string{
		"Birds": "All", "Order by": "Length", "Reporting frequency": "1 in 1,000", "Units": "Metric", "Names": "North American",
	} {
		p := pop(label)
		if strings.Contains(p, "Choose one") {
			t.Errorf("%s still says Choose one, which the owner removed", label)
		}
		seg := regexp.MustCompile(`(?s)<div class="seg" role="group"[^>]*>(.*?)</div>`).FindStringSubmatch(p)
		if seg == nil {
			t.Errorf("%s draws no joined row", label)
			continue
		}
		if all, in := len(link.FindAllString(p, -1)), len(link.FindAllString(seg[1], -1)); all != in || in < 2 {
			t.Errorf("%s has %d choices, %d of them in its joined row; want every one in it", label, all, in)
		}
		cur := regexp.MustCompile(`<a [^>]*aria-current="true"[^>]*>([^<]*)`).FindAllStringSubmatch(seg[1], -1)
		if len(cur) != 1 || !strings.HasPrefix(cur[0][1], chosen) {
			t.Errorf("%s marks %v as current; want exactly %q", label, cur, chosen)
		}
	}
	if !strings.Contains(regexp.MustCompile(`(?s)<div class="seg"[^>]*>.*?</div>`).FindString(pop("Birds")), ">All</a>") {
		t.Error("Birds' All is not in its joined row")
	}
	for _, label := range []string{"Month", "Region", "Family", "Habitat", "Diet"} {
		if p := pop(label); strings.Contains(p, `class="seg"`) {
			t.Errorf("%s takes one or more choices but is drawn as choose one", label)
		}
	}
}

// TestActionsColoredApart asserts product.nextView's rule for Here and Now: they act rather than
// choose, so they carry the act class that colors them apart, as Reset is.
//
// Verifies: product.nextView. Criterion: app.actionsColoredApart.
func TestActionsColoredApart(t *testing.T) {
	body := mustGet(t, newTestServer(t), "/")
	if !regexp.MustCompile(`<a class="btn act" href="[^"]*month=now[^"]*"[^>]*data-now>Now</a>`).MatchString(body) {
		t.Error("Now is not colored apart as an action")
	}
	if !regexp.MustCompile(`<button type="button" class="btn here act"`).MatchString(body) {
		t.Error("Here is not colored apart as an action")
	}
	if n := strings.Count(body, " act\""); n != 2 {
		t.Errorf("%d controls are colored as actions; want exactly Here and Now", n)
	}
}
