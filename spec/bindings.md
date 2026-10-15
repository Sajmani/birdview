# Bindings

How BirdView applies [process.md](process.md). Where the two disagree, this file states the
local amendment and why.

## River adoption — in progress

BirdView adopted specman on 2026-09-23 and is migrating to **River**, its successor. The method
file is River's as of 2026-10-15; the bundle around it is still specman-shaped. Until the
migration finishes, these are deliberate local departures from `process.md`, not drift:

| Area | Current state | Why it has not moved yet |
| --- | --- | --- |
| **Method** | River's `process.md`, pin in its banner | — |
| **Layout** | specman's flat `spec/` | The document server serves flat names only and refuses `/` in a path, and a test asserts the vendored terms under `spec/sources/` are never served (T-004). River's nested layout is a server change with a licence property on it, not a file move |
| **Identifiers** | `P-###`, `T-###`, `G-###`, `AC-###`, `CR-###` | 228 identifiers carry about 3,000 references, over 1,000 of them in Go comments, log output and generated data files |
| **Composition** | Not run. `acceptance.md` sits beside the sources | River puts criteria in composed packages; the split into `app`, `dataset` and `internal/repo` is a judgment still to make |
| **Checks** | specman's `goals.py`, `pins.py`, `refs.py` | They parse the legacy layout and IDs, and they work |
| **Stage adoption** | Not yet declared | River asks each stage to be adopted or declared informal. To be decided by the project owner |

## Project bindings

What this project declares, as [process.md](process.md) asks of every adopter.

| Binding | This project |
| --- | --- |
| **Standing checks** | `go build ./...`, `go vet ./...`, `gofmt -l .` (must print nothing), `go test ./...`. All four, in that order, before any change is called done. |
| **ID citation syntax** | Go line comments, preserving the literal token: `// Verifies: P-042, T-010.` |
| **Artifact locations** | `spec/`, per the standard layout, with one exception below. |
| **Subject vocabulary** | Dotted paths, lowercase, e.g. `diagram.scale.mode`, `badge.contrast.ratio`, `dataset.join.key`. New subjects are added by the project owner. |
| **Hard safety constraints** | Never commit anything under `fetched/`. Never commit the built dataset. Never access or process Wiley Online Library material, by any means including an AI agent (CR-002). Never fetch or process anything from BirdLife's DataZone by any automated means (see `sources.md` and `todo.md`). |
| **External systems** | Tests make no network requests. Any source fetch is exercised against a local fixture or an `httptest` server, never against the live publisher. |
| **Approval authority** | Sameer — gates, waivers, amendments, source deviations, risk acceptances. |
| **Supported runtime** | Whatever the dependencies require (T-007). No minimum Go version is maintained. |
| **Source refresh cadence** | Annually. AviList, AOS and eBird/Clements each revise on roughly yearly cycles, and P-006's synonym table depends on all three. |
| **Existing docs** | `spec/` owns requirements. `PROVENANCE` files under `spec/sources/` own retrieval facts and hashes. A README, when written, owns usage instructions and duplicates neither. |

## Location exception: the third_party directory

**Every vendored third-party file lives under `third_party/<name>/`**, not under
`spec/sources/<name>/`, and its provenance record stays where [process.md](process.md) puts it:
`spec/sources/<name>/PROVENANCE.md` names the out-of-place path and holds its hash. The
directory names match, so `third_party/avonet/` is described by `spec/sources/avonet/`.

| `third_party/` | License | What it is |
| --- | --- | --- |
| `avonet/`, `avilist/`, `fox-wingspan/` | CC BY 4.0, full text beside each file | the three build inputs |
| `inat-terms/`, `inat-api/` | none granted; © iNaturalist | terms BirdView quotes, kept so each quotation can be checked |

Two reasons, and the second is why this moved from a one-file exception to the rule.

The workbooks are build inputs read by the ingest, not rules documents consulted by a reader, and
burying a 21.5 MB spreadsheet inside the spec bundle would misrepresent what it is. That was the
original exception, for AVONET alone.

**And BirdView is published under Apache 2.0, which none of these files are.** A `LICENSE` at
the repository root reads as covering everything beneath it. Putting third-party material in its
own tree, each directory with its own license and README, is how a reader — and a license
scanner — can tell what BirdView may relicense from what it merely carries. `spec/sources/`
keeps only BirdView's own writing: the provenance records and the requirement transcriptions.

The source directory still exists for each, so nobody helpfully copies a file back into it and
creates a second copy to drift.
