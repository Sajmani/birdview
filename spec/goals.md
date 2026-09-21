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

### G-003 — Shape, not only size

Proportion is visible at a glance: which birds are long-billed, long-necked, long-tailed.

Size alone would make a dowitcher and a robin of equal length look identical, and in the field
they look nothing alike. Shape is what makes the comparison recognizable rather than abstract.

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

Two things are *not* non-goals, merely unbuilt: taxonomy browsing beyond the family filter, and
per-species detail pages.
