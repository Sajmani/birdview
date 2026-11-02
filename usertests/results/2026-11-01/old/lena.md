# Usability test: Lena (birder from Germany, phone), old site

## T1. First look

**Answer:** The site draws birds as stick figures to scale so you can compare their sizes and shapes. First I would try "the birds around you now" or the "compare some small/medium/large birds" links. I also noticed right away that units are "imperial" and names are "North Amer[ican]", and I would want to change both.

**Confidence:** 4

**Where:** Front page, first screen.

**Pages:** `http://localhost:8081/`

## T2. Merlin and Blue Jay

**Answer:** The Merlin is bigger, and much heavier:

| | Merlin | Blue Jay | Difference |
| --- | --- | --- | --- |
| Length | 11.5 in (about 29 cm) | 10.7 in (about 27 cm) | Merlin about 2 cm longer, so nearly the same |
| Wingspan | 23.8 in (about 60 cm) | 15.2 in (about 39 cm) | Merlin about 22 cm wider, about 1.5 times |
| Mass | 6.6 oz (about 187 g) | 3.1 oz (about 88 g) | Merlin about twice as heavy |

So my friend is right about the length, but not about the wingspan or the weight. I converted from inches and ounces myself. I could not open the Units menu to switch to metric.

**Confidence:** 4 (the numbers are clear; my conversions are approximate)

**Where:** I searched "Merlin" in the find box and pinned it, pinned the Blue Jay, then clicked Pinned. The mass is only in small text next to the cube in the drawing, not in the number lines at the top of the tile.

**Pages:**
- `http://localhost:8081/?find=Merlin`
- `http://localhost:8081/?find=Merlin&bird=merlin` (became `?find=Merlin&pin=merlin`)
- `http://localhost:8081/?pin=merlin&pin=blujay&show=pinned`

## T3. New York in October

**Answer:** Blue Jay (1 in 24 observations), American Robin (1 in 35), and American Crow (1 in 36). The next is Black-capped Chickadee (1 in 36).

**Confidence:** 4

**Where:** I could not open the Region or Month menus on the front page. I found the "New York in October" example link on the help page ("more ideas"), and it gave me exactly this view.

**Pages:**
- `http://localhost:8081/` (tried the Region menu)
- `http://localhost:8081/docs/using#some-things-to-try`
- `http://localhost:8081/?region=NY&month=Oct`

## T4. Three ducks

**Answer:**
- Widest wingspan: **Mallard**, 34.6 in (about 88 cm). Canvasback 31.3 in, Wood Duck 28.0 in.
- Heaviest: **Canvasback**, 2.6 lb (about 1.2 kg). Mallard 1.9 lb, Wood Duck 1.4 lb.

**Confidence:** 4

**Where:** I searched for each duck and pinned it, then clicked Pinned. On the phone they are not side by side but one under the other, each under its own "Frequency Tier" heading, so I had to scroll up and down to compare.

**Pages:**
- `http://localhost:8081/?find=Canvasback`
- `http://localhost:8081/?find=Wood%20Duck`
- `http://localhost:8081/?pin=mallar3&pin=wooduc&pin=canvas&show=pinned`

## T5. Solid and dashed

**Answer:** Solid lines are measured and dashed lines are estimated. The bill and tail are measured, so they are solid. The neck and body are estimated, so they are dashed, and the head is estimated and drawn hollow. On the wing bar the solid ends are the measured wing chord, and the middle is dashed when the wingspan is estimated. A "~" after a number also means it is estimated.

**Confidence:** 5

**Where:** The help page "Using BirdView", section "Reading a tile". It is also explained in one sentence on the bird's own page. The front page itself does not explain it.

**Pages:** `http://localhost:8081/docs/using#some-things-to-try`

## T6. Sharp-shinned Hawk

**Answer:** 4.6 oz, about 130 g. The site gives one number for the species. In Europe I know the Sparrowhawk females are much heavier than the males, so I expect this one value is an average.

**Confidence:** 4

**Where:** I searched "Sharp-shinned", pinned the bird, and read the mass next to the cube in its tile. The text tool found no "Mass" on the page, so I had to look at the screenshot.

**Pages:**
- `http://localhost:8081/?find=Sharp-shinned`
- `http://localhost:8081/?pin=shshaw&show=pinned`

## T7. A heron's parts

**Answer:** Bill 5.7 in (about 14.5 cm), tail 6.5 in (about 16.5 cm). Both are marked "measured".

**Confidence:** 5

**Where:** On the Great Blue Heron's own page, in the "Measurements and sources" table. I got there from the heron's tile on the front page.

**Pages:** `http://localhost:8081/species/grbher3`

## T8. The wing chord

**Answer:** The wing chord is the solid part at each end of the green wing bar, ending at the small tick marks. The middle of the bar is the rest of the wingspan. It is 18.4 in (about 47 cm) long, and it is measured.

**Confidence:** 4. The explanation text is clear. On my phone the drawing is very small, and I could only guess that the tiny green label on the left of the wing says "Wing chord 18.4 in". I took the number from the table.

**Where:** The sentence under the drawing on the heron's page, and the table row "Wing chord".

**Pages:** `http://localhost:8081/species/grbher3`

## E1. What confused or annoyed me

The worst part was that I could not change the settings. I wanted metric units and world names from the start, because I think in centimeters and grams and say "Rough-legged Buzzard", but the Units and Names menus would not open for me. Region and Month would not open either. So I converted inches, ounces, and pounds in my head the whole time, and for New York I had to find an example link on the help page. Smaller things:
- The weight is not in the number lines at the top of a tile. It is only tiny text next to the cube.
- On the phone the drawings in the tiles are very small: a little cross in a big empty box.
- On the heron's page the drawing is so small that I cannot read any of the labels, and the bird's name is cut off ("Great Blue Her…").
- When I compare pinned birds they are split up under "Frequency Tier" headings, which I do not care about when I compare sizes. "1 in 848 species observations" took me a moment to understand.
- On some pages there was an empty strip on the right side of the screen, as if the page did not fit my phone.

## E2. What I liked

The idea is very good: real sizes side by side, and it is honest about what is measured and what is estimated. The species page with the source for every number is great for someone like me who works with data. The help page is clear and complete. It even says AviList names like Little Auk are an option, so somebody has thought about birders like me. Pinning birds from search and then pressing Pinned to compare them works well. The New York in October list was exactly what I wanted to know.

## E3. One change I would make

Make the settings work with a tap on the phone, especially Units, and remember my choice. Or better, offer metric automatically when the phone is set to German. Every number on the site would then make sense to me without doing arithmetic.

## E4. Ease of use

**3 out of 5.** Once I understood search, pin, and Pinned, finding and comparing birds was easy. But not being able to change units, names, region, or month cost me a lot, and the drawings are too small on a phone. I would come back, probably on a laptop, to check sizes of American birds against the European birds I know, if I can switch to metric and world names.
