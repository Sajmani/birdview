# us-census-boundaries — provenance

**Nothing is vendored here.** The archive is public domain (17 U.S.C. § 105) and could be
committed, but it is 2.6 MB of shapefiles from which BirdView derives compact per-state county
boundary files, so it is pinned by hash and fetched when county boundaries are regenerated. See
[imports.md](../../../imports/imports.md#us-census-boundaries) for the adoption record.

## What is fetched

| Item | Value |
| --- | --- |
| **Documentation** | https://www.census.gov/programs-surveys/geography/technical-documentation/naming-convention/cartographic-boundary-file.html |
| **Download** | `https://www2.census.gov/geo/tiger/GENZ2021/shp/cb_2021_us_county_5m.zip` |
| **Version** | 2021 Cartographic Boundary Files, 1:5,000,000 scale (chosen because Connecticut switched from its 8 counties to 9 planning regions in 2022, while GADM 4.1 uses the 8 counties) |
| **Archive** | 2,707,232 bytes · SHA-256 `27444a787c0fb9c38f665293e99f7d527b7a5b4cef7a87aa3360d467be54ce38` |
| **Read** | `cb_2021_us_county_5m.shp` and `.dbf` (WGS84 / NAD83 geographic coordinates) |
| **Consumer** | `cmd/birdview-boundaries -counties`, which refuses an archive with any other hash |
| **Output** | `data/county-boundaries/USA.*.json.gz`, committed |

**Verified 2026-11-21** by fetching the archive from `www2.census.gov` and hashing it. Matched
against all 3,148 US GADM level-2 units in `data/county-presence/USA.*.tsv.gz`: 3,129 match 1-to-1
by state and normalized name, 7 map via explicit historical renames/splits (South Dakota's Shannon
→ Oglala Lakota, Alaska's Wade Hampton → Kusilvak and split census areas), and the remaining 12
are 10 Great Lakes water units and 2 merged Virginia independent cities (Bedford City, Clifton Forge
City).
