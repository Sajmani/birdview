package status

import (
	"bufio"
	"fmt"
	"io"
	"sort"
	"strings"

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
}

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
		if len(f) != 6 {
			return nil, fmt.Errorf("status: line %d has %d fields, want 6: %q", ln, len(f), line)
		}
		var row Row
		row.Name = strings.ToLower(strings.TrimSpace(f[0]))
		row.Status = Status(f[1])
		row.Region = f[2]
		if _, err := fmt.Sscanf(f[3]+" "+f[4]+" "+f[5], "%d %d %d",
			&row.Episodes, &row.Years, &row.Records); err != nil {
			return nil, fmt.Errorf("status: line %d: %w", ln, err)
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
		"off California; the open record carries no dated ABA-area occurrence",
	"Pterodroma cahow": "Bermuda Petrel: breeds on Bermuda, which is in the ABA Area and " +
		"outside the US/CA country filter the derivation runs under",
	"Pterodroma deserta": "Desertas Petrel: ranges into ABA-area waters from a Madeiran " +
		"breeding colony, outside the country filter",
	"Pelagodroma marina": "White-faced Storm-Petrel: a pelagic the derivation sees in one " +
		"episode because almost nobody is on the water where it is",
	"Pterodroma cervicalis": "White-necked Petrel: the same, two episodes",
	"Pterodroma arminjoniana": "Trindade Petrel: the same, and its records are old enough " +
		"to read as historical",
}
