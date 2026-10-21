# Background notes

Design notes that items in [todo.md](todo.md) and [todo-closed.md](todo-closed.md) refer to, moved
from decisions.md on 2026-10-13. Search `## <heading>`; `grep -n '^## ' river/notes.md` lists them.

Design notes the rows above refer to, moved here from decisions.md with the list on 2026-10-13.

## A cube of water for mass

*Evaluated 2026-10-01, owner's idea, not a decision.* [product.labeledDetailDiagram](requirements/local/product/measurements.md#labeleddetaildiagram) rejects drawing
mass as a shape and gives two reasons. The owner then proposed a specific shape that answers one
of them completely, so the rejection is narrower than it reads and the measurements are recorded
here rather than left to be redone.

**Draw the cube of water that weighs what the bird weighs**, side = ∛(mass). This is not a model
of anything: one gram of water is one cubic centimetre, by the historical definition of the gram,
so the side is a unit conversion and carries no assumption about avian density. It is a real
length, so it can be drawn at the diagram's own scale and does not introduce the second,
arbitrary scale product.crossbarAtVerticalScale refuses. product.labeledDetailDiagram's objection on that count does not apply to it.

**Measured on the real data, at the detail view's geometry** (drawing area 668 × 281):

| | side |
| --- | --- |
| smallest (Fork-tailed Flycatcher, 32 g) | 28 px — 10% of the drawing height |
| median | 52 px — 18% |
| largest (Thick-billed Murre, 964 g) | 80 px — 29% |

It fits. Placed in the lower-left corner with an 8px margin, the crossbar intrudes on it for
**0 of 841 species**; the tightest clearance is 50 px, on Western Grebe.

**What it would add that nothing else shows.** The extremes are not arbitrary: the largest cubes
belong to alcids — Thick-billed Murre, Parakeet Auklet, Crested Auklet, Marbled Murrelet, all
dense diving birds — and the smallest to long-tailed, light-bodied species, Fork-tailed and
Scissor-tailed Flycatcher and Black-billed Magpie. The cube read against the bird's own length is
**bulk against elongation**, which is exactly the thing a length-and-wingspan diagram cannot say.

**What it still does not do**, and the reason product.labeledDetailDiagram's second objection stands: it compresses. A
Bushtit and a Common Raven differ 175× in mass and their cubes differ 5.6×. That is arithmetically
correct for equal-volume cubes and visually understates the gap. It matters less now that the
number is on the figure — the label can carry the magnitude and the cube the tangibility — but a
reader who takes the cubes as the mass comparison will read them wrong.

Open if it is taken up: what labels it, whether it appears on index tiles (it should not — the
tile is dense by design and this is detail-view furniture, as product.labeledDetailDiagram's labels already are), and
whether it is drawn solid or dashed for the 10 near-match masses, which product.measuredVersusModeled's vocabulary already
answers.

## Mass as a proportional line, replacing the cube — parked

*Evaluated 2026-10-02, owner's idea, parked rather than decided.* Everything measured here is
still true; pick it up from this section rather than re-deriving it.

**The complaint is correct and is already in the spec.** The cube root flattens: a Bushtit and a
Common Raven differ in mass by 175× and their cubes by 5.6×. product.labeledDetailDiagram admits this and answers it with
the printed number. The proposal was to answer it with the drawing instead — a horizontal line
below the bird, from the lower-left, ticks at both ends, and the Mass label at the far end as the
Wingspan label already sits at the end of the crossbar.

**Then the owner looked at a rendered section and reported the opposite: the cubes draw the eye to
the heavier birds, and work.** That is evidence, and it is worth more than the argument above, so
it is recorded first. Measured on the section in that screenshot — #9, 51 species:

| | |
| --- | --- |
| length spread | **1.15×** (272–312 mm), by construction |
| mass spread | **8.3×** — Mangrove Cuckoo 63.9 g to Pigeon Guillemot 530 g |
| cube edge spread | **2.02×** |
| apparent area spread | **4.10×** |

A cube twice as wide as its neighbour, in a row of birds that are all the same length, is not
subtle. **The cube root is mild compression over the range a reader actually compares across.**

That reframes the Bushtit-and-Raven example this section opens with. Those two differ 175× in
mass — and they are 11 sections apart, never drawn side by side at a shared scale. The complaint
is strongest exactly where the comparison is not being made, and weakest where it is. Per-section
scale is the default (product.scaleToggle), so the common case is the favourable one.

**It also rescues a comparison the numbers break.** In that same section the Pigeon Guillemot's
mass prints as `1.2 lb` and the Eared Grebe's as `14.5 oz`, because the formatter switches unit at
16 oz. Two birds of the same length, one third heavier, in units a reader has to convert before
comparing. The cubes are directly comparable where the labels are not — an argument for the shape
that nobody made when decisions.massAsWaterCube adopted it.

**Where that leaves the line.** The case for it is weaker than when this section was written: it
would buy proportionality over a range where the cube is already legible, and pay the price set
out below. Not closed — the measurements stand and the design is worked out — but anyone picking
it up should start by disagreeing with the screenshot.

**The finding that decides the shape of the work: a line linear in mass cannot be drawn at one
global scale.** The mass range is **4,258×** (2.6 g – 11,071 g). A 380 px tile has a usable
dynamic range of roughly **180×**. With the Trumpeter Swan's line at 300 px the Calliope
Hummingbird's is **0.07 px**.

Nor is there a physically honest way out. The cube's whole elegance is that one gram of water is
one cubic centimetre *by definition*, so its edge is a true length drawn at the figure's own
scale and invents no second scale. The line equivalent is a water column of fixed cross-section,
where length in centimetres equals mass in grams — and the swan's column is **92× its own body
length**. No choice of cross-section helps, because the range to be spanned exceeds the range a
tile can draw by 24×.

**So a proportional mass line must be normalized per group, and that is the trade.** Measured
within the 20 sections at the default 15% step:

| within one length section | |
| --- | --- |
| median mass range | **4.0×** |
| worst section | **20.0×** — 57 species, 31.9–637 g |
| sections above 50× | **0 of 20** |

Comfortably drawable. Note what is and is not given up: **the bird is already drawn at a
per-section scale** (product.scaleToggle), and so is the cube, so a per-section mass scale is not a new kind of
relativity. What the cube has that a normalized bar would not is that its size *means* something
absolute — you could measure it against the bird with one ruler. A normalized bar is a bar-chart
row: very readable, making no physical claim.

**It therefore needs a CR, because it reverses rather than extends decisions.massAsWaterCube.** product.labeledDetailDiagram currently
states that "a disc or bar sized by mass — a second, arbitrary scale — is still refused", and
product.crossbarAtVerticalScale's rationale is that a mark on its own convenient scale "looks authoritative and says
something false". Those are the sentences a CR would have to overturn, not work around.

**The hard case is the comparison view**, where a reader may pick a hummingbird against a swan and
the lighter line goes sub-pixel. That is arguably the correct message — the cube is precisely what
hides it — but invisible reads as broken. The mitigation is to draw the origin tick
unconditionally, so there is always a mark, and to let the number carry the magnitude as product.labeledDetailDiagram
already says it must.

**Styling, as far as it was worked out:**

- **Solid.** Mass is measured for 831 of 841, and product.measuredVersusModeled draws `near-match` solid too.
- **Not a heavier weight.** Stroke weight carries no meaning in this diagram today, and the
  heaviest mark on a bird diagram should not be the one that is not the bird.
- **A different colour, and this is the load-bearing one.** A horizontal solid ticked line will
  otherwise read as a second crossbar — another length to scale, which is the exact false claim
  product.crossbarAtVerticalScale exists to prevent. Contrast measured against the `#0f172a` background, and against the
  wing's `#0ea5e9`, because a hue-only difference fails for colour-blind readers:

  | candidate | vs background | vs wing |
  | --- | ---: | ---: |
  | lime `#a3e635` | 11.8:1 | **1.8:1** |
  | emerald `#34d399` | 9.3:1 | 1.4:1 |
  | fuchsia `#e879f9` | 7.3:1 | 1.1:1 |
  | violet `#a78bfa` | 6.6:1 | 1.0:1 |

  Lime has the best luminance separation from the crossbar; the others differ by hue alone.
- **Ticks at both ends** — consistent with the chord-tick vocabulary of product.measuredVersusModeled.
- **Label at the far end** — consistent with the wingspan label in product.labeledDetailDiagram.

**Two things found along the way that are worth acting on whether or not this is ever built:**

1. **The cube is `#38bdf8` and the crossbar `#0ea5e9`.** The one mark that is *not* a length is
   near-identical in colour to the mark whose entire point is that it *is* a length to scale.
   That is backwards, and it is a defect in what ships today.
2. **There is no clear band below the bird on a tile.** The tallest bird in each section fills the
   tile height exactly. The cube works because it occupies the empty lower-left *beside* a narrow
   figure; a full-width horizontal line would cross the tail unless the tile grows downward, as
   the detail view's bottom strip already does for product.labeledDetailDiagram's labels.

## Cervical vertebra counts, and why they are worth chasing

*Recorded 2026-09-23.* The per-family neck ratios of product.neckModelPerFamily are the largest unforced assumption in
the project: invented, unvalidated, and visibly driving the silhouette of every long-necked family.
Three families have now been corrected by eye — Aramidae and Pelecanidae after a reader noticed a
short-necked Limpkin, Sulidae on review — which is a sign the table is guesswork with a slow error
signal, not that it is converging.

**Cervical vertebra count is the most promising anchor available.** Birds run from roughly 11 to 25
cervicals, against a mammalian fixed seven, and the count tracks neck length closely: swans sit at
the top of the range, passerines near the bottom, galliforms between. Unlike neck length it is a
discrete, published, comparable anatomical trait.

It would not give the ratio directly — ours is neck ÷ trunk, so it needs count × mean vertebra
length ÷ trunk — but it would give a defensible **ordering across all 91 families**. That is enough
to calibrate the table against the few anchors the project is most confident about, instead of
inventing each entry independently and waiting for someone to spot the next Limpkin.

Two things are unknown and should not be assumed: whether a dataset exists at BirdView's coverage,
and whether its license permits use. Both have sunk candidate sources before. Adoption would follow
the usual path — a `module.md` entry and a compose-and-resolve pass.

## Sketch: one dash per vertebra

*Recorded 2026-09-23, owner's idea, not a decision.* If the counts existed, the neck could be drawn
as a **dashed line with one dash per cervical vertebra** — putting a measured, discrete count into
the one segment that currently carries the least support.

What makes it more than decoration: two birds can reach the same neck length by different means. A
swan is near the top of the vertebral range; a flamingo has a comparably long neck built from
fewer, longer vertebrae. That distinction is real, and the diagram cannot presently show it at all.

Three obstacles, recorded so the idea is picked up with its problems attached rather than
rediscovered:

- **The tail is already dashed**, to distinguish it from the body. Dash would then mean two
  different things on one figure. The tail would need another treatment; there are options.
- **It cannot work on the index tiles.** The neck is a fixed fraction of a bird that fills its tile,
  so its drawn length is roughly scale-independent: about 31 px for the longest-necked species in
  the set, and under 2 px for a median bird. Twenty dashes in 31 px is a grey blur.
- **On the detail view it is marginal**, at 1.9× the tile: a heron's neck is about 62 px, giving
  roughly 3 px per vertebra.

That last figure is the interesting one. Nobody will count 24 dashes, but 3 px is enough to read as
**texture** — finely divided against coarsely divided — which is exactly the swan-versus-flamingo
distinction the idea exists to surface. Comparability rather than countability may be the right
goal, and would be the thing to test.

## Why the perched toggle is not simply the other half of decisions.foldedNeckAsS

*Recorded 2026-09-23.* Offering "show them perched" sounds like a rendering switch and is not one.
A perched bird needs two things BirdView does not have.

**Legs.** [decisions.inFlightNoLegs](requirements/decisions.md#inflightnolegs) removed them because a flight silhouette does not show them; a perched
silhouette is mostly leg. Restoring them restores every problem decisions.inFlightNoLegs recorded — the femur, the
toes and the hip position are unmeasured, and the tarsus alone understates the visible leg most in
the birds whose legs define them. A perched view would need the exposed-leg model that decisions.legFromTarsus
invented and decisions.inFlightNoLegs deleted, and it would still have no source to validate against.

**A second neck model.** The current ratios are a standing-posture model used, until decisions.foldedNeckAsS, in a
flight drawing. A perched view needs the perched extension for every family, which is a different
invented table again — a perched heron's neck is neither folded flat nor fully extended, and where
it sits between those is not published at this coverage.

So the toggle is not a view of the same data from another angle. It is a second set of invented
per-family parameters with no empirical check, which is the thing this project has twice decided
not to buy. Worth revisiting only if a source appears that measures posture, not just skins.

## Why the permission asks wait for a demo

*Recorded 2026-09-21.* Both asks are deferred until BirdView runs and can be shown.

A request that describes an intended use is weaker than one that demonstrates it. Showing a
rights-holder the actual application — what it does with their data, how the data is credited,
that it is non-commercial and that every value carries visible provenance — answers most of what
they would otherwise have to take on trust. It also lets them see that BirdView labels a modeled
length as modeled, which is the strongest available evidence that their measured data would be
handled honestly.

Neither ask blocks v1, which is what makes waiting free:

- **Ricklefs / ESA** would *add* capability. Every length is currently `modeled`
  ([decisions.ricklefsProhibited](requirements/decisions.md#ricklefsprohibited)); a grant would make roughly half of them measured. Nothing breaks without
  it.
- **ABA** would *simplify the build*. Fetch-at-build already works under tech.fetchRestrictedInputs, so a grant would
  only remove the network step and let tech.datasetNeverCommitted relax. Nothing breaks without it either.

Sending them together after a demo also means one conversation per organization rather than two
rounds of the same explanation.

**Update, 2026-09-27.** The ABA request was sent, with screenshots rather than a live demo. The
obvious plan — deploy somewhere they can reach and link it — was circular: hosting BirdView
publicly is the very thing the email asks permission for. Screenshots reproduce almost nothing, to
the rights-holder, about their own data, and need no server to stay up. The Ricklefs/ESA ask is
still outstanding.

**The cost of waiting**, recorded so it is a choice rather than a drift: if ESA grants permission
late, the length column changes from wholly `modeled` to half `measured`, which will move diagrams
and re-sort sections. Builders should treat the provenance model (product.provenancePerMetric) as the thing that stays
fixed and the mix of kinds as the thing that moves.
