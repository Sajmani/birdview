# Technical requirements

Constraints on how BirdView is built, and the project bindings that
[process.md](process.md#project-bindings) requires each project to supply.

Normative. Where this file and the code disagree, the code is wrong.

---

## Structure and delivery

### T-001 — Two commands, one module
Subject: `build.layout`
Goals: G-001

A single Go module with an ingest command that produces the dataset, and a server command that
embeds and serves it.

### T-002 — The server is self-contained at runtime
Subject: `runtime.network`
Value: `none`
Goals: G-001

The server binary makes no network request to serve the application. No CDN for styles or scripts,
no remote data fetch.

Rationale: the tool should work on a laptop in a cabin. This supersedes the earlier design's
`cdn.tailwindcss.com` dependency, which would have left the UI unstyled offline.

### T-003 — Serves on localhost:8080
Subject: `server.address`
Value: `localhost:8080`
Goals: G-001

### T-004 — Fetch restricted inputs at build time; commit only what may be redistributed
Subject: `build.inputs`
Goals: G-004

Inputs whose license forbids redistribution — currently the ABA checklist alone — are downloaded
onto the building machine at build time, verified against the hashes recorded in
[sources.md](sources.md), consumed, and never committed. They land under `fetched/`, which
`.gitignore` excludes.

This mechanism is available only where the publisher permits automated retrieval. It is not a
general workaround for a restrictive license: Ricklefs 2017 was rejected partly because Wiley's
terms forbid automated access outright ([CR-002](decisions.md#cr-002)).

The CC BY inputs are committed: the AVONET workbook, and the Fox et al. wingspan data when it is
retrieved. So is the synonym table of P-006, which is the project's own work.

Rationale: the species list BirdView needs is `all rights reserved` and says so in the first row
of its own CSV. Fetching at build keeps the project honest without giving up the data.

### T-005 — The built dataset is an artifact, never committed
Subject: `build.output`
Goals: G-004

The generated dataset is excluded from version control, because it carries values derived from the
non-redistributable inputs of T-004. A clean checkout builds it; it is not shipped in the
repository.

### T-006 — Identify the client on every outbound request
Subject: `http.user_agent`
Goals: G-004

Any HTTP client this project uses to fetch a source sends a descriptive `User-Agent` naming the
application and a contact. Go's default `Go-http-client/1.1` is not acceptable.

Rationale: Wikimedia's User-Agent policy returns HTTP 403 for generic agents, and iNaturalist's
API guidance asks for the same courtesy. Both are relevant when P-050 returns; the rule is stated
now so it is not discovered later.

### T-007 — The Go version tracks whatever dependencies require
Subject: `build.go_version`
Value: `no floor; follow dependencies`
Goals: G-001

BirdView sets no minimum Go version of its own. When a dependency requires a newer toolchain,
`go mod tidy` raises the `go` directive and the toolchain switches, and that is an accepted
routine change rather than one needing a decision.

Rationale: a minimum-version floor exists to protect *dependents* from a bump they did not ask
for. BirdView exports no library and has no dependents — it is two commands — so the floor would
protect nobody while costing the project the ability to take upstream fixes.

Current state, for the record rather than as a constraint: `go 1.25.0`, raised from 1.22 by
excelize v2.11.0, with the toolchain resolving to 1.26.8. Module versions stay pinned in `go.sum`;
it is only the language and toolchain versions that float.

---

## Accessibility

### T-010 — Badge contrast meets WCAG 2.2 SC 1.4.6
Subject: `badge.contrast.ratio`
Value: `>= 7.0`
Goals: G-002

Every generated badge color yields a contrast ratio of at least 7:1 against **the text it actually
carries**, which is black or white per badge, computed per WCAG 2.2 relative luminance.

Requiring black text specifically would be a stricter rule than WCAG asks for, and an expensive
one: it discards the entire dark half of the gamut and costs about a third of the distinguishable
colors available to P-005.

Rationale: WCAG is adopted as a **mandatory** source scoped to the contrast criteria only — see
[sources.md](sources.md#wcag). BirdView makes no broader conformance claim, and saying so is the
point: the earlier design claimed "WCAG AA ≥ 7:1", which conflates AA (4.5:1) with AAA (7:1) and
asserted a conformance level nothing checked. This requirement is checkable and scoped.

---

## Posture

### T-011 — Local now, public later; never commercial
Subject: `deployment.posture`
Value: `local; non-commercial in perpetuity`
Goals: G-004

BirdView runs locally today and may be deployed publicly later. It will not become a commercial
offering, and source obligations are recorded against the public, non-commercial case so that a
later deployment needs no licensing re-litigation.

**Public deployment is gated on written permission from the ABA.** Serving the application
reproduces their checklist — the species list, common names, family names and banding codes — to
every visitor, which the notice in their own file forbids. Keeping the CSV out of the repository
does not address this; the served pages are the reproduction. The request is drafted and awaiting
a demo to accompany it.

Rationale: non-commercial is load-bearing, not incidental. Macaulay Library permits embedding for
non-commercial use only, and iNaturalist's default CC BY-NC corpus is usable on the same basis.
Either would become a licensing breach the day BirdView took revenue.

---

## Process

### T-020 — The spec bundle is normative
Subject: `process.spec_location`
Value: `spec/`
Goals: G-004

The repository follows the loop in [process.md](process.md). `AGENTS.md` names `spec/` as
normative so agents reading the instruction file are pointed at it.

### T-022 — The iNaturalist mapping is built by a separate, paced tool
Goals: G-004

Resolving species to iNaturalist taxon IDs happens in `cmd/birdview-inatmap`, never in
`birdview-ingest` and never while serving. The tool waits at least one second between requests
(`inat-api/R1`) and identifies itself (T-006). Its output is a table committed to the repository;
the ingest joins against that table and makes no network request.

Rationale: 841 requests at the rate iNaturalist asks for is a quarter of an hour, which does not
belong in a build — and a build that needs the network is a build that fails on a plane. Keeping
resolution separate also keeps T-002 and the no-network-in-tests binding intact without exception.
The table going stale is the accepted cost; a stale ID fails visibly and the refresh is one
command.

---

### T-023 — Link and text colors meet a stated contrast floor
Subject: `text.contrast.ratio`
Goals: G-001

Every text color the stylesheet declares for text on a bare surface clears **4.5:1** (WCAG 2.2
SC 1.4.3) against both the page background and the panel background. **Links additionally clear
7:1** (SC 1.4.6) against the page background, which is the surface every plain link sits on.

Rules that set their own background are excluded and governed by T-010 instead: badges and chips
pair a generated background with a matching text color, and the pair is what has to clear 7:1.

**No link may go uncolored.** A link with no rule inherits the browser default — `#0000EE`
unvisited and `#551A8B` visited — which on this background is 1.90:1 and 1.62:1.

Rationale: contrast was adopted for badges only, because badge colors are *generated* and an
unchecked generator is obviously dangerous. Hand-written CSS turned out to be just as dangerous and
had no check at all: the iNaturalist link of P-064 shipped with no color rule and was effectively
invisible. The floor is 4.5:1 rather than 7:1 for ordinary text because `--dim` sits at 6.96:1 on
the page background and 5.71:1 on a panel; raising the floor would mean repainting the secondary
text of the whole interface, which is a bigger change than the defect warrants. That gap is stated
rather than rounded up — see [CR-013](decisions.md).

---

### T-021 — No source is consumed before it is recorded
Goals: G-004

Every input source appears in [sources.md](sources.md) with its tier, pin, license and status
*before* code that reads it is written.

Rationale: this project found, after designing around three datasets, that two of them could not
legally be redistributed and a third did not contain the fields it was chosen for. Recording first
is cheaper than discovering later.

---

## Project bindings

What [process.md](process.md#project-bindings) requires each project to declare.

| Binding | This project |
| --- | --- |
| **Standing checks** | `go build ./...`, `go vet ./...`, `gofmt -l .` (must print nothing), `go test ./...`. All four, in that order, before any change is called done. |
| **ID citation syntax** | Go line comments, preserving the literal token: `// Verifies: P-042, T-010.` |
| **Artifact locations** | `spec/`, per the standard layout, with one exception below. |
| **Subject vocabulary** | Dotted paths, lowercase, e.g. `diagram.scale.mode`, `badge.contrast.ratio`, `dataset.join.key`. New subjects are added by the project owner. |
| **Hard safety constraints** | Never commit anything under `fetched/`. Never commit the built dataset. Never redistribute data derived from the ABA checklist. Never access or process Wiley Online Library material, by any means including an AI agent (CR-002). |
| **External systems** | Tests make no network requests. Any source fetch is exercised against a local fixture or an `httptest` server, never against the live publisher. |
| **Approval authority** | Sameer — gates, waivers, amendments, source deviations, risk acceptances. |
| **Supported runtime** | Whatever the dependencies require (T-007). No minimum Go version is maintained. |
| **Source refresh cadence** | Annually. ABA, AOS and eBird/Clements each revise on roughly yearly cycles, and P-006's synonym table depends on all three. |
| **Existing docs** | `spec/` owns requirements. `PROVENANCE` files under `spec/sources/` own retrieval facts and hashes. A README, when written, owns usage instructions and duplicates neither. |

### Location exception: the AVONET workbook

`//AVONET Supplementary dataset 1.xlsx` lives at the repository root rather than under
`spec/sources/avonet/`, and its provenance record is
[spec/sources/avonet/PROVENANCE.md](sources/avonet/PROVENANCE.md).

The file is a 21.5 MB build input read by the ingest command, not a rules document consulted by a
reader. Burying a build input inside the spec bundle would misrepresent what it is. The source
directory exists and names the out-of-place path, so nobody helpfully copies the workbook into it
and creates a second copy to drift.
