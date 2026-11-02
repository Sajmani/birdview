# Usability test: Dan, experienced field birder, laptop

## T1. First look

**Answer:** A size-comparison tool. It draws each species as a to-scale stick figure, a top-down
flight silhouette with a cube for weight, so you can see how birds compare. The 1-in-N figures
suggest it also ranks birds by how often people record them. First I would search for a bird I
know and click into it to see where the numbers come from, since the header says "Click a bird for
its numbers and sources."

**Confidence:** 4

**Where:** The front page, first screen.

**Pages:** `http://localhost:8080/` (screenshot)

## T2. Merlin and Blue Jay

**Answer:** The Merlin is bigger, and by more than my friend thinks, except in length.

| | Merlin | Blue Jay | Difference |
| --- | --- | --- | --- |
| Length | 11.5 in | 10.7 in | Merlin about 0.8 in longer (both marked ~, estimated) |
| Wingspan | 23.8 in | 15.2 in (~) | Merlin about 8.6 in wider, roughly 1.5 times |
| Mass | 6.6 oz | 3.1 oz | Merlin about twice as heavy |

So "Blue Jay–sized" holds for length, but not for wings or weight. That matches what I see in the
field. The site also tells me the lengths are modeled, which is fair.

**Confidence:** 4. The page gives the numbers clearly. Merlin females run much heavier than males,
and the page shows one number per species without saying which sex.

**Where:** I typed "Merlin" in the find box, pinned it and the Blue Jay, then clicked Pinned.

**Pages:** `/?find=Merlin`, `/?pin=merlin&pin=blujay&show=pinned` (screenshot and text)

## T3. New York in October

**Answer:** Blue Jay (1 in 24), American Robin (1 in 35), and American Crow (1 in 36). After them
came Black-capped Chickadee, White-throated Sparrow, and Canada Goose. A Blue Jay at the top in
October fits the fall jay flight.

**Confidence:** 4

**Where:** Not from the controls. I tried to open the Region and Month dropdowns and nothing opened
for me. I found it under "more ideas" on the Using BirdView page, which has a ready-made "New York
in October" link. The heading then read "Showing birds in New York in October".

**Pages:** `/` (tried the Region and Month panels twice), `/docs/using#some-things-to-try`,
`/?region=NY&month=Oct` (screenshot)

## T4. Three ducks

**Answer:** The **Mallard** has the widest wingspan, 34.6 in, against 31.3 in (~, estimated) for
the Canvasback and 28.0 in for the Wood Duck. The **Canvasback** is the heaviest, 2.6 lb, against
1.9 lb for the Mallard and 1.4 lb for the Wood Duck.

**Confidence:** 4. The Canvasback's wingspan is estimated, but it would have to be off by more than
3 in to change the answer.

**Where:** I pinned all three with the find box and clicked Pinned. They were not side by side,
though. They came out stacked down the left side of the screen in a single column, one per
"Frequency Tier", with the right two-thirds of the screen empty. I had to scroll to compare them.

**Pages:** `/?pin=mallar3&find=Wood%20Duck`, `/?pin=mallar3&pin=wooduc&find=Canvasback`,
`/?pin=mallar3&pin=wooduc&pin=canvas&show=pinned` (two screenfuls and text)

## T5. Solid and dashed

**Answer:** Solid lines are measured and dashed lines are estimated. The bill, the tail, and the
wing-chord ends of the wing are measured. The neck and body are modeled, and the head is drawn
hollow because it is estimated too. On the wing, the middle is dashed only when the wingspan itself
is estimated. On the Canvasback I could see that dashed middle.

**Confidence:** 5

**Where:** The Using BirdView page has a table of each part. Each species page also has a sentence
under the drawing.

**Pages:** `/docs/using`; confirmed on `/species/shshaw` and `/species/grbher3`

## T6. Sharp-shinned Hawk

**Answer:** **4.6 oz**, marked "measured, from AVONET v7".

**Confidence:** 3. I trust where the number comes from. But a sharpie is one of the most dimorphic
birds we have: a female is close to twice a male's weight. One species-wide figure with no sex and
no range hides that. The wingspan gets a "Reported range"; the mass gets none.

**Where:** I found it with the find box, then went to the bird's own page.

**Pages:** `/?find=Sharp-shinned`, `/species/shshaw` (screenshot and text)

## T7. A heron's parts

**Answer:** The bill is **5.7 in** and the tail is **6.5 in**, both marked measured (AVONET v7).

**Confidence:** 5

**Where:** The top line of the Great Blue Heron's page and the "Measurements and sources" table.

**Pages:** `/species/grbher3` (screenshot and text)

## T8. The wing chord

**Answer:** The wing chord is the two outer stretches of the green wing crossbar, from each tick
mark out to the wingtip. It is **18.4 in**, measured, from AVONET. On the screen, tick to tip is
about a quarter of the full 71.9 in span, which agrees.

**Confidence:** 4. The text under the drawing explains it. In the heron drawing itself, though, the
whole bar is solid green, because its wingspan is measured too. The small ticks are the only thing
that marks off the chord, and I would not have known that without reading the text.

**Where:** The Great Blue Heron's page, the drawing and the caption under it.

**Pages:** `/species/grbher3`

## E1. What confused or annoyed me

The Region and Month dropdowns would not open for me. I only got to New York in October because the
help page had a canned link for it. If I had wanted Ohio in May, I would have been stuck.

"Pinned" is how you compare birds, but the pinned birds are not shown side by side. They are split
up by frequency tier and stacked in one narrow column. Comparing is the whole point of the site,
so that is backwards. And "Frequency Tier 2: 1 in 100 – 1 in 999" headings mean nothing to me when
I just want to see three ducks.

The units jump around. The heron's mass is 89.0 oz on the front page and 5.6 lb on its own page.
The Mallard is 29.8 oz on the front page and 1.9 lb when pinned. Pick one.

There is one mass per species, with no sex and no range, even for birds like the Sharp-shinned Hawk
and Merlin, where it matters a lot.

The front-page tiles are busy: banding code, eBird code, five measurements, a frequency figure, a
family chip, and a scientific name. The stick figures for small birds are tiny smudges at that size.

## E2. What I liked

Every number on a species page says where it came from and whether it is measured, modeled, or
derived, down to the AVONET sheet and the formula for the body. That is exactly what I want, and I
almost never get it. The wingspan's reported range is a nice touch. So is the honest note that the
neck/body split is a drawing convention. The folded S-neck on the heron shows somebody knows birds.
The frequency figure explains itself too: it is "how much of what people record there," not
abundance. The Merlin-versus-Blue-Jay comparison is the kind of thing I would actually use it for.

## E3. One change

Make Pinned a true side-by-side view: the chosen birds next to each other in one row, at one scale,
with no tier headings, so I can look across.

## E4. Ease of use

**3 out of 5.** Finding a single bird and reading its page is easy and rewarding. Filtering and
comparing take more work than they should. I would come back, mostly for the species pages and their
sourcing, and especially if the comparison view improves.
