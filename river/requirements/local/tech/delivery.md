package tech

# Structure and delivery

## twoCommandsOneModule

**Two commands, one module**

Subject: `build.layout`
Applies to: repo
Goals: [goals.rapidIteration](../../../goals/goals.md#rapiditeration), [goals.goodDataCitizen](../../../goals/goals.md#gooddatacitizen)

A single Go module with an ingest command that produces the dataset, and a server command that
embeds and serves it.

## selfContainedServer

**The server is self-contained at runtime**

Subject: `runtime.network`
Value: `none`
Applies to: app
Goals: [goals.openToEveryone](../../../goals/goals.md#opentoeveryone), [goals.readerPrivacy](../../../goals/goals.md#readerprivacy), goals.secure

The server binary makes no network request to serve the application. No CDN for styles or scripts,
no remote data fetch.

Rationale: the tool should work with no network connection.

## servesOnlyItsPages

**The server serves its pages and what they load, and nothing else**

Subject: `server.routes`
Applies to: app
Goals: [goals.freedomToChange](../../../goals/goals.md#freedomtochange), goals.secure

Under `/static/` the server serves the stylesheet and the favicon. It never serves a page template
or a directory listing, though the templates are embedded in the same directory. The one route
that is neither a page nor something a page loads is `/robots.txt`
([crawlersSkipFilteredViews](#crawlersskipfilteredviews)).

Rationale: a plain file server over that directory served every page's Go template source as an
HTML page of its own, and the home page's template source at the bare `/static/`.

## crawlersSkipFilteredViews

**Crawlers are asked to skip filtered views**

Subject: `server.robots`
Applies to: app
Goals: [goals.openToEveryone](../../../goals/goals.md#opentoeveryone)

The server answers `/robots.txt` with a policy that lets every crawler fetch each page that has no
query string, and asks it to fetch nothing with a query string, nor the region boundary file.

Rationale: every control is a query parameter, so the home page links to hundreds of variants of
itself, each linking to hundreds more: an unbounded set of near-duplicate pages, each costing the
egress a public deployment pays for. What is worth indexing has no query string: the home page,
about a thousand species pages, the guides, and the added and excluded lists. A `robots.txt` is a request that well-behaved crawlers
honor, not a limit; what bounds the cost of one that ignores it is the deployment's instance cap.

## stableUrls

**These URLs keep working; everything else may change**

Subject: `server.public_interface`
Applies to: app
Goals: [goals.shareableViews](../../../goals/goals.md#shareableviews), [goals.freedomToChange](../../../goals/goals.md#freedomtochange)

BirdView's public interface is its URLs: the home page `/`, a species page `/species/<eBird code>`
(any of a species' codes, the others redirecting to its own; [product.ebirdLink](../product/presentation.md#ebirdlink)),
the guides `/docs` and `/docs/<guide>`, the lists `/added` and `/excluded`, and the query
parameters the pages' own links write (`by`, `rev`, `step`, `units`, `names`, `family`, `region`,
`month`, `habitat`, `niche`, `freq`, `pin`, `show`), with the values those links write. A URL of that form keeps working across releases. Everything
else is internal and may change without notice: page markup, wording and styling, anything under
`/static/` or `/boundaries/`, and the shape of `species.json`.

**Keeps working** means the URL still shows what it showed, or the nearest thing the data now
allows, and says what changed, as a species that leaves scope already does
([decisions.speciesUrlOutlivesSpecies](../../decisions.md#speciesurloutlivesspecies),
[decisions.compareUrlContract](../../decisions.md#compareurlcontract)). Renaming a path,
a parameter, or a value it accepts needs a decision, and the old form keeps answering, read as
before or redirected. A region value that BirdView replaced, or that its source retired, redirects
to the codes that replace it **without a message**, the page's heading naming the regions shown
([decisions.regionCodesFromData](../../decisions.md#regioncodesfromdata)). `common`, the reporting frequency's name until 2026-11-08, redirects to
`freq` ([decisions.reportingFrequencyParam](../../decisions.md#reportingfrequencyparam)). `habitat=Marine` redirects to its two refinements,
`Marine-inshore` and `Marine-offshore` ([decisions.openOceanSplit](../../decisions.md#openoceansplit)). One decision removed URLs outright: the comparison `/compare`, the
current-month link `/now`, and the `scale` and `species` parameters went with the classic view
([decisions.classicViewRemoved](../../decisions.md#classicviewremoved)). One changed what a URL
shows: a home page address with no month showed the current month and now shows every month;
Now chooses the current month by name
([decisions.homeShowsEveryMonth](../../decisions.md#homeshowseverymonth)).

Rationale: every view is a URL ([goals.shareableViews](../../../goals/goals.md#shareableviews)), and
once BirdView is public, people bookmark and send them; a link that breaks is a promise broken.
Naming exactly what is promised is what leaves the rest free to change
([goals.freedomToChange](../../../goals/goals.md#freedomtochange)). `species.json` is a serving
artifact; a dataset others may build on is a separate, documented publication
([publishDataset](../../../todo.md#publishdataset)).

## oneCanonicalHost

**A deployment answers at one address**

Subject: `server.canonical_host`
Value: `-canonical <host>`; unset by default
Applies to: app
Goals: [goals.shareableViews](../../../goals/goals.md#shareableviews)

Given a canonical host, the server answers a request addressed to any other host with a permanent
redirect to the same path and query on the canonical host, over HTTPS. Given none, it serves
whatever host it is asked for.

Rationale: birdview.site answers at `birdview.site`, `www.birdview.site` and its Cloud Run address,
so one view has three URLs, bookmarks split between them, and search engines index each as a
duplicate. Redirecting makes the canonical one the address every share ends up at. The host
belongs to the deployment, not the code, so it is a flag: running locally, or deploying a copy
elsewhere, needs no change.

## regionBoundaries

**Region boundaries are generated from Natural Earth and kept small**

Subject: `data.region_boundaries`
Applies to: app
Goals: [goals.readerPrivacy](../../../goals/goals.md#readerprivacy), [goals.openToEveryone](../../../goals/goals.md#opentoeveryone)

The boundaries Here tests against ([product.hereSelectsLocalRegion](../product/filtering.md#hereselectslocalregion)) are generated by
`cmd/birdview-boundaries` from Natural Earth's 1:10m states and provinces, pinned by hash, into a
committed file holding every land region BirdView covers and its sea boxes. It is at most
**200 KB as sent compressed**, and is served at a URL carrying its content hash.

The sea boxes are written from the same Go table as the region filter's, so the two cannot drift.
The source is fetched and checked against its hash when the file is regenerated, not committed: it
is 14 MB worldwide, and BirdView needs the United States, Canada, and five atolls. See
[imports.md](../../imports.md#natural-earth).

## seaDistanceTable

**Where each species' records lie relative to shore is derived once and committed**

Subject: `data.sea_distance`
Applies to: dataset
Goals: [goals.rapidIteration](../../../goals/goals.md#rapiditeration), [goals.giveBack](../../../goals/goals.md#giveback), [goals.goodDataCitizen](../../../goals/goals.md#gooddatacitizen)

`cmd/birdview-sea` writes `data/sea-distance.tsv` from the offshore-cells GBIF download (records with
no GADM code, in 0.1° cells by species and month; [gbifDownloadsRecorded](process.md#gbifdownloadsrecorded)) and Natural Earth's 1:10m
land and minor islands, pinned by hash and fetched, never committed
([imports.md](../../imports.md#natural-earth)). It counts eBird records only, inside BirdView's sea boxes; drops each
cell whose center lies on land, which GADM leaves uncoded in places; measures each remaining cell's
center to the nearest shore; and writes, per drawn species, its records under 10 km, 10 to 30 km,
and 30 km or more, with the download's key and DOI in the table's header. The table is committed,
so the ingest needs neither the download nor the network.

Rationale: the download is 90 MB and the polygons 3.5 MB; the derived counts are a few kilobytes
that anyone can check against the method. The method was tried as research first
([notes: Distance from shore](../../../notes.md#distance-from-shore-as-evidence-for-the-open-sea-question)).

## servesOnLocalhost

**Listens on localhost:8080 unless told otherwise**

Subject: `server.address`
Value: `localhost:8080` by default; a deployment sets `-addr`
Applies to: app
Goals: [goals.rapidIteration](../../../goals/goals.md#rapiditeration), goals.secure

The server listens on localhost:8080 unless `-addr` names another address. A deployment names one;
a local run never needs to.

Rationale: birdview.site runs with `-addr 0.0.0.0:8080`, because Cloud Run must reach it, while a
local run keeps the default and is reachable from its own machine only
([decisions.publicPersonalDeployment](../../decisions.md#publicpersonaldeployment)).

## fetchRestrictedInputs

**Build inputs that are not vendored are fetched into `fetched/`, hash-checked, and never committed**

Subject: `build.inputs`
Applies to: repo
Goals: [goals.goodDataCitizen](../../../goals/goals.md#gooddatacitizen)

Any build input that BirdView does not vendor is downloaded onto the building machine into
`fetched/`, which `.gitignore` excludes, verified against its recorded hash before use, and never
committed to version control. Vendored inputs (AviList, the AVONET workbook, the Fox et al.
wingspan data, and the USGS banding list under `third_party/`) and BirdView's own derived tables
under `data/` ([product.joinOnBinomial](../product/dataset.md#joinonbinomial)) are committed.

Rationale: keeping `fetched/` gitignored and hash-checked separates inputs read only on a building
machine from files published with the repository. No ingest input has used `fetched/` since
[decisions.scopeFromRecurrence](../../decisions.md#scopefromrecurrence) replaced the ABA checklist;
the directory and `.gitignore` entry stay for build-time fetches. Whether a publisher permits
redistribution or automated retrieval comes from that source's own terms ([imports.md](../../imports.md)).

## datasetNeverCommitted

**The built dataset is an artifact, never committed**

Subject: `build.output`
Applies to: repo
Goals: [goals.giveBack](../../../goals/goals.md#giveback)

The generated dataset is excluded from version control; a clean checkout builds it.

It is reproducible from the committed inputs in one command, and a stale copy in the tree is worse
than none. (It was first excluded for carrying ABA-derived values, which
[decisions.scopeFromRecurrence](../../decisions.md#scopefromrecurrence) removed.)

## containerImageBuilds

**The container image builds from a clean checkout**

Subject: `build.image`
Applies to: repo
Goals: [goals.secure](../../../goals/goals.md#secure), [goals.openToEveryone](../../../goals/goals.md#opentoeveryone)

The `Dockerfile` at the repository root builds BirdView's image from a clean checkout and nothing
else: every stage copies in all the packages and files its command needs, and the ingest runs
inside the build, so the image carries a dataset no one had to supply.

Rationale: the image is how BirdView deploys ([docs/building.md](../../../../docs/building.md)),
and a fix that cannot ship is not quick to ship. Nothing built it, so when `cmd/birdview` began
importing `data/` on 2026-10-18 the server stage stopped building and nobody noticed for four days.

## identifiedClient

**Identify the client on every outbound request**

Subject: `http.user_agent`
Applies to: dataset
Goals: [goals.goodDataCitizen](../../../goals/goals.md#gooddatacitizen)

Any HTTP client this project uses to fetch a source sends a descriptive `User-Agent` naming the
application and a contact. Go's default `Go-http-client/1.1` is not acceptable.

Rationale: GBIF's API asks for it, Wikimedia's User-Agent policy returns HTTP 403 for generic
agents, and iNaturalist's API guidance asked the same while BirdView called it. Wikimedia's matters
if [product.speciesPhotographs](../product/diagram.md#speciesphotographs) returns.

## goVersionFollowsDeps

**The Go version tracks whatever dependencies require**

Subject: `build.go_version`
Value: `no floor; follow dependencies`
Applies to: repo
Goals: [goals.rapidIteration](../../../goals/goals.md#rapiditeration), goals.secure

BirdView sets no minimum Go version of its own. When a dependency requires a newer toolchain,
`go mod tidy` raises the `go` directive and the toolchain switches, and that is an accepted
routine change rather than one needing a decision.

Rationale: a minimum-version floor protects *dependents* from a bump they did not ask for.
BirdView exports no library and has no dependents, so a floor would protect nobody and block
upstream fixes. Module versions stay pinned in `go.sum`; only the language and toolchain versions
float.

## compressedResponses

**Responses are compressed when the client asks**

Subject: `http.content_encoding`
Value: `gzip on Accept-Encoding, Vary on every response`
Applies to: app
Goals: [goals.openToEveryone](../../../goals/goals.md#opentoeveryone)

Every response is gzipped when the request carries `Accept-Encoding: gzip`, and every response
carries `Vary: Accept-Encoding` whether or not it was compressed.

Rationale: the index inlines an SVG diagram for each of 1,083 species, **3.3 MB of markup that
compresses to 270 KB**. Egress is what a serverless deployment bills for: against 1 GiB of free
monthly transfer, that is about **3,980 index loads instead of about 320**.

**Done in the application, not left to a CDN or a load balancer**, so the saving does not depend
on the deployment shape, which is not chosen yet.

**`Vary` goes on uncompressed responses too.** Without it a cache that stored the plain reply
could serve it to clients asking for gzip, or hand a gzip stream to a client that cannot read one.

**The headers are decided at `WriteHeader`, not at the first `Write`.** A handler that sets a
status commits the header map before writing a byte, so a compressor that labels from `Write`
sends a compressed body with no `Content-Encoding`, and the browser renders binary. Static files
and every [product.undrawnSpeciesAnswer](../product/presentation.md#undrawnspeciesanswer) 404
first shipped that way.

**A status that may carry no body — 204, 304 — is not compressed at all**: even an empty gzip
stream has a twenty-byte header, which the protocol says must not be there.

## cacheLifetimeStated

**Every response says how long it may be reused**

Subject: `serving.cache_validation`
Value: `Cache-Control: no-cache and an ETag on every response, except a content-hashed file`
Applies to: app
Goals: [goals.honestData](../../../goals/goals.md#honestdata), [goals.openToEveryone](../../../goals/goals.md#opentoeveryone)

Every response carries a `Cache-Control` directive and a validator, so a cache knows whether
it may reuse what it holds and can ask cheaply when it may not.

**A file whose URL carries a hash of its own content is the one exception**: it is served
`Cache-Control: public, max-age=31536000, immutable`. Its URL changes whenever its content does,
so a cached copy can never be stale. The region boundary file (regionBoundaries) is served this
way; the pages that load it are not
([decisions.hereAndNowCaching](../../decisions.md#hereandnowcaching)).

Rationale: **sending no cache headers is not the same as sending "do not cache".** With no
freshness information and no validator, a browser applies its own heuristic and has nothing to
revalidate against. A reader who reloads after a rebuild is served the old page and cannot tell
it from a bug.

**`no-cache` means store it and ask before reusing it**, not "do not store". It suits every page
here: validation is cheap and staleness is never acceptable, since a figure like
[product.shareOfRecords](../product/presentation.md#shareofrecords)'s must be current.

**The ETag is a fingerprint of the build and the URL, not a hash of the body.** Hashing the body
means buffering it, and the index is 3.3 MB before compression, memory
[compressedResponses](#compressedresponses) exists to avoid spending. This is sound only because a
page is a pure function of the dataset, the templates, the static files, and the request URI, with
no clock in the render path; [app.honestRevalidation](../../composed/internal/app/acceptance.md#honestrevalidation)
asserts that.

**One fingerprint over all assets, not one per asset**, so a stylesheet edit invalidates the HTML
that loads it. Over-invalidating costs one refetch after a deploy; under-invalidating serves a page
with a stylesheet it was not written for.

**The compressed representation carries a different validator**, suffixed rather than dropped. A
gzipped body is a different representation, and a shared tag would let a cache hand gzip bytes to
a client that asked for none.

**`-dev` sends `no-store` and no validator at all.** It re-reads templates per request, so no
fingerprint taken at startup stays true.
