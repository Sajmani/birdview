# avilist — provenance

Vendored copy of the global avian checklist. BirdView's source of **English names, family and
order names, and the per-species eBird code** (product.namesFromAvilist). See
[../../sources.md](../../sources.md#avilist) for the adoption record.

## What was retrieved

| File | Origin | SHA-256 |
| --- | --- | --- |
| `//third_party/avilist/AviList-v2025b-10Jun2026-extended.xlsx` | https://doi.org/10.2173/avilist.v2025b | `2e1fd3374e23af732b04115b033dd9d97fc53ba275c312d02ef5d12cfb85c988` |

- **Size:** 8,954,422 bytes.
- **Retrieved:** 2026-10-02, by `curl` from
  `https://www.avilist.org/wp-content/uploads/2026/06/AviList-v2025b-10Jun2026-extended.xlsx`,
  linked from the publisher's own [v2025b checklist page](https://www.avilist.org/checklist/v2025b/).
- **License:** **CC BY 4.0**, stated on that page: *"AviList: The Global Avian Checklist © 2026 by
  AviList Core Team is licensed under CC BY 4.0."* Redistribution with attribution is permitted,
  so this file is committed — which is the whole reason it can replace the ABA checklist.
- **Citation:** AviList Core Team. 2026. *AviList: The Global Avian Checklist, v2025b.*
  https://doi.org/10.2173/avilist.v2025b

**The extended file is required; the short one will not do.** The short release omits
`English_name_Clements_v2025` and `Species_code_Cornell_Lab`, which are two of the four things
BirdView reads. Checked by downloading it: 14 columns against the extended file's 26.

The file lives at the repository root under the same location exception as the AVONET workbook —
it is a build input, not a rules document. See
[../../bindings.md](../../bindings.md#location-exception-the-third_party-directory).

## What the file contains

Six sheets. `AviList v2025b` carries 33,685 rows covering every rank; filtering `Taxon_rank` to
`species` leaves **11,131 species**.

The columns BirdView reads:

| Column | Use |
| --- | --- |
| `Taxon_rank` | Row filter — only `species` is read |
| `Scientific_name` | Join key against AVONET's `Species2` |
| `English_name_Clements_v2025` | The **North American** common name (product.englishNameChoice) |
| `English_name_AviList` | The **world** common name; IOC 15.1 usage |
| `Species_code_Cornell_Lab` | The eBird species code, which is BirdView's species id and the code shown on a tile |
| `Family` / `Family_English_name` | Family scientific and English names |
| `Order` | Order name |

**`Range` is deliberately not read**, although it looks like a scope source. It is empty for
**4,992 of the 11,131 species** — including Ruddy Duck and Tundra Swan — because range statements
are populated only where the taxon concept matches Clements exactly. Measured 2026-10-02.

## Why the Clements and Cornell columns are usable here

[sources.md](../../sources.md) rejects eBird/Clements as a source, because Cornell's own terms
forbid redistribution without written authorization. That rejection governs **Cornell's
distribution under Cornell's terms**, and this is not it.

AviList is a separate publication by a consortium — the AOS, Cornell, BirdLife, Avibase and the
IOU — with Cornell staff on its core team (Marshall Iliff, Jeff Gerbracht, Tom Schulenberg), and
it is released under CC BY 4.0 in its entirety. BirdView takes these columns from AviList, under
AviList's license, with AviList's attribution. Nothing is taken from Cornell's distribution.

This is written down rather than assumed because the shape of it — *"the source we rejected,
reached through a source we accepted"* — is exactly what ought to attract scrutiny.

## Measured coverage for BirdView's scope

Counted 2026-10-02 against the 841 species then in scope:

| Outcome | Species |
| --- | --- |
| Matched on scientific name | **832** |
| …of those, carrying an eBird species code | **829** |
| …of those, carrying a Clements English name | **827** |
| Resolved by the synonym table (genus or epithet change) | **7** |
| Unresolved | **5** |

The seven synonyms are one-to-one renamings — *Coccothraustes vespertinus* →
*Hesperiphona vespertina*, four *Dryobates* woodpeckers → *Leuconotopicus*, *Icterus bullockii* →
*I. bullockiorum*, *Pica nuttalli* → *P. nuttallii*.

The five that remain are **genuine disagreements about what a species is**, not join failures, and
no table can fix them:

| Species | Why |
| --- | --- |
| Gray-headed (Purple) Swamphen | ABA splits *Porphyrio poliocephalus*; AviList lumps it into *P. porphyrio* |
| Western Warbling-Vireo | ABA splits *Vireo swainsoni*; AviList lumps it into *V. gilvus*, which is already BirdView's Eastern Warbling-Vireo — a 2:1 collision |
| Ruddy Duck | AviList's *Oxyura jamaicensis* concept has no 1:1 Clements counterpart, so no code |
| White-necked Petrel | as above |
| American Barn Owl | AviList recognizes *Tyto furcata*; Clements treats barn owls differently |

These five keep their scientific-name slug as their id, and no eBird code is shown for them.

## Refreshing

Re-download from the DOI, compare the hash, and re-run the coverage count above. Expect the
synonym table to need entries: AviList revises annually and a genus move is the common case. A
species that silently loses its code would change its URL, which is why
[acceptance.stableSpeciesId](../../acceptance.md#stablespeciesid) asserts id stability rather than trusting the join.
