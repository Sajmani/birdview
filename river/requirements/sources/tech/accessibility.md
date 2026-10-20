package tech

# Accessibility

## badgeContrast

**Badge contrast meets WCAG 2.2 SC 1.4.6**

Subject: `badge.contrast.ratio`
Value: `>= 7.0`
Goals: goals.openToEveryone

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
Goals: goals.openToEveryone

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

**A rule that sets its own background is held to 7:1 against that background**, not the page's.
A badge's background and text color are generated together and set inline, so that pair is
badgeContrast's; a pair written in the stylesheet, such as a selected button's dark text on
`--accent`, is this requirement's.

That sentence used to exempt every rule with its own background, on the reasoning that badges and
chips are generated pairs. The selected button and the pick control are not, and nothing read
them: text on a selected button at 2.42:1 passed every stylesheet check and was caught only by
rendering the page (app.renderedTextContrast).

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

## webPages

**What counts as a BirdView web page**

Subject: `pages.domain`
Binds: wcag.WebPage
Goals: goals.openToEveryone

**Every response the server sends as `text/html`, `image/svg+xml` or `text/plain`, at any route,
is a BirdView web page**, with everything drawn into it. Today that is the index, the species
page in its drawn and undrawn forms, the comparison, the guides and their index, the added-values
and excluded-birds pages, the pipeline diagram and the favicon opened at their own URLs, and the
plain-text error responses.

The list is illustrative, not the definition. Nothing in the bundle enumerates the pages: the
checks discover them from the server's routes, so a page added later is a web page the day it
exists.

**This binds WCAG's definition and does not narrow it.** WCAG makes a web page of any resource
obtained from a URI, so an error response or an SVG opened on its own is one whether BirdView
would like it to be or not. A page BirdView chose to leave out would be a page excused from a
mandatory source, and the only way to do that is an accepted gap in
[decisions.md](../../decisions.md), page by page.

Rationale: the contrast checks were written against the stylesheet, which made "the pages" mean
"what the stylesheet styles" without anyone deciding so. Mapping WCAG over the server's actual
responses found two kinds of page that meaning left out: the pipeline diagram served as its own
SVG document with its own colors, and the page templates served as raw source under `/static/`
(tech.servesOnlyItsPages).
