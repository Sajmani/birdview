# wiley — transcribed requirements

Tier: **mandatory**. Scope: the Wiley Online Library Terms of Use, as they govern any use of
Ricklefs 2017 (*Ecology* 98:1472, `10.1002/ecy.1783`) — the only source of measured avian total
length this project identified.

**Outcome: the source is rejected.** Four clauses each independently prohibit what BirdView would
do with it. See [PROVENANCE.md](PROVENANCE.md) for what was retrieved and why the document itself
is not vendored here.

IDs are assigned by this transcription; the page has no numbering. Each entry quotes the source so
the reading can be checked.

---

### `wiley/R1` — No storage in a retrieval system, and non-profit status is explicitly irrelevant

> "Except as expressly provided above or with respect to material published on an open access
> basis under a Creative Commons license for which the license terms apply (including the
> requirement for attribution), Users **may not**:
>
> copy, distribute, transmit or otherwise reproduce, sell or resell material from Electronic
> Product(s);
>
> store such material in any form or medium in a retrieval system;
>
> download and/or store an entire issue of an Electronic Product or its equivalent;
>
> or transmit such material, directly or indirectly, for use in any paid service such as document
> delivery or list serve, or for use by any information brokerage or for systematic distribution,
> whether or not for commercial or non-profit use or for a fee or free of charge."
> — Rules of Use, B

**For BirdView: fatal.** A built dataset holding Ricklefs lengths is material "stored in a
retrieval system", and shipping it is "systematic distribution". The final clause closes the exit
this project might otherwise have taken: T-011's non-commercial posture is expressly no defense —
"whether or not for commercial or non-profit use or for a fee or free of charge".

### `wiley/R2` — No integration with other material, and no derivative works

> "Users may not integrate material from the Electronic Product(s) with other material or
> otherwise create derivative works (including resulting from the use of artificial intelligence
> technologies and tools) in any medium or override any security feature or exclusionary protocol.
> However, brief quotations for purposes of comment, criticism or similar scholarly purposes are
> not prohibited herein."
> — Rules of Use, C

**For BirdView: fatal, and precisely on point.** Joining Ricklefs lengths onto AVONET traits to
produce `species.json` is the definition of integrating this material with other material and
creating a derivative work.

The final sentence is what permits this transcription to exist: brief quotation for comment is
carved out, and that is all this file does.

### `wiley/R3` — No automated access of any kind

> "Users shall not:
>
> Use any scraper, crawler, spider, robot or other automated means of any kind to access or copy
> this site, Wiley Online Library, the Electronic Products, or any material included therein,
> deep-link to any feature or content on Wiley Online Library... bypass our robot exclusion
> headers or other measures we may use to prevent or restrict access"
> — Rules of Use, G

**For BirdView: fatal to the mechanism.** T-004's fetch-at-build design — the arrangement that
lets the ABA checklist be used without redistributing it — is "automated means" and is
prohibited. Note also that the HTTP 403 this project met repeatedly is the "robot exclusion...
measures" the clause refers to, so the block was not incidental.

### `wiley/R4` — No incorporation into any repository

> "Users shall not have the right to incorporate any material from the Electronic Products into
> any institutional or other repository under any circumstances without Wiley's prior written
> approval"
> — Rules of Use, F

**For BirdView:** forecloses committing the data even if the other clauses somehow did not apply.
"Under any circumstances" leaves no room for a non-commercial reading.

### `wiley/R5` — AI tools may not ingest or process the material

> "Use or enable artificial intelligence technologies and tools to ingest, train, test, analyze,
> process, copy, distribute, make publicly accessible, and/or generate output (including resulting
> from the use of artificial intelligence technologies and tools) based on Wiley Online Library,
> the Electronic Products, or any material included therein, or upload, copy, or host the Wiley
> Online Library, the Electronic Products, or any material included therein to or for use in
> artificial intelligence technologies and tools directly or indirectly"
> — Rules of Use, G, second bullet

**For BirdView:** this project is built with an AI coding agent, so the clause is directly
relevant to how the work is done, not only to what ships. No agent on this project should ingest
or process the Ricklefs article or its data. The agent read the Terms of Use themselves, from a
copy a human retrieved and supplied, for the purpose of determining compliance — which is the
brief-quotation-for-comment case carved out in `wiley/R2`, and is the last contact this project
should have with Wiley material.

### `wiley/R6` — What *is* permitted, and why it does not help

> "Customers and their Authorized Users may download, search, retrieve, display and view, copy and
> save to a secure network or other electronic storage media and store or print out single copies
> of individual articles or items for their own personal use, scholarly, educational or scientific
> research or internal business use... In addition, Authorized Users have the right to use, with
> appropriate credit, figures, tables and brief excerpts from individual articles in the Licensed
> Electronic Product(s) in their own scientific, scholarly and educational works."
> — Rules of Use, A.2

**For BirdView:** a person may read the paper and use its tables in their own scholarly work. That
is a genuine and generous permission, and it is not what BirdView needs. BirdView needs to extract
a data column, join it to another dataset, store the result, and distribute it — each step
prohibited above.

### `wiley/R7` — Text and data mining requires a separate agreement

> "If expressly permitted by Wiley under a relevant License or by statutory rights under applicable
> legislation, users who wish to engage in text and data mining of the Electronic Products for
> non-commercial, scholarly research purposes may do so only by using Wiley Online Library and
> consenting to and in accordance with Wiley's Text and Data Mining Agreement. Wiley reserves all
> rights to make reproductions and extractions for text and data mining."
> — Rules of Use, H

**For BirdView:** the TDM endpoint Crossref advertises for this article is not an open door. It
requires a license or statutory right plus assent to a separate agreement, and even then governs
mining rather than redistribution.

---

## Not adopted

Subscription, pay-per-view and e-commerce terms; warranty limitations; the privacy policy. None
reach a project that will not be accessing the platform.

## What this does not decide

Whether the *measurements themselves* are protectable is a separate question from these terms.
These are contract terms accepted "by using Wiley Online Library"; a US court might well hold that
individual measurements are uncopyrightable facts. This project does not rely on that argument,
because nobody here is qualified to make it, and because the contract binds independently of
whether copyright does.

The clean path to this data, if it is ever wanted, is the author or ESA — not the platform. See
[CR-002](../../decisions.md#cr-002).
