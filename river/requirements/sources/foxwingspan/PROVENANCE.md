# fox-wingspan — provenance

Vendored copy of the avian wingspan database, BirdView's only source of measured wingspan. See
[../../requirements.md](../../requirements.md#fox-wingspan) for the adoption record.

## What was retrieved

| File | Origin | SHA-256 |
| --- | --- | --- |
| `//third_party/foxwingspan/wingspan_fox2026_AppendixS1.xlsx` | https://doi.org/10.6084/m9.figshare.32089083 (v1) | `6b55747a5761c08356d92c7a0b524b4e586c28a25316b0a6c1de3738db1d8eda` |

- **Size:** 4,879,919 bytes · **MD5:** `4970a3506c98859c3aeb7e98e5f5ce35`, which **matches the
  `computed_md5` figshare publishes for the file** — so the download is confirmed intact against
  the publisher's own record, not merely self-consistent.
- **Retrieved:** 2026-09-21, by `curl` from `https://ndownloader.figshare.com/files/63975186`.
- **Upstream filename:** `AppendixS1_wingspan_data_all_xlsx.xlsx`, renamed on vendoring to say what
  it is.
- **License:** **CC BY 4.0**, from the figshare deposit record. Redistribution permitted with
  attribution, so this file is committed.
- **Citation:** Fox, M. R., Brunton, E. A., Shephard, J. M. (2026). Estimating avian wingspan from
  wing length: an order-level regression approach. *bioRxiv*. Dataset: "Avian wingspan database:
  empirical and predicted wingspan for 10,618 flighted bird species", published 2026-04-24.

The file lives at the repository root under the same location exception as the AVONET workbook —
it is a build input, not a rules document. See
[../../bindings.md](../../../bindings.md#location-exception-the-third_party-directory).

## What the file contains

Three sheets. `wingspan data` carries 10,618 rows and 86 columns; `metadata` describes all 85
variables; `citation` holds the reference above.

The columns BirdView reads:

| Column | Use |
| --- | --- |
| `Species1` | Join key — **BirdLife taxonomy**, matching AVONET's `AVONET1_BirdLife` sheet |
| `needs_prediction` | `0` marks an empirical value; anything else means the row is modeled |
| `wingspan_reported_mid_all` | The measured value |
| `wingspan_reported_min_all` / `_max_all` | The reported range, stored alongside per product.lengthAndWingspanSources |

**The predicted columns are deliberately not read.** `wingspan_predicted_order` and
`wingspan_predicted_global` are regressions from AVONET wing chord. Ingesting them would
reintroduce modeled-as-measured values under a source that looks authoritative, which is the error
[decisions.noOpenLengthAndWingspan](../../decisions.md#noopenlengthandwingspan) exists to prevent. BirdView models its own estimates and labels
them.

## Measured coverage for BirdView's scope

Counted 2026-09-21 against the 841 ABA code 1–3 species, joining through AVONET's BirdLife-eBird
crosswalk:

| Outcome | Species | Share |
| --- | --- | --- |
| Empirical wingspan (`needs_prediction = 0`) | **301** | 35.8% |
| Predicted only — excluded, BirdView models instead | 510 | 60.6% |
| No row in the dataset at all | 30 | 3.6% |

**All 30 absences are Galliformes.** The dataset covers "flighted bird species" and contains no
`Galliformes` order value at all, so every ABA quail, grouse, turkey, chachalaca and partridge is
missing by design rather than by join failure. Confirmed by searching the workbook's shared-string
table: no `Meleagris`, `Colinus`, `Callipepla`, `Bonasa` or `Ortalis` binomial appears anywhere in
the file.

## Join note

`Species1` is BirdLife taxonomy, while BirdView's spine is `AVONET2_eBird` (`Species2`). The join
therefore runs through the `BirdLife-eBird crosswalk` sheet inside the AVONET workbook — 11,352
pairs — rather than matching names directly. A direct name match would silently lose species
wherever the two taxonomies disagree.

## Refreshing

Re-download from the DOI, compare both hashes, and re-run the coverage count above. A new version
may move rows between empirical and predicted, which changes how many of BirdView's wingspans are
`measured` — the mix is expected to move; the provenance model is not.
