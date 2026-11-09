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

package trait

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

// Review is one species' row of data/habitat-review.tsv (decisions.habitatReviewedPerSpecies): AVONET's habitat as
// it stood when reviewed, the review's status, and the habitats BirdView adds.
type Review struct {
	Name, AVONET   string
	Status         string // pending, proposed, or reviewed
	Reviewer, Date string
	Additions      []Addition
}

// Addition is one habitat BirdView adds, with its reason.
type Addition struct {
	Habitat, Why string
}

var reviewStatuses = map[string]bool{"pending": true, "proposed": true, "reviewed": true}

// ParseHabitatReview reads the review table, keyed by scientific name. It refuses a row it cannot
// trust: an unknown status, a reviewed row with no reviewer or date, an addition outside AVONET's
// habitat vocabulary, or one without a reason.
func ParseHabitatReview(r io.Reader) (map[string]Review, error) {
	known := map[string]bool{}
	for _, t := range HabitatTerms {
		known[t.Value] = true
	}
	out := map[string]Review{}
	s := bufio.NewScanner(r)
	for n := 1; s.Scan(); n++ {
		line := s.Text()
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		f := strings.Split(line, "\t")
		if len(f) != 7 {
			return nil, fmt.Errorf("habitat review line %d: %d fields, want 7", n, len(f))
		}
		rv := Review{Name: f[1], AVONET: f[2], Status: f[3], Reviewer: f[4], Date: f[5]}
		if !reviewStatuses[rv.Status] {
			return nil, fmt.Errorf("habitat review line %d (%s): status %q", n, f[0], rv.Status)
		}
		if rv.Status == "reviewed" && (rv.Reviewer == "" || rv.Date == "") {
			return nil, fmt.Errorf("habitat review line %d (%s): reviewed, but by whom and when?", n, f[0])
		}
		if f[6] != "" {
			for _, part := range strings.Split(f[6], " | ") {
				h, why, ok := strings.Cut(part, ": ")
				if !ok || strings.TrimSpace(why) == "" {
					return nil, fmt.Errorf("habitat review line %d (%s): %q gives no reason", n, f[0], part)
				}
				if !known[h] {
					return nil, fmt.Errorf("habitat review line %d (%s): %q is not one of AVONET's habitats", n, f[0], h)
				}
				rv.Additions = append(rv.Additions, Addition{Habitat: h, Why: why})
			}
		}
		if _, dup := out[f[0]]; dup {
			return nil, fmt.Errorf("habitat review line %d: %s has two rows", n, f[0])
		}
		out[f[0]] = rv
	}
	return out, s.Err()
}
