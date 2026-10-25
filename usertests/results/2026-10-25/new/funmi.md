# Usability test: Dr. Funmi Okafor (ornithologist, laptop)

## T1. First look

**Answer:** It's a size-comparison tool: North American birds drawn as to-scale stick figures, which you can filter by region, month, diet, habitat and family. The line under the title, "what the stick figures show, and what is estimated", caught my eye straight away. I'd open the Presentation panel first to see what they consider estimated, and then search for a bird I know well.

**Confidence:** 4

**Where:** The front page, first screenful: the title, the one-line description, the "New here? Try…" links, five collapsed panels, a search box, a histogram, then the first hummingbird tiles.

**Pages:** `http://localhost:8080/` (screenshot)

## T2. Merlin and Blue Jay

**Answer:** The Merlin is bigger, by a little in length and by a lot in everything else.

| | Merlin | Blue Jay | Difference |
|---|---|---|---|
| Total length | 11.5 in (modeled) | 10.7 in (modeled) | +0.8 in, ~7% longer |
| Wingspan | 23.8 in (**measured**, Fox et al. 2026; range 23.1–24.4) | 15.2 in (modeled from wing chord, regression on Passeriformes, ~4.7% median error) | +8.6 in, ~55% wider |
| Mass | 6.6 oz (measured, AVONET) | 3.1 oz (measured, AVONET) | about 2.1× heavier |

So your friend is right about length, but a Merlin weighs about twice as much and has a much longer wing. Both total lengths are BirdView's own sums of segments, not published lengths. The site says so clearly, so I'd trust the mass and the Merlin's wingspan more than the length comparison.

**Confidence:** 4 on the direction and the mass, 3 on the exact length gap (both lengths are modeled).

**Where:** I found the Merlin with the search box. I then built a compare view, which gave me the drawings with labels. The full numbers and sources were on each species page under "Measurements and sources".

**Pages:**
- `http://localhost:8080/?q=merlin`
- `http://localhost:8080/compare?species=merlin&species=blujay` (text + screenshot)
- `http://localhost:8080/species/blujay?species=merlin&species=blujay`
- `http://localhost:8080/species/merlin?species=merlin&species=blujay`

## T3. New York in October

**Answer:** The top three are **Blue Jay** (1 in 24 species observations), **American Robin** (1 in 35) and **American Crow** (1 in 36). Black-capped Chickadee is tied with the crow at 1 in 36, and White-throated Sparrow and Canada Goose come next.

This is for New York **State**. The site has no city-level option. It also measures the share of records, not abundance, and the site says that itself.

**Confidence:** 4

**Where:** Regions and months panel. I then sorted by frequency. To get there I combined region NY and month Oct with the frequency ordering used by the "birds around you now" link.

**Pages:**
- `http://localhost:8080/` with the Regions panel open (very long output)
- `http://localhost:8080/` with the Regions panel open, find "Oct"
- `http://localhost:8080/?region=NY&month=Oct&by=frequency&scale=global` (and a find-in-page for the frequencies)

## T4. Three ducks

**Answer:**
- **Widest wingspan:** Mallard, 34.6 in. This is measured. The Canvasback has 31.3 in, but that figure is modeled (dashed wing, "~"). The Wood Duck has 28.0 in.
- **Heaviest:** Canvasback, 2.6 lb (42.4 oz). The Mallard is 1.9 lb and the Wood Duck 1.4 lb (23.2 oz).

**Confidence:** 5 for the heaviest. For the wingspan I'm at 4, because the Canvasback's figure is modeled, although a 3-inch gap is bigger than the model's stated error.

**Where:** I searched for each duck. "wood duck" needed a `+` for the space, and I had to dig out its code, "wooduc". Then I used the compare view, and the screenshot showed all three on one scale.

**Pages:**
- `http://localhost:8080/?q=canvasback`
- `http://localhost:8080/?q=wood+duck` (twice, once to find the code)
- `http://localhost:8080/compare?species=mallar3&species=wooduc&species=canvas` (screenshot, 3 screens)

## T5. Solid and dashed

**Answer:** Solid lines are measured and dashed lines are estimated (modeled); the hollow head is estimated too. On the wing, the solid outer parts are the measured wing chord, with ticks where it ends. The dashed middle is the part of the span that has to be inferred when no measured wingspan exists, as with the Canvasback. Labels with a "~" are estimates. The cube is body mass drawn as a volume of water at the same scale. The neck/body split is described as a drawing convention, not a measurement.

**Confidence:** 5

**Where:** The legend under every drawing, on both the compare pages and the species pages.

**Pages:** As in T2 and T4.

## T6. Sharp-shinned Hawk

**Answer:** **4.6 oz (about 130 g)**, marked measured, from AVONET v7 (sheet AVONET2_eBird). I'd add a caveat. This is a single species mean, and Sharp-shins have very strong reversed sexual size dimorphism: females are much heavier than males. The page gives no range and no sex split for mass, although it does give a reported range for wingspan (21.0–23.7 in).

**Confidence:** 4 that this is what the site says and where it comes from. As "what a Sharp-shin weighs", it's a midpoint that no individual bird is likely to match.

**Where:** I searched for "sharp-shinned" and opened the species page, "Measurements and sources".

**Pages:**
- `http://localhost:8080/?q=sharp-shinned` (three find-in-page passes for the mass, the name and the code)
- `http://localhost:8080/species/shshaw?q=sharp-shinned`
- `http://localhost:8080/species/shshaw` with Controls open, find "Units": nothing there

## E1. What confused or annoyed you

- The compare page shows almost no numbers as text; they're only labels on the drawing. When I copied the compare page's text there were no measurements in it, so I had to open each species page to get numbers I could cite.
- Units jump around. The front page tiles give the Canvasback as 42.4 oz, the compare page says 2.6 lb, and there's no metric on the species page. I work in mm and g. There seems to be a units control on the main list, but I couldn't find one on the species page.
- In the month filter every month starts selected, so clicking "Oct" turns October *off*. To get "October only" you have to think backwards.
- Region means state. "New York" is the whole state; there's no NYC.
- You can only get to a bird's code by digging. To compare three named ducks I had to search, find the code (wooduc, mallar3, canvas), and then build the comparison. A "compare these" button or a multi-name search would help.
- On the Merlin/Blue Jay compare page, "Bill" and "Tail" were highlighted in the sort row and I couldn't tell what that meant.
- Mass is a single figure with no range, n or sex. For a dimorphic raptor that hides a lot.

## E2. What you liked

Honestly, the provenance. Every number says measured or modeled. The modeled ones give the formula, for example "trunk modeled as 14.0329 × mass^0.37", "cranium 7.5 × mass^(1/3), isometric, unvalidated", and "log-log regression on 360 Passeriformes, cross-validated median error 4.7%". The footer admits "Every length here is modeled". I've never seen a bird-size site be that upfront that total length is a construct. The dashed lines carry the same information visually, so you can see at a glance which parts are inferred. DOIs and licenses for AVONET, the wingspan database, AviList and GBIF are right there, so I could cite them. Using wing chord as the measured core of the wing is the right call.

## E3. One change

Give every species page and every compare page a plain table of the numbers in metric (mm, g), with the measured/modeled flag, source and, where the source has it, range, n and sex split. And add a "download/cite this" option. The drawings are lovely, but I need the numbers as text I can copy.

## E4. Ease of use

**3.5 out of 5.** Searching and the species pages were easy. Comparing and filtering took some fiddling: codes, the inverted month toggles and units that changed from page to page. Yes, I'd come back. It's the only place I've seen that is this honest about what is measured and what is modeled. I'd use it for teaching, and to check AVONET values quickly with their provenance.
