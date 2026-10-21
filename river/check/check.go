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

// Package check is River's checker, translated into Go so it runs with
// BirdView's own tests (River's decisions.checksInRepoLanguage).
//
// Translated from River's river/check/goals.py at pin 5fad157
// (sso://user/sameer/river). What is normative is what the checker reports,
// not how: a problem's code and path must match the reference exactly, and
// testdata/ -- River's shared cases, copied at the same pin -- holds that to
// account. When River's checker changes, retranslate and move the pin.
package check

import (
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"
)

// Problem is one thing wrong with a bundle. Code and Path are the contract a
// translation shares with the reference; Message is for people.
type Problem struct {
	Code, Path, Message string
}

// Result is what checking a bundle found.
type Result struct {
	Requirements, Goals, Gradients int
	// Counted, never failed: requirements citing no goal, and goals nothing cites.
	Unjustified, Unserved int
	// UnjustifiedItems names the requirements counted in Unjustified, as
	// "path#identifier". Beyond the shared contract, so a project can be
	// stricter than River about which of them it tolerates.
	UnjustifiedItems []string
	Problems         []Problem
	// Notes are symptoms worth a person's look, never failures
	// (River's product.alignmentReview).
	Notes []Problem
	// Stopped is set when a sanity check failed before anything was counted.
	Stopped bool
}

var (
	// An identifier-shaped heading is read as an item and then checked, so
	// `## item2` is reported rather than skipped; `## Non-goals` is left alone.
	h2Item = regexp.MustCompile(`(?m)^## ([A-Za-z][A-Za-z0-9_]*)[ \t]*$`)
	h3Item = regexp.MustCompile(`(?m)^### ([A-Za-z][A-Za-z0-9_]*)[ \t]*$`)
	// The title is the bold line beneath the heading, optionally followed by a
	// status marker such as *(withdrawn)*.
	titleLine     = regexp.MustCompile(`(?m)\A\s*\*\*(.+?)\*\*(?:[ \t]+\*\([a-z ]+\)\*)?[ \t]*$`)
	oldForm       = regexp.MustCompile(`(?m)^#{2,3} [A-Za-z]+ — .+$`)
	descriptiveID = regexp.MustCompile(`^[A-Za-z]+$`)
	packageDecl   = regexp.MustCompile(`^package\s+([A-Za-z][A-Za-z0-9]*)\s*$`)
	moduleBlock   = regexp.MustCompile("(?ms)^## Module\\s+```[^\\n]*\\n(.*?)^```")
	moduleDecl    = regexp.MustCompile(`(?m)^module\s+\S+\s*$`)
	goalsField    = regexp.MustCompile(`(?m)^Goals:\s*(.+)$`)
	servesField   = regexp.MustCompile(`(?m)^Serves:\s*(.+)$`)
	goalRef       = regexp.MustCompile(`\bgoals\.([A-Za-z0-9_]+)\b`)
	// A qualified reference such as product.stableSurface.
	qualifiedRef = regexp.MustCompile(`\b[a-z][a-z0-9]*\.[a-z][a-z0-9]*[A-Z][A-Za-z0-9]*\b`)
	// A goal's text runs from its title to the next heading or rule.
	goalEnd   = regexp.MustCompile(`(?m)^(?:#|---)`)
	paragraph = regexp.MustCompile(`\n\s*\n`)
)

// River's product.seekableArtifacts: an agent reads by searching and opening
// ranges. Lengths count characters, not bytes.
const (
	maxLine         = 500
	maxInstructions = 6000
)

type item struct {
	path, ident, title string
	cited              []string
}

type checker struct {
	bundle   string
	problems []Problem
}

func (c *checker) report(code, path, msg string) {
	c.problems = append(c.problems, Problem{code, filepath.ToSlash(path), msg})
}

func (c *checker) rel(p string) string {
	r, _ := filepath.Rel(c.bundle, p)
	return filepath.ToSlash(r)
}

// skipped reports whether path lies outside the bundle's packages: the
// checks' own test bundles under check/.
func (c *checker) skipped(path string) bool {
	return strings.HasPrefix(c.rel(path), "check/")
}

// imported reports whether path lies under requirements/imported/: another
// module's package, pinned here.
func (c *checker) imported(path string) bool {
	return strings.HasPrefix(c.rel(path), "requirements/imported/")
}

// markdown lists the .md files under dir, sorted, that belong to the bundle.
func (c *checker) markdown(dir string, provenance bool) []string {
	var out []string
	_ = filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(p, ".md") || c.skipped(p) {
			return nil
		}
		if !provenance && d.Name() == "PROVENANCE.md" {
			return nil
		}
		out = append(out, p)
		return nil
	})
	sort.Strings(out)
	return out
}

func firstLine(text string) string {
	for _, l := range strings.Split(text, "\n") {
		if s := strings.TrimSpace(l); s != "" {
			return s
		}
	}
	return ""
}

// packageDir parses the items in one kind of package directory.
func (c *checker) packageDir(dir string, heading, field *regexp.Regexp, expected string) []item {
	var items []item
	seen := map[[2]string]string{}
	pkgsByDir := map[string]map[string]bool{}
	for _, path := range c.markdown(dir, false) {
		rel := c.rel(path)
		b, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		text := string(b)
		pkg := ""
		if m := packageDecl.FindStringSubmatch(firstLine(text)); m == nil {
			c.report("missing-package-decl", rel, "first non-empty line must be `package <name>`")
		} else {
			pkg = m[1]
			if base := filepath.Base(filepath.Dir(path)); base != pkg {
				c.report("package-dir-mismatch", rel, "declares package "+pkg+" in a directory named "+base+
					"; a package's directory carries its name")
			}
			if expected != "" && pkg != expected {
				c.report("unexpected-package", rel, "declares package "+pkg+", expected "+expected)
			}
			d := filepath.Dir(path)
			if pkgsByDir[d] == nil {
				pkgsByDir[d] = map[string]bool{}
			}
			pkgsByDir[d][pkg] = true
		}
		for _, old := range oldForm.FindAllString(text, -1) {
			c.report("superseded-heading", rel, "`"+old+"` uses the superseded heading form; the heading is "+
				"the identifier alone and the title goes on the bold line beneath it")
		}
		heads := heading.FindAllStringSubmatchIndex(text, -1)
		for i, h := range heads {
			ident := text[h[2]:h[3]]
			title := ""
			if tm := titleLine.FindStringSubmatch(text[h[1]:]); tm != nil {
				title = tm[1]
			} else {
				c.report("missing-title", rel, "`"+ident+"` has no title; put it in bold on the line beneath the heading")
			}
			if !descriptiveID.MatchString(ident) {
				c.report("bad-identifier", rel, "identifier `"+ident+"` must be a descriptive CamelCase Go identifier (letters only)")
			}
			key := [2]string{pkg, ident}
			if first, dup := seen[key]; dup {
				c.report("duplicate-identifier", rel, "duplicate identifier `"+ident+"` in package "+pkg+" (also in "+first+")")
			} else {
				seen[key] = rel
			}
			var cited []string
			if field != nil {
				end := len(text)
				if i+1 < len(heads) {
					end = heads[i+1][0]
				}
				if f := field.FindStringSubmatch(text[h[1]:end]); f != nil {
					for _, g := range goalRef.FindAllStringSubmatch(f[1], -1) {
						cited = append(cited, g[1])
					}
				}
			}
			items = append(items, item{rel, ident, title, cited})
		}
	}
	// Reported on the directory, not on whichever file sorts second, so every
	// implementation reports the same problem whatever order it walks in.
	dirs := make([]string, 0, len(pkgsByDir))
	for d := range pkgsByDir {
		dirs = append(dirs, d)
	}
	sort.Strings(dirs)
	for _, d := range dirs {
		if len(pkgsByDir[d]) > 1 {
			names := make([]string, 0, len(pkgsByDir[d]))
			for p := range pkgsByDir[d] {
				names = append(names, p)
			}
			sort.Strings(names)
			c.report("mixed-packages", c.rel(d), "files in one directory declare different packages: "+strings.Join(names, ", "))
		}
	}
	return items
}

// uniquePackageNames reports a package name declared in more than one of this
// module's own directories: with no import block, a bare prefix resolves by
// name alone. Imported packages belong to other modules and are not counted.
func (c *checker) uniquePackageNames() {
	dirs := map[string]map[string]bool{}
	for _, path := range c.markdown(c.bundle, true) {
		if c.imported(path) {
			continue
		}
		b, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		if m := packageDecl.FindStringSubmatch(firstLine(string(b))); m != nil {
			if dirs[m[1]] == nil {
				dirs[m[1]] = map[string]bool{}
			}
			dirs[m[1]][filepath.Dir(path)] = true
		}
	}
	for pkg, ds := range dirs {
		if len(ds) > 1 {
			c.report("duplicate-package-name", "-", "package "+pkg+" is declared in more than one directory; "+
				"package names must be unique within a module so a bare prefix resolves")
		}
	}
}

// seekable reports a .md file with a line over maxLine characters outside a
// code block, once per file, and an AGENTS.md beside the bundle over
// maxInstructions characters.
func (c *checker) seekable() {
	for _, path := range c.markdown(c.bundle, true) {
		b, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		fenced, longest := false, 0
		for _, line := range strings.Split(string(b), "\n") {
			if strings.HasPrefix(strings.TrimSpace(line), "```") {
				fenced = !fenced
			} else if n := utf8.RuneCountInString(line); !fenced && n > longest {
				longest = n
			}
		}
		if longest > maxLine {
			c.report("line-too-long", c.rel(path), strconv.Itoa(longest)+" characters on one line outside a code block; "+
				"the limit is "+strconv.Itoa(maxLine)+", so give a long item a heading")
		}
	}
	if b, err := os.ReadFile(filepath.Join(filepath.Dir(c.bundle), "AGENTS.md")); err == nil {
		if n := utf8.RuneCount(b); n > maxInstructions {
			c.report("instructions-too-long", "../AGENTS.md", strconv.Itoa(n)+" characters; the limit is "+
				strconv.Itoa(maxInstructions)+". Keep commands, damage-preventing rules and a map; move detail out")
		}
	}
}

// Run checks the bundle rooted at dir.
func Run(dir string) Result {
	bundle, _ := filepath.Abs(dir)
	c := &checker{bundle: bundle}
	goalsDir := filepath.Join(bundle, "goals")
	reqMD := filepath.Join(bundle, "requirements", "README.md")
	local := filepath.Join(bundle, "requirements", "local")

	c.uniquePackageNames()
	c.seekable()
	if b, err := os.ReadFile(reqMD); err != nil {
		c.report("missing-readme-md", "requirements/README.md", "missing requirements/README.md")
	} else if m := moduleBlock.FindStringSubmatch(string(b)); m == nil || !moduleDecl.MatchString(m[1]) {
		c.report("missing-module", "requirements/README.md", "missing `## Module` fenced block with `module <path>`")
	}

	goalItems := c.packageDir(goalsDir, h2Item, nil, "goals")
	gradItems := c.packageDir(filepath.Join(bundle, "gradients"), h3Item, servesField, "gradients")
	reqItems := c.packageDir(local, h2Item, goalsField, "")
	// Another module's packages: structure only, not counted, citations not
	// ours to resolve.
	c.packageDir(filepath.Join(bundle, "requirements", "imported"), h2Item, nil, "")

	goals := map[string]string{}
	for _, it := range goalItems {
		goals[it.ident] = it.title
	}
	// No goals at all is a brownfield bundle; no goals while goal files exist
	// means the layout or the parser changed.
	if len(goals) == 0 && len(c.markdown(goalsDir, true)) > 0 {
		return Result{Stopped: true, Problems: []Problem{{"no-goals-parsed", "goals",
			"goal files exist and none parsed -- parser or layout has changed"}}}
	}
	if len(reqItems) == 0 {
		return Result{Stopped: true, Problems: []Problem{{"no-requirements", "requirements/local",
			"no requirements found -- parser or layout has changed"}}}
	}

	r := Result{Requirements: len(reqItems), Goals: len(goals), Gradients: len(gradItems)}
	citedAnywhere := map[string]bool{}
	for _, it := range reqItems {
		if len(it.cited) == 0 {
			r.Unjustified++
			r.UnjustifiedItems = append(r.UnjustifiedItems, it.path+"#"+it.ident)
			continue
		}
		for _, g := range it.cited {
			citedAnywhere[g] = true
			if _, ok := goals[g]; !ok {
				c.report("unknown-goal", it.path, it.ident+" cites goals."+g+", which package goals does not declare")
			}
		}
	}
	for _, it := range gradItems {
		if len(it.cited) == 0 {
			c.report("gradient-cites-no-goal", it.path, it.ident+" ("+it.title+") cites no `goals.<id>` in Serves:")
			continue
		}
		for _, g := range it.cited {
			if _, ok := goals[g]; !ok {
				c.report("gradient-unknown-goal", it.path, it.ident+" serves goals."+g+", which package goals does not declare")
			}
		}
	}
	for g := range goals {
		if !citedAnywhere[g] {
			r.Unserved++
		}
	}
	r.Problems = c.problems
	r.Notes = c.alignmentNotes(goalsDir, reqItems)
	return r
}

// alignmentNotes reports River's four alignment symptoms: a goal cited by one
// requirement, a goal whose citing requirements all cite one other goal too,
// a goal longer than one paragraph, and a goal containing code or a qualified
// reference. Each can be right; each is where a person looks first.
func (c *checker) alignmentNotes(goalsDir string, reqItems []item) []Problem {
	type goal struct{ path, title, body string }
	goals := map[string]goal{}
	for _, path := range c.markdown(goalsDir, true) {
		b, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		text := string(b)
		for _, h := range h2Item.FindAllStringSubmatchIndex(text, -1) {
			start, title := h[1], ""
			if tm := titleLine.FindStringSubmatchIndex(text[h[1]:]); tm != nil {
				start, title = h[1]+tm[1], text[h[1]+tm[2]:h[1]+tm[3]]
			}
			end := len(text)
			if e := goalEnd.FindStringIndex(text[start:]); e != nil {
				end = start + e[0]
			}
			goals[text[h[2]:h[3]]] = goal{c.rel(path), title, strings.TrimSpace(text[start:end])}
		}
	}
	citedBy := map[string]map[string]bool{}
	for g := range goals {
		citedBy[g] = map[string]bool{}
	}
	for _, it := range reqItems {
		for _, g := range it.cited {
			if citedBy[g] != nil {
				citedBy[g][it.path+"#"+it.ident] = true
			}
		}
	}
	ids := make([]string, 0, len(goals))
	for g := range goals {
		ids = append(ids, g)
	}
	sort.Strings(ids)
	subset := func(a, b map[string]bool) bool {
		for k := range a {
			if !b[k] {
				return false
			}
		}
		return true
	}
	var notes []Problem
	for _, g := range ids {
		gl, mine := goals[g], citedBy[g]
		if len(mine) == 1 {
			notes = append(notes, Problem{"thin-goal", gl.path,
				"goals." + g + " is cited by one requirement; is it narrow, or under-served?"})
		}
		// Two or more, so a goal with one citation is reported once, as thin.
		// Equal sets are reported once, on the goal that sorts later.
		if len(mine) >= 2 {
			for _, h := range ids {
				other := citedBy[h]
				if h != g && subset(mine, other) && (len(mine) != len(other) || h < g) {
					notes = append(notes, Problem{"contained-goal", gl.path,
						"every requirement citing goals." + g + " also cites goals." + h + "; are they one goal?"})
					break
				}
			}
		}
		paras := 0
		for _, p := range paragraph.Split(gl.body, -1) {
			if strings.TrimSpace(p) != "" {
				paras++
			}
		}
		if paras > 1 {
			notes = append(notes, Problem{"long-goal", gl.path,
				"goals." + g + " runs to more than one paragraph; a goal needs one"})
		}
		if strings.Contains(gl.title+gl.body, "`") || qualifiedRef.MatchString(gl.title+" "+gl.body) {
			notes = append(notes, Problem{"goal-names-implementation", gl.path,
				"goals." + g + " contains code or a qualified reference; a goal should survive a rewrite of the implementation"})
		}
	}
	return notes
}
