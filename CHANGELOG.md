# Changelog

Major changes to BirdView, newest first: features added and removed, data sources adopted and
dropped, and changes in how the drawings are made. Small fixes and wording changes are left out;
the commit history has them.

BirdView has no numbered releases yet, so changes are grouped by the period they landed in.

## 2026-10-20 — Documentation for first-time visitors

### Changed

- The README introduces BirdView, how to run it and where to read more, instead of reporting
  project status. Data sources and their licenses moved to the building guide, with a current list.
- The guides were edited to describe BirdView as it is now, without project history.

## 2026-10-18 to 2026-10-19 — Here and Now, Bill and Tail, and tiers you can see

### Added

- **Here** and **Now** buttons beside All under Regions and Months. Here selects the region you are
  in, looked up in your browser so your location never reaches the server; Now selects the current
  month. Region outlines come from [Natural Earth](https://www.naturalearthdata.com/), in the public
  domain.
- **Bill** and **Tail** orderings, on the index and in the comparison view. Both are measured for
  every species; they default to the Global scale, since birds with the same bill can be any size.
- A welcome line under the title with things to try, including the birds around you now, ending
  with *reset to defaults*.
- A histogram under the species count: one bar per tier, linking to it, with empty space where a
  stretch of sizes holds no birds.

### Changed

- The species-count line names every setting, defaults included, and *Comparing* while picking
  birds.
- Section headings name the ordering ("Tail Tier 3"), state the range of that tier's own birds in
  one precision for the whole page, and give the species count beside the range.
- The guides flow as paragraphs instead of breaking at every source line.
- The comparison example is six birds of one length: Blue Jay, American Kestrel, California
  Thrasher, Mourning Dove, Belted Kingfisher and Chuck-will's-widow.
- The month buttons sit on their own line, under All and Now.

## 2026-10-14 to 2026-10-16 — Stick figures, accessibility, and a new process

### Added

- Every page is held to WCAG 2.2's contrast criteria, checked on every kind of page, including by
  rendering pages in a browser and measuring the text.
- A container image (Dockerfile).

### Changed

- The drawing is called a stick figure on the site.
- Escaped and kept birds — parrots, waxbills, ornamental waterfowl — are left off the index and
  listed on a page of their own.
- Going from the index to a comparison keeps the family filter.
- The project's requirements moved to the [River](river/README.md) process and are no longer served
  by the site; the guides link to them in the repository.

### Removed

- The server no longer serves its page templates as files under `/static/`.

## 2026-10-07 to 2026-10-13 — Months, open source, and the pipeline

### Added

- Filter by month as well as region.
- A diagram of how the dataset is built, field by field, served with the guides.
- BirdView is licensed under Apache 2.0, with every third-party file under `third_party/` and its
  own license; a code of conduct and a vulnerability reporting note.

### Changed

- How common a bird is now counts eBird records, using iNaturalist only where eBird withholds a
  sensitive species; eBird's withheld months are no longer read as months without the bird.
- Every response carries a validator, so a reload never shows a stale page.

## 2026-10-02 to 2026-10-05 — Open data, regions, and BirdView's own scope

### Added

- Filter by region: US states, Canadian provinces, five US outlying islands and five stretches of
  open sea, grouped two ways — on the map, and by which birds they share.
- Filter by habitat and diet, assigned per species from AVONET, with BirdView's own additions
  marked as such.
- Order by **Frequency**: how much of what people record is that bird, banded by powers of ten.
- Each tile says what share of recorded observations is that bird, following the region filter.
- Each species page lists every category the bird is sorted into and who decided it.
- A species BirdView does not draw answers its old address with a page saying why.
- An **Only** button beside every group; the guides served by the site itself.

### Changed

- Names, families, taxonomic sequence and species codes come from
  [AviList](https://doi.org/10.2173/avilist.v2025b), and a species' address is its eBird code.
- **Scope is BirdView's own**: a species is in when its records recur — repeated episodes across
  years — in at least one region, derived from open occurrence records through GBIF.
- Mass is written in one unit across whatever a page is comparing.
- Control panels start closed; responses are compressed.

### Removed

- **The ABA Checklist.** It supplied names and scope, was all rights reserved, and blocked public
  deployment; it is no longer read at all.
- The owner's email address from the public site.

## 2026-09-25 to 2026-10-01 — Comparing birds, and the mass cube

### Added

- A comparison view that draws any number of chosen birds on one scale, sortable by length,
  wingspan or mass. Birds are picked from the index, with a Compare birds toggle and a pick bar.
- Mass drawn as a cube of water that weighs the same, on the tiles and the species page.
- The drawing shows which lengths are measured (solid) and which are modeled (dashed), and the
  species page labels every segment with its measurement.
- A favicon drawn from the app's own figure.

### Changed

- Per-section scale is the default.
- The help splits into Presentation, which explains the drawing, and Controls, which explain
  themselves.

## 2026-09-20 to 2026-09-23 — The first BirdView

### Added

- The ingest, building the dataset from [AVONET](https://doi.org/10.6084/m9.figshare.16586228.v7)
  morphology and [Fox et al. 2026](https://doi.org/10.6084/m9.figshare.32089083) wingspans, and the
  server drawing each bird as a scaled figure — bill, head, neck, body, tail and a wing crossbar.
- A species page showing where every number came from: measured, near-match or modeled.
- Ordering by length, wingspan or mass, in sections of similar size, at a per-section or global
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
