# Usability test: Lena (birder from Leipzig, phone)

About 23 commands. I could not open any of the drop-down menus (Units, Month, Region, Names), so I
read everything in inches, ounces, and pounds and converted in my head. The centimeter and gram
values below are my own conversions.

## T1. First look

**Answer:** The site draws birds as stick figures at the same scale so you can compare their sizes
and shapes. Above the fold I see "BirdView", a short explanation, "New here? Try…" links, filter
buttons, a small bar chart, and the first bird (American Robin) with a very small drawing. First I
would switch Units to metric, because everything is in inches.

**Confidence:** 4

**Where:** Front page, first screen.

**Pages:** `http://localhost:8080/` (screenshot and text)

## T2. Merlin and Blue Jay

**Answer:** The Merlin is bigger, and most of the difference is in wingspan and weight:

| | Merlin | Blue Jay | Difference |
| --- | --- | --- | --- |
| Length | 11.5 in (~29 cm) | 10.7 in (~27 cm) | Merlin ~2 cm longer, about the same |
| Wingspan | 23.8 in (~60 cm) | 15.2 in (~39 cm) | Merlin ~22 cm wider, about 1.5× |
| Weight | 6.6 oz (~187 g) | 3.1 oz (~88 g) | Merlin about twice as heavy |

My friend is right about the length but not about the size overall.

**Confidence:** 4. The numbers are clear. I trust the length less because of the "~" mark, and the
site says lengths are modeled.

**Where:** I searched "Merlin" in the "Find a bird to pin" box, pinned it, did the same with Blue
Jay, and then tapped "Pinned (2)" to see them together.

**Pages:**
- `http://localhost:8080/?find=Merlin`
- `http://localhost:8080/?find=Merlin&bird=merlin` (redirected to `?find=Merlin&pin=merlin`)
- `http://localhost:8080/?pin=merlin&find=Blue+Jay`
- `http://localhost:8080/?pin=merlin&find=Blue+Jay&bird=blujay`
- `http://localhost:8080/?find=Blue+Jay&pin=merlin&pin=blujay&show=pinned` (screenshot)

## T3. New York in October

**Answer:** Blue Jay (1 in 24), American Robin (1 in 35), and American Crow (1 in 36). Next come
Black-capped Chickadee, White-throated Sparrow, and Canada Goose.

**Confidence:** 4

**Where:** The Month and Region menus would not open for me. Then I found a ready-made "New York in
October" link on the help page ("more ideas").

**Pages:**
- `http://localhost:8080/` with Month, Region, and "controls" opened (nothing opened)
- `http://localhost:8080/docs/using` (find "York")
- `http://localhost:8080/?region=NY&month=Oct` (text and screenshot)

## T4. Three ducks

**Answer:** **Mallard** has the widest wingspan: 34.6 in (~88 cm), against Canvasback 31.3 in
(~80 cm) and Wood Duck 28.0 in (~71 cm). **Canvasback** is the heaviest: 2.6 lb (~1.2 kg), against
Mallard 1.9 lb (~0.86 kg) and Wood Duck 1.4 lb (~0.64 kg).

**Confidence:** 5

**Where:** I searched for and pinned each duck, then opened "Pinned (3)".

**Pages:**
- `http://localhost:8080/?pin=mallar3&find=Wood+Duck`
- `http://localhost:8080/?pin=mallar3&find=Canvasback`
- `http://localhost:8080/?pin=mallar3&pin=wooduc&pin=canvas&show=pinned` (text and screenshot)

## T5. Solid and dashed

**Answer:** Solid lines are measured and dashed lines are estimated. On the wing bar, the solid ends
are the measured wing chord, and the middle is dashed when the wingspan is estimated. The body and
neck are estimated, so they are dashed.

**Confidence:** 5

**Where:** On the help page ("Solid lines are measured; dashed lines are estimated"), and the
Great Blue Heron's page repeats it under the drawing.

**Pages:** `http://localhost:8080/docs/using`, `http://localhost:8080/species/grbher3`

## T6. Sharp-shinned Hawk

**Answer:** 4.6 oz, about 130 g.

**Confidence:** 4

**Where:** I searched "Sharp-shinned", pinned it, and read "Mass" on its tile.

**Pages:**
- `http://localhost:8080/?find=Sharp-shinned`
- `http://localhost:8080/?pin=shshaw&show=pinned`

## T7. A heron's parts

**Answer:** Bill 5.7 in (~14.5 cm), tail 6.5 in (~16.5 cm). Both are marked "measured" (AVONET).

**Confidence:** 5

**Where:** The Great Blue Heron's own page, in the box at the top and in the "Measurements and
sources" table.

**Pages:** `http://localhost:8080/species/grbher3` (screenshot, 3 screens)

## T8. The wing chord

**Answer:** In the drawing, the wing chord is the solid parts at each end of the green wing bar,
ending at the small tick marks. It is 18.4 in (~47 cm), measured.

**Confidence:** 3. The text under the drawing explains it clearly. On my phone the drawing is small,
and I could only just see the ticks and where the solid part stops.

**Where:** Great Blue Heron page, in the explanation under the figure and in the table.

**Pages:** `http://localhost:8080/species/grbher3`

## E1. What confused or annoyed me

The units, most of all. I think in centimeters and grams, and the site shows inches, ounces, and
pounds. The help page says there is a Metric option, but I could not open the Units menu, and the
Month, Region, and Names menus did not open either. So I spent the whole visit doing arithmetic.
The weight unit also changes: Mallard is "29.8 oz" on the front page and "1.9 lb" when pinned, so
it was harder to compare. On the front page the drawings are very small on a phone. On the heron
page the sources column in the tables is cut off at the right edge of the screen. The "1 in 43
species observations" figures took me a moment to understand.

## E2. What I liked

Pinning birds and seeing them one under the other works well, and the figures really are to the
same scale. The heron page is honest: every number says whether it is measured or modeled and where
it comes from. The search box found every bird I typed. The "New York in October" example link
saved me when the menus did not work.

## E3. One change

Use metric for visitors outside the US, or put a big, easy imperial/metric switch at the top of
the page. Also allow IOC names, so that a European birder can type "Rough-legged Buzzard" or
"Little Auk".

## E4. Ease of use

**3 out of 5.** Search and pinning are easy, but I had to work around the menus and the units.
I would come back to compare sizes before a trip if I could get metric units to work.
