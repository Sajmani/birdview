# What is measured, what is modeled, and what is invented

This page exists because BirdView draws confident-looking pictures out of data that is, in places,
very thin. If you are going to trust the diagrams, you should know exactly which parts are
observations and which are arithmetic.

The short version: **three of the five diagram segments are modeled, every total length is modeled,
and two thirds of the wingspans are modeled.** BirdView labels all of it, per value, in the UI.

---

## The provenance system

Every number carries one of three kinds, visible on the detail page for any bird:

| Kind | Means |
| --- | --- |
| **measured** | taken from a source, for *this* species |
| **near-match** | measured on a **different** species this one was recently split from, because no measurement exists for it yet |
| **modeled** | computed. The `model` field names the formula and, where one exists, its error |

A `~` in the tile footer marks anything not measured. The rule the project holds itself to is that
a modeled value is never presented as a measurement — see
[CR-001](../spec/decisions.md) for why that decision came first and shaped everything else.

**near-match** deserves a note. Ten species are recent taxonomic splits that AVONET has not caught
up with — American Goshawk, Vega Gull, Hudsonian Whimbrel, Western Cattle-Egret, American Barn Owl,
Chihuahuan Meadowlark, Cocos Booby, Siberian Pipit, Blue-billed White-Tern, Western
Warbling-Vireo. Each borrows the parent taxon's measurements, and the detail page names the taxon
it borrowed from. Four of those splits are still awaiting a taxonomic check
([CR-006](../spec/decisions.md)).

---

## Segment by segment

| Part | Status | Where it comes from |
| --- | --- | --- |
| **Bill** | **measured** — 831 direct, 10 near-match | AVONET `Beak.Length_Culmen` |
| **Tail** | **measured** — 831 direct, 10 near-match | AVONET `Tail.Length` |
| **Mass** | **measured** — 831 direct, 10 near-match | AVONET `Mass` |
| **Wingspan** | **301 measured**, 7 near-match, **533 modeled** | Fox et al. 2026 where empirical; otherwise a fitted regression |
| **Head** | **modeled, unvalidated** — all 841 | `7.5 × mass^(1/3)` |
| **Body** | **modeled, unvalidated** — all 841 | `18 × mass^0.31` |
| **Neck** | **modeled, invented** — all 841 | a per-family ratio × the trunk |
| **Total length** | **modeled** — all 841 | the sum of the five segments above |

Two of the five segments are observations. Three are not.

### Why every length is modeled

This is the project's central compromise and it is worth stating plainly.

AVONET — the only openly licensed morphology dataset with this coverage — **has no body length and
no wingspan**. Field-guide lengths are not openly licensed. The one comprehensive academic source,
Ricklefs (2017), is published under terms that prohibit extraction, storage and redistribution, so
it was evaluated and **rejected** ([CR-002](../spec/decisions.md)).

So BirdView builds length rather than borrowing it: bill + head + neck + body + tail, with the two
measured parts real and the three modeled parts labeled. The segments sum exactly to the total
([P-041](../spec/product.md)), which is what lets the height of a diagram *be* the bird's
length and makes two diagrams at one scale directly comparable.

The consequence is that **BirdView's lengths will not match your field guide.** They are a
different quantity, built a different way, and they should be read as a consistent basis for
comparison rather than as authoritative measurements.

### The wingspan model, which is the one that was actually validated

For the 533 species without an empirical wingspan, BirdView fits its **own** order-level log-log
regression of wingspan on wing chord, trained on the empirical column of the Fox dataset.

It replaced an invented formula whose mean absolute error was 18.7%. Five-fold cross-validation:

| Model | mean | median | p90 |
| --- | --- | --- | --- |
| The invented constants it replaced | 18.7% | 18.2% | 29.9% |
| Global log-log on wing chord | 9.0% | 7.3% | 19.0% |
| **Order-level log-log (adopted)** | **6.0%** | **4.6%** | **12.6%** |

Asked for the same species Fox et al. predicted, it agrees with their published predictions to a
**mean of 1.32%** — it is, in effect, their model reconstructed from the data they published. That
agreement is a standing check, so a regression in the fitting code shows up as disagreement with an
external reference rather than as a quietly worse dataset. Full reasoning:
[CR-008](../spec/decisions.md).

Each modeled wingspan states its own error. A Passeriform reads:

> log-log regression of wingspan on wing chord, fitted on 360 empirical species in Passeriformes
> from the Fox et al. 2026 dataset; cross-validated median error 4.7%

**Note what is not claimed.** Fox report no predicted value for species that already have an
empirical one, so there is no held-out set on which to compare their predictions against ours
directly. The 1.32% is agreement between two models; the error figures are our own
cross-validation.

One important trap avoided: AVONET's `Wing.Length` is the **wing chord** — a single folded wing. It
is not half a wingspan, and doubling it is wrong by a wide and species-dependent margin. BirdView
never does that.

### The three modeled segments, in descending order of confidence

**Body (trunk) — `18 × mass^0.31`.** An allometric relation. Mass is the only whole-animal
measurement AVONET carries, so the trunk is derived from it rather than from a linear trait. The
exponent is in the plausible range for a linear dimension against mass, but the relation is
**unvalidated** — there is no dataset of avian trunk lengths to fit against.

**Head (cranium) — `7.5 × mass^(1/3)`.** Isometric: a linear dimension scales as the cube root of
mass. The coefficient is chosen to put a 20 g passerine near 20 mm. It is a shape function, not a
fit, and it is **unvalidated**.

It replaced a model based on bill depth, `max(12 mm, depth × 1.65)`, which is worth recording as a
failure mode: the floor bound for **442 of 841 species** and 86% of birds under 200 mm, so half the
dataset shared one constant head that varied with nothing. The current model has no floor
deliberately — a 2.5 g hummingbird lands near 10 mm, so a floor would only mask a bad coefficient.

**Neck — a per-family ratio of the trunk. This one is invented.** No source measures avian neck
length at anything like this coverage. The ratios are a drawing parameter: Ardeidae 0.85,
Phoenicopteridae 0.95, Anhingidae 0.90, Gruidae 0.85, down to Strigidae and Tytonidae at 0.08, with
everything unlisted at 0.18.

They are surfaced anyway, with the function named — a heron's detail page reads *"neck modeled as
0.85 × trunk (family-specific ratio for Ardeidae)"*. The reasoning
([CR-007](../spec/decisions.md)) is that the user sees the neck either way, since it is
drawn at a length in every diagram. Hiding the number hides the explanation, not the claim.

---

## Deliberate fictions in the drawing

These are choices about rendering, not data. None of them produces a number.

**The diagram is a top-down view of a bird in flight.** Stated as a requirement
([P-040](../spec/product.md)) because leaving it unstated made several questions
unanswerable. It is also what makes the wing crossbar honest: seen from above, wingspan and body
length genuinely lie in one plane, so drawing them at one scale is a fact rather than a convention.

**There are no legs.** They were drawn from the measured tarsus, then from a modeled leg, and then
removed entirely. In flight most birds tuck the legs away; for the minority that trail them —
herons, cranes, storks, ibises, flamingos, stilts, rails, loons, grebes — what shows is the
projection past the tail tip, which needs the femur, the toes, the hip position and the *full*
tibiotarsus. AVONET measures only the tarsometatarsus. Attempting it anyway put a White Ibis at
−14 mm of projection, a bird whose trailing legs are unmistakable.
[CR-010](../spec/decisions.md) has the numbers and what the removal cost.

**Herons and pelicans get an S-shaped neck.** Ardeidae and Pelecanidae fly with the head drawn back
onto the shoulders — the field mark separating a heron from a crane, a stork or an ibis in the air.
Their neck is drawn as two semicircles, which run exactly the neck's length of stroke through
`2/π ≈ 63.7%` of that height. The stroke is still to scale; the compression is a consequence of the
shape, not a chosen number. Sorting, sectioning and the detail view all still use the full neck
length ([CR-011](../spec/decisions.md)).

**The head is an oval 0.72 as wide as it is long**, the tail is a stick, and the body is a stick.
No source gives a tail width or a body width, and drawing a fan or a teardrop would assert a shape
the data does not support.

**Colors encode taxonomy only** — related hues within an order, distinguishable between families.
They say nothing about the bird.

---

## Scope

**841 species**: ABA Checklist 8.19, rarity codes 1–3. Codes 4 and 5 — casuals and accidentals —
are excluded, and rarity is used once at build time as a filter and then discarded, so it never
reaches the dataset or the UI.

**Galliformes have no empirical wingspan at all.** All 30 in-scope species are absent from the Fox
dataset, so every one is modeled.

Taxonomy is ABA's, which is the AOS checklist. Where AVONET's BirdLife taxonomy disagrees, ABA
wins, and the join is bridged by an explicit synonym table rather than by fuzzy matching.

---

## Where this is most likely to be wrong

In rough order of how much it would change the picture:

1. **The neck ratios are invented.** They are the largest unforced assumption in the project, and
   they visibly drive the silhouette of every long-necked family.
2. **The head and body models are unvalidated.** Both are plausible shape functions fitted against
   nothing. If either coefficient is off, every length is off in the same direction.
3. **Lengths are not comparable to published lengths.** They are internally consistent, which is
   what the tool needs, but do not quote them.
4. **Two thirds of wingspans are modeled** — accurately, with a stated error, but modeled.
5. **The flight posture is uniform.** Every bird is drawn with the same implied wing position and
   the same body axis. Real birds differ enormously.
6. **One figure per species.** No sex, age, subspecies or individual variation, in a dataset where
   some species vary by 30% between sexes.

**If you can point at a source that fixes any of these** — particularly a validation set for head
or trunk length, or an openly licensed body-length dataset — that is the single most useful
contribution anyone could make. Adopting a source is a defined process, not an ad-hoc patch: see
[spec/sources.md](../spec/sources.md) and the source-adoption path in
[docs/building.md](building.md).

Corrections: **birdview at ajmani dot net**.
