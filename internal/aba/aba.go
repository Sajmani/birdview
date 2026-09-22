// Package aba parses the ABA Checklist.
//
// The file is all rights reserved and is never redistributed; it is fetched to
// fetched/ at build time and read from there (T-004).
//
// Verifies: P-001, P-004.
package aba

import (
	"encoding/csv"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
)

// MaxRarityCode is the highest ABA rarity code BirdView keeps. Codes 4-6 are
// casual, accidental, extinct or extirpated -- 320 species a birder will
// essentially never meet, whose presence among the common birds makes the size
// comparison less useful rather than more.
//
// Verifies: P-001.
const MaxRarityCode = 3

// Species is one row of the checklist. The rarity code is deliberately not
// carried: it is read to select codes 1-3 and then discarded, because it is
// ABA's editorial judgment rather than BirdView's data (P-001, AC-011).
type Species struct {
	CommonName     string
	ScientificName string
	// AltScientificName is the AOS/Clements alternative the file gives in
	// parentheses, where the two taxonomies disagree. Empty when they agree.
	AltScientificName string
	Family            string // scientific, e.g. Anatidae
	FamilyCommon      string // e.g. Ducks, Geese, and Swans
}

// familyHeader matches a section-header row: "Ducks, Geese, and Swans (Anatidae)".
// Family is not a per-species column; it appears only in these headers, so a
// forward fill recovers it (P-004).
var familyHeader = regexp.MustCompile(`^(.*?)\s*\(([A-Z][a-z]+)\)\s*$`)

var spaces = regexp.MustCompile(`\s+`)

// Normalize prepares a scientific name for joining: strip any parenthetical
// alternative, collapse internal whitespace, trim.
//
// Verifies: P-006.
func Normalize(name string) string {
	if i := strings.IndexByte(name, '('); i >= 0 {
		if j := strings.IndexByte(name[i:], ')'); j >= 0 {
			name = name[:i] + name[i+j+1:]
		}
	}
	return strings.TrimSpace(spaces.ReplaceAllString(name, " "))
}

// parenthetical returns the alternative binomial in parentheses, if any.
func parenthetical(name string) string {
	i := strings.IndexByte(name, '(')
	j := strings.IndexByte(name, ')')
	if i < 0 || j < i {
		return ""
	}
	alt := strings.TrimSpace(name[i+1 : j])
	if len(strings.Fields(alt)) != 2 {
		return "" // a note, not a binomial
	}
	return alt
}

// Parse reads the checklist and returns the species at or below MaxRarityCode.
//
// Verifies: P-001, P-004.
func Parse(r io.Reader) ([]Species, error) {
	cr := csv.NewReader(r)
	cr.FieldsPerRecord = -1 // tolerate the copyright and footnote rows
	rows, err := cr.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("aba.Parse: %w", err)
	}

	var out []Species
	var famSci, famCommon string
	for _, row := range rows {
		if len(row) < 6 {
			continue
		}
		first := strings.TrimSpace(row[0])
		code, codeErr := strconv.Atoi(strings.TrimSpace(row[5]))

		// A non-blank first column with no rarity code is either a family
		// header or the copyright/footnote row. Only the former matches.
		if first != "" && codeErr != nil {
			if m := familyHeader.FindStringSubmatch(first); m != nil {
				famCommon = strings.TrimRight(strings.TrimSpace(m[1]), ",")
				famSci = m[2]
			}
			continue
		}
		if first != "" || codeErr != nil {
			continue
		}
		if code > MaxRarityCode {
			continue
		}
		if famSci == "" {
			return nil, fmt.Errorf("aba.Parse: species row before any family header: %q", row[1])
		}
		out = append(out, Species{
			CommonName:        strings.TrimSpace(row[1]),
			ScientificName:    Normalize(row[3]),
			AltScientificName: parenthetical(row[3]),
			Family:            famSci,
			FamilyCommon:      famCommon,
		})
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("aba.Parse: no species found; is this the ABA checklist?")
	}
	return out, nil
}
