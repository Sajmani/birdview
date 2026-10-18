package product

# Ordering and sectioning

## oneSelectedMetric

**One primary metric, user-selected**

Subject: `ordering.metric`
Value: `length | mass | wingspan | frequency`
Goals: goals.sizeIntuition, goals.comparisonAcrossFamilies

The user chooses whether the dataset is ordered by total length, body mass, wingspan, or how
often the bird is recorded ([orderByFrequency](presentation.md#orderbyfrequency), [decisions.frequencyMetric](../../decisions.md#frequencymetric)).

## ascendingOrder

**Ascending order**

Goals: goals.sizeIntuition

Species are sorted ascending by the selected metric, **and the page says which way that
runs**: *smallest first* under length, wingspan and mass, *most common first* under
Frequency, captioned under each group of controls rather than one caption that follows the
selection. The row aligns on the text baseline, so the heading and every button stay on one
line and only the captions hang below.

The direction is stated for all four rather than only for the one that departs. Frequency
ascends a reciprocal — *1 in N* rises as the bird gets scarcer — so ascending the printed
number descends abundance, and a reader who has learned "smallest first" from three metrics
would otherwise read the fourth as a defect. See
[decisions.frequencyMetric](../../decisions.md#frequencymetric) for why renaming it was declined three times.

## sectionBreakRule

**Sections break at N% above the section minimum**

Subject: `sectioning.rule`
Goals: goals.sizeIntuition, goals.comparisonAcrossFamilies

Under the three size metrics, a new section begins at the first species more than N% larger than
the **minimum** species of the previous section.

Frequency is banded by decade instead
([orderByFrequency](presentation.md#orderbyfrequency), [decisions.frequencyMetric](../../decisions.md#frequencymetric)): this rule would give it 128 sections, and its
bands would move whenever the region filter did.

Rationale: bird size follows a power law. Fixed-width bins put hundreds of small birds in one row
and isolate large raptors in rows of one. A ratio rule produces sections of comparable visual
span at every scale — and, because a section is by construction no more than N% wide, makes the
per-section scale of scaleToggle honest.

## sectionThresholdDefault

**N defaults to 15 and is adjustable**

Subject: `sectioning.step_percent`
Value: `15`
Goals: goals.comparisonAcrossFamilies

The section step N defaults to 15% and the reader can change it, among values from 5% to 30%.

Every value the control offers must be one the server accepts, or the control silently falls back
to the default.

**Under Frequency the same control selects a band width in orders of magnitude**, `10×` or
`100×` ([orderByFrequency](presentation.md#orderbyfrequency)). The percentages are not shown there because they would misdescribe
the rule: one decade is +900%, and a button reading "900%" explains nothing.

## sectionHeadersStateRange

**Section headers state range and count**

Goals: goals.sizeIntuition

Each section shows its range in the currently selected metric and the number of species in it.

Rationale: under per-section scaling (scaleToggle) the header is what tells the user that the scale has
changed between one section and the next. Without it, per-section scaling is misleading.

