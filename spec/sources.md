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
| [`fox-wingspan`](#fox-wingspan) | mandatory | by reference, vendored | **adopted** — retrieved, hashed against the publisher's MD5, vendored |
| [`ricklefs`](#ricklefs) | mandatory | **rejected** | **rejected** — Wiley terms prohibit the use, see CR-002 |
| [`wiley`](#wiley) | mandatory | quoted only, not vendored | **adopted** — terms read and transcribed; the reason `ricklefs` is rejected |
| [`wcag`](#wcag) | mandatory | absorbed, partial | **partial** — adopted for two criteria, not vendored |
| [`inat-api`](#inat-api) | governing | by reference, vendored | **adopted** — copied from `birdsync` and hash-verified; governs the mapping tool |
| [`inat-terms`](#inat-terms) | mandatory | by reference, vendored | **adopted** — copied from `birdsync` and hash-verified; almost nothing in it reaches BirdView |
| [`specman`](#specman) | — (method) | project copy | **adopted** — copied, base pin recorded below |

One of the nine is not fully adopted, and it does not block v1. `ricklefs` was evaluated and
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

## `inat-api`

| Field | Value |
| --- | --- |
| **Origin** | <https://www.inaturalist.org/pages/api+recommended+practices> |
| **Version** | Page's own revision line: "Revised on February 27, 2025 06:23 PM by kueda" |
| **Integrity** | [spec/sources/inat-api/PROVENANCE.md](sources/inat-api/PROVENANCE.md) |
| **Tier** | governing (the operator can throttle or block a client that ignores it) |
| **Import style** | By reference, vendored — **copied from the maintainer's `birdsync` repository and verified byte-for-byte against the SHA-256 recorded there** |
| **Scope** | Every request `cmd/birdview-inatmap` makes to `api.inaturalist.org` |
| **Adopted parts** | `inat-api/R1`–`R8`, in [requirements.md](sources/inat-api/requirements.md) |
| **Exclusions** | Bulk access routes — observation exports, the weekly GBIF dataset — which serve consumers of observation data. BirdView reads one integer per species from the taxonomy. |
| **Owner** | iNaturalist |
| **Status** | **Adopted 2026-09-23.** Vendored and hash-verified; transcribed against BirdView's much narrower use. |

Two of the eight are implemented (`R1` pacing as T-022, `R4` User-Agent as the pre-existing
T-006); two are not applicable because BirdView never pages; four are satisfied by construction
because it never authenticates, never posts and downloads no media. `inat-api/R7` is worth noting:
it is the one requirement `birdsync` records as a knowing departure, and BirdView satisfies it
only because its needs are so much smaller.

The human-facing page cannot be retrieved by an agent — `www.inaturalist.org/pages/*` answers with
HTTP 403 and a JavaScript challenge. The API itself serves an identified client normally.

---

## `inat-terms`

| Field | Value |
| --- | --- |
| **Origin** | <https://www.inaturalist.org/pages/terms>, <https://www.inaturalist.org/pages/community+guidelines>, and the linked <https://www.inaturalist.org/pages/machine_generated_content> |
| **Version** | Retrieved 2026-08-09 from a browser; hashes in [PROVENANCE.md](sources/inat-terms/PROVENANCE.md) |
| **Integrity** | [spec/sources/inat-terms/PROVENANCE.md](sources/inat-terms/PROVENANCE.md) |
| **Tier** | mandatory (copyright and platform terms) |
| **Import style** | By reference, vendored — **copied from `birdsync` and hash-verified**, as above |
| **Scope** | BirdView's use of the platform: unauthenticated reads, and outbound links to taxon pages |
| **Adopted parts** | `inat-terms/R1`–`R6`, in [requirements.md](sources/inat-terms/requirements.md) |
| **Exclusions** | Conduct rules unreachable by a client that posts nothing: harassment, sockpuppets, explicit content, commercial AI training |
| **Owner** | iNaturalist |
| **Status** | **Adopted 2026-09-23.** Vendored so that one negative finding is checkable rather than remembered. |

BirdView is subject to very little of this, and the entry exists to say which little. **The
adoption rests on a negative: nothing in either document restricts reusing a taxon identifier or
linking to a taxon page.** `terms.html` contains no occurrence of *taxon*, *taxonomy*, *database*,
*derivative*, *hyperlink* or *scrape*. The CC BY-NC clause governs Content a contributor posts —
observations and photographs — not iNaturalist's taxonomy.

That is exactly the kind of claim that is worthless asserted from memory, which is why hashed
copies are in the tree: anyone can search them and check.

Photographs stay on iNaturalist. BirdView links; it never copies, caches, hotlinks or embeds an
image, which [T-002](tech.md) independently forbids anyway. See
[CR-012](decisions.md) for the adoption reasoning and what was rejected.

---

## `fox-wingspan`

| Field | Value |
| --- | --- |
| **Origin** | https://doi.org/10.6084/m9.figshare.32089083 |
| **Version** | v1, published 2026-04-24 |
| **Integrity** | [spec/sources/fox-wingspan/PROVENANCE.md](sources/fox-wingspan/PROVENANCE.md) |
| **Tier** | mandatory (copyright license) |
| **Import style** | By reference, vendored — CC BY 4.0 permits it |
| **Scope** | Empirical wingspan (P-060) |
| **Adopted parts** | Rows with `needs_prediction = 0` only |
| **Exclusions** | **Predicted wingspans are not ingested as values.** The `wingspan_predicted_*` columns are regressions from AVONET wing chord; taking them as data would reintroduce the modeled-as-measured problem CR-001 resolves. They are used for one other purpose: as the reference BirdView's own fitted model is validated against ([CR-008](decisions.md#cr-008)). Reading them to check our arithmetic is not the same as presenting them as measurements. |
| **Owner** | Fox, Brunton & Shephard |
| **Status** | **Adopted 2026-09-21.** Retrieved, hashed, vendored; its MD5 matches figshare's published `computed_md5`. Coverage for BirdView's scope measured and recorded. |

Coverage, now measured rather than estimated: of BirdView's 841 in-scope species, **301 (35.8%)
have an empirical wingspan**, 510 have only a predicted value that this project excludes, and 30
have no row at all — every one a Galliform, an order the dataset omits entirely.

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
| **Scope** | Family badge colors (T-010), and link and text colors against the page surfaces (T-023) |
| **Adopted parts** | SC 1.4.3 (Contrast Minimum) and SC 1.4.6 (Contrast Enhanced). Badges and links target 1.4.6 at 7:1; other text is held to 1.4.3 at 4.5:1. |
| **Exclusions** | **Everything else.** No keyboard-access, focus-order, timing, reading-level or other criteria are adopted, and BirdView claims no conformance level. The scope widened once, from badges to all text, after an unstyled link shipped at 1.9:1 — see [CR-013](decisions.md). |
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
| **Version** | Base pin `63bfdecb99a256d2819aa6d0e55bfb0bccf19b96db40e28f72572315e81cbdbc`, upstream `2c41af8`, taken 2026-09-23 |
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
