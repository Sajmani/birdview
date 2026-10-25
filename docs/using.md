# Using BirdView

BirdView shows how big birds are relative to one another. Each bird is drawn as a stick figure to
scale, so a hummingbird beside a swan looks as it would in life.

The page is server-rendered and every control is a plain link, so the URL always describes what
you are looking at: bookmark it, share it, or edit the query string.

---

## The page

Under the title, **New here? Try** offers a few views to start with, then **more ideas**, which
leads to [Some things to try](#some-things-to-try), and **reset to defaults**, which returns the
page to how it is when you arrive.

Below the control panels, the **count line** gives how many species are shown and names every
setting, defaults included: Order by, Scale, Units, Names, Section step, Comparing (only while
picking birds), Month, Region, Habitat, Diet and Family.

Under it, a small **histogram** has one bar per section, as tall as its species count, each linking
to its section. A gap in the row marks a stretch of the scale with no birds.

Each section heading names the measurement and section, such as **Tail Tier 3**, then the range of
that section's own birds and its species count, such as *6 species*.

---

## Reading a tile

Each tile has a header with the bird's names and family, a stick figure drawn to scale, a cube
showing its mass, its length, wingspan and mass as numbers, and how many of the species
observations in the selected regions are of this bird, as *1 in N*.

The figure is a **top-down view of a bird in flight**, drawn head-up. Solid lines are measured;
dashed lines are estimated.

| Part | What it is |
| --- | --- |
| **Bill** | the exposed culmen. Measured, so solid |
| **Head** | the cranium, as an oval. Estimated from mass, drawn hollow |
| **Neck** | estimated per family, dashed. Herons and pelicans fly with the neck folded back, so theirs is an **S**: the same length of line in less height |
| **Body** | the trunk. Estimated from mass, dashed |
| **Tail** | measured, so solid, with a tick where it meets the body |
| **Wing crossbar** | the wingspan, at the **same scale as the bird's length**. The solid outer ends are the measured wing chord, with ticks where it reaches; the middle is dashed when the wingspan is estimated |
| **Cube** | a cube of water that weighs the same as the bird, at the same scale. Every cube stands on the floor of its drawing, so two can be compared directly |

There are no legs: most birds tuck them away in flight, and for those that trail them BirdView
cannot work out how far they would project. [docs/data.md](data.md) explains why.

**A `~` after a number means it is estimated.** Most lengths are, because no open dataset publishes
body length for North American birds.

**Click any bird** for the detail view: the same figure, larger, with each part measured out and
the source or formula behind each number.

The detail view links to the species on **iNaturalist** for photographs, range and observations;
BirdView has no pictures of its own. A few species have no link because they could not be matched
to iNaturalist with confidence, usually after a recent taxonomic split; BirdView leaves the link
out rather than risk sending you to the wrong bird.

---

## The controls

### Order by

**Length**, **Wingspan**, **Mass**, **Bill** and **Tail** sort birds by that measurement, smallest
first, in sections of similar size. **Frequency** sorts by how often a bird is recorded, most common
first; see [below](#sorting-by-how-often-a-bird-is-recorded).

**Bill** and **Tail** are measured for every bird, unlike total length, which is partly modeled.
They default to the Global scale, because birds with the same bill can be any size: one bill
section runs from a hummingbird to a pheasant.

Switching orderings shows how differently birds are built. A Northern Harrier and a Cooper's Hawk
sort close together by length and mass, and far apart by wingspan; wingspan separates soaring birds
from flapping ones almost cleanly.

### Scale — per section or global

This is the most important control.

- **Per section** rescales each section to itself, so every bird is big enough to read. Shape is
  easy to compare within a group, but the size gap *between* groups disappears. This is the
  default for Length, Wingspan and Mass.
- **Global** draws everything at one scale, so a kinglet is a speck beside a swan. This is the
  honest view across families, and the reason the site exists. It is the default for Bill, Tail
  and Frequency.

Use Global to be surprised and Per section to study a family. Global is not the default for the
other orderings because at global scale a median bird is a couple of pixels tall, too small to
show the dashed segments, chord ticks or labels.

### Section step

How much larger a bird must be to start a new section: 5, 10, 15, 20 or 30 percent. The default
is 15%, meaning no bird in a section is more than 15% larger than its smallest; smaller values give
more, tighter sections. Under Frequency the choices are 10× and 100×.

### Units

Inches and ounces, or centimeters and grams. Masses switch to pounds or kilograms where a section
holds large birds.

### Names

**N. America** uses the names a birder here would say: Dovekie, Brant, Rough-legged Hawk.
**World** uses AviList's global names: Little Auk, Brant Goose, Rough-legged Buzzard. The
scientific name, species code and page address do not change.

## The filters

Three boxes narrow which birds are drawn: **Regions and months**, **Diet and Habitat**, and
**Families**. Each starts closed, with a one-line note on what it holds. The filters combine, and
the count line names each setting.

Everything is selected when you arrive, so every filter *drops* birds:

- Click one item (a month, region, diet or family) to drop it.
- Click a **group** heading, such as a zone of regions or an order of families, to drop the group.
- Click **Only** beside a group or category to keep just that within its box; from the arrival
  state this takes one click.
- Click **All** to clear the box and start picking, or to select everything again.

**Only** stays inside its own box: Only on a diet leaves your habitat and region choices alone.

### Months

The twelve months sit on their own line, under **All** and **Now**, which selects only the current
month. Months narrow the birds shown, not just the numbers: pick June and you get the birds present
in June, with frequencies counted against June's records. A bird with too few records to tell when
it is present is shown in every month, and says so.

### Regions

States, provinces, and stretches of open sea. A bird is listed in a region when it is reported
there often enough to be worth looking for, from eBird records, and from iNaturalist where eBird
withholds a sensitive species.

**Here**, beside **All**, selects only the region you are in. Your browser asks for your location,
looks up the region itself, and loads the page you would get by choosing that region by hand, so
your location never reaches the server. If you refuse permission, your location cannot be found,
or you are outside every region BirdView covers, Here says so and leaves the regions as they were.
Here needs JavaScript.

The regions are offered **grouped two ways**, and you can pick from either. **Geographic** is the
usual map, drawn by hand, so Alaska is under Alaska. **Common birds** is measured: regions are
clustered by how much their species lists overlap, which puts Arizona with New Mexico and Texas,
and Alaska with British Columbia.

Open-sea birds belong to no state, so they are grouped by ocean. Seabirds also seen from shore
appear both there and in the states they are seen from.

### Diet and Habitat

What a bird mostly eats and where it mostly lives, from **AVONET**, assigned **per species**, not
per family: a Double-crested Cormorant and a Neotropic Cormorant need not share a habitat, and an
Anhinga is a bird of swamps whatever its relatives do.

AVONET records *one* value per species. BirdView adds habitats and diets where that hides something
obvious, such as a Great Blue Heron on a tidal flat or a gull at a landfill, and marks them as its
own. Each detail page lists every category with who assigned it, and
[everything BirdView adds](/added) is listed in one place.

### Families

Families are grouped by taxonomic order in **checklist sequence**, the order field guides use rather
than alphabetical, so related birds sit together.

Each family chip shows its scientific and common name. Each family has its own color, with related
hues for families in the same order, so you can pick out an order at a glance and still tell its
families apart.

---

## Comparing a few birds side by side

The comparison page draws chosen birds together at one scale, which the index can do only for
birds in the same section.

Turn on **Compare birds** in Controls and every tile gets a **+ Pick** button; a bar along the
bottom collects your picks. The filters still work, so you can narrow to one family and pick from
it. Press **Compare** in the bar to see them together, with the largest setting the scale. The
comparison sorts by Length, Wingspan, Mass, Bill or Tail, and each panel links to **Species
details**, the bird's own page, where every number names its source.

The picks are in the URL, so a comparison is a link you can keep or send.

---

## A link to a bird BirdView no longer draws

The species list is derived from occurrence records and re-derived when they are refreshed, so it
can change. A bird that leaves the list does not get a bare `404 page not found`: its page names
the bird and why it is not drawn — seen but not repeatedly, repeatedly but not lately, or too
little in the open record to judge. Try `/species/paspig`.

The page returns *not found* rather than *gone*, because a bird can come back.

---

## Sorting by how often a bird is recorded

**Order by → Frequency** arranges the page by how much of what people record is that bird, most
common first, using the *1 in N* figure on every tile.

The tiers are powers of ten: 1 in 10 to 1 in 99, then 1 in 100 to 1 in 999, and so on. **Section
step** switches between a tier per power of ten (10×) and tiers twice as wide (100×).

The tier edges are fixed, not fitted to what is on screen, so *1 in 1,000 – 1 in 9,999* means the
same with Alaska selected as with everything selected, and two views compare directly. A tier is
not a rarity label: BirdView states the number and leaves the judgement to you, so no tier is
called "common" or "scarce".

Birds with no records in the selected regions are kept in a last section of their own.

**The scale defaults to Global here.** A frequency tier holds birds of every size, so rescaling
each tier to its largest bird would draw a small bird alone in its tier as tall as a swan in
another. Per section is one click away.

---

## Some things to try

- **[The birds around you now](/?by=frequency&scale=global#here-and-now).** The commonest birds
  where you are this month, on one scale. Your browser asks for your location, works out your
  region itself, and selects it with the current month; your location never leaves the browser.
  Without sharing it, try [New York in October](/?by=frequency&scale=global&region=NY&month=Oct),
  or press **Here** and **Now** under **Regions and months**.
- **Compare some [small birds](/compare?species=leasan&species=houspa&species=dowwoo&species=houfin&species=barswa&species=easpho),
  [medium birds](/compare?species=blujay&species=amekes&species=calthr&species=moudov&species=belkin1&species=chwwid),
  [large birds](/compare?species=turvul&species=gbbgul&species=greegr&species=cangoo&species=grecor&species=rosspo1),
  [huge birds](/compare?species=calcon&species=amwpel&species=mutswa&species=magfri&species=baleag&species=bkfalb).**
  All on one scale. The first three sets are each birds of about one length, so what differs is
  their build; the huge birds all have wings over two metres across.
  - *Small:* Least Sandpiper, House Sparrow, Downy Woodpecker, House Finch, Barn Swallow and
    Eastern Phoebe, all 141 to 158 mm long. The swallow is the lightest, 18 g, yet has the widest
    wings, 325 mm against 234 mm for the finch.
  - *Medium:* Blue Jay, American Kestrel, California Thrasher, Mourning Dove, Belted Kingfisher
    and Chuck-will's-widow, all 272 to 292 mm long. Their wingspans run from 307 mm for the
    thrasher to 655 mm for the nightjar, more than double, and their bills from 15 mm for the
    nightjar to 59 mm for the kingfisher, nearly four times.
  - *Large:* Turkey Vulture, Great Black-backed Gull, Great Egret, Canada Goose, Great Cormorant
    and Roseate Spoonbill, all 614 to 728 mm long. The vulture, the shortest, has the widest
    wings, 1805 mm, against 1280 mm for the spoonbill, and the goose weighs more than three times
    as much as the egret, 2.8 kg against 0.87 kg. The spoonbill's bill, 160 mm, is the longest.
  - *Huge:* California Condor, American White Pelican, Mute Swan, Magnificent Frigatebird, Bald
    Eagle and Black-footed Albatross, with wingspans from 2059 mm for the eagle to 2800 mm for
    the condor. The albatross is the shortest, 709 mm, yet spans 2265 mm; the frigatebird, at
    1.5 kg, weighs about a seventh of the swan's 10.7 kg; and the pelican's bill is 320 mm.

  To compare any birds this way, turn on **Compare birds** in Controls and pick them from the
  index.
- **See all [herons and ducks](/?family=Ardeidae&family=Anatidae) together.** Similar lengths, different
  birds: herons are nearly all neck and leg, ducks nearly all body. Herons and pelicans are the
  only families whose neck is drawn as an S, because they fly with it folded back onto the
  shoulders, which separates a heron from a crane or an ibis in the air.
- **"Big" depends on the measure.** Sort by [length](/?by=length), then
  [wingspan](/?by=wingspan), then [mass, scale Global](/?by=mass&scale=global). Albatrosses and
  frigatebirds move up under wingspan, and mass, which grows with the cube of length, spreads the
  range far wider than length does.

---

## What BirdView will not tell you

It is a size and proportion tool, not a field guide.

- **No plumage, color or markings.** Tile colors encode taxonomy, not appearance.
- **No range map, migration or behavior.** The region and month filters list where and when a bird
  is reported often enough to be worth looking for; that is not a range map.
- **No perched or standing view.** A perched view would need leg and posture data that does not
  exist at this coverage.
- **No individual variation.** One figure per species, not a range.

For identification, use a field guide. For "wait, is a Merlin really that small?", use this.

---

## Where the numbers come from

Every detail page names the source or formula behind each number. [docs/data.md](data.md)
describes each input and how far to trust it, and [how BirdView's data is built](pipeline.md)
shows every source and step in one picture.

Escaped cage, aviary and farm birds, such as parrots, waxbills and a farmed Emu, pass BirdView's
scope rule but are not drawn on the index. [Every one is listed](/excluded) with the reason, and
each keeps its own page.

BirdView's requirements, its product and technical specifications and the checks that hold the
site to them, are [in its repository](../river/).
