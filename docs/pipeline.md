# How BirdView's data is built

Every number on a BirdView page comes from a published source, through a fixed sequence of
commands. This page shows that sequence: each source and table as a box listing its fields, and
each step as an arrow saying what it does — a filter, a join, a model.

**How to read the fields.** ★ marks a field that reaches a page. · marks one the pipeline uses
to decide something — a join key, a filter, a denominator — without showing it. ✗ marks one that
is read and then not used, which is listed because a reader auditing the pipeline would
otherwise wonder.

Rounded boxes are commands. Square boxes are tables. Dotted arrows are facts copied by hand into
code, rather than data read by a program.

```mermaid
flowchart TB
  subgraph SRC["Published sources"]
    AVL["AviList v2025b
    third_party/avilist, CC BY 4.0
    ─────────────
    Taxon_rank · keep species rows
    Sequence · family order
    Scientific_name ★ join key
    English_name_Clements_v2025 ★ North American name
    English_name_AviList ★ world name
    Species_code_Cornell_Lab ★ species code
    Family ★
    Family_English_name ★
    Order ★"]
    AVO["AVONET v7
    third_party/avonet, CC BY 4.0
    sheets AVONET1_BirdLife, AVONET2_eBird, AVONET3_BirdTree
    ─────────────
    Species1, Species2, Species3 · join key
    Order1-3 · wingspan model groups
    Beak.Length_Culmen ★ bill
    Tail.Length ★ tail
    Wing.Length ★ wing chord
    Mass ★ mass, and the head and body models
    Habitat ★ habitat
    Trophic.Niche ★ diet
    Trophic.Level ★ a second diet, where it says Omnivore
    Family1-3 ✗
    Secondary1 ✗"]
    FOX["Fox, Brunton and Shephard wingspans
    third_party/fox-wingspan, CC BY 4.0
    ─────────────
    Species1 · join key, BirdLife names
    needs_prediction · keep measured rows only
    wingspan_reported_mid_all ★ wingspan
    wingspan_reported_min_all ★ range
    wingspan_reported_max_all ★ range"]
    OCC["GBIF occurrence search
    live API, CC BY 4.0
    eBird EOD and iNaturalist research-grade
    ─────────────
    datasetKey · which source
    verbatimScientificName · the name as published
    gadmGid · state or province
    month · eventDate · when
    decimalLatitude · decimalLongitude · the seas
    gadmLevel0Gid · on land or off
    country = US or CA · filter"]
    SENS["eBird Sensitive Species list
    web page
    ─────────────
    species · region · date window"]
    TAXA["iNaturalist taxa API
    live
    ─────────────
    id ★ link to iNaturalist
    name · common name · match check"]
    BACK["GBIF backbone API
    live
    ─────────────
    name · rank · match tier"]
  end

  subgraph CODE["Hand-kept tables in code"]
    NAMES["internal/taxonomy
    ─────────────
    Occurrence: binomial,
    AviListRenames, CountedAs
    SharedNames · DomesticForms
    Renames · NearMatches for AVONET
    NotInAviList"]
    TSENS["taxonomy.EBirdSensitive
    ─────────────
    Species · Region · From · To
    WithheldByEBird, derived"]
    DECL["status.Declared
    ─────────────
    P-071 exceptions"]
    TRAIT["internal/trait
    ─────────────
    habitat and diet supplements"]
  end

  STATUSCMD(["birdview-status"])
  REGIONSCMD(["birdview-regions"])
  INATCMD(["birdview-inatmap"])
  NAMESCMD(["birdview-names"])
  GBIFCMD(["birdview-gbifmap"])

  subgraph DATA["Committed tables in data/"]
    STATUS["status.tsv
    ─────────────
    name · occurrence spelling
    status ★ established, rare, vagrant, historical, absent
    region ★ · episodes ★ · years ★
    records ★ · last seen ★
    commonest region, its records,
    all records there ✗"]
    REGIONS["region-presence.tsv
    ─────────────
    total rows:
    region, twelve monthly totals ·
    species rows:
    region, twelve monthly counts ★
    withheld flags ★ months eBird withholds"]
    INAT["inat-taxa.tsv
    ─────────────
    scientific name · join key
    taxon id ★
    common name · match check"]
    SN["source-names.tsv
    ─────────────
    id · scientific name
    ebird · inaturalist record counts
    read only by the T-028 check"]
    GT["gbif-taxa.tsv
    ─────────────
    GBIF name · AviList name · tier
    read by nothing yet"]
  end

  subgraph INGEST["birdview-ingest"]
    S1["1 scope
    keep established or rare;
    add Declared and both halves
    of a shared name"]
    S2["2 names
    join AviList on scientific name;
    drop NotInAviList"]
    S3["3 measurements
    join AVONET via
    Renames and NearMatches"]
    S4["4 segments
    head and body modeled from Mass;
    neck from body × family ratio;
    total length = sum of segments"]
    S5["5 wingspan
    Fox's measured row if any;
    else a log-log fit on Wing.Length
    within the order"]
    S6["6 habitat and diet
    AVONET's classes
    plus supplements"]
    S7["7 regions
    attach counts, months, withheld flags;
    peak share gives the top region"]
    S8["8 links
    attach the iNaturalist taxon id"]
    S9["not drawn
    rejected species keep
    their status and evidence"]
  end

  JSON["web/data/species.json
  build artifact, not committed
  ─────────────
  id ★ · names ★ · scientificName ★ · speciesCode ★
  family ★ · familyCommon ★ · order ★ · colors ·
  segments ★ · totalLength ★ · wingspan ★
  wingChord ★ · mass ★ · each with provenance ★
  habitats ★ · niches ★ · each with origin ★
  regions · regionCounts · regionMonths
  regionWithheld ★ · commonest ★
  inatTaxonId ★
  regionTotals · regionMonthlyTotals · denominators
  undrawn: status, region, episodes,
  years, records, lastSeen ★"]
  SERVER(["birdview server
  1 in N and the month filter,
  computed per request"])
  PAGES["pages
  index tiles · species · compare · not drawn"]

  OCC -->|"facet names by region, both sources;
  dates for each species in its top 3 regions"| STATUSCMD
  NAMES -->|"normalize every name"| STATUSCMD
  TSENS -->|"confirm withheld months"| STATUSCMD
  STATUSCMD -->|"recurrence on eBird's dates,
  iNaturalist's only where eBird withholds"| STATUS

  OCC -->|"facet names by region × 13 months;
  off-land counts for each sea"| REGIONSCMD
  NAMES -->|"normalize every name"| REGIONSCMD
  TSENS -->|"confirm withheld months"| REGIONSCMD
  REGIONSCMD -->|"count eBird, iNaturalist only in confirmed gaps;
  present at 1 per million"| REGIONS

  SENS -.->|"restated as facts"| TSENS
  TAXA --> INATCMD -->|"exact name and common name agree"| INAT
  OCC --> NAMESCMD -->|"records per source under each AviList name"| SN
  BACK --> GBIFCMD --> GT

  STATUS --> S1
  DECL --> S1
  S1 --> S2 --> S3 --> S4 --> S5 --> S6 --> S7 --> S8 --> JSON
  STATUS -->|"species left out of scope"| S9 --> JSON
  AVL --> S2
  NAMES --> S3
  AVO --> S3
  AVO --> S4
  FOX --> S5
  AVO --> S6
  TRAIT --> S6
  REGIONS --> S7
  INAT --> S8

  JSON --> SERVER --> PAGES
```

The diagram above is drawn on GitHub straight from this file. On BirdView's own documentation it
is a pre-rendered picture of the same source, so the page needs no network to draw it.

## The steps

**Scope comes from recurrence.** `birdview-status` asks GBIF which names each state and province
records, then fetches the dates behind each species in its three best regions and classifies it:
*established* when its records recur in repeated episodes across years. eBird's dates count
everywhere; iNaturalist's count only in a region and month where eBird is confirmed to withhold
the species. See [where the data comes from](data.md#scope) and
[CR-038](../spec/decisions.md#cr-038).

**Region counts come from the same source, finer.** `birdview-regions` asks GBIF for every name in
every region, once for the year and once for each month, and counts the off-land records inside
each sea's box separately. A species is *present* in a region at 1 record per million of what is
recorded there. eBird's records count everywhere; iNaturalist's fill in only where eBird's
[Sensitive Species list](data.md#when-the-sources-disagree-about-a-name) names the species and the
records confirm the gap ([CR-037](../spec/decisions.md#cr-037)).

**Names are normalized before anything is matched.** Both commands pass every published name
through one function, `taxonomy.Occurrence`: subspecies reduced to their species, hybrids left
unmatched, genus moves and boundary disagreements mapped to the AviList name. The tables it
uses are listed on each species' page under *Recorded as*.

**The ingest joins everything into one file.** It keeps the in-scope species, names them from
AviList, measures them from AVONET, models the segments no source measures, takes wingspan from
Fox where it was measured and estimates it where it was not, and attaches the region counts and
the iNaturalist link. Every metric carries its provenance — *measured*, *near-match* or
*modeled* — and the page shows which. See [what is measured and what is modeled](data.md).

**The server does the last arithmetic.** The *1 in N* on a tile and the month filter depend on the
regions and months a reader selects, so they are computed per request from the counts and totals
in `species.json`, not stored.

## Where to go next

- For what each figure means and how far to trust it: [where the data comes from](data.md).
- For running the pipeline yourself: [building BirdView](building.md).
- For every source's license, version and hash: [the source register](../spec/sources.md).
