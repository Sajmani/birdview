package product

# Diagram

## segmentComposition

**Segment composition**

Subject: `diagram.segments`
Value: `bill, head, neck, body, tail, wing crossbar`
Goals: [goals.shapeNotOnlySize](../../../goals/goals.md#shapenotonlysize)

The diagram is a top-down view of a bird in flight, drawn as a vertical cross: top to bottom a bill
stick, a head oval, a neck, a body stick, and a tail stick, with one horizontal crossbar for the
wings. The tail is visibly distinguished from the body; the neck is an S where the family folds it
in flight; there are no legs.

The flight premise is stated because leaving it unstated made the leg question unanswerable; see
[decisions.inFlightNoLegs](../../decisions.md#inflightnolegs).

**Readers are told it is a stick figure.** Text BirdView serves to readers calls the drawing a
*stick figure*, usually *drawn to scale* where it is introduced, and *the figure* or *the drawing*
after that. *Diagram* stays the word in this spec and in the code, and on the site it means only
the data-pipeline diagram ([pipelineDiagram](measurements.md#pipelinediagram)); see [decisions.stickFigure](../../decisions.md#stickfigure).

The tail is distinguished from the body by stroke style, color, and a mark at the junction between
them. Under measuredVersusModeled the tail is measured and so solid, the body modeled and so
dashed; the junction mark carries the distinction on its own if the two ever share a stroke style.

**The neck is a stick, except where the family folds it in flight.** Ardeidae and Pelecanidae fly
with the head drawn back onto the shoulders, and for those the neck is drawn as an **S of two
semicircular arcs**. The stroke is still exactly the neck long at the diagram's scale, so it
occupies `2/π ≈ 63.7%` of that length vertically — a consequence of the shape, not a chosen
parameter; see [decisions.foldedNeckAsS](../../decisions.md#foldedneckass).

**There are no legs.** Most birds tuck them in flight, and for those that trail them — herons,
cranes, storks, ibises, flamingos, stilts, rails, loons, grebes — what shows is the projection past
the tail tip. That depends on the femur, the toes, and the hip's position, which no source
measures, and on the feathered tibiotarsus, which a standing-posture leg model leaves out.
[decisions.inFlightNoLegs](../../decisions.md#inflightnolegs) records the evidence and
what the removal costs.

Rationale: the tail is a stick rather than a triangle because no trait in any source gives a tail
width, and a triangle would assert a fan shape the data does not support.

## segmentsSumToLength

**Segments abut exactly and sum to total length**

Subject: `diagram.segments.sum`
Goals: [goals.sizeIntuition](../../../goals/goals.md#sizeintuition), [goals.shapeNotOnlySize](../../../goals/goals.md#shapenotonlysize)

The five vertical segments meet without gap or overlap, and their lengths sum to the species'
total length.

Rationale: this makes the vertical extent of the drawing the bird's length, so two diagrams at the
same scale can be compared directly.

**Qualification.** The equality holds for segment *lengths* without exception, and for the drawn
vertical extent of 827 of 841 species. The folded-neck families of segmentComposition draw shorter
than their cataloged length, because the S occupies less height than the neck is long. Everything
that treats length as data — ordering, sectioning, the scale reference of scaleToggle, and the
detail view — uses the full total length, so comparability is unaffected. See
[decisions.foldedNeckAsS](../../decisions.md#foldedneckass).

## scaleToggle

**Scale is a user toggle, defaulting to per-section**

Subject: `diagram.scale.mode`
Value: `per-section (default) | global; global is the default under Frequency, Bill and Tail`
Goals: [goals.sizeIntuition](../../../goals/goals.md#sizeintuition), [goals.flexibleComparison](../../../goals/goals.md#flexiblecomparison)
Status: **Withdrawn 2026-10-27**, with the classic view, which the owner removed when the home page replaced it ([decisions.classicViewRemoved](../../decisions.md#classicviewremoved)).

The user switches between a single scale across the entire filtered set, and a scale shared by all
species within a section. **Per-section is the default, and is listed first.**

**Except under Frequency, where global is the default** ([decisions.frequencyGlobalScale](../../decisions.md#frequencyglobalscale)).
Per-section scaling relies on [sectionBreakRule](ordering.md#sectionbreakrule) bounding a section's width, and a frequency tier
has no such bound: one tier runs from 72 mm to 1,201 mm, and the single 140 mm bird in the last
tier would be drawn the height of the 1,201 mm Trumpeter Swan six tiers above it. Both choices stay one click away under either ordering.

**Under Bill and Tail global is the default too, for the same reason.** A section bounded in bill or
tail length is not bounded in total length: with 15% sections, one bill section and one tail
section each span 7.1× in total length, where length sections stay within 1.1× and mass sections
within 3.1×.

Whichever is chosen, the scale is set by the largest total length in the group. Since the segments
sum to total length (segmentsSumToLength) and nothing is drawn outside them, that is the whole vertical extent.

Rationale: [goals.sizeIntuition](../../../goals/goals.md#sizeintuition) depends on this. A per-tile scale — every bird drawn to fill its
tile — makes a hummingbird and a swan identical on screen.

Per-section is the default because at global scale a median bird is a few pixels tall, too small
to show the stroke styles of measuredVersusModeled, the chord ticks, or the labels of
labeledDetailDiagram ([decisions.perSectionDefault](../../decisions.md#persectiondefault)). Global still makes the cross-family comparison
best, and stays one click away and named first in the help.

## crossbarAtVerticalScale

**The wing crossbar is drawn at the vertical scale**

Subject: `diagram.wingspan.scale`
Value: `same as vertical axis`
Goals: [goals.sizeIntuition](../../../goals/goals.md#sizeintuition), [goals.shapeNotOnlySize](../../../goals/goals.md#shapenotonlysize)

The crossbar's length represents wingspan at the same scale as the vertical axis, so its length
relative to the body is literal.

Rationale: wingspan runs 1.0–2.5× body length, so this forces a wide tile, roughly 2.6:1. The cost
is accepted because a crossbar on its own scale would misstate proportion while looking
authoritative.

## proportionDifferentiates

**Proportion differentiates species**

Goals: [goals.shapeNotOnlySize](../../../goals/goals.md#shapenotonlysize)

A species' drawn proportions distinguish it by shape, not only by size: a dowitcher reads as
long-billed, an accipiter as long-tailed, a heron as long-necked.

## craniumFromMass

**The cranium is modeled from mass**

Subject: `diagram.head_model`
Value: `isometric on mass`
Goals: [goals.shapeNotOnlySize](../../../goals/goals.md#shapenotonlysize)

Head length scales isometrically with body mass. Like every modeled value it names its function
(provenancePerMetric) and claims no measurement.

Rationale: an earlier model derived the cranium from bill depth with a 12 mm floor, and the floor
bound for **442 of 841 species**, giving half the dataset one identical head. Mass is the only
whole-animal measurement available, and a linear dimension scaling as its cube root is the natural
default. There is deliberately **no floor**: a floor would only conceal a bad coefficient.

Known limitation: birds of similar mass get similar heads, so within a size tier the shape signal
comes from bill, tail, neck, and wingspan.

**[decisions.trunkExponentCalibrated](../../decisions.md#trunkexponentcalibrated) leaves the cube root alone.** Single skeletal elements
scale at 0.314–0.325, chains of them higher; the head is one element and keeps 1/3, the trunk is a
chain and does not. See [trunkFromMass](#trunkfrommass).

Still open: an Ostrich draws a 360 mm cranium against a real skull of roughly 200 mm, suggesting
negative allometry. No head dataset at this coverage exists to fit against, so the coefficient
stands unvalidated. It is no longer moot: Emu, now in scope at 34 kg, draws a 243 mm cranium.

## trunkFromMass

**The trunk is modeled allometrically from mass**

Subject: `diagram.trunk_model`
Value: `14.0329 x mass^0.37`
Goals: [goals.sizeIntuition](../../../goals/goals.md#sizeintuition), [goals.shapeNotOnlySize](../../../goals/goals.md#shapenotonlysize)

Trunk length scales with body mass by a power law. Like every modeled value it names its function
(provenancePerMetric) and claims no measurement.

The **exponent is calibrated against measured skeletons** rather than chosen. It must lie inside
the 95% confidence interval that published skeletal data give for a *multi-element* linear
dimension ([`bohmer-neck`](../../imports.md#bohmer-neck)) — **0.338 to 0.401**, the intersection of the
intervals for neck length and total leg length. Single bones scale near isometry, chains of them
higher, and a trunk is a chain.

The **coefficient is a consequence, not a second free parameter.** It is set so the model is
unchanged at **63.4 g**, the median mass of the ABA dataset when the exponent was calibrated, so
recalibration changes proportion across the size range rather than overall scale.

That anchor is a **fixed constant, not a value read from the dataset at build time.** If scope
changes the median moves, and the drawing must not move with it silently; re-anchoring would be a
decision, recorded as one.

Rationale: three of the five segments are driven by mass, so this exponent governs how much longer
a big bird is drawn than a small one, which is what [goals.sizeIntuition](../../../goals/goals.md#sizeintuition) exists to convey. The
original unvalidated 0.31 understated the large end of the current scope by about 1.65×. See
[decisions.trunkExponentCalibrated](../../decisions.md#trunkexponentcalibrated) for the measurement, and for what it does not establish.

## neckModelPerFamily

**The neck/body split is a named per-family model**

Subject: `diagram.neck_ratio`
Value: `per-family modeling function`
Goals: [goals.shapeNotOnlySize](../../../goals/goals.md#shapenotonlysize), [goals.honestData](../../../goals/goals.md#honestdata)

The division of the remaining body axis between neck and body uses a per-family ratio. The
resulting **neck length is presented as a metric like any other**, in the detail view, carrying a
provenance record whose `kind` is `modeled` and whose `model` **names the specific function and
the ratio applied to that bird's family** — for example, "per-family neck ratio, Ardeidae = 0.35".

The ratio table is recorded in the code with this requirement cited beside it.

Rationale: no source measures avian neck length at this coverage, and one ratio for every species
would make every bird the same shape, defeating [goals.shapeNotOnlySize](../../../goals/goals.md#shapenotonlysize). Naming the per-family
constant is what makes it acceptable: the reader sees the neck either way, and can read the exact
function that drew it. Supersedes the condition in [decisions.neckSplitUnmeasured](../../decisions.md#necksplitunmeasured); see
[decisions.neckLengthAsMetric](../../decisions.md#necklengthasmetric).

## oneDiagramVocabulary

**One diagram vocabulary, on every surface that draws one**

Subject: `diagram.vocabulary.shared`
Goals: [goals.sizeIntuition](../../../goals/goals.md#sizeintuition), [goals.flexibleComparison](../../../goals/goals.md#flexiblecomparison)

Three surfaces draw the diagram: the **favicon**, the **home page**'s tiles, and the **detail
view**. A reader meets the same figure on all three, so the figure means the same thing on all
three.

**Color is shared exactly.** Each part of the diagram — bill, head, neck, body, tail, wing
crossbar, mass cube — has one color, and every surface uses that color for that part. A surface
may not pick its own shade of a part, however good its local reason.

**Size-dependent geometry is not shared.** Stroke weight, dash pattern, and which parts appear may
be tuned per surface: the favicon drops the dashes and collapses neck, body, and tail into one
stroke, and the detail view retunes the body's dash for its scale. Any such divergence carries its
reason beside it.

**The mechanism is most of the guarantee.** The home page and the detail view share one
`diagram` template and one stylesheet, so they cannot diverge. The favicon is the only exception: a
standalone asset that restates the palette by hand. **That restatement is what has to be
checked**, by [acceptance.faviconFromPalette](../../acceptance.md#faviconfrompalette).

Rationale: the favicon once drew its crossbar `#5ccbfa` while the app drew `#0ea5e9`, lightened
locally because the app's blue failed the icon's contrast standard. The app's blue also failed
for its own labels ([tech.textContrastFloor](../tech/accessibility.md#textcontrastfloor)), so the local fix concealed a real defect. When one
surface cannot live with a shared decision, that is evidence about the decision, not license to
fork it.

## englishNameChoice

**The reader chooses which English name, and the URL does not change**

Subject: `presentation.names.locale`
Value: `na (default) | world`
Goals: [goals.shareableViews](../../../goals/goals.md#shareableviews), [goals.openToEveryone](../../../goals/goals.md#opentoeveryone)

A setting beside Units, **Names**, chooses between North American English names (eBird/Clements,
the default) and world English names (AviList, following IOC). A species' id and URL are the same
under both.

| locale | source | *Alle alle* |
| --- | --- | --- |
| **`na`**, the default | eBird/Clements | Dovekie |
| `world` | AviList, following IOC | Little Auk |

**North American is the default** because this dataset's reader calls that bird a Dovekie. The two
differ for **82 species**, mostly in hyphenation.

Both names are compiled into the dataset, since the locale is a request-time choice. Five species
have no eBird/Clements concept and so no North American name; they **fall back to the world name**
rather than rendering blank.

**Family English names are not localized.** AviList publishes one per family, used in either
locale; a family label is a reading aid, not an identifier.

### The species id is the eBird code, and is locale-independent

A species' URL is `/species/amerob`, and the home page pins birds the same way:
`/?pin=amerob&pin=doveki`. The id is the **eBird species code** from
[namesFromAvilist](dataset.md#namesfromavilist).

**This is what makes the locale toggle possible.** An id derived from the English name would move
when the display preference moved; the eBird code is assigned, so `doveki` is `doveki` under both
names. It is also short, typable, and survives a genus move.

Three species have no eBird code. They keep a **scientific-name slug** as their id.

**Ids must be unique and stable across rebuilds.** Uniqueness is not automatic: AviList lumps two
BirdView species into taxa already in scope, and keeping them would give two birds one URL. That
is one of the two reasons namesFromAvilist drops them.

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
