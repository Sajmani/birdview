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

// Package view computes what the page shows: ordering, sectioning, and diagram
// geometry. It is deliberately pure -- no HTML, no HTTP -- so the rules in
// product.oneSelectedMetric through product.sectionHeadersStateRange and product.segmentComposition through product.crossbarAtVerticalScale can be property-tested directly.
//
// Verifies: product.oneSelectedMetric, product.ascendingOrder, product.sectionBreakRule, product.sectionThresholdDefault, product.sectionHeadersStateRange, product.scaleToggle, product.crossbarAtVerticalScale.
package view

import (
	"math"
	"sort"

	"birdview/internal/dataset"
)

// Metric selects what the page is ordered and sectioned by (product.oneSelectedMetric).
type Metric string

const (
	ByLength   Metric = "length"
	ByWingspan Metric = "wingspan"
	ByMass     Metric = "mass"
	// ByBill and ByTail order by the two measured segments (product.oneSelectedMetric).
	ByBill Metric = "bill"
	ByTail Metric = "tail"
	// ByFrequency orders by how often a species is recorded (product.orderByFrequency). Its
	// value is not a property of the species: it is a ratio over whatever
	// regions the reader has selected, so Value cannot compute it and
	// SectionsBy takes it as a key function instead.
	ByFrequency Metric = "frequency"
)

// DefaultStepPercent is N in product.sectionThresholdDefault: a new section starts at the first bird more
// than N% larger than the minimum of the previous section.
const DefaultStepPercent = 15.0

// Value returns the species' value for this metric, in millimeters or grams.
func (m Metric) Value(s *dataset.Species) float64 {
	switch m {
	case ByWingspan:
		return s.Wingspan.Value
	case ByMass:
		return s.Mass.Value
	case ByBill:
		return s.Segments.Bill.Value
	case ByTail:
		return s.Segments.Tail.Value
	default:
		return s.TotalLength.Value
	}
}

// Section is one run of similarly sized species.
type Section struct {
	Species  []*dataset.Species
	Min, Max float64 // in the selected metric (product.sectionHeadersStateRange)
}

// Sections orders species ascending by metric and breaks them into sections.
//
// The rule is stated against the **minimum of the previous section**, not
// against a running maximum or the immediately preceding bird. Those two read
// almost identically and behave differently: keying on the previous maximum
// lets a section drift arbitrarily wide, one bird at a time, which is exactly
// what a ratio rule is supposed to prevent.
//
// Verifies: product.ascendingOrder, product.sectionBreakRule, product.sectionThresholdDefault, product.sectionHeadersStateRange.
func Sections(species []*dataset.Species, m Metric, stepPercent float64) []Section {
	factor := 1 + stepPercent/100
	return SectionsBy(species, m.Value, func(min float64) float64 { return min * factor })
}

// SectionsBy is Sections with the ordering key and the break rule supplied.
//
// Two rules exist and they differ in what they are anchored to. product.sectionBreakRule's is
// anchored to the DATA -- the first section starts at the smallest bird
// present -- which suits a quantity that is a property of the bird. product.orderByFrequency's
// is anchored to powers of ten, because frequency is a ratio over the
// reader's region selection: a data-anchored band would denote one thing
// nationally and another in Alaska, and a reader comparing two views of the
// same site would be comparing two different tiers of the same name.
//
// next reports the first value that belongs to the section AFTER one whose
// minimum is min. Returning something not greater than min would loop, so it
// is asserted rather than trusted.
//
// Verifies: product.ascendingOrder, product.sectionBreakRule, product.sectionThresholdDefault, product.sectionHeadersStateRange, product.orderByFrequency.
func SectionsBy(species []*dataset.Species, key func(*dataset.Species) float64,
	next func(min float64) float64) []Section {

	if len(species) == 0 {
		return nil
	}
	sorted := append([]*dataset.Species(nil), species...)
	sort.SliceStable(sorted, func(i, j int) bool {
		return key(sorted[i]) < key(sorted[j])
	})

	var out []Section
	first := key(sorted[0])
	cur := Section{Species: []*dataset.Species{sorted[0]}, Min: first, Max: first}
	edge := next(first)

	for _, s := range sorted[1:] {
		v := key(s)
		// An edge of +Inf means there is no section after this one, so
		// everything remaining belongs here. Without the guard the
		// comparison is Inf >= Inf, which is true, and the species with no
		// frequency at all each start a section of their own -- 24 of them,
		// as this first did.
		if !math.IsInf(edge, 1) && v >= edge {
			out = append(out, cur)
			cur = Section{Species: []*dataset.Species{s}, Min: v, Max: v}
			edge = next(v)
			continue
		}
		cur.Species = append(cur.Species, s)
		cur.Max = v
	}
	return append(out, cur)
}

// Decade returns the first value in the band above the one holding v: the
// next power of ten (product.orderByFrequency).
//
// bands is how many decades a section spans -- 1 for "1 in 100 to 1 in 999",
// 2 for "1 in 100 to 1 in 9,999".
//
// +Inf has no next band and returns +Inf, which SectionsBy reads as "no
// section follows" -- that is what keeps the species with no records at all
// in one final section rather than one section each.
func Decade(v float64, bands int) float64 {
	if math.IsInf(v, 1) || v <= 0 {
		return math.Inf(1)
	}
	if bands < 1 {
		bands = 1
	}
	// Floor to the band this value sits in, then step past it. Computed on
	// the exponent rather than by repeated multiplication so that the
	// boundary is exactly a power of ten and not 999.9999999999999.
	k := int(math.Floor(math.Log10(v)))
	k -= ((k % bands) + bands) % bands // align downward to a band edge
	return math.Pow(10, float64(k+bands))
}

// ScaleMode is the user toggle of product.scaleToggle.
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

// Scales returns the reference length each section is drawn against: the
// largest total length in the section, or in the whole set when global.
//
// Returning a reference *length* rather than a pixels-per-mm factor keeps this
// independent of tile size, which the renderer owns.
//
// Verifies: product.scaleToggle.
func Scales(sections []Section, mode ScaleMode) []float64 {
	refs := make([]float64, len(sections))
	var overall float64
	for i, sec := range sections {
		var m float64
		for _, s := range sec.Species {
			if e := s.TotalLength.Value; e > m {
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
