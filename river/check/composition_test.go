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
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

// pkgItem is one heading-identified item in a package, with the text under it.
type pkgItem struct {
	pkg, name, file, body string
}

// inTree reports whether file lies under requirements/<tree>/.
func inTree(file, tree string) bool {
	return strings.Contains(filepath.ToSlash(file), "/requirements/"+tree+"/")
}

// packageItems reads every item in the bundle's packages: the sources, the
// composed packages, and the decisions.
func packageItems(t *testing.T) (map[string]pkgItem, map[string]bool) {
	t.Helper()
	bundle, _ := filepath.Abs("..")
	req := filepath.Join(bundle, "requirements")
	files := map[string][]string{"decisions": {filepath.Join(req, "decisions.md")}}
	local, _ := filepath.Glob(filepath.Join(req, "local", "*", "*.md"))
	imported, _ := filepath.Glob(filepath.Join(req, "imported", "*", "*", "*.md"))
	composed, _ := filepath.Glob(filepath.Join(req, "composed", "*", "*", "*.md"))
	composedPkgs := map[string]bool{}
	for _, f := range composed {
		composedPkgs[filepath.Base(filepath.Dir(f))] = true
	}
	for _, f := range append(append(local, imported...), composed...) {
		if filepath.Base(f) == "PROVENANCE.md" {
			continue
		}
		pkg := filepath.Base(filepath.Dir(f))
		files[pkg] = append(files[pkg], f)
	}
	lower := regexp.MustCompile(`(?m)^#{2,3} ([a-z][A-Za-z0-9]*)[ \t]*$`)
	upper := regexp.MustCompile(`(?m)^#{2,3} ([A-Z][A-Za-z0-9]*)[ \t]*$`)
	items := map[string]pkgItem{}
	for pkg, fs := range files {
		for _, f := range fs {
			text := read(t, f)
			heading := lower
			if inTree(f, "imported") {
				heading = upper
			}
			locs := heading.FindAllStringSubmatchIndex(text, -1)
			for i, l := range locs {
				end := len(text)
				if i+1 < len(locs) {
					end = locs[i+1][0]
				}
				name := text[l[2]:l[3]]
				items[pkg+"."+name] = pkgItem{pkg, name, f, text[l[1]:end]}
			}
		}
	}
	return items, composedPkgs
}

// title is an item's bold title line.
func (it pkgItem) title() string {
	m := regexp.MustCompile(`^\s*\*\*(.+?)\*\*`).FindStringSubmatch(it.body)
	if m == nil {
		return ""
	}
	return m[1]
}

// listed is one line of a composed package's list of requirements that apply
// unchanged, or that it excludes.
type listed struct {
	ref, title, decision, file string
}

var listLine = regexp.MustCompile(`(?m)^- \[([a-z]+\.[A-Za-z0-9]+)\]\([^)]*\)(?: \*\(excluded\)\*)? — (.+?)(?:\. Excluded by \[(decisions\.[A-Za-z0-9]+)\]\([^)]*\))?\s*$`)

// composedLists reads every list line in the composed packages.
func composedLists(t *testing.T) []listed {
	t.Helper()
	bundle, _ := filepath.Abs("..")
	files, _ := filepath.Glob(filepath.Join(bundle, "requirements", "composed", "*", "*", "*.md"))
	var out []listed
	for _, f := range files {
		if filepath.Base(f) == "acceptance.md" {
			continue
		}
		for _, m := range listLine.FindAllStringSubmatch(read(t, f), -1) {
			out = append(out, listed{m[1], m[2], m[3], f})
		}
	}
	return out
}

var fieldLink = regexp.MustCompile(`^\[([^\]]+)\]\([^)]*\)$`)

// field returns the comma-separated names in an item's field (joining indented
// continuation lines), or nil.
func field(body, name string) []string {
	lines := strings.Split(body, "\n")
	prefix := name + ":"
	var raw string
	found := false
	for i, line := range lines {
		if strings.HasPrefix(line, prefix) {
			parts := []string{strings.TrimSpace(strings.TrimPrefix(line, prefix))}
			for j := i + 1; j < len(lines); j++ {
				if len(lines[j]) == 0 || (lines[j][0] != ' ' && lines[j][0] != '\t') {
					break
				}
				if s := strings.TrimSpace(lines[j]); s != "" {
					parts = append(parts, s)
				}
			}
			raw = strings.Join(parts, " ")
			found = true
			break
		}
	}
	if !found {
		return nil
	}
	var out []string
	for _, v := range strings.Split(raw, ",") {
		v = strings.TrimSpace(v)
		if l := fieldLink.FindStringSubmatch(v); l != nil {
			v = l[1]
		}
		if v = strings.Trim(v, "`"); v != "" {
			out = append(out, v)
		}
	}
	return out
}

func qualify(pkg, ref string) string {
	if strings.Contains(ref, ".") {
		return ref
	}
	return pkg + "." + ref
}

// TestCompositionResolves checks the fields composition adds, following River
// 2.0's composition and domain-binding rules:
//
//   - Applies to: names either a composed package in this module (`app`,
//     `dataset`, `repo`) or a domain an imported source defines;
//   - Binds: names an imported domain, and each domain is bound once;
//   - every imported domain an adopted requirement applies to is bound locally,
//     since a requirement over a domain nobody bound applies to nothing;
//   - Composed from: and Covers: name items that exist;
//   - every listed title matches its source verbatim, and every local
//     requirement is accounted for across `app`, `dataset`, and `repo`.
//
// Verifies: tech.bundleIsNormative. Criterion: repo.compositionResolves.
func TestCompositionResolves(t *testing.T) {
	items, composedPkgs := packageItems(t)
	if len(items) == 0 {
		t.Fatal("no items parsed; the layout has changed and this check proves nothing")
	}
	exists := func(ref string) bool { _, ok := items[ref]; return ok }

	applied := map[string][]string{}
	bound := map[string][]string{}
	covered := map[string]bool{}
	checked := 0
	for key, it := range items {
		if it.pkg == "decisions" {
			continue
		}
		for _, d := range field(it.body, "Applies to") {
			checked++
			if composedPkgs[d] {
				continue
			}
			d = qualify(it.pkg, d)
			if !exists(d) {
				t.Errorf("%s applies to %s, which is neither a composed package nor a defined domain", key, d)
			}
			applied[d] = append(applied[d], key)
		}
		for _, d := range field(it.body, "Binds") {
			checked++
			if !strings.Contains(d, ".") || !exists(d) {
				t.Errorf("%s binds %s, which is not an imported domain that exists", key, d)
			}
			bound[d] = append(bound[d], key)
		}
		for _, f := range []string{"Composed from", "Covers"} {
			for _, ref := range field(it.body, f) {
				checked++
				ref = qualify(it.pkg, ref)
				if !exists(ref) {
					t.Errorf("%s: %s names %s, which is not defined", key, f, ref)
				}
				if f == "Covers" {
					covered[ref] = true
				}
			}
		}
	}
	for d, reqs := range applied {
		switch n := len(bound[d]); {
		case n == 0:
			sort.Strings(reqs)
			t.Errorf("nothing binds %s, so %s apply to nothing; add a local requirement with Binds: %s",
				d, strings.Join(reqs, ", "), d)
		case n > 1:
			t.Errorf("%s is bound %d times (%s); a domain has one local definition",
				d, n, strings.Join(bound[d], ", "))
		}
	}
	if checked == 0 {
		t.Fatal("no composition fields found; the layout has changed and this check proves nothing")
	}

	accounted := map[string]bool{}
	for _, it := range items {
		if !inTree(it.file, "local") && !inTree(it.file, "imported") && it.pkg != "decisions" {
			for _, ref := range field(it.body, "Composed from") {
				accounted[qualify(it.pkg, ref)] = true
			}
		}
	}
	inPackage := map[string]string{}
	lists := composedLists(t)
	for _, l := range lists {
		checked++
		src, ok := items[l.ref]
		switch {
		case !ok:
			t.Errorf("%s lists %s, which is not defined", filepath.Base(l.file), l.ref)
			continue
		case src.title() != l.title:
			t.Errorf("%s lists %s as %q; its title is %q", filepath.Base(l.file), l.ref, l.title, src.title())
		case l.decision != "" && !exists(l.decision):
			t.Errorf("%s excludes %s by %s, which is not defined", filepath.Base(l.file), l.ref, l.decision)
		}
		key := l.file + " " + l.ref
		if prev, dup := inPackage[key]; dup {
			t.Errorf("%s lists %s twice (%s)", filepath.Base(l.file), l.ref, prev)
		}
		inPackage[key] = l.title
		accounted[l.ref] = true
	}
	if len(lists) == 0 {
		t.Fatal("no listed requirements found in any composed package; the list form has changed")
	}
	var unaccounted []string
	for key, it := range items {
		if (it.pkg == "product" || it.pkg == "tech") && !accounted[key] {
			unaccounted = append(unaccounted, key)
		}
	}
	sort.Strings(unaccounted)
	if len(unaccounted) > 0 {
		t.Errorf("%d local requirements in no composed package: %v", len(unaccounted), unaccounted)
	}

	var uncovered []string
	for key, it := range items {
		isSource := inTree(it.file, "local") || inTree(it.file, "imported")
		isComposed := !isSource && it.pkg != "decisions" && filepath.Base(it.file) != "acceptance.md"
		if isComposed && field(it.body, "Covers") == nil && !covered[key] {
			uncovered = append(uncovered, key)
		}
	}
	sort.Strings(uncovered)
	t.Logf("%d composition references resolve; uncovered composed requirements: %v", checked, uncovered)
}
