# Usability test: Dr. Funmi Okafor (ornithologist, laptop)

## T1. First look

**Answer:** It draws North American birds as scaled stick figures so you can compare their sizes, with the measurements on each card. First I would click into one bird to see where its numbers come from, since the footer says "every length here is modeled," and I want to know what that means before I trust anything.

**Confidence:** 4

**Where:** Front page, first screen.

**Pages:** `http://localhost:8080/`

## T2. Merlin and Blue Jay

**Answer:** The Merlin is bigger. On length it is only a little bigger: 11.5 in vs 10.7 in, about 0.8 in more (both lengths are marked `~`, modeled). On wingspan it is much bigger: 23.8 in vs 15.2 in, about 8.6 in or 57% more. The Merlin's wingspan is unmarked, so measured; the Blue Jay's is `~`, modeled. By weight the Merlin is about twice as heavy: 6.6 oz vs 3.1 oz. So your friend is right on body length and wrong on everything else. Note that the site gives one mass for the Merlin with no sex split, and in falcons the sexes differ a lot.

**Confidence:** 4. The numbers are clear; I trust the masses and the Merlin's wingspan, and the lengths less.

**Where:** I searched "merlin", pinned it, pinned Blue Jay, and viewed the pinned birds.

**Pages:** `/?find=merlin`, `/?find=merlin&bird=merlin`, `/?pin=merlin&pin=blujay&show=pinned`

## T3. New York in October

**Answer:** Blue Jay (1 in 24 species observations), American Robin (1 in 35), and American Crow (1 in 36). Black-capped Chickadee, White-throated Sparrow, and Canada Goose come right after.

**Confidence:** 4. The page heading says "Showing birds in New York in October." I would have trusted this more if I had been able to set the filters myself.

**Where:** I could not get the Region menu on the front page to open. The help page ("more ideas" → `/docs/using`) has a "New York in October" example link, and I used that.

**Pages:** `/` (Region menu tried four times), `/docs/using`, `/?region=NY&month=Oct`

## T4. Three ducks

**Answer:** The Mallard has the widest wingspan, 34.6 in (measured), against 31.3 in for the Canvasback (modeled, `~`) and 28.0 in for the Wood Duck (measured). The Canvasback is the heaviest at 2.6 lb, against 1.9 lb for the Mallard and 1.4 lb for the Wood Duck. Since the Canvasback's wingspan is modeled, the Mallard-versus-Canvasback ranking rests partly on an estimate.

**Confidence:** 4

**Where:** I searched for each duck, pinned all three, and sorted by length. They appear side by side in one row.

**Pages:** `/?find=canvasback`, `/?find=wood+duck`, `/?pin=mallar3&pin=wooduc&pin=canvas&show=pinned&by=length`

## T5. Solid and dashed

**Answer:** Solid lines are measured and dashed lines are estimated. The neck and body are always dashed because they are modeled, the first from a family ratio and the second from mass. On the wing, the solid outer ends are the measured wing chord, out to the ticks. The middle section is dashed when the wingspan is estimated. You can see this on the Canvasback, whose wing is dashed between the ticks, while the Mallard's is solid.

**Confidence:** 5

**Where:** The help page, and the legend under the drawing on each species page.

**Pages:** `/docs/using`

## T6. Sharp-shinned Hawk

**Answer:** 4.6 oz, marked measured, from AVONET v7. This is a single species value. Sharp-shins are among the most sexually dimorphic North American raptors, with females much heavier than males, so the figure is close to neither sex. The site does not say whether it is a mean of both sexes, and I would want that before citing it.

**Confidence:** 4 that this is what the site says; 3 that it answers the question as a biologist would.

**Where:** Species page, "Measurements and sources" table.

**Pages:** `/?find=sharp-shinned`, `/species/shshaw`

## T7. A heron's parts

**Answer:** The bill is 5.7 in and the tail is 6.5 in, both marked measured, from AVONET v7 (AVONET2_eBird sheet). The page does not say which bill measurement this is: AVONET has culmen-to-skull and bill-from-nares, and I could not tell which one this is. Separately, the total length of 34.7 in is labeled modeled, as the sum of drawn segments, and that is short for a Great Blue Heron. The site is honest that it is an estimate.

**Confidence:** 4

**Where:** Great Blue Heron species page, top line and the sources table.

**Pages:** `/species/grbher3`

## T8. The wing chord

**Answer:** The green horizontal crossbar is the wingspan. The wing chord is the solid stretch of that bar from each wingtip inward to the small tick mark. On this heron it is 18.4 in, measured, from AVONET. It does not say whether that is flattened or natural chord.

**Confidence:** 4

**Where:** The legend under the drawing ("the solid ends are the measured wing chord, and the ticks mark where it ends") and the sources table.

**Pages:** `/species/grbher3` (screenshot)

## E1. What confused or annoyed me

The Region menu would not open for me at all. I only got to New York because the help page had that exact example as a link. Mass jumps between ounces and pounds depending on which birds are on screen, which is a nuisance when comparing. My bigger complaint is about the trait values. They are single species numbers with no sex, no sample size, and no definition of which bill length or which wing chord is meant. For a Sharp-shinned Hawk or a Merlin, a single mass hides most of the story. With Merlin and Blue Jay pinned, the two cards sat in different "frequency tiers," one above the other, so they were not side by side.

## E2. What I liked

The honesty. Every number carries `measured` or `modeled`, the `~` follows modeled values onto the cards, and the drawing itself switches to dashed lines for anything estimated. Each value names its source, down to the AVONET sheet, and gives the model formula with exponents, even when the honest word is "unvalidated." Reported wingspan ranges from the Fox database are a nice touch. The citations in the footer have DOIs and licenses, so I could actually cite the underlying data. It also refuses to call total length a measurement, which most field guides quietly do.

## E3. One change

On the species page, say exactly which measurement each AVONET value is (culmen vs nares bill length, wing chord type), and give the sample size and sex basis. Show male and female masses where AVONET or the literature has them. That would turn the page from honest into citable.

## E4. Ease of use

3 out of 5. Search, pinning, and the species pages were easy, but I hit a dead end with the Region menu and needed the help page to get around it. I would come back, mainly for the species source tables and to show students what is measured and what is modeled. I would cite AVONET and Fox directly rather than BirdView.
