# BirdView usability test: Dan (experienced field birder, laptop)

About 21 commands in all.

## T1. First look

**Answer:** It's a reference site that draws North American birds as stick figures to scale so you can compare their sizes. There are 1,041 species, ordered by length from the hummingbirds up. First I'd type a bird I know into the search box, or click one of the "compare some medium birds" links to see what a comparison looks like.

**Confidence:** 4

**Where:** Front page, first screenful. The one-line description and the "New here? Try…" row say it plainly. Five collapsed panels and a row of filter chips sit between me and the birds. On first load the drawings are cut off at the bottom of the screen.

**Pages:** `http://localhost:8080/` (screenshot)

## T2. Merlin and Blue Jay

**Answer:** The Merlin is bigger, and clearly so, except in length.
- Length: Merlin 11.5 in, Blue Jay 10.7 in. Merlin is about 0.8 in (~7%) longer, so nearly the same. Both lengths are marked "~" (estimated).
- Wingspan: Merlin 23.8 in, Blue Jay 15.2 in. Merlin is about 8.6 in wider (~1.6×). The Merlin figure is measured; the Blue Jay's is estimated ("~", dashed line).
- Weight: Merlin 6.6 oz, Blue Jay 3.1 oz. Merlin is about twice as heavy (+3.5 oz).

So your friend is right on length and wrong on everything else. That fits what I see in the field: a Merlin is a compact, heavy-chested falcon, and a jay is mostly tail and crest.

**Confidence:** 4. The numbers are plain. I docked one point because the jay's wingspan is modeled; Sibley has it a bit bigger.

**Where:** Search box ("Merlin") gave one tile with length, wingspan and mass. I copied the URL pattern of the "medium birds" compare link (which already includes the Blue Jay) to put the two side by side.

**Pages:**
- `http://localhost:8080/?q=Merlin`
- `http://localhost:8080/compare?species=merlin&species=blujay` (page text and two screenshots)
- `http://localhost:8080/?q=Blue+Jay` (my first try with a space in it failed outright)

## T3. New York in October

**Answer:** Blue Jay (1 in 24 of observations), American Robin (1 in 35), and American Crow (1 in 36). Right behind them: Black-capped Chickadee (1 in 36), White-throated Sparrow and Canada Goose (1 in 37 each).

**Confidence:** 4 for the list, which looks right to anyone who has birded the Northeast in fall. Only 3 for whether I used the site as intended: I had to build the address myself.

**Where:** The "Regions and months" panel has month buttons. With all months showing, clicking **Oct** *removed* October and showed every month except October (tiles then said "Nov-Sep"). That's the opposite of what I wanted. I didn't find an obvious "only this month" control. So I put the address together by hand, using month=Oct, region=NY and the by=frequency setting from the "birds around you now" link. That gave "344 of 1041 species … Order by Frequency … Month Oct · Region New York", with the most often recorded birds first.

**Pages:**
- `http://localhost:8080/ --open regions` (the list of regions is enormous; the output was truncated)
- `http://localhost:8080/ --open regions --find Oct`
- `http://localhost:8080/?month=Jan&…&month=Sep&month=Nov&month=Dec` (what clicking Oct gave me)
- `http://localhost:8080/?month=Oct&region=NY&by=frequency&scale=global`

## T4. Three ducks

**Answer:**
- **Widest wingspan: Mallard**, 34.6 in (measured). Next come Canvasback at 31.3 in (estimated) and Wood Duck at 28.0 in.
- **Heaviest: Canvasback**, 2.6 lb. Mallard is 1.9 lb, Wood Duck 1.4 lb.

**Confidence:** 4. The Canvasback wingspan is an estimate, but it's 3 in short of the Mallard, so I'd still trust the ranking.

**Where:** Compare page. To get the species codes I had to search, because searching "duck" didn't turn up the Mallard. I found the Mallard by name, then found Wood Duck and Canvasback in the Ducks family list (from "herons and ducks").

**Pages:**
- `http://localhost:8080/?q=duck`
- `http://localhost:8080/?q=Mallard`
- `http://localhost:8080/?family=Anatidae` (three finds)
- `http://localhost:8080/compare?species=mallar3&species=wooduc&species=canvas` (screenshot)

"Side by side" turned out to be stacked one above the other. I had to scroll through three full-width cards. Weights came out in pounds here, while everything else so far was in ounces.

## T5. Solid and dashed

**Answer:** Solid lines are measured; dashed lines are estimated (the hollow head is estimated too). On the wing, the solid ends are the measured wing chord, and the tick marks show where the chord ends. Numbers with a "~" are the estimated ones.

**Confidence:** 5

**Where:** The legend under the drawings on the compare page and the species page. It's also in the "Presentation" panel title ("what is estimated").

**Pages:** `http://localhost:8080/compare?species=merlin&species=blujay`, `http://localhost:8080/species/shshaw?q=Sharp-shinned`

## T6. Sharp-shinned Hawk

**Answer:** 4.6 oz (about 130 g). The site says this is measured, from AVONET v7.

**Confidence:** 3. The number has a source, which I appreciate. But Sharpies are about as sexually dimorphic as North American raptors get: females can weigh nearly twice what males do. One average with no sex split or range hides that. The wingspan line gives a range (21.0–23.7 in), so why not the mass?

**Where:** Search tile, then the species page's "Measurements and sources" section.

**Pages:**
- `http://localhost:8080/?q=Sharp-shinned` (two finds)
- `http://localhost:8080/species/shshaw?q=Sharp-shinned`

## E1. What confused or annoyed me

The month buttons work backwards. Clicking "Oct" took October *out*, and I only caught it because the tiles said "Nov-Sep". Someone less fussy would come away with the wrong birds. Typing "duck" into the search box didn't find the Mallard. Units flip between oz and lb depending on the page. The compare page stacks the birds instead of putting them side by side, and the top of the front page is five collapsed panels plus eleven filter chips before you see a single bird. Single averages for dimorphic species (Sharpie, Merlin) bother me. And "Top region: Yukon" for Sharp-shinned Hawk is technically explained as a share of records, but it reads oddly.

## E2. What I liked

Every number traces to a source, and the species page says which ones are measured and which are modeled, down to the formula. That's rare and exactly what I want. The solid-versus-dashed convention is clear once you read it. The drawings really do make size differences jump out: you see at a glance that the Merlin's wings dwarf the jay's. The New York October list rang true.

## E3. One change

Make clicking a month mean "show only this month" when all months are showing, or add an obvious "only" option. Close second: show a mass range, or male/female values, wherever the source has them.

## E4. Ease of use

**3 out of 5.** Once I knew the tricks (search box, compare page, species page) it was quick, and the sourcing earns my trust. But I had to hand-build an address to get a simple "this state, this month" list, and that's a basic question for a birder. I'd come back to settle size arguments and to check where a number comes from. I wouldn't use it for trip planning until the month filter makes sense.
