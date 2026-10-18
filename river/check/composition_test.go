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

// packageItems reads every item in the bundle's packages: the sources, the
// composed packages, and the decisions.
func packageItems(t *testing.T) map[string]pkgItem {
	t.Helper()
	bundle, _ := filepath.Abs("..")
	req := filepath.Join(bundle, "requirements")
	files := map[string][]string{"decisions": {filepath.Join(req, "decisions.md")}}
	// Sources are two levels down, composed packages one; PROVENANCE.md
	// declares no package.
	sources, _ := filepath.Glob(filepath.Join(req, "sources", "*", "*.md"))
	composed, _ := filepath.Glob(filepath.Join(req, "*", "*.md"))
	for _, f := range append(sources, composed...) {
		if filepath.Base(f) == "PROVENANCE.md" {
			continue
		}
		pkg := filepath.Base(filepath.Dir(f))
		files[pkg] = append(files[pkg], f)
	}
	// A transcribed source exports capitalized names; every other package's
	// names start lower case, which also keeps a section heading such as
	// "### Diagram" from reading as an item.
	lower := regexp.MustCompile(`(?m)^#{2,3} ([a-z][A-Za-z]+)[ \t]*$`)
	upper := regexp.MustCompile(`(?m)^#{2,3} ([A-Z][A-Za-z]+)[ \t]*$`)
	items := map[string]pkgItem{}
	for pkg, fs := range files {
		for _, f := range fs {
			text := read(t, f)
			heading := lower
			if filepath.Base(f) == "requirements.md" && filepath.Base(filepath.Dir(filepath.Dir(f))) == "sources" {
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
	return items
}

// title is an item's bold title line.
func (it pkgItem) title() string {
	m := regexp.MustCompile(`^\s*\*\*(.+?)\*\*\s*\n`).FindStringSubmatch(it.body)
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

var listLine = regexp.MustCompile(`(?m)^- \[([a-z]+\.[A-Za-z]+)\]\([^)]*\) — (.+?)(?:\. Excluded by \[(decisions\.[A-Za-z]+)\]\([^)]*\))?\s*$`)

// composedLists reads every list line in the composed packages.
func composedLists(t *testing.T) []listed {
	t.Helper()
	bundle, _ := filepath.Abs("..")
	files, _ := filepath.Glob(filepath.Join(bundle, "requirements", "*", "*.md"))
	var out []listed
	for _, f := range files {
		if filepath.Base(filepath.Dir(f)) == "sources" || filepath.Base(f) == "acceptance.md" {
			continue
		}
		for _, m := range listLine.FindAllStringSubmatch(read(t, f), -1) {
			out = append(out, listed{m[1], m[2], m[3], f})
		}
	}
	return out
}

// field returns the comma-separated names in an item's field, or nil.
func field(body, name string) []string {
	m := regexp.MustCompile(`(?m)^` + name + `:\s*(.+)$`).FindStringSubmatch(body)
	if m == nil {
		return nil
	}
	var out []string
	for _, v := range strings.Split(m[1], ",") {
		if v = strings.Trim(strings.TrimSpace(v), "`"); v != "" {
			out = append(out, v)
		}
	}
	return out
}

// qualify resolves a reference made from inside pkg: bare names are pkg's own.
func qualify(pkg, ref string) string {
	if strings.Contains(ref, ".") {
		return ref
	}
	return pkg + "." + ref
}

// TestCompositionResolves checks the fields composition adds, following River's
// requirements-over-domains rule:
//
//   - Applies to: names a domain the source defines;
//   - Binds: names an imported domain, and each domain is bound once;
//   - every domain an adopted requirement applies to is bound locally, since a
//     requirement over a domain nobody bound applies to nothing, silently;
//   - Composed from: and Covers: name items that exist.
//
// A composed requirement no criterion covers is counted, not failed, as River
// counts an uncovered requirement.
//
// Verifies: process integrity. Criterion: acceptance.compositionResolves.
func TestCompositionResolves(t *testing.T) {
	items := packageItems(t)
	if len(items) == 0 {
		t.Fatal("no items parsed; the layout has changed and this check proves nothing")
	}
	exists := func(ref string) bool { _, ok := items[ref]; return ok }

	applied := map[string][]string{} // domain -> requirements over it
	bound := map[string][]string{}   // domain -> local requirements binding it
	covered := map[string]bool{}
	checked := 0
	for key, it := range items {
		for _, d := range field(it.body, "Applies to") {
			checked++
			d = qualify(it.pkg, d)
			if !exists(d) {
				t.Errorf("%s applies to %s, which is not defined", key, d)
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

	// The lists of what applies unchanged, and what is excluded.
	accounted := map[string]bool{}
	for _, it := range items {
		if filepath.Base(filepath.Dir(filepath.Dir(it.file))) != "sources" {
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
	// Every local requirement is included, excluded or composed from by some
	// package. Counted rather than failed while dataset and internal/repo are
	// still to compose (river/todo.md).
	var unaccounted []string
	for key, it := range items {
		if (it.pkg == "product" || it.pkg == "tech") && !accounted[key] {
			unaccounted = append(unaccounted, key)
		}
	}
	sort.Strings(unaccounted)
	t.Logf("%d local requirements in no composed package yet: %v", len(unaccounted), unaccounted)

	// Coverage of the composed packages: everything under requirements/ that
	// is not a source and not the decisions.
	var uncovered []string
	for key, it := range items {
		isSource := strings.Contains(it.file, string(filepath.Separator)+"sources"+string(filepath.Separator))
		isComposed := !isSource && it.pkg != "decisions" && filepath.Base(it.file) != "acceptance.md"
		if isComposed && field(it.body, "Covers") == nil && !covered[key] {
			uncovered = append(uncovered, key)
		}
	}
	sort.Strings(uncovered)
	t.Logf("%d composition references resolve; uncovered composed requirements: %v", checked, uncovered)
}
