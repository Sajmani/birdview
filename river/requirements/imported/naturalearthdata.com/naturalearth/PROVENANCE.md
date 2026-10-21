# natural-earth — provenance

**Nothing is vendored here.** The archive is public domain and could be committed, but it is
14 MB of worldwide boundaries of which BirdView uses 69 polygons, so it is pinned by hash and
fetched when the boundary file is regenerated. See [imports.md](../../../imports.md#natural-earth)
for the adoption record.

## What is fetched

| Item | Value |
| --- | --- |
| **Landing page** | https://www.naturalearthdata.com/downloads/10m-cultural-vectors/10m-admin-1-states-provinces/ |
| **Download** | `https://naciscdn.org/naturalearth/10m/cultural/ne_10m_admin_1_states_provinces.zip` |
| **Version** | 5.1.1, per the archive's own `ne_10m_admin_1_states_provinces.VERSION.txt` |
| **Archive** | 14,909,524 bytes · SHA-256 `efc59726337323058f9446210adc96673179cd344e053666ee3d28cb58ba2b05` |
| **Read** | `ne_10m_admin_1_states_provinces.shp` and `.dbf`, 4,596 records worldwide |
| **Consumer** | `cmd/birdview-boundaries`, which refuses an archive with any other hash |
| **Output** | `data/region-boundaries.json`, committed |

**Verified 2026-10-18** by fetching the archive with `curl` and hashing it. The same day its
attribute table was read to confirm the five atolls BirdView covers are present, under ISO
3166-2 codes UM-67 (Johnston Atoll), UM-71 (Midway Islands), UM-76 (Navassa Island), UM-95
(Palmyra Atoll) and UM-79 (Wake Atoll). The US and Canadian units carry ISO 3166-2 codes whose
suffix is the postal abbreviation BirdView's region slugs use.

## Refreshing

Natural Earth versions the whole release. Fetch the new archive, update the hash here and in the
generator, regenerate, and read the diff of `data/region-boundaries.json`: a boundary that moved
by more than simplification would explain is worth a look before it ships.
