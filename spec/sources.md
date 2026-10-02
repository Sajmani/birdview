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
| [`bohmer-neck`](#bohmer-neck) | mandatory | consulted, not ingested | **adopted** — CC BY 4.0 verified; calibrates one constant, contributes no data |
| [`ricklefs`](#ricklefs) | mandatory | **rejected** | **rejected** — Wiley terms prohibit the use, see CR-002 |
| [`wiley`](#wiley) | mandatory | quoted only, not vendored | **adopted** — terms read and transcribed; the reason `ricklefs` is rejected |
| [`wcag`](#wcag) | mandatory | absorbed, partial | **partial** — adopted for two criteria, not vendored |
| [`inat-api`](#inat-api) | governing | by reference, vendored | **adopted** — copied from `birdsync` and hash-verified; governs the mapping tool |
| [`inat-terms`](#inat-terms) | mandatory | by reference, vendored | **adopted** — copied from `birdsync` and hash-verified; almost nothing in it reaches BirdView |
| [`specman`](#specman) | — (method) | project copy | **adopted** — copied, base pin recorded below |

One of the ten is not fully adopted, and it does not block v1. `ricklefs` was evaluated and
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
| **Geographic scope** | **Worldwide** — 10,661 species across 41 orders, every continent. Measured 2026-10-01. |
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
| **Geographic scope** | **Worldwide** — 10,618 species. Its gaps are functional, not regional: flightless birds (Ostrich, Emu, penguins, kiwi) and Galliformes are absent, while flying birds from every continent are present. Measured 2026-10-01. |
| **Adopted parts** | Rows with `needs_prediction` false only. The column is a spreadsheet boolean: `FALSE`/`TRUE` as excelize renders it, `0`/`1` in the underlying XML, and the reader accepts both. Matching only `"0"` yields **zero** measured wingspans and looks like a coverage problem rather than a parsing one — see `isEmpirical`. |
| **Exclusions** | **Predicted wingspans are not ingested as values.** The `wingspan_predicted_*` columns are regressions from AVONET wing chord; taking them as data would reintroduce the modeled-as-measured problem CR-001 resolves. They are used for one other purpose: as the reference BirdView's own fitted model is validated against ([CR-008](decisions.md#cr-008)). Reading them to check our arithmetic is not the same as presenting them as measurements. |
| **Owner** | Fox, Brunton & Shephard |
| **Status** | **Adopted 2026-09-21.** Retrieved, hashed, vendored; its MD5 matches figshare's published `computed_md5`. Coverage for BirdView's scope measured and recorded. |

Coverage, now measured rather than estimated: of BirdView's 841 in-scope species, **301 (35.8%)
have an empirical wingspan**, 510 have only a predicted value that this project excludes, and 30
have no row at all — every one a Galliform, an order the dataset omits entirely.

Worldwide the empirical rate is **1,424 of 10,618 (13.4%)**, so BirdView's current scope is
served about **2.7× better than the global average**. That is a fact about where ornithology has
been done, and it is worth knowing before any change of scope: widening beyond North America
would not break the pipeline, but it would thin the measured wingspans sharply.

---

## `bohmer-neck`

| Field | Value |
| --- | --- |
| **Origin** | https://doi.org/10.1098/rsos.181588 — data at https://doi.org/10.6084/m9.figshare.8025971 |
| **Version** | figshare v1, file `rsos181588_si_001.docx`, 706,190 bytes |
| **Integrity** | **Not vendored, so no hash here.** The figshare DOI pins the version, and the file is re-downloadable. |
| **Tier** | mandatory (copyright license) |
| **Import style** | **Consulted, not ingested.** No value from it enters the dataset. |
| **Scope** | Calibration evidence only: the mass exponent of [P-047](product.md#p-047) |
| **Geographic scope** | Worldwide — 103 species across 68 families and 34 orders, MNHN Paris skeletal collection, one specimen per species |
| **Adopted parts** | Table S4's published scaling, used to choose one constant. Nothing is joined to BirdView's species. |
| **Exclusions** | Everything else. In particular the per-species neck lengths are **not** used as BirdView's neck values — Böhmer measures the full cervical column, including the part inside the body contour, where [P-045](product.md#p-045) models the *exposed* neck of a silhouette. The two are different quantities and must not be equated; see [CR-021](decisions.md#cr-021). |
| **Owner** | Böhmer, Rössner, Fabre et al. |
| **Status** | **Adopted 2026-10-02 as calibration evidence.** CC BY 4.0 verified from the figshare API record itself, not inferred from the article. |

Why it is a source at all when no number from it is stored: [CR-021](decisions.md#cr-021) changes a
constant in the trunk model, and the evidence for that constant has to be citable. A source entry
is the honest place for it. The alternative — a magic number in a code comment saying "measured
elsewhere" — is exactly what this manifest exists to prevent.

What was measured from it, by regressing each log₁₀ dimension on log₁₀ body mass (n = 103):

| dimension | exponent | 95% CI |
| --- | ---: | :---: |
| neck length | 0.380 | 0.338 – 0.421 |
| total leg length | 0.364 | 0.328 – 0.401 |
| tibiotarsus | 0.356 | 0.321 – 0.391 |
| tarsometatarsus | 0.419 | 0.360 – 0.477 |
| femur | 0.325 | 0.297 – 0.352 |
| central cervical vertebra | 0.314 | 0.280 – 0.348 |

**Single bones scale near isometry; multi-element chains scale with positive allometry.** That
split is the finding, and total length is a chain.

Units in the supplement are unstated. They are kilograms and metres, recovered two independent
ways — *Amazona amazonica* gives (10^−0.13874)³ = 0.383 kg and a 67.6 mm neck. The fit does not
depend on that recovery being right: a slope is invariant under a change of units.

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

## Where geography enters

Only in one place. Recorded because it is the question the ABA decision turns on, and the answer
is not obvious from the table above.

| Source | Geographic scope |
| --- | --- |
| **AVONET** | worldwide — 10,661 species, 41 orders |
| **Fox wingspan** | worldwide — 10,618 species; gaps are flightless birds and Galliformes, not regions |
| **iNaturalist taxon IDs** | worldwide taxonomy; BirdView stores one id per in-scope species |
| **ABA checklist** | **the ABA Area** — North America north of Mexico, plus Hawaii |

So **the ABA checklist is the only thing in BirdView that knows where it is.** Everything else is
global and merely gets subsetted by whatever species list is in force. Changing scope therefore
costs nothing in the trait pipeline — no ingest, model or diagram depends on the region — and
everything in the list itself. The one measurable consequence is coverage: empirical wingspan runs
at 35.8% inside the current scope against 13.4% worldwide, so a wider list would be a thinner one.

---

## Considered and not adopted

Recorded so nobody re-litigates them annually.

| Source | Why not |
| --- | --- |
| **eBird / Clements taxonomy** | Cornell terms clause 2 forbids reproduction or redistribution without written authorization. Superseded anyway: ABA supplies common names, scientific names and families, so no second taxonomy is needed (P-004). |
| **eBird API region species lists** | Would supply scope and geography in AVONET's own taxonomy, but requires an API key and `api.ebird.org/robots.txt` is `Disallow: /`, with the API terms page returning 403. Not evaluated further; a human would need to read the terms. |
| **AOS / NACC check-list** | Best-shaped data of the three checklists — family, order and subfamily as real columns, plus machine-readable status flags — but covers North *and Middle* America, which is the wrong scope, and carries no license statement anywhere. Absence of a statement is not permission. Re-checked 2026-10-01 against the live CSV: 2,212 species, and filtering to neither accidental nor extinct leaves **1,957** — 1,131 more than the ABA area's 841, because the area runs to Panama and the West Indies. The status flags cannot reconstruct ABA scope; they are relative to a far larger region. |
| **AviList v2025b** | **CC BY 4.0**, verified on the publisher's own download page 2026-10-01, and the consensus successor to IOC, Clements and BirdLife with the AOS, Cornell, BirdLife, Avibase and the IOU behind it. It would solve names, families and taxonomic sequence outright — 832 of BirdView's 841 match on scientific name. It does **not** solve scope, which is the only thing ABA is actually blocking, and its English names are wrong for this audience. See [below](#why-avilist-does-not-replace-the-aba-checklist). |
| **iNaturalist species counts** | Would give regional scope plus an abundance signal, but observation counts overstate rarities: a mega-rarity draws every birder in the state, so the signal inverts exactly where it matters. Superseded as a candidate by GBIF below, which carries eBird and is far deeper for birds. |
| **GBIF — eBird Observation Dataset** | **CC BY 4.0**, verified from GBIF's dataset record 2026-10-01, no key required, 1.07 billion US records. This is *not* the eBird API rejected above: that was blocked by `robots.txt` and unreadable terms, while the same observations published through GBIF are openly licensed. Technically excellent and the only serious candidate for deriving scope. It still does not work — see [below](#why-occurrence-data-cannot-reproduce-the-aba-scope). |
| **iNaturalist taxonomy** (as a names source) | Distinct from the taxon IDs already adopted, and not covered by them. The `inat-terms` adoption rests on a narrow negative — nothing restricts reusing an *identifier* — while the same terms state the agreement "does not transfer to You any iNaturalist or third party intellectual property", and the CC BY-NC default governs contributor Content, not the taxonomy. No affirmative grant to republish their name list was found. |
| **Lislevand et al. 2007** | Copyright restrictions: none — genuinely open. Does not contain total length or wingspan, which is what it was originally brought in for. Removed from the project. |
| **EltonTraits, Amniote DB, BIRDBASE, Dunning** | None carries body length or wingspan. BIRDBASE was checked column by column against a claim that it did. |
| **Audubon field guide** | Carries exactly the two needed dimensions, in near-parseable form. Terms forbid reproduction, distribution, storage in a retrieval system, and derivative works. |
| **All About Birds / Birds of the World** | Carries the data; blocks automated retrieval (403) and is subscription-gated respectively. Not assessed further. |
| **Macaulay Library** | Permits *embedding* for non-commercial use — which BirdView qualifies for — but offers no discovery API, and its media-request path explicitly prunes species lists over 100. Cannot scale to the dataset. Viable later for hand-curated species. |
| **Wikimedia Commons, iNaturalist, Flickr** | Photo sources, all researched in detail. Out of scope while P-050 is withdrawn; Commons is the recommendation when it returns. |

### Why AviList does not replace the ABA checklist

*Pass run 2026-10-01, after the ABA permission request went unanswered.* The question was whether
the checklist could be dropped entirely, since redistributing it is what blocks public deployment.

**What ABA actually supplies is three separable things**, and they have different answers:

| | Replaceable? |
| --- | --- |
| **Scope** — which 841 species, by rarity codes 1–3 (P-001) | **no open source found** |
| **Common names**, species and family (P-004) | yes, but at a cost |
| **Taxonomic sequence** for the order grouping (P-010) | yes |

**Scope is the one that matters and nothing supplies it.** AviList carries a `Range` column, but
it is free-text prose — *"desert to savanna plains of northeastern Africa, from northern
Ethiopia…"* — and cannot be filtered on. AOS has real status flags but for the wrong region, as
measured above. The ABA codes are editorial judgment about one specific area, and that judgment
is the product. No taxonomy supplies it because it is not taxonomy.

**Names are replaceable, and should not be replaced.** 832 of 841 match AviList on scientific
name; the 9 that do not are genus reassignments (*Coccothraustes vespertinus*, four *Dryobates*,
*Icterus bullockii*, *Pica nuttalli*, *Porphyrio poliocephalus*, *Vireo swainsoni*). But **86
common names change and 55 of 91 family English names change**, and they change *away* from this
audience: Dovekie becomes Little Auk, Rough-legged Hawk becomes Rough-legged Buzzard, Brant
becomes Brant Goose, every Scrub-Jay loses its hyphen. AviList is a global consensus list and
names birds for the world; BirdView's reader is in the ABA Area and calls them what the ABA
calls them. Adopting it would read as a clean licensing win in this table and be a regression on
the page. (The extended file carries an `English_name_Clements_v2025` column, which is closer to
North American usage and is distributed under AviList's own CC BY 4.0 — worth a look if names
ever do need to move.)

**The taxonomy would also shift.** Four species change order — the three New World vultures move
from Accipitriformes to a Cathartiformes of their own, which adds an order group to the filter —
and two change family: the Wrentit to Paradoxornithidae, the Yellow-breasted Chat from
Icteriidae to Icteridae, dissolving a one-species family.

**Conclusion: the ABA ask cannot be routed around.** Either the permission arrives, or BirdView
defines its own scope and owns it — the spread statistic noted above is the only candidate, and
it trades a curated list for a model. Nothing here changes what is deployed; it is recorded so
the question is not reopened from scratch.

### Why occurrence data cannot reproduce the ABA scope

*Measured 2026-10-01, against the eBird Observation Dataset on GBIF — CC BY 4.0, 1.07 billion US
records.* The spread statistic was the last candidate for replacing the ABA codes, so it was
tried rather than argued about.

**Method.** Facet species by region across **64 regions** — 51 US states and territories plus 13
Canadian provinces, which is the ABA Area to within St. Pierre et Miquelon. The whole matrix
takes **39 seconds**, one query per region. Regions are not the hard part; GBIF has them as a
field.

**Result: no threshold gets close enough to call it a replacement.** Against the ABA's 841:

| Rule | Species | Agree | ABA has, rule misses | Rule adds | F1 |
| --- | --- | --- | --- | --- | --- |
| present in ≥2 regions | 1063 | 724 | 117 | 339 | 0.76 |
| present in ≥3 regions | 931 | 703 | 138 | 228 | 0.79 |
| present in ≥5 regions | 783 | 643 | 198 | 140 | 0.79 |
| ≥100 records in ≥1 region | 997 | 771 | 70 | 226 | **0.84** |
| ≥100 records in ≥2 regions | 772 | 669 | 172 | 103 | 0.83 |
| ≥100 records in ≥3 regions | 692 | 626 | 215 | 66 | 0.82 |

**The failures are structural, not a tuning problem**, which is why no row wins:

- **It adds escaped cage birds.** The additions are Hyacinth Macaw, six *Amazona* parrots,
  lovebirds, Mandarin Duck, Red-legged Partridge. eBird faithfully records what people see; the
  ABA list is a judgment about which populations are *established*. No occurrence statistic
  separates a self-sustaining population from a steady trickle of escapees, because the
  observations look identical.
- **It drops single-region endemics.** Hawaii is one region, so any multi-region rule deletes
  the Hawaiian avifauna — Millerbird, Laysan Duck, Hawaiian Duck, Mariana Swiftlet — and Hawaii
  is explicitly inside the ABA Area. Range-restricted mainland birds go the same way: Five-striped
  Sparrow, Whiskered Auklet.
- **Some misses are nomenclature, not biology.** *Anarhynchus montanus*, *nivosus* and *wilsonia*
  and *Amphispizopsis quinquestriata* fail to match because GBIF's backbone has not followed the
  recent genus moves — the same rot CR-018 handles for links, reappearing as a data-quality
  problem.

Each of these can be patched — special-case Hawaii, filter on establishment, pin a taxonomy —
and every patch is one more editorial judgment BirdView would be making and defending. At that
point the project has written its own checklist, which is a different undertaking from drawing
birds to scale.

**So the choice is real and narrow:** obtain the ABA's permission, or publish a list that is
about 84% of theirs, visibly different at the edges, and owned by BirdView. There is no third
option where the data simply comes from somewhere else.
