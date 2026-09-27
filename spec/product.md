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
`Tail.Length`, `Wing.Length`, `Secondary1`, `Mass`.

Rationale: this is the set the diagram and its derivations actually consume. AVONET's other
columns — `Beak.Width`, `Beak.Depth`, `Hand-Wing.Index` — are deliberately not ingested; carrying
unused data invites someone to invent a use for it.

The set moves both ways, and a column that stops earning its place leaves. `Beak.Depth` was
ingested until the cranium model left bill depth for mass. `Tarsus.Length` made the whole round
trip — excluded as unused, ingested when the diagram gained legs, then dropped again when the
flight premise removed them ([CR-010](decisions.md#cr-010)).

### P-004 — Names and families come from the ABA checklist
Subject: `dataset.names.source`
Value: `ABA Checklist`
Goals: G-002

English common name, scientific name, the **four-letter banding code**, family scientific name
and family common name all come from the ABA checklist. Its section-header rows are formatted
`Sandpipers and Allies (Scolopacidae)`, supplying both halves of the family name; the banding code
is a per-species column.

Rationale: one source for all naming, rather than joining a second taxonomy for common names
alone. AVONET carries no common name at all, and no English-name column exists on any of its
sheets.

### P-005 — Family badge colors are generated
Subject: `badge.color.derivation`
Value: `generated`
Goals: G-002

Colors are generated as **two palettes**: one per taxonomic order, and one per family *within* its
order. Both satisfy T-010, and families are visibly different from those listed next to them.

Two families in different orders may share a color. Their order badge distinguishes them.

Rationale: the ABA area spans 91 families. A hand-curated palette does not scale to that and
breaks on every taxonomic revision, so generation makes the palette a function of the taxonomy
rather than a table to maintain.

The two-palette split is what makes the problem solvable at all. A legible badge bounds the usable
gamut at roughly 43 mutually distinguishable colors with black text, or 69 if white text is also
allowed — against 91 families. No assignment fits 91 into 43. Scoping family colors to their order
reduces the requirement to the largest order, Passeriformes at 44 families, which does fit.

Distinctness must be measured perceptually, and between the entries a reader actually compares.
Evenly spaced hues satisfy "distinct colors" on paper and failed completely in practice: at 91
families that is a four-degree step, and the selector listed families in that same order, so every
neighbor was the same shade.

**Known residual:** within Passeriformes, 9 of 946 family pairs remain below ΔE 15. The dark half
of the gamut has less perceptual room than the light half, and 44 families is close to what the
space holds. Those pairs are never adjacent in the selector, and every badge carries its family
name.

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

The user can filter to any subset of families, multi-select. **Everything is selected on arrival**
— All, every order and every family show as on. An **All** control toggles like an order does,
over the whole set: untoggling it deselects everything and **no birds are displayed**.

Selection is therefore always literal: what is shown is what is selected, and the controls say so.
There is no separate "unfiltered" state that happens to look like everything being chosen.

The selector is **grouped by taxonomic order, in the ABA checklist's own sequence** rather than
alphabetically. Each group is headed by a **toggle that selects or clears every family in that
order**, labeled with the order's scientific name, its common name, and its species count, and
styled like the other controls rather than like a family chip.

Order common names are **the project's own editorial labels**. No source BirdView uses publishes
vernacular names for orders, and several are deliberately loose — Charadriiformes really is
shorebirds plus gulls plus auks. Nothing is derived from them. Each family is labeled with
**both its scientific and its common name** and the number of species it contributes.

Rationale: the selector is where the common name earns its place. A birder looking for nightjars
should not have to know they are Caprimulgidae to find them, and this is the one surface where
they are choosing a family rather than reading about one they have already been shown.

Checklist sequence is how the ABA checklist, Sibley and every other field guide arrange birds, so
it is the order a birder already has in their head — and alphabetizing actively discards an
ordering the source file provides for free.

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

Offered values run from 5% to 30%. Every value the control offers must be one the server accepts,
or the control silently falls back to the default.

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

### P-031 — Tile contents
Subject: `presentation.tile`
Goals: G-001, G-002, G-003

Each species is one **tile**, laid out top to bottom:

1. a header whose first line carries the **common name followed by its four-letter banding code in
   parentheses**, with the **scientific name right-justified on the same line**; and below it the
   **family badge**, showing the family's scientific name and its common name together;
2. the diagram, **vertically centered** in its area;
3. the main metrics — **length, wingspan, mass** — in the selected unit system.

Tiles wrap to fill the available display width.

Rationale for centering: under global scale a small bird occupies a fraction of the tile's height,
and bottom-aligning it reads as a rendering fault rather than as the size difference it is.
Centering makes the empty space symmetric, so it reads as headroom rather than as misalignment.

Family common names are **not** unique across families — several groups share a vernacular
label — so they are a reading aid beside the scientific name, never an identifier.

### P-032 — Imperial by default, metric available
Subject: `units.default`
Value: `imperial`
Goals: G-001

One toggle switches the whole interface between:

| | length, wingspan | mass |
| --- | --- | --- |
| imperial (default) | inches | ounces, pounds above 16 oz |
| metric | centimeters | grams, kilograms above 1 kg |

**Every** displayed number honors it — tile footers, section headers and the detail view alike.

Rationale: the ABA Area audience is predominantly US; the toggle serves Canadian users and anyone
who thinks in centimeters. Ounces below one are shown to two decimals, because a 3 g hummingbird
and a 12 g kinglet both round to 0.1 oz otherwise, erasing the distinction at exactly the size
where the app is trying to make it.

### P-033 — First load shows everything
Goals: G-002

The initial view has every family selected and shows all species, ordered by length.

Rationale: the full sweep from hummingbird to swan is the clearest statement of what the app is
for. Opening filtered or empty would hide it.

### P-038 — Toggling a control keeps your place
Subject: `presentation.scroll`
Goals: G-002

Changing any control preserves the scroll position **and the collapsed state of the Instructions
and Controls panels**, as does returning from a detail view.

Rationale: every control is a link, so each toggle loads a new URL and the browser lands at the
top — which discards your place in a list of 841 birds. It matters most for the controls that do
not change the list at all: switching units or scale should leave you looking at the same birds.

The panels are the same defect seen from another angle. Their markup says open, so every reload
reopened one the reader had deliberately collapsed and pushed the page down. Panel state is
restored **before** the scroll, since collapsing one afterwards would move everything beneath it.

The restoration is the only client-side script in BirdView and holds no application logic;
ordering, sectioning and geometry stay in Go where they are property-tested (see
[arch.md](arch.md)). Without JavaScript the page behaves as it did before.

### P-037 — The empty state says what to do
Subject: `presentation.empty_state`
Goals: G-002

With nothing selected, the display area reads *"Select birds to display in Controls above"*.

Rationale: an empty page is reachable in one click, and "no species match that filter" would be
untrue and unhelpful — nothing failed to match, the user turned everything off. The message names
the control that fixes it.

### P-036 — The controls collapse
Subject: `presentation.controls`
Goals: G-002

The controls panel is collapsible and starts open, like the instructions above it. Once collapsed
it **stays** collapsed across navigations (P-038).

Rationale: the family selector is 91 chips in 25 groups, which pushes the birds well below the
fold. A reader who has chosen their filter should be able to put it away — and have it stay away.

### P-035 — The page explains itself
Subject: `presentation.help`
Goals: G-001, G-002, G-003

Below the title, the page carries help text in three sections, in this order: **the diagram**,
**where the numbers come from**, then **the controls**.

The diagram section is **a real labeled diagram**, the same one the detail view draws, rather than
a list of colour swatches: every part named and measured against an actual bird. The numbers
section explains solid against dashed, the wing chord and its ticks, and the estimate marker. The
controls section covers every control — ordering metric, scale mode, section step, units, family
filter.

The help is collapsible and starts open, so a returning user can dismiss it without losing it.

Rationale: nothing about the diagram is self-evident. A reader cannot be expected to know that the
amber stroke is the bill, that a dashed stroke means estimated, or that the crossbar is a true
wingspan rather than a decoration — and the scale toggle, the one control that changes what the
app is *saying*, is the least discoverable of all.

Showing the diagram rather than describing it is the point of the change: a swatch list asks the
reader to map six colour chips onto a drawing, where a labeled example simply *is* the drawing with
the answers on it. The order runs from the thing on screen, to what its strokes mean, to how to
change it.

### P-034 — A tile opens a detail view
Subject: `presentation.detail_view`
Goals: G-001, G-003

Clicking a tile opens a detail view that is **a larger version of the same tile** — same header,
same diagram, same main metrics, same layout — with the bird drawn **as large as the view allows,
whatever scale the index is using**, plus:

- the measurement of **every individual diagram segment**: bill, head, neck, body, tail;
- the **wing chord**, which the crossbar draws as its solid outer portions (P-065) and which would
  otherwise be the one length visible in the drawing with no row explaining it;
- the **provenance of every metric shown**, main metrics and segments alike (P-063).

Rationale: the tile is deliberately dense, because it exists to be compared against dozens of
others at a glance. That leaves no room to say where each number came from. Keeping the detail
view the *same* layout at a larger size means the user is reading a familiar thing with more of it
exposed, rather than a different screen they have to re-orient to.

The detail view does **not** inherit the index's scale. Under the global default a kinglet is a
few pixels tall, which is the honest comparison — but the one view devoted to that single bird has
nothing to compare it against, so carrying the scale there would leave nothing to look at.

---

## Diagram

### P-040 — Segment composition
Subject: `diagram.segments`
Value: `bill, head, neck, body, tail, wing crossbar`
Goals: G-003

**The diagram is a top-down view of a bird in flight.** Everything below follows from that, and it
is stated first because leaving it unstated is what made the leg question unanswerable for as long
as it was; see [CR-010](decisions.md#cr-010).

The diagram is a vertical cross. Top to bottom: a bill stick, a head oval, a neck, a body stick,
and a tail stick — plus one horizontal crossbar representing the wings. The tail is visually
distinguished from the body by stroke style, colour and a mark at the junction between them. That
remains true under P-065, but for a different reason: the tail is measured and so solid, the body
modeled and so dashed. The junction mark carries the distinction on its own if the two ever share a
stroke style again.

**The neck is a stick, except where the family folds it in flight.** Ardeidae and Pelecanidae fly
with the head drawn back onto the shoulders, and for those the neck is drawn as an **S of two
semicircular arcs**. The stroke is still exactly the neck long at the diagram's scale — the S is
the same neck, curved — so it occupies `2/π ≈ 63.7%` of that length vertically. That ratio is a
consequence of the shape and not a chosen parameter; see [CR-011](decisions.md#cr-011).

**There are no legs.** In flight most birds tuck them out of sight, and for the minority that
trail them — herons, cranes, storks, ibises, flamingos, stilts, rails, loons, grebes — what shows
is the leg's projection past the tail tip, a small residual of quantities BirdView does not have:
the femur, the toes and the hip's position are unmeasured by any source, and the feathered
tibiotarsus that straightens out in flight is exactly what a standing-posture leg model leaves out.
Legs were drawn from the measured tarsus, then from a modeled leg, and are now removed; CR-010
records the evidence and what the removal costs.

Rationale: the tail is a stick rather than a triangle because no trait in any source gives a tail
width, and a triangle would assert a fan shape the data does not support.

### P-041 — Segments abut exactly and sum to total length
Subject: `diagram.segments.sum`
Goals: G-001, G-003

The five vertical segments meet without gap or overlap, and their lengths sum to the species'
total length.

Rationale: this is what lets the vertical extent of the drawing *be* the bird's length, so that
two diagrams at the same scale can be compared directly.

**Qualification.** That equality is a property of the segment *lengths*, and it holds without
exception. It is also true of the drawn vertical extent for 827 of 841 species — but not for the
folded-neck families of P-040, whose S occupies less height than the neck is long. Their silhouette
is genuinely shorter than their catalogued length, which is a fact about a flying heron and not a
rendering error. Everything that treats length as data — ordering, sectioning, the scale reference
of P-042, and the detail view — uses the full total length regardless of how the neck is drawn, so
comparability is unaffected. See [CR-011](decisions.md#cr-011).

### P-042 — Scale is a user toggle, defaulting to per-section
Subject: `diagram.scale.mode`
Value: `per-section (default) | global`
Goals: G-001, G-002

The user switches between a single scale across the entire filtered set, and a scale shared by all
species within a section. **Per-section is the default, and is listed first.**

Whichever is chosen, the scale is set by the largest total length in the group. Since the segments
sum to total length (P-041) and nothing is drawn outside them, that is the whole vertical extent.

Rationale: this is the requirement G-001 lives or dies by. A per-tile scale — every bird drawn to
fill its tile — makes a hummingbird and a swan identical on screen and destroys size intuition
entirely.

**The default was global and is now per-section** ([CR-017](decisions.md)). Global still makes the
cross-family argument better than anything else, and it is one click away and named first in the
help. But the diagram has since acquired detail a reader has to be able to see — provenance in the
stroke style (P-065), the chord ticks, the labels of P-066 — and at global scale a median bird is a
few pixels tall, where none of it is legible. A default that hides the drawing's own vocabulary
teaches the reader nothing, and a reader who has learned to read one bird is far better placed to
be surprised by the swan.

### P-043 — The wing crossbar is drawn at the vertical scale
Subject: `diagram.wingspan.scale`
Value: `same as vertical axis`
Goals: G-001

The crossbar's length represents wingspan at the same scale as the vertical axis, so its length
relative to the body is literal.

Rationale: wingspan runs 1.0–2.5× body length, so honoring this forces a wide tile — roughly
2.6:1. That is a real layout cost, accepted because a crossbar on its own arbitrary scale tells
the user something false about proportion while looking authoritative.

### P-044 — Proportion differentiates species
Goals: G-003

A dowitcher reads as long-billed, an accipiter as long-tailed, a heron as long-necked.

### P-046 — The cranium is modeled from mass
Subject: `diagram.head_model`
Value: `isometric on mass`
Goals: G-003

Head length scales isometrically with body mass. Like every modeled value it names its function
(P-062) and claims no measurement.

Rationale: an earlier model derived the cranium from bill depth with a 12 mm floor, and the floor
bound for **442 of 841 species and 86% of birds under 200 mm**. Half the dataset drew one identical
head that varied with nothing, consumed a fixed share of every small bird's length, and
differentiated no species from another — the opposite of what a shape diagram is for. Bill depth is
in any case a poor predictor of skull length; unfloored it gave a sandpiper a 7 mm head.

Mass is the only whole-animal measurement available, and a linear dimension scaling as its cube
root is the natural default. There is deliberately **no floor**: a floor would only conceal a bad
coefficient.

Known limitation: because both cranium and trunk scale on mass, birds of similar mass get similar
heads. Within a size tier the head therefore differentiates little — the shape signal there comes
from bill, tail, neck and wingspan, which are measured or family-specific.

### P-045 — The neck/body split is a named per-family model
Subject: `diagram.neck_ratio`
Value: `per-family modeling function`
Goals: G-001, G-003, G-004

The division of the remaining body axis between neck and body uses a per-family ratio. The
resulting **neck length is presented as a metric like any other**, in the detail view, carrying a
provenance record whose `kind` is `modeled` and whose `model` **names the specific function and
the ratio applied to that bird's family** — for example, "per-family neck ratio, Ardeidae = 0.35".

The ratio table is recorded in the code with this requirement cited beside it.

Rationale: no source measures avian neck length at this coverage, and one ratio for every species
would make every bird the same shape, defeating G-003. What makes a per-family constant
acceptable is not hiding it but naming it: a user who wants to know why a heron is drawn
long-necked can read the exact function that made it so. Hiding the number while still drawing it
would be the less honest option — the reader sees the neck either way, and only the explanation
would be missing. Supersedes the condition in [CR-004](decisions.md#cr-004); see
[CR-007](decisions.md#cr-007).

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

- **Wingspan, measured** — Fox et al. 2026, filtered to `needs_prediction = 0` so only empirical
  values are taken. Stored as the reported min/mid/max triple rather than a point value.
- **Wingspan, modeled** — where no empirical value exists, a log-log regression of wingspan on
  wing chord, fitted **by the ingest** on those same empirical rows, per taxonomic order, with a
  global fit for orders holding fewer than five empirical species. See
  [CR-008](decisions.md#cr-008).
- **Length** — **modeled for every species.** No usable source of measured avian total length
  was found: Ricklefs 2017 was the only candidate and its terms prohibit the use
  ([CR-002](decisions.md#cr-002)).
- **Traits** — AVONET, read from whichever of its three taxonomy sheets carries the species as its
  own taxon ([CR-006](decisions.md#cr-006)).
- **Everything not covered above** — modeled.

**Total length is a composite.** It is the sum of the five segments (P-041), of which bill and
tail are measured from AVONET while head, neck and body are modeled. Its `kind` is therefore
`modeled`, and its `model` says so — a reader who sees a measured bill inside a modeled total
should be able to find out how both are true at once.

Rationale: AVONET records neither macroscopic dimension, and no open source supplies both across
the ABA avifauna. Mixing sources of differing authority is acceptable *only* under P-062. See
[CR-001](decisions.md#cr-001).

### P-061 — The tile marks any metric that is not measured
Goals: G-001

On the tile itself — where numbers and the scaled diagram appear — any metric whose `kind` is not
`measured` is visibly marked as such, without the tile having to explain why.

Rationale: this is what makes P-060 honest rather than merely convenient. If it is dropped, the
provenance record becomes metadata nobody sees and BirdView presents modeled numbers as data,
which is the thing the project explicitly rejected. It is load-bearing, not decoration.

### P-062 — Every presented metric carries a provenance record
Subject: `measurement.provenance.record`
Value: `{kind, source, sourceTaxon, model}`
Goals: G-001, G-004

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
- **`modeled`** — computed rather than measured. `model` names the computation, and where the
  model's accuracy has been measured it states that too, so a reader can judge how far to trust
  the number rather than only knowing that it is an estimate.

Rationale: this generalizes what CR-001 established for lengths and CR-006 for split taxa. Those
were solved twice, differently, because there was no single place to say how confident a number
is. One record covers all of it, and makes each later gap a labeling problem rather than a
redesign.

### P-065 — The diagram shows which lengths are measured and which are modeled
Subject: `diagram.provenance`
Goals: G-001, G-004

Stroke style carries provenance. A **solid** stroke is a measured length; a **dashed** stroke is a
modeled one. `near-match` draws solid: it is a real measurement, of a real bird, taken with
callipers — of a sibling taxon, which the detail view says and the drawing does not.

That gives, for every species: **bill and tail solid**, **neck and body dashed**, and the **head
drawn hollow** — an unfilled outline, because dash is meaningless on a closed shape and absence of
fill says the same thing.

**The wing crossbar is split.** Its outer portions are the measured wing chord — carpal joint to
the tip of the longest primary — and are **always solid**. Only the inner span between them is
dashed, and only when the wingspan is modeled. A tick marks where the two meet. Where no chord is
available the crossbar is drawn whole, dashed if the span is modeled.

**Both pages explain the convention**, because a visual code nobody documents is decoration.

Rationale: the diagram asserted five proportions and explained none of them. A reader saw a heron's
long neck with no cue that it is entirely invented — the `~` markers of P-061 sit on the *numbers*,
and a reader looking at the shape never meets them. For a project whose central commitment is that
a modeled value is never presented as a measurement (CR-001), the drawing being silent was the
largest remaining gap.

The crossbar split earns its complexity: the outer segment's length stays true even when the span
model is wrong, so all the model's error is attributed to the inner segment, where it belongs. See
[CR-015](decisions.md).

### P-066 — The detail diagram is labeled with each segment and its measurement
Subject: `presentation.detail_view.diagram_labels`
Goals: G-001, G-003

In the detail view, each of the five segments is named and measured **beside the part it
describes**, joined to it by a leader line. The **wing chord** is labeled too, on the crossbar
rather than in the stack, because it names a horizontal length and belongs beside one. An estimated value carries the same `~` the tile footer
uses. The provenance table below the diagram is unchanged: it remains the place where the source or
model behind each number is explained.

Labels never overlap one another. Where two segments are too close together for their labels to
sit at both midpoints, the text moves and the leader line stretches — the leader, not the text
position, is what ties a label to its segment.

Rationale: the table already lists every segment with its value, but reading it means mapping a row
onto a part of a drawing and holding the correspondence in your head. Labeling the drawing directly
removes that step, and the detail view is the one place with room to do it — the tile is
deliberately dense because it exists to be compared against dozens of others.

The **wing chord** is labeled centred over the measured outer segment it names, and the
**wingspan** just beyond the right wingtip and on the crossbar's own line, so it reads as naming
the whole bar rather than any part of it.

This is why the detail diagram is drawn wider than the tile's proportions, and why the bird is laid
out in only part of that width. The right of the box is reserved for the label column and for the
wingspan label beyond the wingtip; a wide-winged bird is drawn slightly smaller rather than running
under its own annotations.

### P-064 — The detail view links to the species on iNaturalist
Subject: `presentation.detail_view.external_link`
Goals: G-001, G-004

The detail view carries a link to the species' iNaturalist taxon page, so a reader who has just
learned how big a bird is can go and see one.

The link is shown **only where the match is confident**, which requires **both** names to agree:

- **Scientific name** — identical, or differing only in genus while sharing a species epithet, which
  is a generic reassignment one authority has made and the other has not.
- **Common name** — ABA's matches **any** English name iNaturalist holds for the taxon, not merely
  its preferred one, once case, punctuation and British spelling are normalized and any ABA
  parenthetical alternate is set aside.

A species that fails **either** test shows **no link at all** — not a search URL, not a genus page,
not a guess. Neither name alone is sufficient: the scientific name identifies the species and the
common name is the independent second opinion that stops a near miss becoming a link to the wrong
bird.

**Coverage is expected to be total.** Every in-scope species resolves under these rules, and a
species that stops resolving is a fault to investigate rather than an absence to accept — see
[CR-016](decisions.md).

BirdView **links and never copies**. No photograph is embedded, hotlinked, cached or reproduced;
the reader arrives on iNaturalist, where each image carries its own license. T-002 forbids the
embedding independently.

Rationale: BirdView will never have photographs, and a reader surprised by a size wants to see the
bird. Linking is the whole answer. The confidence gate exists because a link to the *wrong* bird is
worse than no link on a page whose purpose is being careful about what it claims — and an absent
link is legible, where a search URL that always "works" hides the failure. See
[CR-012](decisions.md).

### P-063 — The detail view explains each metric's provenance
Goals: G-001, G-004

In the detail view (P-034), **every** metric — the three main metrics and all five diagram
segments — is shown with its value and its provenance stated in plain language: naming the source,
naming the taxon actually measured for a `near-match`, and naming the model for a `modeled` value.

Rationale: the tile marks *that* a value is uncertain; this says *how*. "Wingspan estimated from
wing chord", "mass measured on Herring Gull, a species this was split from in 2024", and "neck
length from the per-family ratio for Ardeidae, 0.35 of the body axis" are very different caveats,
and a birder who cares enough to click deserves the difference.

Showing the segments individually is what makes the diagram auditable. Without it, the drawing
asserts five proportions and explains none of them.
