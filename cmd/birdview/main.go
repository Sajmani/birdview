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

// Command birdview serves the size explorer on localhost.
//
// Rendering happens on the server, in html/template. That keeps the ordering,
// sectioning and geometry rules in Go where they are property-tested, rather
// than duplicating them in a language the acceptance criteria cannot reach.
// Filters and toggles are query parameters, and a tile's detail view is a URL.
//
// Verifies: tech.twoCommandsOneModule, tech.selfContainedServer, tech.servesOnLocalhost.
package main

import (
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"log"
	"math"
	"net"
	"net/http"
	"net/url"
	"os"
	pathpkg "path"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"bytes"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"

	"embed"

	goldmarkast "github.com/yuin/goldmark/ast"
	goldmarkparser "github.com/yuin/goldmark/parser"
	goldmarktext "github.com/yuin/goldmark/text"
	goldmarkutil "github.com/yuin/goldmark/util"

	"birdview/data"
	"birdview/docs"
	"birdview/internal/dataset"
	"birdview/internal/inat"
	"birdview/internal/month"
	"birdview/internal/region"
	"birdview/internal/status"
	"birdview/internal/taxonomy"
	"birdview/internal/trait"
	"birdview/internal/view"
	"birdview/web"
)

func main() {
	addr := flag.String("addr", "localhost:8080", "listen address")
	data := flag.String("data", "web/data/species.json", "compiled dataset (a build artifact)")
	dev := flag.String("dev", "", "serve templates and static files from this directory instead of the embed, so edits show on reload")
	// Where readers are asked to send feedback. A URL, not an address: a
	// mailto on a public page is a mailbox harvested within the week, and
	// the owner's personal address is not the project's contact point
	// (product.noPersonalAddress).
	//
	// Empty by default, and an empty value publishes NOTHING rather than a
	// dead link -- a site can go up before its tracker exists, but it must
	// not invite a reader to a page that does not.
	feedback := flag.String("feedback", "", "issue tracker URL for reader feedback; omitted from the page when unset")
	// The deployment's one public host, e.g. birdview.site; requests for any
	// other host are redirected to it (tech.oneCanonicalHost). A flag, not a
	// constant, because the host belongs to whoever deploys, not to the code.
	canonical := flag.String("canonical", "", "the one host to answer at; other hosts are redirected to it over HTTPS. Unset serves every host")
	flag.Parse()

	if *dev != "" {
		if err := devAssets(*dev); err != nil {
			log.Fatalf("birdview: %v", err)
		}
		log.Printf("birdview: DEV MODE, assets from %s — the embed is not being used", *dev)
	}

	ds, err := load(*data)
	if err != nil {
		log.Fatalf("birdview: %v\n\nRun birdview-ingest first; the dataset is a build artifact "+
			"and is not committed (tech.datasetNeverCommitted).", err)
	}
	srv, err := newServer(ds)
	if err == nil {
		srv.feedback = *feedback
		srv.canonical = *canonical
	}
	if err != nil {
		log.Fatalf("birdview: %v", err)
	}
	fmt.Printf("BirdView: %d species, serving http://%s\n", len(ds.Species), *addr)
	if err := http.ListenAndServe(*addr, srv); err != nil {
		log.Fatalf("birdview: %v", err)
	}
}

func load(path string) (*dataset.Dataset, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var ds dataset.Dataset
	if err := json.NewDecoder(f).Decode(&ds); err != nil {
		return nil, fmt.Errorf("decoding %s: %w", path, err)
	}
	if len(ds.Species) == 0 {
		return nil, fmt.Errorf("%s contains no species", path)
	}
	return &ds, nil
}

type server struct {
	mux    *http.ServeMux
	routes []route
	// onRender, when set, is told the name of every template rendered. Only
	// the checks set it, to confirm they have seen every kind of page.
	onRender func(template string)
	// hereJS is web/static/here.js, inlined into the home page.
	hereJS template.JS
	// nextFindJS is web/static/next-find.js, inlined into the home page (product.nextView).
	nextFindJS template.JS
	// search is each species' normalized search text, by id, which the
	// home page's find box matches against (product.nextView).
	search map[string]string
	// boundaryURL is where the region boundary file is served: its path
	// carries the file's hash, so it can be cached for a year
	// (tech.regionBoundaries, decisions.hereAndNowCaching).
	boundaryURL string
	// ds is what the home page draws: every species except the exotics
	// (product.escapedBirdsListed). all is the whole dataset, for the detail pages and the lists
	// that must name every bird.
	ds   *dataset.Dataset
	all  *dataset.Dataset
	tmpl *template.Template
	byID map[string]*dataset.Species
	// idOf maps an AviList scientific name to the served id, so a page can
	// name another species without searching the slice (product.namesRecordsFiledUnder).
	idOf map[string]string
	// sameAs maps a scientific name to the other drawn species whose traits
	// come from the same AVONET row, sorted (product.borrowedMeasurementsSaySo).
	sameAs map[string][]string
	// undrawnByID holds the species the derivation classified and rejected.
	// Consulted only after byID misses, so a URL that stopped working can say
	// why instead of 404ing into silence (product.undrawnSpeciesAnswer).
	undrawnByID map[string]*dataset.Undrawn
	fams        []familyOption
	// devDir is empty for a real binary. When set, render re-parses the
	// templates on every request so an edit shows on reload.
	devDir string
	// feedback is the issue tracker readers are pointed at, or empty, in
	// which case no page says anything about feedback at all (product.noPersonalAddress).
	feedback string
	// canonical is the one host this deployment answers at; a request for
	// any other is redirected there (tech.oneCanonicalHost). Empty serves every host.
	canonical string
	// build fingerprints everything a response is rendered from: the
	// dataset, the templates, and the static files. It is the first half of
	// every ETag (tech.cacheLifetimeStated).
	build string
}

type familyOption struct {
	Name, Common     string
	Color, TextColor string
	Order            string
	OrderCommon      string
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if s.canonical != "" && !sameHost(r.Host, s.canonical) {
		// Permanent, so bookmarks and search engines move to the one address;
		// path and query kept, so every view lands where it was going
		// (tech.oneCanonicalHost).
		http.Redirect(w, r, "https://"+s.canonical+r.URL.RequestURI(), http.StatusMovedPermanently)
		return
	}
	compressed(s.revalidated(s.mux)).ServeHTTP(w, r)
}

// sameHost reports whether a request's Host names the canonical host,
// ignoring case and any port.
func sameHost(requested, canonical string) bool {
	if h, _, err := net.SplitHostPort(requested); err == nil {
		requested = h
	}
	return strings.EqualFold(requested, canonical)
}

// revalidated gives every response a validator and tells caches to check it
// before reusing it (tech.cacheLifetimeStated).
//
// BirdView sent no Cache-Control, no ETag, and no Last-Modified at all, which
// does not mean "do not cache" -- it means the browser may apply its own
// heuristic and has nothing to revalidate against when it does. The symptom
// is a reader reloading a page after a rebuild and being served the old one,
// with no way to tell and no way to force the issue. It cost a review round
// on a page that was correct on the server.
//
// `no-cache` is the counter-intuitive half: it means *store this, but ask
// before reusing it*, which is exactly right here. The page is cheap to
// validate and must never be stale, since the whole point of a figure like
// product.shareOfRecords's is that it is current.
//
// The ETag is computed from the build and the URL rather than by hashing the
// body, because hashing the body means buffering it, and the index is 3.3 MB
// before compression. That is sound only because a page is a pure function
// of (dataset, templates, static files, request URI) -- there is no clock in
// the render path -- and acceptance.honestRevalidation asserts that rather than trusting it.
func (s *server) revalidated(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// -dev re-reads templates per request, so no fingerprint taken at
		// startup stays true. Sending no validator is correct there: the
		// whole point of dev mode is that an edit shows on reload.
		// -dev mode has no fingerprint to trust, and /now answers from the
		// clock: neither may be stored, and neither gets a validator that
		// would let a stale answer be confirmed (decisions.hereAndNowCaching).
		// The home page's month=now redirects by the clock as well (product.nextView).
		if s.build == "" || r.URL.Path == "/now" ||
			(r.URL.Path == "/" && slices.Contains(r.URL.Query()["month"], "now")) {
			w.Header().Set("Cache-Control", "no-store")
			h.ServeHTTP(w, r)
			return
		}
		sum := sha256.Sum256([]byte(s.build + "\x00" + r.Method + "\x00" + r.URL.RequestURI()))
		etag := `"` + hex.EncodeToString(sum[:16]) + `"`
		w.Header().Set("ETag", etag)
		w.Header().Set("Cache-Control", "no-cache")
		// Compared against the bare tag: the client echoes whatever we sent,
		// and the gzip layer may have suffixed it to distinguish the encoded
		// representation from the identity one.
		if m := r.Header.Get("If-None-Match"); m != "" {
			for _, c := range strings.Split(m, ",") {
				c = strings.TrimSpace(c)
				c = strings.TrimPrefix(c, "W/")
				if c == etag || c == etag[:len(etag)-1]+`-gz"` {
					w.WriteHeader(http.StatusNotModified)
					return
				}
			}
		}
		h.ServeHTTP(w, r)
	})
}

// renamed says how two binomials for one taxon differ: a short noun for the
// row's own line, and a sentence naming both halves next to each other.
//
// Two shapes in AviListRenames and they are not the same event. Five are
// genus moves, where the epithet is untouched and the eye catches the
// difference at once. Two are the same genus with a corrected epithet, and
// there the two names can differ by a single letter -- *Pica nuttalli* and
// *Pica nuttallii* -- so a sentence that names both without saying what
// changed reads as though it has printed the same name twice. It was
// reported as a bug in the page, which is what it looked like.
//
// Both halves are therefore put side by side in one clause, because the
// only way to show a one-letter difference in running text is to make the
// reader's eye compare two adjacent words.
//
// It says eBird and not "the occurrence data", which would be wrong in both
// directions. iNaturalist follows AviList for all five woodpeckers and the
// grosbeak -- 53,705 records under *Leuconotopicus villosus* -- and follows
// neither spelling for *Pica nuttallii*, where both databases hold nothing.
// eBird is the one thing true of every row: it holds zero under the AviList
// name in all seven, which is why the join needs the old one at all.
func renamed(old, current string) (what, how string) {
	og, oe, _ := strings.Cut(old, " ")
	ng, ne, _ := strings.Cut(current, " ")
	switch {
	case og != ng && oe != ne:
		return "name", "AviList places this species in the genus " + ng +
			" and spells the epithet " + ne + "; eBird files it as " + old + "."
	case og != ng:
		return "genus", "AviList places this species in the genus " + ng +
			" and eBird files it under " + og + "."
	case oe != ne:
		return "spelling", "AviList spells the epithet " + ne +
			" and eBird spells it " + oe + "; the genus is the same in both."
	}
	return "name", "AviList and eBird differ over this name."
}

// fingerprint hashes everything a rendered page depends on besides the URL:
// the dataset, and every template and static file the binary serves.
//
// All of them together rather than one per asset, so a CSS edit invalidates
// the HTML that loads it. Over-invalidating by a little is the safe
// direction -- the cost is one refetch after a deploy, and the alternative
// is a page styled by a stylesheet it was not written for, which is the
// failure that is hard to recognize as a caching problem.
func fingerprint(ds *dataset.Dataset) (string, error) {
	h := sha256.New()
	paths, err := fs.Glob(assetFS, "static/*")
	if err != nil {
		return "", err
	}
	sort.Strings(paths)
	for _, p := range paths {
		b, err := fs.ReadFile(assetFS, p)
		if err != nil {
			return "", err
		}
		fmt.Fprintf(h, "%s\x00%d\x00", p, len(b))
		h.Write(b)
	}
	// The boundary file too: every index page carries its hashed URL, so a
	// new file must mean a new page validator or a cached index would point
	// at a URL that no longer exists (tech.regionBoundaries).
	fmt.Fprintf(h, "boundaries\x00%d\x00", len(data.Boundaries))
	h.Write(data.Boundaries)
	// Marshal rather than a file hash: the server is handed a parsed
	// dataset and never sees the bytes it came from. Go sorts map keys when
	// marshalling, so this is stable across runs.
	b, err := json.Marshal(ds)
	if err != nil {
		return "", err
	}
	h.Write(b)
	return hex.EncodeToString(h.Sum(nil)[:16]), nil
}

// gzipPool reuses compressors. A gzip.Writer carries a ~64 KiB window and a
// hash table, so allocating one per request is the kind of garbage a server
// should not make.
var gzipPool = sync.Pool{New: func() any { return gzip.NewWriter(io.Discard) }}

// gzipResponse compresses what a handler writes.
//
// The headers are decided in WriteHeader, not at the first Write, and that
// ordering is the whole of the type. Setting Content-Encoding from Write is
// too late for any handler that calls WriteHeader itself -- the header map is
// already committed -- so the body goes out compressed and unlabeled, and a
// browser renders binary. Both of BirdView's static files and every page of
// product.undrawnSpeciesAnswer's 404 did exactly that in the first version of this.
type gzipResponse struct {
	http.ResponseWriter
	gz *gzip.Writer
	// on is false for a status that may not carry a body, where emitting a
	// gzip stream would be a protocol error rather than a saving.
	on      bool
	started bool
}

func (g *gzipResponse) start(code int) {
	if g.started {
		return
	}
	g.started = true
	// 204 and 304 carry no body. Compressing nothing would produce the
	// twenty-byte header of an empty gzip stream where the protocol says
	// there must be no bytes at all.
	if code == http.StatusNoContent || code == http.StatusNotModified {
		g.ResponseWriter.WriteHeader(code)
		return
	}
	g.on = true
	// The gzipped body is a different representation of the resource and
	// may not carry the identity representation's validator: a cache that
	// mixed them would eventually hand gzip bytes to a client that asked
	// for none. Suffixed rather than dropped, so revalidation still works
	// for the clients that do send Accept-Encoding (tech.cacheLifetimeStated).
	if et := g.Header().Get("ETag"); strings.HasSuffix(et, `"`) {
		g.Header().Set("ETag", et[:len(et)-1]+`-gz"`)
	}
	// Content-Length describes the uncompressed body and is now a lie.
	// Deleted rather than corrected: the compressed length is not known
	// until the stream closes, and Go will chunk happily without it.
	g.Header().Del("Content-Length")
	g.Header().Set("Content-Encoding", "gzip")
	g.ResponseWriter.WriteHeader(code)
}

func (g *gzipResponse) WriteHeader(code int) { g.start(code) }

func (g *gzipResponse) Write(b []byte) (int, error) {
	g.start(http.StatusOK)
	if !g.on {
		return g.ResponseWriter.Write(b)
	}
	return g.gz.Write(b)
}

func (g *gzipResponse) Flush() {
	if g.on {
		g.gz.Flush()
	}
	if f, ok := g.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

// compressed gzips a response when the client asks for it (tech.compressedResponses).
//
// Worth doing here and not left to a CDN, because the pages are enormous and
// the saving is on the one axis that costs money to serve. The index inlines
// an SVG diagram per species -- 3.3 MB of markup for 1,083 birds -- and
// compresses to 270 KB, a 92% saving. Against Cloud Run's 1 GiB of free
// monthly egress that is the difference between about 320 index loads and
// about 3,980.
//
// Vary is set on every response, not only compressed ones: a cache that
// stored the uncompressed reply without it would go on serving that reply to
// clients that asked for gzip, and worse, the reverse.
func compressed(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Vary", "Accept-Encoding")
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			h.ServeHTTP(w, r)
			return
		}
		gz := gzipPool.Get().(*gzip.Writer)
		defer gzipPool.Put(gz)
		gz.Reset(w)
		g := &gzipResponse{ResponseWriter: w, gz: gz}
		h.ServeHTTP(g, r)
		if g.on {
			// Close only a stream that was opened. Closing an unused writer
			// emits a gzip header for a body nobody wrote.
			gz.Close()
		}
	})
}

// staticFiles serves the files a page loads -- the stylesheet and the
// favicon -- and nothing else from the same directory. The page templates live
// beside them, and a plain file server would serve each one's source as an
// HTML page, and a listing for the bare directory (tech.servesOnlyItsPages).
func staticFiles(assets fs.FS) http.Handler {
	files := http.FileServer(http.FS(assets))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name := r.URL.Path
		if name == "" || strings.HasSuffix(name, "/") || pathpkg.Ext(name) == ".html" {
			http.NotFound(w, r)
			return
		}
		files.ServeHTTP(w, r)
	})
}

// assetFS is where templates and static files are read from.
//
// Embedded for a real binary, and overridable from disk for development. That
// exists because the embed has an edge nobody notices until it bites: a
// running server holds the templates and CSS as they were when it STARTED, so
// editing one and reloading shows the old page with no sign anything is stale.
// It cost three wrong observations in one afternoon -- a recalibrated dataset,
// a recolored favicon and a footer credit, each reported from a page that no
// longer matched the repository.
var assetFS fs.FS = web.Static

// devAssets points the server at the working tree, so an edit shows up on
// reload. Never used by a deployed binary: the flag defaults off and the
// embed is what ships.
func devAssets(dir string) error {
	sub := os.DirFS(dir)
	if _, err := fs.Stat(sub, "static/next.html"); err != nil {
		return fmt.Errorf("-dev %s: no static/next.html below it: %w", dir, err)
	}
	assetFS = sub
	devDir = dir
	return nil
}

// devDir is non-empty in dev mode, and is what makes render re-parse.
var devDir string

func newServer(ds *dataset.Dataset) (*server, error) {
	tmpl, err := template.New("").Funcs(funcs).ParseFS(assetFS, "static/*.html")
	if err != nil {
		return nil, fmt.Errorf("parsing templates: %w", err)
	}
	s := &server{mux: http.NewServeMux(), ds: ds, tmpl: tmpl, devDir: devDir,
		byID: map[string]*dataset.Species{}, undrawnByID: map[string]*dataset.Undrawn{},
		idOf: map[string]string{}}
	// Inlined into the home page rather than fetched, as its other scripts are.
	js, err := fs.ReadFile(assetFS, "static/here.js")
	if err != nil {
		return nil, fmt.Errorf("reading here.js: %w", err)
	}
	s.hereJS = template.JS(js)
	js, err = fs.ReadFile(assetFS, "static/next-find.js")
	if err != nil {
		return nil, fmt.Errorf("reading next-find.js: %w", err)
	}
	s.nextFindJS = template.JS(js)
	s.search = searchTexts(ds)

	// Once, at startup: every response's ETag is built from this (tech.cacheLifetimeStated).
	// In dev mode the templates are re-read per request, so a fingerprint
	// taken now would go stale on the first edit -- which is the precise
	// bug -dev exists to avoid. Left empty there, and revalidated skips.
	if devDir == "" {
		s.build, err = fingerprint(ds)
		if err != nil {
			return nil, fmt.Errorf("fingerprinting the build: %w", err)
		}
	}

	counts := map[string]int{}
	for i := range ds.Species {
		sp := &ds.Species[i]
		s.byID[sp.ID] = sp
		s.idOf[sp.ScientificName] = sp.ID
		counts[sp.Family]++
	}
	s.sameAs = sharedRows(ds)
	// What the home page draws: the same dataset without the exotics, which keep
	// their pages through byID (product.escapedBirdsListed). A shallow copy, so the totals and
	// the undrawn records are shared rather than duplicated.
	s.all = ds
	shown := *ds
	shown.Species = nil
	for i := range ds.Species {
		if ds.Species[i].Exotic == "" {
			shown.Species = append(shown.Species, ds.Species[i])
		}
	}
	s.ds = &shown
	for i := range ds.Undrawn {
		u := &ds.Undrawn[i]
		// The ingest refuses to emit a collision, so one here means the
		// dataset on disk was not built by it. Dropping the undrawn record is
		// the safe direction: a drawn bird keeps its page.
		if _, drawn := s.byID[u.ID]; drawn {
			return nil, fmt.Errorf("id %q is both drawn and not drawn; "+
				"one URL cannot mean two birds (product.undrawnSpeciesAnswer)", u.ID)
		}
		s.undrawnByID[u.ID] = u
	}
	// ds.Families is already in AviList's taxonomic sequence; preserve it
	// rather than alphabetizing, which is the order every field guide uses.
	drawn := map[string]bool{}
	for i := range s.ds.Species {
		drawn[s.ds.Species[i].Family] = true
	}
	for _, f := range ds.Families {
		// A family every one of whose birds is off the home page is not
		// offered: a control that selects nothing is noise (product.escapedBirdsListed).
		if !drawn[f.Name] {
			continue
		}
		s.fams = append(s.fams, familyOption{
			Name: f.Name, Common: f.Common, Color: f.Color, TextColor: f.TextColor,
			Order: f.Order, OrderCommon: f.OrderCommon,
		})
	}

	assets, err := fs.Sub(assetFS, "static")
	if err != nil {
		return nil, err
	}
	sum := sha256.Sum256(data.Boundaries)
	s.boundaryURL = "/boundaries/" + hex.EncodeToString(sum[:8]) + ".json"
	s.routes = []route{
		{"/static/", http.StripPrefix("/static/", staticFiles(assets))},
		{"/added", http.HandlerFunc(s.added)},
		{"/excluded", http.HandlerFunc(s.excluded)},
		// Both the exact path and the subtree, so the bare name serves the
		// guides' index instead of a 301 to it.
		{"/docs", http.HandlerFunc(s.document)},
		{"/docs/", http.HandlerFunc(s.document)},
		{"/birds.json", http.HandlerFunc(s.nextBirds)},
		{"/boundaries/", http.HandlerFunc(s.boundaries)},
		{"/robots.txt", http.HandlerFunc(robots)},
		{"/species/", http.HandlerFunc(s.detail)},
		{"/", http.HandlerFunc(s.next)},
	}
	for _, rt := range s.routes {
		s.mux.Handle(rt.pattern, rt.handler)
	}
	return s, nil
}

// route is one pattern the server answers. Kept as a list, not only
// registered, because the checks enumerate every route to find every web page
// WCAG applies to (tech.webPages), and a ServeMux cannot list its patterns.
type route struct {
	pattern string
	handler http.Handler
}

// options are the user's current selections, all carried in the query string so
// that any view is a shareable URL.
type options struct {
	Metric view.Metric
	// Step is product.sectionThresholdDefault's section-step percentage, used by the three size
	// metrics. Frequency ignores it and bands by Decades instead.
	Step float64
	// Decades is how many orders of magnitude a frequency section spans
	// (product.orderByFrequency): 1 for "1 in 100 to 1 in 999", 2 for a band four times wider.
	// A separate field rather than an overloaded Step, because one float
	// meaning "15 percent" in three views and "one decade" in a fourth is
	// the kind of thing that reads fine and divides wrong.
	Decades int
	// Families holds the explicitly named families. HasSelection says the
	// address names the parameter at all; the home page reads an empty
	// selection as all (nextOptions).
	Families     map[string]bool
	HasSelection bool
	// Regions and HasRegions draw the same distinction for product.regionAndMonthFilter's filter.
	Regions    map[string]bool
	HasRegions bool
	// Months and HasMonths draw the same distinction over the twelve
	// (product.regionAndMonthFilter, decisions.monthPresence). Indexed 0 to 11, January first.
	Months    map[int]bool
	HasMonths bool
	// Habitats and Niches are product.habitatAndDietFilters's categorical filters, and draw the same
	// distinction. A species matches when ANY of its
	// assignments is selected: a bird of two habitats is in both, and
	// requiring all of them would mean selecting two habitats returned only
	// birds that use both, which is not what picking two places asks for.
	Habitats    map[string]bool
	HasHabitats bool
	Niches      map[string]bool
	HasNiches   bool
	Metric3     bool // imperial when false

	// Locale selects which English name is served (product.englishNameChoice). North American is
	// the default: this is an ABA-area dataset and Dovekie is what its reader
	// calls that bird.
	Locale dataset.Locale
}

func parseOptions(r *http.Request) options {
	q := r.URL.Query()
	o := options{Metric: view.ByLength, Step: view.DefaultStepPercent, Decades: 1,
		Locale: dataset.LocaleNA, Families: map[string]bool{}, Regions: map[string]bool{},
		Months:   map[int]bool{},
		Habitats: map[string]bool{}, Niches: map[string]bool{}}
	switch view.Metric(q.Get("by")) {
	case view.ByWingspan:
		o.Metric = view.ByWingspan
	case view.ByMass:
		o.Metric = view.ByMass
	case view.ByBill:
		o.Metric = view.ByBill
	case view.ByTail:
		o.Metric = view.ByTail
	case view.ByFrequency:
		o.Metric = view.ByFrequency
	}
	// One `step` parameter, read differently by the two sectioning rules
	// (decisions.frequencyMetric): under Frequency it is a band width in
	// orders of magnitude, so the value is 10 or 100 rather than a
	// percentage -- and 10 is a legal percentage too, which is why the metric
	// decides first instead of the range of the number.
	if o.Metric == view.ByFrequency {
		if v := q.Get("step"); v == "100" {
			o.Decades = 2
		}
	} else if v, err := strconv.ParseFloat(q.Get("step"), 64); err == nil && v >= 2 && v <= 100 {
		o.Step = v
	}
	o.Metric3 = q.Get("units") == "metric"
	// Recognize the non-default: writing this the other way round disables
	// the setting the moment the default moves.
	if dataset.Locale(q.Get("names")) == dataset.LocaleWorld {
		o.Locale = dataset.LocaleWorld
	}
	_, o.HasSelection = q["family"]
	for _, f := range q["family"] {
		if f != "" {
			o.Families[f] = true
		}
	}
	_, o.HasHabitats = q["habitat"]
	for _, v := range q["habitat"] {
		if v != "" {
			o.Habitats[v] = true
		}
	}
	_, o.HasNiches = q["niche"]
	for _, v := range q["niche"] {
		if v != "" {
			o.Niches[v] = true
		}
	}
	_, o.HasRegions = q["region"]
	for _, g := range q["region"] {
		// A URL carries slugs ("CA", "GulfOfMexico"); everything downstream
		// works in GADM ids, so translate here and nowhere else (decisions.regionUrlIsPostalCode).
		// An unknown slug is dropped rather than guessed at.
		if rg, ok := region.BySlug(g); ok {
			o.Regions[rg.ID] = true
		}
	}
	_, o.HasMonths = q["month"]
	for _, m := range q["month"] {
		// "Jun", not 6: a month number in a URL is off by one as often as
		// not, and an unknown abbreviation is dropped as a region slug is.
		if i, ok := month.Index(m); ok {
			o.Months[i] = true
		}
	}
	return o
}

// tile is one rendered species. The three footer metrics are formatted here
// rather than in the template: a template that does its own arithmetic cannot
// see the unit toggle, which is exactly how the footer came to ignore it.
type tile struct {
	S *dataset.Species
	// Name is the species' English name in the reader's chosen locale,
	// resolved once here so no template has to know about locales (product.englishNameChoice).
	Name    string
	Geom    view.Geometry
	Box     view.Box
	Length  string
	Wing    string
	MassStr string
	// Commonest is the share line: where the bird is the biggest part of
	// what gets recorded, and how big a part (product.shareOfRecords). Nil when unknown.
	Commonest *shareCell
	// SameAs names the other drawn species with the same measurements
	// (product.borrowedMeasurementsSaySo). Outside Commonest, because it is about the drawing and must
	// show whether or not there is a count.
	SameAs []seam

	// WingModeled drives the crossbar split: the inner span is dashed only when
	// the wingspan is modeled (product.measuredVersusModeled).
	WingModeled bool

	// Labels and ChordLabel are empty on a home page tile: the annotations of product.labeledDetailDiagram
	// need room, so they appear only in the detail view, which builds this
	// struct through labeledTile.
	Labels     []diagramLabel
	TopOffset  float64
	ChordLabel string
	// WingspanLabel sits at the right wingtip, mirroring the chord label at the
	// left, so the crossbar names the whole span as well as its measured part.
	WingspanLabel string
	WingspanEst   bool
	// FootLabels name the whole bird rather than any part of it, stacked on
	// the drawing axis below the tail tip (product.labeledDetailDiagram): its total length and its
	// mass. A slice rather than named fields because product.labeledDetailDiagram requires every
	// metric the detail view reports to appear in the drawing, so the next
	// whole-bird metric should be one more entry and not a new layout.
	//
	// Total length is the bird's reported length, never the drawn height: a
	// folded neck makes those differ by 6-9%.
	FootLabels []footLabel

	// MassCube is the bird's mass drawn as water: one face of the cube of
	// water that weighs the same (product.labeledDetailDiagram, decisions.massAsWaterCube). Side zero means none.
	MassCube massCube

	// DetailURL is the tile's own link. Built in Go rather than in the
	// template so it carries the units and names the species page draws with
	// (product.nextView).
	DetailURL template.URL

	// PinURL pins the bird on the home page, or unpins it when Pinned
	// (product.nextView). Empty everywhere else.
	PinURL template.URL
	Pinned bool
	// Anchor is this tile's fragment id, which the pinned list links to.
	Anchor string
}

// tileAnchor is the fragment id of one tile. Declared once because the href
// and the element's id must agree, and they are written in different files.
func tileAnchor(id string) string { return "s-" + id }

// sectionView is one tier as the histogram reads it: its number, its range
// as written, and how many birds it holds.
type sectionView struct {
	Index    int
	Min, Max string
	Count    int
	// Note replaces the range where there is no range to state: the birds
	// with no records in the regions selected have no ratio and so no tier
	// (product.orderByFrequency).
	Note string
}

// frequencyBand labels a decade section (product.orderByFrequency).
//
// The label states the BAND, not the extremes of the birds in it: a tier
// reading "1 in 1,000 - 1 in 9,999" means the same thing on every page,
// where "1 in 1,184 - 1 in 9,781" would be a different pair of numbers under
// every region filter and would invite a reader to compare two tiers that
// are not the same tier. product.sectionHeadersStateRange asks for the range the section covers, and
// under this rule the band IS that range.
func frequencyBand(min, max float64, decades int) (lo, hi, note string) {
	if math.IsInf(min, 1) {
		return "", "", "No records in the regions selected"
	}
	k := int(math.Floor(math.Log10(min)))
	if decades > 1 {
		k -= ((k % decades) + decades) % decades
	}
	start := math.Pow(10, float64(k))
	end := view.Decade(min, decades) - 1
	// The last band runs off the end of the data rather than to a boundary
	// nothing reaches, so it says so instead of naming an empty ceiling.
	if math.IsInf(max, 1) || end > 1e12 {
		return "1 in " + comma(int64(start)), "", ""
	}
	return "1 in " + comma(int64(start)), "1 in " + comma(int64(end)), ""
}

// tileBox is the drawing area for one diagram. Roughly 2.6:1, because a
// to-scale wingspan runs up to ~2.5x body length and product.crossbarAtVerticalScale draws it at the
// vertical scale rather than shrinking it to fit.
var tileBox = view.Box{W: 380, H: 148}

// filterSpecies applies every filter the address carries to the home page's
// birds, and returns the ones every active filter admits, in dataset order,
// with the families selected (product.filtersCompose, product.nextView).
func (s *server) filterSpecies(opt options) ([]*dataset.Species, map[string]bool) {
	// No family named selects every family. The home page reads an emptied
	// selection as none named (nextOptions).
	selected := opt.Families
	if !opt.HasSelection {
		selected = make(map[string]bool, len(s.fams))
		for _, f := range s.fams {
			selected[f.Name] = true
		}
	}
	// Region works the same way, and composes with family (product.filtersCompose): a bird is
	// shown when every active filter admits it.
	inRegion := func(sp *dataset.Species) bool {
		if !opt.HasRegions {
			return true
		}
		for _, id := range sp.Regions {
			if opt.Regions[id] {
				return true
			}
		}
		return false
	}
	// Months narrow the list as every other filter does (product.regionAndMonthFilter, decisions.monthPresence).
	// A species is shown when SOME selected region admits it in SOME
	// selected month, which is the same "any" the region and habitat
	// filters use: picking two places asks for birds in either.
	allMonths := !opt.HasMonths || len(opt.Months) == month.Count
	inMonth := func(sp *dataset.Species) bool {
		if allMonths {
			return true
		}
		for i, id := range sp.Regions {
			if opt.HasRegions && !opt.Regions[id] {
				continue
			}
			if i >= len(sp.RegionMonths) {
				continue
			}
			// A month eBird withholds this species here is missing by
			// policy, so it is not judged: the species is shown, and its
			// tile says why (decisions.ebirdSilenceNotAbsence). Painted Bunting was hidden from
			// Florida in January this way -- eBird publishes none of its
			// Florida records from October to March.
			withheld := sp.WithheldIn(id)
			for m := range withheld {
				if withheld[m] && opt.Months[m] {
					return true
				}
			}
			in, ok := month.Present(sp.RegionMonths[i], s.ds.RegionMonthlyTotals[id])
			if !ok {
				// Too thin to say when. Shown in every month rather than
				// hidden from any, and the tile says so (decisions.monthPresence).
				return true
			}
			for m := range in {
				if in[m] && opt.Months[m] {
					return true
				}
			}
		}
		return false
	}
	// Habitat and niche compose the same way (product.habitatAndDietFilters, product.filtersCompose). A species
	// matches when ANY of its assignments is selected, and the assignment's
	// origin is irrelevant here: a habitat BirdView added is still a habitat
	// the bird uses, and filtering it out would make the addition pointless.
	anyAssigned := func(as []dataset.Assignment, want map[string]bool, active bool) bool {
		if !active {
			return true
		}
		for _, a := range as {
			if want[a.Value] {
				return true
			}
		}
		return false
	}
	var filtered []*dataset.Species
	for i := range s.ds.Species {
		sp := &s.ds.Species[i]
		if selected[sp.Family] && inRegion(sp) && inMonth(sp) &&
			anyAssigned(sp.Habitats, opt.Habitats, opt.HasHabitats) &&
			anyAssigned(sp.Niches, opt.Niches, opt.HasNiches) {
			filtered = append(filtered, sp)
		}
	}
	return filtered, selected
}

type detailData struct {
	S *dataset.Species
	// Name is the species' English name in the reader's locale (product.englishNameChoice).
	Name    string
	Geom    view.Geometry
	Box     view.Box
	Metrics []metricRow
	// Categories is every group the bird is sorted into, with who decided
	// each (product.categoryProvenance).
	Categories []categoryRow
	// SameAs names the other drawn species with the same measurements
	// (product.borrowedMeasurementsSaySo).
	SameAs []seam
	Opt    options
	// Home is the home page with the address's settings, where Back goes
	// from a shared link or without JavaScript (product.nextView).
	Home template.URL

	// INatURL is empty when no confident iNaturalist match exists, and the
	// template then renders nothing at all rather than a placeholder (product.inaturalistLink).
	INatURL string

	// WingModeled drives the crossbar split: the inner span is dashed only when
	// the wingspan is modeled (product.measuredVersusModeled).
	WingModeled bool

	// FoldNote explains why a folded-neck bird's figure is shorter than the total length labeled
	// beneath it, with its own numbers; empty for a bird drawn at full extent
	// (product.labeledDetailDiagram).
	FoldNote string

	// Labels annotate the detail diagram (product.labeledDetailDiagram); empty on a tile.
	Labels        []diagramLabel
	TopOffset     float64
	ChordLabel    string
	WingspanLabel string
	WingspanEst   bool
	FootLabels    []footLabel
	MassCube      massCube
}

// diagramLabel is one annotation: where view put it, what it says, and whether
// the value it reports is estimated.
type diagramLabel struct {
	view.Label
	Text string
	Est  bool
}

type metricRow struct {
	Label, Value, Kind, Detail string
}

func (s *server) detail(w http.ResponseWriter, r *http.Request) {
	wanted := strings.TrimPrefix(r.URL.Path, "/species/")
	sp, ok := s.byID[wanted]
	if !ok {
		// A bird the derivation classified and left out of scope gets a page
		// naming it and the measured reason, at 404 (product.undrawnSpeciesAnswer, decisions.speciesUrlOutlivesSpecies). Anything
		// else gets the bare 404: this page is for a bird BirdView knows of
		// and chose not to draw, and it would be a fabrication for a typo.
		if u, known := s.undrawnByID[wanted]; known {
			s.undrawn(w, r, u)
			return
		}
		http.NotFound(w, r)
		return
	}
	opt := parseOptions(r)
	// Wider than the tile's 2.6:1 so the labels of product.labeledDetailDiagram have margin outside the
	// wingtips. The extra width costs nothing: the scale is already limited by
	// height for almost every bird, so the bird does not grow, the margin does.
	box := view.Box{W: tileBox.W * 2.6, H: tileBox.H * 1.9}
	// The detail view always draws the bird as large as the box allows, whatever
	// scale the home page drew it at. Beside a swan a small bird is a few
	// pixels tall by design, and carrying that into the one view devoted to
	// that single bird would leave nothing to look at (product.tileOpensDetail).

	// A detail page compares nothing, so the bird is its own scope and the
	// unit is whichever suits it.
	one := unitsFor(opt.Metric3, sp.Mass.Value)
	rows := []metricRow{
		row("Total length", sp.TotalLength, view.ByLength, one),
		row("Wingspan", sp.Wingspan, view.ByWingspan, one),
		row("Wing chord", sp.WingChord, view.ByLength, one),
		row("Mass", sp.Mass, view.ByMass, one),
		row("Bill", sp.Segments.Bill, view.ByLength, one),
		row("Head", sp.Segments.Head, view.ByLength, one),
		row("Neck", sp.Segments.Neck, view.ByLength, one),
		row("Body", sp.Segments.Body, view.ByLength, one),
		row("Tail", sp.Segments.Tail, view.ByLength, one),
	}
	lt := labeledTile(sp, box, one, opt.Locale)

	// Only a confident match gets a link; zero means none was found (product.inaturalistLink).
	var inatURL string
	if sp.INatTaxonID > 0 {
		inatURL = inat.TaxonURL(sp.INatTaxonID)
	}
	s.render(w, "detail.html", detailData{
		S: sp, Name: lt.Name, Geom: lt.Geom, Box: lt.Box, Labels: lt.Labels, TopOffset: lt.TopOffset,
		ChordLabel: lt.ChordLabel, WingspanLabel: lt.WingspanLabel, WingspanEst: lt.WingspanEst,
		FootLabels: lt.FootLabels, MassCube: lt.MassCube,
		Metrics: rows, Categories: s.categories(sp, opt), Opt: opt, Home: homeLink(r), INatURL: inatURL,
		SameAs:      s.sameMeasurements(sp, opt.Locale),
		WingModeled: sp.Wingspan.Prov.Kind == dataset.Modeled,
		FoldNote:    foldNote(sp, lt.Geom, one),
	})
}

// foldNote is the sentence that explains a folded neck on the detail page, or "" when the neck is
// drawn straight. The figure of a folded-neck bird spans less height than its total length
// (decisions.foldedNeckAsS), and the page labels that total length beneath it, so without this a
// careful reader sees two numbers that disagree.
func foldNote(sp *dataset.Species, g view.Geometry, u units) string {
	group := view.FoldedNeckGroup(sp.Family)
	if !g.NeckFolded || group == "" {
		return ""
	}
	return fmt.Sprintf("%s fly with the neck folded, so the %s neck is drawn as an S and the figure is "+
		"shorter than the bird's total length of %s.",
		group, u.format(sp.Segments.Neck.Value, view.ByLength), u.format(sp.TotalLength.Value, view.ByLength))
}

// undrawnData is the page for a bird the derivation rejected (product.undrawnSpeciesAnswer).
//
// The prose is assembled in Go rather than branched in the template, because
// the three verdicts say genuinely different things and a template that tries
// to interleave them ends up asserting the intersection — which for `absent`
// would be counts that were never measured.
type undrawnData struct {
	Feedback string
	Home     template.URL
	Opt      options

	Name string
	// AlsoCalled is the other locale's English name, present only when it
	// differs. A reader who followed a stale link may have it under the other
	// one (product.englishNameChoice).
	AlsoCalled string
	U          *dataset.Undrawn
	// Verdict is the one-line answer, and Why the paragraph behind it.
	Verdict string
	Why     string
	// Evidence is the measured sentence, empty for `absent` — where there is
	// no measurement, and printing the zeros would present the absence of a
	// measurement as a measurement of absence.
	Evidence string
}

// undrawn renders the page for a species the status table classified and
// scope rejected, at 404 (product.undrawnSpeciesAnswer, decisions.speciesUrlOutlivesSpecies).
//
// 404 and not 410: scope is regenerated from live occurrence data, so a bird
// that misses the recurrence bar on one run can clear it on the next, and
// "gone permanently" is a claim this derivation does not support.
//
// Verifies: product.undrawnSpeciesAnswer.
func (s *server) undrawn(w http.ResponseWriter, r *http.Request, u *dataset.Undrawn) {
	opt := parseOptions(r)
	name := u.Name(opt.Locale)
	other := u.Names[dataset.LocaleNA]
	if opt.Locale == dataset.LocaleNA {
		other = u.Names[dataset.LocaleWorld]
	}
	if other == name {
		other = ""
	}

	where := u.Region
	if rg, ok := region.Get(u.Region); ok {
		where = rg.Name
	}
	// Both thresholds are read from the derivation rather than written out
	// here. A page quoting "three episodes across three years" beside a rule
	// that had moved to four would be a confident, checkable lie.
	bar := fmt.Sprintf("%d episodes across %d years", status.MinEpisodes, status.MinYears)
	recurs := fmt.Sprintf("%s episodes across %s years in %s, from %s records",
		comma(int64(u.Episodes)), comma(int64(u.Years)), where, comma(u.Records))

	d := undrawnData{
		Feedback: s.feedback, Home: homeLink(r), Opt: opt,
		Name: name, AlsoCalled: other, U: u,
	}
	switch u.Status {
	case string(status.Vagrant):
		d.Verdict = "it has turned up, but it does not come back"
		d.Evidence = "Recorded in " + recurs + "."
		d.Why = "BirdView draws the birds whose records recur, at least " + bar +
			" in one region. This bird has not reached that."
	case string(status.Historical):
		d.Verdict = "its records recur, but they stopped"
		d.Evidence = "Recorded in " + recurs + ", the last of them on " + prettyDate(u.LastSeen) + "."
		d.Why = fmt.Sprintf("A bird last recorded more than %d years before the records "+
			"end is counted as historical, not present.", status.RecentYears)
	default:
		d.Verdict = "the open record holds nothing that can be measured"
		d.Why = "There are too few dated records of it, anywhere BirdView covers, to judge."
	}

	// The body says what the status code says, and the status code is set
	// before a byte of it is written.
	w.WriteHeader(http.StatusNotFound)
	s.render(w, "undrawn.html", d)
}

// prettyDate turns an ISO date from the dataset into one a reader reads.
// Falls back to the input, which is already legible, rather than to a blank.
func prettyDate(iso string) string {
	t, err := time.Parse("2006-01-02", iso)
	if err != nil {
		return iso
	}
	return t.Format("2 January 2006")
}

// homeLink is the home page with the request's settings (product.nextView).
func homeLink(r *http.Request) template.URL {
	return nextURL(r.URL.Query())
}

// labelGutter is the width reserved at the right of an annotated diagram: the
// segment label column, plus room for the wingspan label that sits just beyond
// the wingtip.
//
// The bird is laid out in the remaining width, so a wide-winged bird is drawn
// slightly smaller rather than running under its own labels. Before this, a
// Black-footed Albatross's wingtip reached x=988 in a 988-wide box while the
// label column started at 838, so the crossbar crossed every label on that
// side.
// 320 rather than the 290 that merely fits: at 290 the widest-winged species
// clear the column by 2px, which is not clearance, it is luck.
const labelGutter = 320

// The strip reserved at the bottom of an annotated diagram for the whole-bird
// labels (product.labeledDetailDiagram), the vertical counterpart of labelGutter.
//
// Without it they are clipped rather than merely cramped: the detail view
// scales the bird to fill the box, so any bird whose height binds has its tail
// tip exactly on the bottom edge and nothing at all below it.
//
// footGap is where the first baseline sits below the tail tip and footLine is
// the spacing between them; footDescender keeps the glyphs of the last line
// inside the viewBox, not merely its baseline.
const (
	footGap       = 15
	footLine      = 13
	footDescender = 9
)

// footLabel is one whole-bird annotation under the tail.
type footLabel struct {
	Key  string // total, mass -- also the CSS hook
	Text string
	Y    float64
	Est  bool
}

// massCube is the cube of water that weighs what the bird weighs, drawn in the
// lower-left in oblique projection (product.labeledDetailDiagram, decisions.massAsWaterCube).
//
// X,Y is the top-left of the FRONT FACE, and the face's base sits on the
// bottom of the drawing area in every diagram. That is what makes two cubes
// comparable at a glance: the projection grows up and to the right from a
// corner that never moves, so only the cube's size differs between birds.
type massCube struct {
	X, Y, Side, Depth float64
	// The mass label sits centered under the face rather than in the stack
	// under the tail, so the number is beside the shape it describes.
	Label          string
	LabelX, LabelY float64
	Est            bool // mass not measured -- dashed, per product.measuredVersusModeled
	Title          string
}

// massCubeDepth is the oblique projection's depth, as a fraction of the side,
// offset at 45 degrees.
//
// Cabinet projection's conventional 0.5 also fits -- measured over all 841
// species, neither value lets the cube reach the crossbar -- but it leaves
// only 14px of clearance on Western Grebe against this one's 21px, and the
// unit tests cannot see that margin because the fixture holds five birds and
// no grebe. The looser value is the one that survives a data change.
const massCubeDepth = 0.4

// massLabelGap is how far the mass label's baseline sits below the face.
//
// Wider than the footGap the stack under the tail uses, and deliberately: the
// tail is a hairline and the cube is a solid block, so the same gap that looks
// airy under one looks cramped under the other.
const massLabelGap = 18

// massLabelCharWidth approximates one character of the 11px label, used only
// to stop a small cube's centered label running off the left edge of the
// drawing -- a Scissor-tailed Flycatcher's cube is 29px wide and its label
// about 70px, which rendered as "ass 1.4 oz".
//
// Deliberately generous: overestimating only nudges the label right, while
// underestimating clips it again.
const massLabelCharWidth = 6.2

// massCubeSide is the side of that cube, in millimeters, for a mass in grams.
//
// One gram of water is one cubic centimeter, by the historical definition of
// the gram, so this is a unit conversion and not a model: it assumes nothing
// about the density of birds. That is what lets the square share the diagram's
// own scale instead of needing one of its own, which product.crossbarAtVerticalScale would refuse.
func massCubeSide(massG float64) float64 {
	if massG <= 0 {
		return 0
	}
	return 10 * math.Cbrt(massG) // cm -> mm
}

// massCubeMargin insets the cube from the corner of the drawing area.
const massCubeMargin = 8

// newMassCube places the cube of water in the lower-left of a drawing area
// drawH tall, at pxPerMm (product.labeledDetailDiagram, decisions.massAsWaterCube).
//
// label is empty on a home page tile, which carries no labels at all and names
// the mass in its footer instead (product.tileContents, decisions.massCubeOnTiles). Everything else about the
// cube is identical there, including the face standing on the floor, so the
// two surfaces cannot drift.
func newMassCube(sp *dataset.Species, drawH, pxPerMm float64, label string) massCube {
	side := massCubeSide(sp.Mass.Value) * pxPerMm
	if side <= 0 {
		return massCube{}
	}
	base := drawH - massCubeMargin
	c := massCube{
		X: massCubeMargin, Y: base - side, Side: side, Depth: side * massCubeDepth,
		Est:   sp.Mass.Prov.Kind != dataset.Measured,
		Title: "Mass as water: a cube of water weighing the same, to scale",
	}
	if label != "" {
		c.Label = label
		// Centered under the face, but never far enough left to be clipped.
		c.LabelX = math.Max(massCubeMargin+side/2, float64(len(label))*massLabelCharWidth/2)
		c.LabelY = base + massLabelGap
	}
	return c
}

// footStrip is the height n stacked whole-bird labels need.
func footStrip(n int) float64 {
	if n == 0 {
		return 0
	}
	return footGap + float64(n-1)*footLine + footDescender
}

// labeledTile builds the annotated diagram for a bird drawn at its own scale:
// the detail view, which shows one bird with nothing to measure it against
// (product.tileOpensDetail).
func labeledTile(sp *dataset.Species, box view.Box, u units, loc dataset.Locale) tile {
	// draw is the area the bird gets; the emitted box is that plus the strip the
	// total length label sits in, so the label costs the bird nothing. Going the
	// other way -- keeping the box and shrinking the bird into it -- would take
	// 8.5% of the detail view's height and contradict product.tileOpensDetail, which requires the
	// bird drawn as large as the view allows. Width has no such freedom, which
	// is why labelGutter does shrink the bird and this does not.
	// Every metric the detail view reports has to appear in the drawing
	// (product.labeledDetailDiagram). The two that name the whole bird rather than a part of it go
	// under the tail; mass is here and not in the stack because it is not a
	// length and has no part to point at.
	//
	// The reported total length, never g.TotalHeight(): for a folded neck the
	// drawing is shorter than the bird (decisions.foldedNeckAsS), and the label names the bird.
	foot := []footLabel{{
		Key:  "total",
		Text: "Total length  " + u.format(sp.TotalLength.Value, view.ByLength),
		Est:  sp.TotalLength.Prov.Kind != dataset.Measured,
	}}

	draw := box
	// The strip has to clear whichever reaches lower: the stack under the tail
	// or the mass label under the cube.
	box.H += math.Max(footStrip(len(foot)), massLabelGap+footDescender)

	// The bird occupies everything left of the gutter; draw stays the full
	// height, because the label column is positioned against the emitted box.
	inner := view.Box{W: draw.W - labelGutter, H: draw.H}
	g := view.Layout(sp, sp.TotalLength.Value, sp.Wingspan.Value, inner)
	top := (draw.H - g.TotalHeight()) / 2
	segOf := map[string]dataset.Metric{
		"bill": sp.Segments.Bill, "head": sp.Segments.Head, "neck": sp.Segments.Neck,
		"body": sp.Segments.Body, "tail": sp.Segments.Tail,
	}
	t := tile{S: sp, Name: sp.Name(loc), Geom: g, Box: box, TopOffset: top,
		WingModeled: sp.Wingspan.Prov.Kind == dataset.Modeled}
	// draw, not box: the segment label column belongs in the drawing area, and
	// must not spill into the strip the total length label owns.
	for _, l := range g.DetailLabels(draw, top) {
		m := segOf[l.Key]
		t.Labels = append(t.Labels, diagramLabel{
			Label: l,
			Text:  strings.ToUpper(l.Key[:1]) + l.Key[1:] + "  " + u.format(m.Value, view.ByLength),
			Est:   m.Prov.Kind == dataset.Modeled,
		})
	}
	if sp.WingChord.Value > 0 {
		t.ChordLabel = "Wing chord  " + u.format(sp.WingChord.Value, view.ByLength)
	}
	t.WingspanLabel = "Wingspan  " + u.format(sp.Wingspan.Value, view.ByWingspan)
	t.WingspanEst = sp.Wingspan.Prov.Kind == dataset.Modeled
	for i := range foot {
		foot[i].Y = top + g.YTailTip + footGap + float64(i)*footLine
	}
	t.FootLabels = foot

	// Lower-left of the DRAWING area, not the emitted box: the box is taller by
	// the foot strip and wider by the label gutter, and the square belongs
	// inside the bird's own area. Measured over the whole dataset before this
	// was built -- the crossbar reaches into that corner for 0 of 841 species,
	// the tightest clearance being 50px on Western Grebe.
	t.MassCube = newMassCube(sp, draw.H, g.PxPerMm,
		"Mass  "+u.format(sp.Mass.Value, view.ByMass))
	return t
}

// row turns a metric into a display line, stating its provenance in plain
// language (product.metricProvenanceExplained).
func row(label string, m dataset.Metric, as view.Metric, u units) metricRow {
	r := metricRow{Label: label, Value: u.format(m.Value, as), Kind: string(m.Prov.Kind)}
	switch m.Prov.Kind {
	case dataset.Measured:
		r.Detail = "measured, from " + m.Prov.Source
		if m.Prov.SourceTaxon != "" {
			r.Detail += " (" + m.Prov.SourceTaxon + ")"
		}
	case dataset.NearMatch:
		r.Detail = "measured on " + m.Prov.SourceTaxon +
			", a different species this one was split from \u2014 no measurement exists for it. Source: " +
			m.Prov.Source
	default:
		r.Detail = "estimated: " + m.Prov.Model
	}
	if m.Min > 0 && m.Max > 0 {
		r.Detail += fmt.Sprintf(" Reported range %s\u2013%s.",
			u.format(m.Min, as), u.format(m.Max, as))
	}
	return r
}

func (s *server) render(w http.ResponseWriter, name string, data any) {
	tmpl := s.tmpl
	// In dev mode the templates are re-parsed per request, which is the whole
	// point: parsing once at startup is exactly how an edited template goes on
	// serving its old self until someone restarts. Serving from a directory
	// without this fixes the CSS and leaves the HTML stale, which is worse
	// than not having the flag -- it looks like it works.
	if s.devDir != "" {
		t, err := template.New("").Funcs(funcs).ParseFS(assetFS, "static/*.html")
		if err != nil {
			log.Printf("birdview: dev re-parse failed, serving the last good templates: %v", err)
		} else {
			tmpl = t
		}
	}
	if s.onRender != nil {
		s.onRender(name)
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.ExecuteTemplate(w, name, data); err != nil {
		log.Printf("birdview: rendering %s: %v", name, err)
	}
}

// format renders a value in the selected unit system.
//
// Mass in imperial means ounces and pounds. An earlier version returned
// kilograms above a kilogram regardless of the toggle, so "imperial" showed a
// metric unit for exactly the birds where it mattered most.
// gramsPerOunce converts to the avoirdupois ounce.
const gramsPerOunce = 28.349523125

// units fixes how every number in one comparison is written.
//
// The mass unit is a property of the SCOPE, not of the value. Chosen per
// value, a section could print "14.5 oz" beside "1.2 lb" for two birds of
// the same length -- arithmetic the reader has to do, on the page whose
// whole purpose is comparison. Length never switches, so only mass needs
// this.
//
// The scope is what the page is comparing: a section or the whole set on the
// index, depending on the scale setting, and the birds named in the address
// on the comparison page. A detail page compares nothing, so its scope is
// the one bird.
type units struct {
	// metric selects centimeters and grams over inches and ounces.
	metric bool
	// big selects kilograms or pounds over grams or ounces.
	big bool
}

// unitsFor picks the unit a scope shares.
//
// The SMALLEST mass decides, so a scope holding one small bird is written in
// the small unit throughout: a swan beside a sparrow reads as "4.2 lb" and
// "0.9 oz" otherwise, and the two cannot be compared by eye at all. Writing
// the swan as 67 oz keeps one scale across the set, which is the point.
// Taking the largest instead would round every small bird to 0.0 lb.
func unitsFor(metric bool, masses ...float64) units {
	u := units{metric: metric, big: true}
	seen := false
	for _, m := range masses {
		if m <= 0 {
			continue // no mass recorded; it cannot decide the unit
		}
		seen = true
		if metric && m < 1000 {
			u.big = false
		}
		if !metric && m/gramsPerOunce < 16 {
			u.big = false
		}
	}
	if !seen {
		u.big = false
	}
	return u
}

func (u units) format(v float64, as view.Metric) string {
	return u.formatDigits(v, as, 0)
}

// histBar is one tier's bar in the histogram under the count line
// (product.tierHistogram): its height as a percentage of the largest tier's
// count, and a label saying which tier it is. An Empty bar is a step of the
// scale that no bird falls in.
type histBar struct {
	Index, Percent int
	Label          string
	Empty          bool
}

// histogram builds one bar per section, each labeled as its heading reads.
// Under the size orderings every tier is one step of the scale, so a stretch
// with no birds becomes one empty bucket per whole step it spans: without
// them two tiers far apart would stand side by side as if they abutted.
// Frequency's bands are fixed and need none.
func histogram(metric string, secs []sectionView, raw []view.Section, m view.Metric, stepPercent float64) []histBar {
	largest := 1 // not 0: a search can hide every bird, and a bar is then 0%, not NaN
	for _, sv := range secs {
		largest = max(largest, sv.Count)
	}
	out := make([]histBar, 0, len(secs))
	step := 1 + stepPercent/100
	for i, sv := range secs {
		if m != view.ByFrequency && i > 0 && i < len(raw) {
			gap := math.Floor(math.Log(raw[i].Min/(raw[i-1].Min*step)) / math.Log(step))
			for k := 0; k < int(gap); k++ {
				out = append(out, histBar{Empty: true})
			}
		}
		var span string
		switch {
		case sv.Note != "":
			span = sv.Note
		case sv.Min == sv.Max:
			span = sv.Min
		case sv.Max != "":
			span = sv.Min + " – " + sv.Max
		default:
			span = sv.Min + " and beyond"
		}
		out = append(out, histBar{Index: sv.Index,
			Percent: int(math.Round(100 * float64(sv.Count) / float64(largest))),
			Label:   fmt.Sprintf("%s Tier %d: %s, %d species", metric, sv.Index, span, sv.Count)})
	}
	return out
}

// tierSpan is a section heading's range before it is written: the tier's
// smallest and largest bird, in the section's units.
type tierSpan struct {
	lo, hi float64
	u      units
}

// headingDigits is the fewest extra decimals, up to three, that keep every
// stated value apart when all the page's headings share them: each tier's two
// ends, and each tier's end from the next tier's start, which is always
// higher. One decimal of an inch is 2.5 mm, and a tier of small bills is
// narrower than that; written alike, two neighbors would read as sharing a
// bird (product.sectionHeadersStateRange).
func headingDigits(spans []tierSpan, m view.Metric) int {
	type val struct {
		v float64
		u units
	}
	var seq []val
	for _, sp := range spans {
		seq = append(seq, val{sp.lo, sp.u}, val{sp.hi, sp.u})
	}
	extra := 0
	for ; extra < 3; extra++ {
		apart := true
		for i := 1; i < len(seq); i++ {
			a, b := seq[i-1], seq[i]
			if a.v != b.v && a.u.formatDigits(a.v, m, extra) == b.u.formatDigits(b.v, m, extra) {
				apart = false
				break
			}
		}
		if apart {
			break
		}
	}
	return extra
}

// formatDigits is format with extra decimal places.
func (u units) formatDigits(v float64, as view.Metric, extra int) string {
	f := func(decimals int, x float64, unit string) string {
		return strconv.FormatFloat(x, 'f', decimals+extra, 64) + " " + unit
	}
	if as == view.ByMass {
		if u.metric {
			if u.big {
				return f(2, v/1000, "kg")
			}
			return f(0, v, "g")
		}
		oz := v / gramsPerOunce
		switch {
		case u.big:
			return f(1, oz/16, "lb")
		case oz >= 1:
			return f(1, oz, "oz")
		default:
			// A 3 g hummingbird is 0.11 oz; one decimal would erase the
			// difference between the smallest birds entirely. This is a
			// precision choice within one unit, not a second unit, so it
			// stays per value.
			return f(2, oz, "oz")
		}
	}
	if u.metric {
		return f(1, v/10, "cm")
	}
	return f(1, v/25.4, "in")
}

// hereSlot is the placeholder the browser replaces with the reader's region
// in the Here button's and the welcome line's templates. URL encoding leaves
// it alone once restored, and no region slug contains it.
const hereSlot = "__REGION__"

// nowSlot is the placeholder the browser replaces with its own month ("Oct")
// in the welcome line's template, as the month= value.
const nowSlot = "__MONTH__"

// boundaries serves the region boundary file at its hashed URL, and only
// there. The URL changes whenever the file does, so it is cached for a year
// (tech.cacheLifetimeStated's one exception).
func (s *server) boundaries(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != s.boundaryURL {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	w.Write(data.Boundaries)
}

// robotsPolicy asks crawlers to skip every URL with a query string, which is
// every filtered view of the home page, and the route that is no page to index.
// Everything else stays open, /static/ included, since a crawler renders a
// page with its stylesheet. Wildcards are RFC 9309's.
const robotsPolicy = `# BirdView: every control is a query parameter, so filtered views are endless.
# Index the pages without one: the home page, each species, the guides.
User-agent: *
Disallow: /*?
Disallow: /boundaries/
`

// robots serves robotsPolicy (tech.crawlersSkipFilteredViews).
func robots(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	io.WriteString(w, robotsPolicy)
}

var funcs = template.FuncMap{
	"f1":  func(v float64) string { return strconv.FormatFloat(v, 'f', 1, 64) },
	"add": func(a, b float64) float64 { return a + b },
	"sub": func(a, b float64) float64 { return a - b },
	"div": func(a, b float64) float64 { return a / b },
	// whynot explains, in a tooltip, why a value is not this bird's own
	// measurement. The three footer cells used to say "estimated" for
	// anything not measured, which is false for a near-match and said so
	// right beside a stroke product.measuredVersusModeled draws SOLID because it IS a measurement.
	"whynot": whyNotMeasured,
}

// whyNotMeasured turns a provenance into a short tooltip.
//
// It names the donor for a near-match, because that is the fact a reader can
// act on: two tiles showing an identical mass are not a coincidence when one
// of them borrowed it from the other, and nothing else on the home page says so.
//
// Verifies: product.provenancePerMetric, product.measuredVersusModeled.
func whyNotMeasured(p dataset.Provenance) string {
	switch p.Kind {
	case dataset.NearMatch:
		if p.SourceTaxon != "" {
			return "measured, but on " + p.SourceTaxon + " \u2014 not this species"
		}
		return "measured on a related species, not this one"
	case dataset.Modeled:
		if p.Model != "" {
			return "estimated: " + p.Model
		}
		return "estimated"
	}
	return "estimated"
}

// categoryRow is one line of the detail view's category table (product.categoryProvenance): what
// the bird is sorted as, and who decided.
type categoryRow struct {
	// Label is blank on a continuation row, so two habitats read as one
	// entry with two values rather than as two unrelated facts.
	Label  string
	Value  string
	Origin string
	Detail string
	// Doc, when set, is a document BirdView serves that explains the rule.
	Doc, DocLabel string
}

// categories lists every category a species is in, with its provenance.
//
// **Regions are one row, not sixty.** Every region a bird is listed in was
// decided by the same rule over the same data, so a row apiece would repeat
// one sentence sixty times and bury the two facts that differ -- habitat and
// diet, where BirdView sometimes disagrees with AVONET. The regions
// themselves are listed inline by their postal codes.
func (s *server) categories(sp *dataset.Species, opt options) []categoryRow {
	loc := opt.Locale
	out := []categoryRow{
		{Label: "Family", Value: sp.Family + " — " + sp.FamilyCommon,
			Origin: string(dataset.Published), Detail: "AviList v2025b, CC BY 4.0"},
		{Label: "Order", Value: sp.Order,
			Origin: string(dataset.Published), Detail: "AviList v2025b, CC BY 4.0"},
	}

	add := func(label string, as []dataset.Assignment, terms []trait.Term) {
		for i, a := range as {
			row := categoryRow{Value: trait.LabelOf(terms, a.Value), Origin: string(a.Origin)}
			if i == 0 {
				row.Label = label
			}
			switch a.Origin {
			case dataset.Published:
				row.Detail = a.Source + ", as " + a.Value
			default:
				row.Detail = a.Why
			}
			if a.Doc != "" {
				row.Doc, row.DocLabel = "/docs/"+a.Doc, "how BirdView adds one"
			}
			out = append(out, row)
		}
	}
	add("Habitat", sp.Habitats, trait.Habitats)
	add("Diet", sp.Niches, trait.Niches)

	// One row for every region, sharing the one rule that put them all there.
	if len(sp.Regions) > 0 {
		codes := make([]string, 0, len(sp.Regions))
		for _, id := range sp.Regions {
			if rg, ok := region.Get(id); ok {
				codes = append(codes, rg.Code)
				if rg.Code == "" {
					codes[len(codes)-1] = rg.Name
				}
			}
		}
		sort.Strings(codes)
		out = append(out, categoryRow{
			Label:  fmt.Sprintf("Regions (%d)", len(codes)),
			Value:  strings.Join(codes, " "),
			Origin: string(dataset.Derived),
			Detail: "reported there in at least 1 record per million, from eBird and " +
				"iNaturalist records via GBIF",
			Doc: "/docs/data#regions", DocLabel: "how regions are decided",
		})
	}

	// Every way this bird's name or record differs between AviList, which
	// supplies the names, and the occurrence data, which supplies the
	// evidence (product.namesRecordsFiledUnder). Placed before the count rows deliberately: all four
	// change how the numbers below should be read.
	//
	// All of them, not the first that matches. This was an if/else-if over
	// two of the four tables, which would have shown a shared name OR a
	// withheld note had any species ever been both. None is today, and the
	// chain was still the wrong shape: a disclosure silently suppressed by
	// another disclosure is exactly what goals.honestData forbids.
	//
	// Why a reader wants these and not just the number: a visitor who knows
	// the bird as Dryobates villosus and finds a page headed Leuconotopicus
	// has no way to tell a taxonomic update from a mistake, and no way to go
	// back to eBird and check. Naming the alias answers both.
	var seams []categoryRow
	seam := func(value, detail string) {
		seams = append(seams, categoryRow{
			Value: value, Origin: string(dataset.Derived), Detail: detail,
			Doc:      "/docs/data#when-the-sources-disagree-about-a-name",
			DocLabel: "how names are joined",
		})
	}

	// An alias: one taxon, two labels. No count is affected, which is why
	// this says so outright -- a row that merely noted a discrepancy would
	// leave a reader wondering what it cost.
	if old, ok := taxonomy.AviListRenamedFrom[sp.ScientificName]; ok {
		what, how := renamed(old, sp.ScientificName)
		seam(old+" — the "+what+" eBird still uses",
			how+" BirdView publishes AviList's name and joins on eBird's, so one taxon "+
				"carries two labels here: every record on this page is this bird's, and "+
				"the counts are unaffected.")
	}

	// A name whose records are this species' under the taxonomy BirdView
	// serves, because the sources draw the boundary somewhere else.
	//
	// The reason is carried per entry and not written here, because it is
	// not the same reason twice: AviList split the Whimbrel and eBird did
	// not follow, while AviList declined to split the Siberian Stonechat
	// and both databases did. One sentence covering both named the wrong
	// taxonomic event for two of the three, and claimed AviList does not
	// carry Numenius phaeopus, which it does -- as the Eurasian bird.
	for _, c := range taxonomy.CountedInto(sp.ScientificName) {
		seam(c.Published+" — the name eBird files these birds under",
			c.Why+" BirdView follows AviList and counts them as this species, rather "+
				"than choosing between the two databases or publishing the bird twice.")
	}

	// A shared name: the one case where the count is not this bird's alone.
	if other, src, why, ok := taxonomy.SharedWith(sp.ScientificName); ok {
		name := other
		if o := s.byID[s.idOf[other]]; o != nil {
			name = fmt.Sprintf("%s (%s)", o.Name(loc), other)
		}
		seam(src+" — a name covering two species",
			why+" The counts on this page are therefore shared with "+name+
				", and BirdView does not divide them: no source says how they divide.")
	}

	// A form filed below the species. Newly true of the counts rather than
	// merely of the data: until the rank fold these trinomials matched no
	// species at all and were discarded, and now they are counted here.
	if why, ok := taxonomy.DomesticForms[sp.ScientificName]; ok {
		seam("domestic and feral birds, counted here",
			why+", below the species rank. BirdView folds a form into the species it is a "+
				"form of, so the counts on this page include those birds. They are not "+
				"separated, because no source separates them.")
	}

	// Withheld is NOT a name disagreement and must not read as one: telling
	// a reader to look for another name would send them after one that does
	// not exist (product.namesRecordsFiledUnder).
	if taxonomy.WithheldByEBird[sp.ScientificName] {
		seam("withheld by eBird",
			"eBird suppresses "+sp.Name(loc)+" from its public export, as it does for "+
				"species vulnerable to disturbance. The counts here are iNaturalist's alone, "+
				"so they are far lower than the bird's actual record. This is a gap in what "+
				"BirdView can see, not a statement about the bird.")
	}

	// Withheld in particular regions or seasons, under eBird's Sensitive
	// Species list and confirmed against the records (decisions.ebirdSilenceNotAbsence). Every region,
	// not just the ones on screen: this page is the complete record, and the
	// tile -- which names only what is on screen -- must never say more
	// (product.tileOpensDetail). A species eBird withholds everywhere already has its row above.
	if !taxonomy.WithheldByEBird[sp.ScientificName] && len(sp.RegionWithheld) > 0 {
		type where struct{ name, months string }
		var ws []where
		for id := range sp.RegionWithheld {
			name := id
			if rg, ok := region.Get(id); ok {
				name = rg.Name
			}
			ws = append(ws, where{name, maskScope(sp.WithheldIn(id))})
		}
		sort.Slice(ws, func(i, j int) bool { return ws[i].name < ws[j].name })
		value := fmt.Sprintf("withheld by eBird in %d regions", len(ws))
		if len(ws) == 1 {
			value = "withheld by eBird in " + ws[0].name + ", " + ws[0].months
		}
		parts := make([]string, len(ws))
		for i, w := range ws {
			parts[i] = w.name + " (" + w.months + ")"
		}
		seam(value, "eBird keeps this species out of its public records in "+
			strings.Join(parts, ", ")+", to protect it from trapping, disturbance, or targeted "+
			"hunting. In those months the counts here are iNaturalist's alone and far below "+
			"the bird's real record, and the month filter does not judge them: a month with "+
			"no eBird records by policy is not a month without the bird.")
	}

	// One label over the group, continuation rows blank, as the trait rows
	// above do it.
	if len(seams) > 0 {
		seams[0].Label = "Recorded as"
		out = append(out, seams...)
	}

	// The month filter declined to judge this bird, which the tile marks
	// and so this must too (product.tileOpensDetail, decisions.monthPresence).
	if where, n, yes := s.abstains(sp, opt); yes {
		out = append(out, categoryRow{
			Label:  "Months",
			Value:  "shown in every month — too few records to say",
			Origin: string(dataset.Derived),
			Detail: fmt.Sprintf("BirdView holds %s records of this bird in %s, under the %d "+
				"it takes to tell the months a bird is present from the months nobody was "+
				"looking. Hiding it from a month on evidence that thin would state "+
				"something about the bird that is really something about the survey.",
				comma(n), where, month.MinRecords),
			Doc: "/docs/data#how-common-a-bird-is", DocLabel: "what this measures",
		})
	}

	// The figure the tile showed, for the regions the reader had selected
	// when they clicked it (product.tileOpensDetail).
	//
	// Built by shareWithSeams, the same function the tile calls, rather
	// than recomputed here. Two call sites deriving one number from the
	// same inputs is how the index and the detail page came to disagree
	// about where a bird is commonest (decisions.honestSeaDenominator); one function cannot.
	//
	// It is NOT the row below. "Top region" is where the bird peaks, which
	// is a different scope and usually a smaller N, and a reader who saw
	// 1 in 23,393 on the tile and found only 1 in 2,986 here had no way to
	// tell that the page had changed the question.
	if cell := s.shareWithSeams(sp, opt); cell != nil {
		value, detail := cell.Note, cell.Title
		if cell.Ratio != "" {
			value = cell.Ratio + " species observations " + cell.Where
			detail = cell.Title
			for _, sm := range cell.Seams {
				detail += " " + sm.Title
			}
		}
		out = append(out, categoryRow{
			Label:  "Frequency",
			Value:  value,
			Origin: string(dataset.Derived),
			Detail: detail,
			Doc:    "/docs/data#how-common-a-bird-is", DocLabel: "what this measures",
		})
	}

	// Where the bird is the biggest share of what gets recorded (product.shareOfRecords).
	//
	// Stated as "1 in N species observations there" rather than as a percentage: the
	// share runs from 1 in 2 to 1 in 117 million across the dataset, and a
	// percentage is 0.000001% at that end and unreadable well before it.
	// The records it rests on are shown so a thin number discloses itself
	// instead of needing a minimum-evidence cutoff.
	if c := sp.Commonest; c != nil && c.OneIn() > 0 {
		where := c.Region
		if rg, ok := region.Get(c.Region); ok {
			where = rg.Name
		}
		out = append(out, categoryRow{
			// "Top region". It began as the bare "Commonest in", which said
			// nothing about what it ranged over, and spent a while as
			// "Commonest land region" while the seas had no honest
			// denominator (decisions.seaShareDiffers). They have one now (decisions.honestSeaDenominator), so the row
			// ranges over everything again and the qualifier went -- and
			// "top" says plainly that this is the maximum of the same
			// figure the index shows, which selecting this one region there
			// reproduces exactly (acceptance.peakAndShareOneCalculation).
			Label:  "Top region",
			Value:  fmt.Sprintf("%s — 1 in %s species observations there", where, comma(int64(c.OneIn()))),
			Origin: string(dataset.Derived),
			// Both halves of the contrast, because the first on its own only
			// says what the number is NOT. A share is where the bird is.
			Detail: fmt.Sprintf("%s of the %s species observations from %s, from eBird and "+
				"iNaturalist via GBIF: how much of what people record there is this bird, not how "+
				"many there are.",
				comma(c.Records), comma(c.RegionRecords), where),
			Doc: "/docs/data#how-common-a-bird-is", DocLabel: "what this measures",
		})
	}
	return out
}

// shareCell is the tile's share line, pre-rendered: the region's short code,
// the share as one-in-N, and the full sentence as a tooltip.
type shareCell struct {
	// Where names the scope the share is over, already prefixed with "in":
	// the one region when exactly one is on, "every region" when all are,
	// and "the selected regions" for any set in between.
	//
	// Not "the ABA Area", which this is not. That area includes St-Pierre
	// et Miquelon, which BirdView has no region for, and excludes Johnston,
	// Palmyra, Wake, and Navassa, which it has regions for. Naming someone
	// else's boundary would also re-attribute BirdView's scope to the
	// authority decisions.scopeFromRecurrence just removed it from. The selection describes itself.
	Where string
	// Ratio is the whole left-hand phrase, "1 in 883,328" or "0 in
	// 1,225,176,035". Both forms are spelled out rather than the second
	// collapsing to an em-dash: a dash makes a reader wonder whether the
	// number is missing or the bird is, where "0 in 1,225,176,035 bird
	// records across every land region" says which, and says how hard anyone
	// looked before concluding it.
	Ratio string
	Title string
	// Note replaces the ratio when there is nothing to divide by at all --
	// the reader has selected only seas, which carry presence and no share
	// (product.shareOfRecords, decisions.seaShareDiffers). Empty whenever Ratio is set, and never both: a cell
	// either divides or explains.
	Note string
	// Seams are caveats on the figure, one line each under it: a count
	// shared with another species (product.namesRecordsFiledUnder, decisions.sharedCountShownOnBoth), a bird the month filter
	// declined to judge (decisions.monthPresence), a species eBird withholds in the regions
	// on screen (decisions.ebirdSilenceNotAbsence). They accompany the Ratio rather than replacing it:
	// suppressing a count would leave a reader unable to tell it from a bird
	// genuinely absent (goals.honestData).
	//
	// A list, and every caveat that applies is on it. The first version held
	// one, and the abstention marker replaced the shared-name marker when
	// both applied -- a disclosure silently suppressed by another, which is
	// the failure goals.honestData names.
	Seams []seam
}

// seam is one caveat line on a tile: the short text, and the tooltip that
// explains it.
type seam struct{ Text, Title string }

// shareIn gives a species' share of the species observations held by the regions
// currently on screen: its records there, over every bird observation from the
// same regions.
//
// **Computed per request, and no subaggregate is precomputed.** The obvious
// worry is combinatorial -- 69 regions is 2^69 possible selections -- but it
// never arises, because the metric is a ratio of two SUMS over the selection
// and sums compose from their parts. The per-region counts are already the
// finest-grained subaggregate there is; any selection is one pass over them.
// Measured at 2.6 ms for the worst case in a Python sketch, which is 1,081
// species against all 69 regions; in Go it does not register.
//
// A region the reader selected where the species was not reported adds to
// the denominator and nothing to the numerator, which is correct: widening
// the selection to places a bird is absent from should lower its share.
//
// **A region with no entry in totals contributes to NEITHER side** (decisions.seaShareDiffers).
// That is the whole guard, and it replaces a pair of conditions that drifted:
// the numerator ranged over the species' regions and the denominator over the
// totals, so when the marine regions turned out to have counts on one side
// and no honest denominator on the other, every pelagic's share was computed
// as sea records over land records. Deriving both sides from "is there a
// denominator for this region" makes that mismatch unrepresentable rather
// than merely fixed, which matters because the two loops read as a matching
// pair and were not one.
func shareIn(sp *dataset.Species, ds *dataset.Dataset, opt options) (num, den int64) {
	totals := ds.RegionTotals
	all := !opt.HasRegions || len(opt.Regions) == region.Len()
	counted := func(id string) bool {
		if _, ok := totals[id]; !ok {
			return false
		}
		return all || opt.Regions[id]
	}
	// Every month unless the reader said otherwise. Both halves of the
	// ratio are narrowed together, which is the whole point: a June figure
	// divided by June's birding is comparable with a December one, where
	// June's records over the year's birding would not be (product.regionAndMonthFilter).
	allMonths := !opt.HasMonths || len(opt.Months) == month.Count
	inMonth := func(m int) bool { return allMonths || opt.Months[m] }

	// The species' regions and the regions where it stays under the bar
	// both count: the bar decides where it is listed, not what it counts
	// (product.regionAndMonthFilter, decisions.belowBarCounted).
	add := func(ids []string, counts []int64, months [][12]int64) {
		for i, id := range ids {
			if !counted(id) {
				continue
			}
			if allMonths {
				if i < len(counts) {
					num += counts[i]
				}
				continue
			}
			if i < len(months) {
				for m, n := range months[i] {
					if inMonth(m) {
						num += n
					}
				}
			}
		}
	}
	add(sp.Regions, sp.RegionCounts, sp.RegionMonths)
	add(sp.BelowRegions, sp.BelowCounts, sp.BelowMonths)
	for id, n := range totals {
		if !counted(id) {
			continue
		}
		if allMonths {
			den += n
			continue
		}
		for m, v := range ds.RegionMonthlyTotals[id] {
			if inMonth(m) {
				den += v
			}
		}
	}
	return num, den
}

// shareOf builds the tile's "1 in N" line for the regions currently on
// or nil when there is nothing to divide.
//
// It deliberately does NOT reuse the detail page's "commonest in" figure.
// That one answers "where in North America is this bird most of what people
// see", which is trivia a reader looking at a filtered index did not ask:
// filtering to Arizona and being told about Yukon is a non sequitur. This
// answers the question the page is actually posing.
func shareOf(sp *dataset.Species, ds *dataset.Dataset, opt options) *shareCell {
	all := !opt.HasRegions || len(opt.Regions) == region.Len()
	if opt.HasRegions && len(opt.Regions) == 0 {
		return nil // nothing selected; the page shows no birds anyway
	}
	if opt.HasMonths && len(opt.Months) == 0 {
		return nil // likewise: no month is selected, so nothing is in view
	}
	num, den := shareIn(sp, ds, opt)
	// No denominator at all means nothing on screen is a land region --
	// in practice, the reader has selected only seas, which carry presence
	// and no share (decisions.seaShareDiffers). Named for the selection rather than for the
	// sea, because that is the thing the reader can act on.
	// No denominator for anything selected. Unreachable from the UI today,
	// every region having a total -- but it is how the marine regions
	// behaved for months before decisions.honestSeaDenominator measured one, so a region added to
	// internal/region ahead of its data lands here rather than dividing by
	// zero.
	if den == 0 {
		return &shareCell{
			// NOT "no records here". The bird may have plenty: this fires
			// when BirdView lacks the REGION's total, and the numerator is
			// zero only because a region with no denominator is skipped on
			// both sides. Buller's Shearwater has 39,143 records in the
			// North Pacific, and would land here reading "no records" if
			// that sea's total ever went missing again -- which is a
			// statement about the bird, and false.
			Note:  "No data for these regions yet",
			Title: "BirdView has no record totals for these regions yet.",
		}
	}
	where, scope := "in the selected regions", "the selected regions"
	if !all && len(opt.Regions) == 1 {
		for id := range opt.Regions {
			if rg, ok := region.Get(id); ok {
				where, scope = "in "+rg.Name, rg.Name
			}
		}
	} else if all {
		where, scope = "(all regions)", "every region BirdView covers"
	}
	// The months the figure ranges over, where that is not the whole year.
	// Appended to both halves rather than replacing either: a reader needs
	// to know the figure narrowed, and narrowing in two dimensions at once
	// is exactly when a scope label stops being obvious (product.regionAndMonthFilter).
	if ms := monthScope(opt); ms != "" {
		if all {
			where = "(all regions, " + ms + ")"
		} else {
			where += ", " + ms
		}
		scope += " in " + ms
	}
	// Zero is stated, not dashed. "0 in 1,225,176,035 species observations across
	// every land region" carries the effort behind the absence; an em-dash
	// left a reader unable to tell a missing figure from a missing bird.
	ratio, detail := "0 in "+comma(den), "None"
	if num > 0 {
		ratio, detail = "1 in "+comma(den/num), comma(num)
	}
	return &shareCell{
		Where: where,
		Ratio: ratio,
		Title: fmt.Sprintf("%s of the %s species observations from %s are this species.",
			detail, comma(den), scope),
	}
}

// monthScope names the months a figure ranges over, or "" for the whole
// year, where the scope is the default and saying so is noise.
//
// Runs of three or more collapse to "Jun-Aug", and a run is allowed to WRAP
// the year end, so a winter selection reads "Nov-Feb" rather than
// "Jan, Feb, Nov, Dec". Winter is the one selection a birder is most likely
// to make and the one a naive renderer mangles worst.
func monthScope(opt options) string {
	if !opt.HasMonths || len(opt.Months) == 0 || len(opt.Months) == month.Count {
		return ""
	}
	var mask [month.Count]bool
	for m := range opt.Months {
		mask[m] = true
	}
	return maskScope(mask)
}

// maskScope names a set of months the way monthScope does -- runs of three
// collapsed, the year end allowed to wrap -- with two ends of its own: no
// months is "", and all twelve is "all year", because a withheld-by-eBird
// caveat covering the whole year should say so rather than say nothing.
func maskScope(mask [month.Count]bool) string {
	n := 0
	for _, on := range mask {
		if on {
			n++
		}
	}
	switch n {
	case 0:
		return ""
	case month.Count:
		return "all year"
	}
	// Find a start: a set month whose predecessor is not set. With fewer
	// than twelve set one always exists.
	start := -1
	for i := 0; i < month.Count; i++ {
		if mask[i] && !mask[(i+month.Count-1)%month.Count] {
			start = i
			break
		}
	}
	var runs [][]int
	var cur []int
	for k := 0; k < month.Count; k++ {
		i := (start + k) % month.Count
		if mask[i] {
			cur = append(cur, i)
			continue
		}
		if len(cur) > 0 {
			runs = append(runs, cur)
			cur = nil
		}
	}
	if len(cur) > 0 {
		runs = append(runs, cur)
	}
	parts := make([]string, 0, len(runs))
	for _, r := range runs {
		if len(r) >= 3 {
			parts = append(parts, month.Abbrev[r[0]]+"-"+month.Abbrev[r[len(r)-1]])
			continue
		}
		for _, i := range r {
			parts = append(parts, month.Abbrev[i])
		}
	}
	return strings.Join(parts, ", ")
}

// shareWithSeams is shareOf plus the disclosure of product.namesRecordsFiledUnder: where an
// occurrence source files two BirdView species under one name, the tile says
// so rather than presenting the count as this bird's alone.
//
// On the tile and not only on the detail page, because the index is where
// the misreading happens. A Caribbean mangrove resident ranked 51st
// most-recorded in North America looks like a fact until you open it.
func (s *server) shareWithSeams(sp *dataset.Species, opt options) *shareCell {
	cell := shareOf(sp, s.ds, opt)
	if cell == nil || cell.Ratio == "" {
		return cell
	}
	if text, title, yes := s.withheldSeam(sp, opt); yes {
		cell.Seams = append(cell.Seams, seam{text, title})
	}
	if where, n, yes := s.abstains(sp, opt); yes {
		// The reason is on the tile and not only in the tooltip: "shown in
		// every month" alone reads as a fact about the bird, which is the
		// one misreading the abstention exists to prevent (decisions.monthPresence).
		cell.Seams = append(cell.Seams, seam{"Shown in every month due to too few records",
			fmt.Sprintf("BirdView holds %s records of this bird in %s, which "+
				"is too few to tell the months it is here from the months nobody looked. It is "+
				"shown whatever months you select, rather than hidden from some on evidence "+
				"that thin.", comma(n), where)})
	}
	if other, src, _, ok := taxonomy.SharedWith(sp.ScientificName); ok {
		name := other
		if o := s.byID[s.idOf[other]]; o != nil {
			name = o.Name(opt.Locale)
		}
		// Sentence case, like the other caveats: each sits on its own line
		// under the figure and reads as a sentence, not as a fragment of it.
		cell.Seams = append(cell.Seams, seam{"Shared with " + name,
			fmt.Sprintf("This count is not this bird's alone. The occurrence data "+
				"files both it and %s under %s, and does not separate them, so BirdView does not "+
				"divide the records either.", name, src)})
	}
	return cell
}

// sharedRows groups the drawn species whose traits come from one AVONET row:
// the same sheet and the same filed taxon. Such species are drawn alike
// because their measurements are the same, not because the birds are (product.borrowedMeasurementsSaySo).
//
// Derived from the provenance rather than from taxonomy.NearMatches, so a
// borrowing reached any other way still gets its note.
func sharedRows(ds *dataset.Dataset) map[string][]string {
	byRow := map[string][]string{}
	for i := range ds.Species {
		p := ds.Species[i].Segments.Bill.Prov
		if p.SourceTaxon == "" {
			continue
		}
		key := p.Source + "\x00" + p.SourceTaxon
		byRow[key] = append(byRow[key], ds.Species[i].ScientificName)
	}
	out := map[string][]string{}
	for _, names := range byRow {
		if len(names) < 2 {
			continue
		}
		sort.Strings(names)
		for _, n := range names {
			for _, o := range names {
				if o != n {
					out[n] = append(out[n], o)
				}
			}
		}
	}
	return out
}

// sameMeasurements gives one caveat per other drawn species that shares this
// one's AVONET row, for the tile and the detail page (product.borrowedMeasurementsSaySo). The tooltip
// says which way the borrowing runs, because the two species are not in the
// same position: one has no measurements of its own.
func (s *server) sameMeasurements(sp *dataset.Species, loc dataset.Locale) []seam {
	var out []seam
	row := sp.Segments.Bill.Prov
	where := fmt.Sprintf("%s, %s", row.SourceTaxon, strings.TrimPrefix(row.Source, "AVONET v7, "))
	for _, other := range s.sameAs[sp.ScientificName] {
		o := s.byID[s.idOf[other]]
		if o == nil {
			continue
		}
		name := o.Name(loc)
		var why string
		switch {
		case sp.Segments.Bill.Prov.Kind == dataset.NearMatch && o.Segments.Bill.Prov.Kind != dataset.NearMatch:
			why = fmt.Sprintf("AVONET has no measurements of this bird, so it is drawn from %s's (%s).", name, where)
		case o.Segments.Bill.Prov.Kind == dataset.NearMatch && sp.Segments.Bill.Prov.Kind != dataset.NearMatch:
			why = fmt.Sprintf("AVONET has no measurements of %s, so it is drawn from this bird's (%s).", name, where)
		default:
			why = fmt.Sprintf("This bird and %s are both drawn from one AVONET row (%s).", name, where)
		}
		out = append(out, seam{"Same measurements as " + name, why})
	}
	return out
}

// withheldIn returns, for the regions and months on screen, the months eBird
// withholds a species in each region that has any (decisions.ebirdSilenceNotAbsence). Months outside
// the selection are dropped, so a June filter does not report Florida's
// winter.
func (s *server) withheldIn(sp *dataset.Species, opt options) map[string][12]bool {
	allMonths := !opt.HasMonths || len(opt.Months) == month.Count
	out := map[string][12]bool{}
	for id := range sp.RegionWithheld {
		if opt.HasRegions && !opt.Regions[id] {
			continue
		}
		w := sp.WithheldIn(id)
		set := false
		for m := range w {
			if !allMonths && !opt.Months[m] {
				w[m] = false
			}
			set = set || w[m]
		}
		if set {
			out[id] = w
		}
	}
	return out
}

// withheldSeam says, on the tile, that eBird withholds this species in the
// regions and months on screen, so its figure is short by however many
// records eBird kept back (decisions.ebirdSilenceNotAbsence).
//
// One region is named with its months, because that is the case a reader
// can act on: "Withheld by eBird in Florida, Oct-Mar". Several are counted,
// because thirty-eight state names do not fit on a tile, and the detail page
// lists them.
func (s *server) withheldSeam(sp *dataset.Species, opt options) (text, title string, yes bool) {
	w := s.withheldIn(sp, opt)
	if len(w) == 0 {
		return "", "", false
	}
	title = "eBird keeps this species out of its public records here, to protect it -- " +
		"from trapping, disturbance, or targeted hunting -- so the figure counts only what " +
		"iNaturalist holds in those months and is far below the bird's real record. That " +
		"is a gap in what BirdView can see, not a statement about the bird."
	if len(w) == 1 {
		for id, months := range w {
			name := id
			if rg, ok := region.Get(id); ok {
				name = rg.Name
			}
			text = "Withheld by eBird in " + name
			if sc := maskScope(months); sc != "" {
				text += ", " + sc
			}
		}
		return text, title, true
	}
	return fmt.Sprintf("Withheld by eBird in %d regions", len(w)), title, true
}

// abstains reports whether the month filter declined to judge this species,
// and in which region it has the most to go on (decisions.monthPresence).
//
// A species under 200 records in a region cannot have its months told apart
// from its surveys: Millerbird holds 28 in Hawaii, all in five months, and
// lives on Nihoa the year round. It is shown in every selected month, and
// saying so is the whole of the honesty here — the alternative is a page
// that hides a bird from June on the strength of nobody having sailed to
// Nihoa in June.
//
// Reported only when a month filter is actually narrowing something. With
// the whole year selected nothing was judged, so there is nothing to
// disclose and a marker would be noise on every tile.
func (s *server) abstains(sp *dataset.Species, opt options) (where string, most int64, yes bool) {
	if !opt.HasMonths || len(opt.Months) == 0 || len(opt.Months) == month.Count {
		return "", 0, false
	}
	var best string
	var bestN int64 = -1
	for i, id := range sp.Regions {
		if opt.HasRegions && !opt.Regions[id] {
			continue
		}
		if i >= len(sp.RegionMonths) {
			continue
		}
		if _, ok := month.Present(sp.RegionMonths[i], s.ds.RegionMonthlyTotals[id]); ok {
			// Some region in view had enough to judge on, so the filter did
			// its job and the species is here on its merits.
			return "", 0, false
		}
		if n := sp.RegionCounts[i]; n > bestN {
			best, bestN = id, n
		}
	}
	if bestN < 0 {
		return "", 0, false
	}
	if rg, ok := region.Get(best); ok {
		best = rg.Name
	}
	return best, bestN, true
}

// comma groups an integer with thousands separators. The shares reach the
// hundred millions, where an ungrouped run of digits cannot be read at all.
func comma(n int64) string {
	s := strconv.FormatInt(n, 10)
	if len(s) <= 3 {
		return s
	}
	var b strings.Builder
	for i, c := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			b.WriteByte(',')
		}
		b.WriteRune(c)
	}
	return b.String()
}

// served names the two document sets BirdView publishes: the user-facing
// guides and the normative spec. Both are embedded, so a provenance
// citation resolves without the network (product.categoryProvenance, tech.selfContainedServer).
var served = map[string]struct {
	fs   embed.FS
	kind string
}{
	"docs": {docs.Files, "documentation"},
}

// document serves one embedded Markdown file, rendered.
func (s *server) document(w http.ResponseWriter, r *http.Request) {
	set, rest, _ := strings.Cut(strings.TrimPrefix(r.URL.Path, "/"), "/")
	src, ok := served[set]
	if !ok {
		http.NotFound(w, r)
		return
	}
	if rest == "" {
		s.documentIndex(w, set)
		return
	}
	// A diagram, pre-rendered beside the page it belongs to.
	if name, ok := strings.CutSuffix(rest, ".svg"); ok && name != "" && !strings.ContainsAny(name, "/.") {
		svg, err := src.fs.ReadFile(name + ".svg")
		if err != nil {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "image/svg+xml")
		// Opened full size, an SVG is a document of its own; it carries no
		// script, and this says so to the browser as well (tech.selfContainedServer).
		w.Header().Set("Content-Security-Policy", "default-src 'none'; style-src 'unsafe-inline'")
		w.Write(svg)
		return
	}
	// Bare names only, so there is no traversal to reason about.
	if strings.ContainsAny(rest, "/.") {
		http.NotFound(w, r)
		return
	}
	body, err := src.fs.ReadFile(rest + ".md")
	if err != nil {
		http.NotFound(w, r)
		return
	}
	title := markdownTitle(body, rest)
	s.renderMarkdown(w, set, title, withDiagram(src.fs, set, rest, title, body))
}

// mermaidBlock is a page's diagram source. render-diagrams.py reads the same
// pattern, one block per page.
var mermaidBlock = regexp.MustCompile("(?s)```mermaid\n.*?\n```")

// withDiagram replaces a page's Mermaid source with the picture rendered
// from it.
//
// GitHub draws the source itself; BirdView would need Mermaid's script to,
// which tech.selfContainedServer rules out. A page without its rendered SVG keeps the source,
// shown as code, and TestDiagramsAreCurrent fails.
//
// The picture links to itself because it is several times the page's width:
// shrunk to fit, its field names are too small to read.
func withDiagram(fsys embed.FS, set, name, title string, body []byte) []byte {
	if _, err := fsys.ReadFile(name + ".svg"); err != nil {
		return body
	}
	src := "/" + set + "/" + name + ".svg"
	img := fmt.Sprintf("[![Diagram: %s](%s)](%s)\n\n*Select the diagram to open it full size.*",
		title, src, src)
	return mermaidBlock.ReplaceAllLiteral(body, []byte(img))
}

// documentIndex lists a set, so /docs is not a dead end.
func (s *server) documentIndex(w http.ResponseWriter, set string) {
	src := served[set]
	entries, _ := src.fs.ReadDir(".")
	var b bytes.Buffer
	fmt.Fprintf(&b, "# BirdView %s documents\n\n", src.kind)
	for _, e := range entries {
		name, ok := strings.CutSuffix(e.Name(), ".md")
		if !ok {
			continue // a diagram, listed by the page it belongs to
		}
		body, _ := src.fs.ReadFile(e.Name())
		fmt.Fprintf(&b, "- [%s](/%s/%s)\n", markdownTitle(body, name), set, name)
	}
	s.renderMarkdown(w, set, "BirdView "+src.kind+" documents", b.Bytes())
}

// markdownTitle is a document's first top-level heading, falling back to
// its file name.
//
// **The first heading, not the first line.** river/README.md opens with an
// HTML comment recording where the file came from, and reading line one
// listed it in the index as "<!--".
func markdownTitle(src []byte, fallback string) string {
	for _, line := range strings.Split(string(src), "\n") {
		if t, ok := strings.CutPrefix(line, "# "); ok {
			return strings.TrimSpace(t)
		}
	}
	return fallback
}

func (s *server) renderMarkdown(w http.ResponseWriter, set, title string, src []byte) {
	var buf bytes.Buffer
	if err := markdown[set].Convert(src, &buf); err != nil {
		log.Printf("birdview: rendering %q: %v", title, err)
		http.Error(w, "could not render that document", http.StatusInternalServerError)
		return
	}
	s.render(w, "docs.html", struct {
		Title string
		Body  template.HTML
	}{title, template.HTML(buf.String())})
}

// repoLinks rewrites the repository-relative links these documents carry
// into links a browser can follow.
//
// **Rewritten at render time, not in the files.** `../river/requirements/local/product/` is
// the correct link in a repository checkout, and editing the file to suit a
// browser would break it for the checkout; these documents are served
// precisely because they are the files the repository carries.
//
// **Only the guides are served** (decisions.requirementsNotServed). A link to another guide becomes
// its /docs path. A link to anything else in the repository -- the
// requirements under river/, source code, data tables, vendored inputs --
// becomes a link to that file in the published repository, because the
// site does not serve it and a citation to a 404 is worse than none.
type repoLinks struct{ base string }

// repository is where BirdView's source is published.
const repository = "https://github.com/google/birdview"

// repoURL links to path, relative to the repository root, on the code host.
// A path with no extension in its last element is a directory.
func repoURL(path, frag string) string {
	kind := "blob"
	if !strings.Contains(pathpkg.Base(path), ".") {
		kind = "tree"
	}
	return repository + "/" + kind + "/main/" + path + frag
}

func (t repoLinks) Transform(doc *goldmarkast.Document, reader goldmarktext.Reader, _ goldmarkparser.Context) {
	_ = goldmarkast.Walk(doc, func(n goldmarkast.Node, entering bool) (goldmarkast.WalkStatus, error) {
		link, ok := n.(*goldmarkast.Link)
		if !entering || !ok {
			return goldmarkast.WalkContinue, nil
		}
		if dest := t.rewrite(string(link.Destination)); dest != "" {
			link.Destination = []byte(dest)
		}
		return goldmarkast.WalkContinue, nil
	})
}

// rewrite maps one repository-relative destination to where a browser
// should go, or returns "" to leave it alone: anything absolute, external,
// or a fragment within the page.
func (t repoLinks) rewrite(dest string) string {
	u, err := url.Parse(dest)
	if err != nil || u.Scheme != "" || u.Host != "" || u.Path == "" || strings.HasPrefix(u.Path, "/") {
		return ""
	}
	frag := ""
	if u.Fragment != "" {
		frag = "#" + u.Fragment
	}
	// Resolve against the directory the document lives in, as a checkout would.
	path := pathpkg.Clean(t.base + "/" + u.Path)
	if strings.HasPrefix(path, "../") || path == ".." {
		return "" // outside the repository; nothing sensible to point at
	}
	if name, ok := strings.CutPrefix(path, "docs/"); ok && !strings.Contains(name, "/") {
		if stem, isDoc := strings.CutSuffix(name, ".md"); isDoc {
			return "/docs/" + stem + frag
		}
	}
	if path == "docs" {
		return "/docs/" + frag
	}
	return repoURL(path, frag)
}

// markdown renders the served documents. GFM tables are enabled because the
// documents use them heavily -- the source tables and the work-arising list
// are tables -- and without the extension they render as runs of pipes.
//
// Raw HTML in the source is NOT passed through: these files are ours, but a
// renderer that emits whatever HTML it is given is a habit worth not having.
var markdown = map[string]goldmark.Markdown{
	"docs": newMarkdown("docs"),
}

func newMarkdown(base string) goldmark.Markdown {
	return goldmark.New(
		goldmark.WithExtensions(extension.Table, extension.Strikethrough, extension.Linkify),
		goldmark.WithParserOptions(
			// Heading ids, so a provenance citation can point at the
			// section that answers it rather than at the whole document
			// and a hope (product.categoryProvenance).
			goldmarkparser.WithAutoHeadingID(),
			goldmarkparser.WithASTTransformers(
				goldmarkutil.Prioritized(repoLinks{base}, 100))),
		// No hard wraps: the sources are wrapped at 100 columns, and turning
		// each newline into <br> broke every paragraph mid-sentence, unlike
		// the same file on a code host.
	)
}

// added lists every category BirdView assigned itself, in one place.
//
// The detail pages already say it bird by bird, which answers "why is this
// robin a frugivore" and not "what has this project decided on its own
// account". Those are different questions, and the second is the one
// somebody auditing the project would ask. Published assignments are
// excluded: there are 1,678 of them and they are AVONET's, not news.
func (s *server) added(w http.ResponseWriter, r *http.Request) {
	type row struct{ kind, species, value, origin, why string }
	var rows []row
	for i := range s.all.Species {
		sp := &s.all.Species[i]
		for _, a := range sp.Habitats {
			if a.Origin != dataset.Published {
				rows = append(rows, row{"Habitat", sp.Name(dataset.LocaleNA),
					trait.LabelOf(trait.Habitats, a.Value), string(a.Origin), a.Why})
			}
		}
		for _, a := range sp.Niches {
			if a.Origin != dataset.Published {
				rows = append(rows, row{"Diet", sp.Name(dataset.LocaleNA),
					trait.LabelOf(trait.Niches, a.Value), string(a.Origin), a.Why})
			}
		}
	}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].kind != rows[j].kind {
			return rows[i].kind > rows[j].kind // Habitat before Diet
		}
		if rows[i].origin != rows[j].origin {
			return rows[i].origin < rows[j].origin
		}
		return rows[i].species < rows[j].species
	})

	var b bytes.Buffer
	b.WriteString("# What BirdView adds to its sources\n\n")
	fmt.Fprintf(&b, "Every category on this site comes from a published source except these %d, "+
		"which are BirdView's own. `derived` means computed from published data by a stated rule; "+
		"`birdview` means BirdView decided it.\n\n", len(rows))
	b.WriteString("| | Species | Category | Origin | Why |\n|---|---|---|---|---|\n")
	for _, r := range rows {
		fmt.Fprintf(&b, "| %s | %s | %s | `%s` | %s |\n",
			r.kind, r.species, r.value, r.origin, r.why)
	}
	s.renderMarkdown(w, "docs", "What BirdView adds to its sources", b.Bytes())
}

// excluded lists every bird left off the index, with its reason and its
// records, each linked to its detail page (product.escapedBirdsListed). A bird a reader has seen
// should not vanish without a word; this is the word.
func (s *server) excluded(w http.ResponseWriter, r *http.Request) {
	var rows []*dataset.Species
	for i := range s.all.Species {
		if sp := &s.all.Species[i]; sp.Exotic != "" {
			rows = append(rows, sp)
		}
	}
	sort.Slice(rows, func(i, j int) bool {
		return rows[i].Name(dataset.LocaleNA) < rows[j].Name(dataset.LocaleNA)
	})
	var b bytes.Buffer
	b.WriteString("# Birds left off the index\n\n")
	fmt.Fprintf(&b, "These %d birds pass BirdView's scope rule, but their records here are of "+
		"escaped, released, or kept birds, with no population of their own, so they are left off "+
		"the index. Each keeps its own page. The list is BirdView's judgment; long-feral "+
		"populations, such as Indian Peafowl in Florida, are not on it. See "+
		"[how the data is made](/docs/data#scope).\n\n", len(rows))
	b.WriteString("| Bird | Scientific name | Records | Why it is left off |\n|---|---|---:|---|\n")
	for _, sp := range rows {
		var n int64
		for _, c := range sp.RegionCounts {
			n += c
		}
		for _, c := range sp.BelowCounts {
			n += c
		}
		fmt.Fprintf(&b, "| [%s](/species/%s) | *%s* | %s | %s |\n",
			sp.Name(dataset.LocaleNA), sp.ID, sp.ScientificName, comma(n), sp.Exotic)
	}
	s.renderMarkdown(w, "docs", "Birds left off the index", b.Bytes())
}
