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

