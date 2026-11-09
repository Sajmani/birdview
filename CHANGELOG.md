# Changelog

Major changes to BirdView, newest first: features, data sources, and how the drawings are made.
Small fixes are in the commit history. With no numbered releases yet, changes are grouped by period.

## 2026-11-08 — Reporting frequency, and reporting data issues

### Added

- **Report a data issue** under each table on a bird's page opens a GitHub issue with the bird, the
  table's values, and the data's versions filled in. The home page's **Report it** now fills in the
  page's address, its settings, and the versions too, and both name the exact build of the site.
  So does **Report a problem** on the page for a bird BirdView no longer draws: every report link
  on the site now fills in what it is about.

### Changed

- A bird's neck, head, and body say whether their proportions are proposed or reviewed.
- The 26 woodpeckers, loons, and grebes whose modeled wingspan stated its error as "NaN%" now state
  a measured one.
- **Diving ducks and grebes** are drawn closer to how they fly: scaup, goldeneyes, scoters, and
  eiders with shorter necks, grebes with longer necks and smaller heads. A Lesser Scaup's neck was
  longer than a Horned Grebe's.
- **How common** is now **Reporting frequency**, since it measures how often a bird is reported,
  not how many are there. Addresses use `freq=` in place of `common=`, and old ones still work.
  A bird's page and the guide call the share of reports its reporting frequency too.

## 2026-11-07 — How common, and habitats in groups

### Added

- **How common** keeps only the birds that are at least 1 in 100, 1 in 1,000, or 1 in 10,000 of the
  observations in the regions and months chosen. Sorted by length or wingspan, the familiar birds of
  a place and season line up by size across families: in Brooklyn in November, Ring-billed Gull
  shares a wingspan tier with Red-tailed Hawk and Common Raven.

### Changed

- **Habitat** arranges AVONET's eleven habitats in five groups, a line each: Shore and sea, Fresh
  water, Woods, Open country, and Towns and farms. A group's name chooses all its habitats at once,
  and each can still be chosen alone. The welcome line's grassland hunters are now the open-country
  hunters.
- **Habitats for 60 water birds**, reviewed one by one: shorebirds, herons, rails, gulls, and terns
  now count as birds of the shore, and sea ducks, loons, and grebes as birds of the shore and sea as
  well, where AVONET files many of them under the habitat they breed in.
- Every setting that can be all always offers **All**, marked when chosen, so the choices no longer
  shift. How common no longer shows All twice.

## 2026-11-03 — A sky-blue cube, and regions by their data codes

### Changed

- A figure or note limited to some months names them in parentheses after the place: *1 in 31
  species observations in New York (Nov)*, *Withheld by eBird in Florida (Oct–Mar)*.

- A region in an address is now the code its records are filed under: `?region=USA.33_1` for New
  York rather than `?region=NY`. Old addresses still work and go to the new ones. This makes room
  for counties, which have no postal codes. The sea regions keep their names for now.
- Every number's tooltip, and its row on a bird's own page, now begins by saying what it measures in
  everyday words, such as "One folded wing, from the bend of the wing to its tip" for Chord. The
  guide lists them all. Wingspan follows the definition in Fox and colleagues' paper.
- A bird found by name links to its own page, so one bird's measurements no longer need a pin.
- The credits cite GBIF's two datasets in full, by DOI, and give iNaturalist's research-grade
  observations their actual license, CC BY-NC 4.0; every page had said CC BY 4.0 for both.

- The mass cube and its name are sky blue again, a shade light enough to read as text. For
  color-blind readers it can look like the wing or the chord, but it stands apart from the figure
  with its name beside it.

## 2026-11-01 — Larger type, and one tile on both pages

### Changed

- No text is smaller than 13px, and running text is 15px, 16px on a phone. The species page's
  diagram labels had rendered at 3.5px on a phone.
- The species page shows the home page's tile, larger, naming every part and its measurement, in
  place of a diagram labeled with leader lines. Bill, Tail, Inner, Chord, and Mass are named in their
  parts' colors.
- Measurements are set in a proportional face with even-width figures instead of monospace.
- The tail and the wing chord have colors of their own: amber bill, white tail, emerald wing with
  an orchid chord, blue mass cube, and gray modeled parts, every two distinct for color-blind
  readers too.

### Added

- Pinned birds lead the list as a **Pinned** group, side by side, and stay in their tiers below.
- Tiles name the crossbar's emerald middle **Inner**: the wingspan less the two wing chords, marked
  `~` because it is worked out rather than measured. **Wingspan**, which names the whole bar, is no
  longer emerald. The icon shows the orchid chords too.
- Every number on a tile names its source when you hover over it, measured ones included. Only
  estimates had a tooltip before, and the mass had none.
- **Tiny birds** joins the things to try: six birds around 10 cm long, from Winter Wren to Bushtit.
- The line of things to try begins **Try**, without "New here?".
- The line under the histogram says it counts all the birds, then how many are shown: *Length
  distribution of all 1041 birds, showing the first 100*. It said *of the first 100 of 1041
  birds*, though the histogram has always counted them all.
- **Grassland hunters** joins them too: the birds of grassland that hunt animals, from Burrowing
  Owl to Golden Eagle.

## 2026-11-01 — eBird links

### Changed

- The link on every tile and species page goes to the species on **eBird**, by the eBird code
  beside its name, instead of iNaturalist. The tile's button reads **eBird**. Every species has a
  code now: Ruddy Duck, White-necked Petrel, and American Barn Owl, which AviList gives none, link
  to eBird too.
- Where a bird has two eBird codes, the simpler (fewer digits, then shorter) is shown, linked, and
  its page address; the other still finds it, and its old address redirects.

### Removed

- The iNaturalist taxon table and the tool that built it from iNaturalist's API: the eBird link
  needs no lookup. iNaturalist's observations still count toward where each bird is found, through
  GBIF.

## 2026-10-30 to 2026-10-31 — Tiles put the numbers first

### Changed

- Each tile leads with the bird's name and codes, then its length, bill, and tail, its wingspan
  and wing chord, and its *1 in N* share, above the drawing. The mass is named beside its cube.
  Under the drawing, the family, by its scientific name, and the bird's scientific name. The pin
  moves to the tile's top-right corner.

### Added

- An **iNat** link in every tile's lower-right corner, to the species on iNaturalist.

## 2026-10-26 to 2026-10-28 — A new home page: find, pin, and compare

### Added

- A **find box** above the birds: type part of a name, code, or scientific name to count and list
  matching birds.
- **Pins**: pin any bird, from the find list, its tile, or its species page, to keep it on the page
  whatever the settings. **Pinned**, always in view under the heading, shows only them,
  side by side at one scale ([#5](https://github.com/google/birdview/issues/5)).
- **New here? Try** compares small, medium, large, and huge birds with pins.
- Tiers with headings and a histogram with a bar per tier, under every ordering.
- Four-letter **banding codes** from the USGS Bird Banding Laboratory's public-domain list, beside
  the eBird code on tiles, species pages, and the find list. Birds the Laboratory gives no code
  show none; BirdView does not make codes up.

### Changed

- The home page draws one set of birds at one scale, the first 100 by default. A heading says
  what is shown ("Showing birds in New York in October"); under it, **First**, **All**, and
  **Pinned** are always in view, then your pins, the find box, and the
  settings, each opening its choices beneath it.
- The home page shows **every month** by default. **Now** chooses the current month by your
  device's clock, as does *the birds around you now*.
- Units' choices read **Imperial** and **Metric**.
- A species page's Back returns by browser history, so the home page comes back as you left it.

### Removed

- The classic index, the comparison view (`/compare`), `/now`, and pick mode. Pins replace picks.

## 2026-10-25 — Phones and search

### Added

- A search box that narrows the index as you type.

### Changed

- Every page fits a phone held upright.
- Shorter page text: design reasoning moved out of the pages into the project's records.

## 2026-10-22 to 2026-10-23 — BirdView is public at birdview.site

### Added

- [birdview.site](https://birdview.site), served from Google Cloud Run; `www` and the Cloud Run
  address redirect to it.
- `robots.txt`, asking crawlers to skip filtered views.
- Every deploy is tested first and checked from outside after.

### Changed

- The addresses readers bookmark and share are now promised to keep working: the home page,
  species pages, the guides, and the parameters the pages' own links write.

## 2026-10-20 — Documentation for first-time visitors

### Changed

- The README introduces BirdView, how to run it, and where to read more, instead of project
  status. Data sources and their licenses moved to the building guide, with a current list.
- The guides describe BirdView as it is now, without project history.

## 2026-10-18 to 2026-10-19 — Here and Now, Bill and Tail, and tiers you can see

### Added

- **Here** and **Now** buttons beside All under Regions and Months. Here selects your region,
  looked up in your browser so your location never reaches the server; Now selects the current
  month. Region outlines come from [Natural Earth](https://www.naturalearthdata.com/), public
  domain.
- **Bill** and **Tail** orderings, on the index and in the comparison view. Both are measured for
  every species; they default to Global scale, since birds with the same bill can be any size.
- A welcome line under the title with things to try, including the birds around you now, ending
  with *reset to defaults*.
- A histogram under the species count: one bar per tier, linking to it, with gaps where no birds
  fall.

### Changed

- The species-count line names every setting, defaults included, and *Comparing* while picking.
- Section headings name the ordering ("Tail Tier 3"), give the range of that tier's own birds at
  one precision for the whole page, and the species count.
- The guides flow as paragraphs instead of breaking at every source line.
- The comparison example is six birds of one length: Blue Jay, American Kestrel, California
  Thrasher, Mourning Dove, Belted Kingfisher, and Chuck-will's-widow.
- The month buttons sit on their own line, under All and Now.

## 2026-10-14 to 2026-10-16 — Stick figures, accessibility, and a new process

### Added

- Every kind of page is held to WCAG 2.2's contrast criteria, checked partly by rendering pages in
  a browser and measuring the text.
- A container image (Dockerfile).

### Changed

- The drawing is called a stick figure on the site.
- Escaped and kept birds — parrots, waxbills, ornamental waterfowl — are left off the index and
  listed on a page of their own.
- Going from the index to a comparison keeps the family filter.
- The requirements moved to the [River](river/README.md) process and are no longer served by the
  site; the guides link to them in the repository.

### Removed

- The server no longer serves its page templates as files under `/static/`.

## 2026-10-07 to 2026-10-13 — Months, open source, and the pipeline

### Added

- Filter by month as well as region.
- A diagram of how the dataset is built, field by field, served with the guides.
- BirdView is licensed under Apache 2.0, every third-party file under `third_party/` with its own
  license; a code of conduct and a vulnerability reporting note.

### Changed

- How common a bird is counts eBird records, using iNaturalist only where eBird withholds a
  sensitive species; eBird's withheld months are no longer read as months without the bird.
- Every response carries a validator, so a reload never shows a stale page.

## 2026-10-02 to 2026-10-05 — Open data, regions, and BirdView's own scope

### Added

- Filter by region: US states, Canadian provinces, five US outlying islands, and five stretches of
  open sea, grouped two ways — on the map, and by which birds they share.
- Filter by habitat and diet, assigned per species from AVONET, with BirdView's own additions
  marked as such.
- Order by **Frequency**: how much of what people record is that bird, banded by powers of ten.
- Each tile says what share of recorded observations is that bird, following the region filter.
- Each species page lists every category the bird is sorted into and who decided it.
- A species BirdView does not draw answers its old address with a page saying why.
- An **Only** button beside every group; the guides are served by the site.

### Changed

- Names, families, taxonomic sequence, and species codes come from
  [AviList](https://doi.org/10.2173/avilist.v2025b), and a species' address is its eBird code.
- **Scope is BirdView's own**: a species is in when its records recur — repeated episodes across
  years — in at least one region, derived from open occurrence records through GBIF.
- Mass is written in one unit across whatever a page is comparing.
- Control panels start closed; responses are compressed.

### Removed

- **The ABA Checklist.** It supplied names and scope, was all rights reserved, and blocked public
  deployment; it is no longer read.
- The owner's email address from the public site.

## 2026-09-25 to 2026-10-01 — Comparing birds, and the mass cube

### Added

- A comparison view drawing any number of chosen birds on one scale, sortable by length, wingspan,
  or mass, picked from the index with a Compare birds toggle and a pick bar.
- Mass drawn as a cube of water that weighs the same, on the tiles and the species page.
- The drawing marks measured lengths solid and modeled ones dashed; the species page labels every
  segment with its measurement.
- A favicon drawn from the app's own figure.

### Changed

- Per-section scale is the default.
- The help splits into Presentation, explaining the drawing, and self-explanatory Controls.

## 2026-09-20 to 2026-09-23 — The first BirdView

### Added

- The ingest, building the dataset from [AVONET](https://doi.org/10.6084/m9.figshare.16586228.v7)
  morphology and [Fox et al. 2026](https://doi.org/10.6084/m9.figshare.32089083) wingspans, and the
  server drawing each bird as a scaled figure — bill, head, neck, body, tail, and a wing crossbar.
- A species page showing where every number came from: measured, near-match, or modeled.
- Ordering by length, wingspan, or mass, in sections of similar size, at a per-section or global
  scale; a family filter with generated colors per order and family.
- A link from each species to iNaturalist, for photographs.
- Text held to a 7:1 contrast floor everywhere.

### Changed

- Wingspans missing from the data are estimated by a regression fitted to measured ones, replacing
  an invented model.
- The figure became a bird in flight: legs drawn from the measured tarsus were removed, and herons
  and pelicans draw the neck folded as an S.

### Removed

- Ricklefs 2017 as a source of body lengths: its publisher's terms forbid the use.
