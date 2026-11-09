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
// head factor against photographs, writing data/neck-review.tsv
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

// reviewFig is one figure on the review page: the bird drawn at a neck ratio and head factor, and
// its proportions as numbers.
type reviewFig struct {
	Tile           tile
	Neck, Head     float64 // ratio and factor drawn
	NeckMm, HeadMm float64
	TotalMm        float64
	NeckShare      float64
	HeadShare      float64
}

// proportions recomputes a bird's head, neck, and total length at a neck ratio and head factor; bill,
// body, and tail stay as the dataset has them.
func proportions(sp *dataset.Species, neck, head float64) (headMm, neckMm, total float64) {
	body, _ := morphology.BodyLength(sp.Mass.Value)
	base, _ := morphology.HeadLength(sp.Mass.Value, "", "", nil)
	headMm, neckMm = base*head, body*neck
	return headMm, neckMm, sp.Segments.Bill.Value + headMm + neckMm + body + sp.Segments.Tail.Value
}

// figure draws sp at a neck ratio and head factor, to the scale of refLength.
func (s *server) figure(sp *dataset.Species, neck, head, refLength float64) reviewFig {
	h, n, total := proportions(sp, neck, head)
	body, _ := morphology.BodyLength(sp.Mass.Value)
	c := *sp
	c.Segments.Head.Value, c.Segments.Neck.Value, c.Segments.Body.Value = h, n, body
	c.TotalLength.Value = total
	g := view.Layout(&c, refLength, c.Wingspan.Value, reviewBox)
	t := newTile(&c, g, reviewBox, unitsFor(true, c.Mass.Value), dataset.LocaleNA)
	return reviewFig{Tile: t, Neck: neck, Head: head, NeckMm: n, HeadMm: h, TotalMm: total,
		NeckShare: n / total, HeadShare: h / total}
}

// current is the neck ratio and head factor a genus draws with now, and the words for where they
// come from.
func (rv *reviewState) current(g reviewGenus) (neck, head float64, source string) {
	neck, source = rv.table.NeckRatioFor(g.Family, g.Genus)
	head, hs := rv.table.HeadFactorFor(g.Family, g.Genus)
	if hs != "" {
		source += "; head: " + hs
	}
	return neck, head, source
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
	Neck, Head          float64
	Status              string
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
			neck, head, _ := rv.current(g)
			status := "family table"
			if p, ok := rv.table[g.Genus]; ok {
				status = p.Status
			} else if p, ok := rv.table[g.Family]; ok {
				status = p.Status + " (family)"
			}
			d.Queue = append(d.Queue, reviewQueueRow{Family: g.Family, Genus: g.Genus,
				Name: g.Rep.Name(dataset.LocaleNA), Neck: neck, Head: head, Status: status})
		}
		s.render(w, "review.html", d)
		return
	}
	g := rv.queue[i]
	neck, head, source := rv.current(g)
	_, _, total := proportions(g.Rep, neck, head)
	d.G, d.Name, d.Source, d.Pos = g, g.Rep.Name(dataset.LocaleNA), source, i+1
	d.Current = s.figure(g.Rep, neck, head, total)
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
			n, h, _ := rv.current(k)
			_, _, t := proportions(k.Rep, n, h)
			longest = math.Max(longest, t)
		}
	}
	for _, k := range kin {
		n, h, _ := rv.current(k)
		d.Strip = append(d.Strip, reviewStripFig{Genus: k.Genus, Fig: s.figure(k.Rep, n, h, longest), Here: k.Genus == g.Genus})
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
	curNeck, curHead, _ := rv.current(g)
	rv.mu.Unlock()
	neck, err1 := strconv.ParseFloat(r.URL.Query().Get("neck"), 64)
	head, err2 := strconv.ParseFloat(r.URL.Query().Get("head"), 64)
	if err1 != nil || err2 != nil || neck < 0.02 || neck > 2 || head < 0.3 || head > 2 {
		http.Error(w, "neck from 0.02 to 2 and head from 0.3 to 2", http.StatusBadRequest)
		return
	}
	_, _, curTotal := proportions(g.Rep, curNeck, curHead)
	_, _, total := proportions(g.Rep, neck, head)
	ref := math.Max(curTotal, total)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	var buf bytes.Buffer
	if err := s.tmpl.ExecuteTemplate(&buf, "reviewpair", struct{ Current, Adjusted reviewFig }{
		s.figure(g.Rep, curNeck, curHead, ref), s.figure(g.Rep, neck, head, ref)}); err != nil {
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
	neck, head, _ := rv.current(g)
	if r.FormValue("choice") == "adjusted" {
		var err1, err2 error
		neck, err1 = strconv.ParseFloat(r.FormValue("neck"), 64)
		head, err2 = strconv.ParseFloat(r.FormValue("head"), 64)
		if err1 != nil || err2 != nil {
			http.Error(w, "neck and head must be numbers", http.StatusBadRequest)
			return
		}
	}
	taxon := g.Genus
	if r.FormValue("family") == "on" {
		taxon = g.Family
	}
	p := morphology.Proportion{Taxon: taxon, Neck: round2(neck), Status: "reviewed", Reviewer: rv.reviewer,
		Date: time.Now().Format("2006-01-02"), Reason: strings.Join(strings.Fields(r.FormValue("reason")), " ")}
	if h := round2(head); h != 1 {
		p.Head = h
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
