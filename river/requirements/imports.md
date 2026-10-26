# Imports — BirdView

Every body of rules or data BirdView is subject to or has evaluated: its origin, the version it is
pinned to, its tier, and how far adoption has got. Rejected and retired sources stay listed with
the reason, so nobody evaluates them again.

Search for a source's heading, such as ``## `avonet` ``; `grep -n '^## '` lists them all. Each opens
with a field table. `Status` is the field that distinguishes an adopted source from a named one —
all the others can be filled in plausibly from a URL alone. `Integrity` names **where** a content
hash lives. It never contains one.

## Summary

| Name | Tier | Import | Status |
| --- | --- | --- | --- |
| [`avonet`](#avonet) | mandatory | by reference, vendored | **adopted** — pinned, vendored, hashed; no criteria yet |
| [`aba`](#aba) | mandatory | by reference, **not** vendored | **partial** — terms read and quoted, pinned by hash; redistribution refused, so fetched at build |
| [`fox-wingspan`](#fox-wingspan) | mandatory | by reference, vendored | **adopted** — retrieved, hashed against the publisher's MD5, vendored |
| [`bohmer-neck`](#bohmer-neck) | mandatory | consulted, not ingested | **adopted** — CC BY 4.0 verified; calibrates one constant, contributes no data |
| [`avilist`](#avilist) | mandatory | by reference, vendored | **adopted** — CC BY 4.0, vendored and hashed; supplies every name and the species code |
| [`gbif`](#gbif) | mandatory | by reference, queried, not vendored | **adopted** — CC BY 4.0 verified from the API; supplies the GBIF→AviList name bridge only |
| [`ricklefs`](#ricklefs) | mandatory | **rejected** | **rejected** — Wiley terms prohibit the use, see [decisions.ricklefsProhibited](decisions.md#ricklefsprohibited) |
| [`wiley`](#wiley) | mandatory | quoted only, not vendored | **adopted** — terms read and transcribed; the reason `ricklefs` is rejected |
| [`wcag`](#wcag) | mandatory | by reference, vendored | **partial by design** — three contrast criteria transcribed; no conformance level claimed |
| [`inat-api`](#inat-api) | governing | by reference, vendored | **adopted** — copied from `birdsync` and hash-verified; governs the mapping tool |
| [`inat-terms`](#inat-terms) | mandatory | by reference, vendored | **adopted** — copied from `birdsync` and hash-verified; almost nothing in it reaches BirdView |
| [`specman`](#specman) | — (method) | project copy | **retired** 2026-10-15 — replaced by River, whose pin is in `process.md`'s banner |
| [`ebird-sensitive`](#ebird-sensitive) | mandatory | absorbed as facts, not vendored | **adopted** — the US and Canada entries restated in `taxonomy.EBirdSensitive`, each confirmed against the export ([decisions.ebirdSilenceNotAbsence](decisions.md#ebirdsilencenotabsence)) |
| [`natural-earth`](#natural-earth) | mandatory | by reference, pinned by hash, **not** vendored | **adopted** 2026-10-18 — public domain; fetched and hash-checked by `cmd/birdview-boundaries` for the Here button's region boundaries |
| [`birdlife-checklist`](#birdlife-checklist) | mandatory | **rejected** | **rejected** — DataZone terms forbid automated access; [decisions.vegaGullBorrows](decisions.md#vegagullborrows) re-based 2026-10-14 onto the vendored `avilist` and `avonet` workbooks |

Two of the fifteen are rejected, and neither blocks v1. `birdlife-checklist` was consulted in a
way DataZone's terms forbid; its quotations were removed and [decisions.vegaGullBorrows](decisions.md#vegagullborrows) was
re-based on the vendored CC BY 4.0 workbooks. `ricklefs` was evaluated and
**rejected**: its platform terms prohibit extraction, integration, storage, and redistribution
alike. Both entries are kept so the evaluations are not repeated.

---

## `avonet`

| Field | Value |
| --- | --- |
| **Origin** | https://doi.org/10.6084/m9.figshare.16586228 |
| **Version** | v7 |
| **Integrity** | [river/requirements/imported/figshare.com/avonet/PROVENANCE.md](imported/figshare.com/avonet/PROVENANCE.md) |
| **Tier** | mandatory (copyright license) |
| **Import style** | By reference; vendored at `//third_party/avonet/AVONET Supplementary dataset 1.xlsx` under the location exception in [bindings.md](../bindings.md#location-exception-the-third_party-directory) |
| **Scope** | All morphometric traits ([product.traitsFromAvonet](local/product/dataset.md#traitsfromavonet), [product.traitsPerSpecies](local/product/dataset.md#traitsperspecies)) |
| **Geographic scope** | **Worldwide** — 10,661 species across 41 orders, every continent. Measured 2026-10-01. |
| **Adopted parts** | Sheets `AVONET2_eBird`, `AVONET1_BirdLife`, and `AVONET3_BirdTree` ingested ([decisions.abaSpeciesWithoutAvonet](decisions.md#abaspecieswithoutavonet), [decisions.whiteTernBorrows](decisions.md#whiteternborrows), [decisions.herringGullEuropeanRow](decisions.md#herringgulleuropeanrow)); `AVONET_Raw_Data` and the crosswalk sheets consulted |
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
| **Integrity** | [river/requirements/imported/aba.org/aba/PROVENANCE.md](imported/aba.org/aba/PROVENANCE.md) |
| **Tier** | mandatory (copyright) |
| **Import style** | **Retired 2026-10-05.** Not read at build, not vendored, not referenced by any code. |
| **Scope** | *(formerly)* species list, common and scientific names, families, rarity filter |
| **Adopted parts** | None. |
| **Exclusions** | Everything. |
| **Owner** | American Birding Association |
| **Status** | **Retired 2026-10-05**; see below the table |

**Status.** **Retired 2026-10-05, and this is the entry that unblocked deployment.** Names,
families, and sequence moved to AviList
([decisions.avilistReplacesAba](decisions.md#avilistreplacesaba));
scope moved to BirdView's own derivation
([decisions.scopeFromRecurrence](decisions.md#scopefromrecurrence)). `internal/aba` is deleted and
nothing fetches the file. The entry stays because the record of why a source was adopted and then
dropped is worth more than the space it takes — and because anyone re-reading the terms below should
see at the top that they no longer bind this project.

License, quoted verbatim from row 0 of the CSV itself:

> "The ABA Checklist is a copyrighted work owned by American Birding Association, Inc. and cannot
> be reproduced without the express written permission of American Birding Association, Inc."

No other terms document governs it: `aba.org/terms-of-use/`, `/copyright/`, and `/legal/` all 404,
and the checklist landing page carries no license statement at all.

**Open action:** ask ABA for written permission. It is the only input v1 genuinely depends on that
cannot be committed, and a non-commercial, attributed, open-source request costs an email.

---

## `inat-api`

| Field | Value |
| --- | --- |
| **Origin** | <https://www.inaturalist.org/pages/api+recommended+practices> |
| **Version** | Page's own revision line: "Revised on February 27, 2025 06:23 PM by kueda" |
| **Integrity** | [river/requirements/imported/inaturalist.org/inatapi/PROVENANCE.md](imported/inaturalist.org/inatapi/PROVENANCE.md) |
| **Tier** | governing (the operator can throttle or block a client that ignores it) |
| **Import style** | By reference, vendored — **copied from the maintainer's `birdsync` repository and verified byte-for-byte against the SHA-256 recorded there** |
| **Scope** | Every request `cmd/birdview-inatmap` makes to `api.inaturalist.org` |
| **Adopted parts** | all eight, in [inatapi.md](imported/inaturalist.org/inatapi/inatapi.md) |
| **Exclusions** | Bulk access routes — observation exports, the weekly GBIF dataset — which serve consumers of observation data. BirdView reads one integer per species from the taxonomy. |
| **Owner** | iNaturalist |
| **Status** | **Adopted 2026-09-23.** Vendored and hash-verified; transcribed against BirdView's much narrower use. |

Two of the eight are implemented ([`inatapi.RequestRate`](imported/inaturalist.org/inatapi/inatapi.md#requestrate) pacing as [tech.pacedInaturalistMapping](local/tech/process.md#pacedinaturalistmapping), [`inatapi.UserAgent`](imported/inaturalist.org/inatapi/inatapi.md#useragent) as the pre-existing
[tech.identifiedClient](local/tech/delivery.md#identifiedclient)); two are not applicable because BirdView never pages; four are satisfied by construction
because it never authenticates, never posts, and downloads no media. [`inatapi.AuthenticateOnlyWhenNecessary`](imported/inaturalist.org/inatapi/inatapi.md#authenticateonlywhennecessary) is worth noting:
it is the one requirement `birdsync` records as a knowing departure, and BirdView satisfies it
only because its needs are so much smaller.

The human-facing page cannot be retrieved by an agent — `www.inaturalist.org/pages/*` answers with
HTTP 403 and a JavaScript challenge. The API itself serves an identified client normally.

---

## `inat-terms`

| Field | Value |
| --- | --- |
| **Origin** | <https://www.inaturalist.org/pages/terms>, <https://www.inaturalist.org/pages/community+guidelines>, and the linked <https://www.inaturalist.org/pages/machine_generated_content> |
| **Version** | Retrieved 2026-08-09 from a browser; hashes in [PROVENANCE.md](imported/inaturalist.org/inatterms/PROVENANCE.md) |
| **Integrity** | [river/requirements/imported/inaturalist.org/inatterms/PROVENANCE.md](imported/inaturalist.org/inatterms/PROVENANCE.md) |
| **Tier** | mandatory (copyright and platform terms) |
| **Import style** | By reference, vendored — **copied from `birdsync` and hash-verified**, as above |
| **Scope** | BirdView's use of the platform: unauthenticated reads, and outbound links to taxon pages |
| **Adopted parts** | all six, in [inatterms.md](imported/inaturalist.org/inatterms/inatterms.md) |
| **Exclusions** | Conduct rules unreachable by a client that posts nothing: harassment, sockpuppets, explicit content, commercial AI training |
| **Owner** | iNaturalist |
| **Status** | **Adopted 2026-09-23.** Vendored so that one negative finding is checkable rather than remembered. |

BirdView is subject to very little of this, and the entry exists to say which little. **The
adoption rests on a negative: nothing in either document restricts reusing a taxon identifier or
linking to a taxon page.** `terms.html` contains no occurrence of *taxon*, *taxonomy*, *database*,
*derivative*, *hyperlink*, or *scrape*. The CC BY-NC clause governs Content a contributor posts —
observations and photographs — not iNaturalist's taxonomy.

That is exactly the kind of claim that is worthless asserted from memory, which is why hashed
copies are in the tree: anyone can search them and check.

Photographs stay on iNaturalist. BirdView links; it never copies, caches, hotlinks, or embeds an
image, which [tech.selfContainedServer](local/tech/delivery.md#selfcontainedserver) independently forbids anyway. See
[decisions.inaturalistLinkAdopted](decisions.md#inaturalistlinkadopted) for the adoption reasoning and what was rejected.

---

## `avilist`

| Field | Value |
| --- | --- |
| **Origin** | https://doi.org/10.2173/avilist.v2025b |
| **Version** | v2025b, published 2026-06-11, **extended** release |
| **Integrity** | [river/requirements/imported/avilist.org/avilist/PROVENANCE.md](imported/avilist.org/avilist/PROVENANCE.md) |
| **Tier** | mandatory (copyright license) |
| **Import style** | By reference, vendored — CC BY 4.0 permits it, and that permission is the point |
| **Scope** | Every English name, family and order name, and the eBird species code ([product.namesFromAvilist](local/product/dataset.md#namesfromavilist), [product.englishNameChoice](local/product/diagram.md#englishnamechoice)) |
| **Geographic scope** | Worldwide — 11,131 species, 252 families |
| **Adopted parts** | `Taxon_rank`, `Scientific_name`, `English_name_Clements_v2025`, `English_name_AviList`, `Species_code_Cornell_Lab`, `Family`, `Family_English_name`, `Order`. The **extended** release is required: the short one omits the Clements name and the species code. |
| **Exclusions** | `Range`, which looks like a scope source and is not — empty for **4,992 of 11,131** species, including Ruddy Duck and Tundra Swan, because it is populated only on an exact Clements concept match. |
| **Owner** | AviList Core Team |
| **Status** | **Adopted 2026-10-02**, replacing the ABA checklist for names ([decisions.avilistReplacesAba](decisions.md#avilistreplacesaba)). Retrieved, hashed, vendored. |

**This is the source that unblocks deployment.** The ABA checklist is all-rights-reserved and
serving BirdView reproduces it, so going public was blocked on a permission that has not arrived.
AviList supplies the same four things under CC BY 4.0 and may therefore be served.

It also answers the objection that defeated it a day earlier. AviList's own English names are IOC
usage — Little Auk for Dovekie — which is wrong for this audience. But the extended release
**also carries the Clements name**, which *is* this audience's usage: measured over BirdView's
scope, the Clements column reproduces the ABA name for **827 of 827** species where both exist,
the only five differences being ABA's own `X (alternate)` notation. One file, both vocabularies,
which is what makes [product.englishNameChoice](local/product/diagram.md#englishnamechoice) possible rather than aspirational.

### Why the Clements and Cornell columns are usable

"Considered and not adopted" below rejects eBird/Clements because Cornell's terms forbid
redistribution without written authorization. **That rejection stands, and this is not it.**

AviList is a separate publication by a consortium — the AOS, Cornell, BirdLife, Avibase, and the
IOU — with Cornell staff on its core team, released CC BY 4.0 in its entirety. BirdView takes
these columns from AviList, under AviList's license, with AviList's attribution. Nothing is taken
from Cornell's distribution, and nothing here grants access to it.

Recorded explicitly because the shape of it — *the source we rejected, reached through a source we
accepted* — is exactly what should attract scrutiny rather than escape it.

---

## `gbif`

| Field | Value |
| --- | --- |
| **Origin** | <https://api.gbif.org/v1> — the GBIF Backbone Taxonomy, https://doi.org/10.15468/39omei, and the EOD – eBird Observation Dataset, https://doi.org/10.15468/aomfnb |
| **Version** | Queried 2026-10-02. Backbone dataset record: published 2023-08-28, modified 2023-11-17. EOD record: published 2025-08-08, modified 2025-10-29. The API serves whatever is current; the pin is the retrieval date and the generated table. |
| **Integrity** | **Not vendored, so no hash here.** What is committed is `//data/gbif-taxa.tsv`, which is the project's own reduction and not a copy of anything GBIF publishes. |
| **Tier** | mandatory (copyright license) |
| **Import style** | By reference, **queried not vendored** — `cmd/birdview-gbifmap` reads the API and writes the bridge table, which the repository commits |
| **Scope** | Names only: mapping a GBIF backbone scientific name onto the AviList name for the same taxon |
| **Geographic scope** | The backbone is worldwide. The species set the bridge covers is scoped to `country=US` and `country=CA` — 1,378 species over 1,238,404,492 records. |
| **Adopted parts** | `/occurrence/search` faceted on `speciesKey` (the bridge) and on `verbatimScientificName` per GADM region (region membership, [product.regionAndMonthFilter](local/product/filtering.md#regionandmonthfilter)); `/species/{key}` for the accepted name and family; `/species/{key}/synonyms` for the names it absorbed. The **iNaturalist research-grade dataset**, `50c9509d-22c7-4a22-a47d-8c48425ef4a7`, is queried the same way and for one reason: eBird withholds sensitive species. |
| **Exclusions** | Occurrence records themselves — no coordinate, date, observer, or media reaches the dataset. `/species/match` is excluded as unfit, below. **The occurrence counts are still not used for scope**, which remains an open product decision; they *are* used for region membership ([product.regionAndMonthFilter](local/product/filtering.md#regionandmonthfilter)), where only the derived list of regions per species is stored, never a count. |
| **Owner** | GBIF Secretariat |
| **Status** | **Adopted 2026-10-02** for the name bridge. License verified from the API's own dataset records; coverage measured and recorded below; no criteria in `acceptance.md` yet. |

License: **CC BY 4.0** on both datasets, read from `GET /v1/dataset/{key}` rather than from a web
page, because `www.gbif.org` answers an agent with a Cloudflare challenge. The committed table
names both DOIs in its header, which is where the attribution lives while nothing from it is
served.

**No API terms are transcribed, and the gap is deliberate.** `www.gbif.org/terms` returns HTTP 403
to an identified client, and [README.md](../README.md#sources) forbids supplying from memory a
rule that could not be retrieved — a plausible rate limit is indistinguishable from a real one.
What *was* retrievable is `api.gbif.org/robots.txt`, which disallows one image path and nothing
else. So the pacing in `cmd/birdview-gbifmap` — 100 ms between requests across four workers, about
five minutes for a full run — is a courtesy BirdView sets for itself, not a published limit it is
meeting.

### Why this is not the eBird rejection, and not the GBIF scope rejection either

Two entries in [Considered and not adopted](#considered-and-not-adopted) look like this one and
are not:

- **eBird / Clements** is rejected because Cornell's terms forbid redistribution. Nothing here
  comes from Cornell's distribution. The observations reach BirdView through GBIF, openly
  licensed, and only as a list of names.
- **GBIF — eBird Observation Dataset** is rejected *as a scope source*, and that rejection stands.
  This adoption takes the taxonomy and leaves the scope question exactly where it was.

### What the bridge is for, and what it measures

GBIF's backbone lags current taxonomy by several revisions, so the two name sets do not join:
**62 of BirdView's 839 species have no verbatim match among GBIF's accepted names.** That is a
join failure, not an absence — GBIF files Cooper's Hawk under `Accipiter cooperii` where AviList
has `Astur cooperii`.

`/species/match` does not fix it and must not be reached for. It maps an *old* name onto the
current backbone, which is the opposite direction. Asked for `Astur cooperii` on 2026-10-02 it
answered `{"matchType":"NONE"}` with `strict=true`, and without it `Aster cooperi (A.Gray)
Kuntze` — a shrub in the Asteraceae — at `matchType: FUZZY`, confidence 84. Both answers are
useless and the second is worse than useless.

The backbone's own synonymy does fix it. Measured over the 1,378 species:

| Tier | Resolved | What it is |
| --- | ---: | --- |
| exact | 1,301 | GBIF's accepted name is an AviList species name |
| synonym | 71 | one name GBIF files as a synonym of it is |
| epithet | 3 | one AviList species in the same family shares the epithet, after folding gender agreement and Latin genitive endings |
| ambiguous | 1 | more than one corresponded, so none was chosen |
| none | 2 | nothing corresponded |

**1,375 of 1,378 resolve**, and BirdView's own coverage goes from 777 to 813 of 839. The 26 that
remain are not name failures: GBIF's species facet holds no taxon for them at all.

### Where a resolved row is name-correct and concept-wrong

**A resolved row says the two authorities use different names for one taxon. It does not say their
species concepts agree.** Where GBIF has not followed a split, its name maps to whichever half
kept the name, and the occurrences behind it are mostly the other half.

Nine of these touch BirdView's own species, found by asking GBIF what it makes of each AviList
name the bridge failed to reach, on 2026-10-02. **Two mechanisms, and the distinction matters:**

**Seven, where GBIF's backbone holds the AviList name as a synonym of the lump** — it has not
followed the split at all, and `/species/match` says so:

| GBIF taxon | records | the bridge maps it to | the bird those records mostly are |
| --- | ---: | --- | --- |
| `Circus cyaneus` | 3,815,109 | `Circus cyaneus` (Hen Harrier) | `Circus hudsonius`, Northern Harrier |
| `Gallinula chloropus` | 1,474,223 | `Gallinula chloropus` (Eurasian Moorhen) | `Gallinula galeata`, Common Gallinule |
| `Cistothorus platensis` | 333,113 | `Cistothorus platensis` (Grass Wren) | `Cistothorus stellaris`, Sedge Wren |
| `Buteo nitidus` | 141,856 | `Buteo nitidus` (Gray-lined Hawk) | `Buteo plagiatus`, Gray Hawk |
| `Oceanodroma leucorhoa` | 66,497 | `Hydrobates leucorhous` (Leach's Storm-Petrel) | mostly right; subsumes `Hydrobates socorroensis` |
| `Calonectris diomedea` | 26,230 | `Calonectris diomedea` (Scopoli's Shearwater) | mostly `Calonectris borealis`, Cory's Shearwater |
| `Gygis alba` | 16,008 | `Gygis alba` (White Tern) | `Gygis candida`, Blue-billed White-Tern |

**Two, where GBIF's backbone accepts both halves and the dataset still reports only the parent.**
The backbone is not the problem here; what eBird published to GBIF is:

| GBIF taxon | records | the bridge maps it to | the bird those records mostly are | backbone status of the other half |
| --- | ---: | --- | --- | --- |
| `Setophaga petechia` | 6,245,227 | `Setophaga petechia` (Mangrove Warbler) | `Setophaga aestiva`, Northern Yellow Warbler | accepted, zero records |
| `Numenius phaeopus` | 556,569 | `Numenius phaeopus` (Whimbrel) | `Numenius hudsonicus`, Hudsonian Whimbrel | accepted, zero records |

**Six of the nine are the dangerous shape**, and they span both tables: the name the bridge lands
on belongs to a bird BirdView does not carry — a vagrant in the ABA Area, or absent from it —
while the bird actually seen gets no row at all. Those are `Circus cyaneus`, `Gallinula
chloropus`, `Numenius phaeopus`, `Cistothorus platensis`, `Buteo nitidus`, and `Gygis alba`,
carrying 6,336,878 records between them.

In the other three — `Setophaga petechia`, `Oceanodroma leucorhoa`, `Calonectris diomedea` — the
name does land on a species BirdView carries, so the row is usable and merely absorbs a second
bird's records: 6,337,954 more. All nine together are 12,674,832 records, 1.02% of the corpus and
concentrated in common species.

Nine is a floor, not a census. The method finds only lumps that collide with a species BirdView
already carries; a lump among the 561 GBIF species BirdView does not have would not show up.
Nothing detects them automatically — it is one query per unreached species, by hand.

Anything that later derives *scope* from these counts has to face this. The bridge only records
it, because the table has one AviList column and a lump needs two.

### Two findings that govern any future use of these counts for scope

Measured 2026-10-02, after the bridge, and recorded here because they decide the shape of the
scope work rather than the bridge's.

**1. The species facet must not be used to enumerate species.** Faceting `/occurrence/search` on
`speciesKey` silently omits every record GBIF's backbone could not resolve — about **4.7 million
records, 0.4% of the corpus, filed above species rank**. The casualty list is not marginal: the
facet reports **zero** US+CA records for Cooper's Hawk.

Querying by `verbatimScientificName`, which is the name **eBird published** rather than the name
GBIF matched it to, recovers them. Cooper's Hawk returns **4,148,628**. Over BirdView's 839
species the inverted query finds **827 (98.6%)** against the facet's 777.

So the direction of the question matters: *ask per species by name*, never *ask the taxonomy what
exists*. The second form fails silently and the failure looks like absence.

*(A first run of this measurement reported ~40 species at zero, including European Starling. That
was API throttling at concurrency 8 with failures counted as zeros. At concurrency 3 with backoff
there were no failures. Noted because a rate-limited query returning 0 is indistinguishable from
an absent bird unless the two paths are kept apart.)*

**2. eBird's export withholds conservation-sensitive species entirely.** The species the
inverted query cannot find are not a taxonomy problem. Every one is present in iNaturalist and
all but absent from eBird:

*(Corrected 2026-10-05 by [product.sourceNamingMeasured](local/product/dataset.md#sourcenamingmeasured)'s ratio band. This said **12 species** and
listed nine, and one of the nine was wrong: **Blue-billed White-Tern is not withheld** — eBird
holds 16,008 records of it under *Gygis alba*, a lump, not a suppression. The band also found
two the table missed, because neither reads as zero: **Great Gray Owl**, eBird 1 against
iNaturalist's 2,859, and **Black Rail**, 16 against 101. Eleven species are withheld, and
`taxonomy.WithheldByEBird` is the list.)*

| species | eBird | iNaturalist |
| --- | ---: | ---: |
| Spotted Owl | **0** | 1,061 |
| Northern Hawk Owl | **0** | 1,644 |
| Gyrfalcon | **0** | 511 |
| Lesser Prairie-Chicken | **0** | 228 |
| Gunnison Sage-Grouse | **0** | 110 |
| Akohekohe, Maui Parrotbill, Puaiohi | **0** | 7, 10, 21 |
| *Great Gray Owl* | *1* | *2,859* |
| *Black Rail* | *16* | *101* |
| *Barred Owl — control, not sensitive* | *1,242,780* | *55,832* |

The control is what makes it legible: Barred Owl is in both, and its sensitive congener Spotted
Owl is in neither GBIF index while iNaturalist holds a thousand records. The pattern reads as
eBird's sensitive-species policy — raptors and grouse vulnerable at known sites, plus critically
endangered Hawaiian endemics. *(Inference from the pattern and the control. eBird's published
sensitive-species list was not consulted, and `www.gbif.org/terms` is unreachable to an
identified client, so this is not confirmed from a source.)*

It is **systematic and adverse**: an eBird-only list omits precisely the rare and protected birds
a reader is most likely to look up, with no signal that anything is missing.

**eBird ∪ iNaturalist covers 839 of 839.** Which revises this manifest's own judgment below that
iNaturalist is superseded — see that entry.

---

## `fox-wingspan`

| Field | Value |
| --- | --- |
| **Origin** | https://doi.org/10.6084/m9.figshare.32089083 |
| **Version** | v1, published 2026-04-24 |
| **Integrity** | [river/requirements/imported/figshare.com/foxwingspan/PROVENANCE.md](imported/figshare.com/foxwingspan/PROVENANCE.md) |
| **Tier** | mandatory (copyright license) |
| **Import style** | By reference, vendored — CC BY 4.0 permits it |
| **Scope** | Empirical wingspan ([product.lengthAndWingspanSources](local/product/measurements.md#lengthandwingspansources)) |
| **Geographic scope** | **Worldwide** — 10,618 species. Its gaps are functional, not regional: flightless birds (Ostrich, Emu, penguins, kiwi) and Galliformes are absent, while flying birds from every continent are present. Measured 2026-10-01. |
| **Adopted parts** | Rows with `needs_prediction` false only. The column is a spreadsheet boolean: `FALSE`/`TRUE` as `internal/xlsx` renders it, `0`/`1` in the underlying XML, and the reader accepts both. Matching only `"0"` yields **zero** measured wingspans and looks like a coverage problem rather than a parsing one — see `isEmpirical`. |
| **Exclusions** | Predicted wingspans; see below the table |
| **Owner** | Fox, Brunton & Shephard |
| **Status** | **Adopted 2026-09-21.** Retrieved, hashed, vendored; its MD5 matches figshare's published `computed_md5`. Coverage for BirdView's scope measured and recorded. |

**Exclusions.** **Predicted wingspans are not ingested as values.** The `wingspan_predicted_*`
columns are regressions from AVONET wing chord; taking them as data would reintroduce the
modeled-as-measured problem [decisions.noOpenLengthAndWingspan](decisions.md#noopenlengthandwingspan) resolves. They are used for one other
purpose: as the reference BirdView's own fitted model is validated against
([decisions.fittedWingspanModel](decisions.md#fittedwingspanmodel)). Reading them to check our
arithmetic is not the same as presenting them as measurements.

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
| **Scope** | Calibration evidence only: the mass exponent of [product.trunkFromMass](local/product/diagram.md#trunkfrommass) |
| **Geographic scope** | Worldwide — 103 species across 68 families and 34 orders, MNHN Paris skeletal collection, one specimen per species |
| **Adopted parts** | Table S4's published scaling, used to choose one constant. Nothing is joined to BirdView's species. |
| **Exclusions** | Everything else. In particular the per-species neck lengths are **not** used as BirdView's neck values — Böhmer measures the full cervical column, including the part inside the body contour, where [product.neckModelPerFamily](local/product/diagram.md#neckmodelperfamily) models the *exposed* neck of a silhouette. The two are different quantities and must not be equated; see [decisions.trunkExponentCalibrated](decisions.md#trunkexponentcalibrated). |
| **Owner** | Böhmer, Rössner, Fabre et al. |
| **Status** | **Adopted 2026-10-02 as calibration evidence.** CC BY 4.0 verified from the figshare API record itself, not inferred from the article. |

Why it is a source at all when no number from it is stored: [decisions.trunkExponentCalibrated](decisions.md#trunkexponentcalibrated) changes a
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

Units in the supplement are unstated. They are kilograms and meters, recovered two independent
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
| **Scope** | Passerine total length ([product.lengthAndWingspanSources](local/product/measurements.md#lengthandwingspansources)) |
| **Adopted parts** | `Length` column, passerines |
| **Exclusions** | — |
| **Owner** | Ricklefs / Wiley / ESA |
| **Status** | **Rejected 2026-09-21.** The terms were read and prohibit extraction, integration with other data, storage in a retrieval system, and redistribution — explicitly regardless of non-commercial use. See [`wiley`](#wiley) and [decisions.ricklefsProhibited](decisions.md#ricklefsprohibited). |

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
| **Integrity** | [river/requirements/imported/onlinelibrary.wiley.com/wiley/PROVENANCE.md](imported/onlinelibrary.wiley.com/wiley/PROVENANCE.md) — records the hash of the maintainer's local printout, **not** a vendored file |
| **Tier** | mandatory |
| **Import style** | **Quoted only.** The document is deliberately not vendored — see below |
| **Scope** | Any use of material from Wiley Online Library; in practice, Ricklefs 2017 |
| **Adopted parts** | Rules of Use A.2, B, C, F, G, H |
| **Exclusions** | Subscription, PPV, and e-commerce terms; warranty limitations; privacy policy |
| **Owner** | John Wiley & Sons, Inc. |
| **Status** | **Adopted.** Retrieved by a human, transcribed with quotations, and acted on — it is why `ricklefs` is rejected. |

**This source is not vendored, and the departure is deliberate.** Rules of Use B forbids storing
Wiley material "in any form or medium in a retrieval system" and F forbids incorporating it "into
any institutional or other repository under any circumstances". A git repository is both.
Committing the terms in order to evidence compliance with the terms would breach them. Rules of
Use C permits "brief quotations for purposes of comment, criticism or similar scholarly
purposes", so the transcription quotes the governing clauses and the document stays out.

Recorded here as a location and import-style exception, in the same spirit as the AVONET workbook
exception in [bindings.md](../bindings.md#location-exception-the-third_party-directory).

---

## `wcag`

| Field | Value |
| --- | --- |
| **Origin** | https://www.w3.org/TR/2024/REC-WCAG22-20241212/ |
| **Version** | WCAG 2.2, W3C Recommendation of 12 December 2024 |
| **Integrity** | [river/requirements/imported/w3.org/wcag/PROVENANCE.md](imported/w3.org/wcag/PROVENANCE.md) |
| **Tier** | **mandatory** |
| **Import style** | By reference, vendored at `//third_party/wcag/` under the location exception in [bindings.md](../bindings.md#location-exception-the-third_party-directory); transcribed in [imported/w3.org/wcag/wcag.md](imported/w3.org/wcag/wcag.md) |
| **Scope** | Every BirdView web page, in full |
| **Adopted parts** | SC 1.4.6 Contrast (Enhanced) for all text, which meets SC 1.4.3 Contrast (Minimum) as well; SC 1.4.11 Non-text Contrast for graphics and interface components; conformance requirement 5.2.2, Full pages, as the reading of each; and the definitions of *web page* and *contrast ratio* they use |
| **Exclusions** | **Everything else.** No keyboard-access, focus-order, timing, reading-level, or other criteria are adopted, and BirdView claims no conformance level. The scope widened once, from badges to all text, after an unstyled link shipped at 1.9:1 — see [decisions.twoContrastFloors](decisions.md#twocontrastfloors) — and the separate 4.5:1 floor for ordinary text was then dropped for one 7:1 floor ([decisions.oneContrastFloor](decisions.md#onecontrastfloor)). |
| **Owner** | W3C; project owner grants any deviation |
| **Status** | **Partial by design.** Vendored, hashed, and transcribed 2026-10-16, with every quotation checked against the vendored copy. 1.4.11 has no check yet. |

Scoping this narrowly is deliberate. The earlier design claimed "a minimum WCAG AA contrast ratio
of ≥ 7:1", which conflates two conformance levels and asserts an accessibility posture nothing
verified. A narrow adopted scope that is actually checked is worth more than a broad claim that
is not.

---

## `specman`

| Field | Value |
| --- | --- |
| **Origin** | https://github.com/Sajmani/specman — `process.md`; the repository has been private since 2026-10-21 |
| **Version** | Base pin `63bfdecb99a256d2819aa6d0e55bfb0bccf19b96db40e28f72572315e81cbdbc`, upstream `2c41af8`, taken 2026-09-23 |
| **Integrity** | Not applicable — see below |
| **Tier** | — it states how requirements are made, and produces none for criteria to cite |
| **Import style** | Project copy at `spec/process.md`, carrying a banner that points here |
| **Scope** | How changes to this repository are worked |
| **Owner** | Project owner |
| **Status** | **Retired 2026-10-15.** Replaced by River, specman's successor. |

The `Version` field above is a **base pin, not an integrity hash**, and is kept as history: it
records which method this project was worked under from 2026-09-23 to 2026-10-15.

**River gets no entry here.** River's `README.md` holds that the method is not a requirements
source — it produces nothing for a criterion to cite, which this entry already observed — and
records the pin in the file's own banner, so that one pin has one record. Local amendments to
it are in [bindings.md](../bindings.md).

---

## `natural-earth`

| Field | Value |
| --- | --- |
| **Origin** | https://www.naturalearthdata.com/downloads/10m-cultural-vectors/10m-admin-1-states-provinces/ — archive at `https://naciscdn.org/naturalearth/10m/cultural/ne_10m_admin_1_states_provinces.zip` |
| **Version** | 5.1.1 |
| **Integrity** | [river/requirements/imported/naturalearthdata.com/naturalearth/PROVENANCE.md](imported/naturalearthdata.com/naturalearth/PROVENANCE.md) |
| **Tier** | mandatory (copyright), though a public-domain source imposes nothing |
| **Import style** | By reference, **pinned by hash, not vendored**: `cmd/birdview-boundaries` fetches the archive, checks its hash, and writes `data/region-boundaries.json` ([tech.regionBoundaries](local/tech/delivery.md#regionboundaries)) |
| **Scope** | Region boundaries for the Here button ([product.hereSelectsLocalRegion](local/product/filtering.md#hereselectslocalregion)) |
| **Adopted parts** | Admin-1 polygons for the 51 US units (50 states and DC), the 13 Canadian provinces and territories, and five US Minor Outlying Islands: Johnston, Midway, Navassa, Palmyra, and Wake |
| **Exclusions** | Every other country, and Jarvis, Baker, and Howland, which are not BirdView regions |
| **Owner** | Natural Earth (Tom Patterson, Nathaniel Vaughn Kelso, and contributors) |
| **Status** | **Adopted 2026-10-18.** Retrieved, hashed, and the five atolls confirmed present. Credited in the index footer though no credit is required. |

License, quoted from <https://www.naturalearthdata.com/about/terms-of-use/>, read 2026-10-18:

> "All versions of Natural Earth raster + vector map data found on this website are in the public
> domain. … No permission is needed to use Natural Earth. Crediting the authors is unnecessary."

Its borders and coastlines are not GADM's, which the occurrence records were filed under, so near
a border or a shore the Here button may pick a neighboring region. GADM itself was not usable: its
license forbids redistribution, and the boundaries are sent to every reader's browser.

---

## `birdlife-checklist`

| Field | Value |
| --- | --- |
| **Origin** | HBW and BirdLife International digital checklist of the birds of the world, archived versions linked from <http://datazone.birdlife.org/species/taxonomy> |
| **Version** | Version 5 (December 2020), the version AVONET's `Metadata` sheet names for its BirdLife taxonomy; version 7 (December 2022) was also read |
| **Integrity** | **Not vendored, so no hash here.** No copy is kept, in the repository or elsewhere |
| **Tier** | mandatory (the site's terms bind anyone who uses it) |
| **Import style** | **Rejected.** Nothing from it enters the dataset, the repository, or the decision record |
| **Scope** | Evaluated for [decisions.vegaGullBorrows](decisions.md#vegagullborrows) (Vega Gull's AVONET donor); superseded by the vendored `avilist` and `avonet` workbooks |
| **Adopted parts** | None |
| **Exclusions** | Everything |
| **Owner** | BirdLife International |
| **Status** | **Rejected 2026-10-14.** Consulted 2026-10-13 by an agent in a way DataZone's terms forbid; quotations removed and [decisions.vegaGullBorrows](decisions.md#vegagullborrows) re-based on `avilist` and `avonet` |

**What the terms say.** The DataZone *Terms and Conditions of Use* (read 2026-10-13) govern the
site and the checklist files it links. Website clause 3 permits viewing on screen and copies of
extracts "for your personal use or private study"; any other use "requires the written prior
permission of BirdLife International", which it says it is "generally happy to authorise" for
non-commercial conservation use. **Website clause 11 forbids "any text or data mining or web
scraping"**, including any "automated device, program, tool … to access, obtain, copy, monitor
or republish any portion of the website" or its data, and any "automated analytical technique".
Data clause 11 forbids redistributing the Data, whole or in part, alone or combined with other
data.

**How it was consulted, and how that was resolved.** On 2026-10-13 an agent fetched the taxonomy
page, downloaded the version 5 and version 7 archives with `curl`, and searched the spreadsheets
with a program for *vegae* — automated access and automated analysis under clause 11. The copies
were deleted the same evening. On 2026-10-14 the three short quotations in [decisions.vegaGullBorrows](decisions.md#vegagullborrows)
were removed and [decisions.vegaGullBorrows](decisions.md#vegagullborrows) was re-based on the two CC BY 4.0 workbooks BirdView already vendors:
AviList v2025b's `Decision_summary` for Proposal `2025-1136` (which records that *L. vegae* is
sometimes treated as a subspecies of *L. smithsonianus* and that BirdLife's *L. argentatus* is the
European Herring Gull alone) and AVONET's `AVONET_Raw_Data` sheet (which shows that all 6
specimens of `AVONET1_BirdLife`'s *Larus argentatus* are `argentatus/argenteus` while 4 of the 8
specimens of `AVONET1_BirdLife`'s *Larus smithsonianus* are from Russia and Japan).

**What does not change.** AVONET's BirdLife sheet is AVONET's work, under CC BY 4.0, and is
unaffected; it carries BirdLife's species *names*, not its checklist. **No agent may fetch or
process anything from BirdLife's DataZone**, as for Wiley ([decisions.ricklefsProhibited](decisions.md#ricklefsprohibited)).

---

## `ebird-sensitive`

| Field | Value |
| --- | --- |
| **Origin** | <https://support.ebird.org/en/support/solutions/articles/48000803210-sensitive-species-in-ebird> |
| **Version** | Page modified 2026-05-15; retrieved 2026-10-09 by an agent with a plain HTTP fetch |
| **Integrity** | Not vendored |
| **Tier** | **mandatory** — it determines what the eBird half of the occurrence data *cannot* say |
| **Import style** | Absorbed as facts: the entries touching the US and Canada, for species BirdView draws, restated as `taxonomy.EBirdSensitive` |
| **Scope** | Where and when eBird withholds a species from public output, so that BirdView can tell eBird's silence from a bird's absence ([decisions.ebirdSilenceNotAbsence](decisions.md#ebirdsilencenotabsence)) |
| **Adopted parts** | 9 species from the global list, and 27 from the regional list as 68 entries of species, region, and date window |
| **Exclusions** | Everything outside the US and Canada, and every species BirdView does not draw. The rationale eBird stores for each listing is not published and not needed |
| **Owner** | Cornell Lab of Ornithology |
| **Status** | **Adopted 2026-10-09.** The list is not taken on trust: each entry is confirmed against the export before BirdView acts on it, and 26 of 230 region cells were not applied there |

**Why it matters.** GBIF's eBird export applies this list, including its *regional and
seasonal* entries, and nothing in BirdView knew. Painted Bunting has no eBird records in Florida
from October to March and 12,483 in April; the month filter read that as absence and hid the
bird from Florida in January, the heart of its season there.

**The list is not the export.** Checked cell by cell against the records GBIF serves, **42 of
the region cells the list names hold eBird records where it says none should be**: Greater
Sage-Grouse in Utah from February to May (1,689 records), Ferruginous Pygmy-Owl in Texas
(4,809), Spruce Grouse in Vermont (1,303). The list describes eBird's policy today; the export
reflects what was applied when it was built. Most of the 42 are a handful of stray records — a
Long-eared Owl or two per state, for reasons the export does not record — but some are plainly not suppressed at
all. So BirdView uses the list only to say *where to look*, and the records to say *whether it happened* — the rule is in [decisions.ebirdSilenceNotAbsence](decisions.md#ebirdsilencenotabsence).

**Why restating it is permitted.** Cornell's terms forbid redistributing eBird **data** —
checklists and observations — and that is why eBird/Clements is rejected below. This page is
neither: eBird's own data-use article classes "eBird.org products (web pages…)" as usable with
citation, and what BirdView takes from it is a set of facts — this species, this region, these
dates — restated in its own structure and cited here. No text from the page is copied. *Owner to
confirm*, as for every licensing judgment.

---

## Where geography enters

**Only through BirdView's own derivation, never through a source.** Recorded because it is the
question any change of geographic scope turns on, and the answer is not obvious from the table
above.

Every source BirdView reads is worldwide:

| Source | Geographic scope |
| --- | --- |
| **AviList** | worldwide — 11,131 species, 252 families |
| **AVONET** | worldwide — 10,661 species, 41 orders |
| **Fox wingspan** | worldwide — 10,618 species; gaps are flightless birds and Galliformes, not regions |
| **iNaturalist taxon IDs** | worldwide taxonomy; BirdView stores one id per in-scope species |
| **eBird and iNaturalist occurrence, via GBIF** | worldwide, at **very uneven density**: the US and Canada alone hold 1,238,404,492 eBird records ([decisions.regionJoinNormalizes](decisions.md#regionjoinnormalizes)) |

What makes BirdView North American is entirely its own:

| Where | What it fixes |
| --- | --- |
| `internal/region` | the 69 land regions — the 50 US states and DC, Canada's 13 provinces and territories, and five US outlying islands — and the five marine boxes. `birdview-status` derives scope over exactly these, and `birdview-regions` derives presence over them |
| the `US,CA` country filter | in `birdview-regions` (the pelagic share and the marine boxes), `birdview-names`, and `birdview-gbifmap` |
| the calibrations | 1 record per million for regional presence ([product.regionAndMonthFilter](local/product/filtering.md#regionandmonthfilter)), the recurrence thresholds of [product.scopeIsRecurringAvifauna](local/product/dataset.md#scopeisrecurringavifauna), and the 200-record month floor of [decisions.monthPresence](decisions.md#monthpresence) — every one measured against US and Canadian record density |

**The ABA checklist used to be the only thing in BirdView that knew where it was.** That stopped
being true when [decisions.scopeFromRecurrence](decisions.md#scopefromrecurrence) replaced it with scope derived from occurrence
records, and this section said otherwise until 2026-10-09.

Changing scope therefore costs nothing in the trait pipeline — no ingest, model, or diagram
depends on the region — and a great deal in the derivation: a new region table, the country
filter, and above all **re-measuring the calibrations**, which were tuned where records are
plentiful. The trait-side consequence is coverage: empirical wingspan runs at 35.8% inside the
current scope against 13.4% worldwide, so a wider list would be a thinner one.

---

## Considered and not adopted

Recorded so nobody re-litigates them annually.

### eBird / Clements taxonomy

Cornell terms clause 2 forbids reproduction or redistribution without written authorization.
Superseded anyway: ABA supplies common names, scientific names, and families, so no second taxonomy
is needed ([product.namesFromAvilist](local/product/dataset.md#namesfromavilist)).

### eBird API region species lists

Would supply scope and geography in AVONET's own taxonomy, but requires an API key and
`api.ebird.org/robots.txt` is `Disallow: /`, with the API terms page returning 403. Not evaluated
further; a human would need to read the terms.

### AOS / NACC check-list

Best-shaped data of the three checklists — family, order, and subfamily as real columns, plus
machine-readable status flags — but covers North *and Middle* America, which is the wrong scope, and
carries no license statement anywhere. Absence of a statement is not permission. Re-checked
2026-10-01 against the live CSV: 2,212 species, and filtering to neither accidental nor extinct
leaves **1,957** — 1,131 more than the ABA area's 841, because the area runs to Panama and the West
Indies. The status flags cannot reconstruct ABA scope; they are relative to a far larger region.

### AviList v2025b

**CC BY 4.0**, verified on the publisher's own download page 2026-10-01, and the consensus successor
to IOC, Clements, and BirdLife with the AOS, Cornell, BirdLife, Avibase, and the IOU behind it. It
would solve names, families, and taxonomic sequence outright — 832 of BirdView's 841 match on
scientific name. It does **not** solve scope, which is the only thing ABA is actually blocking, and
its English names are wrong for this audience. See
[below](#why-avilist-does-not-replace-the-aba-checklist).

### iNaturalist species counts

**Partly reinstated 2026-10-02, and this entry was wrong.** The original objection stands on its own
terms — counts overstate rarities, and for birds iNaturalist is about **80× shallower** than eBird
(13.4M US bird records against 1.07 billion), so it cannot rank anything. But "superseded by GBIF"
does not follow, and was refuted by measurement: eBird's export **withholds conservation-sensitive
species entirely**, and iNaturalist is the only open source measured that has them. Spotted Owl: 0
in eBird, 1,061 in iNaturalist. It is not a replacement for eBird; it is the complement that makes
eBird usable. See [`gbif`](#gbif).

### GBIF — eBird Observation Dataset (as a *scope* source)

**CC BY 4.0**, verified from GBIF's dataset record 2026-10-01, no key required, 1.07 billion US
records. This is *not* the eBird API rejected above: that was blocked by `robots.txt` and unreadable
terms, while the same observations published through GBIF are openly licensed. Technically excellent
and the only serious candidate for deriving scope. It still does not work — see
[below](#why-occurrence-data-cannot-reproduce-the-aba-scope). **Its taxonomy is a different question
and has since been adopted**: [`gbif`](#gbif) above takes the names, not the scope.

### iNaturalist taxonomy (as a names source)

Distinct from the taxon IDs already adopted, and not covered by them. The `inat-terms` adoption
rests on a narrow negative — nothing restricts reusing an *identifier* — while the same terms state
the agreement "does not transfer to You any iNaturalist or third party intellectual property", and
the CC BY-NC default governs contributor Content, not the taxonomy. No affirmative grant to
republish their name list was found.

### Lislevand et al. 2007

Copyright restrictions: none — genuinely open. Does not contain total length or wingspan, which is
what it was originally brought in for. Removed from the project.

### EltonTraits, Amniote DB, BIRDBASE, Dunning

None carries body length or wingspan. BIRDBASE was checked column by column against a claim that it
did.

### Audubon field guide

Carries exactly the two needed dimensions, in near-parseable form. Terms forbid reproduction,
distribution, storage in a retrieval system, and derivative works.

### All About Birds / Birds of the World

Carries the data; blocks automated retrieval (403) and is subscription-gated respectively. Not
assessed further.

### Macaulay Library

Permits *embedding* for non-commercial use — which BirdView qualifies for — but offers no discovery
API, and its media-request path explicitly prunes species lists over 100. Cannot scale to the
dataset. Viable later for hand-curated species.

### Wikimedia Commons, iNaturalist, Flickr

Photo sources, all researched in detail. Out of scope while [product.speciesPhotographs](local/product/diagram.md#speciesphotographs) is withdrawn;
Commons is the recommendation when it returns.

### Why AviList does not replace the ABA checklist

*Pass run 2026-10-01, after the ABA permission request went unanswered.* The question was whether
the checklist could be dropped entirely, since redistributing it is what blocks public deployment.

**What ABA actually supplies is three separable things**, and they have different answers:

| | Replaceable? |
| --- | --- |
| **Scope** — which 841 species, by rarity codes 1–3 ([product.scopeIsRecurringAvifauna](local/product/dataset.md#scopeisrecurringavifauna)) | **no open source found** |
| **Common names**, species and family ([product.namesFromAvilist](local/product/dataset.md#namesfromavilist)) | yes, but at a cost |
| **Taxonomic sequence** for the order grouping ([product.familyFilter](local/product/filtering.md#familyfilter)) | yes |

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
  recent genus moves — the same rot [decisions.compareUrlContract](decisions.md#compareurlcontract) handles for links, reappearing as a data-quality
  problem.

Each of these can be patched — special-case Hawaii, filter on establishment, pin a taxonomy —
and every patch is one more editorial judgment BirdView would be making and defending. At that
point the project has written its own checklist, which is a different undertaking from drawing
birds to scale.

**So the choice is real and narrow:** obtain the ABA's permission, or publish a list that is
about 84% of theirs, visibly different at the edges, and owned by BirdView. There is no third
option where the data simply comes from somewhere else.

**Update, 2026-10-02 — the conclusion stands and the framing does not.** The owner chose the
second branch: stop waiting on the ABA, publish a list BirdView owns. Two things above are now
misleading to read as current.

*The target was wrong.* Every F1 here measures agreement with the ABA's editorial judgment,
which is no longer the goal. A rule that admits the Red-crowned Parrot — 89,723 records, an
established population a birder in Texas will certainly see — is not 'wrong' because the ABA
treats it differently. The new question is whether BirdView's own rule can be stated in one
sentence and defended, not how closely it imitates someone else's.

*The method was unsafe.* These numbers come from the **species facet**, which was later found to
drop every record GBIF's backbone cannot resolve — including all of Cooper's Hawk. The agreement
figures are therefore depressed by an unknown amount, and the 70 'ABA has, rule misses' includes
species that were present all along. Re-measure with the inverted query before quoting any of
them; see [`gbif`](#gbif).
