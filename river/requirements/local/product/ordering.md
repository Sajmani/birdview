package product

# Ordering and sectioning

## oneSelectedMetric

**One primary metric, user-selected**

Subject: `ordering.metric`
Value: `length | wingspan | mass | bill | tail | frequency`
Goals: [goals.sizeIntuition](../../../goals/goals.md#sizeintuition), [goals.shapeNotOnlySize](../../../goals/goals.md#shapenotonlysize), [goals.flexibleComparison](../../../goals/goals.md#flexiblecomparison)

The user chooses whether the dataset is ordered by total length, wingspan, body mass, bill length,
tail length, or how often the bird is recorded ([orderByFrequency](presentation.md#orderbyfrequency), [decisions.frequencyMetric](../../decisions.md#frequencymetric)).
The comparison view offers the same size orderings for the birds it holds.

**Bill and tail are absolute lengths, not proportions.** They are measured for every species; a
ratio to total length would put modeled values under every ratio, since three of total length's
segments are modeled from mass.

## ascendingOrder

**Ascending order**

Goals: [goals.sizeIntuition](../../../goals/goals.md#sizeintuition)

Species are sorted ascending by the selected metric, **and the page says which way that
runs**: *smallest first* under length, wingspan, mass, bill, and tail, *most common first* under
Frequency, captioned under each group of controls rather than one caption that follows the
selection. The row aligns on the text baseline, so the heading and every button stay on one
line and only the captions hang below.

The direction is stated for every metric, not only the one that departs: Frequency ascends a
reciprocal — *1 in N* rises as the bird gets scarcer — so a reader who has learned "smallest
first" elsewhere would otherwise read it as a defect. See
[decisions.frequencyMetric](../../decisions.md#frequencymetric) for why renaming it was declined.

## sectionBreakRule

**Sections break at N% above the section minimum**

Subject: `sectioning.rule`
Goals: [goals.sizeIntuition](../../../goals/goals.md#sizeintuition), [goals.flexibleComparison](../../../goals/goals.md#flexiblecomparison)

Under the three size metrics, a new section begins at the first species more than N% larger than
the **minimum** species of the previous section.

Frequency is banded by decade instead
([orderByFrequency](presentation.md#orderbyfrequency), [decisions.frequencyMetric](../../decisions.md#frequencymetric)): this rule would give it 128 sections, and its
bands would move whenever the region filter did.

Rationale: bird size follows a power law, so fixed-width bins crowd small birds and isolate large
ones. A ratio rule gives sections of comparable visual span at every scale, and, because a
section is no more than N% wide, makes the per-section scale of scaleToggle honest.

## sectionThresholdDefault

**N defaults to 15 and is adjustable**

Subject: `sectioning.step_percent`
Value: `15`
Goals: [goals.flexibleComparison](../../../goals/goals.md#flexiblecomparison)

The section step N defaults to 15% and the reader can change it, among values from 5% to 30%.

Every value the control offers must be one the server accepts, or the control silently falls back
to the default.

**Under Frequency the same control selects a band width in orders of magnitude**, `10×` or
`100×` ([orderByFrequency](presentation.md#orderbyfrequency)). Percentages would misdescribe it: one decade is +900%.

## sectionHeadersStateRange

**Section headers state range and count**

Goals: [goals.sizeIntuition](../../../goals/goals.md#sizeintuition)

Each section shows its range in the currently selected metric and the number of species in it,
under a heading that names the ordering as Order by does and the section's number: *Tail Tier 1*
when ordering by Tail. The count follows the range directly, as *N species*, in the style of the
*K of N species* line at the top of the index.

**The range is the tier's own birds**, smallest to largest; a tier of one bird states one value.
Tiers do not abut, and the stretch between them can be wide. So every heading on a page carries
the same number of decimals, the fewest — up to three more than usual — that keep every stated
value apart, within a heading and from its neighbors, and a gap between tiers shows as a gap
between the numbers.

Rounding heading by heading made adjacent ranges read as overlapping ("0.2 in – 0.3 in", "0.3 in
– 0.4 in"); running each range on to the next tier's start overstated it.

Rationale: under per-section scaling (scaleToggle) the header is what tells the user that the scale has
changed between one section and the next. Without it, per-section scaling is misleading.

## tierHistogram

**A histogram shows how many species each tier holds**

Subject: `sectioning.histogram`
Goals: [goals.localRelevance](../../../goals/goals.md#localrelevance)

Under the species-count line, a histogram has one bar per tier, in tier order, each as tall as its
tier's species count against the largest tier's, and each linking to its tier. A bar's label names
the tier as its heading does, its range, and its count: *Bill Tier 4: 0.35 in – 0.41 in, 72
species*. It follows the current ordering and filters, so it is the shape of the view on screen.

Every tier is the same 15% step, so the bars are a histogram of size on a log scale, and **a
stretch with no birds shows as empty space**: one empty bucket for each whole step between where a
tier's step ends and the next tier begins, so tiers far apart do not stand side by side as if they
abutted. Frequency's bands are fixed decades and need none.
