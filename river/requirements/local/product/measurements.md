package product

# Measurements

## lengthAndWingspanSources

**Where length and wingspan come from**

Subject: `measurement.source`
Goals: [goals.sizeIntuition](../../../goals/goals.md#sizeintuition), [goals.honestData](../../../goals/goals.md#honestdata)

- **Wingspan, measured** — Fox et al. 2026, filtered to `needs_prediction = 0` so only empirical
  values are taken. Stored as the reported min/mid/max triple rather than a point value.
- **Wingspan, modeled** — where no empirical value exists, a log-log regression of wingspan on
  wing chord, fitted **by the ingest** on those same empirical rows, per taxonomic order, with a
  global fit for orders holding fewer than five empirical species. Each fit states its own
  out-of-sample median error, by five-fold cross-validation, or by leave-one-out for a group of
  fewer than ten species, too few for five folds. See
  [decisions.fittedWingspanModel](../../decisions.md#fittedwingspanmodel). Where the proportions
  review sets a wingspan factor for the genus or family, the modeled value is multiplied by it and
  the provenance says so; a measured wingspan is never adjusted
  ([decisions.allModeledProportionsReviewed](../../decisions.md#allmodeledproportionsreviewed)).
- **Length** — **modeled for every species.** No usable source of measured avian total length
  was found: Ricklefs 2017 was the only candidate and its terms prohibit the use
  ([decisions.ricklefsProhibited](../../decisions.md#ricklefsprohibited)).
- **Traits** — AVONET, read from whichever of its three taxonomy sheets carries the species as its
  own taxon ([decisions.abaSpeciesWithoutAvonet](../../decisions.md#abaspecieswithoutavonet)).
  A near-match borrows its parent's row from the eBird sheet, except where another sheet holds a
  narrower concept of that parent: Blue-billed White-Tern borrows `AVONET1_BirdLife`'s *Gygis
  alba*, which excludes the Little White Tern
  ([decisions.whiteTernBorrows](../../decisions.md#whiteternborrows)). The same holds for a
  species' own row: European Herring Gull is read from `AVONET1_BirdLife`'s *Larus argentatus*,
  because the eBird sheet's row of that name contains every herring gull
  ([decisions.herringGullEuropeanRow](../../decisions.md#herringgulleuropeanrow)).
- **Everything not covered above** — modeled.

**Total length is a composite.** It is the sum of the five segments (segmentsSumToLength), of
which bill and tail are measured from AVONET while head, neck, and body are modeled. Its `kind` is
therefore `modeled`, and its `model` says so, so a reader can see how a measured bill sits inside a
modeled total.

Rationale: AVONET records neither dimension, and no open source supplies both across the ABA
avifauna. Mixing sources of differing authority is acceptable *only* under provenancePerMetric.
See [decisions.noOpenLengthAndWingspan](../../decisions.md#noopenlengthandwingspan).

## tileMarksModeled

**The tile marks any metric that is not measured**

Goals: [goals.honestData](../../../goals/goals.md#honestdata)

On the tile itself — where numbers and the scaled diagram appear — any metric whose `kind` is not
`measured` is visibly marked as such, without the tile having to explain why.

Rationale: this is what makes lengthAndWingspanSources honest. Without it the provenance record is
metadata nobody sees, and BirdView presents modeled numbers as data, which the project rejected.

## provenancePerMetric

**Every presented metric carries a provenance record**

Subject: `measurement.provenance.record`
Value: `{kind, source, sourceTaxon, model}`
Goals: [goals.honestData](../../../goals/goals.md#honestdata), [goals.goodDataCitizen](../../../goals/goals.md#gooddatacitizen)

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
  name such as *Charadrius nivosus* for *Anarhynchus nivosus*, is measured. `sourceTaxon` records
  the name so the claim is auditable.
- **`near-match`** — a published measurement of a **different taxon**, used because none exists
  for this species; typically a pre-split parent, such as *Larus argentatus* for *Larus vegae*.
- **`modeled`** — computed rather than measured. `model` names the computation, and where the
  model's accuracy has been measured it states that too.

**Every value a tile shows says where it came from in its tooltip**, measured values included: a
measured value names its source, and the name it is filed under where that is not the species'
own; a near-match names its donor and its source; a modeled value names its model. A reader can
then learn where a measurement came from as easily as an estimate, without opening the species
page.

Rationale: one record covers what [decisions.noOpenLengthAndWingspan](../../decisions.md#noopenlengthandwingspan) solved for lengths and
[decisions.abaSpeciesWithoutAvonet](../../decisions.md#abaspecieswithoutavonet) for split taxa, so each later gap is a labeling problem rather
than a redesign.

## everydayDefinitions

**Every measurement says what it is, in everyday words, before where it came from**

Subject: `measurement.definition`
Goals: [goals.honestData](../../../goals/goals.md#honestdata)

Each measurement BirdView names has **one plain sentence saying what it measures**, written for a
reader who knows no anatomy, and **consistent with its source's own definition**. It leads the
measurement's tooltip on a tile, before where the value came from
([provenancePerMetric](#provenancepermetric)); it leads the measurement's row in the species page's table; and the guide
lists them all. A reader asked for exactly this: "Inner" and "Chord" were unfamiliar, and the
explanations described how values were calculated "rather than what they actually represent"
([todo.plainMeasurementDefinitions](../../../todo-closed.md#plainmeasurementdefinitions)).

| Name | Definition | The source's own |
| --- | --- | --- |
| Length | From the tip of the bill to the tip of the tail. | the sum of the drawn segments ([decisions.noOpenLengthAndWingspan](../../decisions.md#noopenlengthandwingspan)) |
| Bill | From the tip of the bill to where it meets the skull. | AVONET: "Length from the tip of the beak to the base of the skull" |
| Tail | From where the tail leaves the body to the tip of its longest feather. | AVONET: "Distance between the tip of the longest rectrix and the point at which the two central rectrices protrude from the skin" |
| Wingspan | From the tip of one wing to the tip of the other, across the body, with both wings fully extended. | Fox et al.: "Wingspan, the tip-to-tip distance of fully extended wings"; the body between the wings included |
| Chord | One folded wing, from the bend of the wing to its tip. | AVONET: "Length from the carpal joint to the tip of the longest primary on the unflattened wing" |
| Inner | The middle of the wingspan, between the bends of the two wings: the body and the inner wings. | derived, the wingspan less two chords ([labeledDetailDiagram](#labeleddetaildiagram)) |
| Mass | How much the bird weighs; the cube is the water that weighs the same. | AVONET body mass |
| Head, Neck, Body | The parts of the length between the bill and the tail. | modeled ([diagram.segmentComposition](diagram.md#segmentcomposition)) |

The definitions say what a value represents, not how it was obtained; the provenance that follows
says that, and an estimate keeps its `~`.

## measuredVersusModeled

**The diagram shows which lengths are measured and which are modeled**

Subject: `diagram.provenance`
Goals: [goals.honestData](../../../goals/goals.md#honestdata)

Stroke style carries provenance. A **solid** stroke is a measured length; a **dashed** stroke is a
modeled one. `near-match` draws solid: it is a real measurement of a sibling taxon, which the
detail view says and the drawing does not.

**That holds even where two tiles come out identical** (settled 2026-10-05). Of nine near-match
species, three have their donor on screen beside them — Chihuahuan and Eastern Meadowlark, Cocos
and Brown Booby, Siberian and American Pipit — and render the same length, wingspan, and mass. The
home page adds no marker: the detail page names the donor on every affected row (*"measured on
Sturnella magna, a different species this one was split from — no measurement exists for it"*),
and a tile marker would add a mark to the six whose donor is not on screen.

What a tile must not do is say something **false** about it — labeling a near-match `estimated`
would contradict its solid stroke. A near-match names its donor in the tooltip
([provenancePerMetric](#provenancepermetric)).

For every species, then: **bill and tail solid**, **neck and body dashed**, and the **head drawn
hollow** — an unfilled outline, because dash is meaningless on a closed shape.

**The wing crossbar is split.** Its outer portions are the measured wing chord — carpal joint to
the tip of the longest primary — and are **always solid**. Only the inner span between them is
dashed, and only when the wingspan is modeled. A tick marks where the two meet. Where no chord is
available the crossbar is drawn whole, dashed if the span is modeled.

**Both pages explain the convention**, because an undocumented visual code is decoration.

Rationale: the `~` markers of tileMarksModeled sit on the numbers, so a reader looking at the shape
— a heron's long neck, entirely modeled — had no cue. A silent drawing was the largest remaining
gap in the commitment that a modeled value is never presented as a measurement
([decisions.noOpenLengthAndWingspan](../../decisions.md#noopenlengthandwingspan)). The crossbar split keeps the outer segment true even when the
span model is wrong, so the model's error falls on the inner segment. See
[decisions.drawingCarriesProvenance](../../decisions.md#drawingcarriesprovenance).

## labeledDetailDiagram

**The detail tile names each segment and its measurement, keyed to the drawing by color**

Subject: `presentation.detail_view.diagram_labels`
Goals: [goals.sizeIntuition](../../../goals/goals.md#sizeintuition), [goals.shapeNotOnlySize](../../../goals/goals.md#shapenotonlysize), [goals.honestData](../../../goals/goals.md#honestdata)

**Every metric the detail view reports appears on its tile.** The provenance table may explain
more about a number but may not be the only place it appears, so adding a metric to the table
means adding it to the tile. It is a rule rather than a list because a list let mass go undrawn.

The detail view's tile is the home page's tile, larger ([tileOpensDetail](presentation.md#tileopensdetail)), and its measurement lines
carry every segment: **length, bill, head, neck, body, and tail**, then **wingspan, inner span, and
wing chord**, named Wingspan, Inner, and Chord. An estimated value carries the same `~` as everywhere else. The drawing itself carries
one label, the mass beside its cube, as on the home page ([decisions.textSizeFloor](../../decisions.md#textsizefloor)).

**A measurement's name is set in its part's color** ([oneDiagramVocabulary](diagram.md#onediagramvocabulary)): **Bill** amber, **Tail**
white, **Inner** emerald, **Chord** orchid, and **Mass** blue. **Wingspan** names the whole
bar, not one colored part, so it is plain text, like Length.

**The inner span is the crossbar between the chord ticks**, the wingspan less two wing chords
([decisions.innerSpanDerived](../../decisions.md#innerspanderived)). It is **derived, never shown as a measurement**: it subtracts measurements from two
sources taken on different specimens and different axes, a spread wing tip to tip against a closed
wing wrist to tip, so the number measures nothing anyone measured, and subtraction enlarges their
uncertainty. The tile names it **Inner**, short enough that the wing line fits a phone; its tooltip and its
row in the measurements table, named Inner span, say how it is derived and from what. It carries the `~`, and it is left out where twice the chord reaches the wingspan, where the
drawing does not split the bar. The name says
which part it measures, so color is never the only cue. Head, neck, and body are modeled and drawn
gray, and their names are plain text.

Labels in the drawing, with leader lines to each part, were the design until 2026-11-01. Drawn in
an SVG scaled to the page, they rendered at 8px on a laptop and 3.5px on a phone, below
[tech.textSizeFloor](../tech/accessibility.md#textsizefloor), and a label column wide enough to read would leave a phone's drawing no
room.

Rationale: reading a table means mapping rows onto the drawing in your head; a measurement named
on the tile, in its part's color, maps itself. One tile on both pages is one design to keep
readable.

### Total length and mass

**Total length is the tile's Length.** It states the bird's length, not the drawing's height (below).

**Where the figure is shorter than the total length, the page says why.** A bird that flies with
its neck folded has the neck drawn as an S ([decisions.foldedNeckAsS](../../decisions.md#foldedneckass)), so the figure spans less
height than the total length labeled beneath it. Its detail page says so beside the neck note, with
the species' own neck length and total length. That the S is the same length of line in less
height, and that sorting, sections, and every measurement use the full length, stays here rather
than on the page (Sameer, 2026-10-25). A bird drawn at full extent carries no such note.

**Mass is drawn as a number, and also as water.** The **cube of water that weighs what the bird
weighs** is drawn in the **lower-left corner**, in oblique projection, at the diagram's own scale.
Its side is ∛(mass): one gram of water is one cubic centimeter, so this is a unit conversion, not a
model, and adds no second scale. It is drawn away from the figure and named in the diagram key, so
it does not read as part of the bird.

**The front face stands on the floor of the drawing, in the same place in every diagram**, and the
projection grows up and to the right, so two cubes share an origin and differ only in size. The
depth is 0.4 of the side: cabinet projection's 0.5 also clears the crossbar everywhere, but by
14px where 0.4 leaves 21, a margin no unit test sees because the fixture holds five birds.

The **Presentation panel explains the cube**, not only the diagram key, since that panel teaches
the drawing ([pageExplainsItself](presentation.md#pageexplainsitself)) and the cube is its one mark
that is not a length. Both say **what the cube is — the bird's mass drawn as water, at the same
scale as the bird — and no more.** That every cube stands on the floor of its own diagram so two
can be compared, and that a cube root flattens (a cube twice as wide is a bird eight times as
heavy), is design reasoning, recorded here and not on the page; the number under the cube gives the
mass (Sameer, 2026-10-25).

**Mass is labeled beside its cube**, its baseline on the cube's foot, on both pages.

The **home page's tiles carry the cube too** ([tileContents](presentation.md#tilecontents),
[decisions.massCubeOnTiles](../../decisions.md#masscubeontiles)), for what length and wingspan cannot show: **bulk against elongation**.
Read against the bird's own length, it is about **27%** of the drawing height for alcids, dense
divers, and **10–13%** for long-tailed light-bodied species like the Scissor-tailed Flycatcher and
the magpie (measured after
[decisions.trunkExponentCalibrated](../../decisions.md#trunkexponentcalibrated); the ratio is
scale-invariant, so it holds at any scale).

**The cube compresses, so the number is required.** A Bushtit and a Common Raven differ in mass by
**175×** and their cubes by 5.6×. The number carries the magnitude and the cube the sense of it.
See [decisions.massAsWaterCube](../../decisions.md#massaswatercube), and
[crossbarAtVerticalScale](diagram.md#crossbaratverticalscale) for why a disc or bar sized by mass —
a second, arbitrary scale — is refused.

Mass is **measured for 831 of 841 species**, so it is usually the only unmarked number on the
figure and the only solid-stroked whole-bird mark.

**Total length states the bird's length, not the drawing's height.** For the folded-neck families
of [decisions.foldedNeckAsS](../../decisions.md#foldedneckass) these differ: the neck's S spans 2/π
of its length, so the figure is 6–9% shorter than the Length across the 14 species affected. The
neck measurement follows the same rule ([decisions.foldedNeckAsS](../../decisions.md#foldedneckass)); a number taken from the drawn extent
would name a quantity no source states.

## metricProvenanceExplained

**The detail view explains each metric's provenance**

Goals: [goals.honestData](../../../goals/goals.md#honestdata), [goals.goodDataCitizen](../../../goals/goals.md#gooddatacitizen)

In the detail view (tileOpensDetail), **every** metric — the three main metrics and all five diagram
segments — is shown with its value and its provenance stated in plain language: naming the source,
naming the taxon actually measured for a `near-match`, and naming the model for a `modeled` value.

Rationale: the tile marks *that* a value is uncertain; this says *how*. "Wingspan estimated from
wing chord", "mass measured on Herring Gull, a species this was split from in 2024", and "neck
length from the per-family ratio for Ardeidae, 0.35 of the body axis" are different caveats.
Showing the segments individually is what makes the diagram auditable.

## categoryProvenance

**The detail view lists every category the bird is in, and where each came from**

Subject: `presentation.detail.categories`
Goals: [goals.honestData](../../../goals/goals.md#honestdata), [goals.goodDataCitizen](../../../goals/goals.md#gooddatacitizen)

The detail view names every category a bird is in — its family and order, the regions it is
listed in, its habitat and niche, and any group BirdView puts it in — and marks each as
**published**, **derived**, or **`birdview`**, saying where it came from.

[metricProvenanceExplained](#metricprovenanceexplained) does this for numbers; a category is as
much a claim as a measurement.

- **Published** — the source said so. AviList gives the family; AVONET gives the habitat.
- **Derived** — BirdView computed it from published data by a stated rule. A region comes from
  the 1-ppm presence rule over GBIF records; a zone comes from clustering those lists.
- **`birdview`** — this project decided it, and no source says it; for instance, a supplementary
  habitat added because AVONET records only a primary one. Labeled with the project's name rather
  than "ours", which on a public page has no clear referent.

**An unmarked category is the failure this exists to prevent.** Listing "Habitat: Wetland" beside
"Habitat: Coastal" without saying the first is AVONET's and the second is ours presents an opinion
as a finding, the error [lengthAndWingspanSources](#lengthandwingspansources) forbids for numbers.

**Where the explanation is an algorithm, link to it rather than inline it.** The rule is named in
one line and links to the fuller account in `docs/`.

**The guides are served by BirdView itself, at `/docs`**, so a provenance link works wherever the
site does, including offline ([tech.selfContainedServer](../tech/delivery.md#selfcontainedserver)).

**The requirement documents are not served**
([decisions.requirementsNotServed](../../decisions.md#requirementsnotserved)). They are working
documents, published with the repository at `github.com/google/birdview`, and a guide that cites a
requirement links to it there.

**Links are rewritten when the document is rendered, never in the file.**
`../river/requirements/local/product/` is correct in a checkout; in a browser it becomes the file
in the published repository, and a link to another guide becomes its `/docs` path. Every rewritten
repository link must name a file that exists.

**Nothing under `river/` or `third_party/` is served.** They hold the provenance of other people's
terms — the ABA checklist's, Wiley's, iNaturalist's — and the vendored iNaturalist pages kept for
hashing. Serving them would be the redistribution
[tech.fetchRestrictedInputs](../tech/delivery.md#fetchrestrictedinputs) exists to prevent.

## pipelineDiagram

**The documentation draws the data pipeline, field by field**

Subject: `presentation.docs.pipeline`
Value: `docs/pipeline.md: every source and table, its fields, and every step between them`
Goals: [goals.honestData](../../../goals/goals.md#honestdata), [goals.giveBack](../../../goals/goals.md#giveback)

The documentation carries **one diagram of how the dataset is built**: every published source and
every intermediate table as a box giving its name and the fields BirdView reads from it, and every
step between them — a filter, a join, a model — as an arrow saying what it does. It ends at the
dataset the server reads and the pages it serves.

**Each field says what becomes of it**: shown on a page, used without being shown (a join key, a
filter, a denominator), or read and then not used, so an auditor need not wonder whether it was
forgotten.

**Linked from wherever the documentation says where the data comes from**, and from the guides for
readers and for developers.

**Kept in step with the code.** A change to what the pipeline reads, joins, models, or writes
updates the diagram and its step descriptions in the same change, as
[product.selfDescriptionStaysTrue](presentation.md#selfdescriptionstaystrue) requires of all
documentation. Tests compare what they can: every source column the diagram names is one a reader
asks for, and every `internal/taxonomy` table the ingest reads is named in the diagram. What a step
*does* stays a reading task.

**One source, drawn two ways.** The diagram is Mermaid in the Markdown, which GitHub draws.
BirdView serves a picture pre-rendered from that source rather than loading Mermaid's script, which
[tech.selfContainedServer](../tech/delivery.md#selfcontainedserver) rules out; the picture records
a hash of its source, so a stale picture fails a test. It is plain SVG with no script and no link
off the server.

## borrowedMeasurementsSaySo

**A drawing made from another drawn species' measurements says so**

Subject: `presentation.shared_measurements`
Value: `a note on the tile and the detail page of each species, naming the others`
Goals: [goals.honestData](../../../goals/goals.md#honestdata)

Where two or more drawn species take their AVONET traits from **the same row** — a near-match
borrowing a relative's measurements under [provenancePerMetric](#provenancepermetric), and the
relative drawn from that same row — their drawings are alike because their measurements are, not
because the birds are. Borrowing is honest when labeled
([decisions.abaSpeciesWithoutAvonet](../../decisions.md#abaspecieswithoutavonet)), but two
identical tiles look like a finding, and the near-match label does not say the other is the source.

**Every such species says so where it is drawn.** Its tile carries the line *Same measurements
as* the other species, styled as the tile's other caveats, and its detail page says whose
measurements were used and that the other species' drawing comes from the same row.

**The groups come from the dataset, not from a table.** Two species share measurements when their
traits name the same source sheet and source taxon, so a new near-match gets its note
automatically.

Rationale: [goals.honestData](../../../goals/goals.md#honestdata) — otherwise the reason two
drawings coincide is visible only by comparing two detail pages.
[decisions.vegaGullBorrows](../../decisions.md#vegagullborrows).

## escapedBirdsListed

**Escaped and kept birds are left off the home page, and listed**

Subject: `presentation.index.exotic`
Value: `declared list; off the home page, listed at /excluded, detail pages kept`
Goals: [goals.localRelevance](../../../goals/goals.md#localrelevance), [goals.honestData](../../../goals/goals.md#honestdata)

**A bird whose records here are escapes is not drawn on the home page.** Recurrence cannot tell a
wild population from a supply of escapes ([declaredExceptions](dataset.md#declaredexceptions)), so
parrots, waxbills, ornamental waterfowl, zoo birds, and a farmed Emu crowded out the birds a reader
can expect to meet.

**Which birds are left off is a declared list, with a reason for each** — the kind of bird, where
it comes from, and how its records here arise. The rule: no self-sustaining population in any
region BirdView covers, so that its records are escapes, releases, or birds kept by people. A
long-feral population, such as Indian Peafowl in Florida or Red Junglefowl on Kauai, does not
qualify; nor does a wild vagrant, however rare.

**Left off, not deleted.** Every listed bird keeps its detail page, which says it is not on the
home page and why. `/excluded` lists them all, each linked to that page, with its reason and its
records; the home page links to the list. A bird listed here appears in no family, habitat, diet or
region control, and is not counted by them.

Rationale: [goals.honestData](../../../goals/goals.md#honestdata) — removing a bird a reader may
have seen, with no trace, is what the goals forbid; a list that says why is the honest version.
[decisions.escapedBirdsOffIndex](../../decisions.md#escapedbirdsoffindex).
