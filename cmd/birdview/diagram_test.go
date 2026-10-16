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

package main

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"strings"
	"testing"

	"birdview/docs"
)

// TestDiagramsAreCurrent fails when a page's Mermaid source and the picture
// BirdView serves for it disagree. Edit the Markdown, run
// docs/render-diagrams.py, commit both.
//
// Verifies: product.pipelineDiagram. Criterion: acceptance.pipelineDiagramServed.
func TestDiagramsAreCurrent(t *testing.T) {
	entries, err := docs.Files.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	diagrams := 0
	for _, e := range entries {
		name, ok := strings.CutSuffix(e.Name(), ".md")
		if !ok {
			continue
		}
		md, _ := docs.Files.ReadFile(e.Name())
		m := regexp.MustCompile("(?s)```mermaid\n(.*?)\n```").FindSubmatch(md)
		if m == nil {
			continue
		}
		diagrams++
		svg, err := docs.Files.ReadFile(name + ".svg")
		if err != nil {
			t.Errorf("docs/%s.md has a diagram and no docs/%s.svg; run docs/render-diagrams.py", name, name)
			continue
		}
		sum := sha256.Sum256(m[1])
		want := "<!-- source sha256 " + hex.EncodeToString(sum[:]) + " -->\n"
		if !strings.HasPrefix(string(svg), want) {
			first, _, _ := strings.Cut(string(svg), "\n")
			t.Errorf("docs/%s.svg was drawn from a different source than docs/%s.md holds "+
				"(%s, want %s); run docs/render-diagrams.py", name, name, first, strings.TrimSpace(want))
		}
	}
	if diagrams == 0 {
		t.Error("no page in docs/ has a diagram; product.pipelineDiagram asks for the pipeline's")
	}
}

// TestDiagramsAreSelfContained keeps the served pictures inert and local:
// opened full size an SVG is a document of its own, and one that ran a
// script or fetched a font would break tech.selfContainedServer where no test of the pages
// would look.
//
// Verifies: product.pipelineDiagram, tech.selfContainedServer. Criterion: acceptance.pipelineDiagramServed.
func TestDiagramsAreSelfContained(t *testing.T) {
	entries, _ := docs.Files.ReadDir(".")
	checked := 0
	for _, e := range entries {
		if !strings.HasSuffix(e.Name(), ".svg") {
			continue
		}
		checked++
		svg, _ := docs.Files.ReadFile(e.Name())
		s := string(svg)
		for _, bad := range []string{"<script", "<foreignObject", "@import"} {
			if strings.Contains(s, bad) {
				t.Errorf("docs/%s carries %s", e.Name(), bad)
			}
		}
		if m := regexp.MustCompile(`\son[a-z]+\s*=`).FindString(s); m != "" {
			t.Errorf("docs/%s carries an event handler:%s", e.Name(), m)
		}
		for _, u := range regexp.MustCompile(`(?:https?:)?//[^"'\s)<>]+`).FindAllString(s, -1) {
			if !strings.HasPrefix(u, "http://www.w3.org/") {
				t.Errorf("docs/%s refers off the server, to %s", e.Name(), u)
			}
		}
		for _, u := range regexp.MustCompile(`(?:url\(|href=")([^)"]*)`).FindAllStringSubmatch(s, -1) {
			if !strings.HasPrefix(u[1], "#") {
				t.Errorf("docs/%s refers to %q, outside itself", e.Name(), u[1])
			}
		}
	}
	if checked == 0 {
		t.Error("no SVG is embedded in docs; the pipeline diagram is not being served")
	}
}

// TestPipelineDiagramIsServed checks the page shows the picture rather than
// the Mermaid source, and the picture answers as one.
//
// Verifies: product.pipelineDiagram, product.categoryProvenance. Criterion: acceptance.pipelineDiagramServed.
func TestPipelineDiagramIsServed(t *testing.T) {
	s := newTestServer(t)
	page := mustGet(t, s, "/docs/pipeline")
	if !strings.Contains(page, `<a href="/docs/pipeline.svg"><img src="/docs/pipeline.svg"`) {
		t.Error("/docs/pipeline does not show the diagram, linked to itself full size")
	}
	if strings.Contains(page, "flowchart TB") {
		t.Error("/docs/pipeline shows the Mermaid source instead of the picture")
	}

	w := httptest.NewRecorder()
	s.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/docs/pipeline.svg", nil))
	if w.Code != http.StatusOK || w.Header().Get("Content-Type") != "image/svg+xml" ||
		!strings.Contains(w.Body.String(), "<svg") {
		t.Errorf("/docs/pipeline.svg answers %d as %q; want the picture as image/svg+xml",
			w.Code, w.Header().Get("Content-Type"))
	}
	for _, path := range []string{"/docs/pipeline.md", "/docs/.svg", "/docs/x/pipeline.svg",
		"/docs/../README.svg", "/spec/product.svg"} {
		if code, _ := get(t, s, path); code == http.StatusOK {
			t.Errorf("%s is served; only a page's own diagram should be", path)
		}
	}

	index := mustGet(t, s, "/docs/")
	if !strings.Contains(index, `href="/docs/pipeline"`) {
		t.Error("the docs index does not list the pipeline page")
	}
	if strings.Contains(index, "pipeline.svg") {
		t.Error("the docs index lists the picture as a document of its own")
	}
}

// TestPipelineDiagramNamesRealColumns checks every source column the diagram
// names is one the reader code asks for, by its quoted name. A column
// renamed or dropped in the code would otherwise leave the diagram
// describing a pipeline that no longer exists.
//
// Verifies: product.pipelineDiagram. Criterion: acceptance.pipelineDiagramServed.
func TestPipelineDiagramNamesRealColumns(t *testing.T) {
	md, err := docs.Files.ReadFile("pipeline.md")
	if err != nil {
		t.Fatal(err)
	}
	for _, src := range []struct{ box, reader string }{
		{"AVL", "../../internal/avilist/avilist.go"},
		{"AVO", "../../internal/avonet/avonet.go"},
		{"FOX", "../../internal/wingspan/wingspan.go"},
	} {
		code, err := os.ReadFile(src.reader)
		if err != nil {
			t.Fatal(err)
		}
		m := regexp.MustCompile("(?s)\n\\s*" + src.box + "\\[\"`[^`]*?─+\n(.*?)`\"\\]").FindSubmatch(md)
		if m == nil {
			t.Errorf("no %s box with a field list in docs/pipeline.md", src.box)
			continue
		}
		var cols []string
		for _, line := range strings.Split(string(m[1]), "\n") {
			// The column name comes before the mark; one per line, which
			// TestPipelineFieldsCarryOneMark enforces.
			c, _, _ := strings.Cut(strings.TrimSpace(line), " ")
			if base, ok := strings.CutSuffix(c, "1-3"); ok {
				cols = append(cols, base+"1", base+"2", base+"3")
			} else if c != "" {
				cols = append(cols, c)
			}
		}
		if len(cols) < 5 {
			t.Errorf("%s box: only %d columns parsed (%v); the field list was not read", src.box, len(cols), cols)
		}
		for _, c := range cols {
			if !strings.Contains(string(code), `"`+c+`"`) {
				t.Errorf("docs/pipeline.md names %s column %q, which %s never asks for",
					src.box, c, src.reader)
			}
		}
	}
}

// TestPipelineDiagramNamesEveryTaxonomyTable checks that each table in
// internal/taxonomy the ingest reads appears in the diagram, so a new table
// steering the build cannot go undrawn (product.pipelineDiagram).
//
// Verifies: product.pipelineDiagram. Criterion: acceptance.pipelineDiagramServed.
func TestPipelineDiagramNamesEveryTaxonomyTable(t *testing.T) {
	md, err := docs.Files.ReadFile("pipeline.md")
	if err != nil {
		t.Fatal(err)
	}
	code, err := os.ReadFile("../../cmd/birdview-ingest/main.go")
	if err != nil {
		t.Fatal(err)
	}
	// Not data, so not drawn. Each says why.
	notData := map[string]string{
		"AviListRenamedFrom":     "the inverse of AviListRenames, which the diagram names",
		"AwaitingTaxonomicCheck": "a list printed in the build report",
	}
	seen := map[string]bool{}
	for _, m := range regexp.MustCompile(`taxonomy\.([A-Z]\w*)`).FindAllStringSubmatch(string(code), -1) {
		seen[m[1]] = true
	}
	if len(seen) < 5 {
		t.Fatalf("only %d taxonomy references found in the ingest (%v); the scan read nothing", len(seen), seen)
	}
	for name := range seen {
		if _, ok := notData[name]; ok {
			continue
		}
		if !regexp.MustCompile(`\b` + name + `\b`).Match(md) {
			t.Errorf("the ingest reads taxonomy.%s, which docs/pipeline.md does not name", name)
		}
	}
}

// TestPipelineDiagramIsLinked checks the diagram can be reached from where
// each audience reads about the data, not only from the index: the guide
// for readers, the guide for developers, the using guide's word on where
// the numbers come from, the site's source list, and the README.
//
// Verifies: product.pipelineDiagram. Criterion: acceptance.pipelineDiagramServed.
func TestPipelineDiagramIsLinked(t *testing.T) {
	s := newTestServer(t)
	for _, path := range []string{"/docs/data", "/docs/building", "/docs/using", "/"} {
		if !strings.Contains(mustGet(t, s, path), `href="/docs/pipeline"`) {
			t.Errorf("%s does not link to the pipeline diagram", path)
		}
	}
	readme, err := os.ReadFile("../../README.md")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(readme), "(docs/pipeline.md)") {
		t.Error("README.md does not link to docs/pipeline.md")
	}
}

// TestPipelineFieldsCarryOneMark holds every box in the diagram to the
// format its legend states: one field per line, then exactly one mark, then
// an optional note in italics. The first draft used · both as a mark and as
// a separator between fields on one line, so "species · region · date
// window" read as one field with two marks, and the legend's "one per
// line" was true of some boxes and not others.
//
// Lines under a box's name and above its rule say where it lives, and are
// italic too.
//
// Verifies: product.pipelineDiagram. Criterion: acceptance.pipelineDiagramServed.
func TestPipelineFieldsCarryOneMark(t *testing.T) {
	md, err := docs.Files.ReadFile("pipeline.md")
	if err != nil {
		t.Fatal(err)
	}
	if m := regexp.MustCompile("(?m)^\\s*\\w+\\[\"[^`\"][^\"]*─").Find(md); m != nil {
		t.Errorf("a box with fields is a plain label, so its notes cannot be italic: %q", m)
	}
	field := regexp.MustCompile(`^([^★·✗*]+?) ([★·✗])(?: \*([^*]+)\*)?$`)
	italic := regexp.MustCompile(`^\*[^*]+\*$`)
	boxes := 0
	for _, b := range regexp.MustCompile("(?s)(\\w+)\\[\"`(.*?)`\"\\]").FindAllSubmatch(md, -1) {
		name, label := string(b[1]), string(b[2])
		head, fields, ok := strings.Cut(label, "─────────────\n")
		if !ok {
			continue // the legend's key to the marks, which lists marks, not fields
		}
		boxes++
		heads := strings.Split(strings.TrimSpace(head), "\n")
		for _, line := range heads[1:] {
			if line = strings.TrimSpace(line); !italic.MatchString(line) {
				t.Errorf("%s: %q under the box's name is not italic", name, line)
			}
		}
		for _, line := range strings.Split(fields, "\n") {
			line = strings.TrimSpace(line)
			if !field.MatchString(line) {
				t.Errorf("%s: %q is not one field, one mark and an optional italic note", name, line)
			}
		}
	}
	if boxes < 15 {
		t.Errorf("only %d boxes with fields found; the diagram was not read", boxes)
	}
}

// TestPipelineTablesAreLeftAligned checks the rendered picture has every
// table's lines left-aligned, and the legend's key to the marks. Mermaid
// centers every label, so
// render-diagrams.py moves them; a render that skipped that step would
// still match its source's hash, and nothing else would notice.
//
// Verifies: product.pipelineDiagram. Criterion: acceptance.pipelineDiagramServed.
func TestPipelineTablesAreLeftAligned(t *testing.T) {
	md, _ := docs.Files.ReadFile("pipeline.md")
	svg, _ := docs.Files.ReadFile("pipeline.svg")
	tables := 0
	for _, b := range regexp.MustCompile("(?s)\\w+\\[\"`(.*?)`\"\\]").FindAllSubmatch(md, -1) {
		if regexp.MustCompile("─────|[★·✗]").Match(b[1]) {
			tables++
		}
	}
	if got := strings.Count(string(svg), `style="text-anchor: start"`); tables == 0 || got != tables {
		t.Errorf("docs/pipeline.svg left-aligns %d boxes of the %d tables and keys in docs/pipeline.md; "+
			"run docs/render-diagrams.py", got, tables)
	}
}
