package tech

# Accessibility

## badgeContrast

**Badge contrast meets WCAG 2.2 SC 1.4.6**

Subject: `badge.contrast.ratio`
Value: `>= 7.0`
Goals: goals.comparisonAcrossFamilies

Every generated badge color yields a contrast ratio of at least 7:1 against **the text it actually
carries**, which is black or white per badge, computed per WCAG 2.2 relative luminance.

Requiring black text specifically would be a stricter rule than WCAG asks for, and an expensive
one: it discards the entire dark half of the gamut and costs about a third of the distinguishable
colors available to product.generatedBadgeColors.

Rationale: WCAG is adopted as a **mandatory** source scoped to the contrast criteria only — see
[sources.md](../../requirements.md#wcag). BirdView makes no broader conformance claim, and saying so is the
point: the earlier design claimed "WCAG AA ≥ 7:1", which conflates AA (4.5:1) with AAA (7:1) and
asserted a conformance level nothing checked. This requirement is checkable and scoped.

