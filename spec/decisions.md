# Decisions

Append-only. Entries are never rewritten; a decision that turns out badly is superseded by a new
entry, not edited.

Each record states the subject, the conflicting requirements, the kind of conflict, the options
considered, the decision, who made it, and the resulting status of everything it touched.

---

## CR-001 — No open source supplies both body length and wingspan

**Subject:** `measurement.provenance`
**Kind:** Over-constraint — no value satisfies every applicable constraint
**Date:** 2026-09-20 · **Decided by:** Sameer

### The conflict

Three requirements could not hold at once:

1. Lengths and wingspans shown to the user are real published measurements, not model output
   (owner's stated position, after rejecting the earlier design's invented allometry).
2. The dataset covers the regularly occurring ABA avifauna — 841 species (P-001).
3. Inputs are freely licensed and machine-readable (T-004, T-021).

Research against actual sources established that no combination satisfies all three. AVONET
records neither dimension. Of the candidates that do: Fox et al. 2026 is CC BY 4.0 but carries
empirical wingspan for only ~1,424 species globally, and 6 of 20 common North American landbirds
sampled; Ricklefs 2017 has genuine ruler-measured total length but only for passerines and is not
openly licensed; Wikidata is CC0 with 1,110 wingspans and **nine** lengths; Audubon and All About
Birds have both dimensions and prohibit or block reuse. EltonTraits, the Amniote database,
BIRDBASE, Lislevand and Dunning were each checked and carry neither.

### Options considered

| Option | Satisfies | Violates | Cost |
| --- | --- | --- | --- |
| Real data only, accept gaps | 1, 3 | 2 | ~30% wingspan coverage, passerine-only length. Most cards incomplete or absent. |
| Return to a modeled source | 2, 3 | 1 | Complete coverage; every number synthetic, which is the position already rejected. |
| Per-value provenance | 2, 3, and 1 *conditionally* | — | Full coverage with each value labeled measured or estimated. Requires the labeling to be real. |
| Pursue licenses first | possibly all | — | Indefinite delay, uncertain outcome. |

### Decision

**Per-value provenance.** Real measurements where they exist, modeled estimates elsewhere, each
tagged, with the interface distinguishing them.

The conditional matters: this only satisfies (1) because **P-061 makes the labeling
user-visible**. Without P-061 the decision degenerates into presenting modeled numbers as data,
which is the thing that was rejected. P-061 is therefore not a nice-to-have and must not be
dropped as a simplification.

### Resulting status

- **P-060** Active, as amended.
- **P-061** Active, and load-bearing for this decision.
- Honest expectation: most v1 lengths will be estimates. Ricklefs covers passerines only, so
  roughly half of the 841 species have no measured length available at any price.

---

## CR-002 — Ricklefs 2017 is not openly licensed *(open)*

**Subject:** `dataset.length.source`
**Kind:** Cross-source — a mandatory source's terms versus a local requirement
**Date raised:** 2026-09-20 · **Status: OPEN — owner action required**

Ricklefs 2017 is the only source of real total-length measurements for North American birds that
this project found. Crossref reports only Wiley version-of-record terms and no CC license, and the
Wiley terms page returns HTTP 403 to automated retrieval — so **the terms have not been read**.

Per [process.md](process.md#some-sources-refuse-to-be-fetched), the gap is left unfilled rather
than guessed at. Specifically *not* relied upon: that the `RS-eco/traitdata` R package
redistributes the data under GPL-3 (a package license does not establish the data's license), and
the argument that measurements are uncopyrightable facts (a legal judgment nobody here is
qualified to make).

**Three exits, none yet taken:**

1. A human reads the Wiley terms and records what they permit.
2. Ask Ricklefs or ESA for permission for non-commercial open-source use.
3. Drop the measured-length half of P-060, making every length an estimate.

Until one is taken, the measured-length branch of P-060 is **Deferred (owner: Sameer)**. The
modeled branch is unaffected, so this does not block v1 — it bounds how good v1 can be.

---

## CR-003 — Location filtering has no redistributable data source

**Subject:** `filter.location`
**Kind:** Over-constraint
**Date:** 2026-09-20 · **Decided by:** Sameer

The owner asked for filtering by location. The species-list source chosen for P-001 carries no
geography, and no freely redistributable per-region occurrence data was found at useful
granularity. eBird's region species-list endpoint would serve, but requires an API key, and
`api.ebird.org/robots.txt` is `Disallow: /` with the API terms page returning 403 — so what it
permits is unestablished.

**Decision:** withdraw P-011 and P-012 from v1 rather than defer them, and revisit with a range
source. Withdrawing rather than deferring is deliberate: a `Deferred` entry needs an owner and a
date, and there is no date until a source exists.

**Resulting status:** P-011 Withdrawn · P-012 Withdrawn · family is the only filter.

---

## CR-004 — The neck/body split has no measured basis

**Subject:** `diagram.neck_ratio`
**Kind:** Over-constraint, and a consistency question against CR-001
**Date:** 2026-09-20 · **Decided by:** Sameer

With a total length in hand, AVONET supplies culmen and tail directly and head follows from bill
depth — but nothing measures where the neck ends and the body begins. The earlier design used
hand-assigned per-family neck ratios, which is the same class of invented number that CR-001
rejected for lengths.

| Option | Consequence |
| --- | --- |
| Per-family ratio | Herons look like herons; the split is a guess with no source |
| One fixed ratio | No invented numbers; every bird the same shape, largely defeating G-003 |
| Find a source | No per-species neck-length data is believed to exist at this coverage |
| Merge neck into body | Drops a segment the owner explicitly asked to keep |

**Decision:** keep the per-family ratio, **declared as a visual design parameter** carrying no
measurement claim and never surfaced as data (P-045).

This is consistent with CR-001 rather than in tension with it, on one condition: CR-001 governs
*numbers presented to the user as measurements*, and this governs *pixels in a schematic*. The
distinction holds only while the ratio stays out of the dataset and out of the UI's numbers. If a
neck length ever appears as a value, this decision is void and must be reopened.

---

## CR-005 — Photographs withdrawn from v1

**Subject:** `presentation.mode`
**Kind:** Scope
**Date:** 2026-09-20 · **Decided by:** Sameer

Photographs were researched in full before being scoped out, and the findings are recorded here so
the work is not repeated:

- **Wikimedia Commons** is the recommended primary. Its licensing policy guarantees commercial-use
  compatibility, the Action API returns structured license metadata (`extmetadata`) sufficient to
  build a credit line automatically, no API key is needed, caching is permitted and preferred over
  hot-linking, and per-species `in flight` categories exist. There is **no "perched" equivalent**
  — 30 category hits in total, mostly cats — so in-flight is a positive filter and perched is the
  residual default.
- **iNaturalist** is a viable fallback *because BirdView is non-commercial* (T-011), which makes
  its default CC BY-NC corpus usable. Its API supplies a ready-made attribution string per photo.
  It has a **verified absence** of behavior annotation: the controlled-terms endpoint returns
  seven terms, none of them about flight or perching.
- **Macaulay Library** permits embedding for non-commercial use, quoted from its own page: *"The
  appropriate data and credit will automatically display along with the image… embedding/sharing
  media should be used for non-commercial purposes only."* Raw CDN hot-linking is a different act
  and is not covered by that permission. The blocker is not licensing but discovery — there is no
  public API to find a good photo of a given species, and the media-request path prunes species
  lists over 100.
- **Flickr** is legally usable but operationally hostile: caching is forbidden, galleries are
  capped at 30 photos per page, and there is a 24-hour takedown-propagation duty.

**Decision:** ship v1 with diagrams only. P-050 through P-053 are withdrawn, not abandoned.

**Resulting status:** P-030 reduced to a single mode · P-050–P-053 Withdrawn.

---

## CR-006 — ABA species with no AVONET row of their own *(resolved)*

**Subject:** `dataset.join.key`
**Kind:** Derived — no pair of requirements contradicts; the conflict appears only when P-006's
mapping meets the provenance rule
**Date raised:** 2026-09-21 · **Resolved:** 2026-09-21 · **Decided by:** Sameer

### Resolution

Two decisions, together:

1. **Read each species from whichever AVONET taxonomy carries it as its own taxon**, recording the
   sheet and the filed name in its provenance record. This reduces the problem from 9 species to
   6.
2. **For the remaining 6, borrow the parent taxon's traits and label them `near-match`** under
   P-062. All 841 species stay in the dataset, and every borrowed value says whose measurement it
   actually is.

A value read from another sheet, or found under a pre-revision name, is **`measured`** — the
specimens are the right birds and only the filing changed. `near-match` is reserved for a value
measured on a genuinely different taxon.

This mirrors CR-001 exactly: mixed authority is acceptable when, and only when, each value is
labeled. P-062 now provides the vocabulary both decisions needed, and the detail view of P-034 /
P-063 is where it reaches the user.

**Resulting status:** P-006 Active as amended · P-062 Active · P-063 Active · the six near-match
species are `Larus vegae`, `Numenius hudsonicus`, `Sturnella lilianae`, `Vireo swainsoni`,
`Sula brewsteri`, `Anthus japonicus`.

**Carried forward, not closed by this:** the mass caveat below, and the four renames-in-disguise
at the end, still need a human taxonomic check before the mapping table is written.

---

### Original analysis

### What was measured

Joining ABA 8.19 codes 1–3 against `AVONET2_eBird` on normalized binomial: **819 of 841 = 97.38%**.
AVONET carries the 2022-era eBird taxonomy, so four years of revision separate the two sources.
Full classification is in [sources/aba/PROVENANCE.md](sources/aba/PROVENANCE.md#taxonomy).

Twelve misses look like renames and can be mapped. **Nine are splits**, where the pre-split parent
is still a separate ABA species in its own right:

| ABA species | Would borrow traits from |
| --- | --- |
| American Herring Gull *Larus smithsonianus* | *Larus argentatus* |
| Vega Gull *Larus vegae* | *Larus argentatus* |
| Hudsonian Whimbrel *Numenius hudsonicus* | *Numenius phaeopus* |
| Cory's Shearwater *Calonectris borealis* | *Calonectris diomedea* |
| Northern Yellow Warbler *Setophaga aestiva* | *Setophaga petechia* |
| Chihuahuan Meadowlark *Sturnella lilianae* | *Sturnella magna* |
| Western Warbling-Vireo *Vireo swainsoni* | *Vireo gilvus* |
| Cocos Booby *Sula brewsteri* | *Sula leucogaster* |
| Siberian Pipit *Anthus japonicus* | *Anthus rubescens* |

### Why this is a conflict rather than a chore

Mapping a split onto its parent silently attributes one species' measurements to another. Two of
the rows above resolve to the *same* parent, so American Herring Gull and Vega Gull would render
as pixel-identical birds in an application whose entire purpose is showing that birds differ in
size and shape.

It also collides with [CR-001](#cr-001). P-060 divides every value into *measured* or *estimated*;
a trait borrowed from a pre-split parent is neither. It is a real measurement of a different
taxon, which is a third category the provenance model does not currently have.

### Options

| Option | Effect |
| --- | --- |
| Borrow parent traits, tag as a third provenance class | All 841 species present; honest labeling; P-060/P-061 grow a category |
| Borrow parent traits silently | Complete dataset, false precision, and identical gull diagrams |
| Exclude the 9 | 832 species, every one genuinely its own measurement; some common birds missing |
| Source the 9 elsewhere | Correct and slow; no identified source |

### Amended 2026-09-21 — most of them are in AVONET after all

The analysis above searched only `AVONET2_eBird`. AVONET ships **three parallel taxonomies**, and
searching all of them changes the numbers materially:

| Resolution | Count | Species |
| --- | --- | --- |
| Present in `AVONET2_eBird` under an older name | 11 | the four *Anarhynchus* (ex-*Charadrius*), *Astur* ×2 (ex-*Accipiter*), *Ardea ibis* (ex-*Bubulcus*), *Botaurus exilis* (ex-*Ixobrychus*), *Gygis candida* (ex-*G. alba*), *Tyto furcata* (ex-*T. alba*), *Ramosomyia violiceps* (ex-***Leucolia violiceps***) |
| Own row in `AVONET1_BirdLife` | 4 | *Larus smithsonianus*, *Calonectris borealis*, *Pterodroma deserta*, *Spilopelia chinensis* |
| Own row in `AVONET3_BirdTree` | 1 | *Setophaga aestiva*, as *Dendroica aestiva* |
| **Absent from every AVONET taxonomy** | **6** | *Larus vegae*, *Numenius hudsonicus*, *Sturnella lilianae*, *Vireo swainsoni*, *Sula brewsteri*, *Anthus japonicus* |

*Ramosomyia violiceps* is resolved: the earlier lookup guessed *Amazilia violiceps*, which is the
BirdLife and BirdTree name. `AVONET2_eBird` calls it *Leucolia violiceps*.

**A new option therefore exists:** read trait rows from whichever AVONET taxonomy sheet carries
the species as its own taxon, recording which sheet each row came from. For the five species
above this yields genuine own-taxon measurements instead of a borrowed parent's.

**And a new caveat.** Where BirdLife splits a pair, the *linear* traits differ but **mass is
identical**:

```
Larus smithsonianus   culmen 69.3  depth 18.3  tail 166.0  wing 429.5  mass 1090.99
Larus argentatus      culmen 74.0  depth 19.6  tail 170.2  wing 430.8  mass 1090.99
```

The same holds for *Calonectris borealis*/*diomedea* (both 534.99 g) and *Pterodroma
deserta*/*feae* (both 311 g). AVONET's mass column is sourced separately from its morphometrics
and was evidently not split. So diagrams would differentiate correctly while **ordering by mass
(P-020) would still show the pair as identical** — a smaller version of the original problem,
confined to one of the three sort metrics.

### Options considered

| Option | Effect | |
| --- | --- | --- |
| Read across all three AVONET sheets, recording the source sheet per species | Reduces the problem from 9 to 6; mixes taxonomies, which must be recorded per row | **chosen** |
| For the remaining 6, borrow parent traits as `near-match` | All 841 present, honestly labeled | **chosen** |
| For the remaining 6, exclude them | 835 species, every trait its own taxon's | rejected — loses common birds for a labeling problem |
| Borrow silently | Complete dataset, false precision, identical gull diagrams | rejected |

### Still unresolved

The rename/split classification remains provisional. *Astur atricapillus*, *Ardea ibis*, *Gygis
candida* and *Tyto furcata* are listed as renames because the parent is absent from ABA, but each
is really a split whose `AVONET2_eBird` row represents the combined taxon. Mapping them is
therefore the same borrowing problem in a quieter form, and each needs a human taxonomic check
before entering the mapping table.

---

## Work arising

Not conflicts; obligations this bundle creates that nothing yet discharges.

| Item | Owner |
| --- | --- |
| Resolve CR-002 — read the Wiley terms, or ask, or drop measured length | Sameer |
| Ask ABA for written permission to redistribute the checklist | Sameer |
| Retrieve, hash and vendor `fox-wingspan`; write its `PROVENANCE.md` | — |
| Write `spec/sources/aba/PROVENANCE.md` with the CSV hash and the fetch procedure | — |
| Write `acceptance.md` — Gate 2, before any code | — |
