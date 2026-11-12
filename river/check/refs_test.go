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
// criterion, or decision resolves, across every tracked file outside
// third_party/ and .agents/skills/:
//
//   - every qualified reference -- a goal, product, tech, decision, todo, or
//     composed-package name, and a transcribed source's exported name -- names
//     an item its package defines;
//   - every anchored link into the bundle lands on a heading that exists;
//   - a link whose text names an item lands on that item: the file that
//     defines it, at its anchor;
//   - no identifier in a superseded form (G/P/T/AC/CR-NNN, source/RN, or
//     unquoted acceptance.foo) survives anywhere.
//
// Verifies: tech.bundleIsNormative. Criterion: repo.decisionReferencesExist.
func TestReferencesResolve(t *testing.T) {
	bundle, _ := filepath.Abs("..")
	repo := filepath.Dir(bundle)

	lowerItem := regexp.MustCompile(`(?m)^#{2,3} ([a-z][A-Za-z0-9]*)[ \t]*$`)
	upperItem := regexp.MustCompile(`(?m)^#{2,3} ([A-Z][A-Za-z0-9]*)[ \t]*$`)
	glob := func(pattern string) []string {
		m, _ := filepath.Glob(filepath.Join(bundle, pattern))
		sort.Strings(m)
		return m
	}
	packages := map[string][]string{
		"goals":     glob("goals/*.md"),
		"product":   glob("requirements/local/product/*.md"),
		"tech":      glob("requirements/local/tech/*.md"),
		"decisions": {filepath.Join(bundle, "requirements", "decisions.md")},
		"todo":      {filepath.Join(bundle, "todo.md"), filepath.Join(bundle, "todo-closed.md")},
		"app":       glob("requirements/composed/internal/app/*.md"),
		"dataset":   glob("requirements/composed/internal/dataset/*.md"),
		"repo":      glob("requirements/composed/internal/repo/*.md"),
	}
	exported := map[string]bool{}
	dirs, _ := filepath.Glob(filepath.Join(bundle, "requirements", "imported", "*", "*"))
	for _, d := range dirs {
		pkg := filepath.Base(d)
		f := filepath.Join(d, pkg+".md")
		if _, err := os.Stat(f); err == nil {
			packages[pkg] = []string{f}
			exported[pkg] = true
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
	qualified := regexp.MustCompile(`\b(goals|product|tech|decisions|app|dataset|repo)\.([a-z][A-Za-z0-9]+)\b`)
	var names []string
	for pkg := range exported {
		names = append(names, pkg)
	}
	sort.Strings(names)
	var exportedRef *regexp.Regexp
	if len(names) > 0 {
		exportedRef = regexp.MustCompile(`\b(` + strings.Join(names, "|") + `)\.([A-Z][A-Za-z0-9]+)\b`)
	}
	legacy := regexp.MustCompile(`\b(?:G|P|T|AC|CR)-\d{3}\b|\b[a-z][a-z-]*/R\d+\b|\bacceptance\.[a-z][A-Za-z0-9]+\b`)
	inlineCode := regexp.MustCompile("`[^`\n]+`")
	link := regexp.MustCompile(`\]\(([^)\s#]*)#([^)\s]+)\)`)
	citation := regexp.MustCompile("\\[`?([a-z]+)\\.([A-Za-z0-9]+)`?\\]\\(([^)\\s]*)\\)")

	out, err := exec.Command("git", "-C", repo, "ls-files", "--cached", "--others", "--exclude-standard").Output()
	if err != nil {
		t.Fatalf("listing tracked files: %v; this check reads what git tracks", err)
	}
	refs := 0
	for _, rel := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if strings.HasPrefix(rel, "third_party/") || strings.HasPrefix(rel, ".agents/skills/") {
			continue
		}
		path := filepath.Join(repo, filepath.FromSlash(rel))
		b, err := os.ReadFile(path)
		if err != nil || !utf8.Valid(b) {
			continue
		}
		text := string(b)
		scanText := text
		if strings.HasSuffix(rel, ".md") {
			scanText = inlineCode.ReplaceAllString(fenced.ReplaceAllString(text, ""), "")
		}
		for _, re := range []*regexp.Regexp{qualified, exportedRef} {
			if re == nil {
				continue
			}
			for _, m := range re.FindAllStringSubmatch(scanText, -1) {
				if extensions[m[2]] {
					continue
				}
				refs++
				if _, ok := defined[m[1]][m[2]]; !ok {
					t.Errorf("%s: %s.%s names nothing package %s defines", rel, m[1], m[2], m[1])
				}
			}
		}
		if !strings.HasPrefix(rel, "river/check/") {
			for _, m := range legacy.FindAllString(scanText, -1) {
				if m == "acceptance.md" {
					continue
				}
				t.Errorf("%s: %s is a superseded identifier", rel, m)
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
		noFence := fenced.ReplaceAllString(text, "")
		for _, m := range citation.FindAllStringSubmatch(noFence, -1) {
			pkg, name, url := m[1], m[2], m[3]
			want, ok := defined[pkg][name]
			if !ok || extensions[name] {
				continue
			}
			target, frag, _ := strings.Cut(url, "#")
			// river/conformance.md headings are "## <pkg>.<criterion>", not the criterion's definition.
			if resolve(target) != want || frag != strings.ToLower(name) {
				r, _ := filepath.Rel(bundle, want)
				t.Errorf("%s: [%s.%s](%s) does not land on it; link to %s#%s", rel, pkg, name, url,
					filepath.ToSlash(r), strings.ToLower(name))
			}
		}
		for _, m := range link.FindAllStringSubmatch(noFence, -1) {
			dest := resolve(m[1])
			if !strings.HasPrefix(dest, bundle+string(filepath.Separator)) || !strings.HasSuffix(dest, ".md") {
				continue
			}
			if !anchors[dest][m[2]] {
				r, _ := filepath.Rel(bundle, dest)
				t.Errorf("%s: link to %s#%s lands on no heading", rel, filepath.ToSlash(r), m[2])
			}
		}
	}
	counts := []string{}
	for _, pkg := range append([]string{"goals", "product", "tech", "decisions", "app", "dataset", "repo"}, names...) {
		counts = append(counts, strconv.Itoa(len(defined[pkg]))+" "+pkg)
	}
	t.Logf("%s; %d references resolve", strings.Join(counts, ", "), refs)
}

// TestRelativeLinksResolve checks that every relative link in a tracked
// Markdown file lands on a file or directory that exists, anchored or not,
// including the READMEs this project writes for its vendored files under
// third_party/.
//
// Verifies: tech.bundleIsNormative. Criterion: repo.decisionReferencesExist.
func TestRelativeLinksResolve(t *testing.T) {
	bundle, _ := filepath.Abs("..")
	repo := filepath.Dir(bundle)
	out, err := exec.Command("git", "-C", repo, "ls-files", "--cached", "--others", "--exclude-standard", "*.md").Output()
	if err != nil {
		t.Fatalf("listing tracked files: %v; this check reads exactly what git tracks", err)
	}
	link := regexp.MustCompile(`\]\(([^)\s]+)\)`)
	checked := 0
	for _, rel := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if strings.HasPrefix(rel, ".agents/skills/") {
			continue
		}
		path := filepath.Join(repo, filepath.FromSlash(rel))
		for _, m := range link.FindAllStringSubmatch(fenced.ReplaceAllString(read(t, path), ""), -1) {
			target, _, _ := strings.Cut(m[1], "#")
			if target == "" || strings.HasPrefix(target, "/") || strings.Contains(target, ":") {
				continue
			}
			checked++
			dest := filepath.Join(filepath.Dir(path), filepath.FromSlash(target))
			if _, err := os.Stat(dest); err != nil {
				t.Errorf("%s: link to %s lands on nothing", rel, m[1])
			}
		}
	}
	if checked == 0 {
		t.Fatal("no relative links found; the layout or the pattern has changed and this check proves nothing")
	}
	t.Logf("%d relative links resolve", checked)
}

var fenced = regexp.MustCompile("(?ms)^```.*?^```[ \t]*$")

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
