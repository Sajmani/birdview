# Using BirdView

BirdView draws birds as stick figures to scale, so a hummingbird beside a swan looks as it would in
life. Every setting is in the page address, so you can bookmark or share any view.

---

## The page

Under the title, **New here? Try** offers a few views to start with, and **more ideas** leads to
[Some things to try](#some-things-to-try).

The page opens on the birds recorded this month, most often recorded first, the first 100 of
them. Every bird shown is drawn to one scale.

Above the birds, a **find box** finds birds by English or scientific name, code, family, or order
as you type, and lists them once there are 10 or fewer. **Pin** a bird, from that list or from its
tile, to keep it on the page whatever the settings; the pinned birds are listed under the box.

The **count line** says how many birds are drawn and names every setting. Click a setting to change
it. Under it, a **histogram** has one bar per tier, each linking to its tier. Each tier heading
names the order and the tier, such as **Tail Tier 3**, its range, and how many birds it holds.

---

## Reading a tile

Each tile has the bird's names and family, a stick figure to scale, a cube for its mass, its length,
wingspan, and mass, and how many of the species observations in the selected regions are of this
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

## The settings

- **Show:** the first 100 birds (type another number), **All** the birds the settings match, or
  **Pinned only**. **Clear pins** unpins every bird; **Reset** returns every setting to its default
  and keeps the pins.
- **Order by:** **Frequency**, how often a bird is recorded, most often first (see
  [below](#sorting-by-how-often-a-bird-is-recorded)), or **Length**, **Wingspan**, **Mass**,
  **Bill**, or **Tail**, smallest first. Choose the order again to reverse it. A Northern Harrier
  and a Cooper's Hawk sort close together by length and mass, and far apart by wingspan.
- **Month:** **Now**, the default, is this month; **All** is every month. A bird with too few
  records to tell is shown in every month, and says so.
- **Region:** states, provinces, and stretches of open sea. A bird is listed in a region when at
  least 1 in a million of the region's bird records are of it. **Here** chooses the region you are
  in; your browser works it out, so your location never reaches the server. Here needs JavaScript.
- **Family:** grouped by taxonomic order, in field-guide sequence.
- **Habitat** and **Diet:** where a bird mostly lives and what it mostly eats, per species, from
  **AVONET**. BirdView adds a few more and marks them as its own; [everything BirdView adds](/added)
  is listed in one place.
- **Units:** inches and ounces, or centimeters and grams.
- **Names:** **North American** names, such as Dovekie and Rough-legged Hawk, or **World** names,
  AviList's: Little Auk, Rough-legged Buzzard.

Choosing one option selects only it, choosing more adds them, and choosing a selected option again
removes it. Nothing selected means all. The settings combine.

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

- **[The birds around you now](/#here-and-now).** The birds most often recorded where you are this
  month. Without sharing your location, try
  [New York in October](/?region=NY&month=Oct).
- **Compare some [small birds](/?by=length&pin=leasan&pin=houspa&pin=dowwoo&pin=houfin&pin=barswa&pin=easpho&show=pinned),
  [medium birds](/?by=length&pin=blujay&pin=amekes&pin=calthr&pin=moudov&pin=belkin1&pin=chwwid&show=pinned),
  [large birds](/?by=length&pin=turvul&pin=gbbgul&pin=greegr&pin=cangoo&pin=grecor&pin=rosspo1&show=pinned),
  [huge birds](/?by=length&pin=calcon&pin=amwpel&pin=mutswa&pin=magfri&pin=baleag&pin=bkfalb&show=pinned).**
  Each pins six birds and shows only them, by length.
  - *Small:* Least Sandpiper, House Sparrow, Downy Woodpecker, House Finch, Barn Swallow, and
    Eastern Phoebe, all 141 to 158 mm long. The swallow is the lightest, 18 g, yet has the widest
    wings, 325 mm against 234 mm for the finch.
  - *Medium:* Blue Jay, American Kestrel, California Thrasher, Mourning Dove, Belted Kingfisher,
    and Chuck-will's-widow, all 272 to 292 mm long. Their wingspans run from 307 mm for the
    thrasher to 655 mm for the nightjar, and their bills from 15 mm for the nightjar to 59 mm for
    the kingfisher.
  - *Large:* Turkey Vulture, Great Black-backed Gull, Great Egret, Canada Goose, Great Cormorant,
    and Roseate Spoonbill, all 614 to 728 mm long. The vulture, the shortest, has the widest
    wings, 1805 mm, against 1280 mm for the spoonbill, and the goose weighs more than three times
    as much as the egret.
  - *Huge:* California Condor, American White Pelican, Mute Swan, Magnificent Frigatebird, Bald
    Eagle, and Black-footed Albatross, all with wings over two meters. The albatross is the
    shortest, 709 mm, yet spans 2265 mm; the frigatebird weighs about a seventh of the swan.
- **See all [herons and ducks](/?family=Ardeidae&family=Anatidae&month=all) together.** Similar
  lengths, different builds: herons are nearly all neck and leg, ducks nearly all body.
- **"Big" depends on the measure.** Order every bird by [length](/?by=length&month=all&show=all),
  then [wingspan](/?by=wingspan&month=all&show=all), then [mass](/?by=mass&month=all&show=all).

---

## What BirdView will not tell you

It is a size and proportion tool, not a field guide: no plumage, range maps, behavior, perched
view, or individual variation. For identification, use a field guide. For "wait, is a Merlin really
that small?", use this.

---

## Where the numbers come from

Every bird's page gives the source of each number. [How the data is made](data.md) describes each
input, and [one diagram](pipeline.md) shows how they become the drawings. Escaped cage, aviary, and
farm birds are [listed separately](/excluded).
