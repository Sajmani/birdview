// Package status derives BirdView's own occurrence status for a species from
// the dates on which it was recorded, rather than from how many records it
// has.
//
// The distinction is the whole point (CR-028). A record count measures
// observer effort: it under-reports pelagic species and remote-island
// endemics because nobody is there, and over-reports vagrants because a
// twitchable rarity draws a crowd. Those pull in opposite directions, so no
// threshold on a count corrects both. Recurrence is insensitive to both --
// Akikiki has 376 records spread over 45 years and is a resident, while the
// 2023 American Flamingo in Pennsylvania has 931 records inside 13 days and
// is not.
//
// What this package cannot do is tell where a bird came from. Indian Peafowl
// in New York recurs on every measure here -- 47 episodes across 33 years --
// because ornamental birds are replenished by people. Origin is declared, not
// measured; see P-071.
//
// Verifies: P-001, P-070.
package status

import "time"

// Status is BirdView's own occurrence status for a species in a region.
type Status string

const (
	// Established: the species keeps turning up, year after year.
	Established Status = "established"
	// Rare: it recurs, but thinly -- a handful of episodes across years.
	Rare Status = "rare"
	// Vagrant: it has turned up, but not repeatedly. One or two arrivals,
	// however many observers went to see them.
	Vagrant Status = "vagrant"
	// Absent: no open record at all.
	Absent Status = "absent"
)

// InScope reports whether a status puts a species in the dataset (P-001).
func (s Status) InScope() bool { return s == Established || s == Rare }

// Gap is the silence, in days, that separates one arrival from the next.
//
// Thirty days is declared, not fitted. It has to be long enough that a bird
// present for a season is one episode rather than thirty -- observers do not
// report daily, and a week's hole in a resident's record is ordinary -- and
// short enough that two arrivals in different seasons are not merged. The
// measured cases sit nowhere near it: the Pennsylvania flamingo spans 13 days
// and Lucy's Warbler in Virginia 7, while a resident's gaps are days.
const Gap = 30

// Shape is what the dates say about a species in one region.
type Shape struct {
	Records int64
	Dates   int
	// Episodes is the number of separate arrivals: runs of dates with no
	// silence longer than Gap between them.
	Episodes int
	Years    int
	SpanDays int
	// Truncated records that the date facet hit its page limit, so Dates,
	// Episodes and SpanDays are FLOORS and not counts. A species with more
	// distinct dates than a facet page is established by inspection, so this
	// is sufficient rather than a defect -- but it must never be read as a
	// measurement.
	Truncated bool
}

// Measure reduces the dates a species was recorded on in one region to the
// shape Classify reads. Dates need not be sorted and may repeat.
func Measure(records int64, dates []time.Time, truncated bool) Shape {
	s := Shape{Records: records, Truncated: truncated}
	if len(dates) == 0 {
		return s
	}
	// Copy before sorting: the caller's slice is not ours to reorder.
	ds := make([]time.Time, len(dates))
	copy(ds, dates)
	for i := 1; i < len(ds); i++ {
		for j := i; j > 0 && ds[j].Before(ds[j-1]); j-- {
			ds[j], ds[j-1] = ds[j-1], ds[j]
		}
	}
	years := map[int]bool{}
	uniq := ds[:0:0]
	for i, d := range ds {
		years[d.Year()] = true
		if i == 0 || !d.Equal(ds[i-1]) {
			uniq = append(uniq, d)
		}
	}
	s.Dates = len(uniq)
	s.Years = len(years)
	s.SpanDays = int(uniq[len(uniq)-1].Sub(uniq[0]).Hours()/24) + 1
	s.Episodes = 1
	for i := 1; i < len(uniq); i++ {
		if uniq[i].Sub(uniq[i-1]).Hours()/24 > Gap {
			s.Episodes++
		}
	}
	return s
}

// Thresholds separating the statuses. Declared coefficients, each with a
// reason; nothing here is fitted to reproduce anyone else's list, because the
// only available training signal would have been the ABA checklist and
// fitting to it would launder its editorial judgement into our weights
// (CR-028).
const (
	// MinEpisodes and MinYears are the bar for recurrence at all. Two
	// arrivals in two years is the least evidence that can distinguish
	// "comes back" from "turned up once"; below it the species is a vagrant
	// in this region however many records the arrival drew.
	MinEpisodes = 3
	MinYears    = 3

	// EstablishedEpisodes and EstablishedYears mark a bird that is simply
	// part of the region's avifauna rather than a scarce visitor to it.
	EstablishedEpisodes = 12
	EstablishedYears    = 8
)

// Classify turns a shape into a status.
func Classify(s Shape) Status {
	switch {
	case s.Dates == 0:
		return Absent
	// A truncated facet means more distinct dates than a page holds, which no
	// vagrant has. Decided before the thresholds so a floor is never compared
	// against a bar it cannot reach.
	case s.Truncated:
		return Established
	case s.Episodes >= EstablishedEpisodes && s.Years >= EstablishedYears:
		return Established
	case s.Episodes >= MinEpisodes && s.Years >= MinYears:
		return Rare
	default:
		return Vagrant
	}
}

// Best returns the strongest status among a species' regions, which is what
// decides scope: a bird resident anywhere in the ABA Area is in scope, even
// where it is a vagrant everywhere else. An American Flamingo is a vagrant in
// Pennsylvania and a resident in Florida, and belongs in the dataset.
func Best(ss []Status) Status {
	best := Absent
	rank := map[Status]int{Absent: 0, Vagrant: 1, Rare: 2, Established: 3}
	for _, s := range ss {
		if rank[s] > rank[best] {
			best = s
		}
	}
	return best
}
