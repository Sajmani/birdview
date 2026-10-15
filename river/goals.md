# Goals

Why BirdView exists. Goals are objectives in the OKR sense — directional and deliberately not
measurable. Requirements in [product.md](product.md) and [tech.md](tech.md) cite them in a
`Goals:` field; a goal never lists its requirements, so the two cannot drift apart.

Whether a requirement genuinely *serves* the goal it names is a Gate 1 judgment. No criterion can
check it.

---

### G-001 — Size intuition

A birder using BirdView comes away knowing how big a bird actually is, relative to birds they
already know.

Field guides give numbers, and numbers do not build intuition: almost nobody can picture the
difference between an 8.5-inch bird and an 11-inch one, even though one is noticeably larger in
the field. Seeing them drawn to a shared scale, side by side, does what the numbers do not.

### G-002 — Comparison within and across families

The comparison works *within* a family — how do these sparrows actually differ? — and *across*
families — is a Merlin bigger than a Blue Jay?

These are different questions with different failure modes. Within a family, birds are close in
size and the interesting variation is proportional. Across families the size range is enormous, and
the honest answer is often surprising. A design that serves one can easily defeat the other, which
is why scale is a user choice rather than a fixed decision (P-042).

### G-004 — A good citizen of the data commons it depends on

BirdView honors the terms of every source it uses, credits them, and never presents someone
else's work as its own or as better-grounded than it is.

Almost nothing in BirdView is BirdView's. The measurements are AVONET's, the names are
AviList's, the wingspans are Fox et al.'s, and the occurrence records are eBird's and
iNaturalist's through GBIF. Each arrives with conditions — attribution, non-commercial
use, no redistribution — and honoring them is a condition of the project existing at all, not
overhead on top of it. This goal is why the project rejected a dataset it wanted
([CR-002](decisions.md#cr-002)), fetches one source at build rather than committing it (T-004),
and records where every number came from (P-062).

It is also why BirdView is worth more than the sum of its sources: a compilation that quietly
misrepresented their provenance would be worth less than nothing to the people who produced them.

### G-003 — Shape, not only size

Proportion is visible at a glance: which birds are long-billed, long-necked, long-tailed.

Size alone would make a dowitcher and a robin of equal length look identical, and in the field
they look nothing alike. Shape is what makes the comparison recognizable rather than abstract.

### G-005 — The data as it is, not as it would be tidier

BirdView shows what its sources actually record — including where they disagree with each
other — and declines to smooth it over.

Every convenience here is a small lie that compounds. A rarity label puts somebody's threshold
where a number the reader could have judged belongs. Merging records across a taxonomic split
asserts a division nobody measured. Dropping a name the sources disagree about makes the
disagreement invisible rather than settled. The honest move is the same in each case: show the
figure, name the disagreement, and let the reader see the seam.

The hard half is that this sometimes means **publishing a worse-looking answer**. Where the
sources cannot be reconciled, the page says they cannot, rather than picking one and reading
smoothly. A number that is quietly wrong costs more than a gap that is loudly honest, because
only one of them can be checked.

### G-006 — Nothing disappears quietly

BirdView does not silently drop data. Every omission, merge or filter that hides information is
stated where a reader would otherwise be misled by its absence.

The three ways data goes missing here are a species filtered out of scope, two sources merged
under one name, and a source that withholds records outright — and all three are invisible by
construction. A reader cannot see a bird that is not drawn, cannot tell that one tile's count
covers two species, and cannot distinguish a bird nobody records from a bird eBird refuses to
publish. Absence leaves no mark unless something marks it.

This is the operational half of [G-005](#g-005). That goal says show the data as it is; this one
says the gaps are part of the data. It is why a species out of scope keeps a page saying why
([P-075](product.md#p-075)), why a zero is printed against its denominator rather than dashed
([P-074](product.md#p-074)), and why a count covering two species says so on both
([P-077](product.md#p-077)).

---

## Non-goals

Stated as deliberately as the goals. Each of these is a plausible thing a bird app does, and
BirdView does not do it.

- **Bird identification.** No "what did I see?" flow, no ID keys, no photo matching. BirdView
  assumes you know what the bird is and want to know how big it is.
- **Songs and calls.** No audio.
- **Range maps.** No cartography, and as of v1 no location filtering either (P-011, withdrawn).
- **Life lists and personal data.** No accounts, no sightings, no eBird list integration, nothing
  persisted about the user.
- **Commercial operation.** BirdView is not and will not become a commercial offering. This is a
  standing constraint, not a current circumstance — several sources that BirdView may later draw
  on are usable only on that basis (T-011).

Taxonomy browsing beyond the family filter is *not* a non-goal, merely unbuilt.

A per-species detail view **is** in scope, as of P-034 — it is where a metric's provenance gets
explained (P-063), which the tile has no room for.
