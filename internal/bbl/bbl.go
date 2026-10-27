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

// Package bbl reads the USGS Bird Banding Laboratory's species list, the
// source of BirdView's four-letter banding codes (imports.md `bbl`). The
// list is a U.S. Government work in the public domain; it is vendored as
// the page the Laboratory serves, third_party/bbl/speclist.html.
package bbl

import (
	"html"
	"regexp"
	"strings"
)

// Entry is one species row: the Laboratory's species number, its
// four-letter alpha code, and the English and scientific names it gives.
type Entry struct {
	Number     string
	Alpha      string
	Common     string
	Scientific string
}

var (
	row   = regexp.MustCompile(`(?s)<tr>(.*?)</tr>`)
	cell  = regexp.MustCompile(`(?s)<td[^>]*>(.*?)</td>`)
	tag   = regexp.MustCompile(`<[^>]+>`)
	space = regexp.MustCompile(`\s+`)
	alpha = regexp.MustCompile(`^[A-Z]{4}$`)
)

// Parse reads the species list's table. The columns are, in order: species
// number, alpha code, common name, band size, French name, scientific name,
// threatened or endangered, comments, and taxonomic order. Rows that are not
// one species are left out: an unidentified pair ("Unidentified Gray-cheeked/
// Bicknell's Thrush"), a hybrid, or a row with no four-letter code.
func Parse(page string) []Entry {
	var out []Entry
	for _, r := range row.FindAllStringSubmatch(page, -1) {
		var c []string
		for _, m := range cell.FindAllStringSubmatch(r[1], -1) {
			c = append(c, strings.TrimSpace(space.ReplaceAllString(html.UnescapeString(tag.ReplaceAllString(m[1], "")), " ")))
		}
		if len(c) < 9 || !alpha.MatchString(c[1]) {
			continue
		}
		common, sci := c[2], c[5]
		low := strings.ToLower(common)
		if strings.HasPrefix(low, "unidentified") || strings.Contains(common, "/") ||
			strings.Contains(low, "hybrid") || strings.Contains(sci, " x ") || sci == "" {
			continue
		}
		out = append(out, Entry{Number: c[0], Alpha: c[1], Common: common, Scientific: sci})
	}
	return out
}
