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

package dataset

import (
	"compress/gzip"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"io"
	"sort"
)

// Counties is county presence held apart from the species (decisions.countiesServedApart): one
// compact table per county, read only when a county is chosen, so a page walks the counties on
// screen rather than every county of every species. The ingest writes it beside species.json and
// the server loads it once, with -counties.
type Counties struct {
	// Counties are the county regions, each named as GBIF names it and placed in its state, with
	// its records over every species (product.regionAndMonthFilter, decisions.countyListingRule).
	Counties []County
	// Species names the species Entries refer to, by scientific name, so a county file is joined
	// to the dataset by name rather than trusted to share its order.
	Species []string
	// Entries are each county's species with records there, parallel to Counties and ordered by
	// species index.
	Entries [][]CountyEntry

	byID map[string]int
	sum  string
}

// Sum is the SHA-256 of the file the counties were read from, empty for counties built in memory,
// so a server's validators change when its county data does (tech.cacheLifetimeStated).
func (c *Counties) Sum() string { return c.sum }

// CountyEntry is one species' records in one county: twelve monthly counts, January first, whose
// sum is its records there, and whether it stays under the bar and which months eBird withholds.
type CountyEntry struct {
	Species uint16
	Flags   uint16
	Months  [12]uint32
}

const countyBelow = 1 << 15

// Below says the species has records in the county but stays under its bar: it counts toward the
// county's figures but is not listed there (decisions.belowBarCounted).
func (e CountyEntry) Below() bool { return e.Flags&countyBelow != 0 }

// Withheld marks the months eBird withholds the species in the county (decisions.ebirdSilenceNotAbsence).
func (e CountyEntry) Withheld() [12]bool {
	var w [12]bool
	for m := range w {
		w[m] = e.Flags&(1<<m) != 0
	}
	return w
}

// Monthly widens the twelve counts to the dataset's width.
func (e CountyEntry) Monthly() [12]int64 {
	var out [12]int64
	for m, n := range e.Months {
		out[m] = int64(n)
	}
	return out
}

// Records is the species' records in the county, every month.
func (e CountyEntry) Records() int64 {
	var n int64
	for _, v := range e.Months {
		n += int64(v)
	}
	return n
}

// NewCountyEntry packs one species' records in a county.
func NewCountyEntry(species int, months [12]int64, below bool, withheld [12]bool) (CountyEntry, error) {
	if species < 0 || species > 0xffff {
		return CountyEntry{}, fmt.Errorf("species index %d does not fit a county entry", species)
	}
	e := CountyEntry{Species: uint16(species)}
	for m, n := range months {
		if n < 0 || n > 0xffffffff {
			return CountyEntry{}, fmt.Errorf("%d records in one county-month does not fit a county entry", n)
		}
		e.Months[m] = uint32(n)
		if withheld[m] {
			e.Flags |= 1 << m
		}
	}
	if below {
		e.Flags |= countyBelow
	}
	return e, nil
}

// County is one county region: its GADM code, its name, its state's GADM code, and every bird
// record there, in all and by month, the denominator of a county's share (product.shareOfRecords).
type County struct {
	ID      string    `json:"id"`
	Name    string    `json:"name"`
	State   string    `json:"state"`
	Total   int64     `json:"total,omitempty"`
	Monthly [12]int64 `json:"monthly,omitempty"`
}

// Index readies the lookups, sorting each county's entries by species; call it after filling
// Counties and Entries by hand.
func (c *Counties) Index() error {
	if len(c.Entries) != len(c.Counties) {
		return fmt.Errorf("%d counties but entries for %d", len(c.Counties), len(c.Entries))
	}
	c.byID = make(map[string]int, len(c.Counties))
	for i, ct := range c.Counties {
		if _, dup := c.byID[ct.ID]; dup {
			return fmt.Errorf("county %s appears twice", ct.ID)
		}
		c.byID[ct.ID] = i
		es := c.Entries[i]
		sort.Slice(es, func(a, b int) bool { return es[a].Species < es[b].Species })
		for j, e := range es {
			if int(e.Species) >= len(c.Species) {
				return fmt.Errorf("county %s names species %d of %d", ct.ID, e.Species, len(c.Species))
			}
			if j > 0 && es[j-1].Species == e.Species {
				return fmt.Errorf("county %s lists %s twice", ct.ID, c.Species[e.Species])
			}
		}
	}
	return nil
}

// County finds a county by its GADM code.
func (c *Counties) County(id string) (County, bool) {
	i, ok := c.byID[id]
	if !ok {
		return County{}, false
	}
	return c.Counties[i], true
}

// Entry finds a species' records in a county, by the species' index in Species.
func (c *Counties) Entry(id string, species int) (CountyEntry, bool) {
	i, ok := c.byID[id]
	if !ok {
		return CountyEntry{}, false
	}
	es := c.Entries[i]
	j := sort.Search(len(es), func(k int) bool { return int(es[k].Species) >= species })
	if j < len(es) && int(es[j].Species) == species {
		return es[j], true
	}
	return CountyEntry{}, false
}

// Write encodes the counties gzipped, as the ingest leaves them beside species.json.
func (c *Counties) Write(w io.Writer) error {
	zw := gzip.NewWriter(w)
	if err := gob.NewEncoder(zw).Encode(c); err != nil {
		return err
	}
	return zw.Close()
}

// ReadCounties decodes what Write wrote and readies its lookups.
func ReadCounties(r io.Reader) (*Counties, error) {
	h := sha256.New()
	zr, err := gzip.NewReader(io.TeeReader(r, h))
	if err != nil {
		return nil, err
	}
	defer zr.Close()
	var c Counties
	if err := gob.NewDecoder(zr).Decode(&c); err != nil {
		return nil, err
	}
	if err := c.Index(); err != nil {
		return nil, err
	}
	if _, err := io.Copy(io.Discard, zr); err != nil {
		return nil, err
	}
	c.sum = hex.EncodeToString(h.Sum(nil))
	return &c, nil
}
