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
[requirements.md](../../requirements.md#wcag). BirdView makes no broader conformance claim, and saying so is the
point: the earlier design claimed "WCAG AA ≥ 7:1", which conflates AA (4.5:1) with AAA (7:1) and
asserted a conformance level nothing checked. This requirement is checkable and scoped.

## textContrastFloor

**Link and text colors meet a stated contrast floor**

Subject: `text.contrast.ratio`
Goals: goals.sizeIntuition

Every text color the stylesheet declares for text on a bare surface clears **7:1** (WCAG 2.2
SC 1.4.6, Contrast Enhanced) against **both** the page background and the panel background. Links
and ordinary text are held to the same floor, which is also the floor badgeContrast holds badges to.

**"Text color" means `fill` as well as `color`.** SVG text is painted with `fill`, so a diagram
label is text and is held to this floor exactly like a paragraph. Shapes are not: a stroke or a
fill on a line, a polygon or the mass cube is a graphical object, governed by WCAG 1.4.11's 3:1
rather than by this requirement. What separates the two is whether the rule styles a text element,
which the stylesheet already declares by giving that element a font property.

That distinction went unwritten, and the check read `color` alone. The wing chord and wingspan
labels of [product.labeledDetailDiagram](../product/measurements.md#labeleddetaildiagram) shipped painted `#0ea5e9` — **6.44:1 on the page background
and 5.28:1 on a panel** — and nothing saw them, because SVG text does not use `color`.

Rules that set their own background are excluded and governed by badgeContrast instead: badges and chips
pair a generated background with a matching text color, and the pair is what has to clear 7:1.

**No link may go uncolored.** A link with no rule inherits the browser default — `#0000EE`
unvisited and `#551A8B` visited — which on this background is 1.90:1 and 1.62:1.

Rationale: contrast was adopted for badges only, because badge colors are *generated* and an
unchecked generator is obviously dangerous. Hand-written CSS turned out to be just as dangerous and
had no check at all: the iNaturalist link of product.inaturalistLink shipped with no color rule and was effectively
invisible.

One floor rather than two. The first version of this requirement set 4.5:1 for ordinary text,
because `--dim` sat at 6.96:1 on the page background and 5.71:1 on a panel and repainting the
interface's secondary text looked disproportionate to a single broken link. The owner then reported
the dim text as hard to read too, which settled it: `--dim` and `--accent` moved, and the exception
went with them. See [decisions.oneContrastFloor](../../decisions.md#onecontrastfloor).

---
