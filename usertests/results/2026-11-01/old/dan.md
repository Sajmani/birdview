# Usability test: Dan (experienced field birder, laptop)

## T1. First look

**Answer:** It draws birds as to-scale stick figures so you can compare sizes, with numbers on
every card and "click a bird for its numbers and sources." First I'd type a bird I know into "Find a
bird to pin" and check its numbers against my Sibley.

**Confidence:** 4

**Where:** Front page, first screenful. The one-line intro and the "New here?" links told me what it
was. The drawings on the cards are tiny on a laptop. A robin is a little green cross, and I couldn't
tell what any of it meant until I read the help.

**Pages:** `http://localhost:8081/`

## T2. Merlin and Blue Jay

**Answer:** The Merlin is bigger, by a little in length and by a lot in everything else.

| | Merlin | Blue Jay | Difference |
| --- | --- | --- | --- |
| Length | 11.5 in (modeled, ~) | 10.7 in (modeled, ~) | Merlin about 0.8 in longer (~7%) |
| Wingspan | 23.8 in (measured) | 15.2 in (modeled, ~) | Merlin about 8.6 in wider (~55%) |
| Weight | 6.6 oz | 3.1 oz | Merlin about twice as heavy |

So my friend is right about length and wrong about everything else. In the air a Merlin is a much
bigger bird.

**Confidence:** 4. These roughly match my field guide. I'm holding back one point because both
lengths and the jay's wingspan carry the little "~" that means estimated.

**Where:** I searched "Merlin" in the find box and pinned it, then pinned Blue Jay. The Pinned view
showed the two cards together. Weight isn't in the text line at the top of the card; it's next to the
little cube under the drawing.

**Pages:** `http://localhost:8081/?find=Merlin`, `http://localhost:8081/?find=Merlin&bird=merlin`,
`http://localhost:8081/?pin=merlin&show=pinned`, `http://localhost:8081/?pin=merlin&pin=blujay&show=pinned`

## T3. New York in October

**Answer:** Blue Jay (1 in 24 species observations), American Robin (1 in 35), and American Crow
(1 in 36). After them come Black-capped Chickadee, White-throated Sparrow, and Canada Goose.

**Confidence:** 4 that this is what the site says. It's a believable October list for New York.

**Where:** This was the hard one. The page has Month and Region dropdowns, but I couldn't get them
to open. The "birds around you now" link wanted my location, and when I said no it told me to
"choose yours under Region," which I'd already tried. In the end I found a ready-made "New York in
October" link on the "more ideas" help page. That page heading reads "Showing birds in New York in
October."

**Pages:** `http://localhost:8081/` (Region and Month panels, which wouldn't open),
`http://localhost:8081/#here-and-now`, `http://localhost:8081/docs/using#some-things-to-try`,
`http://localhost:8081/?region=NY&month=Oct`

## T4. Three ducks

**Answer:** Mallard has the widest wingspan, 34.6 in (Canvasback 31.3 in, Wood Duck 28.0 in).
Canvasback is the heaviest, 2.6 lb (Mallard 1.9 lb, Wood Duck 1.4 lb).

**Confidence:** 5 for what the site shows. The Canvasback's wingspan is marked estimated (~ and a
dashed middle on the wing), and it's close enough to the Mallard's that I'd want a measured figure
before I'd bet on it.

**Where:** I found each duck with the find box and pinned all three. They weren't really side by
side: the Pinned view stacked them in one column, split up under "Frequency Tier" headings, so I had
to scroll up and down to compare. Sorting by frequency makes no sense when I've picked three birds to
compare.

**Pages:** `http://localhost:8081/?pin=mallar3&find=Wood%20Duck`,
`http://localhost:8081/?pin=mallar3&find=Canvasback`,
`http://localhost:8081/?pin=mallar3&pin=wooduc&pin=canvas&show=pinned`

## T5. Solid and dashed

**Answer:** Solid lines are measured and dashed lines are estimated. On the wing crossbar, the solid
ends are the measured wing chord, out to the ticks, and the middle is dashed when the wingspan is
estimated. The neck and body are always dashed because they're modeled. The hollow head is
estimated too.

**Confidence:** 5

**Where:** The "Using" help page, and the same explanation appears under the drawing on each bird's
own page. Nothing on the front page explains it.

**Pages:** `http://localhost:8081/docs/using`

## T6. Sharp-shinned Hawk

**Answer:** 4.6 oz, marked measured, from AVONET v7.

**Confidence:** 4. I'm sure that's the number. My reservation is that Sharpies vary a lot by sex,
with females far heavier than males, and the site gives one number with no range and no male/female
split. For a wingspan it does show a reported range. I'd like the same for mass.

**Where:** I searched "Sharp-shinned" and opened the bird's own page, where the Measurements and
sources table has the number and its source.

**Pages:** `http://localhost:8081/?find=Sharp-shinned`, `http://localhost:8081/species/shshaw`

## T7. A heron's parts

**Answer:** Bill 5.7 in, tail 6.5 in. Both are marked measured, from AVONET.

**Confidence:** 5

**Where:** Great Blue Heron's page, in the Measurements and sources table. They're also labeled on
the drawing in very small print.

One thing jumped out: total length is 34.7 in, "modeled." My Sibley says about 46 in for a Great
Blue Heron. The site does label it modeled and explains that the S-curved neck makes the figure
shorter, but the 34.7 is a total length, not a drawing length, and it's a foot short.

**Pages:** `http://localhost:8081/species/grbher3`

## T8. The wing chord

**Answer:** The wing chord is the solid green part at each end of the horizontal wing line, from
the wingtip in to the small tick mark. It's labeled "Wing chord 18.4 in" above the left end. It's
18.4 in, measured, from AVONET.

**Confidence:** 5

**Where:** Same heron page: the drawing, the explanation under it ("the solid ends are the measured
wing chord, and the ticks mark where it ends"), and the table.

**Pages:** `http://localhost:8081/species/grbher3`

## E1. What confused or annoyed me

The Region and Month dropdowns. I couldn't get them to open, and that's the main thing I'd use the
site for. I only got New York in October through a canned link buried in the help. On the front
page the drawings are so small you can't read anything from them, and weight is hidden down by a
little cube instead of sitting with the other numbers. "Compare" puts my three ducks in a single
column under frequency tier headings, so I'm scrolling instead of comparing. And the Great Blue
Heron's total length of 34.7 in is well short of what any field guide says. It's labeled modeled,
but a number that far off makes me doubt the other modeled lengths, like the Merlin's.

## E2. What I liked

The species page is first-rate. Every number says whether it's measured or modeled, where it came
from, and for wingspans the reported range. I can trace a figure to AVONET or the Fox wingspan
database, which almost nobody else lets you do. Solid for measured and dashed for estimated is a
good, honest convention once you know it. The frequency numbers ("1 in 24 species observations in
New York, Oct") are clearly explained as how much people record, not how many birds there are. I
appreciated that.

## E3. One change

Put pinned birds next to each other in one row, in the order I pinned them or sorted by size, with
weight on the same line as length and wingspan. That's the comparison the site promises, and right
now it makes you work for it.

## E4. Ease of use

**3 out of 5.** Looking up one bird and its sources was easy and worth the trip. Comparing birds and
filtering by place and month took more fiddling than it should have. I'd come back for the species
pages and the sourcing, and I'd keep my Sibley handy for lengths.
