package product

# Measurements

## lengthAndWingspanSources

**Where length and wingspan come from**

Subject: `measurement.source`
Goals: goals.sizeIntuition

- **Wingspan, measured** — Fox et al. 2026, filtered to `needs_prediction = 0` so only empirical
  values are taken. Stored as the reported min/mid/max triple rather than a point value.
- **Wingspan, modeled** — where no empirical value exists, a log-log regression of wingspan on
  wing chord, fitted **by the ingest** on those same empirical rows, per taxonomic order, with a
  global fit for orders holding fewer than five empirical species. See
  [decisions.fittedWingspanModel](../../decisions.md#fittedwingspanmodel).
- **Length** — **modeled for every species.** No usable source of measured avian total length
  was found: Ricklefs 2017 was the only candidate and its terms prohibit the use
  ([decisions.ricklefsProhibited](../../decisions.md#ricklefsprohibited)).
- **Traits** — AVONET, read from whichever of its three taxonomy sheets carries the species as its
  own taxon ([decisions.abaSpeciesWithoutAvonet](../../decisions.md#abaspecieswithoutavonet)). A near-match borrows its parent's row from the eBird
  sheet, except where another sheet holds a narrower concept of that parent: Blue-billed
  White-Tern borrows `AVONET1_BirdLife`'s *Gygis alba*, which excludes the Little White Tern
  ([decisions.whiteTernBorrows](../../decisions.md#whiteternborrows)). The same holds for a species' own row: European Herring Gull
  is read from `AVONET1_BirdLife`'s *Larus argentatus*, because the eBird sheet's row of that
  name contains every herring gull ([decisions.herringGullEuropeanRow](../../decisions.md#herringgulleuropeanrow)).
- **Everything not covered above** — modeled.

**Total length is a composite.** It is the sum of the five segments (segmentsSumToLength), of which bill and
tail are measured from AVONET while head, neck and body are modeled. Its `kind` is therefore
`modeled`, and its `model` says so — a reader who sees a measured bill inside a modeled total
should be able to find out how both are true at once.

Rationale: AVONET records neither macroscopic dimension, and no open source supplies both across
the ABA avifauna. Mixing sources of differing authority is acceptable *only* under provenancePerMetric. See
[decisions.noOpenLengthAndWingspan](../../decisions.md#noopenlengthandwingspan).

## tileMarksModeled

**The tile marks any metric that is not measured**

Goals: goals.sizeIntuition

On the tile itself — where numbers and the scaled diagram appear — any metric whose `kind` is not
`measured` is visibly marked as such, without the tile having to explain why.

Rationale: this is what makes lengthAndWingspanSources honest rather than merely convenient. If it is dropped, the
provenance record becomes metadata nobody sees and BirdView presents modeled numbers as data,
which is the thing the project explicitly rejected. It is load-bearing, not decoration.

## provenancePerMetric

**Every presented metric carries a provenance record**

Subject: `measurement.provenance.record`
Value: `{kind, source, sourceTaxon, model}`
Goals: goals.sizeIntuition, goals.goodDataCitizen

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

Rationale: this generalizes what decisions.noOpenLengthAndWingspan established for lengths and decisions.abaSpeciesWithoutAvonet for split taxa. Those
were solved twice, differently, because there was no single place to say how confident a number
is. One record covers all of it, and makes each later gap a labeling problem rather than a
redesign.

## measuredVersusModeled

**The diagram shows which lengths are measured and which are modeled**

Subject: `diagram.provenance`
Goals: goals.sizeIntuition, goals.goodDataCitizen

Stroke style carries provenance. A **solid** stroke is a measured length; a **dashed** stroke is a
modeled one. `near-match` draws solid: it is a real measurement, of a real bird, taken with
callipers — of a sibling taxon, which the detail view says and the drawing does not.

**That holds even where two tiles come out identical, which was asked and settled 2026-10-05.**
Nine species carry a near-match value and three have their donor on screen beside them, so
Chihuahuan Meadowlark and Eastern Meadowlark, Cocos Booby and Brown Booby, Siberian Pipit and
American Pipit each render the same length, wingspan and mass. A birder who knows a Chihuahuan
Meadowlark is the smaller bird will notice. The index still says nothing: the detail page names
the donor on every affected row — *"measured on Sturnella magna, a different species this one
was split from — no measurement exists for it"* — and a marker on the tile would explain an
oddity to the three readers who can see it while adding a mark to six who cannot.

What a tile must not do is say something **false** about it. All three footer cells once
labelled any unmeasured value `estimated`, which contradicted the solid stroke beside it; a
near-match now names its donor in the tooltip ([provenancePerMetric](#provenancepermetric)).

That gives, for every species: **bill and tail solid**, **neck and body dashed**, and the **head
drawn hollow** — an unfilled outline, because dash is meaningless on a closed shape and absence of
fill says the same thing.

**The wing crossbar is split.** Its outer portions are the measured wing chord — carpal joint to
the tip of the longest primary — and are **always solid**. Only the inner span between them is
dashed, and only when the wingspan is modeled. A tick marks where the two meet. Where no chord is
available the crossbar is drawn whole, dashed if the span is modeled.

**Both pages explain the convention**, because a visual code nobody documents is decoration.

Rationale: the diagram asserted five proportions and explained none of them. A reader saw a heron's
long neck with no cue that it is entirely invented — the `~` markers of tileMarksModeled sit on the *numbers*,
and a reader looking at the shape never meets them. For a project whose central commitment is that
a modeled value is never presented as a measurement (decisions.noOpenLengthAndWingspan), the drawing being silent was the
largest remaining gap.

The crossbar split earns its complexity: the outer segment's length stays true even when the span
model is wrong, so all the model's error is attributed to the inner segment, where it belongs. See
[decisions.drawingCarriesProvenance](../../decisions.md#drawingcarriesprovenance).

## labeledDetailDiagram

**The detail diagram is labeled with each segment and its measurement**

Subject: `presentation.detail_view.diagram_labels`
Goals: goals.sizeIntuition, goals.shapeNotOnlySize

**Every metric the detail view reports appears in the drawing.** The provenance table below it may
explain more about a number, but it may not be the only place a number appears: a reader looking
at the diagram should not have to discover that something was measured and left out of it. Adding
a metric to the table therefore means adding it to the diagram. This is stated as a rule rather
than a list because mass was reported and undrawn for days, and every check at the time named the
metrics it expected — mass among them nowhere.

In the detail view, each of the five segments is named and measured **beside the part it
describes**, joined to it by a leader line. The **wing chord** is labeled too, on the crossbar
rather than in the stack, because it names a horizontal length and belongs beside one. The
**total length** and the **mass** are labeled as well — see below. An estimated value carries the same `~` the tile footer
uses. The provenance table below the diagram is unchanged: it remains the place where the source or
model behind each number is explained.

The `~` must be in the **markup**. Writing it as CSS generated content is the obvious way and does
not work: `::after` generates no box on an SVG `<text>`, so the rule is silently inert. This is not
a hypothetical — it is how the marker was built in `a826b96`, the commit that introduced these
labels, and it shipped drawing every modeled value identically to a measured one. See
[acceptance.segmentLabelsDontCollide](../../acceptance.md#segmentlabelsdontcollide).

Labels never overlap one another. Where two segments are too close together for their labels to
sit at both midpoints, the text moves and the leader line stretches — the leader, not the text
position, is what ties a label to its segment.

**Text draws over lines.** Every line in the diagram is emitted before every label, so a leader
that crosses a label passes behind the words rather than through them.

Rationale: the table already lists every segment with its value, but reading it means mapping a row
onto a part of a drawing and holding the correspondence in your head. Labeling the drawing directly
removes that step, and the detail view is the one place with room to do it — the tile is
deliberately dense because it exists to be compared against dozens of others.

The **wing chord** is labeled centred over the **left** outer segment — both outer segments are the
measured chord, and the left keeps the label clear of the wingspan. The **wingspan** sits just
beyond the right wingtip and on the crossbar's own line, so it reads as naming the whole bar rather
than any part of it.

This is why the detail diagram is drawn wider than the tile's proportions, and why the bird is laid
out in only part of that width. The right of the box is reserved for the label column and for the
wingspan label beyond the wingtip; a wide-winged bird is drawn slightly smaller rather than running
under its own annotations.

### The whole-bird labels: total length and mass

**Total length and mass are labeled horizontally, centred on the axis, below the tail tip.** They
are the labels that name the whole bird rather than a part of it, and their placement says so:
under the foot of the figure, on the centreline, spanning nothing. They take no leader line —
there is no part to point at, and proximity to the end of the drawing is the whole of their
meaning.

**Mass is drawn as a number, and also as water.** The **cube of water that weighs what the bird
weighs** is drawn in the **lower-left corner**, in oblique projection, at the diagram's own
scale. Its side is ∛(mass), which is a unit conversion and not a model — one gram of water is
one cubic centimetre by definition — so it assumes nothing about the density of birds and
introduces no second scale. It is drawn away from the figure, and named in the diagram key,
because an unexplained shape beside a bird would read as part of the bird.

**The front face stands on the floor of the drawing, in the same place in every diagram**, and
the projection grows up and to the right from there. That fixed corner is what makes two cubes
comparable at a glance: only their size differs between birds, so the eye has a common origin to
measure from. A base that moved with the bird would cost that without changing a single number.
The depth is 0.4 of the side; cabinet projection's conventional 0.5 also clears the crossbar
everywhere, but leaves 14px where this leaves 21, and no unit test can see that margin because
the fixture holds five birds.

The **Presentation panel explains the cube**, not only the diagram key: that panel is where a
reader is taught to read the drawing ([pageExplainsItself](presentation.md#pageexplainsitself)), and the cube is the one mark on the
figure that is not a length. It says what the cube is, that every cube stands on the floor of its
own diagram so two can be held against each other, and that a cube root flattens — a cube twice
as wide is a bird eight times as heavy.

**Mass is labeled under its cube, not in the stack under the tail**, so the number sits with the
shape it describes. On a small cube the label is wider than the face and is held clear of the
left edge rather than centred past it.

The **index tiles carry the cube too**, unlabelled ([tileContents](presentation.md#tilecontents), decisions.massCubeOnTiles). It is there for
what the rest of the figure cannot say: **bulk against elongation**. Read against
the bird's own length it is about **27%** of the drawing height for alcids, which are dense
divers, and **10–13%** for long-tailed light-bodied species like the Scissor-tailed Flycatcher
and the magpie. Length and wingspan cannot tell those apart.

*(Re-measured after [decisions.trunkExponentCalibrated](../../decisions.md#trunkexponentcalibrated), which lengthened heavy birds and so shrank the
cube's share of them; the alcid figure was 28–29%. The ratio is scale-invariant, so it is the same
under either mode of scaleToggle.)*

**The cube compresses, and the number is there because it does.** A Bushtit and a Common Raven
differ in mass by **175×** and their cubes differ by 5.6×, which is correct for equal-volume cubes
and understates the gap badly. So the number carries the magnitude and the square carries the
sense of it; a reader who takes the squares alone as the mass comparison will read them wrong.
See [decisions.massAsWaterCube](../../decisions.md#massaswatercube), which reverses an earlier decision to draw no shape at all, and
[crossbarAtVerticalScale](diagram.md#crossbaratverticalscale) for why a disc or bar sized by mass — a second, arbitrary scale — is still
refused.

Mass also behaves differently from its neighbours on provenance, and the drawing shows it:
**measured for 831 of 841 species**, against a total length modeled for every one. It is usually
the only unmarked number on the figure, and the only solid-stroked whole-bird mark.

It belongs on the diagram for the reason the segments do, only more so. Total length is the number
BirdView exists to give a reader intuition about (goals.sizeIntuition); every other label names a component of
it. Leaving it only in the table below made the drawing annotate its own parts while staying silent
about the thing it is a drawing *of*.

A **bottom strip is added** for it. Without one the label is clipped: the detail view scales the
bird to fill the box, so a bird whose height binds has its tail tip exactly on the bottom edge and
nothing below it.

The strip **extends the box downward rather than shrinking the bird into it** — the opposite of
what the label gutter does horizontally, and deliberately so. Width is fixed by the page column, so
the gutter has to come out of the bird. Height is not: the drawing can simply be taller, and the
bird keeps the size [tileOpensDetail](presentation.md#tileopensdetail) requires. Taking it out of the bird instead would cost 8.5% of
the detail view's height to buy a one-line label, which is a bad trade and would put this
requirement in conflict with tileOpensDetail rather than beside it.

**Total length states the bird's length, not the drawing's height.** For the folded-neck families of
[decisions.foldedNeckAsS](../../decisions.md#foldedneckass) these differ: the neck's S spans 2/π of its length, so the figure is
6–9% shorter than the labeled number across the 14 species affected. That is the rule the neck
label already follows — decisions.foldedNeckAsS reports neck length in full while drawing it compressed — and the
total length label inherits it rather than introducing it. A label derived from the drawn extent
would be a different and worse thing: it would name a quantity no source states and no table lists.

## metricProvenanceExplained

**The detail view explains each metric's provenance**

Goals: goals.sizeIntuition, goals.goodDataCitizen

In the detail view (tileOpensDetail), **every** metric — the three main metrics and all five diagram
segments — is shown with its value and its provenance stated in plain language: naming the source,
naming the taxon actually measured for a `near-match`, and naming the model for a `modeled` value.

Rationale: the tile marks *that* a value is uncertain; this says *how*. "Wingspan estimated from
wing chord", "mass measured on Herring Gull, a species this was split from in 2024", and "neck
length from the per-family ratio for Ardeidae, 0.35 of the body axis" are very different caveats,
and a birder who cares enough to click deserves the difference.

Showing the segments individually is what makes the diagram auditable. Without it, the drawing
asserts five proportions and explains none of them.

## categoryProvenance

**The detail view lists every category the bird is in, and where each came from**

Subject: `presentation.detail.categories`
Goals: goals.sizeIntuition, goals.goodDataCitizen

The detail view names every category a bird is in — its family and order, the regions it is
listed in, its habitat and niche, and any group BirdView puts it in — and marks each as
**published**, **derived** or **`birdview`**, saying where it came from.

[metricProvenanceExplained](#metricprovenanceexplained) does this for numbers. A bird is also **sorted**, and the sorting is as much a
claim as a measurement is: its family and order, the regions it is listed in, its habitat and
its niche, and any group BirdView puts it in. Each is a statement about the bird that someone or
something decided, and the detail view **names every one of them and says where it came from**.

Three provenances, and the UI distinguishes them:

- **Published** — the source said so. AviList gives the family; AVONET gives the habitat.
- **Derived** — BirdView computed it from published data by a stated rule. A region comes from
  the 1-ppm presence rule over GBIF records; a zone comes from clustering those lists.
- **`birdview`** — this project decided it, and no source says it. Labelled with the
  project's name rather than "ours", which on a public page has no clear referent. A supplementary habitat added because
  AVONET records only a primary one, for instance.

**An unmarked category is the failure this exists to prevent.** A page that lists "Habitat:
Wetland" beside "Habitat: Coastal" without saying the first is AVONET's and the second is ours
presents an opinion as a finding, which is the same error [lengthAndWingspanSources](#lengthandwingspansources) forbids for numbers.

**Where the explanation is an algorithm, link to it rather than inline it.** "Clustered on
Jaccard similarity of species lists, cut at k=26, each remaining region assigned to its nearest
seed" does not belong in a table cell. The rule is named in one line and links to the fuller
account in `docs/`.

**The guides are served by BirdView itself, at `/docs`.** A reader following a provenance link
lands on an explanation that works wherever the site does, including offline, which is the
point of [tech.selfContainedServer](../tech/delivery.md#selfcontainedserver)'s laptop in a cabin.

**The requirement documents are not served** ([decisions.requirementsNotServed](../../decisions.md#requirementsnotserved)). They are working
documents — the decision log, the open-work list, the method itself — and they are published
with the repository at `github.com/google/birdview`, which is where someone who wants them
looks. A guide that cites a requirement links to it there.

**Links are rewritten when the document is rendered, never in the file.** `../river/requirements/sources/product/`
is the correct link in a repository checkout; in a browser it becomes the file in the published
repository, and a link to another guide becomes its `/docs` path. Editing the file to suit a
browser would break it for the checkout, and these documents are served precisely because they
are the files the repository carries. Every rewritten repository link must name a file that
exists, or it is a 404 waiting for publication.

**Nothing under `river/` or `third_party/` is served.** Between them they hold the provenance
of other people's terms — the ABA checklist's, Wiley's, iNaturalist's — and the vendored
iNaturalist pages themselves, kept so a provenance record can be hashed and checked. They are
quoted for a record, not ours to republish, and serving them would be the redistribution
[tech.fetchRestrictedInputs](../tech/delivery.md#fetchrestrictedinputs) exists to prevent.

## pipelineDiagram

**The documentation draws the data pipeline, field by field**

Subject: `presentation.docs.pipeline`
Value: `docs/pipeline.md: every source and table, its fields, and every step between them`
Goals: goals.goodDataCitizen, goals.dataAsItIs

The documentation carries **one diagram of how the dataset is built**: every published source and
every intermediate table as a box giving its name and the fields BirdView reads from it, and every
step between them — a filter, a join, a model — as an arrow saying what it does. It ends at the
dataset the server reads and the pages it serves.

**Each field says what becomes of it**: shown on a page, used without being shown (a join key, a
filter, a denominator), or read and then not used. The last is listed because a reader auditing
the pipeline would otherwise wonder whether it was forgotten.

**Linked from wherever the documentation says where the data comes from**, and from the guide for
readers and the guide for developers, so neither audience has to find it by browsing the index.

**Kept in step with the code.** A change to what the pipeline reads, joins, models or writes
updates the diagram and the step descriptions under it in the same change, as
[product.selfDescriptionStaysTrue](presentation.md#selfdescriptionstaystrue) requires of all documentation. The parts a test can compare are compared:
every source column the diagram names is one a reader asks for, and every `internal/taxonomy`
table the ingest reads is named in the diagram. What a step *does* stays a reading task.

**One source, drawn two ways.** The diagram is Mermaid in the Markdown, which GitHub draws
directly. BirdView serves a picture pre-rendered from that same source rather than loading
Mermaid's script, which [tech.selfContainedServer](../tech/delivery.md#selfcontainedserver) rules out; the picture records a hash of the source
it was drawn from, so a stale picture is a failing test rather than a quiet disagreement. It is
plain SVG with no script and no link off the server.

## borrowedMeasurementsSaySo

**A drawing made from another drawn species' measurements says so**

Subject: `presentation.shared_measurements`
Value: `a note on the tile and the detail page of each species, naming the others`
Goals: goals.shapeNotOnlySize, goals.nothingDisappearsQuietly

Where two or more drawn species take their AVONET traits from **the same row** — a near-match
borrowing a relative's measurements under [provenancePerMetric](#provenancepermetric), and the relative drawn from that same
row — their drawings are alike because their measurements are the same, not because the birds are.
That is acceptable: borrowing is honest when it is labeled ([decisions.abaSpeciesWithoutAvonet](../../decisions.md#abaspecieswithoutavonet)). But
two identical tiles look like a finding about the birds, and the near-match label on one of them
does not tell a reader that the other is the source.

**Every such species says so where it is drawn.** Its tile carries the line *Same measurements
as* the other species, styled as the tile's other caveats, and its detail page says whose
measurements were used and that the other species' drawing comes from the same row.

**The groups come from the dataset, not from a table.** Two species share measurements when their
traits name the same source sheet and source taxon. A new near-match then gets its note without
anyone remembering to add it.

Rationale: [goals.nothingDisappearsQuietly](../../../goals/goals.md#nothingdisappearsquietly) — a borrowed value is already labeled, but the reason two
drawings coincide was visible only to someone who opened both detail pages and compared the
provenance rows. [decisions.vegaGullBorrows](../../decisions.md#vegagullborrows).

## escapedBirdsListed

**Escaped and kept birds are left off the index, and listed**

Subject: `presentation.index.exotic`
Value: `declared list; off the index, listed at /excluded, detail pages kept`
Goals: goals.comparisonAcrossFamilies, goals.nothingDisappearsQuietly

**A bird whose records here are escapes is not drawn on the index.** Recurrence cannot tell a
wild population from a supply of escapes ([declaredExceptions](dataset.md#declaredexceptions)), so parrots, waxbills, ornamental
waterfowl, zoo birds and a farmed Emu were drawn among the birds a reader can expect to meet,
and they crowd out the ones that answer the question BirdView is for.

**Which birds are left off is a declared list, with a reason for each** — the kind of bird, where
it comes from, and how its records here arise. The rule: no self-sustaining population in any
region BirdView covers, so that its records are escapes, releases or birds kept by people. A
long-feral population, such as Indian Peafowl in Florida or Red Junglefowl on Kauai, does not
qualify; nor does a wild vagrant, however rare.

**Left off, not deleted.** Every listed bird keeps its detail page, which says it is not on the
index and why. `/excluded` lists them all, each linked to that page, with its reason and its
records; the index links to the list. A bird listed here appears in no family, habitat, diet or
region control, and is not counted by them.

Rationale: [goals.nothingDisappearsQuietly](../../../goals/goals.md#nothingdisappearsquietly) — removing a bird a reader may have seen, with no trace, is
the failure the goals forbid; a list that says why is the honest version of the same edit.
[decisions.escapedBirdsOffIndex](../../decisions.md#escapedbirdsoffindex).
