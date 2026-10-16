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
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"
	"unicode/utf8"
)

// TestReferencesResolve checks that every reference to a goal, requirement,
// criterion or decision resolves, across every tracked file outside
// third_party/:
//
//   - every qualified reference -- a goal, product, tech, decision or
//     criterion name, and a transcribed source's exported name -- names an
//     item its package defines;
//   - every anchored link into the bundle lands on a heading that exists,
//     an item's anchor being its identifier lowercased;
//   - a link whose text names an item lands on that item: the file that
//     defines it, at its anchor;
//   - no identifier in a superseded form survives anywhere: a letter code and
//     three digits, or a source name, a slash and an R number.
//
// It catches a name that was renamed, one invented in a decision record and
// never written down, and a link that lands nowhere.
//
// Verifies: process integrity. Criterion: acceptance.decisionReferencesExist.
func TestReferencesResolve(t *testing.T) {
	bundle, _ := filepath.Abs("..")
	repo := filepath.Dir(bundle)
	sources := filepath.Join(bundle, "requirements", "sources")

	lowerItem := regexp.MustCompile(`(?m)^#{2,3} ([a-z][A-Za-z]+)[ \t]*$`)
	upperItem := regexp.MustCompile(`(?m)^#{2,3} ([A-Z][A-Za-z]+)[ \t]*$`)
	glob := func(pattern string) []string {
		m, _ := filepath.Glob(filepath.Join(bundle, pattern))
		sort.Strings(m)
		return m
	}
	packages := map[string][]string{
		"goals":      glob("goals/*.md"),
		"product":    glob("requirements/sources/product/*.md"),
		"tech":       glob("requirements/sources/tech/*.md"),
		"decisions":  {filepath.Join(bundle, "requirements", "decisions.md")},
		"acceptance": {filepath.Join(bundle, "requirements", "acceptance.md")},
	}
	// A transcribed source is its own package, with exported names.
	exported := map[string]bool{}
	dirs, _ := os.ReadDir(sources)
	for _, d := range dirs {
		f := filepath.Join(sources, d.Name(), "requirements.md")
		if _, err := os.Stat(f); err == nil {
			packages[d.Name()] = []string{f}
			exported[d.Name()] = true
		}
	}
	defined := map[string]map[string]string{}
	for pkg, files := range packages {
		item := lowerItem
		if exported[pkg] {
			item = upperItem
		}
		defined[pkg] = map[string]string{}
		for _, f := range files {
			for _, m := range item.FindAllStringSubmatch(read(t, f), -1) {
				defined[pkg][m[1]] = f
			}
		}
	}
	if len(defined["product"]) == 0 || len(defined["decisions"]) == 0 {
		t.Fatal("no requirements or decisions parsed; the layout has changed and this check proves nothing")
	}

	heading := regexp.MustCompile(`(?m)^#{1,6} (.+?)[ \t]*$`)
	anchors := map[string]map[string]bool{}
	_ = filepath.WalkDir(bundle, func(p string, d os.DirEntry, err error) error {
		if err == nil && !d.IsDir() && strings.HasSuffix(p, ".md") {
			anchors[p] = map[string]bool{}
			for _, h := range heading.FindAllStringSubmatch(read(t, p), -1) {
				anchors[p][slug(h[1])] = true
			}
		}
		return nil
	})

	extensions := map[string]bool{"md": true, "py": true, "go": true, "svg": true, "json": true, "tsv": true, "html": true, "css": true}
	qualified := regexp.MustCompile(`\b(goals|product|tech|decisions|acceptance)\.([a-z][A-Za-z]+)\b`)
	var names []string
	for pkg := range exported {
		names = append(names, pkg)
	}
	sort.Strings(names)
	var exportedRef *regexp.Regexp
	if len(names) > 0 {
		exportedRef = regexp.MustCompile(`\b(` + strings.Join(names, "|") + `)\.([A-Z][A-Za-z]+)\b`)
	}
	legacy := regexp.MustCompile(`\b(?:G|P|T|AC|CR)-\d{3}\b|\b[a-z][a-z-]*/R\d+\b`)
	link := regexp.MustCompile(`\]\(([^)\s#]*)#([^)\s]+)\)`)
	citation := regexp.MustCompile("\\[`?([a-z]+)\\.([A-Za-z]+)`?\\]\\(([^)\\s]*)\\)")
	// Translated or copied from River, so their names are River's.
	fromRiver := map[string]bool{"river/README.md": true, "river/check/check.go": true, "river/check/check_test.go": true}

	out, err := exec.Command("git", "-C", repo, "ls-files").Output()
	if err != nil {
		t.Fatalf("listing tracked files: %v; this check reads exactly what git tracks", err)
	}
	refs := 0
	for _, rel := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if strings.HasPrefix(rel, "third_party/") || strings.HasPrefix(rel, "river/check/testdata/") {
			continue
		}
		path := filepath.Join(repo, filepath.FromSlash(rel))
		b, err := os.ReadFile(path)
		if err != nil || !utf8.Valid(b) {
			continue
		}
		text := string(b)
		if !fromRiver[rel] {
			for _, re := range []*regexp.Regexp{qualified, exportedRef} {
				if re == nil {
					continue
				}
				for _, m := range re.FindAllStringSubmatch(text, -1) {
					if extensions[m[2]] {
						continue // goals.md, refs.go: a file, not a reference
					}
					refs++
					if _, ok := defined[m[1]][m[2]]; !ok {
						t.Errorf("%s: %s.%s names nothing package %s defines", rel, m[1], m[2], m[1])
					}
				}
			}
			// The checks have to describe the form they reject.
			if !strings.HasPrefix(rel, "river/check/") {
				for _, m := range legacy.FindAllString(text, -1) {
					t.Errorf("%s: %s is a superseded identifier", rel, m)
				}
			}
		}
		if !strings.HasSuffix(rel, ".md") {
			continue
		}
		resolve := func(target string) string {
			if target == "" {
				return path
			}
			return filepath.Clean(filepath.Join(filepath.Dir(path), filepath.FromSlash(target)))
		}
		for _, m := range citation.FindAllStringSubmatch(text, -1) {
			pkg, name, url := m[1], m[2], m[3]
			want, ok := defined[pkg][name]
			if !ok || extensions[name] {
				continue
			}
			target, frag, _ := strings.Cut(url, "#")
			if resolve(target) != want || frag != strings.ToLower(name) {
				r, _ := filepath.Rel(bundle, want)
				t.Errorf("%s: [%s.%s](%s) does not land on it; link to %s#%s", rel, pkg, name, url,
					filepath.ToSlash(r), strings.ToLower(name))
			}
		}
		for _, m := range link.FindAllStringSubmatch(text, -1) {
			dest := resolve(m[1])
			if !strings.HasPrefix(dest, bundle+string(filepath.Separator)) || !strings.HasSuffix(dest, ".md") {
				continue // only links into the bundle are held to its anchors
			}
			if !anchors[dest][m[2]] {
				r, _ := filepath.Rel(bundle, dest)
				t.Errorf("%s: link to %s#%s lands on no heading", rel, filepath.ToSlash(r), m[2])
			}
		}
	}
	counts := []string{}
	for _, pkg := range append([]string{"goals", "product", "tech", "decisions", "acceptance"}, names...) {
		counts = append(counts, strconv.Itoa(len(defined[pkg]))+" "+pkg)
	}
	t.Logf("%s; %d references resolve", strings.Join(counts, ", "), refs)
}

// slug is a heading's anchor as a code host derives it: lowercased, markup and
// punctuation removed, spaces made hyphens. Letters outside ASCII survive, as
// they do in Python's \w and on GitHub.
func slug(heading string) string {
	s := strings.ToLower(strings.TrimSpace(heading))
	s = regexp.MustCompile("[`*]").ReplaceAllString(s, "")
	s = regexp.MustCompile(`[^\p{L}\p{N}\p{M}_\- ]`).ReplaceAllString(s, "")
	return strings.ReplaceAll(s, " ", "-")
}

func read(t *testing.T, path string) string {
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}
