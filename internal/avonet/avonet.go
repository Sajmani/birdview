// Package avonet reads morphometric traits from the AVONET workbook.
//
// AVONET ships three parallel taxonomies. BirdView's spine is AVONET2_eBird,
// but several ABA species that sheet lumps have their own row in
// AVONET1_BirdLife or AVONET3_BirdTree, and reading across all three is what
// CR-006 decided to do.
//
// Verifies: P-002, P-003, P-006.
package avonet

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/xuri/excelize/v2"
)

// Sheet names, used verbatim as the Source in a provenance record so a reader
// can tell which taxonomy a value came from.
const (
	SheetEBird    = "AVONET2_eBird"
	SheetBirdLife = "AVONET1_BirdLife"
	SheetBirdTree = "AVONET3_BirdTree"
	SheetCross    = "BirdLife-eBird crosswalk"
)

// Traits are the five AVONET measurements BirdView consumes. Other columns --
// Beak.Width, Beak.Depth, Hand-Wing.Index -- are deliberately not read: carrying
// unused data invites someone to invent a use for it (P-003).
//
// A column that stops earning its place leaves. Beak.Depth was read until the
// cranium model moved from bill depth to mass. Tarsus.Length made the round
// trip -- unused, then ingested when the diagram gained legs, then dropped
// again when the flight premise removed them (CR-010).
type Traits struct {
	Culmen    float64 // Beak.Length_Culmen, mm
	Tail      float64 // Tail.Length, mm
	WingChord float64 // Wing.Length, mm -- folded wing, NOT wingspan
	Secondary float64 // Secondary1, mm
	Mass      float64 // grams

	// Habitat and Niche are AVONET's own categorical classifiers, taken
	// verbatim. They are **primary** values: AVONET records the one habitat
	// and the one feeding niche a species mostly uses, not every one it
	// uses. Anything BirdView adds beyond them is its own and is marked as
	// such (P-069).
	Habitat string // Habitat: Forest, Wetland, Marine, Grassland, ...
	Niche   string // Trophic.Niche: Invertivore, Aquatic predator, ...
}

// Workbook holds the trait rows of each taxonomy, keyed by that taxonomy's
// binomial, plus the BirdLife->eBird name crosswalk.
type Workbook struct {
	EBird    map[string]Traits
	BirdLife map[string]Traits
	BirdTree map[string]Traits

	// FamilyOf gives the eBird-sheet family for a species, used only as a
	// cross-check against ABA's family; ABA is authoritative (P-004).
	FamilyOf map[string]string

	// OrderOf gives the taxonomic order, which the wingspan model needs: the
	// regression is fitted per order, not per family.
	OrderOf map[string]string

	// EBirdToBirdLife maps a Species2 name to the Species1 names AVONET pairs
	// it with. Needed because the Fox wingspan data is keyed on BirdLife names.
	EBirdToBirdLife map[string][]string
}

// Open reads the sheets BirdView needs. The workbook is 21 MB, so each sheet is
// streamed rather than materialized.
func Open(path string) (*Workbook, error) {
	f, err := excelize.OpenFile(path)
	if err != nil {
		return nil, fmt.Errorf("avonet.Open %s: %w", path, err)
	}
	defer f.Close()

	wb := &Workbook{
		EBird:           map[string]Traits{},
		BirdLife:        map[string]Traits{},
		BirdTree:        map[string]Traits{},
		FamilyOf:        map[string]string{},
		OrderOf:         map[string]string{},
		EBirdToBirdLife: map[string][]string{},
	}

	if err := wb.readTraits(f, SheetEBird, "Species2", "Family2", "Order2", wb.EBird, wb.FamilyOf, wb.OrderOf); err != nil {
		return nil, err
	}
	if err := wb.readTraits(f, SheetBirdLife, "Species1", "Family1", "Order1", wb.BirdLife, nil, wb.OrderOf); err != nil {
		return nil, err
	}
	if err := wb.readTraits(f, SheetBirdTree, "Species3", "Family3", "Order3", wb.BirdTree, nil, wb.OrderOf); err != nil {
		return nil, err
	}
	if err := wb.readCrosswalk(f); err != nil {
		return nil, err
	}
	return wb, nil
}

func (wb *Workbook) readTraits(f *excelize.File, sheet, spCol, famCol, ordCol string, into map[string]Traits, fams, ords map[string]string) error {
	rows, err := f.Rows(sheet)
	if err != nil {
		return fmt.Errorf("avonet: sheet %s: %w", sheet, err)
	}
	defer rows.Close()

	var idx map[string]int
	for rows.Next() {
		cells, err := rows.Columns()
		if err != nil {
			return fmt.Errorf("avonet: sheet %s: %w", sheet, err)
		}
		if idx == nil {
			idx = headerIndex(cells)
			for _, want := range []string{spCol, "Beak.Length_Culmen", "Tail.Length", "Wing.Length", "Secondary1", "Mass"} {
				if _, ok := idx[want]; !ok {
					return fmt.Errorf("avonet: sheet %s is missing required column %q", sheet, want)
				}
			}
			continue
		}
		name := strings.TrimSpace(at(cells, idx[spCol]))
		if name == "" {
			continue
		}
		// A categorical column is optional: AVONET's raw-data sheet has
		// none, and a missing one must leave the field empty rather than
		// fail the read.
		cat := func(col string) string {
			i, ok := idx[col]
			if !ok {
				return ""
			}
			v := strings.TrimSpace(at(cells, i))
			if v == "NA" {
				return ""
			}
			return v
		}
		into[name] = Traits{
			Culmen:    num(at(cells, idx["Beak.Length_Culmen"])),
			Tail:      num(at(cells, idx["Tail.Length"])),
			WingChord: num(at(cells, idx["Wing.Length"])),
			Secondary: num(at(cells, idx["Secondary1"])),
			Mass:      num(at(cells, idx["Mass"])),
			Habitat:   cat("Habitat"),
			Niche:     cat("Trophic.Niche"),
		}
		if fams != nil {
			if i, ok := idx[famCol]; ok {
				fams[name] = strings.TrimSpace(at(cells, i))
			}
		}
		if ords != nil {
			if i, ok := idx[ordCol]; ok {
				if v := strings.TrimSpace(at(cells, i)); v != "" {
					ords[name] = v
				}
			}
		}
	}
	return nil
}

func (wb *Workbook) readCrosswalk(f *excelize.File) error {
	rows, err := f.Rows(SheetCross)
	if err != nil {
		return fmt.Errorf("avonet: sheet %s: %w", SheetCross, err)
	}
	defer rows.Close()
	var idx map[string]int
	for rows.Next() {
		cells, err := rows.Columns()
		if err != nil {
			return err
		}
		if idx == nil {
			idx = headerIndex(cells)
			continue
		}
		s1 := strings.TrimSpace(at(cells, idx["Species1"]))
		s2 := strings.TrimSpace(at(cells, idx["Species2"]))
		if s1 != "" && s2 != "" {
			wb.EBirdToBirdLife[s2] = append(wb.EBirdToBirdLife[s2], s1)
		}
	}
	return nil
}

func headerIndex(cells []string) map[string]int {
	m := make(map[string]int, len(cells))
	for i, c := range cells {
		m[strings.TrimSpace(c)] = i
	}
	return m
}

func at(cells []string, i int) string {
	if i < 0 || i >= len(cells) {
		return ""
	}
	return cells[i]
}

func num(s string) float64 {
	v, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
	if err != nil || v <= 0 {
		return 0
	}
	return v
}
