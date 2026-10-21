package wcag

# wcag — transcribed requirements

Tier: **mandatory**. Scope: the contrast of what BirdView draws — text, and the graphics a reader
needs to read a diagram. BirdView adopts these criteria and **claims no conformance level**; the
other success criteria are not adopted (see [module.md](../../../module.md#wcag)).

Quotations are verbatim from the vendored Recommendation, with its markup removed. Each item says
what the quotation means for BirdView. Copyright © 2020-2024 World Wide Web Consortium; this file
includes material copied from *Web Content Accessibility Guidelines (WCAG) 2.2*,
<https://www.w3.org/TR/2024/REC-WCAG22-20241212/>. See [PROVENANCE.md](PROVENANCE.md).

Each success criterion below applies to every web page, in an `Applies to:` field naming
`WebPage`. BirdView says which of its responses are web pages in a requirement of its own, and
composition holds every one of them to every criterion here, without listing them.

---

## WebPage

**A web page, as WCAG defines it**

> a non-embedded resource obtained from a single URI using HTTP plus any other resources
> that are used in the rendering or intended to be rendered together with it by a user agent
> — Glossary

**For BirdView: every response a reader can open at its own URL**, together with everything drawn
into it. An SVG embedded in a guide is part of that guide's page; the same SVG opened at its own
URL is a page of its own. A plain-text error response is a page too: it is a resource obtained from
a URI using HTTP.

## FullPages

**A page meets a criterion in full, or not at all**

> Conformance (and conformance level) is for full web page(s) only, and cannot be achieved if part
> of a web page is excluded.
> — Conformance requirement 5.2.2, Full pages

**For BirdView:** BirdView makes no conformance claim, but adopts this as the reading of every
criterion below. A criterion holds for a page only if it holds for every part of it — a diagram,
a badge and a footer included. A check that covers the stylesheet and not a diagram's own colors
has not shown that the page meets the criterion.

## ContrastRatio

**How contrast is measured**

> (L1 + 0.05) / (L2 + 0.05), where L1 is the relative luminance of the lighter of
> the colors, and L2 is the relative luminance of the darker of the colors.
> — Glossary

> For the purpose of Success Criteria 1.4.3 and 1.4.6, contrast is measured with respect to the
> specified background over which the text is rendered in normal usage. If no background color is
> specified, then white is assumed.
> — Glossary, contrast ratio, note 3

> It is a failure if no background color is specified when the text color is specified, because
> the user's default background color is unknown and cannot be evaluated for sufficient contrast.
> For the same reason, it is a failure if no text color is specified when a background color is
> specified.
> — Glossary, contrast ratio, note 4

**For BirdView:** the formula `internal/palette` implements. Note 4 is a rule of its own: a page
that sets a text color must set the background behind it, and the reverse. A page that sets
neither, such as a plain-text error response, is drawn in the reader's own colors and has nothing
to measure.

## ContrastMinimum

**Text contrast of at least 4.5:1 (Level AA)**

Applies to: WebPage

> The visual presentation of text and images of text has a contrast ratio of at least 4.5:1, except
> for the following:
> Large Text
> Large-scale text and images of large-scale text have a contrast ratio of at least
> 3:1;
> — Success Criterion 1.4.3 Contrast (Minimum)

The criterion goes on to exempt incidental text and logotypes, in the same words as
ContrastEnhanced below.

**For BirdView: met by meeting ContrastEnhanced.** Every ratio here is lower than the one
ContrastEnhanced sets for the same text, so it is transcribed for completeness and needs no check
of its own.

## ContrastEnhanced

**Text contrast of at least 7:1 (Level AAA)**

Applies to: WebPage

> The visual presentation of text and images of text has a contrast ratio of at least 7:1, except
> for the following:
> Large Text
> Large-scale text and images of large-scale text have a contrast ratio of at least
> 4.5:1;
> Incidental
> Text or images of text that are part of an inactive user interface component, that
> are pure decoration, that are not visible to anyone, or that are part of a picture that contains
> significant other visual content, have no contrast requirement.
> Logotypes
> Text that is part of a logo or brand name has no contrast requirement.
> — Success Criterion 1.4.6 Contrast (Enhanced)

**For BirdView: the floor for all text, including large text.** BirdView holds every text color to
7:1 and does not use the large-text allowance, which is stricter than this criterion asks and is
BirdView's own choice ([decisions.oneContrastFloor](../../../decisions.md#onecontrastfloor)). Diagram labels are text, painted with
SVG `fill`. BirdView has no logo text and no images of text.

## NonTextContrast

**Contrast of at least 3:1 for interface components and graphics (Level AA)**

Applies to: WebPage

> The visual presentation of the following have a contrast ratio of at least 3:1 against adjacent
> color(s):
> User Interface Components
> Visual information required to identify user interface components and
> states, except for inactive components or where the appearance of the component is determined by
> the user agent and not modified by the author;
> Graphical Objects
> Parts of graphics required to understand the content, except when a particular
> presentation of graphics is essential to the information being conveyed.
> — Success Criterion 1.4.11 Non-text Contrast

**For BirdView:** the strokes and fills of the bird diagrams, the mass cube, the wing crossbar and
the boxes and arrows of the pipeline diagram are graphical objects, and controls such as the
section-step buttons are interface components. The stylesheet already distinguishes these from
text for [tech.textContrastFloor](../../../local/tech/accessibility.md#textcontrastfloor). **No check holds them to 3:1 yet.**
