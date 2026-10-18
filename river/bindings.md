# Bindings

How BirdView applies [README.md](README.md). Where the two disagree, this file states the
local amendment and why.

## River adoption — in progress

BirdView adopted specman on 2026-09-23 and is migrating to **River**, its successor. The method
file is River's as of 2026-10-15; the bundle around it is still specman-shaped. Until the
migration finishes, these are deliberate local departures from `README.md`, not drift:

| Area | Current state | Why it has not moved yet |
| --- | --- | --- |
| **Method** | River's `README.md`, pin in its banner | — |
| **Layout** | River's nested layout: `goals/`, `requirements/`, one package per source under `requirements/sources/` ([decisions.nestedLayout](requirements/decisions.md#nestedlayout)) | — |
| **Identifiers** | Descriptive camelCase, qualified by kind outside their own file: `goals.*`, `product.*`, `tech.*`, `decisions.*`, `acceptance.*`; a transcribed source is its own package with exported names, `inatapi.RequestRate` ([decisions.descriptiveIdentifiers](requirements/decisions.md#descriptiveidentifiers)) | — |
| **Composition** | **`app` composed** in River's normalized form: it links the 52 local requirements that apply to the web application unchanged, excludes the four withdrawn photograph requirements, and writes out WCAG's contrast criteria over every web page and the stylesheet's contract. Its criteria for the composed items are in `requirements/app/acceptance.md`; every other criterion still waits in `requirements/acceptance.md` | `dataset` and `internal/repo`, which account for the remaining 24 local requirements; the four mixed requirements in [todo.md](todo.md) are split first |
| **Checks** | All in **Go**, run by `go test`: River's checker, translated at River pin `30ac54a`, with River's shared cases copied at the same pin; BirdView's stricter goal rule; the vendored-file hashes; every quotation against its vendored source; every reference, anchor and relative link; and the composition fields. **The rendered contrast check needs Chrome** (`BIRDVIEW_CHROME` to point at one) and skips without it, so it gates on the owner's machine and nowhere else until there is CI | River's checker does not read composed packages yet, so BirdView checks its own |
| **Stage adoption** | **Requirements**, **Implementation** and **Verification** adopted; **Release informal** | BirdView is not deployed. It has a Dockerfile but no rollout policy, monitoring or alerting — none of the release guardrails River's Release stage produces. Decided 2026-10-16; revisit when the site goes live |

## Project bindings

What this project declares, as [README.md](README.md) asks of every adopter.

| Binding | This project |
| --- | --- |
| **Standing checks** | `go build ./...`, `go vet ./...`, `gofmt -l .` (must print nothing), `go test ./...`. All four, in that order, before any change is called done. |
| **ID citation syntax** | Go line comments, preserving the literal token: `// Verifies: product.scaleToggle, tech.badgeContrast.` |
| **Artifact locations** | `river/`, with one exception below. Not served; the guides link to it in the published repository ([decisions.requirementsNotServed](requirements/decisions.md#requirementsnotserved)). |
| **Subject vocabulary** | Dotted paths, lowercase, e.g. `diagram.scale.mode`, `badge.contrast.ratio`, `dataset.join.key`. New subjects are added by the project owner. |
| **Hard safety constraints** | Never commit anything under `fetched/`. Never commit the built dataset. Never access or process Wiley Online Library material, by any means including an AI agent (decisions.ricklefsProhibited). Never fetch or process anything from BirdLife's DataZone by any automated means (see `requirements/requirements.md` and `todo.md`). |
| **External systems** | Tests make no network requests. Any source fetch is exercised against a local fixture or an `httptest` server, never against the live publisher. |
| **Approval authority** | Sameer — gates, waivers, amendments, source deviations, risk acceptances. |
| **Supported runtime** | Whatever the dependencies require (tech.goVersionFollowsDeps). No minimum Go version is maintained. |
| **Source refresh cadence** | Annually. AviList, AOS and eBird/Clements each revise on roughly yearly cycles, and product.joinOnBinomial's synonym table depends on all three. |
| **Existing docs** | `river/` owns requirements. `PROVENANCE` files under `river/requirements/sources/` own retrieval facts and hashes. A README, when written, owns usage instructions and duplicates neither. |

## Location exception: the third_party directory

**Every vendored third-party file lives under `third_party/<name>/`**, not under
`river/requirements/sources/<name>/`, and its provenance record stays where [README.md](README.md) puts it:
`river/requirements/sources/<name>/PROVENANCE.md` names the out-of-place path and holds its hash. The
directory names match, so `third_party/avonet/` is described by `river/requirements/sources/avonet/`.

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
scanner — can tell what BirdView may relicense from what it merely carries. `river/requirements/sources/`
keeps only BirdView's own writing: the provenance records and the requirement transcriptions.

The source directory still exists for each, so nobody helpfully copies a file back into it and
creates a second copy to drift.
