# bbl — provenance

Vendored copy of the USGS Bird Banding Laboratory's species list, BirdView's source of four-letter
banding codes. See [imports.md](../../../imports.md#bbl) for the adoption record.

## What was retrieved

| File | Origin | SHA-256 |
| --- | --- | --- |
| `//third_party/bbl/speclist.html` | https://www.pwrc.usgs.gov/BBL/Bander_Portal/login/speclist.php | `a9a1224f4c4bb5862540512d6a55dc118f3a915f2b4979e2dc0ae13996a45b79` |

- **Size:** 550,213 bytes, UTF-8 HTML.
- **Retrieved:** 2026-10-27, by `curl` with a browser user agent (the USGS site refuses curl's
  own). The page carries per-request markup (analytics and a challenge script), so a second fetch
  hashes differently; the table itself is what is read, and this copy is what is checked.
- **License:** U.S. public domain. Quoted from
  <https://www.usgs.gov/information-policies-and-instructions/copyrights-and-credits>, read
  2026-10-27: "USGS-authored or produced data and information are considered to be in the U.S.
  Public Domain." The page asks that credit be given ("Credit: U.S. Geological Survey").

## What the file contains

One table of 1,227 rows: species number, alpha code, common name, band size, French name,
scientific name, threatened or endangered status, comments, and taxonomic order. BirdView reads
the alpha code, the common name, and the scientific name, and skips rows that are not one species
(unidentified pairs, hybrids, intergrades): 1,213 rows remain.

The page says of its codes: "The alpha codes provided here are required for reporting bandings to
the Bird Banding Laboratory in Bandit (preferred) or Band manager. They are not required or
recommended for any other purpose." That is advice, not a condition of use. It also mentions a
second list, of gallinaceous and other birds not banded with federal bands, whose link is
commented out of the page and whose address no longer serves it; those birds have no code here.
