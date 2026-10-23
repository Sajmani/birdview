# Architecture

How the code is actually put together. Descriptive: where this and the code disagree, this file
is wrong.

## Shape

Two commands over a shared set of internal packages, plus six generators that write the committed
tables under `data/`, and a checker for deployments. Generators run by hand, never by the build.

```
cmd/birdview-ingest      compiles sources and data/ tables into web/data/species.json (a build artifact)
cmd/birdview             serves it on localhost, rendering HTML server-side
cmd/birdview-status      generator: data/status.tsv, BirdView's own occurrence status (the scope)
cmd/birdview-regions     generator: data/region-presence.tsv, per-region and per-month presence
cmd/birdview-inatmap     generator: data/inat-taxa.tsv, one iNaturalist taxon ID per species
cmd/birdview-gbifmap     generator: data/gbif-taxa.tsv, the GBIF -> AviList name bridge
cmd/birdview-boundaries  generator: data/region-boundaries.json, from Natural Earth
cmd/birdview-names       generator: data/source-names.tsv; only a test reads it
cmd/birdview-smoke       checks a running deployment from outside, after each deploy
internal/avilist         names, families, orders, sequence and eBird codes from AviList
internal/avonet          traits from all three AVONET taxonomy sheets, with habitat and niche
internal/wingspan        empirical wingspans, and the regression fitted from them
internal/taxonomy        the name mapping: renames, near-matches, synonyms in both directions
internal/status          recurrence: episodes, the status they imply, and the committed table
internal/region          the regions, their two groupings, and their IDs
internal/month           the twelve months and the rule for presence in one
internal/boundary        the region boundary file and the reference lookup over it
internal/trait           the habitat and diet vocabularies, and BirdView's additions
internal/gbif            the GBIF -> AviList name bridge, and the paced API client
internal/inat            the iNaturalist taxon lookup, and the links built from it
internal/morphology      the segment models, and the rounding that keeps them summing
internal/palette         generated order and family colors under the contrast rule
internal/dataset         the record and the provenance vocabulary
internal/view            ordering, sectioning, scale and diagram geometry; pure, no HTML
internal/xlsx            the standard-library workbook reader behind the three source readers
internal/smoke           the post-deploy checks; a package so the server's tests can run them
web/static/              embedded templates, CSS, here.js and the favicon
data/                    the committed generated tables; region-boundaries.json is embedded
docs/                    the guides, embedded and served at /docs
.github/workflows/       CI: the standing checks on every push and pull request
Dockerfile               ingest, test, build; the test stage gates every image
```

## Why rendering is server-side

The ordering, sectioning and geometry rules carry acceptance criteria, and property tests over
pure Go functions are the cheapest checker that would actually catch a violation. Putting those
rules in JavaScript would have meant either duplicating them in Go for the tests, or testing them
through a browser harness the project does not have.

So the server computes sections and geometry and renders HTML. Every control is a query parameter
and every detail view is a URL, which also makes any view shareable. The cost is a page load per
interaction, which is unnoticeable on localhost.

**Two exceptions, both deliberate.** A page load per interaction means the browser lands at the
top each time, discarding your place in a list of over a thousand birds. An inline script in
`index.html` remembers, for the session, the scroll position and which panels are collapsed, and
restores them on return. It holds no application logic.

The other is `here.js`, behind the Here button. It looks up the reader's region in the browser so
their position never reaches the server ([product.locationStaysInBrowser](requirements/local/product/filtering.md#locationstaysinbrowser)), and a test holds
it to the Go lookup in `internal/boundary`. The rule both protect is that *logic* stays in Go, not
that the client is inert; without JavaScript the page still works, minus these two.

This is what let [acceptance.familyFilterSelectsAndClears](requirements/acceptance.md#familyfilterselectsandclears) and [acceptance.unitToggleSwitches](requirements/acceptance.md#unittoggleswitches) be real checks rather than declared gaps.

## Data flow

The ingest, in outline; [docs/pipeline.md](../docs/pipeline.md) draws every source, field and step.

```
data/status.tsv ─────────► scope: which species are drawn, and which are not and why
AviList ─────────────────► names, family, order, sequence, eBird code
AVONET ──────────────────► resolve each species to a trait row ──► segments ──┐
                             (own row, then renamed row, then a near-match)   │
Fox wingspans ───────────► empirical value, else the fitted order-level ──────┤
                             regression                                       ├─► species.json ─► server
data/inat-taxa.tsv ──────► iNaturalist link ──────────────────────────────────┤
data/region-presence.tsv ► regions, months, share of records ─────────────────┘
```

The generators that write those `data/` tables query GBIF (eBird and iNaturalist records) and
iNaturalist over the network; the ingest itself reads only files in the repository.

Resolution order (`resolve` in `cmd/birdview-ingest`) is the substance of [decisions.abaSpeciesWithoutAvonet](requirements/decisions.md#abaspecieswithoutavonet): prefer any row that is genuinely *this* taxon —
whatever name or sheet it is filed under — over borrowing a relative's measurements. In order: a
sheet pinned per species (`taxonomy.OwnRowSheets`, [decisions.herringGullEuropeanRow](requirements/decisions.md#herringgulleuropeanrow)); the AviList
name in `AVONET2_eBird`; then that name, the name AviList renamed it from, and AVONET's older
labels, each across the eBird, BirdLife and BirdTree sheets. All of these are `measured`. Only when
no such row exists does a near-match to the pre-split parent apply (`taxonomy.NearMatches`, with
`NearMatchSheets` pinning a narrower sheet, [decisions.whiteTernBorrows](requirements/decisions.md#whiteternborrows)), and the label then says so.

## The invariant that shapes the code

Segments must sum **exactly** to total length, because the drawn height of a diagram is what makes
two birds comparable. `morphology.BalanceToSum` rounds the total and four segments, then derives
the trunk as the remainder, so the rounded parts always reconcile. Rounding each part
independently leaves up to 0.2 mm adrift; the previous design did that and broke the invariant for
27.6% of its species.

## Where the scale decisions live

`view.Sections` groups by a metric's step rule, and `view.SectionsBy` by any key and break rule
(frequency uses it, breaking by decade). `view.Scales` picks one reference length per group — the
section's largest bird, or the whole set's in global mode. `view.Layout` turns a reference into pixels. Keeping the
reference separate from the pixel conversion is what lets [acceptance.sharedScale](requirements/acceptance.md#sharedscale) assert that every bird in a
section shares a scale factor, which is the property [product.scaleToggle](requirements/local/product/diagram.md#scaletoggle) actually requires.

The server also computes a *width* reference per group, because a to-scale crossbar on a
wide-winged bird in a section of short ones would otherwise overflow. Both references come from
the same group, so one scale still serves the whole section.

## The network generators, and why they are not in the ingest

`birdview-inatmap`, `birdview-gbifmap`, `birdview-regions`, `birdview-status` and
`birdview-names` all ask somebody else's server about BirdView's species, and all write a committed
table rather than running at build time
([tech.pacedInaturalistMapping](requirements/local/tech/process.md#pacedinaturalistmapping)). The reasons are the same in each case: tests make no network request, a build that needs
the network fails on a plane, and a quarter of an hour of someone else's server time does not
belong in a `go build`. The tables go stale; a stale row fails visibly and the refresh is one
command. `birdview-boundaries` makes one hash-checked download of Natural Earth.

The two mapping tools differ in what a miss means. iNaturalist supplies a *link*, so a miss costs a link and
nothing else, and `inat.Table` simply omits the species. GBIF supplies a *name correspondence*,
where the set of things that did not resolve is the interesting output — so `gbif.Table` carries
every species GBIF reported, with a tier saying how it resolved or why it did not.

`gbif.Resolve` tries three tiers and refuses ambiguity at each: the accepted name, then the
synonyms GBIF files under it, then the epithet within GBIF's family after folding gender agreement
and Latin case. The last is the only fuzzy one, it contributes three rows out of 1,378, and it is
scoped and uniqueness-checked because AviList holds 369 (family, folded epithet) buckets with more
than one species in them.

**Nothing joins against `data/gbif-taxa.tsv` yet.** The bridge exists so that a later decision
about where species scope comes from can be made on measured ground; making that decision is not
this code's job.

## Known rough edges

Listed in [todo.md](todo.md) with the rest of the open work, since 2026-10-13.

## Gotchas

- **The neck/body ratio has no measured basis, and is surfaced anyway** ([product.neckModelPerFamily](requirements/local/product/diagram.md#neckmodelperfamily)).
  The neck length is serialized and shown with `kind: modeled` and a `model` naming the function
  and that family's ratio. Do not drop the model string: naming it is what makes showing the number
  defensible ([decisions.neckLengthAsMetric](requirements/decisions.md#necklengthasmetric), which
  superseded [decisions.neckSplitUnmeasured](requirements/decisions.md#necksplitunmeasured)).
- **Per-tile diagram scaling destroys the product** ([product.scaleToggle](requirements/local/product/diagram.md#scaletoggle)). It makes a hummingbird
  and a swan identical on screen. Scale is per-section or global, never per-tile.
- **The wing crossbar is to scale with the vertical axis** ([product.crossbarAtVerticalScale](requirements/local/product/diagram.md#crossbaratverticalscale)), which
  is why tiles are wide.
- **AVONET has no common names and no body length or wingspan.** Three designs foundered on
  assuming otherwise. Names come from AviList ([product.namesFromAvilist](requirements/local/product/dataset.md#namesfromavilist)); length and wingspan from
  [product.lengthAndWingspanSources](requirements/local/product/measurements.md#lengthandwingspansources).
- **Scope is recurrence, not abundance** ([product.scopeIsRecurringAvifauna](requirements/local/product/dataset.md#scopeisrecurringavifauna),
  [decisions.scopeFromRecurrence](requirements/decisions.md#scopefromrecurrence)). A record count measures observer effort, which under-reports
  pelagics and island endemics and over-reports vagrants.
- **Identify the client on outbound requests** ([tech.identifiedClient](requirements/local/tech/delivery.md#identifiedclient)). Go's default user agent
  earns a 403 from Wikimedia.

## Running the local server

`-data` defaults to the relative `web/data/species.json`, so the server must start from the repo
root. Build to a fixed path and stop that process before restarting; `kill %1` on `go run` leaves
its child holding the port. Start it detached, with its output in a file, so the shell that
launched it returns at once:

```
pkill -f /tmp/birdview-serve; go build -o /tmp/birdview-serve ./cmd/birdview && \
  nohup /tmp/birdview-serve > /tmp/birdview-serve.log 2>&1 &
```

A `-dev web` server reads templates and static files from `web/static/` per request but its Go
code is frozen at start, so `can't evaluate field X` after a struct change usually means a stale
process.
