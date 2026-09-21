# Data Provenance

Source records for the input datasets consumed by `birdview-ingest`. Checksums
let a future run confirm it is working against the same bytes that were
inspected here.

## `AVONET Supplementary dataset 1.xlsx`

- **Source:** https://figshare.com/s/b990722d72a26b5bfead?file=34480856 — a
  private share link. The canonical public record is
  https://figshare.com/articles/dataset/16586228, DOI
  `10.6084/m9.figshare.16586228.v7`, whose file listing matches these bytes
  exactly. Cite the canonical DOI, not the share link.
- **Dataset:** AVONET — morphological, ecological and geographical data for all
  birds (Tobias et al. 2022, *Ecology Letters* 25:581–597), supplementary
  dataset 1.
- **License:** **CC BY 4.0** (https://creativecommons.org/licenses/by/4.0/), as
  recorded on the figshare deposit. Redistribution is permitted with
  attribution, so this file may be committed to a public repository.
- **Retrieved:** 2026-09-20
- **Size:** 21,524,673 bytes
- **SHA-256:** `eb645e83dddb40f1654a3e8d721998dbca76eff540231b7b809267c1e96f8d3e`
- **Sheets present:** `Metadata`, `AVONET1_BirdLife`, `AVONET2_eBird`,
  `AVONET3_BirdTree`, `Pigot2020_BirdTree`, `AVONET_Raw_Data`, `Mass_Sources`,
  `Data_Sources`, `Measurers`, `BirdLife-eBird crosswalk`,
  `BirdLife–BirdTree crosswalk`.
- **Notes:** The filename as downloaded is "dataset 1", not "data 1". There is
  no `AVONET1_PEL` sheet, and no sheet carries an English common-name column;
  the trait columns named in the spec (`Beak.Length_Culmen`, `Beak.Depth`,
  `Beak.Width`, `Tail.Length`, `Wing.Length`, `Mass`) are present on
  `AVONET1_BirdLife` / `AVONET2_eBird`.

## `Supplementary dataset 2.xlsx`

- **Source:** https://figshare.com/s/b990722d72a26b5bfead — the same figshare
  deposit as supplementary dataset 1 above, a different file within it. (The
  per-file id recorded for dataset 1, `file=34480856`, addresses that file only;
  no per-file id was captured for this one.)
- **Dataset:** AVONET supplementary dataset 2 (Tobias et al. 2022, *Ecology
  Letters* 25:581–597).
- **License:** **CC BY 4.0**, same deposit as dataset 1 above.
- **Retrieved:** 2026-09-20
- **Size:** 1,241,301 bytes
- **SHA-256:** `b4eeaad59841f8fb35e458dea830503f173dd89f298cc6007b2c88493d8ec490`
- **Sheets present:** `Metadata`, `Duplicate measurements`.
- **Notes:** Repeatability data — 9,891 duplicate measurements of individual
  specimens (keyed on `Unique.Specimen.Identifier`, with `Measurer` and
  `Avonet.Protocol.Used`), intended for measurement-error analysis rather than
  species-level traits. Same eight morphometric columns as dataset 1; no
  species-level rows, no total length, no wingspan. Not read by the pipeline.

## `avian_ssd_jan07.txt` — no longer used, record retained

Removed from the working directory when the project was restructured. It is kept
here because the record explains why: the dataset does not contain the two
fields it was brought in for. Do not re-acquire it expecting total length or
wingspan.

- **Source:** https://esapubs.org/archive/ecol/E088/096/avian_ssd_jan07.txt
- **Dataset:** Lislevand, Figuerola & Székely, *Avian body sizes in relation to
  fecundity, mating system, display behavior, and resource sharing* (Ecological
  Archives E088-096).
- **Retrieved:** 2026-09-20
- **Size:** 848,026 bytes
- **SHA-256:** `351c0e72dc44d30404b967ce6cc73c005d3f64e6ad56ed16c5124bf8bec18ef3`
- **Format as downloaded:** tab-separated, CRLF line endings, pure ASCII (no
  Latin-1 high bytes), `-999` as the missing-value sentinel, 3,769 species rows.
- **Notes:** Carries `Family`, `Species_number`, `Species_name`, `English_name`,
  `Subspecies`, then mass / wing / tarsus / bill / tail measurements split by
  sex (`M_`, `F_`, `unsexed_`) each with an `_N` sample-size column, plus
  `Clutch_size`, `Egg_mass`, `Mating_System`, `Display`, `Resource`,
  `References`. It contains **no total-length and no wingspan columns**, so the
  `M_length` / `Span` / `Wingspan` fields named in spec §2.2 do not exist in
  this source.
