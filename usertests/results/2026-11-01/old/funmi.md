# Usability test: Dr. Funmi Okafor (ornithologist, laptop)

## T1. First look

**Answer:** A size-comparison tool: each North American species drawn as a stick figure to one scale, with length, bill, tail, wingspan, chord, and mass on every tile. First I would click a bird to see where its numbers come from, since the intro says I can.

**Confidence:** 4

**Where:** Front page, first screenful.

**Pages:** http://localhost:8081/

## T2. Merlin and Blue Jay

**Answer:** The Merlin is bigger, and the difference is mostly in wing and weight. Length is only a little more: 11.5 in vs 10.7 in, about 0.8 in, and both are modeled (marked ~). Wingspan is much wider: 23.8 in vs 15.2 in, about 8.6 in or roughly 55% more. The Merlin's span is a measurement; the Jay's is marked ~, so it is modeled. Weight is about double: 6.6 oz vs 3.1 oz. So "about the size of a Blue Jay" holds for length and fails for the other two.

**Confidence:** 4. The numbers are clear. I knocked a point off because the lengths are modeled and a single mass hides the Merlin's sexual dimorphism.

**Where:** Found Merlin with the find box, pinned it and the Blue Jay, then clicked Pinned.

**Pages:** http://localhost:8081/?find=Merlin, http://localhost:8081/?find=Merlin&bird=merlin, http://localhost:8081/?pin=merlin&pin=blujay&show=pinned (text and screenshot)

## T3. New York in October

**Answer:** Blue Jay (1 in 24 species observations), American Robin (1 in 35), and American Crow (1 in 36). Next come Black-capped Chickadee, White-throated Sparrow, and Canada Goose.

**Confidence:** 4

**Where:** I couldn't get to this from the front page. The Month and Region dropdowns would not open for me (I tried several times). I found a ready-made "New York in October" link under "Some things to try" on the Using BirdView page, and the heading then read "Showing birds in New York in October".

**Pages:** http://localhost:8081/ (tried to open the Region/Month panels four times), http://localhost:8081/docs/using (twice with find, once in full), http://localhost:8081/?region=NY&month=Oct (screenshot)

## T4. Three ducks

**Answer:** The Mallard has the widest wingspan, 34.6 in (measured), against 31.3 in for the Canvasback (modeled, ~) and 28.0 in for the Wood Duck (measured). The Canvasback is the heaviest at 2.6 lb, against 1.9 lb for the Mallard and 1.4 lb for the Wood Duck.

**Confidence:** 5 for mass. 4 for wingspan, because the Canvasback figure is modeled. Still, it would take a 3 in error to change the answer.

**Where:** Searched each by name, then pinned all three and sorted by wingspan.

**Pages:** http://localhost:8081/?find=Wood%20Duck, http://localhost:8081/?find=Canvasback, http://localhost:8081/?pin=mallar3&pin=wooduc&pin=canvas&show=pinned&by=wingspan (screenshot)

## T5. Solid and dashed

**Answer:** Solid lines are measured and dashed lines are estimated. Bill, tail, and the wing chord (the solid outer ends of the wing bar, up to the ticks) are measured. Neck and body are dashed because they are modeled from mass or a family ratio, and the head is drawn hollow because it is modeled. The middle of the wing bar is dashed when the wingspan itself is modeled, as on the Canvasback.

**Confidence:** 5

**Where:** The "Reading a tile" table on the Using BirdView page, repeated in the note under every species drawing.

**Pages:** http://localhost:8081/docs/using, http://localhost:8081/species/shshaw, http://localhost:8081/species/grbher3

## T6. Sharp-shinned Hawk

**Answer:** 4.6 oz, marked measured, from AVONET v7 (sheet AVONET2_eBird). As someone who knows the species: that number is a species mean. Female sharpies are much heavier than males, so no single bird will weigh 4.6 oz, and the site doesn't say that.

**Confidence:** 4 that the site says 4.6 oz. For what a given hawk weighs, it's only an average.

**Where:** Found by name, then the species page's "Measurements and sources" table.

**Pages:** http://localhost:8081/?find=Sharp-shinned, http://localhost:8081/species/shshaw

## T7. A heron's parts

**Answer:** Bill 5.7 in and tail 6.5 in, both measured, from AVONET v7.

**Confidence:** 5

**Where:** Great Blue Heron's own page. They're in the labels beside the drawing and in the Measurements and sources table.

**Pages:** http://localhost:8081/species/grbher3 (text and screenshot)

## T8. The wing chord

**Answer:** The wing chord is the solid green outer section of the wing crossbar, from the wingtip in to the tick mark. On the heron's drawing the left one is labeled "Wing chord 18.4 in". It is 18.4 in, measured, from AVONET.

**Confidence:** 5

**Where:** Same page: the drawing label and the explanatory note under it.

**Pages:** http://localhost:8081/species/grbher3

## E1. What confused or annoyed me

The filter dropdowns. I wanted New York and October, and I could not get the Region or Month panel to open at all. I only got there because the help page happened to have a "New York in October" example link. If that link hadn't been there I would have given up on T3.

Comparing pinned birds is clumsy. Three ducks are one column of three tall tiles, each under its own "tier" heading with a single bird in it, and I had to scroll to compare them. I'd like them in a row with a shared baseline. The tier headings mean nothing for three birds.

On the science: a total length of 34.7 in for a Great Blue Heron is well short of what field guides publish, because it is a sum of modeled segments. To be fair, it is labeled modeled and explained, and that matters. But a birder will read it as "length". Mass and wing values are species means with no sex split, which matters for raptors like the sharpie and the Merlin. The source column also says "measured measured" on every row.

The labels on the species drawing are very small on a laptop. I could read them, but I had to squint.

## E2. What I liked

It is honest about provenance, more than almost anything else in this space. Every number on the species page is tagged measured or modeled, and the source names the dataset, version, and sheet. Modeled values give their formula, and one is even marked "unvalidated". The wingspan gives a reported range. The ~ on tiles and the solid/dashed convention follow from the same idea. The footer gives full citations with DOIs and licenses, so I could cite it. Choosing the wing chord as the measured part of the wing is a sensible use of AVONET.

## E3. One change

Make the Region and Month filters work, or at least easy to find. They're how you get the "what's around me" question answered. After that: show pinned birds side by side on one baseline instead of one per tier.

## E4. Ease of use

3 out of 5. Reading tiles and species pages was easy. The filters defeated me, and comparing birds took more scrolling than it should. I would come back, mainly for the per-number provenance, and I might point students to the species pages. I'd want the length caveat in front of anyone who quotes those totals.
