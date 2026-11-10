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
	"sort"
	"strings"
	"testing"
)

// TestEveryLocalRequirementCitesAGoal is BirdView's own rule, stricter than
// River's checker, which counts a requirement citing no goal and a goal
// nothing cites rather than failing them (River's goalsMayPrecedeRequirements
// and requirementsMayPrecedeGoals). BirdView holds itself to both: every
// active product and tech requirement cites a declared goal, and every
// declared goal is cited by at least one active requirement.
//
// A transcribed source's rules are exempt -- they are an imported source's,
// not BirdView's to justify -- and so is a withdrawn requirement.
//
// Verifies: tech.bundleIsNormative. Criterion: repo.everyRequirementCitesGoal.
func TestEveryLocalRequirementCitesAGoal(t *testing.T) {
	goalFiles, _ := filepath.Glob(filepath.Join("..", "goals", "*.md"))
	if len(goalFiles) == 0 {
		t.Fatal("no goal files found")
	}
	h2 := regexp.MustCompile(`(?m)^## ([a-z][A-Za-z0-9]*)[ \t]*$`)
	withdrawn := regexp.MustCompile(`(?mi)^(?:\*\*)?Status:[ \t]*(?:\*\*)?withdrawn\b`)
	goalsField := regexp.MustCompile(`(?m)^Goals:[ \t]*(.+)$`)
	linkTarget := regexp.MustCompile(`\]\([^)]*\)`)
	goalRef := regexp.MustCompile(`\bgoals\.([a-z][A-Za-z0-9]*)\b`)

	declared := map[string]bool{}
	served := map[string]int{}
	for _, gf := range goalFiles {
		txt := read(t, gf)
		for _, m := range h2.FindAllStringSubmatch(txt, -1) {
			if !withdrawn.MatchString(section(txt, m[1])) {
				declared[m[1]] = true
			}
		}
	}
	if len(declared) == 0 {
		t.Fatal("no active goals parsed")
	}

	localFiles, _ := filepath.Glob(filepath.Join("..", "requirements", "local", "*", "*.md"))
	activeReqs := 0
	for _, lf := range localFiles {
		txt := read(t, lf)
		rel, _ := filepath.Rel("..", lf)
		rel = filepath.ToSlash(rel)
		for _, m := range h2.FindAllStringSubmatch(txt, -1) {
			ident := m[1]
			sec := section(txt, ident)
			if withdrawn.MatchString(sec) {
				continue
			}
			activeReqs++
			gm := goalsField.FindStringSubmatch(sec)
			if gm == nil {
				t.Errorf("%s: %s cites no goal; every active BirdView requirement must", rel, ident)
				continue
			}
			refs := goalRef.FindAllStringSubmatch(linkTarget.ReplaceAllString(gm[1], "]"), -1)
			if len(refs) == 0 {
				t.Errorf("%s: %s has Goals: field naming no goals.* identifier", rel, ident)
			}
			for _, r := range refs {
				if !declared[r[1]] {
					t.Errorf("%s: %s cites undeclared goal goals.%s", rel, ident, r[1])
				} else {
					served[r[1]]++
				}
			}
		}
	}
	if activeReqs == 0 {
		t.Fatal("no active local requirements parsed")
	}
	var unserved []string
	for g := range declared {
		if served[g] == 0 {
			unserved = append(unserved, "goals."+g)
		}
	}
	sort.Strings(unserved)
	if len(unserved) > 0 {
		t.Errorf("%d goal(s) are cited by no active requirement (%v); BirdView requires every goal to be served",
			len(unserved), unserved)
	}
}

// section returns the text of one item: from its heading to the next.
func section(text, ident string) string {
	norm := "\n" + text
	start := strings.Index(norm, "\n## "+ident+"\n")
	if start < 0 {
		return ""
	}
	rest := norm[start+1:]
	if next := strings.Index(rest[1:], "\n## "); next >= 0 {
		return rest[:next+1]
	}
	return rest
}

// TestActiveTodoHasIssue asserts tech.activeWorkHasIssue: every todo item with Status: active has an
// Issue: field linking one GitHub issue of this project. It reads the file, not GitHub, so it runs
// offline; that the issue links back is checked when the item is marked active.
//
// Verifies: tech.activeWorkHasIssue. Criterion: repo.activeTodoHasIssue.
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
