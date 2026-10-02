// Package region names the places a bird can be filtered to.
//
// Two kinds sit side by side. **Land** regions are GADM level-1 units — US
// states, Canadian provinces and the outlying islands — identified by GADM's
// own id and labeled with the standard postal abbreviation. **Marine** regions
// are BirdView's own, and exist because six species belong to no land region
// at all: Bermuda Petrel has 384 records in the ABA Area and zero inside any
// land polygon (P-011).
//
// Zones group regions for navigation and claim nothing about birds. Clustering
// all 64 land regions by how much their species lists overlap produces one
// 42-member block covering the whole east — Maine and Georgia share 0.75 of
// their species — so there is no data-driven way to subdivide it. Any
// subdivision is for finding your state quickly, and this package says so
// rather than dressing it up.
//
// Verifies: P-011.
package region

import "sort"

// Kind separates a real administrative region from one BirdView drew.
type Kind string

const (
	// Land is a GADM level-1 unit. Its boundary is somebody else's.
	Land Kind = "land"
	// Marine is BirdView's own approximation, defined by a coordinate box.
	// Boundaries are ours and are stated to be approximate.
	Marine Kind = "marine"
)

// Region is one filterable place.
type Region struct {
	// ID is GADM's own id for a land region ("USA.5_1"), or a SEA.* id for a
	// marine one. It is what appears in a URL and in the presence table.
	ID string
	// Code is the short label on a chip: the standard postal abbreviation
	// where one exists, a short name where none does.
	Code string
	// Name is the full name, shown beside the code.
	Name string
	Kind Kind
	// Zone is the navigation group this sits in.
	Zone string
}

// Zones in the order they are offered. Roughly west to east, then outward,
// which is how a reader scans for their own state.
var Zones = []string{
	"Alaska & the North",
	"Pacific",
	"Mountain West",
	"Southwest",
	"Great Plains",
	"Midwest",
	"Northeast",
	"Southeast",
	"Outlying Islands",
	"At Sea",
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

func add(id, code, name string, kind Kind, zone string) {
	all[id] = Region{ID: id, Code: code, Name: name, Kind: kind, Zone: zone}
}

func init() {
	land := []struct{ id, code, name, zone string }{
		// Alaska & the North
		{"USA.2_1", "AK", "Alaska", "Alaska & the North"},
		{"CAN.13_1", "YT", "Yukon", "Alaska & the North"},
		{"CAN.6_1", "NT", "Northwest Territories", "Alaska & the North"},
		{"CAN.8_1", "NU", "Nunavut", "Alaska & the North"},
		// Pacific
		{"CAN.2_1", "BC", "British Columbia", "Pacific"},
		{"USA.48_1", "WA", "Washington", "Pacific"},
		{"USA.38_1", "OR", "Oregon", "Pacific"},
		{"USA.5_1", "CA", "California", "Pacific"},
		{"USA.12_1", "HI", "Hawaii", "Pacific"},
		// Mountain West
		{"CAN.1_1", "AB", "Alberta", "Mountain West"},
		{"USA.13_1", "ID", "Idaho", "Mountain West"},
		{"USA.27_1", "MT", "Montana", "Mountain West"},
		{"USA.51_1", "WY", "Wyoming", "Mountain West"},
		{"USA.29_1", "NV", "Nevada", "Mountain West"},
		{"USA.45_1", "UT", "Utah", "Mountain West"},
		{"USA.6_1", "CO", "Colorado", "Mountain West"},
		// Southwest
		{"USA.3_1", "AZ", "Arizona", "Southwest"},
		{"USA.32_1", "NM", "New Mexico", "Southwest"},
		{"USA.44_1", "TX", "Texas", "Southwest"},
		{"USA.37_1", "OK", "Oklahoma", "Southwest"},
		// Great Plains
		{"CAN.12_1", "SK", "Saskatchewan", "Great Plains"},
		{"CAN.3_1", "MB", "Manitoba", "Great Plains"},
		{"USA.35_1", "ND", "North Dakota", "Great Plains"},
		{"USA.42_1", "SD", "South Dakota", "Great Plains"},
		{"USA.28_1", "NE", "Nebraska", "Great Plains"},
		{"USA.17_1", "KS", "Kansas", "Great Plains"},
		// Midwest
		{"CAN.9_1", "ON", "Ontario", "Midwest"},
		{"USA.24_1", "MN", "Minnesota", "Midwest"},
		{"USA.50_1", "WI", "Wisconsin", "Midwest"},
		{"USA.23_1", "MI", "Michigan", "Midwest"},
		{"USA.16_1", "IA", "Iowa", "Midwest"},
		{"USA.26_1", "MO", "Missouri", "Midwest"},
		{"USA.14_1", "IL", "Illinois", "Midwest"},
		{"USA.15_1", "IN", "Indiana", "Midwest"},
		{"USA.36_1", "OH", "Ohio", "Midwest"},
		// Northeast
		{"CAN.11_1", "QC", "Québec", "Northeast"},
		{"CAN.4_1", "NB", "New Brunswick", "Northeast"},
		{"CAN.7_1", "NS", "Nova Scotia", "Northeast"},
		{"CAN.10_1", "PE", "Prince Edward Island", "Northeast"},
		{"CAN.5_1", "NL", "Newfoundland and Labrador", "Northeast"},
		{"USA.20_1", "ME", "Maine", "Northeast"},
		{"USA.30_1", "NH", "New Hampshire", "Northeast"},
		{"USA.46_1", "VT", "Vermont", "Northeast"},
		{"USA.33_1", "NY", "New York", "Northeast"},
		{"USA.22_1", "MA", "Massachusetts", "Northeast"},
		{"USA.40_1", "RI", "Rhode Island", "Northeast"},
		{"USA.7_1", "CT", "Connecticut", "Northeast"},
		{"USA.31_1", "NJ", "New Jersey", "Northeast"},
		{"USA.39_1", "PA", "Pennsylvania", "Northeast"},
		// Southeast
		{"USA.18_1", "KY", "Kentucky", "Southeast"},
		{"USA.43_1", "TN", "Tennessee", "Southeast"},
		{"USA.49_1", "WV", "West Virginia", "Southeast"},
		{"USA.47_1", "VA", "Virginia", "Southeast"},
		{"USA.21_1", "MD", "Maryland", "Southeast"},
		{"USA.8_1", "DE", "Delaware", "Southeast"},
		{"USA.9_1", "DC", "District of Columbia", "Southeast"},
		{"USA.34_1", "NC", "North Carolina", "Southeast"},
		{"USA.41_1", "SC", "South Carolina", "Southeast"},
		{"USA.11_1", "GA", "Georgia", "Southeast"},
		{"USA.10_1", "FL", "Florida", "Southeast"},
		{"USA.1_1", "AL", "Alabama", "Southeast"},
		{"USA.25_1", "MS", "Mississippi", "Southeast"},
		{"USA.4_1", "AR", "Arkansas", "Southeast"},
		{"USA.19_1", "LA", "Louisiana", "Southeast"},
		// Outlying Islands -- no postal code exists, so the code is a short name.
		{"UMI.4_1", "Johnston", "Johnston Atoll", "Outlying Islands"},
		{"UMI.6_1", "Midway", "Midway Atoll", "Outlying Islands"},
		{"UMI.7_1", "Navassa", "Navassa Island", "Outlying Islands"},
		{"UMI.8_1", "Palmyra", "Palmyra Atoll", "Outlying Islands"},
		{"UMI.9_1", "Wake", "Wake Island", "Outlying Islands"},
	}
	for _, r := range land {
		add(r.id, r.code, r.name, Land, r.zone)
	}

	sea := []struct{ id, code, name string }{
		{"SEA.PAC", "N Pacific", "North Pacific, off the west coast"},
		{"SEA.CPA", "C Pacific", "Central Pacific, around Hawaii"},
		{"SEA.ATL", "N Atlantic", "North Atlantic, off the east coast"},
		{"SEA.GULF", "Gulf", "Gulf of Mexico"},
		{"SEA.ARC", "Arctic", "Arctic Ocean"},
	}
	for _, r := range sea {
		add(r.id, r.code, r.name, Marine, "At Sea")
	}
}

// Get returns a region by id.
func Get(id string) (Region, bool) {
	r, ok := all[id]
	return r, ok
}

// Len is how many regions exist.
func Len() int { return len(all) }

// InZone lists a zone's regions, in the order they were declared for land and
// in test order for the sea.
func InZone(zone string) []Region {
	var out []Region
	for _, r := range all {
		if r.Zone == zone {
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
