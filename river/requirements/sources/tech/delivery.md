package tech

# Structure and delivery

## twoCommandsOneModule

**Two commands, one module**

Subject: `build.layout`
Goals: goals.sizeIntuition

A single Go module with an ingest command that produces the dataset, and a server command that
embeds and serves it.

## selfContainedServer

**The server is self-contained at runtime**

Subject: `runtime.network`
Value: `none`
Goals: goals.sizeIntuition

The server binary makes no network request to serve the application. No CDN for styles or scripts,
no remote data fetch.

Rationale: the tool should work on a laptop in a cabin. This supersedes the earlier design's
`cdn.tailwindcss.com` dependency, which would have left the UI unstyled offline.

## servesOnLocalhost

**Serves on localhost:8080**

Subject: `server.address`
Value: `localhost:8080`
Goals: goals.sizeIntuition

## fetchRestrictedInputs

**Fetch restricted inputs at build time; commit only what may be redistributed**

Subject: `build.inputs`
Goals: goals.goodDataCitizen

Inputs whose license forbids redistribution are downloaded onto the building machine at build time,
verified against the hashes recorded in [sources.md](../../requirements.md), consumed, and never committed.
They land under `fetched/`, which `.gitignore` excludes. Since [decisions.scopeFromRecurrence](../../decisions.md#scopefromrecurrence)
replaced the ABA checklist, no build input uses this path; the rule and the `.gitignore` entry stay
in place in case a future source requires them.

This mechanism is available only where the publisher permits automated retrieval. It is not a
general workaround for a restrictive license: Ricklefs 2017 was rejected partly because Wiley's
terms forbid automated access outright ([decisions.ricklefsProhibited](../../decisions.md#ricklefsprohibited)).

The CC BY inputs are committed: AviList, the AVONET workbook, and the Fox et al. wingspan data. So
are the synonym tables of product.joinOnBinomial and the occurrence tables under `data/`, which are the project's
own work.

## datasetNeverCommitted

**The built dataset is an artifact, never committed**

Subject: `build.output`
Goals: goals.goodDataCitizen

The generated dataset is excluded from version control. It was originally excluded because it
carried values derived from the ABA checklist ([decisions.scopeFromRecurrence](../../decisions.md#scopefromrecurrence) ended that); it stays
excluded because it is reproducible from the committed inputs in one command and a stale copy in the
tree is worse than none. A clean checkout builds it; it is not shipped in the repository.

## identifiedClient

**Identify the client on every outbound request**

Subject: `http.user_agent`
Goals: goals.goodDataCitizen

Any HTTP client this project uses to fetch a source sends a descriptive `User-Agent` naming the
application and a contact. Go's default `Go-http-client/1.1` is not acceptable.

Rationale: Wikimedia's User-Agent policy returns HTTP 403 for generic agents, and iNaturalist's
API guidance asks for the same courtesy. Both are relevant when product.speciesPhotographs returns; the rule is stated
now so it is not discovered later.

## goVersionFollowsDeps

**The Go version tracks whatever dependencies require**

Subject: `build.go_version`
Value: `no floor; follow dependencies`
Goals: goals.sizeIntuition

BirdView sets no minimum Go version of its own. When a dependency requires a newer toolchain,
`go mod tidy` raises the `go` directive and the toolchain switches, and that is an accepted
routine change rather than one needing a decision.

Rationale: a minimum-version floor exists to protect *dependents* from a bump they did not ask
for. BirdView exports no library and has no dependents — it is two commands — so the floor would
protect nobody while costing the project the ability to take upstream fixes.

Current state, for the record rather than as a constraint: `go 1.26.9`, raised from 1.25.0 by
`golang.org/x/net` v0.60.0 and standard-library `net/http` security fixes. Module versions stay
pinned in `go.sum`; it is only the language and toolchain versions that float.

## compressedResponses

**Responses are compressed when the client asks**

Subject: `http.content_encoding`
Value: `gzip on Accept-Encoding, Vary on every response`
Goals: goals.sizeIntuition, goals.goodDataCitizen

Every response is gzipped when the request carries `Accept-Encoding: gzip`, and every response
carries `Vary: Accept-Encoding` whether or not it was compressed.

Rationale: the index inlines a full SVG diagram for each of 1,083 species, which is **3.3 MB of
markup compressing to 270 KB — a 92% saving**. Nothing else BirdView could do moves a number
that far. It matters beyond politeness because egress is the one axis of a serverless
deployment that bills: against 1 GiB of free monthly transfer, this is the difference between
about **320 index loads and about 3,980**.

**Done in the application, not left to a CDN or a load balancer**, because the saving should
not depend on which of several deployment shapes is chosen, and because the shape is not chosen
yet.

**`Vary` goes on uncompressed responses too.** Without it a cache that stored the plain reply
would serve it to clients asking for gzip, and — the damaging direction — hand a gzip stream to
a client that cannot read one.

**The headers are decided at `WriteHeader`, not at the first `Write`.** Any handler that sets a
status itself has already committed the header map by the time a byte is written, so a
compressor that labels from `Write` emits a correctly compressed body with no
`Content-Encoding` and the browser renders binary. Both static files and every
[product.undrawnSpeciesAnswer](../product/presentation.md#undrawnspeciesanswer) 404 shipped that way in the first draft of this.

**A status that may carry no body — 204, 304 — is not compressed at all**, an empty gzip
stream's twenty-byte header being bytes the protocol says must not be there.

## cacheLifetimeStated

**Every response says how long it may be reused**

Subject: `serving.cache_validation`
Value: `Cache-Control: no-cache and an ETag on every response`
Goals: goals.sizeIntuition, goals.nothingDisappearsQuietly

Every response carries a `Cache-Control` directive and a validator, so a cache knows whether
it may reuse what it holds and can ask cheaply when it may not.

Rationale: **sending no cache headers is not the same as sending "do not cache".** A response
with no freshness information and no validator lets a browser apply its own heuristic, and
leaves it nothing to revalidate against when it does. The failure is a reader who reloads
after a rebuild, is served the old page, and has no way to tell — which is indistinguishable
from a bug in the page and was reported as one.

**`no-cache` means store it and ask before reusing it**, not "do not store". That is the right
directive for every page here: validation is cheap and staleness is never acceptable, since
the point of a figure like [product.shareOfRecords](../product/presentation.md#shareofrecords)'s is that it is current.

**The ETag is a fingerprint of the build and the URL, not a hash of the body.** Hashing the
body means buffering it, and the index is 3.3 MB before compression — the memory this would
cost on every concurrent request is the thing [compressedResponses](#compressedresponses) exists to avoid spending. The
substitute is sound only because a page is a pure function of the dataset, the templates, the
static files and the request URI, with no clock anywhere in the render path, and
[acceptance.honestRevalidation](../../acceptance.md#honestrevalidation) asserts that rather than trusting it.

**One fingerprint over all assets, not one per asset.** A stylesheet edit invalidates the HTML
that loads it. Over-invalidating costs one refetch after a deploy; under-invalidating serves a
page styled by a stylesheet it was not written for, which is the failure hardest to recognize
as a caching problem.

**The compressed representation carries a different validator**, suffixed rather than dropped.
A gzipped body is a different representation of the resource, and a cache that let the two
share a tag would eventually hand gzip bytes to a client that asked for none — the same class
of error `Vary: Accept-Encoding` exists to prevent, one layer down.

**`-dev` sends `no-store` and no validator at all.** It re-reads templates per request, so no
fingerprint taken at startup stays true, and a stale page under the one flag that exists to
make edits visible would defeat the flag.
