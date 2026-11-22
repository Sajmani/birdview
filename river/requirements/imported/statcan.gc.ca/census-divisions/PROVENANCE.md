# statcan-boundaries — provenance

**Nothing is vendored here.** The archive is 133 MB (high-resolution Lambert Conformal Conic
shapefile) from which BirdView derives compact per-province census-division boundary files, so it is
pinned by hash and never committed. See [imports.md](../../../imports/imports.md#statcan-boundaries)
for the adoption record and attribution requirement.

## What is fetched

| Item | Value |
| --- | --- |
| **Landing page** | https://www.statcan.gc.ca/en/subjects/standard/geography/boundary-files |
| **Download** | `https://www12.statcan.gc.ca/census-recensement/2021/geo/sip-pis/boundary-limites/files-fichiers/lcd_000b21a_e.zip` |
| **Licence** | https://www.statcan.gc.ca/en/terms-conditions/open-licence (Statistics Canada Open Licence, modified 2025-07-10) |
| **Version** | 2021 Census – Cartographic Boundary File, Census divisions (`lcd_000b21a_e`) |
| **Archive** | 139,814,225 bytes · SHA-256 `1b638cea73fbf8349ba908d86be678272b2abcc74240fb0dfd9daf7a28457ad0` |
| **Read** | `lcd_000b21a_e.shp` and `.dbf` (293 census divisions, projected in `NAD83_Statistics_Canada_Lambert` / EPSG:3347, inverse-projected to WGS84/NAD83 lat/lon) |
| **Consumer** | `cmd/birdview-boundaries -counties`, which refuses an archive with any other hash |
| **Output** | `data/county-boundaries/CAN.*.json.gz`, committed |

**Verified 2026-11-21** after manual browser download (`www12.statcan.gc.ca` sits behind a
Cloudflare check that rejects `curl`) and SHA-256 check. Matched against all 293 Canadian GADM
level-2 units in `data/county-presence/CAN.*.tsv.gz`: 265 match 1-to-1 by province and normalized
name, 24 map via explicit StatCan 2021 → GADM 4.1 renames/splits (Québec municipal mergers,
BC's Comox Valley + Strathcona → Comox-Strathcona, Nunavut Inuktitut names, NWT 6 regions →
Fort Smith / Inuvik, and Greater Sudbury), and the remaining 4 are Ontario Great Lakes water units.
