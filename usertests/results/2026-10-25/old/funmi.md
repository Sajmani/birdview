# Usability test: Dr. Funmi Okafor (ornithologist, laptop)

About 21 commands in all.

## T1. First look

**Answer:** It draws North American birds as stick figures to scale, using published and modelled data, so you can compare sizes across families. The opening line says straight away that some of the data is modelled, which I liked. The first thing I would do is open **Presentation** to see what the drawing claims to show and what it estimates, and only then compare a few birds.

**Confidence:** 4

**Where:** The front page as it first loads: the header text, five collapsed panels, a filter-chip bar, a histogram, and then "Length Tier 1" with hummingbirds.

**Pages:** `http://localhost:8081/` (screenshot)

## T2. Merlin and Blue Jay

**Answer:** The Merlin is bigger, mostly in wingspan and weight. Its length is only a little greater.

| | Merlin | Blue Jay | Difference |
|---|---|---|---|
| Length | 11.5 in (modelled) | 10.7 in (modelled) | Merlin about 0.8 in longer (~7%) |
| Wingspan | 23.8 in (measured, Fox et al.; range 23.1–24.4) | 15.2 in (estimated) | Merlin about 8.6 in wider (~1.6×) |
| Mass | 6.6 oz (AVONET) | 3.1 oz (AVONET) | Merlin about 2.1× heavier |

So the friend is roughly right on length and wrong on everything else.

The lengths are both modelled: they are the sum of drawn segments. I would not quote an 0.8 in difference between two modelled numbers as if it were real.

The mass is a single value with no sex given. Merlins are strongly reverse-dimorphic, so a female Merlin would outweigh a jay by even more than this.

**Confidence:** 4 for the direction and for wingspan and mass. 3 for the size of the length difference.

**Where:** I opened a compare view of the two birds. The figures are in the drawing labels; the text version showed only "[drawing]", so I needed a screenshot. I then opened the Merlin's species page, which names the source for every number. That is excellent.

**Pages:**
- `/compare?species=merlin&species=blujay` (text, then screenshots at 1 and 2 screens)
- `/species/merlin`

## T3. New York in October

**Answer:** Blue Jay (1 in 24 species observations in NY in October), American Robin (1 in 35) and American Crow (1 in 36). Close behind: Black-capped Chickadee, White-throated Sparrow and Canada Goose.

**Confidence:** 4. The site is careful to say this is a share of what people record, not abundance, which is the right framing.

**Where:** In Regions and months I chose New York. The month buttons all start selected and clicking one *removes* it. I only understood that after reading the help text, which says "one month is two clicks": All, then Oct. Then I went to Controls and set Order by to Frequency. The view switched itself to Global scale and a 10× section step, which was fine but unannounced.

**Pages:**
- `/?region=NY` (regions panel)
- `/?month=&region=NY`
- `/?month=Oct&region=NY` (controls panel)
- `/?by=frequency&month=Oct&region=NY` (screenshot)

## T4. Three ducks

**Answer:** **Mallard** has the widest wingspan, at 34.6 in (measured). Canvasback is 31.3 in (estimated) and Wood Duck 28.0 in (measured). **Canvasback** is the heaviest, at 2.6 lb. Mallard is 1.9 lb and Wood Duck 1.4 lb.

**Confidence:** 4. The mass ranking is solid. The wingspan ranking compares a measured Mallard with an estimated Canvasback, but the gap of 3.3 in is large enough that I'd accept it.

**Where:** The compare view. Getting there was the hard part. There is no search box I could find: adding `q=duck` or `q=Mallard` to the address did nothing, and all 1041 tiles stayed. In the end I used the "herons and ducks" shortcut, turned on Compare birds, and picked from the list. I had to read species codes (`mallar3`, `canvas`, `wooduc`) off the tiles.

**Pages:**
- `/?q=duck` (controls panel)
- `/?q=Mallard&species=`
- `/?family=Ardeidae&family=Anatidae&species=`
- `/?family=Ardeidae&family=Anatidae&species=mallar3`
- `/compare?species=mallar3&species=wooduc&species=canvas` (screenshot)

## T5. Solid and dashed

**Answer:**
- **Solid** lines are measured values. Bill and tail come from AVONET.
- **Dashed** lines are estimated. Neck and body are always dashed, and the head is estimated too but drawn as a hollow oval.
- On the wing bar, the solid outer ends are the measured wing chord, and ticks mark where it ends. The middle span is dashed when the wingspan is modelled, which the site says is about two thirds of the birds.
- A "~" after a number marks it as estimated.

**Confidence:** 5

**Where:** The Presentation panel on the front page. The same legend is repeated under every compare view and species page.

**Pages:** `/` (Presentation panel opened, find "dashed")

## T6. Sharp-shinned Hawk

**Answer:** **4.6 oz (about 130 g)**, measured, from AVONET v7 (sheet AVONET2_eBird, *Accipiter striatus*).

**Confidence:** 4 that this is what the site says. As a biologist I'd add a caveat: Sharp-shins are among the most dimorphic raptors, with males around 100 g and females around 170 g. A single species mean describes almost no actual bird. The site doesn't say whether the figure is a mean of the sexes or what its sample size is.

**Where:** The species page, "Measurements and sources".

**Pages:** `/species/shshaw`

---

## E1. What confused or annoyed me

- **No search by name.** This was the biggest friction. To look at three ducks I had to go through a family shortcut and picking mode, and read four-to-seven-letter codes. Typing "Canvasback" should just work.
- **Month toggles work backwards.** All months start on and a click turns one *off*. It is documented, but my first instinct, clicking "Oct", would have given me eleven months.
- **Figures live only in the drawings.** On the compare page the numbers are labels inside the figure. The text and accessible version just said "[drawing]". I couldn't copy them into a spreadsheet without opening each species page.
- **Odd sort highlight.** On the Merlin/Blue Jay compare page, both "Bill" and "Tail" were highlighted in Sort by, so I couldn't tell how it was sorted.
- **Small wording slips.** The species pages say "measured measured" and "modeled estimated". It reads like a badge and its tooltip merged together.
- **One mass per species, no sex, no n.** For raptors and some ducks that's a real limitation, and it isn't flagged the way the modelled lengths are.
- **Neck lengths.** The neck lengths look strange (a 4.3 in neck on a Wood Duck). The site is honest that the neck/body split is a drawing convention, but the numbers are still printed as if they meant something. I'd rather not see them printed at all.
- **Thin region thresholds.** "1 record per million" puts Merlin in Hawaii and Midway Atoll. That's defensible but a little surprising.

## E2. What I liked

- **It is honest about measured versus modelled.** That's the thing I care about most, and it's rare. The site says outright that no published source gives body length for North American birds, so every length is modelled.
- **Per-number provenance on species pages.** Each number has its dataset, sheet and taxon. Modelled values show their formula, e.g. "14.0329 × mass^0.37, calibrated against skeletal data" and "isometric, unvalidated". I could cite AVONET and Fox et al. directly from this.
- **Wing chord shown as the measured part of the wingspan.** It's a clever and correct way to show what is known.
- **The frequency framing.** It says "how much of what people record", not abundance, and notes that "a plain count says where the birdwatchers are". That's exactly right.
- **Full credits with DOIs and licences** on the front page.

## E3. One change

Add a name search box, for common or scientific names, that filters the tiles and lets me add a bird to the comparison directly. Close second: under each compare drawing, a plain table of the numbers, with measured/modelled marked and min–max or a sex split where AVONET has it.

## E4. Ease of use

**3 / 5.** Once I was on a compare page or a species page it was clear and trustworthy. Getting there was the slow part, mainly because there's no search and the month filter works backwards. Yes, I would come back, mainly for the sourced species pages and to show students the difference between measured and modelled data. I'd want search before recommending it widely.
