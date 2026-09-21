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
and collapsing internal whitespace, with a committed mapping table covering the taxonomic
divergence between AOS/NACC (which ABA follows) and the 2022-era eBird/Clements taxonomy AVONET
carries.

Rationale: the measured join rate on binomial alone, over the P-001 scope, is **97.38%** — 819 of
841 — leaving 22 species needing a mapping entry. English name is the worse key: hyphenation and
Eastern/Western prefixes churn on every taxonomy update.

The 22 are not all renames, and that distinction is the subject of [CR-006](decisions.md#cr-006).
A rename may be mapped freely. A **split** cannot be mapped without borrowing the parent species'
measurements, which makes the value no longer a measurement *of that species* and therefore
engages P-060 and P-061.

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

### P-034 — A card opens a detail view
Subject: `presentation.detail_view`
Goals: G-001, G-003

Clicking a species card opens a fuller view of that species: its diagram at a larger size, every
metric with its value, and the provenance of each metric (P-063).

Rationale: the card is deliberately dense — it exists to be compared against dozens of others at a
glance, which leaves no room to say where each number came from. The detail view is where that is
said. It is the surface that makes P-062's provenance record worth recording, rather than metadata
nobody ever sees.

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

### P-060 — Where length and wingspan come from
Subject: `measurement.source`
Goals: G-001

- **Wingspan** — Fox et al. 2026, filtered to `needs_prediction = 0` so only empirical values are
  taken. Stored as the reported min/mid/max triple rather than a point value.
- **Length** — **modeled for every species.** No usable source of measured avian total length
  was found: Ricklefs 2017 was the only candidate and its terms prohibit the use
  ([CR-002](decisions.md#cr-002)).
- **Traits** — AVONET, read from whichever of its three taxonomy sheets carries the species as its
  own taxon ([CR-006](decisions.md#cr-006)).
- **Everything not covered above** — modeled.

Rationale: AVONET records neither macroscopic dimension, and no open source supplies both across
the ABA avifauna. Mixing sources of differing authority is acceptable *only* under P-062. See
[CR-001](decisions.md#cr-001).

### P-061 — The card marks any metric that is not measured
Goals: G-001

On the card itself — where numbers and the scaled diagram appear — any metric whose `kind` is not
`measured` is visibly marked as such, without the card having to explain why.

Rationale: this is what makes P-060 honest rather than merely convenient. If it is dropped, the
provenance record becomes metadata nobody sees and BirdView presents modeled numbers as data,
which is the thing the project explicitly rejected. It is load-bearing, not decoration.

### P-062 — Every presented metric carries a provenance record
Subject: `measurement.provenance.record`
Value: `{kind, source, sourceTaxon, model}`
Goals: G-001

Every metric BirdView presents — total length, wingspan, mass, and each diagram segment — carries
a record of where that specific value came from:

| Field | Meaning |
| --- | --- |
| `kind` | `measured`, `near-match`, or `modeled` |
| `source` | The dataset and its pinned version, including the sheet where that distinguishes rows |
| `sourceTaxon` | The binomial the value was actually filed under |
| `model` | For `modeled` values, the named model that produced it |

The three kinds:

- **`measured`** — a published measurement *of this species*, whatever name it is filed under and
  whichever sheet it sits in. A value read from `AVONET1_BirdLife`, or found under a pre-revision
  name such as *Charadrius nivosus* for *Anarhynchus nivosus*, is measured: the specimens are the
  right birds. `sourceTaxon` records the name so the claim is auditable.
- **`near-match`** — a published measurement of a **different taxon**, used because none exists
  for this species. Typically a pre-split parent: *Larus argentatus* standing in for *Larus
  vegae*. Real data, wrong bird, and the record says so.
- **`modeled`** — computed rather than measured. `model` names the computation.

Rationale: this generalizes what CR-001 established for lengths and CR-006 for split taxa. Those
were solved twice, differently, because there was no single place to say how confident a number
is. One record covers all of it, and makes each later gap a labeling problem rather than a
redesign.

### P-063 — The detail view explains each metric's provenance
Goals: G-001

In the detail view (P-034), each metric is shown with its provenance stated in plain language —
naming the source, and for a `near-match` naming the taxon actually measured, for a `modeled`
value naming the model.

The detail view also discloses that the **neck/body division is a drawing convention rather than a
measurement** (P-045). It is the only honest place to say so, since the diagram gives no hint that
one of its five segments is differently grounded from the other four.

Rationale: the card marks *that* a value is uncertain; this says *how*. "Wingspan estimated from
wing chord" and "mass measured on Herring Gull, a species this was split from in 2024" are very
different caveats, and a birder who cares enough to click deserves the difference.
