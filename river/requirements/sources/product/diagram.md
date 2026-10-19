package product

# Diagram

## segmentComposition

**Segment composition**

Subject: `diagram.segments`
Value: `bill, head, neck, body, tail, wing crossbar`
Goals: goals.shapeNotOnlySize

The diagram is a top-down view of a bird in flight, drawn as a vertical cross: top to bottom a bill
stick, a head oval, a neck, a body stick and a tail stick, with one horizontal crossbar for the
wings. The tail is visibly distinguished from the body; the neck is an S where the family folds it
in flight; there are no legs.

**The diagram is a top-down view of a bird in flight.** Everything below follows from that, and it
is stated because leaving it unstated is what made the leg question unanswerable for as long as it
was; see [decisions.inFlightNoLegs](../../decisions.md#inflightnolegs).

**Readers are told it is a stick figure.** Text BirdView serves to readers calls the drawing a
*stick figure*, usually *drawn to scale* where it is introduced, and *the figure* or *the drawing*
after that. *Diagram* stays the word in this spec and in the code, and on the site it means only
the data-pipeline diagram ([pipelineDiagram](measurements.md#pipelinediagram)); see [decisions.stickFigure](../../decisions.md#stickfigure).

The diagram is a vertical cross. Top to bottom: a bill stick, a head oval, a neck, a body stick,
and a tail stick — plus one horizontal crossbar representing the wings. The tail is visually
distinguished from the body by stroke style, colour and a mark at the junction between them. That
remains true under measuredVersusModeled, but for a different reason: the tail is measured and so solid, the body
modeled and so dashed. The junction mark carries the distinction on its own if the two ever share a
stroke style again.

**The neck is a stick, except where the family folds it in flight.** Ardeidae and Pelecanidae fly
with the head drawn back onto the shoulders, and for those the neck is drawn as an **S of two
semicircular arcs**. The stroke is still exactly the neck long at the diagram's scale — the S is
the same neck, curved — so it occupies `2/π ≈ 63.7%` of that length vertically. That ratio is a
consequence of the shape and not a chosen parameter; see [decisions.foldedNeckAsS](../../decisions.md#foldedneckass).

**There are no legs.** In flight most birds tuck them out of sight, and for the minority that
trail them — herons, cranes, storks, ibises, flamingos, stilts, rails, loons, grebes — what shows
is the leg's projection past the tail tip, a small residual of quantities BirdView does not have:
the femur, the toes and the hip's position are unmeasured by any source, and the feathered
tibiotarsus that straightens out in flight is exactly what a standing-posture leg model leaves out.
Legs were drawn from the measured tarsus, then from a modeled leg, and are now removed; decisions.inFlightNoLegs
records the evidence and what the removal costs.

Rationale: the tail is a stick rather than a triangle because no trait in any source gives a tail
width, and a triangle would assert a fan shape the data does not support.

## segmentsSumToLength

**Segments abut exactly and sum to total length**

Subject: `diagram.segments.sum`
Goals: goals.sizeIntuition, goals.shapeNotOnlySize

The five vertical segments meet without gap or overlap, and their lengths sum to the species'
total length.

Rationale: this is what lets the vertical extent of the drawing *be* the bird's length, so that
two diagrams at the same scale can be compared directly.

**Qualification.** That equality is a property of the segment *lengths*, and it holds without
exception. It is also true of the drawn vertical extent for 827 of 841 species — but not for the
folded-neck families of segmentComposition, whose S occupies less height than the neck is long. Their silhouette
is genuinely shorter than their catalogued length, which is a fact about a flying heron and not a
rendering error. Everything that treats length as data — ordering, sectioning, the scale reference
of scaleToggle, and the detail view — uses the full total length regardless of how the neck is drawn, so
comparability is unaffected. See [decisions.foldedNeckAsS](../../decisions.md#foldedneckass).

## scaleToggle

**Scale is a user toggle, defaulting to per-section**

Subject: `diagram.scale.mode`
Value: `per-section (default) | global; global is the default under Frequency, Bill and Tail`
Goals: goals.sizeIntuition, goals.comparisonAcrossFamilies

The user switches between a single scale across the entire filtered set, and a scale shared by all
species within a section. **Per-section is the default, and is listed first.**

**Except under Frequency, where global is the default** ([decisions.frequencyGlobalScale](../../decisions.md#frequencyglobalscale)).
Per-section scaling is justified by [sectionBreakRule](ordering.md#sectionbreakrule) guaranteeing a section is at most N% wide,
and a frequency tier has no such bound — it is banded by how often a bird is recorded, which
is uncorrelated with size, so one tier runs from 72 mm to 1,201 mm. Each tier would be drawn
against its own largest bird, and across the nine tiers that reference swings by 0.15× to
1.36×: the single 140 mm bird in the last tier would be drawn the height of the 1,201 mm
Trumpeter Swan six tiers above it. Both choices stay one click away under either ordering.

**Under Bill and Tail global is the default too, for the same reason.** A section bounded in bill or
tail length is not bounded in total length: with 15% sections, one bill section runs from the
109 mm Green-breasted Mango to the 769 mm Golden Pheasant and one tail section from the 109 mm
Black-capped Gnatcatcher to the 774 mm Common Loon, 7.1× in each, where length sections stay within
1.1× and mass sections within 3.1×.

Whichever is chosen, the scale is set by the largest total length in the group. Since the segments
sum to total length (segmentsSumToLength) and nothing is drawn outside them, that is the whole vertical extent.

Rationale: this is the requirement goals.sizeIntuition lives or dies by. A per-tile scale — every bird drawn to
fill its tile — makes a hummingbird and a swan identical on screen and destroys size intuition
entirely.

**The default was global and is now per-section** ([decisions.perSectionDefault](../../decisions.md#persectiondefault)). Global still makes the
cross-family argument better than anything else, and it is one click away and named first in the
help. But the diagram has since acquired detail a reader has to be able to see — provenance in the
stroke style (measuredVersusModeled), the chord ticks, the labels of labeledDetailDiagram — and at global scale a median bird is a
few pixels tall, where none of it is legible. A default that hides the drawing's own vocabulary
teaches the reader nothing, and a reader who has learned to read one bird is far better placed to
be surprised by the swan.

## crossbarAtVerticalScale

**The wing crossbar is drawn at the vertical scale**

Subject: `diagram.wingspan.scale`
Value: `same as vertical axis`
Goals: goals.sizeIntuition

The crossbar's length represents wingspan at the same scale as the vertical axis, so its length
relative to the body is literal.

Rationale: wingspan runs 1.0–2.5× body length, so honoring this forces a wide tile — roughly
2.6:1. That is a real layout cost, accepted because a crossbar on its own arbitrary scale tells
the user something false about proportion while looking authoritative.

## proportionDifferentiates

**Proportion differentiates species**

Goals: goals.shapeNotOnlySize

A species' drawn proportions distinguish it by shape, not only by size: a dowitcher reads as
long-billed, an accipiter as long-tailed, a heron as long-necked.

## craniumFromMass

**The cranium is modeled from mass**

Subject: `diagram.head_model`
Value: `isometric on mass`
Goals: goals.shapeNotOnlySize

Head length scales isometrically with body mass. Like every modeled value it names its function
(provenancePerMetric) and claims no measurement.

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

**The cube root is deliberately left alone by [decisions.trunkExponentCalibrated](../../decisions.md#trunkexponentcalibrated)**, which raised the
*trunk's* exponent on the evidence of measured skeletons. The same evidence supports isometry for
a cranium: single skeletal elements scale at 0.314–0.325, where chains of them scale higher. The
head is one element and keeps 1/3; the trunk is a chain and does not. See [trunkFromMass](#trunkfrommass).

Still open, and recorded because it is the one place the head model is visibly wrong: an Ostrich
draws a 360 mm cranium against a real skull of roughly 200 mm, which implies bird heads scale with
*negative* allometry rather than isometry. No head dataset exists at this coverage to fit against
— the one large skull collection is passerines-only — so the coefficient stands unvalidated. It
costs nothing inside the ABA scope, where the heaviest bird is 11 kg rather than 111 kg.

## trunkFromMass

**The trunk is modeled allometrically from mass**

Subject: `diagram.trunk_model`
Value: `14.0329 x mass^0.37`
Goals: goals.sizeIntuition, goals.shapeNotOnlySize

Trunk length scales with body mass by a power law. Like every modeled value it names its function
(provenancePerMetric) and claims no measurement.

The **exponent is calibrated against measured skeletons** rather than chosen. It must lie inside
the 95% confidence interval that published skeletal data give for a *multi-element* linear
dimension ([`bohmer-neck`](../../requirements.md#bohmer-neck)) — **0.338 to 0.401**, the intersection of the
intervals for neck length and total leg length. The distinction matters: single bones scale near
isometry, chains of them scale higher, and a trunk is a chain.

The **coefficient is a consequence, not a second free parameter.** It is set so the model is
unchanged at **63.4 g**, the median mass of the ABA dataset when the exponent was calibrated.
Recalibration is therefore a change of *proportion* across the size range rather than of overall
scale: birds near the middle are drawn as they were, small birds slightly shorter, large birds
substantially longer.

That anchor is a **fixed constant, not a value read from the dataset at build time.** If scope
changes the median moves, and the drawing must not move with it silently; re-anchoring would be a
decision, recorded as one.

Rationale: three of the five segments are driven by mass, so this exponent governs how much longer
a big bird is drawn than a small one — which is the one thing goals.sizeIntuition exists to convey. The original
0.31 was an unvalidated constant and it was **too shallow**, understating the large end of the
current scope by about 1.65×. See [decisions.trunkExponentCalibrated](../../decisions.md#trunkexponentcalibrated) for the measurement, and for what
it does not establish.

## neckModelPerFamily

**The neck/body split is a named per-family model**

Subject: `diagram.neck_ratio`
Value: `per-family modeling function`
Goals: goals.sizeIntuition, goals.shapeNotOnlySize, goals.goodDataCitizen

The division of the remaining body axis between neck and body uses a per-family ratio. The
resulting **neck length is presented as a metric like any other**, in the detail view, carrying a
provenance record whose `kind` is `modeled` and whose `model` **names the specific function and
the ratio applied to that bird's family** — for example, "per-family neck ratio, Ardeidae = 0.35".

The ratio table is recorded in the code with this requirement cited beside it.

Rationale: no source measures avian neck length at this coverage, and one ratio for every species
would make every bird the same shape, defeating goals.shapeNotOnlySize. What makes a per-family constant
acceptable is not hiding it but naming it: a user who wants to know why a heron is drawn
long-necked can read the exact function that made it so. Hiding the number while still drawing it
would be the less honest option — the reader sees the neck either way, and only the explanation
would be missing. Supersedes the condition in [decisions.neckSplitUnmeasured](../../decisions.md#necksplitunmeasured); see
[decisions.neckLengthAsMetric](../../decisions.md#necklengthasmetric).

## oneDiagramVocabulary

**One diagram vocabulary, on every surface that draws one**

Subject: `diagram.vocabulary.shared`
Goals: goals.sizeIntuition, goals.comparisonAcrossFamilies

Four surfaces draw the diagram: the **favicon**, the **index** (its tiles and its legend), the
**detail view**, and the **comparison view**. A reader meets the same figure on all four, so the
figure means the same thing on all four.

**Colour is the shared part, and it is shared exactly.** Each part of the diagram — bill, head,
neck, body, tail, wing crossbar, mass cube — has one colour, and every surface uses that colour
for that part. A surface may not pick its own shade of a part, however good its local reason.

**Size-dependent geometry is not shared, and is not expected to be.** Stroke weight, dash pattern
and which parts appear at all may be tuned per surface, because a 16px icon and a 1000px detail
diagram are not the same drawing problem: the favicon drops the dashes entirely and collapses
neck, body and tail into one stroke, and the detail view retunes the body's dash so it stays
legible at its own scale. Any such divergence carries its reason beside it.

**The mechanism is most of the guarantee.** The index, detail and comparison views cannot diverge:
one `diagram` template emits the markup and one stylesheet colours it, so there is nowhere for a
per-page shade to live. The favicon is the exception, and the only one — a standalone asset that
cannot reference the stylesheet and therefore restates the palette by hand. **It is the restatement
that has to be checked**, which is what [acceptance.faviconFromPalette](../../acceptance.md#faviconfrompalette) does.

Rationale: this requirement exists because the thing it forbids had already happened. The favicon
drew its crossbar `#5ccbfa` while the app drew `#0ea5e9` — a hand-lightened variant adopted
because the app's own blue was too dark to clear the icon's contrast standard. That is the
failure in miniature: a surface hit a problem with the shared vocabulary and fixed it locally
instead of fixing the vocabulary. The app's blue was *also* too dark for its own labels (tech.textContrastFloor),
so the local fix concealed a real defect for as long as it stood.

The lesson generalizes past colour. When one surface cannot live with a shared decision, that is
evidence about the decision, not licence to fork it.

## englishNameChoice

**The reader chooses which English name, and the URL does not change**

Subject: `presentation.names.locale`
Value: `na (default) | world`
Goals: goals.sizeIntuition, goals.comparisonAcrossFamilies

A toggle beside Units and Scale chooses between North American English names (eBird/Clements,
the default) and world English names (AviList, following IOC). A species' id and URL are the same
under both.

English bird names are regional. The same bird is **Dovekie** to a North American birder and
**Little Auk** to everyone else, and neither is a mistake. BirdView serves both:

| locale | source | *Alle alle* |
| --- | --- | --- |
| **`na`**, the default | eBird/Clements | Dovekie |
| `world` | AviList, following IOC | Little Auk |

**North American is the default**, this being an ABA-area dataset whose reader calls that bird a
Dovekie. Over the current scope the two differ for **82 species**, most of them hyphenation; the
genuine vocabulary changes are a much smaller set.

Both names are compiled into the dataset, because the locale is a request-time choice and
re-ingesting per locale would be absurd. Five species have no eBird/Clements concept and so no
North American name; they **fall back to the world name** rather than rendering blank.

**Family English names are not localized.** AviList publishes one per family, so "Fringilline and
Cardueline Finches and Allies" becomes "Finches, Euphonias, and Allies" in either locale.
Accepted: a family label is a reading aid, not an identifier.

### The species id is the eBird code, and is locale-independent

A species' URL is `/species/amerob`, and the comparison view names its birds the same way:
`/compare?species=amerob&species=doveki`. The id is the **eBird species code** from
[namesFromAvilist](dataset.md#namesfromavilist).

**This is what makes the locale toggle possible at all.** The four-letter alpha code it replaced
was *derived from the English name*, so under the world locale `DOVE` would have to become
`LIAU` — and an identifier that moves when a display preference moves is not an identifier. The
eBird code is assigned rather than derived, so `doveki` is `doveki` under both names.

It is also simply better as a URL: `/species/amerob` over `/species/turdus_migratorius` is
shorter, typable, and survives a genus move — and genus moves are common, with seven in this
dataset alone.

Three species have no eBird code. They keep a **scientific-name slug** as their id, which is the
old scheme surviving where the new one has nothing to offer.

**Ids must be unique and stable across rebuilds.** Uniqueness is not automatic: AviList lumps two
BirdView species into taxa already in scope, and lumping them in place rather than dropping them
would have given two birds one URL. That is one of the two reasons namesFromAvilist drops them.

## speciesPhotographs

**Species photographs**

Status: **Withdrawn from v1.** See [decisions.photographsWithdrawn](../../decisions.md#photographswithdrawn) for the research and the reasons
it is deferred rather than abandoned.

## perchedOrInFlight

**Perched and in-flight selection**

Status: **Withdrawn from v1** with speciesPhotographs. Note for whoever revives it: Wikimedia Commons has
per-species `in flight` categories but no "perched" equivalent, so the design must treat in-flight
as a positive filter and perched as the residual default.

## photoAttribution

**Photograph attribution**

Status: **Withdrawn from v1** with speciesPhotographs.

## photoCaching

**Local caching of photographs**

Status: **Withdrawn from v1** with speciesPhotographs.

