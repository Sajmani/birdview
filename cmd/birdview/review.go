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

package main

import (
	"bytes"
	"fmt"
	"html/template"
	"math"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"birdview/internal/dataset"
	"birdview/internal/morphology"
	"birdview/internal/view"
)

// The review tool (tech.proportionsReviewTool): a local mode for judging each genus's neck ratio and
// head factor against photographs, writing data/proportions-review.tsv
// (decisions.proportionsReviewedPerTaxon).

// checkReviewFlags refuses review mode on a deployment, which a canonical host marks, and review mode
// without a reviewer to record.
func checkReviewFlags(path, reviewer, canonical string) error {
	switch {
	case path == "":
		return nil
	case canonical != "":
		return fmt.Errorf("-review is for one machine, not a deployment; it cannot be used with -canonical (tech.proportionsReviewTool)")
	case reviewer == "":
		return fmt.Errorf("-review needs -reviewer, the name each reviewed row records")
	}
	return nil
}

// reviewState is the review tool's table and where it is written.
type reviewState struct {
	mu       sync.Mutex
	path     string
	header   string // the table's comment lines, kept when it is written back
	reviewer string
	table    morphology.Review
	queue    []reviewGenus
}

// reviewGenus is one genus in the review queue and the bird that stands for it.
type reviewGenus struct {
	Family, Genus string
	Rep           *dataset.Species
}

// reviewFamilyOrder is the order families are reviewed in (decisions.proportionsReviewedPerTaxon); the
// rest of the non-passerines follow in taxonomic order.
var reviewFamilyOrder = []string{"Anatidae", "Ardeidae", "Scolopacidae", "Threskiornithidae",
	"Phalacrocoracidae", "Laridae", "Rallidae", "Podicipedidae"}

// enableReview turns review mode on over the table at path.
func (s *server) enableReview(path, reviewer string) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	table, err := morphology.ParseReview(bytes.NewReader(b))
	if err != nil {
		return fmt.Errorf("%s: %v", path, err)
	}
	var header strings.Builder
	for _, line := range strings.SplitAfter(string(b), "\n") {
		if !strings.HasPrefix(line, "#") {
			break
		}
		header.WriteString(line)
	}
	s.rv = &reviewState{path: path, header: header.String(), reviewer: reviewer, table: table,
		queue: s.reviewQueue()}
	for _, rt := range []route{
		{"/review", localOnly(http.HandlerFunc(s.reviewPage))},
		{"/review/figure", localOnly(http.HandlerFunc(s.reviewFigure))},
		{"/review/save", localOnly(http.HandlerFunc(s.reviewSave))},
	} {
		s.mux.Handle(rt.pattern, rt.handler)
	}
	return nil
}

// localOnly answers only requests from the machine the server runs on.
func localOnly(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		host, _, err := net.SplitHostPort(r.RemoteAddr)
		if ip := net.ParseIP(host); err != nil || ip == nil || !ip.IsLoopback() {
			http.Error(w, "the review tool answers only this machine", http.StatusForbidden)
			return
		}
		h.ServeHTTP(w, r)
	})
}

// reviewQueue lists every genus of every non-passerine family BirdView draws, families in review
// order, each genus represented by its most reported species.
func (s *server) reviewQueue() []reviewGenus {
	reports := func(sp *dataset.Species) int64 {
		var n int64
		for i, r := range sp.Regions {
			if strings.Count(r, ".") == 1 && i < len(sp.RegionCounts) {
				n += sp.RegionCounts[i]
			}
		}
		return n
	}
	rep := map[string]*dataset.Species{}
	familyOf := map[string]string{}
	for i := range s.ds.Species {
		sp := &s.ds.Species[i]
		if sp.Order == "Passeriformes" {
			continue
		}
		g := strings.Fields(sp.ScientificName)[0]
		familyOf[g] = sp.Family
		if cur, ok := rep[g]; !ok || reports(sp) > reports(cur) {
			rep[g] = sp
		}
	}
	rank := map[string]int{}
	for i, f := range reviewFamilyOrder {
		rank[f] = i - len(reviewFamilyOrder)
	}
	for i, f := range s.fams {
		if _, ok := rank[f.Name]; !ok {
			rank[f.Name] = i
		}
	}
	var q []reviewGenus
	for g, sp := range rep {
		q = append(q, reviewGenus{Family: familyOf[g], Genus: g, Rep: sp})
	}
	sort.Slice(q, func(i, j int) bool {
		if q[i].Family != q[j].Family {
			return rank[q[i].Family] < rank[q[j].Family]
		}
		return q[i].Genus < q[j].Genus
	})
	return q
}

// reviewBox is the drawing area of each figure on the review page.
var reviewBox = view.Box{W: 300, H: 340}

// setting is one value for every modeled measurement: the neck ratio and the head, body, and
// wingspan factors (decisions.allModeledProportionsReviewed).
type setting struct{ Neck, Head, Body, Wing float64 }

// reviewFig is one figure on the review page: the bird drawn at a setting, and its proportions as
// numbers.
type reviewFig struct {
	Tile tile
	setting
	NeckMm, HeadMm, BodyMm float64
	TotalMm, WingMm        float64
	NeckShare, HeadShare   float64
	BodyShare              float64
}

// sized is a bird's modeled measurements at a setting; bill, tail, and wing chord stay as measured,
// and a measured wingspan stays as it is.
type sized struct{ head, neck, body, total, wing float64 }

func proportions(sp *dataset.Species, at setting) sized {
	body, _ := morphology.BodyLength(sp.Mass.Value)
	head, _ := morphology.HeadLength(sp.Mass.Value, "", "", nil)
	z := sized{head: head * at.Head, body: body * at.Body, wing: sp.Wingspan.Value}
	z.neck = z.body * at.Neck
	z.total = sp.Segments.Bill.Value + z.head + z.neck + z.body + sp.Segments.Tail.Value
	if sp.Wingspan.Prov.Kind == dataset.Modeled {
		built := sp.Wingspan.Prov.Factor
		if built == 0 {
			built = 1
		}
		z.wing = sp.Wingspan.Value / built * at.Wing
	}
	return z
}

// figure draws sp at a setting, to the scale of refLength.
func (s *server) figure(sp *dataset.Species, at setting, refLength float64) reviewFig {
	z := proportions(sp, at)
	c := *sp
	c.Segments.Head.Value, c.Segments.Neck.Value, c.Segments.Body.Value = z.head, z.neck, z.body
	c.TotalLength.Value, c.Wingspan.Value = z.total, z.wing
	g := view.Layout(&c, refLength, c.Wingspan.Value, reviewBox)
	t := newTile(&c, g, reviewBox, unitsFor(true, c.Mass.Value), dataset.LocaleNA)
	return reviewFig{Tile: t, setting: at, NeckMm: z.neck, HeadMm: z.head, BodyMm: z.body,
		TotalMm: z.total, WingMm: z.wing,
		NeckShare: z.neck / z.total, HeadShare: z.head / z.total, BodyShare: z.body / z.total}
}

// current is the setting a genus draws with now, and the words for where it comes from.
func (rv *reviewState) current(g reviewGenus) (setting, string) {
	var at setting
	var source string
	at.Neck, source = rv.table.NeckRatioFor(g.Family, g.Genus)
	for _, f := range []struct {
		name string
		v    *float64
		of   func(string, string) (float64, string)
	}{
		{"head", &at.Head, rv.table.HeadFactorFor},
		{"body", &at.Body, rv.table.BodyFactorFor},
		{"wingspan", &at.Wing, rv.table.WingspanFactorFor},
	} {
		var src string
		*f.v, src = f.of(g.Family, g.Genus)
		if src != "" {
			source += "; " + f.name + ": " + src
		}
	}
	return at, source
}

func (rv *reviewState) find(genus string) (int, bool) {
	for i, g := range rv.queue {
		if g.Genus == genus {
			return i, true
		}
	}
	return 0, false
}

// reviewPageData is the review page: the queue when no genus is chosen, one genus otherwise.
type reviewPageData struct {
	Queue []reviewQueueRow

	G                 reviewGenus
	Name, EBirdURL    string
	WingModeled       bool
	Current, Adjusted reviewFig
	Source            string
	Strip             []reviewStripFig
	Prev, Next        string
	Pos, Total        int
	Saved             string
	Reviewer, Table   string
}

type reviewQueueRow struct {
	Family, Genus, Name string
	setting
	Status string
}

type reviewStripFig struct {
	Genus string
	Fig   reviewFig
	Here  bool
}

func (s *server) reviewPage(w http.ResponseWriter, r *http.Request) {
	rv := s.rv
	rv.mu.Lock()
	defer rv.mu.Unlock()
	d := reviewPageData{Total: len(rv.queue), Reviewer: rv.reviewer, Table: filepath.Base(rv.path),
		Saved: r.URL.Query().Get("saved")}
	i, ok := rv.find(r.URL.Query().Get("genus"))
	if !ok {
		for _, g := range rv.queue {
			at, _ := rv.current(g)
			status := "family table"
			if p, ok := rv.table[g.Genus]; ok {
				status = p.Status
			} else if p, ok := rv.table[g.Family]; ok {
				status = p.Status + " (family)"
			}
			d.Queue = append(d.Queue, reviewQueueRow{Family: g.Family, Genus: g.Genus,
				Name: g.Rep.Name(dataset.LocaleNA), setting: at, Status: status})
		}
		s.render(w, "review.html", d)
		return
	}
	g := rv.queue[i]
	at, source := rv.current(g)
	d.G, d.Name, d.Source, d.Pos = g, g.Rep.Name(dataset.LocaleNA), source, i+1
	d.WingModeled = g.Rep.Wingspan.Prov.Kind == dataset.Modeled
	d.Current = s.figure(g.Rep, at, proportions(g.Rep, at).total)
	d.Adjusted = d.Current
	d.EBirdURL = d.Current.Tile.EBirdURL
	if i > 0 {
		d.Prev = rv.queue[i-1].Genus
	}
	if i+1 < len(rv.queue) {
		d.Next = rv.queue[i+1].Genus
	}
	// The family's genera side by side, each at its current values and one scale.
	var kin []reviewGenus
	longest := 0.0
	for _, k := range rv.queue {
		if k.Family == g.Family {
			kin = append(kin, k)
			kat, _ := rv.current(k)
			longest = math.Max(longest, proportions(k.Rep, kat).total)
		}
	}
	for _, k := range kin {
		kat, _ := rv.current(k)
		d.Strip = append(d.Strip, reviewStripFig{Genus: k.Genus, Fig: s.figure(k.Rep, kat, longest), Here: k.Genus == g.Genus})
	}
	s.render(w, "review.html", d)
}

// reviewFigure redraws a genus's bird at the neck ratio and head factor asked for, to the scale of the
// longer of it and the current figure, so the two compare.
func (s *server) reviewFigure(w http.ResponseWriter, r *http.Request) {
	rv := s.rv
	rv.mu.Lock()
	i, ok := rv.find(r.URL.Query().Get("genus"))
	if !ok {
		rv.mu.Unlock()
		http.NotFound(w, r)
		return
	}
	g := rv.queue[i]
	cur, _ := rv.current(g)
	rv.mu.Unlock()
	at, err := settingOf(r.URL.Query().Get, cur)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	ref := math.Max(proportions(g.Rep, cur).total, proportions(g.Rep, at).total)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	var buf bytes.Buffer
	if err := s.tmpl.ExecuteTemplate(&buf, "reviewpair", struct{ Current, Adjusted reviewFig }{
		s.figure(g.Rep, cur, ref), s.figure(g.Rep, at, ref)}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(buf.Bytes())
}

// reviewSave writes the genus's row, or its family's, as reviewed, and goes on to the next genus.
func (s *server) reviewSave(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}
	rv := s.rv
	rv.mu.Lock()
	defer rv.mu.Unlock()
	i, ok := rv.find(r.FormValue("genus"))
	if !ok {
		http.NotFound(w, r)
		return
	}
	g := rv.queue[i]
	at, _ := rv.current(g)
	if r.FormValue("choice") == "adjusted" {
		var err error
		if at, err = settingOf(r.FormValue, at); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}
	taxon := g.Genus
	if r.FormValue("family") == "on" {
		taxon = g.Family
	}
	p := morphology.Proportion{Taxon: taxon, Neck: round2(at.Neck), Status: "reviewed", Reviewer: rv.reviewer,
		Date: time.Now().Format("2006-01-02"), Reason: strings.Join(strings.Fields(r.FormValue("reason")), " ")}
	factor := func(v float64) float64 {
		if v = round2(v); v == 1 {
			return 0
		}
		return v
	}
	p.Head, p.Body = factor(at.Head), factor(at.Body)
	// A wingspan factor is kept only where the wingspan is modeled (decisions.allModeledProportionsReviewed).
	if g.Rep.Wingspan.Prov.Kind == dataset.Modeled {
		p.Wing = factor(at.Wing)
	}
	next := morphology.Review{}
	for k, v := range rv.table {
		next[k] = v
	}
	next[taxon] = p
	var buf bytes.Buffer
	if err := next.Write(&buf, rv.header); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if _, err := morphology.ParseReview(bytes.NewReader(buf.Bytes())); err != nil {
		http.Error(w, "the row would not read back: "+err.Error(), http.StatusBadRequest)
		return
	}
	tmp := rv.path + ".tmp"
	if err := os.WriteFile(tmp, buf.Bytes(), 0o644); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := os.Rename(tmp, rv.path); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	rv.table = next
	to := "/review?saved=" + template.URLQueryEscaper(taxon)
	if i+1 < len(rv.queue) {
		to += "&genus=" + template.URLQueryEscaper(rv.queue[i+1].Genus)
	}
	http.Redirect(w, r, to, http.StatusSeeOther)
}

func round2(v float64) float64 { return math.Round(v*100) / 100 }

// settingOf reads a setting from a form or query, keeping a value of cur for any field not given.
func settingOf(get func(string) string, cur setting) (setting, error) {
	at := cur
	for _, f := range []struct {
		name   string
		v      *float64
		lo, hi float64
	}{
		{"neck", &at.Neck, 0.02, 2}, {"head", &at.Head, 0.3, 2}, {"body", &at.Body, 0.5, 2}, {"wing", &at.Wing, 0.5, 2},
	} {
		raw := get(f.name)
		if raw == "" {
			continue
		}
		v, err := strconv.ParseFloat(raw, 64)
		if err != nil || v < f.lo || v > f.hi {
			return cur, fmt.Errorf("%s must be a number from %g to %g", f.name, f.lo, f.hi)
		}
		*f.v = v
	}
	return at, nil
}
