# avonet — provenance

Vendored copy of the AVONET morphometric dataset. See [../../sources.md](../../sources.md#avonet)
for the adoption record.

## What was retrieved

| File | Origin | SHA-256 |
| --- | --- | --- |
| `//AVONET Supplementary dataset 1.xlsx` | https://doi.org/10.6084/m9.figshare.16586228 (v7) | `eb645e83dddb40f1654a3e8d721998dbca76eff540231b7b809267c1e96f8d3e` |

- **Size:** 21,524,673 bytes
- **Retrieved:** 2026-09-20, via a figshare private share link
  (`figshare.com/s/b990722d72a26b5bfead?file=34480856`). The canonical public deposit is the DOI
  above, whose file listing matches these bytes; cite the DOI, not the share link.
- **License:** CC BY 4.0, recorded on the deposit and confirmed through the figshare API.
- **Citation:** Tobias, J. A. et al. 2022. AVONET: morphological, ecological and geographical data
  for all birds. *Ecology Letters* 25: 581–597.

## Why the file lives at the repository root

The path above is anchored at the repository root with `//` because the workbook is a build input
read by the ingest command, not a rules document. Moving it under this directory would
misrepresent what it is. The location exception is recorded as a project binding in
[../../tech.md](../../tech.md#location-exception-the-avonet-workbook).

This directory exists so that nobody helpfully copies the workbook into it and creates a second
copy to drift.

## What the file contains

Sheets: `Metadata`, `AVONET1_BirdLife`, `AVONET2_eBird`, `AVONET3_BirdTree`, `Pigot2020_BirdTree`,
`AVONET_Raw_Data`, `Mass_Sources`, `Data_Sources`, `Measurers`, `BirdLife-eBird crosswalk`,
`BirdLife–BirdTree crosswalk`.

BirdView reads `AVONET2_eBird` only — 10,661 species rows in eBird/Clements taxonomy, with
`Species2`, `Family2` and the trait columns named in P-003. Trait coverage on that sheet is
complete for the families checked: zero missing values for culmen, bill depth, tail, wing chord,
secondary or mass.

**Three things it does not contain**, each of which an earlier design assumed it did:

- No total body length.
- No wingspan. `Wing.Length` is wing *chord* — the folded wing from the carpal flexure to the tip
  of the longest primary — which is not a wingspan and is not convertible to one without a model.
- No English common name, on any sheet.

## Deliberately not vendored

**Supplementary dataset 2** from the same deposit: 9,891 duplicate specimen measurements for
repeatability analysis, keyed on `Unique.Specimen.Identifier`. It carries no species-level rows
and neither of the macroscopic dimensions BirdView needs. It was present early in the project and
was removed. Re-obtainable from the same DOI.

## Refreshing

Re-download from the DOI, compare the hash, read the diff. A new AVONET version may revise
taxonomy, in which case P-006's synonym table needs revisiting.
