# Product requirements

What BirdView must do, from a user's point of view. Normative: where this file and the code
disagree, the code is wrong — or this file needs an acknowledged amendment.

IDs are permanent. Withdrawn requirements stay in place so references to them still resolve.

---

## Dataset and taxonomy

### P-001 — Scope is the regularly occurring ABA Area avifauna
Subject: `dataset.scope`
Value: `ABA Area, rarity codes 1-3`
Goals: G-001, G-002

The dataset is the species of the ABA Area — North America north of Mexico, plus Hawaii —
restricted to ABA rarity codes 1 through 3, which is 841 of the 1,161 species in ABA Checklist
8.19.

Rationale: codes 4–6 are casual, accidental, extinct or extirpated — 320 species a birder will
essentially never encounter, many of them Old World strays whose presence among the common birds
makes the comparison less useful, not more. The rarity column is read at ingest purely as a scope
filter; it is not carried into the dataset and not shown in the UI (see P-002 for why that
distinction matters to the license).

### P-002 — Morphometric traits come from AVONET
Subject: `dataset.traits.source`
Value: `AVONET2_eBird`
Goals: G-001, G-003

Bill, wing, tail and mass measurements come from AVONET's `AVONET2_eBird` sheet, which uses
eBird/Clements taxonomy. See [sources.md](sources.md#avonet).

### P-003 — Traits carried per species
Goals: G-003

Each species record carries scientific name, family, and these AVONET traits: `Beak.Length_Culmen`,
`Beak.Depth`, `Tail.Length`, `Wing.Length`, `Secondary1`, `Mass`.

Rationale: this is the set the diagram and its derivations actually consume. AVONET's other
columns — `Beak.Width`, `Hand-Wing.Index`, `Tarsus.Length` — are deliberately not ingested;
carrying unused data invites someone to invent a use for it.

### P-004 — Names and families come from the ABA checklist
Subject: `dataset.names.source`
Value: `ABA Checklist`
Goals: G-002

English common name, scientific name, family scientific name and family common name all come from
the ABA checklist. Its section-header rows are formatted `Sandpipers and Allies (Scolopacidae)`,
supplying both halves of the family name.

Rationale: one source for all naming, rather than joining a second taxonomy for common names
alone. AVONET carries no common name at all, and no English-name column exists on any of its
sheets.

### P-005 — Family badge colors are generated
Subject: `badge.color.derivation`
Value: `generated`
Goals: G-002

Each family gets a badge color derived algorithmically — evenly spaced hues, or a stable hash of
the family name — constrained to satisfy T-010.

Rationale: the ABA area spans roughly 80 families. A hand-curated palette does not scale to that
and breaks on every taxonomic revision. Generation makes the palette a function of the taxonomy
rather than a table to maintain.

### P-006 — Sources join on normalized scientific binomial
Subject: `dataset.join.key`
Value: `scientific binomial, normalized`
Goals: G-001

Species are joined across sources on scientific binomial, after stripping `(...)` parentheticals
and collapsing internal whitespace, with a committed synonym table covering known genus-level
disagreements between AOS/NACC (which ABA follows) and eBird/Clements (which AVONET follows).

Rationale: measured join rate on binomial before any synonym table is 99.4%; all misses are genus
reshuffles such as *Dryobates* vs *Leuconotopicus*. English name is the worse key despite scoring
marginally higher on one measurement — hyphenation and Eastern/Western prefixes churn on every
taxonomy update.

### P-007 — Unmatched species are reported, not silently dropped
Goals: G-001

Ingest reports every species that fails to join, by name and count, and fails the build if the
unmatched fraction exceeds a stated threshold.

Rationale: ABA, AOS and eBird/Clements revise on independent annual cycles. Without a loud
failure, a taxonomy update silently shrinks the dataset and nobody notices.

---

## Filtering

### P-010 — Filter by family
Goals: G-002

The user can filter to any subset of families, multi-select, with a single control that clears the
filter and shows everything.

### P-011 — Filter by location
Status: **Withdrawn** — no freely redistributable per-region occurrence data exists at useful
granularity. See [CR-003](decisions.md#cr-003).

### P-012 — Filters compose
Status: **Withdrawn** with P-011. Family is the only filter in v1.

---

## Ordering and sectioning

### P-020 — One primary metric, user-selected
Subject: `ordering.metric`
Value: `length | mass | wingspan`
Goals: G-001, G-002

The user chooses whether the dataset is ordered by total length, body mass, or wingspan.

### P-021 — Ascending order
Goals: G-001

Species are sorted ascending by the selected metric.

### P-022 — Sections break at N% above the section minimum
Subject: `sectioning.rule`
Goals: G-001, G-002

A new section begins at the first species more than N% larger than the **minimum** species of the
previous section.

Rationale: bird size follows a power law. Fixed-width bins put hundreds of small birds in one row
and isolate large raptors in rows of one. A ratio rule produces sections of comparable visual
span at every scale — and, because a section is by construction no more than N% wide, makes the
per-section scale of P-042 honest.

### P-023 — N defaults to 15 and is adjustable
Subject: `sectioning.step_percent`
Value: `15`
Goals: G-002

### P-024 — Section headers state range and count
Goals: G-001

Each section shows its range in the currently selected metric and the number of species in it.

Rationale: under per-section scaling (P-042) the header is what tells the user that the scale has
changed between one section and the next. Without it, per-section scaling is misleading.

---

## Presentation

### P-030 — Diagram is the only presentation mode in v1
Subject: `presentation.mode`
Value: `diagram`
Goals: G-001, G-003

Rationale: photographs were scoped out of v1 (P-050). The diagram is what makes size and shape
comparable; photographs were always the enrichment.

### P-031 — Card contents
Goals: G-001, G-002, G-003

Each species is one card showing its common name, its family badge, the diagram, and its length,
wingspan and mass. Cards wrap to fill the available display width.

### P-032 — Imperial by default, metric available
Subject: `units.default`
Value: `imperial`
Goals: G-001

Lengths in inches and mass in grams by default, with a toggle to metric.

Rationale: the ABA Area audience is predominantly US; the toggle serves Canadian users and anyone
who thinks in centimeters.

### P-033 — First load shows everything
Goals: G-002

With no filter applied, the initial view is all species, ordered by length.

Rationale: the full sweep from hummingbird to swan is the clearest statement of what the app is
for. Opening filtered or empty would hide it.

---

## Diagram

### P-040 — Segment composition
Subject: `diagram.segments`
Value: `bill, head, neck, body, tail, wing crossbar`
Goals: G-003

The diagram is a vertical cross. Top to bottom: a bill stick, a head oval, a neck stick, a body
stick, and a tail stick — plus one horizontal crossbar representing the wings. The tail is
visually distinguished from the body by stroke style and a mark at the junction between them.

Rationale: the tail is a stick rather than a triangle because no trait in any source gives a tail
width, and a triangle would assert a fan shape the data does not support.

### P-041 — Segments abut exactly and sum to total length
Subject: `diagram.segments.sum`
Goals: G-001, G-003

The five vertical segments meet without gap or overlap, and their lengths sum to the species'
total length.

Rationale: this is what lets the vertical extent of the drawing *be* the bird's length, so that
two diagrams at the same scale can be compared directly.

### P-042 — Scale is a user toggle
Subject: `diagram.scale.mode`
Value: `per-section (default) | global`
Goals: G-001, G-002

The user switches between a scale shared by all species within a section, and a single scale
across the entire filtered set. Per-section is the default.

Rationale: this is the requirement G-001 lives or dies by. A per-card scale — every bird drawn to
fill its card — makes a hummingbird and a swan identical on screen and destroys size intuition
entirely. Per-section keeps every card legible while remaining honest within the section, since a
section spans at most N%. Global is brutally honest across the whole range and renders small birds
as specks; it is the view that produces the cross-family surprise, so it is offered rather than
imposed.

### P-043 — The wing crossbar is drawn at the vertical scale
Subject: `diagram.wingspan.scale`
Value: `same as vertical axis`
Goals: G-001

The crossbar's length represents wingspan at the same scale as the vertical axis, so its length
relative to the body is literal.

Rationale: wingspan runs 1.0–2.5× body length, so honoring this forces a wide card — roughly
2.6:1. That is a real layout cost, accepted because a crossbar on its own arbitrary scale tells
the user something false about proportion while looking authoritative.

### P-044 — Proportion differentiates species
Goals: G-003

A dowitcher reads as long-billed, an accipiter as long-tailed, a heron as long-necked.

### P-045 — The neck/body split is a declared design parameter
Subject: `diagram.neck_ratio`
Value: `per-family constant, non-measured`
Goals: G-003

The division of the remaining body axis between neck and body uses a hand-assigned per-family
ratio. **It is a visual design parameter, not a measurement.** It carries no provenance claim, is
never surfaced to the user as data, and is recorded as a table in the code with this requirement
cited beside it.

Rationale: no source measures avian neck length at this coverage. The alternative — one ratio for
every species — is more honest and makes every bird the same shape, which defeats G-003. This is
the one number in BirdView that is invented, and it is confined to pixels. See
[CR-004](decisions.md#cr-004).

---

## Photographs

### P-050 — Species photographs
Status: **Withdrawn from v1.** See [CR-005](decisions.md#cr-005) for the research and the reasons
it is deferred rather than abandoned.

### P-051 — Perched and in-flight selection
Status: **Withdrawn from v1** with P-050. Note for whoever revives it: Wikimedia Commons has
per-species `in flight` categories but no "perched" equivalent, so the design must treat in-flight
as a positive filter and perched as the residual default.

### P-052 — Photograph attribution
Status: **Withdrawn from v1** with P-050.

### P-053 — Local caching of photographs
Status: **Withdrawn from v1** with P-050.

---

## Measurements

### P-060 — Length and wingspan carry per-value provenance
Subject: `measurement.provenance`
Value: `measured | estimated, per value`
Goals: G-001

Every total length and every wingspan is tagged as either a published measurement or a modeled
estimate:

- **Wingspan, measured** — Fox et al. 2026, filtered to `needs_prediction = 0` so that only
  empirical values are taken. Stored as the reported min/mid/max triple rather than a point value.
- **Length, measured** — Ricklefs 2017, passerines only. Subject to the unresolved license
  question in [CR-002](decisions.md#cr-002).
- **Everything else** — modeled from AVONET traits and tagged as an estimate.

Rationale: AVONET records neither dimension, and no open source supplies both across the ABA
avifauna. Mixing measured and modeled values is acceptable *only* because each is labeled. See
[CR-001](decisions.md#cr-001).

### P-061 — The interface distinguishes measured from estimated
Goals: G-001

Wherever a length or wingspan is shown — as a number, and as a scaled diagram — the interface
makes clear whether it is measured or estimated.

Rationale: this requirement is what makes P-060 honest rather than merely convenient. If it is
dropped, P-060 becomes a way of presenting modeled numbers as data, which is the thing the project
explicitly rejected. It is load-bearing, not decoration.
