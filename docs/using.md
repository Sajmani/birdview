# Using BirdView

BirdView draws birds as stick figures to scale, so a hummingbird beside a swan looks as it would in
life. Every setting is in the page address, so you can bookmark or share any view.

---

## The page

Under the title, **New here? Try** offers a few views to start with, **more ideas** leads to
[Some things to try](#some-things-to-try), and **reset to defaults** returns the page to how it was
when you arrived.

Below the panels, the **count line** says how many species are shown and names every setting. Under
it, a **histogram** has one bar per section, each linking to its section. Each section heading names
the measurement and section, such as **Tail Tier 3**, its range and its species count.

---

## Reading a tile

Each tile has the bird's names and family, a stick figure to scale, a cube for its mass, its length,
wingspan and mass, and how many of the species observations in the selected regions are of this
bird, as *1 in N*.

The figure is a **top-down view of a bird in flight**, head up. Solid lines are measured; dashed
lines are estimated.

| Part | What it is |
| --- | --- |
| **Bill** | measured, so solid |
| **Head** | estimated from mass, drawn hollow |
| **Neck** | estimated per family, dashed. Herons and pelicans fly with the neck folded, so theirs is an **S** |
| **Body** | estimated from mass, dashed |
| **Tail** | measured, so solid |
| **Wing crossbar** | the wingspan, at the same scale as the bird. The solid ends are the measured wing chord, ending at the ticks; the middle is dashed when the wingspan is estimated |
| **Cube** | a cube of water that weighs the same as the bird, at the same scale |

Legs are not drawn. **A `~` after a number means it is estimated**, as most lengths are.

**Click any bird** for its own page: the same figure, larger, with every part measured and the source
of each number, and a link to the species on **iNaturalist** for photographs.

---

## The controls

### Order by

**Length**, **Wingspan**, **Mass**, **Bill** and **Tail** sort birds by that measurement, smallest
first, in sections of similar size. **Frequency** sorts by how often a bird is recorded, most common
first; see [below](#sorting-by-how-often-a-bird-is-recorded).

Switching orderings shows how differently birds are built: a Northern Harrier and a Cooper's Hawk
sort close together by length and mass, and far apart by wingspan.

### Scale — per section or global

- **Per section** scales each section to its largest bird, so every bird is big enough to read. Use
  it to compare shape within a group. It is the default for Length, Wingspan and Mass.
- **Global** draws everything at one scale, so a kinglet is a speck beside a swan. Use it to compare
  size across families. It is the default for Bill, Tail and Frequency.

### Section step

How much larger a bird must be to start a new section: 5, 10, 15, 20 or 30 percent. Under Frequency
the choices are 10× and 100×.

### Units

Inches and ounces, or centimeters and grams.

### Names

**N. America** uses North American names: Dovekie, Brant, Rough-legged Hawk. **World** uses
AviList's: Little Auk, Brant Goose, Rough-legged Buzzard.

## The filters

**Regions and months**, **Diet and Habitat** and **Families** narrow which birds are drawn. They
combine, and everything is selected when you arrive:

- Click a month, region, diet or family to drop it.
- Click a **group** heading, such as a zone or an order, to drop the group.
- Click **Only** beside a group or category to keep just that, within its own panel.
- Click **All** to clear the panel, or to select everything again.

### Months

**Now** selects the current month. June shows the birds present in June. A bird with too few records
to tell is shown in every month, and says so.

### Regions

States, provinces and stretches of open sea. A bird is listed where it is reported often enough to
be worth looking for.

**Here** selects the region you are in. Your browser works out the region itself, so your location
never reaches the server. Here needs JavaScript.

The regions are **grouped two ways**: **Geographic**, by the map, and **Common birds**, by how much
their species lists overlap.

### Diet and Habitat

What a bird mostly eats and where it mostly lives, per species, from **AVONET**. BirdView adds a few
more and marks them as its own; [everything BirdView adds](/added) is listed in one place.

### Families

Grouped by taxonomic order, in field-guide sequence. Families in the same order share related
colors.

---

## Comparing a few birds side by side

Turn on **Compare birds** in Controls and every tile gets a **+ Pick** button; a bar along the bottom
collects your picks. Press **Compare** to draw them together at one scale. The comparison is a link
you can keep or send.

---

## A link to a bird BirdView no longer draws

Its page names the bird and says why it is not drawn. Try `/species/paspig`.

---

## Sorting by how often a bird is recorded

**Order by → Frequency** sorts by how much of what people record is that bird, most common first,
using the *1 in N* figure on every tile. The tiers are powers of ten: 1 in 10 to 1 in 99, then 1 in
100 to 1 in 999, and so on. Birds with no records in the selected regions come last.

---

## Some things to try

- **[The birds around you now](/?by=frequency&scale=global#here-and-now).** The commonest birds
  where you are this month, on one scale. Without sharing your location, try
  [New York in October](/?by=frequency&scale=global&region=NY&month=Oct).
- **Compare some [small birds](/compare?species=leasan&species=houspa&species=dowwoo&species=houfin&species=barswa&species=easpho),
  [medium birds](/compare?species=blujay&species=amekes&species=calthr&species=moudov&species=belkin1&species=chwwid),
  [large birds](/compare?species=turvul&species=gbbgul&species=greegr&species=cangoo&species=grecor&species=rosspo1),
  [huge birds](/compare?species=calcon&species=amwpel&species=mutswa&species=magfri&species=baleag&species=bkfalb).**
  - *Small:* Least Sandpiper, House Sparrow, Downy Woodpecker, House Finch, Barn Swallow and
    Eastern Phoebe, all 141 to 158 mm long. The swallow is the lightest, 18 g, yet has the widest
    wings, 325 mm against 234 mm for the finch.
  - *Medium:* Blue Jay, American Kestrel, California Thrasher, Mourning Dove, Belted Kingfisher
    and Chuck-will's-widow, all 272 to 292 mm long. Their wingspans run from 307 mm for the
    thrasher to 655 mm for the nightjar, and their bills from 15 mm for the nightjar to 59 mm for
    the kingfisher.
  - *Large:* Turkey Vulture, Great Black-backed Gull, Great Egret, Canada Goose, Great Cormorant
    and Roseate Spoonbill, all 614 to 728 mm long. The vulture, the shortest, has the widest
    wings, 1805 mm, against 1280 mm for the spoonbill, and the goose weighs more than three times
    as much as the egret.
  - *Huge:* California Condor, American White Pelican, Mute Swan, Magnificent Frigatebird, Bald
    Eagle and Black-footed Albatross, all with wings over two meters. The albatross is the
    shortest, 709 mm, yet spans 2265 mm; the frigatebird weighs about a seventh of the swan.
- **See all [herons and ducks](/?family=Ardeidae&family=Anatidae) together.** Similar lengths,
  different builds: herons are nearly all neck and leg, ducks nearly all body.
- **"Big" depends on the measure.** Sort by [length](/?by=length), then
  [wingspan](/?by=wingspan), then [mass, scale Global](/?by=mass&scale=global).

---

## What BirdView will not tell you

It is a size and proportion tool, not a field guide: no plumage, range maps, behavior, perched
view or individual variation. For identification, use a field guide. For "wait, is a Merlin really
that small?", use this.

---

## Where the numbers come from

Every bird's page gives the source of each number. [How the data is made](data.md) describes each
input, and [one diagram](pipeline.md) shows how they become the drawings. Escaped cage, aviary and
farm birds are [listed separately](/excluded).
