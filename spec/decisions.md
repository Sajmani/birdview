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

## Work arising

Not conflicts; obligations this bundle creates that nothing yet discharges.

| Item | Owner |
| --- | --- |
| Resolve CR-002 — read the Wiley terms, or ask, or drop measured length | Sameer |
| Ask ABA for written permission to redistribute the checklist | Sameer |
| Retrieve, hash and vendor `fox-wingspan`; write its `PROVENANCE.md` | — |
| Write `spec/sources/aba/PROVENANCE.md` with the CSV hash and the fetch procedure | — |
| Write `acceptance.md` — Gate 2, before any code | — |
