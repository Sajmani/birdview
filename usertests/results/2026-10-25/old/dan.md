# Usability test: Dan, experienced field birder, laptop

## T1. First look

**Answer:** It's a size-comparison tool. It draws North American birds as scaled stick figures, made from published and modeled data, so you can compare species across families. First I'd try one of the "compare some medium birds" links, or look for a way to put two birds I know side by side.

**Confidence:** 4

**Where:** Front page, first screenful: the intro line plus the "New here? Try…" row.

**Pages:** `http://localhost:8081/`

## T2. Merlin and Blue Jay

**Answer:** The Merlin is bigger, mostly in wingspan and weight. Length is about the same.

| | Merlin | Blue Jay | Difference |
|---|---|---|---|
| Length | 11.5 in ~ (modeled) | 10.7 in ~ (modeled) | Merlin about 0.8 in longer (but both are modeled) |
| Wingspan | 23.8 in (measured) | 15.2 in ~ (estimated) | Merlin about 8.6 in wider, roughly 1.5× |
| Weight | 6.6 oz | 3.1 oz | Merlin about twice as heavy |

So my friend is right on length and wrong on everything else. The Merlin is a much bigger bird in hand.

**Confidence:** 4 on wingspan and weight. 3 on length: the site says every length is modeled, and my Sibley puts the Merlin slightly *shorter* than a Blue Jay.

**Where:** A compare page with the two birds. The front page has no search box, and `?q=Merlin` did nothing (it still showed all 1041 species, and find-in-page found no Merlin). I only got there because I know the eBird codes: I copied the "medium birds" compare link and swapped in `merlin`. On the Merlin's own page, length is labelled "sum of the five drawn segments; bill and tail measured, head, neck and trunk modeled".

**Pages:**
- `http://localhost:8081/?q=Merlin` (twice, once with find)
- `http://localhost:8081/compare?species=blujay&species=amekes&species=calthr&species=moudov&species=belkin1&species=chwwid`
- `http://localhost:8081/compare?species=blujay&species=merlin` (two screenshots)
- `http://localhost:8081/species/merlin`

## T3. New York in October

**Answer:** Blue Jay (1 in 24 species observations), American Robin (1 in 35) and American Crow (1 in 36). Black-capped Chickadee and White-throated Sparrow come right after them.

**Confidence:** 4. The page states the numbers and says they come from eBird/iNaturalist records via GBIF. A Blue Jay at the top in October fits the fall jay flight.

**Where:** Front page, Regions and months panel, set to New York and Oct, with Order by Frequency. The month buttons were confusing: every month starts selected, so clicking "Oct" *removes* October instead of picking it. I built the address by hand instead (`region=NY&month=Oct`, plus `by=frequency`, which I took from the "birds around you now" link).

**Pages:**
- `http://localhost:8081/` with the regions panel open (find "New York", then find "Oct")
- `http://localhost:8081/?region=NY&month=Oct&by=frequency`

## T4. Three ducks

**Answer:**
- **Widest wingspan:** Mallard, 34.6 in (measured). Canvasback is 31.3 in (estimated) and Wood Duck 28.0 in.
- **Heaviest:** Canvasback, 2.6 lb. Mallard is 1.9 lb and Wood Duck 1.4 lb.

**Confidence:** 4. The Canvasback's wingspan is dashed (estimated), but it would have to be off by more than 3 in to beat the Mallard.

**Where:** A compare page with all three, on one scale.

**Pages:** `http://localhost:8081/compare?species=mallar3&species=wooduc&species=canvas` (screenshot, 3 screens)

## T5. Solid and dashed

**Answer:** Solid lines are measured and dashed lines are estimated. The head is always estimated and drawn hollow. On the wing, the solid outer parts are the measured wing chord, ticks mark where the chord ends, and the span between the ticks is dashed when the wingspan is estimated. A "~" after a number means the same thing.

**Confidence:** 5. It's spelled out under every compare page.

**Where:** The legend paragraph under the compare drawings (it also appears on species pages).

**Pages:** Same compare pages as T2 and T4.

## T6. Sharp-shinned Hawk

**Answer:** 4.6 oz, labelled measured, from AVONET v7 (eBird sheet, *Accipiter striatus*).

**Confidence:** 4 as an average. Sharpies differ a lot by sex: females are much heavier than males. The page doesn't mention sex anywhere (find "female" found nothing), so 4.6 oz is presumably a pooled average. It fits roughly between a male and a female in my Sibley.

**Where:** The species page, typed straight in with the eBird code.

**Pages:**
- `http://localhost:8081/species/shshaw` (find "Mass")
- `http://localhost:8081/species/shshaw` (find "female")

## E1. What confused or annoyed me

There's no way to look up a bird by name. No search box I could see, and adding `q=` to the address did nothing. I found every bird by guessing eBird codes into the address bar, which most people can't do. Without the codes I'd have been scrolling through 1041 tiles sorted by length. The month buttons work backwards: everything starts on, and clicking October turns it *off*. The compare page shows its numbers only as small labels on the drawing, so it's hard to read the figures off in a row. Weights for strongly dimorphic birds like Sharpies and Merlins are one number, with no note that females run much heavier.

## E2. What I liked

It's honest about its numbers. Every figure says whether it's measured or modeled, and the species page names the source (AVONET, Fox et al.) or the formula. That's exactly what I want, and more than most bird sites give you. Solid versus dashed lines show it at a glance. The compare-on-one-scale view does what it says, and the Merlin against the Blue Jay was a real eye-opener on wingspan. The New York October frequency list, "1 in 24 observations", is useful and makes sense.

## E3. One change

Put a name search box at the top of the front page, and on the compare page ("add a bird…"), that accepts common names, scientific names or eBird codes.

## E4. Ease of use

**2 out of 5.** It was easy once I knew the address tricks, but a regular user would get stuck on the very first task. I would come back for the trustworthy measured-versus-modeled numbers and the side-by-side comparisons, but only once there's a search box.
