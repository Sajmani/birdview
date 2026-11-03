# gbif-terms — provenance

Vendored copies of GBIF's data user agreement and citation guidelines. See
[imports.md](../../../imports.md#gbif-terms) for the adoption record and [gbifterms.md](gbifterms.md)
for the transcription.

## What was retrieved

| File | Origin | SHA-256 |
| --- | --- | --- |
| `//third_party/gbifterms/data-user.html` | <https://www.gbif.org/terms/data-user> | `cb6cedbab01babb4a0ae4f3ad3e2137b945bd5d9de993fb5fcfc85b64a863113` |
| `//third_party/gbifterms/citation-guidelines.html` | <https://www.gbif.org/citation-guidelines> | `91bf93264252f433392845e98fff252f12352a2f51bd63682a60e888fe31848f` |

- **Retrieved:** 2026-11-03, **by the maintainer, saved from a browser** as complete web pages,
  while signed in to GBIF.org. The data user agreement reads "last updated: 15 September 2026".
- **Modified** as recorded in `third_party/gbifterms/README.md`: scripts, styles, and stylesheet
  links removed, and the signed-in account name replaced. The hashes are of the modified copies.
- **Not vendored:** the pages' image and asset folders. One image carries content: the citation
  guidelines' "GBIF download citation example", a screenshot of a download page's citation, which
  the vendored HTML references but does not contain.

## Why this copy was not fetched by an agent

`www.gbif.org` answers an automated request with HTTP 403 and a Cloudflare block page ("Sorry, you
have been blocked"), with or without a browser user agent. Working around it was not attempted, as
with eBird's and iNaturalist's challenges. GBIF's API at `api.gbif.org` is not behind it.

## Refreshing

Save both pages again from a browser, strip them the same way, and replace the hashes above. A
changed "last updated" date on the agreement means re-reading [gbifterms.md](gbifterms.md) against
it.
