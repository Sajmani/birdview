# Using BirdView

BirdView shows you how big birds actually are relative to one another. Everything on screen is
drawn to one shared scale, so a hummingbird beside a swan looks the way it would look in life.

The page is server-rendered and every control is a plain link, so the URL always describes exactly
what you are looking at. Bookmark it, share it, or hand-edit the query string.

---

## Reading a tile

Each bird is one tile: a header, a stick figure drawn to scale, and three numbers.

The figure is a **top-down view of a bird in flight**, drawn head-up:

| Part | What it is |
| --- | --- |
| **Bill** | the exposed culmen — measured |
| **Head** | the cranium, as an oval — estimated from mass |
| **Neck** | estimated per family. Herons and pelicans fly with the neck folded back, so theirs is drawn as an **S** — the same length of line in less height |
| **Body** | the trunk — estimated from mass |
| **Tail** | dashed, so you can tell it from the body, with a tick at the junction |
| **Wing crossbar** | the wingspan, at the **same scale as the bird's height** — so a wing really is that long relative to the body |

There are no legs. In flight most birds tuck them away, and for the ones that trail them BirdView
cannot work out how far they would project. [docs/data.md](data.md) explains why.

**A `~` after a number means it is estimated rather than measured.** You will see a lot of them:
most lengths are modeled, because no open dataset publishes body length for North American birds.

**Click any bird** for the detail view — the same tile, larger, with every segment measured out
individually and the source or formula behind each number named.

The detail view also links to the species on **iNaturalist**, for photographs, range and
observations. BirdView has no pictures of its own and never will. A handful of species show no
link: those are ones iNaturalist could not be matched to with confidence, usually a recent
taxonomic split it has not followed yet, and an absent link is more honest than a guess that lands
on the wrong bird.

---

## The controls

### Order by — length, wingspan or mass

Sorts every bird by that measurement, then groups them into sections of similar size.

Switching between the three is the fastest way to see how differently birds are built. A Northern
Harrier and a Cooper's Hawk sort close together by length and far apart by mass; sorting by
wingspan separates soaring birds from flapping ones almost cleanly.

### Scale — per section or global

This is the important one.

- **Per section** (the default) rescales each section to itself, so every bird is big enough to
  read. Shape is easy to compare within a group, but the size gap *between* groups disappears.
- **Global** draws everything at one scale across the whole filtered set. A kinglet really is a
  speck beside a swan. This is the honest view across families, and the reason the site exists.

Use Global to be surprised; use Per section to study a family. Global is one click away and not
the default because at global scale a median bird is a couple of pixels tall, and none of the
figure's vocabulary — the dashed modeled segments, the chord ticks, the labels — is legible
there.

### Section step

How much larger a bird has to be to start a new section — a percentage. Smaller values give more,
tighter sections. The default is 15%.

A section always spans less than the step, so at 15% no bird in a section is more than 15% larger
than the smallest one in it.

### Units

Inches and ounces, or centimeters and grams. Large birds switch up to pounds and kilograms.

### Names

**N. America** uses the names a birder here would say — Dovekie, Brant, Rough-legged Hawk.
**World** uses AviList's global names for the same birds — Little Auk, Brant Goose, Rough-legged
Buzzard.

## The filters

Three boxes narrow which birds are drawn: **Regions**, **Diet and Habitat**, and **Families**.
Each arrives closed, with a one-line note saying what it holds; open the one you want. They
combine, and the count line always names what is currently on.

Everything is selected when you arrive, so every filter is a way of *dropping* birds:

- Click one thing — a family, a region, a diet — to drop it.
- Click a **group** heading to drop that whole group.
- Click **Only** beside a group to go the other way: keep that group and drop everything else in
  that box. From the arrival state this is the fast path, and it is one click.
- Click **All** to clear the box and start picking.

**Only** stays inside its own box. Only on a diet leaves your habitat and region choices alone.

### Regions

States, provinces, and stretches of open sea. A bird is listed in a region when it is reported
there often enough to be worth looking for, from eBird records — and from iNaturalist where eBird
withholds a sensitive species.

The same regions are offered **grouped two ways**, and you can pick from either. **Geographic**
is the usual map, drawn by hand, so someone looking for Alaska finds it under Alaska. **Common
birds** is measured — regions are clustered by how much their species lists actually overlap,
which puts Arizona with New Mexico and Texas, and Alaska with British Columbia.

Birds of the open sea belong to no state at all, so they are grouped by ocean. Seabirds that also
come within sight of shore appear both there and in the states they are seen from.

### Diet and Habitat

What a bird mostly eats and where it mostly lives, from **AVONET**, assigned **per species** and
not per family — a Double-crested Cormorant and a Neotropic Cormorant need not share a habitat,
and an Anhinga is a bird of swamps whatever its relatives do.

AVONET records *one* value per species. BirdView adds further habitats and diets where that hides
something obvious — a Great Blue Heron on a tidal flat, a gull at a landfill — and marks those as
its own rather than AVONET's. Every bird's detail page lists each category with who assigned it,
and [everything BirdView adds](/added) is listed in one place.

### Families

Families are grouped by taxonomic order in **checklist sequence** — the order field guides use, not
alphabetical — so related birds sit together and the list reads the way you already think.

Each family chip carries both its scientific and common name plus a count, and each family has its
own color, shaded by order: families in the same order get related hues, so you can pick out an
order at a glance while still telling its families apart.

---

## Comparing a few birds side by side

Picking birds puts them on one page at one shared scale, which is the comparison the index can
only approximate across sections.

Click **Compare birds** in Controls to start picking, then click the birds you want. The compare
page draws them together, largest setting the scale, and lets you sort the set by length,
wingspan or mass. Each panel links to that bird's own page, where every number names its source.

The birds you picked live in the URL, so a comparison is a link you can keep or send.

---

## A link to a bird BirdView stopped drawing

BirdView's list of species is derived from occurrence records and re-derived when those records
are refreshed, so it can change. If a bird leaves the list, its page does not vanish into
`404 page not found`: it tells you which bird you asked for and why it is no longer drawn —
seen but not repeatedly, repeatedly but not lately, or with too little in the open record to
judge. Try `/species/paspig`.

The address still answers *not found* rather than *gone*, because the list is re-derived and a
bird can come back.

---

## Sorting by how often a bird is recorded

**Order by → Frequency** arranges the page by how much of what people record is that bird,
most common first, using the figure already on every tile.

The tiers are powers of ten: one for the birds that are about one record in a thousand, the
next for one in ten thousand, and so on — nine of them across the whole list. **Section step**
switches between a tier per power of ten and tiers twice as wide.

Two things worth knowing. The tier edges are fixed, not fitted to whatever is on screen, so
*1 in 1,000 – 1 in 9,999* means the same thing with Alaska selected as it does with everything
selected — you can compare two views directly. And a tier is not a rarity label: BirdView
states the number and leaves the judgement to you, which is why no tier is called "common" or
"scarce".

Birds with no records in the regions you have selected are kept, in a last section of their
own, rather than dropped from the page.

**Scale switches to Global here, and that is deliberate.** Everywhere else each section is
rescaled to its own largest bird, which is safe because a section spans birds of similar size.
A frequency tier does not — it holds whatever happens to be about as often recorded, from
hummingbirds to swans — so rescaling tier by tier would draw the one 140 mm bird in the last
tier at the height of a 1,201 mm Trumpeter Swan six tiers above. One scale across the whole
page is the honest way to read this view. Per section is still one click away.

---

## Some things to try

- **Sort by mass, scale Global.** Mass is a cube of a linear dimension, so the range is brutal in a
  way length never shows.
- **Select just Ardeidae, then just Anatidae.** Herons are nearly all neck and leg; ducks are nearly
  all body. Same length, entirely different birds.
- **Sort by wingspan, then by length.** Watch the albatrosses and frigatebirds climb.
- **Look at the folded necks.** Herons and pelicans are the only families whose neck draws as an S,
  because they are the only ones that fly with the head drawn back onto the shoulders. It is the
  field mark that separates a heron from a crane or an ibis in the air.

---

## What BirdView will not tell you

It is a size and proportion tool, not a field guide.

- **No plumage, color or markings.** Tile colors encode taxonomy, nothing about the bird's
  appearance.
- **No range map, season or behavior.** You can filter by region, but that is a list of where a
  bird is reported often enough to be worth looking for — not a range map, not a migration, and
  not a date.
- **No perched or standing view.** The stick figures show a bird in flight only. A perched view would
  need leg and posture data that does not exist at this coverage.
- **No individual variation.** One figure per species, not a range.

For identification, use a field guide. For "wait, is a Merlin really that small?", use this.

---

## Where the numbers come from

Escaped cage, aviary and farm birds — parrots, waxbills, a farmed Emu — are not drawn on the
index, though their records pass BirdView's scope rule. [Every one is listed](/excluded) with
the reason, and each keeps its own page.

Every bird's detail page names the source or formula behind each number it shows. Beyond that,
[docs/data.md](data.md) describes each input and how far to trust it,
[how BirdView's data is built](pipeline.md) draws every source and step as one diagram, and
BirdView's own requirements are [in its repository](../river/) — the product and technical
specifications, the decision record, and the acceptance criteria that check them.
