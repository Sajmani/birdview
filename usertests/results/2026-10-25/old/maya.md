# Usability test: Maya (new birder, phone)

About 26 commands in all. Screenshots: `scratch/ut/old-maya/` (front.png, oct_tap.png, ducks.png).

## T1. First look

**Answer:** It's a site that draws birds as stick figures to scale so you can compare their sizes. First I'd tap "the birds around you now" or one of the "small / medium / large birds" links, because the five grey panels underneath (Presentation, Controls, Regions…) look like settings, not things to look at.

**Confidence:** 4

**Where:** Front page, first screen. That screen is all text, panels and filter chips plus a little bar chart. I didn't see a single bird until I scrolled. On my phone the page also looked wider than the screen, with a white strip down the right side.

**Pages:** `http://localhost:8081/`

## T2. Merlin and Blue Jay

**Answer:** The Merlin is **bigger**, a little longer but much bigger in wingspan and weight:

| | Merlin | Blue Jay | Difference |
|---|---|---|---|
| Length | 11.5 in | 10.7 in | Merlin ~0.8 in longer |
| Wingspan | 23.8 in | 15.2 in | Merlin ~8.6 in wider |
| Weight | 6.6 oz | 3.1 oz | Merlin ~3.5 oz heavier (about twice as heavy) |

Both lengths are marked "modeled", and so is the Blue Jay's wingspan.

**Confidence:** 4 for the numbers. I lost a point because it says the lengths are estimates.

**Where:** Each bird's own species page, under "Measurements and sources".

**How I got there:** I couldn't find a search box anywhere. Adding `q=merlin` to the address did nothing; the page just showed the same hummingbirds. Find-in-page found nothing until I scrolled through the whole list of 1,041 birds. Only then did I find "Merlin (merlin)" and tap it. The "medium birds" comparison has a Blue Jay, but it shows no numbers as text, so I opened the Blue Jay's own page.

**Pages:** `/?q=merlin`, `/` (find "Merlin", whole list), `/compare?species=blujay&…` (medium birds), `/species/merlin`, `/species/blujay`

## T3. New York in October

**Answer:** Blue Jay (1 in 24 observations), American Robin (1 in 35), American Crow (1 in 36). Black-capped Chickadee, White-throated Sparrow, Canada Goose, Song Sparrow and Northern Cardinal are right behind.

**Confidence:** 4

**Where:** Front page with Region = New York, Month = Oct, ordered by Frequency.

**How I got there:** The Regions and Months panel is huge, with a very long list of states, seas and atolls. When I tapped "Oct", the chip changed to "Month 11 of 12 selected". It had taken October *out* instead of picking it. To get October only, I'd have to tap the other eleven months off one by one. I changed the address to month=Oct instead. I also had to work out that "Frequency" under Controls means "most common first". "The birds around you now" didn't use my location or the month. It just sorted the whole continent by frequency.

**Pages:** `/ --open regions`, `/?region=NY` (regions panel), `/?month=Jan…Sep,Nov,Dec&region=NY` (after tapping Oct), `/?by=frequency&scale=global#here-and-now`, `/?month=Oct&region=NY&by=frequency`

## T4. Three ducks

**Answer:** **Mallard has the widest wingspan, Canvasback is the heaviest.**

**Confidence:** 3

**Where:** Compare page with the three ducks, using the Wingspan and Mass sort buttons.

**How I got there:** I opened "herons and ducks" and used find-in-page to get each duck's name and code. Then I put the three side by side on the compare page. On my phone the numbers on the drawings are far too small to read. I went by the sort buttons instead. Sorting by Wingspan put Mallard last, and sorting by Mass put Canvasback last. I'm assuming the list goes smallest to biggest, but nothing on the page says so, so I'm not sure. Looking at the drawings, Canvasback's cube does look the biggest. I couldn't tell the wing lines apart by eye.

**Pages:** `/?family=Ardeidae&family=Anatidae` (find Mallard / Wood Duck / Canvasback), `/compare?species=mallar3&species=wooduc&species=canvas` (screenshot and sort links)

## T5. Solid and dashed

**Answer:** Solid lines are measured, dashed lines are estimated. The head is always an estimate and is drawn hollow. On the wing, the solid ends are the measured "wing chord", and the middle is dashed if the full wingspan was estimated.

**Confidence:** 5 for solid = measured, dashed = estimated. 3 for the wing part, because I don't know what a wing chord is.

**Where:** The paragraph under the drawings on every compare page and species page.

**Pages:** `/compare?species=blujay&…`, `/species/merlin`

## T6. Sharp-shinned Hawk

**Answer:** **4.6 oz** (marked "measured").

**Confidence:** 4

**Where:** Sharp-shinned Hawk species page, Measurements and sources.

**How I got there:** I narrowed to the hawk family and used find-in-page to get its code. Then I opened its page.

**Pages:** `/?family=Accipitridae` (find "Sharp-shinned"), `/species/shshaw`

---

## E1. What confused or annoyed you?

There's no search box. That's the big one. I know the bird I want, and I had to scroll through a thousand birds or guess filters to find it. The first screen is all settings. "Length Tier 3", "Section step 15%", "Scale Per section" mean nothing to me. Tapping October and having it *remove* October was really confusing, and I'd have given up there on the subway. On the compare page the numbers on the drawings are tiny, unreadable on my phone. The species pages are full of things like "AVONET2_eBird" and "mass^0.37". The page also seemed slightly wider than my screen.

## E2. What did you like?

Once I was on a bird's page, the numbers were right there and it said which ones are measured and which are guesses. I liked the honesty. The New York in October list was exactly the kind of thing I want, and it matched what I see at my feeder. Comparing birds side by side at the same scale is a cool idea. The "small / medium / large birds" links were a nice way in.

## E3. One change

Put a search box at the very top, so I can type "Merlin" and go straight to it. Also add "Add a bird…" on the compare page.

## E4. Ease of use

**2 out of 5.** I got my answers, but only by working hard at it. I might come back for the side-by-side size comparisons if there were a search box. As it is, I'd probably stick with Merlin (the app).
