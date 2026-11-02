# Usability test: Maya (new birder, phone)

## T1. First look

**Answer:** It draws birds as stick figures to scale so you can compare their sizes. First I'd tap
"the birds around you now", or "medium birds", because those are the only plain-English things on
the screen. The five gray boxes (Presentation, Controls, ...) and the row of little "Order by /
Scale / Section step 15%" chips look like settings, not something to read. The page is also a bit
wider than my screen: there's a white strip down the right side, and the bird cards are cut off on
the right.

**Confidence:** 4

**Where:** the front page, first screen

**Pages:** `http://localhost:8081/` (one screenful, then three to see the cards)

## T2. Merlin and Blue Jay

**Answer:** The Merlin is bigger, mostly in wings and weight:
- Length: Merlin 11.5 in vs. Blue Jay 10.7 in, so less than an inch longer. About the same, my friend's right.
- Wingspan: Merlin 23.8 in vs. Blue Jay 15.2 in, so about 8.6 in wider (roughly 1.5 times).
- Weight: Merlin 6.6 oz vs. Blue Jay 3.1 oz, so about twice as heavy.

**Confidence:** 4 for the numbers. Finding them was the problem.

**Where:** the bird list filtered to two families (Falcons and Caracaras plus Crows, Jays, and
Magpies) from the Families panel. I couldn't find a search box. Adding `find=merlin` to the address
did nothing (still "1041 of 1041 species"). Find-in-page didn't find Merlin because the list only
loads the first ~25 birds (hummingbirds). The "medium birds" comparison has the Blue Jay and a
Kestrel but no Merlin. I only found it because I already knew a Merlin is a falcon.

**Pages:**
- `http://localhost:8081/?find=merlin`
- `http://localhost:8081/` (find in page "Merlin": 0 results)
- `http://localhost:8081/compare?species=blujay&species=amekes&species=calthr&species=moudov&species=belkin1&species=chwwid`
- `http://localhost:8081/` (Controls panel, Families panel)
- `http://localhost:8081/?family=Falconidae&family=Corvidae`

## T3. New York in October

**Answer:** Blue Jay (1 in 24 records), American Robin (1 in 35), American Crow (1 in 36). Then
Black-capped Chickadee, White-throated Sparrow, and Canada Goose.

**Confidence:** 4

**Where:** Regions and months panel, plus Order by Frequency. This was confusing. Every month and
every state starts out selected, so tapping "Oct" or "NY" **removes** it instead of choosing it.
I had to read the small print to learn you tap "All" first to clear, then tap the one you want, for
both month and region, and then change the order to Frequency. That's five or six taps and a
paragraph of reading. I ended up at the address below.

**Pages:**
- `http://localhost:8081/` (Regions and months panel opened, read)
- `http://localhost:8081/?month=Oct&region=NY&by=frequency`

## T4. Three ducks

**Answer:** The **Mallard** has the widest wingspan (34.6 in, vs. Canvasback 31.3 in and Wood Duck
28.0 in). The **Canvasback** is the heaviest (2.6 lb, vs. Mallard 1.9 lb and Wood Duck 1.4 lb).

**Confidence:** 4 (from the list numbers). From the side-by-side drawing alone I'd say 2: the
cubes and wing lines look close, and the labels on the drawings are too small to read on my phone.
The side-by-side page doesn't print the numbers anywhere I could read them.

**Where:** numbers from the "herons and ducks" style family list (Ducks, Swans, and Geese). The
picture is from the Compare page.

**Pages:**
- `http://localhost:8081/?family=Anatidae`
- `http://localhost:8081/compare?species=mallar3&species=wooduc&species=canvas`

## T5. Solid and dashed

**Answer:** Solid lines are measured; dashed lines are estimated (guessed by the site). The hollow
head is estimated too.

**Confidence:** 5. It says so in bold under the drawings.

**Where:** the paragraph under the drawings on the Compare page (also on the bird pages)

**Pages:** `http://localhost:8081/compare?species=mallar3&species=wooduc&species=canvas`

## T6. Sharp-shinned Hawk

**Answer:** 4.6 oz (about the same length as a Merlin, 11.5 in, but lighter)

**Confidence:** 4

**Where:** the hawks family list (Kites, Old World Vultures, Eagles, and Hawks). Same problem as
the Merlin: no search, so I had to know which family to look in.

**Pages:** `http://localhost:8081/?family=Accipitridae`

## T7. A heron's parts

**Answer:** Bill 5.7 in, tail 6.5 in (both marked "measured").

**Confidence:** 5 (from the table). The labels on the drawing itself are too small to read.

**Where:** "Measurements and sources" table on the Great Blue Heron's page

**Pages:**
- `http://localhost:8081/?family=Ardeidae&family=Anatidae` (to find the heron)
- `http://localhost:8081/species/grbher3`

## T8. The wing chord

**Answer:** 18.4 in. The text says the wing chord is "the solid outer ends" of the wing line, with
little ticks where it stops, so I think it's each end of the green wing line, from the wingtip in to
the tick. I couldn't actually see that on the drawing: on my phone the wing looks like one thin
green line, and the "Wing chord" label is microscopic. I still don't really know what a wing chord
is. Nothing tells me in plain words.

**Confidence:** 4 for 18.4 in, 2 for which part of the drawing it is

**Where:** Great Blue Heron page: the table row "Wing chord 18.4 in measured" and the paragraph under
the drawing

**Pages:** `http://localhost:8081/species/grbher3`

## E1. What confused or annoyed you

No search box. I know the bird's name; I don't always know its family. Typing it in the address did
nothing, and find-in-page fails because the list only loads hummingbirds first. The month and state
buttons work backwards: tapping one turns it off. The drawings are tiny on my phone, and all the
labels on them are too small to read, so the stick figures, which are the whole point, don't help
me much. The numbers in the list did the work. There's a lot of jargon and fine print ("Section step
15%", "per section", "allometric", "AVONET2_eBird", "1 in 11,147,318 species observations"). The page
is a bit wider than my screen and cards get cut off. A Canada Goose's weight shows as "99.2 oz"
instead of pounds.

## E2. What you liked

Once I found the birds, the numbers were right there on each card: length, wingspan, weight. The
New York in October list was great when I got there. Blue Jay, Robin, Crow, Chickadee: yes, that's
my feeder. I liked that it's honest about what's measured and what's guessed, and the "measured" /
"modeled" tags on the heron page were clear. The ready-made "small / medium / large birds" links
are a nice idea.

## E3. One change

A search box at the top: type "Merlin", get the Merlin, with a button to add it to a comparison.

## E4. Ease of use

**2 out of 5.** I'd come back for the "what's around me this month" list and maybe to settle a
size argument, but only if I could search by name. On the subway I'd have given up on the Merlin
question.
