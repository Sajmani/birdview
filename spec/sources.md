# Source manifest

Every body of rules or data BirdView is subject to, what tier it sits at, and how far adoption has
actually got. `Status` is the field that distinguishes an adopted source from a named one — all
the others can be filled in plausibly from a URL alone.

`Integrity` names **where** a content hash lives. It never contains one.

---

## Summary

| Name | Tier | Import | Status |
| --- | --- | --- | --- |
| [`avonet`](#avonet) | mandatory | by reference, vendored | **adopted** — pinned, vendored, hashed; no criteria yet |
| [`aba`](#aba) | mandatory | by reference, **not** vendored | **partial** — terms read and quoted, pinned by hash; redistribution refused, so fetched at build |
| [`fox-wingspan`](#fox-wingspan) | mandatory | by reference, vendored | **named** — license confirmed CC BY 4.0, **not yet retrieved** |
| [`ricklefs`](#ricklefs) | mandatory | **rejected** | **rejected** — Wiley terms prohibit the use, see CR-002 |
| [`wiley`](#wiley) | mandatory | quoted only, not vendored | **adopted** — terms read and transcribed; the reason `ricklefs` is rejected |
| [`wcag`](#wcag) | mandatory | absorbed, partial | **partial** — adopted for two criteria, not vendored |
| [`specman`](#specman) | — (method) | project copy | **adopted** — copied, base pin recorded below |

Two of the seven are not fully adopted, and neither blocks v1. `ricklefs` was evaluated and
**rejected**: its platform terms prohibit extraction, integration, storage and redistribution
alike. Its entry is kept so the evaluation is not repeated.

---

## `avonet`

| Field | Value |
| --- | --- |
| **Origin** | https://doi.org/10.6084/m9.figshare.16586228 |
| **Version** | v7 |
| **Integrity** | [spec/sources/avonet/PROVENANCE.md](sources/avonet/PROVENANCE.md) |
| **Tier** | mandatory (copyright license) |
| **Import style** | By reference; vendored at `//AVONET Supplementary dataset 1.xlsx` under the location exception in [tech.md](tech.md#location-exception-the-avonet-workbook) |
| **Scope** | All morphometric traits (P-002, P-003) |
| **Adopted parts** | Sheet `AVONET2_eBird` only |
| **Exclusions** | Supplementary dataset 2 (specimen repeatability data) — carries no species-level rows and neither of the dimensions BirdView needs. Removed from the working tree; re-obtainable from the same deposit. |
| **Owner** | Tobias et al. / figshare |
| **Status** | **Adopted.** Pinned, vendored, hashed. No acceptance criteria yet. |

License: **CC BY 4.0**, recorded on the deposit. Redistribution is permitted with attribution,
which is why this is the one data input that may be committed.

Required citation: Tobias, J. A. et al. 2022. AVONET: morphological, ecological and geographical
data for all birds. *Ecology Letters* 25: 581–597.

---

## `aba`

| Field | Value |
| --- | --- |
| **Origin** | https://www.aba.org/aba-checklist/ — CSV at `https://www.aba.org/wp-content/uploads/2026/03/ABA_Checklist-8.19.csv.zip` |
| **Version** | 8.19, January 2026 |
| **Integrity** | [spec/sources/aba/PROVENANCE.md](sources/aba/PROVENANCE.md) |
| **Tier** | mandatory (copyright) |
| **Import style** | By reference, **deliberately not vendored** — fetched at build per T-004 |
| **Scope** | Species list, common and scientific names, families, rarity filter (P-001, P-004) |
| **Adopted parts** | Main checklist only. Provisional Lists A and B are not used. |
| **Exclusions** | Rarity codes are not carried into the dataset or the UI; they are read at ingest as a scope filter and discarded. |
| **Owner** | American Birding Association, `info@aba.org` |
| **Status** | **Adopted.** Retrieved and hashed 2026-09-21; terms read and quoted; redistribution refused by those terms, so the file is fetched at build and never committed. |

License, quoted verbatim from row 0 of the CSV itself:

> "The ABA Checklist is a copyrighted work owned by American Birding Association, Inc. and cannot
> be reproduced without the express written permission of American Birding Association, Inc."

No other terms document governs it: `aba.org/terms-of-use/`, `/copyright/` and `/legal/` all 404,
and the checklist landing page carries no license statement at all.

**Open action:** ask ABA for written permission. It is the only input v1 genuinely depends on that
cannot be committed, and a non-commercial, attributed, open-source request costs an email.

---

## `fox-wingspan`

| Field | Value |
| --- | --- |
| **Origin** | https://doi.org/10.6084/m9.figshare.32089083 |
| **Version** | To be pinned on retrieval |
| **Integrity** | To be recorded in `spec/sources/fox-wingspan/PROVENANCE.md` on retrieval |
| **Tier** | mandatory (copyright license) |
| **Import style** | By reference, vendored — CC BY 4.0 permits it |
| **Scope** | Empirical wingspan (P-060) |
| **Adopted parts** | Rows with `needs_prediction = 0` only |
| **Exclusions** | **Predicted wingspans are excluded.** The dataset's `wingspan_predicted_*` columns are regressions from AVONET wing chord; adopting them would reintroduce exactly the modeled-as-measured problem CR-001 resolves. |
| **Owner** | Fox, Brunton & Shephard |
| **Status** | **Named, not retrieved.** License confirmed CC BY 4.0 from the figshare record. The file has not been downloaded, hashed, or parsed by this project. |

Coverage warning carried over from research: roughly 1,424 of 10,618 species have an empirical
value, and a 20-species spot check of common North American landbirds found only 6. Expect the
measured fraction to be thin.

---

## `ricklefs`

| Field | Value |
| --- | --- |
| **Origin** | https://doi.org/10.1002/ecy.1783 — Ricklefs 2017, *Ecology* 98:1472 |
| **Version** | Unpinned |
| **Integrity** | None — nothing vendored |
| **Tier** | mandatory (copyright) |
| **Import style** | **None — rejected.** Fetch-at-build was the plan; Wiley's Rules of Use G prohibits automated access |
| **Scope** | Passerine total length (P-060) |
| **Adopted parts** | `Length` column, passerines |
| **Exclusions** | — |
| **Owner** | Ricklefs / Wiley / ESA |
| **Status** | **Rejected 2026-09-21.** The terms were read and prohibit extraction, integration with other data, storage in a retrieval system, and redistribution — explicitly regardless of non-commercial use. See [`wiley`](#wiley) and [CR-002](decisions.md#cr-002). |

Two further cautions, both from research rather than from the terms:

- The `RS-eco/traitdata` R package redistributes this data under GPL-3, but a package license does
  not establish the data's license. That redistribution is not evidence of permission.
- Measurements are from museum study skins, which stretch or compress relative to a live bird. It
  is a real published measurement of a different quantity than a field guide's "length".

---

## `wiley`

| Field | Value |
| --- | --- |
| **Origin** | https://onlinelibrary.wiley.com/terms-and-conditions#vor |
| **Version** | Retrieved 2026-09-21; page footer 9/21/26 5:29 PM. Wiley reserves the right to change these terms at any time. |
| **Integrity** | [spec/sources/wiley/PROVENANCE.md](sources/wiley/PROVENANCE.md) — records the hash of the maintainer's local printout, **not** a vendored file |
| **Tier** | mandatory |
| **Import style** | **Quoted only.** The document is deliberately not vendored — see below |
| **Scope** | Any use of material from Wiley Online Library; in practice, Ricklefs 2017 |
| **Adopted parts** | Rules of Use A.2, B, C, F, G, H |
| **Exclusions** | Subscription, PPV and e-commerce terms; warranty limitations; privacy policy |
| **Owner** | John Wiley & Sons, Inc. |
| **Status** | **Adopted.** Retrieved by a human, transcribed with quotations, and acted on — it is why `ricklefs` is rejected. |

**This source is not vendored, and the departure is deliberate.** Rules of Use B forbids storing
Wiley material "in any form or medium in a retrieval system" and F forbids incorporating it "into
any institutional or other repository under any circumstances". A git repository is both.
Committing the terms in order to evidence compliance with the terms would breach them. Rules of
Use C permits "brief quotations for purposes of comment, criticism or similar scholarly
purposes", so the transcription quotes the governing clauses and the document stays out.

Recorded here as a location and import-style exception, in the same spirit as the AVONET workbook
exception in [tech.md](tech.md#location-exception-the-avonet-workbook).

---

## `wcag`

| Field | Value |
| --- | --- |
| **Origin** | https://www.w3.org/TR/WCAG22/ |
| **Version** | WCAG 2.2 |
| **Integrity** | Not vendored |
| **Tier** | **mandatory** |
| **Import style** | Absorbed, partial — the adopted criteria are restated as T-010 |
| **Scope** | Family badge colors only |
| **Adopted parts** | SC 1.4.3 (Contrast Minimum) and SC 1.4.6 (Contrast Enhanced). BirdView targets 1.4.6, 7:1. |
| **Exclusions** | **Everything else.** No keyboard-access, focus-order, timing, reading-level or other criteria are adopted, and BirdView claims no conformance level. |
| **Owner** | W3C; project owner grants any deviation |
| **Status** | **Partial by design.** Adopted for two criteria, restated as a checkable requirement, not vendored. |

Scoping this narrowly is deliberate. The earlier design claimed "a minimum WCAG AA contrast ratio
of ≥ 7:1", which conflates two conformance levels and asserts an accessibility posture nothing
verified. A narrow adopted scope that is actually checked is worth more than a broad claim that
is not.

---

## `specman`

| Field | Value |
| --- | --- |
| **Origin** | https://github.com/Sajmani/specman — `process.md` |
| **Version** | Base pin `fe2762837b398a44c6fbf62068774036f65d45557dab75045766647a6d94f556`, upstream `2c41af8`, taken 2026-09-20 |
| **Integrity** | Not applicable — see below |
| **Tier** | — it states how requirements are made, and produces none for criteria to cite |
| **Import style** | Project copy at [spec/process.md](process.md), carrying a banner that points here |
| **Scope** | How changes to this repository are worked |
| **Owner** | Project owner |
| **Status** | **Adopted.** Copied with banner; base pin recorded here and nowhere else. |

The `Version` field above is a **base pin, not an integrity hash**. The project copy is
*supposed* to differ from it: it carries a banner, and may carry `[LOCAL]` amendments. A diff
against the pin should contain the banner and nothing else but `[LOCAL]` rows; anything further
is drift.

---

## Considered and not adopted

Recorded so nobody re-litigates them annually.

| Source | Why not |
| --- | --- |
| **eBird / Clements taxonomy** | Cornell terms clause 2 forbids reproduction or redistribution without written authorization. Superseded anyway: ABA supplies common names, scientific names and families, so no second taxonomy is needed (P-004). |
| **eBird API region species lists** | Would supply scope and geography in AVONET's own taxonomy, but requires an API key and `api.ebird.org/robots.txt` is `Disallow: /`, with the API terms page returning 403. Not evaluated further; a human would need to read the terms. |
| **AOS / NACC check-list** | Best-shaped data of the three checklists — family, order and subfamily as real columns — but covers North *and Middle* America (2,212 species), which is the wrong scope, and carries no license statement anywhere. Absence of a statement is not permission. |
| **iNaturalist species counts** | Would give regional scope plus an abundance signal, but observation counts overstate rarities: a mega-rarity draws every birder in the state, so the signal inverts exactly where it matters. |
| **Lislevand et al. 2007** | Copyright restrictions: none — genuinely open. Does not contain total length or wingspan, which is what it was originally brought in for. Removed from the project. |
| **EltonTraits, Amniote DB, BIRDBASE, Dunning** | None carries body length or wingspan. BIRDBASE was checked column by column against a claim that it did. |
| **Audubon field guide** | Carries exactly the two needed dimensions, in near-parseable form. Terms forbid reproduction, distribution, storage in a retrieval system, and derivative works. |
| **All About Birds / Birds of the World** | Carries the data; blocks automated retrieval (403) and is subscription-gated respectively. Not assessed further. |
| **Macaulay Library** | Permits *embedding* for non-commercial use — which BirdView qualifies for — but offers no discovery API, and its media-request path explicitly prunes species lists over 100. Cannot scale to the dataset. Viable later for hand-curated species. |
| **Wikimedia Commons, iNaturalist, Flickr** | Photo sources, all researched in detail. Out of scope while P-050 is withdrawn; Commons is the recommendation when it returns. |
