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

package check

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// TestEveryLocalRequirementCitesAGoal is BirdView's own rule, stricter than
// River's checker, which counts a requirement citing no goal and a goal
// nothing cites rather than failing them (River's goalsMayPrecedeRequirements
// and requirementsMayPrecedeGoals). BirdView holds itself to both: every
// active product and tech requirement cites a goal, and every goal is cited.
//
// A transcribed source's rules are exempt -- they are an imported source's,
// not BirdView's to justify -- and so is a withdrawn requirement.
//
// Verifies: the structural rules in river/README.md. Criterion: acceptance.everyRequirementCitesGoal.
func TestEveryLocalRequirementCitesAGoal(t *testing.T) {
	r := Run("..")
	if r.Stopped {
		t.Fatalf("the checker stopped: %v", r.Problems)
	}
	if r.Unserved != 0 {
		t.Errorf("%d goal(s) are cited by no requirement; BirdView requires every goal to be served", r.Unserved)
	}
	withdrawn := regexp.MustCompile(`(?m)^Status: \*\*Withdrawn`)
	for _, ref := range r.UnjustifiedItems {
		path, ident, _ := strings.Cut(ref, "#")
		if !strings.HasPrefix(path, "requirements/local/product/") && !strings.HasPrefix(path, "requirements/local/tech/") {
			continue // an imported source's rule
		}
		b, err := os.ReadFile(filepath.Join("..", filepath.FromSlash(path)))
		if err != nil {
			t.Fatal(err)
		}
		if !withdrawn.MatchString(section(string(b), ident)) {
			t.Errorf("%s: %s cites no goal; every active BirdView requirement must", path, ident)
		}
	}
}

// section returns the text of one item: from its heading to the next.
func section(text, ident string) string {
	start := strings.Index(text, "\n## "+ident+"\n")
	if start < 0 {
		return ""
	}
	rest := text[start+1:]
	if next := strings.Index(rest[1:], "\n## "); next >= 0 {
		return rest[:next+1]
	}
	return rest
}

// TestActiveTodoHasIssue asserts tech.activeWorkHasIssue: every todo item with Status: active has an
// Issue: field linking one GitHub issue of this project. It reads the file, not GitHub, so it runs
// offline; that the issue links back is checked when the item is marked active.
//
// Verifies: tech.activeWorkHasIssue. Criterion: acceptance.activeTodoHasIssue.
func TestActiveTodoHasIssue(t *testing.T) {
	b, err := os.ReadFile(filepath.Join("..", "todo.md"))
	if err != nil {
		t.Fatal(err)
	}
	issue := regexp.MustCompile(`(?m)^Issue: \[#\d+\]\(https://github\.com/google/birdview/issues/\d+\)\s*$`)
	for _, item := range strings.Split(string(b), "\n## ")[1:] {
		name, body, _ := strings.Cut(item, "\n")
		if !regexp.MustCompile(`(?m)^Status: active\b`).MatchString(body) {
			continue
		}
		if !issue.MatchString(body) {
			t.Errorf("todo.%s is active but has no Issue: line linking its GitHub issue, as [#N](https://github.com/google/birdview/issues/N)", name)
		}
	}
}
