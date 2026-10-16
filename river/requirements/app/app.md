package app

# app — composed requirements for the BirdView web application

What the web application must satisfy, composed from BirdView's own sources and the sources it
imports. Implementations and releases of the server are verified against this package; its
criteria are in [acceptance.md](acceptance.md).

**This first composition covers accessibility only.** It is the test case for composing a
requirement over every instance of a domain. The application's other requirements are still read
from `product` and `tech` directly until they are composed here.

**Status: composed 2026-10-16 by an agent, awaiting the owner's approval.** Not tagged.

## Composition record

| Source | Tier | Pin |
| --- | --- | --- |
| [`wcag`](../sources/wcag/requirements.md) | mandatory | WCAG 2.2, W3C Recommendation of 12 December 2024; SHA-256 in [PROVENANCE.md](../sources/wcag/PROVENANCE.md) |
| [`tech`](../sources/tech/tech.md) | local | the commit this file is in |
| [`decisions`](../decisions.md) | — | the commit this file is in |

Precedence is the ranking in [requirements.md](../requirements.md): a mandatory source beats a
local one. **No conflict arose.** BirdView's local requirements only tighten what WCAG asks, which
a lower tier may always do; nothing local loosens it.

---

## pagesMeetTextContrast

**Every BirdView web page meets WCAG 1.4.6 in full, without the large-text allowance**

Composed from: wcag.ContrastEnhanced, wcag.FullPages, wcag.ContrastRatio, tech.webPages, decisions.oneContrastFloor

Every web page, as tech.webPages defines one, meets wcag.ContrastEnhanced over every part of the
page (wcag.FullPages), measured as wcag.ContrastRatio defines. All text is held to 7:1, large text
included: wcag.ContrastEnhanced allows 4.5:1 for large text, and decisions.oneContrastFloor gives
that up. Meeting this meets wcag.ContrastMinimum too.

**No page is excepted.** A page that sets no colors, such as a plain-text error response, is drawn
in the reader's own colors and meets this with nothing to measure.

## pagesMeetNonTextContrast

**Every BirdView web page meets WCAG 1.4.11 in full**

Composed from: wcag.NonTextContrast, wcag.FullPages, tech.webPages

Every web page meets wcag.NonTextContrast over every part of the page: the graphics a reader needs
to read a diagram, and the controls and their states, have 3:1 against adjacent colors.

## stylesheetContract

**The stylesheet and the palette carry text contrast for every page that uses only them**

Composed from: tech.textContrastFloor, tech.badgeContrast, wcag.ContrastRatio

The stylesheet is an intermediary: it meets pagesMeetTextContrast for any page that takes its
colors from it, so that requirement is checked once on the stylesheet rather than once per page.

- **Guarantee.** Every text color `styles.css` declares clears 7:1 against both the page and the
  panel background, or against the rule's own background where it sets one, and no link goes
  uncolored (tech.textContrastFloor). Every badge's generated
  background clears 7:1 against the text color generated with it (tech.badgeContrast).
- **Assumption.** The page links `styles.css`, has no `<style>` element, and sets a color inline
  only as a badge's generated pair. A page that breaks this is outside the contract and is
  checked on its own.
- **Scope.** Text, in pages rendered from BirdView's templates. **Not** an SVG document served at
  its own URL, which carries its own styles, and **not** non-text contrast, which this contract
  does not guarantee at all.
- **Override.** Closed. A template may not set a color except through the palette; one that does
  has left the contract, which the assumption check reports.
