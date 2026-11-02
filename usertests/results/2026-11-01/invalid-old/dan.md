# Usability test: Dan, experienced field birder, laptop

## T1. First look

**Answer:** It draws North American birds as stick figures at true scale so you can compare sizes, and it says up front that some of the numbers are modeled. I'd go straight to comparing two birds I know, or try "the birds around you now."

**Confidence:** 4

**Where:** Front page, first screen: the intro line, the "New here? Try" links, five collapsed panels, a row of filter chips, a histogram, and the first hummingbird tiles.

**Pages:** `http://localhost:8081/`

## T2. Merlin and Blue Jay

**Answer:** The Merlin is bigger, and by weight it's a lot bigger.

| | Merlin | Blue Jay | Difference |
| --- | --- | --- | --- |
| Length | 11.5 in (estimated) | 10.7 in (estimated) | Merlin about 0.8 in longer |
| Wingspan | 23.8 in (measured) | 15.2 in (estimated) | Merlin about 8.6 in wider |
| Mass | 6.6 oz | 3.1 oz | Merlin about twice as heavy |

My friend is right only about length. In the air a Merlin is a much bigger bird.

**Confidence:** 4. The numbers are clear, but two of the four lengths have the "~" for estimated.

**Where:** The compare page, side by side at one scale. I couldn't find a search box anywhere, so I made the compare address myself from the eBird codes. One of the front-page compare links used `blujay`, and I know Merlin's code is `merlin`. The `find=` address had no effect on the front page: it still showed all 1041 species.

**Pages:**
- `http://localhost:8081/?find=Merlin` (did nothing)
- `http://localhost:8081/` with Controls opened (no search there either)
- `http://localhost:8081/compare?species=merlin&species=blujay` (1 and 2 screens)

## T3. New York in October

**Answer:** Blue Jay (1 in 24 observations), American Robin (1 in 35), and American Crow (1 in 36). Black-capped Chickadee, White-throated Sparrow, and Canada Goose come right after.

**Confidence:** 4. It looks right for October in New York, though I'd have expected the White-throat to rank a little higher.

**Where:** Regions and months panel, plus Order by Frequency under Controls. In that panel, clicking a state *drops* it, so to get New York alone you clear "All" and then click NY. Months work the same way. That's four clicks before you even sort, and it isn't what I expected.

**Pages:**
- `http://localhost:8081/` with Regions opened
- `http://localhost:8081/?region=NY&month=Oct&by=frequency`

## T4. Three ducks

**Answer:** The Mallard has the widest wingspan, at 34.6 in measured (Canvasback 31.3 in estimated, Wood Duck 28.0 in). The Canvasback is the heaviest, at 2.6 lb (Mallard 1.9 lb, Wood Duck 1.4 lb).

**Confidence:** 4

**Where:** The compare page. Again I typed the codes myself (`mallar3`, `wooduc`, `canvas`). Someone who doesn't know eBird codes would have to turn on "Compare birds" and hunt for each duck among 1041 tiles.

**Pages:** `http://localhost:8081/compare?species=mallar3&species=wooduc&species=canvas`

## T5. Solid and dashed

**Answer:** Solid lines are measured and dashed lines are estimated. The head is always estimated and drawn hollow. On the wing, the solid outer pieces are the measured wing chord, and the middle is dashed when the full span is estimated. In the number labels, a "~" means the same thing.

**Confidence:** 5. The note under every drawing says so plainly. The "~" isn't explained there; I worked it out.

**Where:** The legend under each compare page and species page.

**Pages:** Same as T2 and T4.

## T6. Sharp-shinned Hawk

**Answer:** 4.6 oz, marked as measured, from AVONET.

**Confidence:** 3. The number is sourced, which I like. But a Sharpie female is nearly twice the weight of a male, and the page gives one figure with no range and no word on sex. I'd want to know whether that's an average of both sexes.

**Where:** The species page's "Measurements and sources" table. Again I typed the code, `shshaw`.

**Pages:** `http://localhost:8081/species/shshaw`

## T7. A heron's parts

**Answer:** Bill 5.7 in, tail 6.5 in. Both are marked measured, from AVONET.

**Confidence:** 5

**Where:** The Great Blue Heron's own page. Both numbers are on the drawing labels and in the table.

**Pages:** `http://localhost:8081/species/grbher3`

## T8. The wing chord

**Answer:** The wing chord is the solid green piece at each outer end of the wing line, out past the small tick marks, labeled "Wing chord 18.4 in" at upper left. It's 18.4 in, measured, from AVONET.

**Confidence:** 4. Once I'd read the legend it was clear. The label sits on the left end only, though, and in small print.

**Where:** The drawing and table on the heron page.

**Pages:** `http://localhost:8081/species/grbher3`

## E1. What confused or annoyed me

No search box. With a thousand birds on the page that's the first thing I reach for. I only got anywhere because I know eBird's four- and six-letter codes and could type addresses by hand. Most people can't.

The region and month filters work backwards: a click removes a state instead of choosing it.

The labels on the drawings are tiny on a laptop. I could read "Wingspan 23.8 in" but I was squinting.

On the Merlin and Blue Jay page, both "Bill" and "Tail" were highlighted as the sort, which made no sense.

The modeled total lengths run short. The site says the Great Blue Heron is 34.7 in long, and every field guide I own says around 46. I appreciate that it says "modeled", but a skeptic like me looks at a number like that and starts doubting the rest. I'd rather see no total length than a wrong one.

The front page opens on hummingbirds and a wall of panels. It's busy before you've done anything.

## E2. What I liked

Every number on a species page says where it came from, whether it's measured or modeled, and for wingspans the reported range. That's exactly what I want, and I've never seen a bird site do it. Seeing the Merlin and the Blue Jay to one scale makes the point better than any table: the jay is nearly as long, but the Merlin's wings dwarf it. The October New York list matched what I'd expect, and the "1 in 24 observations" wording is honest about what it's counting.

## E3. One change

Put a search box at the top that finds a bird by common name and lets me add it to a comparison.

## E4. Ease of use

**3 out of 5.** It's easy once you're on a compare or species page and hard to get there. Yes, I'd come back for the sourcing and the side-by-side drawings, but I'd bookmark compare addresses rather than use the front page.
