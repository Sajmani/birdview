# Bindings

How BirdView applies [README.md](README.md) and the River 2.0 skill at
[../.agents/skills/river/SKILL.md](../.agents/skills/river/SKILL.md). Where they leave a choice to
the project, or where BirdView states a local amendment, this file records it and why.

## River 2.0

BirdView adopted specman on 2026-09-23, migrated to River 1.0 on 2026-10-15, and upgraded to
**River 2.0** on 2026-11-10 ([decisions.riverV2PackagesAndSplits](requirements/decisions.md#riverv2packagesandsplits)).

| Area | State in BirdView |
| --- | --- |
| **Skill and module** | Skill installed at [../.agents/skills/river/](../.agents/skills/river/SKILL.md), pinned in [../.agents/skills/PROVENANCE.md](../.agents/skills/PROVENANCE.md); module declared in [README.md](README.md) (`module birdview`, `river 2.0`). |
| **Layout** | `goals/`, `requirements/`, one package per source under `requirements/local/{product,tech}` or `requirements/imported/<origin>/<name>` ([decisions.nestedLayout](requirements/decisions.md#nestedlayout)), source sums in [requirements/sums.md](requirements/sums.md), and package approvals in [requirements/approvals.md](requirements/approvals.md). |
| **Identifiers** | Descriptive camelCase, qualified by package outside their own file: `goals.*`, `product.*`, `tech.*`, `decisions.*`, `app.*`, `dataset.*`, `repo.*`; a transcribed source is its own package with exported names, [`inatapi.RequestRate`](requirements/imported/inaturalist.org/inatapi/inatapi.md#requestrate) ([decisions.descriptiveIdentifiers](requirements/decisions.md#descriptiveidentifiers)). |
| **Composition** | [app](requirements/composed/internal/app/app.md), [dataset](requirements/composed/internal/dataset/dataset.md), and [repo](requirements/composed/internal/repo/repo.md) under `requirements/composed/internal/`, with criteria in each package's `acceptance.md` and guardrails in [conformance.md](conformance.md). |
| **Checks** | `river check` plus BirdView's Go checks in `river/check/` (`go test ./river/check/`); see below. |
| **Stage adoption** | All four adopted: **Requirements**, **Implementation**, **Verification**, and **Release** since 2026-10-23 (guardrails in [tech/release.md](requirements/local/tech/release.md); the deployment's probe, uptime check, alerting, and runbook live in the private `Sajmani/birdview-deploy`). |

**Checks.** `river check` enforces River 2.0's structural, accounting, sums, approval,
host/dependency discovery, and reachability rules. `river/check/` adds BirdView's stricter goal rule
(`TestEveryLocalRequirementCitesAGoal`), active-issue rule (`TestActiveTodoHasIssue`), vendored
hashes (`TestVendoredHashesMatch`), verbatim quotations (`TestQuotationsAreVerbatim`),
link/reference resolution (`TestReferencesResolve`, `TestRelativeLinksResolve`), composition
cross-checks (`TestCompositionResolves`), and Go function reachability (`TestNoUnreachableCode`,
with its test-oracle allowlist in `river/check/deadcode_test.go`). **The rendered contrast and
phone-width checks need Chrome** (`BIRDVIEW_CHROME`) and skip without it; CI has Chrome and fails on
any skip ([tech.checksOnEveryChange](requirements/local/tech/release.md#checksoneverychange)).

## Project bindings

What this project declares, as the River skill asks of every adopter.

| Binding | This project |
| --- | --- |
| **Standing checks** | `go build ./...`, `go vet ./...`, `gofmt -l .` (must print nothing), `go test ./...`, and `river check --allow-unapproved river` (`river check river` once approved). |
| **ID citation syntax** | Go line comments preserving literal tokens: `// Verifies: product.scaleToggle. Criterion: app.foo.` |
| **Artifact locations** | `river/`, with one exception below. Not served; the guides link to it in the published repository ([decisions.requirementsNotServed](requirements/decisions.md#requirementsnotserved)). |
| **Subject vocabulary** | Dotted paths, lowercase, e.g. `diagram.scale.mode`, `badge.contrast.ratio`, `dataset.join.key`. New subjects are added by the project owner. |
| **Hard safety constraints** | Never commit anything under `fetched/`. Never commit the built dataset. Never access or process Wiley Online Library material, by any means including an AI agent ([decisions.ricklefsProhibited](requirements/decisions.md#ricklefsprohibited)). Never fetch or process anything from BirdLife's DataZone by any automated means (see `requirements/imports.md` and `todo.md`). |
| **External systems** | Tests make no network requests. Any source fetch is exercised against a local fixture or an `httptest` server, never against the live publisher. |
| **Approval authority** | Sameer — gates, waivers, amendments, source deviations, risk acceptances, and `river approve --by Sameer river`. |
| **Supported runtime** | Whatever the dependencies require ([tech.goVersionFollowsDeps](requirements/local/tech/delivery.md#goversionfollowsdeps)). No minimum Go version is maintained. |
| **Source refresh cadence** | Annually. AviList, AOS, and eBird/Clements each revise on roughly yearly cycles, and [product.joinOnBinomial](requirements/local/product/dataset.md#joinonbinomial)'s synonym table depends on all three. |
| **Existing docs** | `river/` owns requirements. `PROVENANCE` files under `river/requirements/imported/` own retrieval facts and hashes. A README, when written, owns usage instructions and duplicates neither. |

## Location exception: the third_party directory

**Every vendored third-party file lives under `third_party/<name>/`**, not under
`river/requirements/imported/<origin>/<name>/`, and its provenance record stays where River puts it:
`river/requirements/imported/<origin>/<name>/PROVENANCE.md` names the out-of-place path and holds its hash. The
directory names match, so `third_party/avonet/` is described by `river/requirements/imported/figshare.com/avonet/`.

| `third_party/` | License | What it is |
| --- | --- | --- |
| `avonet/`, `avilist/`, `foxwingspan/` | CC BY 4.0, full text beside each file | the three build inputs |
| `inatterms/`, `inatapi/` | none granted; © iNaturalist | terms of two retired sources, kept so the quotations in their record can be checked |
| `wcag/` | W3C Document License; copied unmodified | the accessibility rules BirdView quotes, kept so each quotation can be checked |

Two reasons, and the second is why this moved from a one-file exception to the rule.

The workbooks are build inputs read by the ingest, not rules documents consulted by a reader, and
burying a 21.5 MB spreadsheet inside the spec bundle would misrepresent what it is. That was the
original exception, for AVONET alone.

**And BirdView is published under Apache 2.0, which none of these files are.** A `LICENSE` at
the repository root reads as covering everything beneath it. Putting third-party material in its
own tree, each directory with its own license and README, is how a reader — and a license
scanner — can tell what BirdView may relicense from what it merely carries. `river/requirements/imported/`
keeps only BirdView's own writing: the provenance records and the requirement transcriptions.

The source directory still exists for each, so nobody helpfully copies a file back into it and
creates a second copy to drift.
