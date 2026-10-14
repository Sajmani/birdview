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

// Package region names the places a bird can be filtered to.
//
// Two kinds sit side by side. **Land** regions are GADM level-1 units — US
// states, Canadian provinces and the outlying islands — identified by GADM's
// own id and labeled with the standard postal abbreviation. **Marine** regions
// are BirdView's own, and exist because six species belong to no land region
// at all: Bermuda Petrel has 384 records offshore and zero inside any
// land polygon (P-011).
//
// Zones group regions, two ways (P-013). **Common birds** is derived by
// clustering regions on how much their species lists overlap; **Geographic**
// is drawn by hand so that a reader looking for Alaska finds it under Alaska.
// Each grouping is a partition, so a region appears exactly once in each and
// exactly twice overall.
//
// An earlier version of this file said there was no data-driven way to
// subdivide the east. That was an artifact of faceting GBIF on `speciesKey`,
// which drops what its backbone cannot resolve; redone against the verbatim
// names in data/region-presence.tsv, the east subdivides into four groups.
// See CR-023.
//
// Verifies: P-011, P-013.
package region

import (
	"bufio"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"
)

// Kind separates a real administrative region from one BirdView drew.
type Kind string

const (
	// Land is a GADM level-1 unit. Its boundary is somebody else's.
	Land Kind = "land"
	// Marine is BirdView's own approximation, defined by a coordinate box.
	// Boundaries are ours and are stated to be approximate.
	Marine Kind = "marine"
)

// The two groupings. Names are exported because the server and its tests name
// them, and a typo in a string literal should not silently produce an empty
// selector.
const (
	// ByBirds is the grouping derived from the presence table.
	ByBirds = "Common birds"
	// ByGeography is the hand-drawn grouping.
	ByGeography = "Geographic"
)

// Region is one filterable place.
type Region struct {
	// ID is GADM's own id for a land region ("USA.5_1"), or a SEA.* id for a
	// marine one. It keys the presence table and is the gadmGid sent to
	// GBIF. It is **not** what appears in a URL: it is a database key, and
	// Slug is the reader's handle (CR-024).
	ID string
	// Code is the standard postal abbreviation, or **empty** where no
	// standard exists. It is deliberately left empty for the islands and the
	// oceans: the obvious abbreviations are taken -- Midway cannot be MI,
	// Navassa cannot be NV, Palmyra cannot be PA, Wake cannot be WA -- and a
	// code invented to dodge a collision would sit beside 64 official ones
	// with nothing marking it as this project's guess (CR-024).
	Code string
	// Name is the full name, shown beside the code, or alone where there is
	// no code.
	Name string
	// Gloss clarifies a name that needs it -- which stretch of ocean a box
	// covers, in practice. It is shown in the chip's tooltip and never
	// enters the slug.
	Gloss string
	Kind  Kind
	// zones maps a grouping to this region's zone within it. Unexported so
	// that every region is built by this package and the partition invariant
	// has somewhere to hold.
	zones map[string]string
}

// Zone returns this region's zone within a grouping, or "" if the grouping is
// not one this package offers.
func (r Region) Zone(grouping string) string { return r.zones[grouping] }

// Slug is how this region appears in a URL: its postal abbreviation where it
// has one, otherwise its name in PascalCase -- "CA", "GulfOfMexico".
//
// **Derived, never stored.** A stored slug can outlive the name it was made
// from, leaving a URL that describes a region BirdView no longer has; deriving
// it means renaming a region renames its URL, which is the honest behavior
// even though it breaks the old link.
func (r Region) Slug() string {
	if r.Code != "" {
		return r.Code
	}
	var b strings.Builder
	for _, word := range strings.Fields(r.Name) {
		for i, ch := range word {
			if i == 0 {
				b.WriteString(strings.ToUpper(string(ch)))
				continue
			}
			b.WriteRune(ch)
		}
	}
	return b.String()
}

// Grouping is one way of carving the regions up. Each is a partition: every
// region sits in exactly one of its zones (P-013).
type Grouping struct {
	// Name is both the identifier and the UI label.
	Name string
	// Note is the one line the UI shows under the label. A derived grouping
	// and a hand-drawn one must not look alike, so each says which it is.
	Note string
	// Zones in the order they are offered.
	Zones []string
}

// Groupings in the order the selector offers them. The hand-drawn one comes
// first: a reader arrives looking for their own state, and finds it fastest
// on the map they already carry. The derived grouping is the more interesting
// answer, but it is the one you go looking for, not the one you need first.
//
// Within each, zones run roughly west to east and then outward, which is how
// a reader scans for their own state. **Pelagic is last** in both: it is the
// only group that is not a place anyone lives, and a reader looking for their
// state should not have to scroll past the open ocean to find it.
var Groupings = []Grouping{
	{
		Name: ByGeography,
		Note: "The usual map, for finding your state quickly. Drawn by hand, not measured.",
		Zones: []string{
			"Alaska & the North",
			"Pacific",
			"Mountain West",
			"Southwest",
			"Great Plains",
			"Midwest",
			"Northeast",
			"Southeast",
			"Outlying Islands",
			"Pelagic",
		},
	},
	{
		Name: ByBirds,
		Note: "Grouped by how much their bird lists actually overlap, measured from the records.",
		Zones: []string{
			"Hawaiian Islands",
			"Tropical Pacific Atolls",
			"Pacific Northwest",
			"California & the Great Basin",
			"Northern Rockies",
			"Southwest Borderlands",
			"Southern Plains",
			"Prairies & the Boreal",
			"The High Arctic",
			"Great Lakes",
			"Ohio & Mississippi Valleys",
			"Gulf Coast",
			"South Atlantic",
			// Both names say "Southern" or "Northern" on purpose. Connecticut
			// and Rhode Island cluster with the coastal plain, not with the
			// states they share a region with, and a label reading plain
			// "Mid-Atlantic" made that look like a mistake.
			//
			// It is not. Rhode Island's single closest region anywhere is
			// Delaware, ahead of Connecticut and New Hampshire. Connecticut
			// is the genuinely marginal one -- its three nearest neighbours
			// are all in New England -- but it still loses the group
			// comparison, because the northern group reaches Newfoundland
			// (0.800 from Connecticut) and Prince Edward Island (0.742) and
			// its mean is dragged down by them.
			//
			// Moving the two was measured and is worse: mean within-zone
			// cohesion over the pair falls from 0.8257 to 0.8239. Splitting
			// the northern group is worse still, leaving five regions nearer
			// some other group than their own. The names were the defect.
			"Mid-Atlantic & Southern New England",
			"Northern New England & the Maritimes",
			"Caribbean",
			"Pelagic",
		},
	},
}

// Marine boxes. **They are disjoint**, touching only at shared edges, so a
// record falls in at most one ocean and a species is assigned every ocean it
// actually occurs in — Sooty Shearwater gets all five, which is right for a
// circumpolar bird. Order is therefore not significant, and a test pins the
// disjointness rather than an ordering.
//
// The boundaries are BirdView's own and approximate, chosen to separate the
// birds rather than to reproduce any published limit. The Gulf box stops at
// 31°N and -80°W, which is roughly Florida's west coast; east of that is
// Atlantic. The gap between -115 and -98 is the continental interior and is
// deliberately no ocean at all.
var marineBoxes = []struct {
	ID             string
	MinLat, MaxLat float64
	MinLon, MaxLon float64
}{
	{"SEA.ARC", 66, 90, -180, -40},
	{"SEA.CPA", -10, 40, -180, -150},
	{"SEA.PAC", 10, 66, -150, -115},
	{"SEA.GULF", 18, 31, -98, -80},
	{"SEA.ATL", 10, 66, -80, -40},
}

// MarineBox returns the coordinate box for a marine region id.
func MarineBox(id string) (minLat, maxLat, minLon, maxLon float64, ok bool) {
	for _, b := range marineBoxes {
		if b.ID == id {
			return b.MinLat, b.MaxLat, b.MinLon, b.MaxLon, true
		}
	}
	return 0, 0, 0, 0, false
}

// MarineIDs lists the marine regions in test order.
func MarineIDs() []string {
	out := make([]string, 0, len(marineBoxes))
	for _, b := range marineBoxes {
		out = append(out, b.ID)
	}
	return out
}

// all is every region BirdView offers, keyed by id.
//
// The postal abbreviations are USPS and Canada Post standards, not this
// project's invention. The outlying islands have no postal code and take a
// short form of their name; the marine regions are named here because nobody
// else names them.
var all = map[string]Region{}

func add(id, code, name, gloss string, kind Kind, birds, geo string) {
	all[id] = Region{
		ID: id, Code: code, Name: name, Gloss: gloss, Kind: kind,
		zones: map[string]string{ByBirds: birds, ByGeography: geo},
	}
}

func init() {
	// Two zone columns: the derived grouping, then the hand-drawn one. The
	// `birds` column is transcribed from the clustering described in CR-023
	// and is re-derived by AC-087, so a hand edit here that the data does not
	// support fails a test.
	land := []struct{ id, code, name, birds, geo string }{
		// Alaska & the North (geographic). The clustering splits these four:
		// Alaska goes to the Pacific Northwest at 0.685, Yukon and the
		// Northwest Territories to the Prairies, and Nunavut matches nothing
		// above the floor. That split is why the geographic grouping exists.
		{"USA.2_1", "AK", "Alaska", "Pacific Northwest", "Alaska & the North"},
		{"CAN.13_1", "YT", "Yukon", "Prairies & the Boreal", "Alaska & the North"},
		{"CAN.6_1", "NT", "Northwest Territories", "Prairies & the Boreal", "Alaska & the North"},
		{"CAN.8_1", "NU", "Nunavut", "The High Arctic", "Alaska & the North"},
		// Pacific
		{"CAN.2_1", "BC", "British Columbia", "Pacific Northwest", "Pacific"},
		{"USA.48_1", "WA", "Washington", "Pacific Northwest", "Pacific"},
		{"USA.38_1", "OR", "Oregon", "Pacific Northwest", "Pacific"},
		{"USA.5_1", "CA", "California", "California & the Great Basin", "Pacific"},
		{"USA.12_1", "HI", "Hawaii", "Hawaiian Islands", "Pacific"},
		// Mountain West
		{"CAN.1_1", "AB", "Alberta", "Northern Rockies", "Mountain West"},
		{"USA.13_1", "ID", "Idaho", "Northern Rockies", "Mountain West"},
		{"USA.27_1", "MT", "Montana", "Northern Rockies", "Mountain West"},
		{"USA.51_1", "WY", "Wyoming", "Northern Rockies", "Mountain West"},
		{"USA.29_1", "NV", "Nevada", "California & the Great Basin", "Mountain West"},
		{"USA.45_1", "UT", "Utah", "California & the Great Basin", "Mountain West"},
		{"USA.6_1", "CO", "Colorado", "Southern Plains", "Mountain West"},
		// Southwest
		{"USA.3_1", "AZ", "Arizona", "Southwest Borderlands", "Southwest"},
		{"USA.32_1", "NM", "New Mexico", "Southwest Borderlands", "Southwest"},
		{"USA.44_1", "TX", "Texas", "Southwest Borderlands", "Southwest"},
		{"USA.37_1", "OK", "Oklahoma", "Southern Plains", "Southwest"},
		// Great Plains
		{"CAN.12_1", "SK", "Saskatchewan", "Prairies & the Boreal", "Great Plains"},
		{"CAN.3_1", "MB", "Manitoba", "Prairies & the Boreal", "Great Plains"},
		{"USA.35_1", "ND", "North Dakota", "Prairies & the Boreal", "Great Plains"},
		{"USA.42_1", "SD", "South Dakota", "Prairies & the Boreal", "Great Plains"},
		{"USA.28_1", "NE", "Nebraska", "Southern Plains", "Great Plains"},
		{"USA.17_1", "KS", "Kansas", "Southern Plains", "Great Plains"},
		// Midwest
		{"CAN.9_1", "ON", "Ontario", "Great Lakes", "Midwest"},
		{"USA.24_1", "MN", "Minnesota", "Great Lakes", "Midwest"},
		{"USA.50_1", "WI", "Wisconsin", "Great Lakes", "Midwest"},
		{"USA.23_1", "MI", "Michigan", "Great Lakes", "Midwest"},
		{"USA.16_1", "IA", "Iowa", "Ohio & Mississippi Valleys", "Midwest"},
		{"USA.26_1", "MO", "Missouri", "Ohio & Mississippi Valleys", "Midwest"},
		{"USA.14_1", "IL", "Illinois", "Ohio & Mississippi Valleys", "Midwest"},
		{"USA.15_1", "IN", "Indiana", "Ohio & Mississippi Valleys", "Midwest"},
		{"USA.36_1", "OH", "Ohio", "Ohio & Mississippi Valleys", "Midwest"},
		// Northeast
		{"CAN.11_1", "QC", "Québec", "Northern New England & the Maritimes", "Northeast"},
		{"CAN.4_1", "NB", "New Brunswick", "Northern New England & the Maritimes", "Northeast"},
		{"CAN.7_1", "NS", "Nova Scotia", "Northern New England & the Maritimes", "Northeast"},
		{"CAN.10_1", "PE", "Prince Edward Island", "Northern New England & the Maritimes", "Northeast"},
		{"CAN.5_1", "NL", "Newfoundland and Labrador", "Northern New England & the Maritimes", "Northeast"},
		{"USA.20_1", "ME", "Maine", "Northern New England & the Maritimes", "Northeast"},
		{"USA.30_1", "NH", "New Hampshire", "Northern New England & the Maritimes", "Northeast"},
		{"USA.46_1", "VT", "Vermont", "Northern New England & the Maritimes", "Northeast"},
		{"USA.33_1", "NY", "New York", "Northern New England & the Maritimes", "Northeast"},
		{"USA.22_1", "MA", "Massachusetts", "Northern New England & the Maritimes", "Northeast"},
		{"USA.40_1", "RI", "Rhode Island", "Mid-Atlantic & Southern New England", "Northeast"},
		{"USA.7_1", "CT", "Connecticut", "Mid-Atlantic & Southern New England", "Northeast"},
		{"USA.31_1", "NJ", "New Jersey", "Mid-Atlantic & Southern New England", "Northeast"},
		{"USA.39_1", "PA", "Pennsylvania", "Mid-Atlantic & Southern New England", "Northeast"},
		// Southeast
		{"USA.18_1", "KY", "Kentucky", "Ohio & Mississippi Valleys", "Southeast"},
		{"USA.43_1", "TN", "Tennessee", "Ohio & Mississippi Valleys", "Southeast"},
		{"USA.49_1", "WV", "West Virginia", "Ohio & Mississippi Valleys", "Southeast"},
		{"USA.47_1", "VA", "Virginia", "Mid-Atlantic & Southern New England", "Southeast"},
		{"USA.21_1", "MD", "Maryland", "Mid-Atlantic & Southern New England", "Southeast"},
		{"USA.8_1", "DE", "Delaware", "Mid-Atlantic & Southern New England", "Southeast"},
		{"USA.9_1", "DC", "District of Columbia", "Mid-Atlantic & Southern New England", "Southeast"},
		{"USA.34_1", "NC", "North Carolina", "South Atlantic", "Southeast"},
		{"USA.41_1", "SC", "South Carolina", "South Atlantic", "Southeast"},
		{"USA.11_1", "GA", "Georgia", "South Atlantic", "Southeast"},
		{"USA.10_1", "FL", "Florida", "South Atlantic", "Southeast"},
		{"USA.1_1", "AL", "Alabama", "Gulf Coast", "Southeast"},
		{"USA.25_1", "MS", "Mississippi", "Gulf Coast", "Southeast"},
		{"USA.4_1", "AR", "Arkansas", "Ohio & Mississippi Valleys", "Southeast"},
		{"USA.19_1", "LA", "Louisiana", "Gulf Coast", "Southeast"},
		// Outlying Islands -- no postal code exists, so the code is a short
		// name. Too small to place against the mainland (4 to 128 species,
		// against a mainland floor of 182), they are grouped against each
		// other: Midway's closest neighbour anywhere is Hawaii at 0.378, and
		// it is in the Hawaiian chain; Johnston, Palmyra and Wake match each
		// other at 0.44 to 0.55; Navassa shares zero species with any of them.
		{"UMI.4_1", "", "Johnston Atoll", "Tropical Pacific Atolls", "Outlying Islands"},
		{"UMI.6_1", "", "Midway Atoll", "Hawaiian Islands", "Outlying Islands"},
		{"UMI.7_1", "", "Navassa Island", "Caribbean", "Outlying Islands"},
		{"UMI.8_1", "", "Palmyra Atoll", "Tropical Pacific Atolls", "Outlying Islands"},
		{"UMI.9_1", "", "Wake Island", "Tropical Pacific Atolls", "Outlying Islands"},
	}
	for _, r := range land {
		add(r.id, r.code, r.name, "", Land, r.birds, r.geo)
	}

	// The oceans have no code either, and their names are kept short enough
	// to slug: "North Pacific, off the west coast" would have become
	// NorthPacificOffTheWestCoast, so the clarifier is a gloss instead.
	sea := []struct{ id, name, gloss string }{
		{"SEA.PAC", "North Pacific", "off the west coast"},
		{"SEA.CPA", "Central Pacific", "around Hawaii"},
		{"SEA.ATL", "North Atlantic", "off the east coast"},
		{"SEA.GULF", "Gulf of Mexico", ""},
		{"SEA.ARC", "Arctic Ocean", ""},
	}
	for _, r := range sea {
		// The open ocean is Pelagic under both groupings. Clustering the
		// marine regions would be clustering BirdView's own boxes.
		add(r.id, "", r.name, r.gloss, Marine, "Pelagic", "Pelagic")
	}

	// Index by slug once, from the same method the links use. Panics on a
	// collision rather than letting one region quietly shadow another and
	// become unreachable from every chip.
	for _, r := range all {
		if other, dup := bySlug[r.Slug()]; dup {
			panic(fmt.Sprintf("region: %s and %s share the slug %q",
				other.ID, r.ID, r.Slug()))
		}
		bySlug[r.Slug()] = r
	}
}

// BySlug resolves a URL token to a region. The lookup is built once from the
// same Slug method the links are rendered with, so a slug that renders always
// resolves.
func BySlug(slug string) (Region, bool) {
	r, ok := bySlug[slug]
	return r, ok
}

var bySlug = map[string]Region{}

// ZonesOf lists a grouping's zones in offer order, or nil if no such grouping
// exists.
func ZonesOf(grouping string) []string {
	for _, g := range Groupings {
		if g.Name == grouping {
			return g.Zones
		}
	}
	return nil
}

// Get returns a region by id.
func Get(id string) (Region, bool) {
	r, ok := all[id]
	return r, ok
}

// Len is how many regions exist.
func Len() int { return len(all) }

// InZone lists the regions of one zone within one grouping, by name.
func InZone(grouping, zone string) []Region {
	var out []Region
	for _, r := range all {
		if r.zones[grouping] == zone {
			out = append(out, r)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

// LandIDs lists every land region id, sorted, for a generator to iterate.
func LandIDs() []string {
	var out []string
	for id, r := range all {
		if r.Kind == Land {
			out = append(out, id)
		}
	}
	sort.Strings(out)
	return out
}

// Presence is the per-species region table built by cmd/birdview-regions:
// scientific name to the regions that species was reported in, with how many
// records put it there.
type Presence map[string][]Reported

// Reported is one region a species was reported in, and its record count
// there. The count is what lets a page say what share of the birds recorded
// in the regions on screen are this one (P-074).
type Reported struct {
	ID      string
	Records int64
	// Months splits Records twelve ways, January first. Carried beside the
	// annual figure rather than replacing it: every record has a month, so
	// the twelve sum to Records exactly, and a reader that does not care
	// about months should not have to add them up (P-011, CR-036).
	Months [12]int64
	// Withheld marks the months eBird keeps this species out of its public
	// output here, under its Sensitive Species list, confirmed against the
	// records (CR-037). In those months the count is missing by policy, so
	// it says nothing about whether the bird was there.
	Withheld [12]bool
	// Below marks a region where the species has records but stays under
	// the presence bar. Its records count in the species' figure; it is not
	// one of the species' regions (P-011, CR-042). Written as a ~ after the
	// region id.
	Below bool
}

// BelowMark follows a region id in the table when the species is under the
// presence bar there (CR-042).
const BelowMark = "~"

// Totals is every bird record a region holds, over every species including
// the ones BirdView does not draw. It is the denominator of that share: a
// species' records divided by all the birds seen in the same place by the
// same observers, which is what cancels out how hard anyone was looking.
type Totals map[string]int64

// MonthlyTotals is Totals split twelve ways, January first, and is the
// denominator of a figure scoped to months (P-011, CR-036). The twelve sum
// to the annual total exactly.
type MonthlyTotals map[string][12]int64

// LoadPresence reads the generated table. Comment lines carry the region
// names for a human reading the file and are not parsed back -- the names
// live in this package, and reading them from the data would let the two
// disagree without anything noticing.
func LoadPresence(r io.Reader) (Presence, Totals, MonthlyTotals, error) {
	out := Presence{}
	totals := Totals{}
	monthly := MonthlyTotals{}
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	line := 0
	for sc.Scan() {
		line++
		t := sc.Text()
		if t == "" || strings.HasPrefix(t, "#") {
			continue
		}
		name, rest, ok := strings.Cut(t, "\t")
		if !ok {
			return nil, nil, nil, fmt.Errorf("region presence line %d: no tab", line)
		}
		// A "total" row carries a region's own count, over every bird
		// recorded there. No species name collides: they are binomials.
		if name == "total" {
			gid, n, ok := strings.Cut(rest, "\t")
			if !ok {
				return nil, nil, nil, fmt.Errorf("region presence line %d: total row needs a region and a count", line)
			}
			ms, sum, err := parseMonths(strings.TrimSpace(n))
			if err != nil {
				return nil, nil, nil, fmt.Errorf("region presence line %d: %w", line, err)
			}
			if _, known := Get(gid); !known {
				return nil, nil, nil, fmt.Errorf("region presence line %d: %q is in no region table entry", line, gid)
			}
			totals[gid] = sum
			monthly[gid] = ms
			continue
		}
		if rest == "" {
			out[name] = nil
			continue
		}
		var rs []Reported
		for _, pair := range strings.Split(rest, ",") {
			id, n, ok := strings.Cut(pair, ":")
			if !ok {
				return nil, nil, nil, fmt.Errorf("region presence line %d: %q is not region:records", line, pair)
			}
			id, below := strings.CutSuffix(id, BelowMark)
			if _, known := Get(id); !known {
				return nil, nil, nil, fmt.Errorf("region presence line %d: %q is in no region table entry", line, id)
			}
			counts, flags, _ := strings.Cut(n, "!")
			ms, sum, err := parseMonths(counts)
			if err != nil {
				return nil, nil, nil, fmt.Errorf("region presence line %d: %w", line, err)
			}
			rep := Reported{ID: id, Records: sum, Months: ms, Below: below}
			if flags != "" {
				if len(flags) != 12 || strings.Trim(flags, "01") != "" {
					return nil, nil, nil, fmt.Errorf("region presence line %d: withheld flags %q "+
						"are not twelve 0s and 1s", line, flags)
				}
				for m, c := range flags {
					rep.Withheld[m] = c == '1'
				}
			}
			rs = append(rs, rep)
		}
		out[name] = rs
	}
	if err := sc.Err(); err != nil {
		return nil, nil, nil, err
	}
	if len(totals) == 0 {
		return nil, nil, nil, fmt.Errorf("region presence has no total rows; the share of a " +
			"region's records cannot be computed without its denominator")
	}
	return out, totals, monthly, nil
}

// parseMonths reads `jan|feb|...|dec` and returns the twelve and their sum.
//
// The annual figure is the sum and is never written down, so the two cannot
// drift: a table carrying both would eventually carry a year that is not
// what its months add to, and nothing would notice (P-011, CR-036).
func parseMonths(s string) (ms [12]int64, sum int64, err error) {
	parts := strings.Split(s, "|")
	if len(parts) != 12 {
		return ms, 0, fmt.Errorf("%q has %d monthly counts, want 12", s, len(parts))
	}
	for i, p := range parts {
		v, err := strconv.ParseInt(p, 10, 64)
		if err != nil {
			return ms, 0, fmt.Errorf("month %d of %q: %w", i+1, s, err)
		}
		if v < 0 {
			return ms, 0, fmt.Errorf("month %d of %q is negative", i+1, s)
		}
		ms[i] = v
		sum += v
	}
	return ms, sum, nil
}
