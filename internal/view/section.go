// Package view computes what the page shows: ordering, sectioning and diagram
// geometry. It is deliberately pure -- no HTML, no HTTP -- so the rules in
// P-020 through P-024 and P-040 through P-043 can be property-tested directly.
//
// Verifies: P-020, P-021, P-022, P-023, P-024, P-042, P-043.
package view

import (
	"sort"

	"birdview/internal/dataset"
)

// Metric selects what the page is ordered and sectioned by (P-020).
type Metric string

const (
	ByLength   Metric = "length"
	ByWingspan Metric = "wingspan"
	ByMass     Metric = "mass"
)

// DefaultStepPercent is N in P-023: a new section starts at the first bird more
// than N% larger than the minimum of the previous section.
const DefaultStepPercent = 15.0

// Value returns the species' value for this metric, in millimeters or grams.
func (m Metric) Value(s *dataset.Species) float64 {
	switch m {
	case ByWingspan:
		return s.Wingspan.Value
	case ByMass:
		return s.Mass.Value
	default:
		return s.TotalLength.Value
	}
}

// Section is one run of similarly sized species.
type Section struct {
	Species  []*dataset.Species
	Min, Max float64 // in the selected metric (P-024)
}

// Sections orders species ascending by metric and breaks them into sections.
//
// The rule is stated against the **minimum of the previous section**, not
// against a running maximum or the immediately preceding bird. Those two read
// almost identically and behave differently: keying on the previous maximum
// lets a section drift arbitrarily wide, one bird at a time, which is exactly
// what a ratio rule is supposed to prevent.
//
// Verifies: P-021, P-022, P-023, P-024.
func Sections(species []*dataset.Species, m Metric, stepPercent float64) []Section {
	if len(species) == 0 {
		return nil
	}
	sorted := append([]*dataset.Species(nil), species...)
	sort.SliceStable(sorted, func(i, j int) bool {
		return m.Value(sorted[i]) < m.Value(sorted[j])
	})

	factor := 1 + stepPercent/100
	var out []Section
	cur := Section{Species: []*dataset.Species{sorted[0]}, Min: m.Value(sorted[0]), Max: m.Value(sorted[0])}

	for _, s := range sorted[1:] {
		v := m.Value(s)
		if v >= cur.Min*factor {
			out = append(out, cur)
			cur = Section{Species: []*dataset.Species{s}, Min: v, Max: v}
			continue
		}
		cur.Species = append(cur.Species, s)
		cur.Max = v
	}
	return append(out, cur)
}

// ScaleMode is the user toggle of P-042.
type ScaleMode string

const (
	// PerSection: every bird in a section shares a scale. The default, because
	// it keeps every tile legible while staying honest within the section --
	// a section spans at most N%.
	PerSection ScaleMode = "section"

	// Global: one scale across the whole filtered set, so a hummingbird really
	// is a speck beside a swan. Honest across the full range, at the cost of
	// legibility for small birds.
	Global ScaleMode = "global"
)

// VerticalExtent is how much vertical room a species needs in millimeters:
// normally its total length, but more when its feet reach below the tail tip.
//
// It is scale-free, which is what lets a section agree on one reference before
// any pixels exist. Using total length alone would clip the feet of the very
// bird that sets the scale.
func VerticalExtent(s *dataset.Species) float64 {
	seg := s.Segments
	toFeet := seg.Bill.Value + seg.Head.Value + seg.Neck.Value +
		seg.Body.Value/2 + s.Leg.Value*legSin
	if toFeet > s.TotalLength.Value {
		return toFeet
	}
	return s.TotalLength.Value
}

// Scales returns the reference extent each section is drawn against: the
// largest vertical extent in the section, or in the whole set when global.
//
// Returning a reference *length* rather than a pixels-per-mm factor keeps this
// independent of tile size, which the renderer owns.
//
// Verifies: P-042.
func Scales(sections []Section, mode ScaleMode) []float64 {
	refs := make([]float64, len(sections))
	var overall float64
	for i, sec := range sections {
		var m float64
		for _, s := range sec.Species {
			if e := VerticalExtent(s); e > m {
				m = e
			}
		}
		refs[i] = m
		if m > overall {
			overall = m
		}
	}
	if mode == Global {
		for i := range refs {
			refs[i] = overall
		}
	}
	return refs
}
