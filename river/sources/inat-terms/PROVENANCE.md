# inat-terms — provenance

Vendored copies of iNaturalist's terms of use, community guidelines, and the definition of
machine generated content that the guidelines link to. See [../../sources.md](../../sources.md)
for the adoption record and [requirements.md](requirements.md) for the transcription.

## What was retrieved

| File | Origin | SHA-256 |
| --- | --- | --- |
| `//third_party/inat-terms/terms.html` | <https://www.inaturalist.org/pages/terms> | `de9737c4108f4058f0f6173ed8ea95894a6d56a224a5e245ae411538a0154afb` |
| `//third_party/inat-terms/community-guidelines.html` | <https://www.inaturalist.org/pages/community+guidelines> | `2f32109e84002e0208ecb86aa18331ed6b62d849c1729a9c4e816690626044b8` |
| `//third_party/inat-terms/machine-generated-content.html` | <https://www.inaturalist.org/pages/machine_generated_content> | `57ebd02cf03f0eebc5823ddea52e7299b086acf0ff54d6b719a666f8fc642160` |

- **Retrieved:** 2026-08-09, **by the maintainer, saved from a browser.**

## Why this copy was not fetched for BirdView

**Copied from the maintainer's `birdsync` repository**, which vendored these on 2026-08-09; every
SHA-256 above was verified byte-for-byte against the hashes recorded there. `birdsync` is subject
to these documents far more heavily than BirdView is — it writes to a user's account — so its
transcription is the more exercised reading, and BirdView's narrower one is below.

`www.inaturalist.org/pages/*` answers an automated request with HTTP 403 and a JavaScript
challenge, so an agent cannot re-retrieve these. Working around that was rejected; see
[../inat-api/PROVENANCE.md](../inat-api/PROVENANCE.md).

## Why BirdView vendors terms it is barely subject to

BirdView contributes nothing to iNaturalist, so most of what these documents govern cannot reach
it. They are vendored anyway for one specific finding that a link on every detail page depends on:
**nothing in them restricts reusing a taxon identifier or linking to a taxon page.** The
CC BY-NC clause governs *Content a user contributes* — observations, photographs, text. The
taxonomy is not that.

That finding is a negative, and a negative asserted from memory is worth nothing. Hashed copies
make it checkable: anyone can search these files and see for themselves what is and is not there.

## Refreshing

Re-save from a browser, compare the hash, read the diff. If `birdsync` refreshes first, copy from
there and verify the hash, as this copy did.

## Active content removed

These files are the documents **with `<script>`, `<style>` and `<link>` elements removed**, for
the credential-leak reason recorded in [../inat-api/PROVENANCE.md](../inat-api/PROVENANCE.md).
Nothing else was altered.

Verified on copy into BirdView: zero `<script>`, `<style>` or `<link>` elements and no
`AIza`-prefixed string in any of the three files.

| File | SHA-256 as served | SHA-256 as stored |
| --- | --- | --- |
| `terms.html` | `3ce69d6eac87fcc93ccca8b67a4b97f93991291ca2509467781ab8adc593122c` | `de9737c4108f4058f0f6173ed8ea95894a6d56a224a5e245ae411538a0154afb` |
| `community-guidelines.html` | `dc84d9b271dabf10dc2c52512e50a5ac2bb4ff23711a36cb73ac26282eb5ce84` | `2f32109e84002e0208ecb86aa18331ed6b62d849c1729a9c4e816690626044b8` |
| `machine-generated-content.html` | `b2453ade68999662bb893f8b1b534b040545821a860f053670b8d88d8ae38cbc` | `57ebd02cf03f0eebc5823ddea52e7299b086acf0ff54d6b719a666f8fc642160` |

As-served hashes are carried over from `birdsync`'s record and have **not** been re-verified here,
because the unstripped files were never in this repository. The as-stored hashes have been.

The paths above are anchored at the repository root with `//` so `river/check/pins.py` can
verify them (AC-003).
