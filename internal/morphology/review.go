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

package morphology

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// Proportion is one row of data/proportions-review.tsv: a genus's or family's neck ratio and its head,
// body, and wingspan factors, any of which may be empty to keep what the models give, and whether
// the owner has reviewed it (decisions.proportionsReviewedPerTaxon,
// decisions.allModeledProportionsReviewed).
type Proportion struct {
	Taxon      string
	Neck, Head float64 // 0 keeps the family table's ratio, or a factor of 1
	Body, Wing float64 // 0 is a factor of 1; Wing applies only to a modeled wingspan
	Status     string  // "proposed" or "reviewed"
	Reviewer   string
	Date       string
	Reason     string
}

// Review is the table by taxon. A nil Review is the code's family table alone.
type Review map[string]Proportion

// IsFamily reports whether a taxon names a family, which every family name ending in -idae does.
func IsFamily(taxon string) bool { return strings.HasSuffix(taxon, "idae") }

// reviewColumns is the table's header row.
const reviewColumns = "taxon\tneck ratio\thead factor\tbody factor\twingspan factor\tstatus\treviewer\tdate\treason"

var isoDate = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)

// ParseReview reads the table, refusing a row that cannot be applied: an unknown status, a reviewed
// row without reviewer or date, a proposed row without a reason, a value out of range, or a taxon
// given twice.
func ParseReview(r io.Reader) (Review, error) {
	rv := Review{}
	s := bufio.NewScanner(r)
	header := false
	for line := 1; s.Scan(); line++ {
		text := s.Text()
		if strings.HasPrefix(text, "#") || strings.TrimSpace(text) == "" {
			continue
		}
		if !header {
			if text != reviewColumns {
				return nil, fmt.Errorf("line %d: header %q, want %q", line, text, reviewColumns)
			}
			header = true
			continue
		}
		f := strings.Split(text, "\t")
		if len(f) != 9 {
			return nil, fmt.Errorf("line %d: %d columns, want 9", line, len(f))
		}
		p := Proportion{Taxon: f[0], Status: f[5], Reviewer: f[6], Date: f[7], Reason: f[8]}
		var err error
		if p.Neck, err = optional(f[1], 0.02, 2); err != nil {
			return nil, fmt.Errorf("line %d: neck ratio: %v", line, err)
		}
		if p.Head, err = optional(f[2], 0.3, 2); err != nil {
			return nil, fmt.Errorf("line %d: head factor: %v", line, err)
		}
		if p.Body, err = optional(f[3], 0.5, 2); err != nil {
			return nil, fmt.Errorf("line %d: body factor: %v", line, err)
		}
		if p.Wing, err = optional(f[4], 0.5, 2); err != nil {
			return nil, fmt.Errorf("line %d: wingspan factor: %v", line, err)
		}
		switch {
		case p.Taxon == "":
			return nil, fmt.Errorf("line %d: no taxon", line)
		case p.Status != "proposed" && p.Status != "reviewed":
			return nil, fmt.Errorf("line %d: status %q, want proposed or reviewed", line, p.Status)
		case p.Status == "reviewed" && (p.Reviewer == "" || !isoDate.MatchString(p.Date)):
			return nil, fmt.Errorf("line %d: a reviewed row needs a reviewer and a date", line)
		case p.Status == "proposed" && p.Reason == "":
			return nil, fmt.Errorf("line %d: a proposed row needs a reason", line)
		}
		if _, dup := rv[p.Taxon]; dup {
			return nil, fmt.Errorf("line %d: %s is given twice", line, p.Taxon)
		}
		rv[p.Taxon] = p
	}
	if !header {
		return nil, fmt.Errorf("no header row")
	}
	return rv, s.Err()
}

func optional(v string, lo, hi float64) (float64, error) {
	if v == "" {
		return 0, nil
	}
	x, err := strconv.ParseFloat(v, 64)
	if err != nil || x < lo || x > hi {
		return 0, fmt.Errorf("%q is not a number from %g to %g", v, lo, hi)
	}
	return x, nil
}

// Write writes the table under the given comment header, rows in taxon order.
func (rv Review) Write(w io.Writer, header string) error {
	if _, err := io.WriteString(w, header+reviewColumns+"\n"); err != nil {
		return err
	}
	taxa := make([]string, 0, len(rv))
	for t := range rv {
		taxa = append(taxa, t)
	}
	sort.Strings(taxa)
	num := func(x float64) string {
		if x == 0 {
			return ""
		}
		return strconv.FormatFloat(x, 'f', 2, 64)
	}
	for _, t := range taxa {
		p := rv[t]
		if _, err := fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n", p.Taxon, num(p.Neck), num(p.Head),
			num(p.Body), num(p.Wing), p.Status, p.Reviewer, p.Date, p.Reason); err != nil {
			return err
		}
	}
	return nil
}

// status is how a row's provenance ends: who reviewed it, or that no one has.
func (p Proportion) status() string {
	if p.Status == "reviewed" {
		return "reviewed by " + p.Reviewer + ", " + p.Date
	}
	return "proposed, not yet reviewed"
}

// NeckRatioFor is the neck ratio a bird of this family and genus draws with, and the words its
// provenance gives the source: a genus row, then a family row, then the code's family table.
func (rv Review) NeckRatioFor(family, genus string) (float64, string) {
	if p, ok := rv[genus]; ok && p.Neck > 0 {
		return p.Neck, fmt.Sprintf("ratio for the genus %s, in %s; %s", genus, family, p.status())
	}
	if p, ok := rv[family]; ok && p.Neck > 0 {
		return p.Neck, fmt.Sprintf("ratio for %s, from the review table; %s", family, p.status())
	}
	ratio, explicit := NeckRatio(family)
	if explicit {
		return ratio, "family-specific ratio for " + family
	}
	return ratio, "family default ratio for " + family
}

// HeadFactorFor is the factor on the modeled cranium, 1 when no row sets one, and the words its
// provenance gives the source.
func (rv Review) HeadFactorFor(family, genus string) (float64, string) {
	return rv.factor(family, genus, func(p Proportion) float64 { return p.Head })
}

// BodyFactorFor is the factor on the modeled trunk, 1 when no row sets one.
func (rv Review) BodyFactorFor(family, genus string) (float64, string) {
	return rv.factor(family, genus, func(p Proportion) float64 { return p.Body })
}

// WingspanFactorFor is the factor on a modeled wingspan, 1 when no row sets one. The caller applies
// it only to a modeled wingspan, never a measured one (decisions.allModeledProportionsReviewed).
func (rv Review) WingspanFactorFor(family, genus string) (float64, string) {
	return rv.factor(family, genus, func(p Proportion) float64 { return p.Wing })
}

func (rv Review) factor(family, genus string, of func(Proportion) float64) (float64, string) {
	if p, ok := rv[genus]; ok && of(p) > 0 {
		return of(p), fmt.Sprintf("the factor for the genus %s; %s", genus, p.status())
	}
	if p, ok := rv[family]; ok && of(p) > 0 {
		return of(p), fmt.Sprintf("the factor for %s; %s", family, p.status())
	}
	return 1, ""
}
