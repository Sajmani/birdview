# Usability test: Maya, new birder, phone

## T1. First look

**Answer:** It shows birds as little stick drawings at true scale so you can compare how big they are. First I'd tap "the birds around you now", or just type a bird I know into "Find a bird to pin".

**Confidence:** 4

**Where:** The front page, first screen. The intro sentence and the "New here?" line made sense. The row of gray dropdowns (Order by, Month, Region, Family, Habitat, Diet, Units, Names) and the bar chart under it didn't mean anything to me yet. The American Robin card at the bottom had a tiny drawing: a green cross with a small head. I couldn't tell what it was until I read more.

**Visited:** `http://localhost:8080/` (screenshot)

## T2. Merlin and Blue Jay

**Answer:** The Merlin is bigger, mostly in wingspan and weight. Length is nearly the same.

| | Merlin | Blue Jay | Difference |
| --- | --- | --- | --- |
| Length | 11.5 in | 10.7 in | Merlin about 0.8 in longer |
| Wingspan | 23.8 in | 15.2 in | Merlin about 8.6 in wider (roughly 1.5×) |
| Weight | 6.6 oz | 3.1 oz | Merlin about twice as heavy |

So my friend is right about length, but a Merlin is a much heavier bird with much longer wings.

**Confidence:** 4

**Where:** I searched "Merlin" in the find box and pinned it, then searched "Blue Jay" and pinned it, then tapped "Pinned (2)". The two cards are stacked one above the other, so I had to scroll to compare them. The Merlin's green wing line is clearly longer and its "mass" box is bigger.

**Visited:**
- `http://localhost:8080/?find=Merlin`
- `http://localhost:8080/?find=Merlin&pin=merlin`
- `http://localhost:8080/?find=Blue+Jay&pin=merlin`
- `http://localhost:8080/?find=Blue+Jay&pin=merlin&pin=blujay`
- `http://localhost:8080/?pin=merlin&pin=blujay&show=pinned` (text and screenshot)

## T3. New York in October

**Answer:** Blue Jay (1 in 24 observations), American Robin (1 in 35), and American Crow (1 in 36). Black-capped Chickadee, White-throated Sparrow, and Canada Goose come right after.

**Confidence:** 4

**Where:** "Showing birds in New York in October". I couldn't get there from the Month and Region dropdowns: tapping them did nothing for me, and they stayed shut. "The birds around you now" needs my location, and it's November for me, not October. I eventually found a "New York in October" link in the "more ideas" help page, and that worked.

**Visited:**
- `http://localhost:8080/` (tried to open Month, Region, and "controls": nothing opened)
- `http://localhost:8080/#here-and-now`
- `http://localhost:8080/docs/using#some-things-to-try`
- `http://localhost:8080/?region=NY&month=Oct`

## T4. Three ducks

**Answer:** The **Mallard** has the widest wingspan (34.6 in, vs. Canvasback 31.3 in and Wood Duck 28.0 in). The **Canvasback** is the heaviest (2.6 lb, vs. Mallard 1.9 lb and Wood Duck 1.4 lb).

**Confidence:** 5

**Where:** I pinned all three and used the Pinned view. They aren't really side by side on my phone. They're stacked, so I scrolled up and down, but the numbers are easy to read. One odd thing: on the main list the Mallard's mass said "29.8 oz", and in the pinned view it said "1.9 lb". I had to work out that those are the same.

**Visited:**
- `http://localhost:8080/?find=Wood+Duck&pin=mallar3`
- `http://localhost:8080/?find=Canvasback&pin=mallar3`
- `http://localhost:8080/?pin=mallar3&pin=wooduc&pin=canvas&show=pinned` (text and screenshot)

## T5. Solid and dashed

**Answer:** Solid lines are measured, and dashed lines are estimated. On the wing, the solid ends are the measured wing chord. The middle of the wing is dashed when the wingspan is estimated, and the neck and body are estimated, so they're dashed too.

**Confidence:** 4

**Where:** The "stick figures" help page linked from the front page intro. It's also written right under the drawing on a bird's own page, which I only saw later on the heron. The list cards don't explain it, and the dashes are very small on a phone.

**Visited:** `http://localhost:8080/docs/using` (searched "dash")

## T6. Sharp-shinned Hawk

**Answer:** 4.6 oz.

**Confidence:** 3. The number was clear. I know female sharpies are much bigger than males, though, and the card gives just one number without saying whether it's an average.

**Where:** The search results only show the name, not the numbers, so I pinned it and looked at the Pinned view card.

**Visited:**
- `http://localhost:8080/?find=Sharp-shinned`
- `http://localhost:8080/?pin=shshaw&show=pinned`

## T7. A heron's parts

**Answer:** The bill is 5.7 in and the tail is 6.5 in. Both are marked "measured".

**Confidence:** 5

**Where:** The Great Blue Heron's own page, in the card at the top and again in the "Measurements and sources" table.

**Visited:** `http://localhost:8080/species/grbher3` (text and screenshot)

## T8. The wing chord

**Answer:** The wing chord is 18.4 in. In the drawing it's the solid green pieces at the two outer ends of the wing crossbar, out to the little tick marks. The part between the ticks is the rest of the wingspan.

**Confidence:** 3. The sentence under the drawing says so, and I can see the ticks. Still, I don't really know what a "wing chord" is. I'm guessing it's one wing measured from the bend to the tip. On the heron the whole wing line looks solid to me, so I can only find the chord by the ticks.

**Where:** Great Blue Heron page: the note under the drawing ("the solid ends are the measured wing chord, and the ticks mark where it ends") and the "Wing chord 18.4 in" row.

**Visited:** `http://localhost:8080/species/grbher3`

## E1. What confused or annoyed you

The Month and Region dropdowns wouldn't open when I tapped them. That's the main way to ask "what's around here in October", and I only got there by digging through a help page. A lot of the front page is jargon to me: "Frequency Tier 1", "1 in 43 species observations", "Chord", the four-letter codes like AMRO, and the Latin names. The drawings on the list are tiny on a phone, and I can't make out solid vs. dashed without zooming. "Compare side by side" really means scrolling up and down a stack of cards. Weight flipped between ounces and pounds for the same duck. The heron page's sources table is a wall of things like "AVONET2_eBird" and "7.5 x mass^(1/3)" that I scrolled right past. There was also a white strip down the right edge of every screenshot, like the page doesn't quite fit my screen.

## E2. What you liked

Pinning birds is simple, and once I had two birds pinned the answer was right there in plain numbers. The Merlin vs. Blue Jay thing surprised me, in a good way. The "New York in October" list is exactly what I'd want before a walk in Prospect Park. The bird's own page explains the drawing in one short paragraph and says what's measured and what's estimated, which I trust.

## E3. One change

Make the pinned birds show up next to each other, small enough to see on one phone screen, with a one-line key under them ("solid = measured, dashed = estimated, ticks = end of one wing"). And make Month and Region open when you tap them.

## E4. Ease of use

**3 out of 5.** Looking up and comparing birds I already know by name was easy. Anything involving the filters, or understanding the drawings, made me work. I'd come back to settle "which is bigger" questions with friends, but probably not to browse.
