package status

import (
	"bufio"
	"fmt"
	"io"
	"sort"
	"strings"
	"time"

	"birdview/internal/dataset"
	"birdview/internal/taxonomy"
)

// Source names this input in provenance records.
const Source = "BirdView's own occurrence status, derived from GBIF occurrence records (CC BY 4.0)"

// Row is one line of the committed status table.
type Row struct {
	// Name is the species as the OCCURRENCE DATA spells it, lowercased. It is
	// not necessarily an AviList name -- eBird still files five species under
	// a pre-move genus -- so Scope resolves it before anyone joins on it.
	Name   string
	Status Status
	Region string // the region the verdict came from
	Episodes,
	Years int
	Records int64

	// Commonest is the region where the species makes up the biggest SHARE
	// of what is recorded there, with both halves of the ratio kept.
	//
	// Share, not record count, and the two disagree for 564 of 1,081
	// species: a count says where the birdwatchers are. California holds 35x
	// the Sharp-shinned Hawk records of Yukon and 118x the birdwatching, so
	// the hawk is one record in 250 in Yukon against one in 835 in
	// California. The Millerbird settles it -- by count its home reads
	// "Hawaii", by share it reads Nihoa, the island it actually lives on.
	//
	// Both numbers are carried rather than the ratio so that a share resting
	// on nine records can say so, which is what a minimum-evidence threshold
	// would otherwise have to do.
	Commonest string
	CommonestRecords,
	CommonestRegionRecords int64

	// LastSeen is the most recent date the species was recorded on in Region,
	// and the zero time where no date was measured at all.
	//
	// Carried because it is the entire content of a `historical` verdict.
	// Without it the table can say Passenger Pigeon recurs across 89 episodes
	// in 52 years AND that it is historical, but not the one fact that
	// reconciles them -- that every one of those records ends in 1902. A
	// verdict whose evidence is withheld is not checkable (P-075).
	LastSeen time.Time
}

// DateLayout is how LastSeen is written in the table: a plain ISO date, or
// "-" where the derivation measured none. Dates are what this table is about,
// so a missing one is written the way the file writes every other blank.
const DateLayout = "2006-01-02"

// Read parses the committed table written by cmd/birdview-status.
func Read(r io.Reader) ([]Row, error) {
	var out []Row
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 1<<20), 1<<20)
	for ln := 0; sc.Scan(); {
		ln++
		line := strings.TrimRight(sc.Text(), "\r")
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		f := strings.Split(line, "\t")
		if len(f) != 10 {
			return nil, fmt.Errorf("status: line %d has %d fields, want 10: %q", ln, len(f), line)
		}
		var row Row
		row.Name = strings.ToLower(strings.TrimSpace(f[0]))
		row.Status = Status(f[1])
		row.Region = f[2]
		if _, err := fmt.Sscanf(f[3]+" "+f[4]+" "+f[5], "%d %d %d",
			&row.Episodes, &row.Years, &row.Records); err != nil {
			return nil, fmt.Errorf("status: line %d: %w", ln, err)
		}
		row.Commonest = f[6]
		if _, err := fmt.Sscanf(f[7]+" "+f[8], "%d %d",
			&row.CommonestRecords, &row.CommonestRegionRecords); err != nil {
			return nil, fmt.Errorf("status: line %d: %w", ln, err)
		}
		if d := strings.TrimSpace(f[9]); d != "" && d != "-" {
			t, err := time.Parse(DateLayout, d)
			if err != nil {
				return nil, fmt.Errorf("status: line %d: last seen %q: %w", ln, d, err)
			}
			row.LastSeen = t
		}
		switch row.Status {
		case Established, Rare, Vagrant, Historical, Absent:
		default:
			return nil, fmt.Errorf("status: line %d: unknown status %q", ln, row.Status)
		}
		out = append(out, row)
	}
	return out, sc.Err()
}

// Scope returns the AviList names BirdView draws, from the table plus the
// declared exceptions (P-001, P-071).
//
// The table is keyed on the name the occurrence data uses, which is not
// always AviList's: eBird still files five woodpeckers and the Evening
// Grosbeak under a pre-move genus. taxonomy.AviListRenames carries those, and
// is applied here so every caller joins on one namespace.
func Scope(rows []Row) []string {
	in := map[string]bool{}
	for _, r := range rows {
		if !r.Status.InScope() {
			continue
		}
		in[aviListName(r.Name)] = true
	}
	for name := range Declared {
		in[name] = true
	}
	out := make([]string, 0, len(in))
	for n := range in {
		out = append(out, n)
	}
	sort.Strings(out)
	return out
}

// OutOfScope returns the rejected species, keyed by AviList name: everything
// the table classified that Scope did not take.
//
// It is the complement of Scope over the same rows and in the same namespace,
// computed from Scope rather than from InScope, so the two cannot drift. The
// difference is not academic -- the nine declared exceptions of P-071 are in
// scope with an out-of-scope status, and a complement built on InScope alone
// would publish an epitaph for a bird the index is drawing.
//
// Where two occurrence names fold onto one AviList name, the strongest
// verdict wins, on the same rule as Best.
//
// Verifies: P-075.
func OutOfScope(rows []Row) map[string]Row {
	in := map[string]bool{}
	for _, n := range Scope(rows) {
		in[n] = true
	}
	out := map[string]Row{}
	for _, r := range rows {
		n := aviListName(r.Name)
		if in[n] {
			continue
		}
		if cur, seen := out[n]; seen && Rank(cur.Status) >= Rank(r.Status) {
			continue
		}
		out[n] = r
	}
	return out
}

// CommonestBy keys each row's peak share by AviList name, so the ingest can
// attach it without redoing the namespace mapping.
//
// Rows with no usable peak are left out rather than given a zero: a species
// with nothing to divide has no "commonest" region, and an entry saying one
// in zero would be worse than silence.
func CommonestBy(rows []Row) map[string]*dataset.Commonest {
	out := make(map[string]*dataset.Commonest, len(rows))
	for _, r := range rows {
		if r.Commonest == "" || r.Commonest == "-" ||
			r.CommonestRecords == 0 || r.CommonestRegionRecords == 0 {
			continue
		}
		out[aviListName(r.Name)] = &dataset.Commonest{
			Region:        r.Commonest,
			Records:       r.CommonestRecords,
			RegionRecords: r.CommonestRegionRecords,
		}
	}
	return out
}

// aviListName maps a name the occurrence data uses onto AviList's.
func aviListName(lower string) string {
	t := title(lower)
	if alt, ok := taxonomy.AviListRenames[t]; ok {
		return alt
	}
	return t
}

func title(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// Declared are the species kept in scope by hand, because the occurrence data
// cannot speak for them (P-071). Each names its reason.
//
// Every one is a bird the ABA area genuinely holds. They are here because the
// open record is empty or stale, not because the bird is: eBird withholds
// conservation-sensitive species outright, iNaturalist obscures dates for
// threatened taxa, and two of these breed outside the US and Canada and so
// fall outside the country filter the derivation runs under.
//
// This table is the editorial surface of the whole scope model, and it is
// nine species. The rule it patches keeps 830 of the previous 839 on measured
// evidence alone.
//
// Verifies: P-001, P-071.
var Declared = map[string]string{
	"Pseudonestor xanthophrys": "Maui Parrotbill: critically endangered Hawaiian endemic; " +
		"no open record at all, eBird withholding it and iNaturalist holding none",
	"Palmeria dolei": "Akohekohe: critically endangered Maui endemic; iNaturalist has records " +
		"but obscures their dates for threatened taxa, so recurrence is unmeasurable",
	"Myadestes palmeri": "Puaiohi: critically endangered Kauai endemic, one episode in the " +
		"open record for the same reason",
	"Hydrobates socorroensis": "Townsend's Storm-Petrel: breeds on Mexican islands and is seen " +
		"off California; the open record carries no dated occurrence on land, and its " +
		"records at sea are not yet reachable by the recurrence test (CR-030)",
	"Pterodroma cahow": "Bermuda Petrel: breeds on Bermuda and ranges into waters " +
		"BirdView covers, but Bermuda is outside the US/CA country filter the " +
		"derivation runs under, so the open record holds nothing for it here",
	"Pterodroma deserta": "Desertas Petrel: ranges into waters BirdView covers from a " +
		"Madeiran breeding colony, which is outside the country filter",
	// The three below said the derivation sees these birds rarely "because
	// almost nobody is on the water where it is". That stated a measurement
	// nobody had made: the recurrence test runs on gadmGid and GBIF has no
	// gadmGid for open ocean, so pass 2 has never asked the water anything
	// (CR-030). All three are 95-100% pelagic with thousands of records at
	// sea. Whether they would clear the bar unaided is unknown, so the
	// exceptions stand and only the reason is corrected.
	"Pelagodroma marina": "White-faced Storm-Petrel: a pelagic whose records are almost all " +
		"at sea, where the recurrence test cannot yet look; on land the derivation sees " +
		"one episode",
	"Pterodroma cervicalis": "White-necked Petrel: the same, two episodes on land",
	"Pterodroma arminjoniana": "Trindade Petrel: the same, and its land records are old " +
		"enough to read as historical",
}
