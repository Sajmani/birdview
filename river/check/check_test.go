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
	"sort"
	"strconv"
	"strings"
	"testing"
)

// TestBundle runs River's checker over BirdView's own bundle, so it runs with
// every other test and cannot be skipped (River's decisions.checksInRepoLanguage).
//
// Verifies: the structural rules in river/README.md.
func TestBundle(t *testing.T) {
	r := Run("..")
	for _, p := range r.Problems {
		t.Errorf("[%s] %s: %s", p.Code, p.Path, p.Message)
	}
	if !r.Stopped {
		t.Logf("%d requirement(s), %d goal(s), %d gradient(s), %d citing no goal, %d goal(s) uncited",
			r.Requirements, r.Goals, r.Gradients, r.Unjustified, r.Unserved)
	}
	for _, n := range r.Notes {
		t.Logf("note [%s] %s: %s", n.Code, n.Path, n.Message)
	}
}

// TestCases holds this translation to River's shared cases: each must report
// the same exit status, counts, problems, and notes, by code and path, as River's
// Python reference does. The cases are copied from River at the pin check.go
// records.
func TestCases(t *testing.T) {
	cases, err := filepath.Glob(filepath.Join("testdata", "*", "want.txt"))
	if err != nil || len(cases) < 20 {
		t.Fatalf("found %d cases (err %v); testdata/ has not been copied from River", len(cases), err)
	}
	for _, wantFile := range cases {
		dir := filepath.Dir(wantFile)
		t.Run(filepath.Base(dir), func(t *testing.T) {
			want := readWant(t, wantFile)
			r := Run(filepath.Join(dir, "bundle"))
			got := map[string]string{"exit": "0"}
			if len(r.Problems) > 0 {
				got["exit"] = "1"
			}
			if !r.Stopped {
				for k, v := range map[string]int{"requirements": r.Requirements, "goals": r.Goals,
					"gradients": r.Gradients, "unjustified": r.Unjustified, "unserved": r.Unserved} {
					got[k] = strconv.Itoa(v)
				}
			}
			var problems []string
			for _, p := range r.Problems {
				problems = append(problems, p.Code+" "+p.Path)
			}
			sort.Strings(problems)
			got["problems"] = strings.Join(problems, "\n")
			var notes []string
			for _, n := range r.Notes {
				notes = append(notes, n.Code+" "+n.Path)
			}
			sort.Strings(notes)
			got["notes"] = strings.Join(notes, "\n")
			for k, w := range want {
				if got[k] != w {
					t.Errorf("%s: want %q, got %q", k, w, got[k])
				}
			}
		})
	}
}

func readWant(t *testing.T, path string) map[string]string {
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]string{}
	var problems, notes []string
	for _, line := range strings.Split(string(b), "\n") {
		if strings.TrimSpace(line) == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, _ := strings.Cut(line, " ")
		if key == "problem" {
			problems = append(problems, value)
		} else if key == "note" {
			notes = append(notes, value)
		} else {
			want[key] = value
		}
	}
	sort.Strings(problems)
	want["problems"] = strings.Join(problems, "\n")
	// A case that lists no notes does not check them.
	if notes != nil {
		sort.Strings(notes)
		want["notes"] = strings.Join(notes, "\n")
	}
	return want
}
