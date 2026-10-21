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
	"html"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// TestQuotationsAreVerbatim checks that every block quotation in a transcribed
// source's <pkg>.md appears word for word in a file that source
// vendors. A transcription is reviewable material, not ground truth, and the
// vendored copy is there so a quotation can be checked against it; this does
// the checking.
//
// Comparison ignores markup and whitespace, and treats curly and straight
// quotation marks alike, since none of those changes a word. An ellipsis, or a
// closing quotation mark followed by an opening one, splits a quotation into
// fragments, each checked on its own. The attribution line, starting with an em
// dash, is not part of the quotation.
//
// Verifies: tech.recordBeforeConsuming. Criterion: acceptance.quotationsAreVerbatim.
func TestQuotationsAreVerbatim(t *testing.T) {
	row := regexp.MustCompile("\\|\\s*`//([^`]+\\.html)`\\s*\\|")
	records, err := filepath.Glob(filepath.Join("..", "requirements", "imported", "*", "*", "PROVENANCE.md"))
	if err != nil {
		t.Fatal(err)
	}
	checked := 0
	for _, record := range records {
		dir := filepath.Dir(record)
		transcription := filepath.Join(dir, filepath.Base(dir)+".md")
		if _, err := os.Stat(transcription); err != nil {
			continue // a source that contributes data, not requirements
		}
		var corpus strings.Builder
		for _, m := range row.FindAllStringSubmatch(read(t, record), -1) {
			corpus.WriteString(htmlText(read(t, filepath.Join("..", "..", filepath.FromSlash(m[1])))))
			corpus.WriteString(" ")
		}
		if corpus.Len() == 0 {
			continue // quoted, not vendored, so there is nothing to check against
		}
		text := corpus.String()
		for _, q := range blockQuotes(read(t, transcription)) {
			for _, frag := range fragment.Split(q, -1) {
				frag = strings.TrimSpace(frag)
				if len(frag) < 4 {
					continue
				}
				checked++
				if !strings.Contains(text, frag) {
					short := frag
					if len(short) > 90 {
						short = short[:90] + "…"
					}
					t.Errorf("%s: quotation not found in the vendored copy: %q",
						filepath.Base(dir), short)
				}
			}
		}
	}
	if checked == 0 {
		t.Fatal("no quotations checked; the layout has changed and this check proves nothing")
	}
	t.Logf("%d quoted passages match their vendored sources", checked)
}

var (
	activeElement = regexp.MustCompile(`(?is)<(script|style)\b.*?</(script|style)>`)
	markupTag     = regexp.MustCompile(`(?s)<[^>]*>`)
	space         = regexp.MustCompile(`\s+`)
	emphasis      = regexp.MustCompile("\\*\\*|[*`]")
	beforePunct   = regexp.MustCompile(`\s+([,.;:!?)])`)
	fragment      = regexp.MustCompile(`…|\.\.\.|\[…\]|"\s+"`)
	quoteMarks    = strings.NewReplacer("“", `"`, "”", `"`, "‘", "'", "’", "'")
)

// htmlText is a document's visible text, normalized for comparison.
func htmlText(doc string) string {
	s := activeElement.ReplaceAllString(doc, " ")
	s = markupTag.ReplaceAllString(s, " ")
	return normalize(html.UnescapeString(s))
}

func normalize(s string) string {
	s = space.ReplaceAllString(quoteMarks.Replace(s), " ")
	// Markup removed as a space leaves one before the punctuation that
	// followed a link.
	return strings.TrimSpace(beforePunct.ReplaceAllString(s, "$1"))
}

// blockQuotes returns each Markdown block quotation's text, without its
// attribution line and without the quotation marks around it.
func blockQuotes(md string) []string {
	var out []string
	var cur []string
	flush := func() {
		if len(cur) > 0 {
			q := normalize(emphasis.ReplaceAllString(strings.Join(cur, " "), ""))
			out = append(out, strings.Trim(q, `" `))
		}
		cur = nil
	}
	for _, line := range strings.Split(md, "\n") {
		body, ok := strings.CutPrefix(line, ">")
		if !ok {
			flush()
			continue
		}
		body = strings.TrimSpace(body)
		if strings.HasPrefix(body, "—") {
			continue
		}
		cur = append(cur, body)
	}
	flush()
	return out
}
