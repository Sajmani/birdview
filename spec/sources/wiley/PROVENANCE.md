# wiley — provenance

The Wiley Online Library Terms of Use, consulted to decide whether Ricklefs 2017 could be used as
a source of measured avian total length. It cannot. See [requirements.md](requirements.md) for the
transcription and [../../decisions.md](../../decisions.md#cr-002) for the decision.

## What was retrieved

| Item | Value |
| --- | --- |
| **Origin** | `https://onlinelibrary.wiley.com/terms-and-conditions#vor` |
| **Retrieved** | 2026-09-21, **by the maintainer, printed to PDF from a browser** |
| **Page footer timestamp** | 9/21/26, 5:29 PM |
| **Local file** | `Wiley.pdf`, 112,989 bytes, 6 pages |
| **SHA-256** | `c2b7fd36ad1b79ca14c2a3b84106b3e4e574a9437bc0dbf86d3cb9aed4141c30` |
| **Copyright line on the page** | "Copyright © 1999-2026 John Wiley & Sons, Inc or related companies. All rights reserved, including rights for text and data mining and training of artificial intelligence technologies or similar technologies." |

The `#vor` fragment is the section Crossref names as governing this article: its metadata for
`10.1002/ecy.1783` lists exactly one license, `content-version: vor`, pointing at this page.

## Why a human retrieved it

`onlinelibrary.wiley.com` answers automated requests with HTTP 403 — confirmed against both
`/termsAndConditions` and `/terms-and-conditions` on 2026-09-21. That block is not incidental:
Rules of Use G prohibits automated access and refers to "robot exclusion headers or other measures
we may use to prevent or restrict access". Working around it to read the rules about not working
around it would be self-defeating, so the maintainer opened the page and printed it.

## Why the document is **not** vendored here

This directory deliberately contains no copy of the terms, which departs from the method's normal
practice of vendoring an adopted source.

The reason is the source's own terms. Rules of Use B forbids storing material from Wiley Online
Library "in any form or medium in a retrieval system", and F forbids incorporating it "into any
institutional or other repository under any circumstances without Wiley's prior written approval".
A git repository is both. Committing Wiley's terms of use in order to document our compliance with
Wiley's terms of use would violate them.

Rules of Use C carves out the alternative explicitly: "brief quotations for purposes of comment,
criticism or similar scholarly purposes are not prohibited herein." So `requirements.md` quotes
the governing clauses and nothing more, and the full document stays outside the repository.

The hash above is recorded so the maintainer's local copy can be identified as the one these
quotations were taken from. It is **not** an integrity pin on a vendored file, because no file is
vendored.

## Refreshing

Only necessary if someone revisits CR-002. Re-print from a browser, compare against the quotations
in `requirements.md`, and note that Wiley reserves the right to change these terms at any time:
"Wiley may modify any of these terms and conditions at any time by providing notice on Wiley
Online Library."
