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

Every text color the stylesheet declares for text on a bare surface clears **7:1** (WCAG 2.2
SC 1.4.6, Contrast Enhanced) against **both** the page background and the panel background. Links
and ordinary text are held to the same floor, which is also the floor T-010 holds badges to.

**"Text color" means `fill` as well as `color`.** SVG text is painted with `fill`, so a diagram
label is text and is held to this floor exactly like a paragraph. Shapes are not: a stroke or a
fill on a line, a polygon or the mass cube is a graphical object, governed by WCAG 1.4.11's 3:1
rather than by this requirement. What separates the two is whether the rule styles a text element,
which the stylesheet already declares by giving that element a font property.

That distinction went unwritten, and the check read `color` alone. The wing chord and wingspan
labels of [P-066](product.md#p-066) shipped painted `#0ea5e9` — **6.44:1 on the page background
and 5.28:1 on a panel** — and nothing saw them, because SVG text does not use `color`.

Rules that set their own background are excluded and governed by T-010 instead: badges and chips
pair a generated background with a matching text color, and the pair is what has to clear 7:1.

**No link may go uncolored.** A link with no rule inherits the browser default — `#0000EE`
unvisited and `#551A8B` visited — which on this background is 1.90:1 and 1.62:1.

Rationale: contrast was adopted for badges only, because badge colors are *generated* and an
unchecked generator is obviously dangerous. Hand-written CSS turned out to be just as dangerous and
had no check at all: the iNaturalist link of P-064 shipped with no color rule and was effectively
invisible.

One floor rather than two. The first version of this requirement set 4.5:1 for ordinary text,
because `--dim` sat at 6.96:1 on the page background and 5.71:1 on a panel and repainting the
interface's secondary text looked disproportionate to a single broken link. The owner then reported
the dim text as hard to read too, which settled it: `--dim` and `--accent` moved, and the exception
went with them. See [CR-014](decisions.md).

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

### T-024 — What the project says about itself stays true
Subject: `docs.currency`
Goals: G-003

**Every CC BY source whose output appears on a page is credited on that page.** This is a licence
condition, not a courtesy: AviList, AVONET and Fox are all CC BY 4.0, and attribution is the
thing they ask for in return.

It was not being met. The index page carried full credits, while **`detail.html` and
`compare.html` credited AVONET, Fox and the ABA and omitted AviList** — on pages whose every
English name, scientific name, family and species code comes from AviList. A reader could use
BirdView for an hour and never see the source of the words on the screen.

**User-facing documentation describes the project as it is, not as it was.** `README.md` and
`docs/` are read before the code and are the first thing a new reader trusts, so a stale claim
there is worse than no claim. Two kinds of staleness are forbidden:

- **A count that no longer matches the dataset.** `docs/data.md` claimed 841 species after
  [CR-022](decisions.md#cr-022) cut the list to 839.
- **A source credited for something it no longer supplies.** `README.md` listed the ABA Checklist
  as the source of "species list, names, families" after CR-022 moved names and families to
  AviList. The ABA checklist supplies **scope** and nothing else.

**Before any public deployment, the docs are checked against the spec**, since a release is the
moment the claims become promises to strangers rather than notes to ourselves.

The check is mechanical where it can be — counts and credits are comparable against the dataset
and the page — and a reading task where it cannot. [AC-090](acceptance.md#ac-090) does the
mechanical part, and records plainly that prose accuracy is not something a test can hold.

### T-025 — No personal address is published
Subject: `deploy.contact`
Goals: G-004

No page BirdView serves carries a contact address. Reader feedback goes to the project's issue
tracker, whose URL is a deployment setting (`-feedback`) rather than a constant in the markup.

**Unset publishes nothing**, not a dead link. A site may legitimately go up before its tracker
exists; what it must not do is invite a reader somewhere that does not.

Rationale: the header carried the owner's personal mailbox for the project's whole life, spelled
out in words rather than written plainly. **That obfuscation is the argument against it.** It
was written that way because a plain address on a public page is a harvested mailbox — which is
a reason not to publish the address at all, not a reason to spell it oddly. Beyond the spam, an
issue tracker is the better channel: a bug report in public can be read, searched, corrected by
someone else and closed, where mail to one person is none of those things. The owner's personal
mailbox is not the project's contact point.

*(This requirement does not quote the address it removed, and that is not squeamishness. The
spec is served, so a quotation here would republish on `/spec/tech` exactly what the rule
forbids — which is how [AC-097](acceptance.md#ac-097) first failed.)*

**The `User-Agent` contact stays, and is a different thing.** [T-006](#t-006) requires outbound
requests to name a contact, and iNaturalist's API guidance asks for it specifically. That
address reaches a handful of API operators who may need to reach whoever is making the requests;
it is not published to visitors. Removing it would be discourteous and risks the project being
blocked.

**This extends to addresses that are not ours.** The source manifest recorded a source owner's
mailbox, which the served spec then republished to every visitor. A provenance record names who
owns a source; it does not need to hand out their mail.

### T-026 — Responses are compressed when the client asks
Subject: `http.content_encoding`
Value: `gzip on Accept-Encoding, Vary on every response`
Goals: G-001, G-004

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
[P-075](product.md#p-075) 404 shipped that way in the first draft of this.

**A status that may carry no body — 204, 304 — is not compressed at all**, an empty gzip
stream's twenty-byte header being bytes the protocol says must not be there.

### T-027 — Every join reports what it dropped
Subject: `build.join_loss`
Value: `named and counted on every run; a declared bar above which the run refuses`
Goals: G-004, G-006

Every join between two datasets reports the records it could not match: how many, under what
names, and whether the loss was intended. Where a loss can be bounded, the bar is declared and
the run **refuses to write** rather than exceed it.

Rationale: [G-006](goals.md#g-006) at the point where data actually disappears. A join is the
one place in this pipeline where a record can cease to exist without anyone deciding it should,
and the failure is invisible on both sides — the input still has the record, the output simply
never mentions it.

**The ingest's joins already did this and the region pass did not**, which is exactly where the
loss was. `birdview-ingest` prints what the AviList join dropped and asserts it against a
declared list, prints what AVONET could not match and fails above a threshold, and requires
iNaturalist coverage to be total. `birdview-regions` matched raw facet names against AviList
names and discarded the rest in silence: **37 names, 10,015,688 records, of which 9,954,362
belonged to ten species BirdView draws** ([CR-035](decisions.md#cr-035)).

**A discarded record is worse than absent, because it still counts.** Every one of those names
reached the denominator of [P-074](product.md#p-074)'s share and no species' numerator — so
Hairy Woodpecker's 7,310,342 eBird records were counted as *some other bird* on every tile
including its own, and the bird showed 0.7% of its own record.

**The bar is a fraction of records, not a count of names.** A new vagrant adds a name and six
records; a genus move adds one name and seven million.

**Normalizing is part of the join, not a step before it.** Two commands read the occurrence
data, and the loss above was not a missing table entry — every entry already existed — but a
second reader that never consulted them. So every reader normalizes through **one function**,
and a criterion asserts the two agree on the species rather than asserting either in isolation
([AC-106](acceptance.md#ac-106)). The three gaps [CR-035](decisions.md#cr-035) found surfaced
one at a time over two days precisely because nothing compared the readers, so each fix left
the next one standing: no rename, then no lump, then no reduction of rank.

**Rank is reduced before any name is looked up**, the naming tables being keyed on binomials
that `dryobates villosus harrisi` matches none of. The ordering that is load-bearing the other
way is the hybrid test, which runs *before* the truncation: cut to two tokens, `larus
smithsonianus × hyperboreus` is an ordinary species name that joins cleanly and credits a real
bird with records of something that is not it.

Measured 2026-10-06, the region pass's unmatched residue fell from **0.0431% under 1,754
names to 0.0113% under 727**, moving **387,515 records into 430 species**. Region totals came
back byte-identical, so this reattributes records rather than changing what they are divided
by, and **114 species-region presences were added and none removed** — the only shape the
change can have, since a trinomial either joins a drawn species or stays where it was. The
largest single beneficiaries are the two declared domestic forms, feral Rock Pigeon
(+109,401) and Muscovy Duck (+37,470), and one genuine subspecies recovery, Yellow-rumped
Warbler (+30,274) from `coronata` and `auduboni`.

### T-028 — How each source names a species is measured, not assumed
Subject: `build.name_correspondence`
Value: `committed per-source record counts; a declared ratio band`
Goals: G-004, G-005, G-006

A committed table records how many records each occurrence source holds for every species
BirdView draws, under the name BirdView serves. A species far outside the usual ratio between
the two sources is reported, and every exception is declared with a reason.

Rationale: a name can fail to join in two ways and only one is visible. A name that **misses**
leaves a species at zero, which somebody eventually notices — Bullock's Oriole in no region at
all. A name that **matches the wrong concept** joins cleanly and stays wrong: eBird's
*Setophaga petechia* is AviList's Mangrove Yellow Warbler by string and the undivided Yellow
Warbler by meaning, and **no name-matching check can see it** —
[`data/gbif-taxa.tsv`](../data/gbif-taxa.tsv) files that pair as `exact`, correctly.

What separates them is asking each database how many records it holds. The ratio is
unexpectedly stable — **median 48×, 5th percentile 10.7×, 95th 148×** — because iNaturalist is
uniformly shallower than eBird rather than differently distributed. Every kind of disagreement
shows as an outlier, and in opposite directions:

| shape | signature |
| --- | --- |
| a rename or lump | eBird ≈ 0; its records are under another name |
| a shared name | ratio enormous — *Setophaga petechia* is **117,834×** |
| a withheld species | eBird ≈ 0 under *every* name, iNaturalist normal |
| a domestic form | ratio large — Rock Pigeon 11,809×, iNaturalist filing under a trinomial |

**Only 7 of 1,019 species fall outside 1×–1000×, and every one is a real finding.** Measured
2026-10-05, the band immediately caught two species nothing had recorded as withheld —
**Great Gray Owl** (eBird 1, iNaturalist 2,859) and **Black Rail** (16 against 101) — and
exposed an error in [`sources.md`](sources.md), which listed Blue-billed White-Tern among the
withheld when eBird holds 16,008 records of it under *Gygis alba*.

**The declared exceptions are the point, not an escape hatch.** Each names which of the four
shapes it is, so the table reads as a statement about taxonomy rather than a list of numbers
to ignore.
