package product

# Ordering and sectioning

## oneSelectedMetric

**One primary metric, user-selected**

Subject: `ordering.metric`
Value: `length | wingspan | mass | bill | tail | frequency`
Goals: goals.sizeIntuition, goals.comparisonAcrossFamilies, goals.shapeNotOnlySize

The user chooses whether the dataset is ordered by total length, wingspan, body mass, bill length,
tail length, or how often the bird is recorded ([orderByFrequency](presentation.md#orderbyfrequency), [decisions.frequencyMetric](../../decisions.md#frequencymetric)).
The comparison view offers the same size orderings for the birds it holds.

**Bill and tail are absolute lengths, not proportions.** They are measured for every species, where
total length is a sum with three segments modeled from mass; a bill or tail "for its size" was
considered and not offered, because the only plain denominator, total length, would put modeled
values under every ratio. Added 2026-10-19 at the owner's request.

## ascendingOrder

**Ascending order**

Goals: goals.sizeIntuition

Species are sorted ascending by the selected metric, **and the page says which way that
runs**: *smallest first* under length, wingspan, mass, bill and tail, *most common first* under
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

Each section shows its range in the currently selected metric and the number of species in it,
under a heading that names the ordering as Order by does and the section's number: *Tail Tier 1*
when ordering by Tail.

**The range is the tier's span, and successive tiers abut.** A tier runs from its first bird up to
where the next tier begins, so each heading ends exactly where the next one starts, written the
same way; the last tier ends at its largest bird, and a tier of one bird states one value. Every
heading on a page carries the same number of decimals, the fewest — up to three more than usual —
that keep each tier's two ends apart, since a tier of small bills can be narrower than a tenth of
an inch. Headings once gave each tier's smallest and largest bird, rounded heading by heading:
"0.2 in – 0.3 in" then "0.3 in – 0.4 in" read as an overlap and hid the gap between them.

Rationale: under per-section scaling (scaleToggle) the header is what tells the user that the scale has
changed between one section and the next. Without it, per-section scaling is misleading.

