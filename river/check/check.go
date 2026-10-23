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
// Translated from River's river/check/goals.py at pin f8252f5 (River is not
// yet published). What is normative is what the checker reports,
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
	fencedBlock   = regexp.MustCompile("(?ms)^\\s*```.*?^\\s*```")
	packageDecl   = regexp.MustCompile(`^package\s+([A-Za-z][A-Za-z0-9]*)\s*$`)
	moduleBlock   = regexp.MustCompile("(?ms)^## Module\\s+```[^\\n]*\\n(.*?)^```")
	moduleDecl    = regexp.MustCompile(`(?m)^module\s+\S+\s*$`)
	goalsField    = regexp.MustCompile(`(?m)^Goals:\s*(.+)$`)
	servesField   = regexp.MustCompile(`(?m)^Serves:\s*(.+)$`)
	// A qualified reference such as product.stableSurface.
	qualifiedRef = regexp.MustCompile(`\b[a-z][a-z0-9]*\.[a-z][a-z0-9]*[A-Z][A-Za-z0-9]*\b`)
	// A goal's text runs from its title to the next heading or rule.
	goalEnd   = regexp.MustCompile(`(?m)^(?:#|---)`)
	paragraph = regexp.MustCompile(`\n\s*\n`)
	// decisions.md fields, one per line (River's README, "Keeping the bundle's files").
	decisionDate   = regexp.MustCompile(`(?m)^Date:\s*\d{4}-\d{2}-\d{2}\s*$`)
	decisionStatus = regexp.MustCompile(`^(accepted|amended by ([A-Za-z]+(?:, [A-Za-z]+)*)|superseded by ([A-Za-z]+))\s*(?:;.*)?$`)
	identList      = regexp.MustCompile(`[A-Za-z]+`)
	// River's product.linkedReferences: a qualified reference is package.identifier, the
	// identifier camelCase or Exported. Go has no lookbehind, so findRefs filters the characters
	// around each match by hand, as goals.py's pattern does with lookarounds.
	refPattern  = regexp.MustCompile(`([a-z][a-z0-9]*)\.([a-z]+[A-Z][A-Za-z0-9]*|[A-Z][A-Za-z0-9]*)\b`)
	goalPattern = regexp.MustCompile(`(goals)\.([A-Za-z][A-Za-z0-9_]*)\b`)
	linkRe      = regexp.MustCompile(`\[([^\]]*)\]\(([^)\s]*)\)`)
	headingRe   = regexp.MustCompile(`(?m)^#{1,6} (.+?)[ \t]*$`)
	codeSpanRe  = regexp.MustCompile("`[^`]*`")
	slugDrop    = regexp.MustCompile(`[^\w\- ]`)
)

type ref struct {
	start, end int
	pkg, ident string
}

func isWordByte(b byte) bool {
	return b == '_' || b >= '0' && b <= '9' || b >= 'a' && b <= 'z' || b >= 'A' && b <= 'Z'
}

// findRefs returns the qualified references in s that are not part of a path, a URL or a longer
// name: not preceded by a word character, '/', '#', '.' or '-', and not followed by a word
// character, '(' or '/'.
func findRefs(s string) []ref { return findWith(refPattern, s) }

func findWith(pattern *regexp.Regexp, s string) []ref {
	var out []ref
	for _, m := range pattern.FindAllStringSubmatchIndex(s, -1) {
		if m[0] > 0 {
			if b := s[m[0]-1]; isWordByte(b) || strings.IndexByte("/#.-", b) >= 0 {
				continue
			}
		}
		if m[1] < len(s) {
			if b := s[m[1]]; isWordByte(b) || b == '(' || b == '/' {
				continue
			}
		}
		out = append(out, ref{m[0], m[1], s[m[2]:m[3]], s[m[4]:m[5]]})
	}
	return out
}

// goalRefs returns the goal identifiers cited in s.
func goalRefs(s string) []string {
	var out []string
	for _, r := range findWith(goalPattern, s) {
		out = append(out, r.ident)
	}
	return out
}

func anchorSlug(heading string) string {
	return strings.ReplaceAll(slugDrop.ReplaceAllString(strings.ToLower(strings.TrimSpace(heading)), ""), " ", "-")
}

func inSpans(pos int, spans [][]int) bool {
	for _, s := range spans {
		if pos >= s[0] && pos < s[1] {
			return true
		}
	}
	return false
}

// symbolTable maps package -> identifier -> the file defining it, for every item the module can
// resolve: goals, gradients, local, imported and composed packages, and standalone files
// such as decisions.md, under the package each declares.
func (c *checker) symbolTable() map[string]map[string]string {
	table := map[string]map[string]string{}
	add := func(pkg, path string, heading *regexp.Regexp) {
		b, err := os.ReadFile(path)
		if err != nil {
			return
		}
		for _, m := range heading.FindAllStringSubmatch(string(b), -1) {
			if table[pkg] == nil {
				table[pkg] = map[string]string{}
			}
			table[pkg][m[1]] = path
		}
	}
	for _, p := range c.markdown(filepath.Join(c.bundle, "goals"), true) {
		add("goals", p, h2Item)
	}
	for _, p := range c.markdown(filepath.Join(c.bundle, "gradients"), true) {
		add("gradients", p, h3Item)
	}
	for _, sub := range []string{"local", "imported", "composed"} {
		for _, p := range c.markdown(filepath.Join(c.bundle, "requirements", sub), false) {
			b, err := os.ReadFile(p)
			if err != nil {
				continue
			}
			if m := packageDecl.FindStringSubmatch(firstLine(string(b))); m != nil {
				add(m[1], p, h2Item)
			}
		}
	}
	// River's tech.declaredPackages: a file outside the package directories is a package named
	// by its declaration. One that lacks it is reported once, and resolved under its file's name
	// so that its references do not fail a second time.
	for _, p := range c.standaloneItemFiles() {
		b, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		text := string(b)
		pkg := filePackage(p)
		if m := packageDecl.FindStringSubmatch(firstLine(text)); m != nil {
			pkg = m[1]
		}
		add(pkg, p, h2Item)
		for _, h := range h3Item.FindAllStringSubmatchIndex(text, -1) {
			if titleLine.MatchString(text[h[1]:]) {
				if table[pkg] == nil {
					table[pkg] = map[string]string{}
				}
				table[pkg][text[h[2]:h[3]]] = p
			}
		}
	}
	return table
}

// filePackage is the package a standalone file is named for: its name up to any
// hyphen, lowercased. decisions.md is decisions; todo.md and todo-closed.md are todo.
func filePackage(path string) string {
	name := strings.TrimSuffix(filepath.Base(path), ".md")
	name, _, _ = strings.Cut(name, "-")
	return strings.ToLower(name)
}

// inPackageDir reports whether path is under goals/, gradients/, or a requirements
// package directory, where a package is named by its directory.
func (c *checker) inPackageDir(path string) bool {
	rel := c.rel(path)
	for _, d := range []string{"goals/", "gradients/", "requirements/local/", "requirements/imported/", "requirements/composed/"} {
		if strings.HasPrefix(rel, d) {
			return true
		}
	}
	return false
}

// definesItems reports an identifier-shaped ## or ### heading with a bold title
// beneath it, outside code blocks.
func definesItems(text string) bool {
	text = fencedBlock.ReplaceAllString(text, "")
	for _, re := range []*regexp.Regexp{h2Item, h3Item} {
		for _, h := range re.FindAllStringIndex(text, -1) {
			if titleLine.MatchString(text[h[1]:]) {
				return true
			}
		}
	}
	return false
}

// standaloneItemFiles lists the files outside the package directories whose
// headings define items (River's tech.declaredPackages). The bundle's README.md is
// the method, copied unchanged, and its items are examples; PROVENANCE.md defines
// nothing.
func (c *checker) standaloneItemFiles() []string {
	var out []string
	for _, p := range c.markdown(c.bundle, false) {
		if c.inPackageDir(p) || c.rel(p) == "README.md" {
			continue
		}
		if b, err := os.ReadFile(p); err == nil && definesItems(string(b)) {
			out = append(out, p)
		}
	}
	return out
}

// standalonePackages reports a standalone item file that does not declare the
// package it is named for.
func (c *checker) standalonePackages() {
	for _, p := range c.standaloneItemFiles() {
		b, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		want := filePackage(p)
		if m := packageDecl.FindStringSubmatch(firstLine(string(b))); m == nil {
			c.report("missing-package-decl", c.rel(p), "defines items, so its first non-empty line must be `package "+want+"`")
		} else if m[1] != want {
			c.report("package-file-mismatch", c.rel(p), "declares package "+m[1]+"; a file outside a package directory "+
				"is named for its package, so this one is "+want)
		}
	}
}

// references reports, outside the method README and the packages that travel between modules,
// a qualified reference the module can resolve that is not a link, one naming nothing its
// package declares, and a link whose file or heading does not exist (River's
// product.linkedReferences).
func (c *checker) references() {
	table := c.symbolTable()
	anchors := map[string]map[string]bool{}
	anchorsOf := func(path string) map[string]bool {
		if a, ok := anchors[path]; ok {
			return a
		}
		b, err := os.ReadFile(path)
		if err != nil {
			anchors[path] = nil
			return nil
		}
		a := map[string]bool{}
		for _, m := range headingRe.FindAllStringSubmatch(string(b), -1) {
			a[anchorSlug(m[1])] = true
		}
		anchors[path] = a
		return a
	}
	for _, path := range c.markdown(c.bundle, true) {
		rel := c.rel(path)
		if rel == "README.md" || strings.HasPrefix(rel, "requirements/composed/exported/") || c.imported(path) {
			continue
		}
		b, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		fenced, openCode := false, false
		for i, line := range strings.Split(string(b), "\n") {
			n := strconv.Itoa(i + 1)
			if strings.HasPrefix(strings.TrimSpace(line), "```") {
				fenced = !fenced
				continue
			}
			if fenced || strings.HasPrefix(strings.TrimLeft(line, " \t"), "#") {
				continue
			}
			// Padding marks a carried piece as part of a longer span, never one bare name.
			if openCode {
				line = "`… " + line
			}
			openCode = strings.Count(line, "`")%2 == 1
			if openCode {
				line += " …`"
			}
			code := codeSpanRe.FindAllStringIndex(line, -1)
			var spans [][]int
			for _, m := range linkRe.FindAllStringSubmatchIndex(line, -1) {
				if inSpans(m[0], code) {
					continue
				}
				spans = append(spans, m[:2])
				text := strings.Trim(line[m[2]:m[3]], "`")
				rs := findRefs(text)
				if len(rs) != 1 || rs[0].start != 0 || rs[0].end != len(text) || table[rs[0].pkg] == nil {
					continue
				}
				r := rs[0]
				if _, ok := table[r.pkg][r.ident]; !ok {
					c.report("unknown-reference", rel, "line "+n+": "+r.pkg+"."+r.ident+" names nothing package "+r.pkg+" declares")
					continue
				}
				dest := line[m[4]:m[5]]
				target, frag, _ := strings.Cut(dest, "#")
				file := path
				if target != "" {
					file = filepath.Join(filepath.Dir(path), filepath.FromSlash(target))
				}
				if a := anchorsOf(file); a == nil || frag != "" && !a[frag] {
					c.report("broken-reference-link", rel, "line "+n+": the link for "+r.pkg+"."+r.ident+" leads to "+dest+", which does not exist")
				}
			}
			for _, r := range findRefs(line) {
				if inSpans(r.start, spans) || table[r.pkg] == nil {
					continue
				}
				// Inside code formatting, only a span that is exactly the name is a reference;
				// longer code (a comment, a field, a quoted message) is code.
				var span []int
				for _, cs := range code {
					if r.start >= cs[0] && r.start < cs[1] {
						span = cs
					}
				}
				if span != nil && line[span[0]+1:span[1]-1] != line[r.start:r.end] {
					continue
				}
				if _, ok := table[r.pkg][r.ident]; ok {
					c.report("unlinked-reference", rel, "line "+n+": "+r.pkg+"."+r.ident+" is not a link to its definition")
				} else if span == nil {
					// In code formatting, a name that resolves to nothing is an example or a name
					// from the past, not a reference.
					c.report("unknown-reference", rel, "line "+n+": "+r.pkg+"."+r.ident+" names nothing package "+r.pkg+" declares")
				}
			}
		}
	}
}

// decisionField returns the value of a `Name:` line in an entry, and whether it is present.
func decisionField(body, name string) (string, bool) {
	m := regexp.MustCompile(`(?m)^` + regexp.QuoteMeta(name) + `:[ \t]*(.*)$`).FindStringSubmatch(body)
	if m == nil {
		return "", false
	}
	return strings.TrimSpace(m[1]), true
}

func decisionIDs(v string) []string {
	if v == "" || v == "none" || v == "—" {
		return nil
	}
	return identList.FindAllString(v, -1)
}

func contains(list []string, s string) bool {
	for _, x := range list {
		if x == s {
			return true
		}
	}
	return false
}

// decisions checks requirements/decisions.md, if present: the fields every entry carries, and
// supersession and amendment recorded at both ends.
func (c *checker) decisions(reqItems []item) {
	path := filepath.Join(c.bundle, "requirements", "decisions.md")
	b, err := os.ReadFile(path)
	if err != nil {
		return
	}
	rel := c.rel(path)
	text := string(b)
	heads := h2Item.FindAllStringSubmatchIndex(text, -1)
	var order []string
	bodies := map[string]string{}
	for i, h := range heads {
		end := len(text)
		if i+1 < len(heads) {
			end = heads[i+1][0]
		}
		id := text[h[2]:h[3]]
		order = append(order, id)
		bodies[id] = text[h[1]:end]
	}
	declared := map[string]map[string]bool{}
	for _, it := range reqItems {
		pkg := filepath.Base(filepath.Dir(it.path))
		if declared[pkg] == nil {
			declared[pkg] = map[string]bool{}
		}
		declared[pkg][it.ident] = true
	}
	type state struct {
		kind string // accepted, amended, superseded
		by   []string
	}
	status := map[string]state{}
	supersedes, amends := map[string][]string{}, map[string][]string{}
	for _, id := range order {
		body := bodies[id]
		var missing []string
		if !decisionDate.MatchString(body) {
			missing = append(missing, "Date (YYYY-MM-DD)")
		}
		for _, n := range []string{"Decided by", "Status", "Affects"} {
			if _, ok := decisionField(body, n); !ok {
				missing = append(missing, n)
			}
		}
		if len(missing) > 0 {
			c.report("decision-missing-field", rel, id+" lacks "+strings.Join(missing, ", "))
		}
		st := state{kind: "accepted"}
		if raw, ok := decisionField(body, "Status"); ok {
			if m := decisionStatus.FindStringSubmatch(raw); m == nil {
				c.report("decision-bad-status", rel, id+" has Status `"+raw+"`; use accepted, amended by <id>, "+
					"or superseded by <id>, optionally followed by `; note`")
			} else if m[2] != "" {
				st = state{"amended", decisionIDs(m[2])}
			} else if m[3] != "" {
				st = state{"superseded", []string{m[3]}}
			}
		}
		status[id] = st
		v, _ := decisionField(body, "Supersedes")
		supersedes[id] = decisionIDs(v)
		v, _ = decisionField(body, "Amends")
		amends[id] = decisionIDs(v)
		aff, _ := decisionField(body, "Affects")
		for _, q := range qualifiedRef.FindAllString(aff, -1) {
			pkg, rid, _ := strings.Cut(q, ".")
			if declared[pkg] != nil && !declared[pkg][rid] {
				c.report("decision-unknown-requirement", rel, id+" affects "+q+", which package "+pkg+" does not declare")
			}
		}
	}
	for _, id := range order {
		for _, link := range []struct {
			kind    string
			targets []string
		}{{"supersedes", supersedes[id]}, {"amends", amends[id]}} {
			for _, t := range link.targets {
				st, ok := status[t]
				if !ok {
					c.report("decision-unknown-ref", rel, id+" "+link.kind+" "+t+", which is not an entry here")
					continue
				}
				agrees := st.kind == "superseded" && contains(st.by, id)
				if link.kind == "amends" {
					agrees = (st.kind == "amended" && contains(st.by, id)) || st.kind == "superseded"
				}
				if !agrees {
					c.report("decision-status-mismatch", rel, id+" "+link.kind+" "+t+", but "+t+"'s Status does not point back")
				}
			}
		}
		st := status[id]
		for _, b := range st.by {
			if _, ok := status[b]; !ok {
				c.report("decision-unknown-ref", rel, id+"'s Status names "+b+", which is not an entry here")
			} else if (st.kind == "superseded" && !contains(supersedes[b], id)) || (st.kind == "amended" && !contains(amends[b], id)) {
				c.report("decision-status-mismatch", rel, id+"'s Status says "+st.kind+" by "+b+", but "+b+" does not say so")
			}
		}
	}
}

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
					cited = append(cited, goalRefs(f[1])...)
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
	c.standalonePackages()
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

	c.decisions(reqItems)
	c.references()
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
