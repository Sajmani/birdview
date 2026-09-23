# Using BirdView

BirdView shows you how big birds actually are relative to one another. Everything on screen is
drawn to one shared scale, so a hummingbird beside a swan looks the way it would look in life.

The page is server-rendered and every control is a plain link, so the URL always describes exactly
what you are looking at. Bookmark it, share it, or hand-edit the query string.

---

## Reading a tile

Each bird is one tile: a header, a diagram, and three numbers.

The diagram is a **top-down view of a bird in flight**, drawn head-up:

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

---

## The controls

### Order by — length, wingspan or mass

Sorts every bird by that measurement, then groups them into sections of similar size.

Switching between the three is the fastest way to see how differently birds are built. A Northern
Harrier and a Cooper's Hawk sort close together by length and far apart by mass; sorting by
wingspan separates soaring birds from flapping ones almost cleanly.

### Scale — global or per section

This is the important one.

- **Global** (the default) draws everything at one scale across the whole filtered set. A kinglet
  really is a speck beside a swan. This is the honest view, and the reason the site exists.
- **Per section** rescales each section to itself, so every bird fills its tile. Shape is easier to
  compare within a group, but the size gap *between* groups disappears.

Use Global to be surprised; use Per section to study a family.

### Section step

How much larger a bird has to be to start a new section — a percentage. Smaller values give more,
tighter sections. The default is 15%.

A section always spans less than the step, so at 15% no bird in a section is more than 15% larger
than the smallest one in it.

### Units

Inches and ounces, or centimeters and grams.

### Family and order selector

Everything is selected when you arrive.

- Click a **family** to drop it.
- Click an **order** to drop that whole group.
- Click **All** to clear everything and start picking.

Families are grouped by taxonomic order in **checklist sequence** — the order field guides use, not
alphabetical — so related birds sit together and the list reads the way you already think.

Each family chip carries both its scientific and common name plus a count, and each family has its
own color, shaded by order: families in the same order get related hues, so you can pick out an
order at a glance while still telling its families apart.

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
- **No range, season or behavior.** Rarity codes are used once, at build time, to pick the species
  list, and then discarded.
- **No perched or standing view.** The diagrams are flight silhouettes only. A perched view would
  need leg and posture data that does not exist at this coverage.
- **No individual variation.** One figure per species, not a range.

For identification, use a field guide. For "wait, is a Merlin really that small?", use this.
