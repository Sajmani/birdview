# inat-api — provenance

Vendored copy of iNaturalist's API recommended practices. See
[../../sources.md](../../sources.md) for the adoption record and
[requirements.md](requirements.md) for the transcription.

## What was retrieved

| File | Origin | SHA-256 |
| --- | --- | --- |
| `//third_party/inat-api/api-recommended-practices.html` | <https://www.inaturalist.org/pages/api+recommended+practices> | `ffb53ae11e1cce9c8bc2cef7743d478318ae7f9525eccde68d56d9eab4f7dbd3` |

- **Retrieved:** 2026-08-09, **by the maintainer, saved from a browser.**
- **Page's own revision line:** "Revised on February 27, 2025 06:23 PM by kueda".

## Why this copy was not fetched for BirdView

It was **copied from the maintainer's `birdsync` repository**, which vendored it on 2026-08-09,
and the SHA-256 above was verified byte-for-byte against the hash recorded there. The two
projects are subject to the same document, so re-retrieving it would gain nothing and the hash
proves the copy is faithful.

Re-retrieval is in any case not available to an agent. `www.inaturalist.org/pages/*` answers an
automated request with HTTP 403 and a JavaScript challenge — confirmed again for BirdView on
2026-09-23, when `allaboutbirds.org` and `birds.cornell.edu` were found to behave the same way.

Working around the challenge was rejected: evading a service's bot protection in order to import
its own rules for being a well-behaved client would be self-defeating. Writing the limits from
memory was also rejected — a plausible invented number is indistinguishable from a real one once
it has a requirement ID beside it.

Note that `api.inaturalist.org` — the API itself, which BirdView actually calls — serves an
identified automated client normally. It is the human-facing `pages/*` documents that are gated.

## Refreshing

Re-save from a browser, compare the hash, and read the diff. The page carries its own revision
date, which is the quickest way to tell whether anything moved. If `birdsync` refreshes first,
copy from there and verify the hash, as this copy did.

## Active content removed

The `.html` file here is the document **with `<script>`, `<style>` and `<link>` elements
removed**. Nothing else was altered: text, structure and attributes are as served.

This is not cosmetic, and the reason is recorded in `birdsync` rather than discovered twice: a
browser-saved iNaturalist page embeds the site's Google Maps JavaScript API key in a
`<script src=...>` tag, and committing it published a third party's credential and tripped
GitHub secret scanning. The key belongs to iNaturalist and appears in every page they serve, so
nothing became public that was not already — but redistributing someone else's credential is not
this repository's business, and a terms snapshot has no use for scripts.

Verified on copy into BirdView: zero `<script>`, `<style>` or `<link>` elements, and no
`AIza`-prefixed string, in this file or in `inat-terms/`.

| File | SHA-256 as served | SHA-256 as stored |
| --- | --- | --- |
| `api-recommended-practices.html` | `04bba1540762ab1c80404bea68f133dec9b6b1152d2cf0f3a0e9e33ff7e2aa62` | `ffb53ae11e1cce9c8bc2cef7743d478318ae7f9525eccde68d56d9eab4f7dbd3` |

The as-served hash is carried over from `birdsync`'s record and has **not** been re-verified
here, because the unstripped file was never in this repository. The as-stored hash has been.

The paths above are anchored at the repository root with `//` so `spec/check/pins.py` can
verify them (AC-003).
