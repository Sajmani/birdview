# Usability test: Lena (birder from Leipzig, on a phone)

## T1. First look
**Answer:** It is a site that draws birds as stick figures to scale, so you can compare their sizes. First I would tap one of the comparison links ("medium birds") or search for a bird I know. I also saw right away "Units in / oz", so I knew I would need to find where to change that.
**Confidence:** 4
**Where:** front page, first screen
**Visited:** http://localhost:8080/ (screenshot)

## T2. Merlin and Blue Jay
**Answer:** The Merlin is bigger. A little longer, but much bigger in wingspan and about twice as heavy.
- Length: Merlin 29.1 cm vs Blue Jay 27.1 cm, so about 2 cm longer
- Wingspan: 60.4 cm vs 38.6 cm, so about 22 cm wider (roughly 1.5×)
- Weight: 188 g vs 88 g, so about 100 g heavier (a bit more than 2×)

**Confidence:** 5. The species pages say which numbers are measured and which are modeled. Both lengths are "modeled", and so is the Blue Jay's wingspan.
**Where:** the two species pages, after switching to cm / g in Controls. The compare page showed the two drawings side by side, but on my phone the numbers on the drawings were far too small to read.
**Visited:**
- http://localhost:8080/?q=merlin
- http://localhost:8080/?q=merlin (Controls opened, to find units)
- http://localhost:8080/compare?species=merlin&species=blujay&units=metric (text + screenshot)
- http://localhost:8080/?species=merlin&species=blujay&units=metric
- http://localhost:8080/species/merlin?species=merlin&species=blujay&units=metric
- http://localhost:8080/species/blujay?species=merlin&species=blujay&units=metric (twice)

## T3. New York in October
**Answer:** Blue Jay (1 in 24 observations), American Robin (1 in 35), American Crow (1 in 36). Black-capped Chickadee, White-throated Sparrow and Canada Goose come right after.
**Confidence:** 4. The list was clear once I had it. I am less sure I set the month filter the way the site means it to be used.
**Where:** main list, Region = New York, Month = Oct, Order by Frequency.
**Visited:**
- http://localhost:8080/?units=metric (Regions and months opened)
- http://localhost:8080/?units=metric&region=NY (Regions opened, found the month buttons)
- http://localhost:8080/?month=Oct&region=NY&units=metric&by=frequency

The month buttons confused me. With "All" months selected, the "Oct" button's link would have *removed* October and kept the other eleven. To get only October I would have had to tap eleven other months. I typed the address with only October instead.

## T4. Three ducks
**Answer:** The **Mallard** has the widest wingspan (87.9 cm, vs Canvasback 79.4 cm and Wood Duck 71.1 cm). The **Canvasback** is the heaviest (1202 g, vs Mallard 843 g and Wood Duck 658 g).
**Confidence:** 4
**Where:** search results, one search per duck. I did not manage to get them truly "side by side": I could not see how to pick birds for Compare from my phone without more tapping, so I searched each one alone.
**Visited:**
- http://localhost:8080/?q=mallard&units=metric
- http://localhost:8080/?q=wood duck&units=metric (failed, the space broke it)
- http://localhost:8080/?q=wood+duck&units=metric (twice)
- http://localhost:8080/?q=canvasback&units=metric

## T5. Solid and dashed
**Answer:** Solid lines are measured. Dashed lines are estimated, and so is the hollow head. On the wing, the solid ends are the measured wing chord, and the ticks show where it ends. The cube is the bird's mass drawn as water at the same scale.
**Confidence:** 5
**Where:** the text under the drawings on the compare page and on each species page.
**Visited:** same as T2 (compare page, Merlin page)

## T6. Sharp-shinned Hawk
**Answer:** 131 g (length 29.1 cm, wingspan 56.7 cm)
**Confidence:** 5
**Where:** search result tile for Sharp-shinned Hawk (*Accipiter striatus*)
**Visited:**
- http://localhost:8080/?q=sharp-shinned&units=metric (twice, the second time to check it was the right bird)

About 22 commands in total.

## E1. What confused or annoyed you
Everything starts in inches and ounces. I had to open "Controls" to find cm / g, and I could not tell if the site would remember it. The compare page is the obvious place to answer "which is bigger", but on a phone the labels on the drawings are tiny, maybe 5 px, and I could not read them. I had to open each bird's page to get the numbers. The month buttons work backwards from what I expected: tapping "Oct" turns October *off*. The compare page's sort buttons also looked strange, with "Bill" and "Tail" highlighted when I had not chosen them. It also took me a while to realise I can use World (IOC) names. That is good for me, but it is hidden in Controls.

## E2. What you liked
The idea is lovely: real sizes, side by side, and a cube of water for weight. I really like that every number says whether it is measured or modeled and where it comes from. As a scientist I trust that. Frequency for one state and one month is exactly what I want as a visitor ("what will I see this week?"). The search is fast and also works on scientific names.

## E3. One change
Show the numbers (length, wingspan, mass) as normal readable text under each bird on the compare page. While you are at it, choose cm / g automatically for visitors outside the US, or put the unit switch next to the search box.

## E4. Ease of use
**3 / 5.** Once I knew where things were it was quick, and the data is very good. But on a phone the important comparison view is hard to read, and the filters needed some guessing. Yes, I would come back, especially to check what is common in a new state each month. I would set it to metric and World names first.
