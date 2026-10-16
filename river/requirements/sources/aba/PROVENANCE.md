# aba — provenance

**Nothing is vendored here, and that is deliberate.** See
[../../requirements.md](../../requirements.md#aba) for the adoption record.

## Why this directory is empty of documents

The ABA Checklist is all rights reserved. Its own CSV says so in row 0, quoted verbatim:

> "The ABA Checklist is a copyrighted work owned by American Birding Association, Inc. and cannot
> be reproduced without the express written permission of American Birding Association, Inc."

So the file is fetched onto the building machine at build time, used, and never committed
(tech.fetchRestrictedInputs). This directory exists to say that out loud, so nobody resolves a missing input by
helpfully checking the CSV in.

## What is fetched

| Item | Value |
| --- | --- |
| **Landing page** | https://www.aba.org/aba-checklist/ |
| **Download** | `https://www.aba.org/wp-content/uploads/2026/03/ABA_Checklist-8.19.csv.zip` |
| **Version** | 8.19, January 2026 — 1,161 species on the main list |
| **Archive** | 30,951 bytes · SHA-256 `f036c47429585a00ff16417fc4d6d152fd3b600044e7b283c6371718d7c91e07` |
| **Extracted file** | `ABA_Checklist-8.19.csv`, 81,041 bytes, 1,301 CSV rows |
| **SHA-256** | `9c9539b70e3d2ce2d2d46eb9354528e649e999990ec2ab416c1db222b86331d3` |
| **Encoding** | UTF-8, no BOM, CRLF line endings. Non-ASCII occurs only in the French name column. |
| **Destination** | `fetched/`, which `.gitignore` excludes (verified: rule at `.gitignore:3`) |

**Verified 2026-09-21** by fetching the archive with `curl` and hashing both files. The CSV hash
matches the value an earlier research pass reported, which is now confirmed rather than assumed.

Retrieval is permitted by `https://www.aba.org/robots.txt`, checked the same day: its `Disallow`
rules cover WooCommerce upload paths, `add-to-cart` query strings and `/wp-admin/`, and a later
`User-agent: * / Disallow:` block allows everything else. `/wp-content/uploads/2026/` is not
excluded.

Two artifacts of the archive worth knowing before writing the ingest: it unpacks a `__MACOSX/`
sidecar directory alongside the CSV, and the CSV's final line ends with a bare LF where every
other line ends CRLF.

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
both the family scientific name for joining and the family common name for badges (product.namesFromAvilist).

Row 0 is the copyright blob quoted above. The final row is a footnote. Together with the 99 family
headers those make up the 101 rows whose first column is non-blank, so "first column non-blank"
alone is not a family-header test — the pattern `... (Familyname)` is.

Every one of the 1,301 rows has exactly 6 fields, so a parser may rely on the column count.

Rarity distribution across the 1,161 species rows: code 1 = 490, 2 = 258, 3 = 93, 4 = 124,
5 = 151, 6 = 45. BirdView keeps codes 1–3, which is 841 species (product.scopeIsRecurringAvifauna).

**All of the above was measured from the fetched file on 2026-09-21**, not taken from the
publisher's description.

## Taxonomy

ABA follows AOS/NACC; AVONET's `AVONET2_eBird` sheet follows eBird/Clements. The CSV's own note:

> "Note that where discrepancies in names occur between Clements/eBird and the AOS, both are
> listed, that of AOS in parentheses."

**Measured 2026-09-21, against `AVONET2_eBird` itself** — the join BirdView actually performs:

| Set | Joined on binomial | Rate |
| --- | --- | --- |
| All 1,161 ABA species | 1,094 | 94.23% |
| Codes 1–3, the product.scopeIsRecurringAvifauna scope | 819 of 841 | **97.38%** |

An earlier research pass reported 99.4%, but that measured ABA against the **eBird taxonomy
v2025** CSV. AVONET was published in 2022 and carries the eBird/Clements taxonomy of that era, so
four years of taxonomic revision sit in the gap. The figure that matters is 97.38%, and it will
decay further with each ABA and eBird revision.

The 22 in-scope misses fall into three kinds, which matter very differently:

- **Apparent renames (12)** — the pre-revision name is in AVONET and is not a separate ABA
  species: the four *Anarhynchus* plovers (ex-*Charadrius*), *Astur cooperii* and
  *Astur atricapillus*, *Ardea ibis*, *Botaurus exilis*, *Gygis candida*, *Pterodroma deserta*,
  *Spilopelia chinensis*, *Tyto furcata*.
- **Splits (9)** — the parent is *still* a separate ABA species, so mapping the child onto it
  makes two species share one set of measurements: *Anthus japonicus*, *Calonectris borealis*,
  *Larus smithsonianus*, *Larus vegae*, *Numenius hudsonicus*, *Setophaga aestiva*,
  *Sturnella lilianae*, *Sula brewsteri*, *Vireo swainsoni*. *Larus smithsonianus* and
  *Larus vegae* both resolve to *Larus argentatus*.
- **Unresolved (1)** — *Ramosomyia violiceps*. **Resolved 2026-09-21:** `AVONET2_eBird` calls it
  *Leucolia violiceps*; the earlier lookup guessed *Amazilia violiceps*, which is the BirdLife and
  BirdTree name.

Searching all three AVONET taxonomy sheets rather than `AVONET2_eBird` alone shows that only
**6** of the 22 are absent from AVONET entirely — *Larus vegae*, *Numenius hudsonicus*,
*Sturnella lilianae*, *Vireo swainsoni*, *Sula brewsteri* and *Anthus japonicus*. See
[decisions.abaSpeciesWithoutAvonet](../../decisions.md#abaspecieswithoutavonet) for the full disposition and the mass caveat.

**The rename/split classification above is provisional.** It was derived mechanically from
whether the parent name still appears in ABA 8.19, and that test under-detects: *Astur
atricapillus* is a split from Eurasian Goshawk, and looks like a rename only because the Eurasian
bird is not on the ABA list. Each entry needs a human taxonomic check before it enters the mapping
table. See [decisions.abaSpeciesWithoutAvonet](../../decisions.md#abaspecieswithoutavonet).

## Open action

Ask `info@aba.org` for written permission to redistribute the checklist in a non-commercial,
attributed, open-source project. Permission would let tech.datasetNeverCommitted relax and make a clean checkout
buildable with no network access. Recorded in [../../decisions.md](../../decisions.md#work-arising).
