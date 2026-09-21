# aba — provenance

**Nothing is vendored here, and that is deliberate.** See
[../../sources.md](../../sources.md#aba) for the adoption record.

## Why this directory is empty of documents

The ABA Checklist is all rights reserved. Its own CSV says so in row 0, quoted verbatim:

> "The ABA Checklist is a copyrighted work owned by American Birding Association, Inc. and cannot
> be reproduced without the express written permission of American Birding Association, Inc."

So the file is fetched onto the building machine at build time, used, and never committed
(T-004). This directory exists to say that out loud, so nobody resolves a missing input by
helpfully checking the CSV in.

## What is fetched

| Item | Value |
| --- | --- |
| **Landing page** | https://www.aba.org/aba-checklist/ |
| **Download** | `https://www.aba.org/wp-content/uploads/2026/03/ABA_Checklist-8.19.csv.zip` |
| **Version** | 8.19, January 2026 — 1,161 species on the main list |
| **Extracted file** | `ABA_Checklist-8.19.csv`, 81,041 bytes, 1,301 lines |
| **Expected SHA-256** | `9c9539b70e3d2ce2d2d46eb9354528e649e999990ec2ab416c1db222b86331d3` |
| **Destination** | `fetched/`, which `.gitignore` excludes |

**The hash above has not been verified by this project.** It was reported by the research pass
that established the licensing position, not computed from a file this repository fetched. Treat
it as expected-value-to-confirm on first retrieval, and correct this record if it differs.

## Shape of the file

Six columns, **no header row**:

| Col | Content |
| --- | --- |
| 1 | blank for species rows; non-blank marks a family section header |
| 2 | English common name |
| 3 | French common name |
| 4 | Scientific name |
| 5 | Four-letter banding code |
| 6 | ABA rarity code, 1–6 |

Family is **not** a per-species column. It appears in 99 section-header rows formatted
`Ducks, Geese, and Swans (Anatidae)`, which a forward-fill parse recovers — and which supplies
both the family scientific name for joining and the family common name for badges (P-004).

Row 0 is the copyright blob quoted above. The final row is a footnote.

Rarity distribution across the 1,161: code 1 = 490, 2 = 258, 3 = 93, 4 = 124, 5 = 151, 6 = 45.
BirdView keeps codes 1–3, which is 841 species (P-001).

## Taxonomy

ABA follows AOS/NACC; AVONET's `AVONET2_eBird` sheet follows eBird/Clements. The CSV's own note:

> "Note that where discrepancies in names occur between Clements/eBird and the AOS, both are
> listed, that of AOS in parentheses."

Measured join rate on scientific binomial, after stripping parentheticals: **99.4%** (1,154 of
1,161). All seven misses are genus-level disagreements — *Dryobates* vs *Leuconotopicus* for four
woodpeckers, *Milvago* vs *Daptrius*, *Coccothraustes* vs *Hesperiphona*, *Basileuterus* vs
*Euthlypis* — which P-006's synonym table covers.

## Open action

Ask `info@aba.org` for written permission to redistribute the checklist in a non-commercial,
attributed, open-source project. Permission would let T-005 relax and make a clean checkout
buildable with no network access. Recorded in [../../decisions.md](../../decisions.md#work-arising).
