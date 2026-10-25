# Usability test — Lena (birder from Leipzig, phone)

I ran 25 commands in total.

## T1. First look

**Answer:** The site draws birds as stick figures to scale so you can compare their sizes. There are also panels to narrow the list by region, month, diet and family. First I would tap "the birds around you now", or look for a search box to type a bird name. There is no search box.

**Confidence:** 4

**Where:** The top of the front page. The intro text, the "New here? Try…" links, five collapsed panels, then a row of filter chips and a small bar chart. On my phone the content only fills part of the screen width.

**Pages:** `http://localhost:8081/` (screenshot)

## T2. Merlin and Blue Jay

**Answer:** The Merlin is **bigger**, a little in length and much more in wingspan and weight. My friend is only right about the length.

| | Merlin | Blue Jay | Difference |
|---|---|---|---|
| Length | 11.5 in (≈29 cm, modelled) | 10.7 in / 27.1 cm (modelled) | Merlin ≈2 cm longer (+7%) |
| Wingspan | 23.8 in (≈60.5 cm, measured) | 15.2 in / 38.6 cm (modelled) | Merlin ≈22 cm wider (+57%) |
| Weight | 6.6 oz (≈187 g, measured) | 3.1 oz / 88 g (measured) | Merlin about twice as heavy (≈+100 g) |

**Confidence:** 4. The numbers are clear on the species pages. I converted the Merlin's numbers to metric myself because I had not switched units yet. Both lengths are "modelled", which surprised me.

**Where:** First I tried typing "Merlin" into the address as a search (`?q=Merlin`). Nothing happened: the page still showed all 1041 birds. Then I used find-in-page on the full list. It showed "Merlin (merlin)", so I guessed the code and built a compare address like the "medium birds" link. On the compare page the numbers in the drawings are far too small to read on a phone, so I opened each "Species details" page.

**Pages:**
- `http://localhost:8081/?q=Merlin`
- `http://localhost:8081/` (find)
- `http://localhost:8081/compare?species=merlin&species=blujay` (text and screenshot)
- `http://localhost:8081/species/merlin?species=merlin&species=blujay`
- `http://localhost:8081/species/blujay?species=merlin&species=blujay`

## T3. New York in October

**Answer:** The birds people record most often are **Blue Jay** (1 in 24 records), **American Robin** (1 in 35) and **American Crow** (1 in 36). Next come Black-capped Chickadee, White-throated Sparrow and Canada Goose.

**Confidence:** 4

**Where:** The "Regions and months" panel, with Order by → Frequency. All months and all regions are selected at the start. Tapping "Oct" switches October *off*, so getting only October means tapping the other eleven months. Regions work the same way, with about 70 of them. I went straight to the address those taps would give: month Oct, region NY, ordered by frequency. The chips then said "Month Oct, Region New York" and the list was sorted by "1 in N species observations in New York, Oct".

**Pages:**
- `http://localhost:8081/?units=metric` (Controls, then Regions panel)
- `http://localhost:8081/?month=Oct&region=NY&by=frequency&units=metric`

## T4. Three ducks

**Answer:** The **Mallard has the widest wingspan** (87.9 cm). The **Canvasback is the heaviest** (1.20 kg).

| | Length | Wingspan | Weight |
|---|---|---|---|
| Mallard | 50.5 cm | 87.9 cm (measured) | 843 g |
| Wood Duck | 47.4 cm | 71.1 cm (measured) | 658 g |
| Canvasback | 52.0 cm | 79.4 cm (modelled) | 1.20 kg |

**Confidence:** 4. The Canvasback wingspan is modelled, but it is 8 cm short of the Mallard's, so I believe the order.

**Where:** I found the codes by filtering to ducks (family Anatidae) and using find-in-page. Then I built a compare page and read the order from the Sort by Wingspan and Sort by Mass buttons. The drawings' labels are unreadable on a phone, so I took the numbers from each species page.

**Pages:**
- `http://localhost:8081/?family=Anatidae&units=metric` (find Mallard, Wood, Canvasback)
- `http://localhost:8081/compare?species=mallar3&species=wooduc&species=canvas&units=metric`
- `http://localhost:8081/?family=Anatidae&units=metric&by=mass` (not useful: numbers without names)
- `http://localhost:8081/species/mallar3?...&units=metric`
- `http://localhost:8081/species/wooduc?...&units=metric`
- `http://localhost:8081/species/canvas?...&units=metric` (twice)

## T5. Solid and dashed

**Answer:** **Solid lines are measured, dashed lines are estimated** (modelled). The head is always estimated and drawn hollow. On the wing, the solid outer ends show the measured wing chord, with ticks where it ends. The span between them is dashed when the wingspan is estimated.

**Confidence:** 5

**Where:** The explanation printed under the drawings on every compare and species page.

**Pages:** `http://localhost:8081/compare?species=merlin&species=blujay` and the species pages.

## T6. Sharp-shinned Hawk

**Answer:** **131 g** (measured, AVONET).

**Confidence:** 3. The number is clear, but it is a single value. In this species the female is much heavier than the male, and the page gives no range and says nothing about the sexes. So I am not sure what 131 g stands for.

**Where:** Find-in-page on the full list gave "Sharp-shinned Hawk (shshaw)". Then I opened its species page with metric units.

**Pages:**
- `http://localhost:8081/?units=metric` (find "Sharp-shinned")
- `http://localhost:8081/species/shshaw?units=metric`

---

## E1. What confused or annoyed me

- **No search.** I know which bird I want. With 1041 species I could only find it with my browser's find-in-page, after scrolling the whole list to the bottom. Typing a name did nothing.
- **Tiny numbers on the compare page.** The comparison is the main feature, but on my phone the labels on the drawings are a few pixels high. I had to open every bird on its own page to read its size, so the comparison was no help.
- **Month and region filters are backwards for me.** Everything starts switched on, and tapping a month switches it off. To see "October only" I would need eleven taps, and for "New York only" about seventy. I want an "only this one" tap.
- **Inches and ounces by default.** I found cm / g under Controls, but it took a while. The Merlin and Blue Jay pages I opened first were still in inches.
- Lengths are always "modelled", even for birds like the Mallard. That surprised me a little. The site is honest about it, though.
- The list headings like "Length Tier 3 3.96 in – 4.53 in" and the "Section step 15%" chip don't mean anything to me.

## E2. What I liked

- Every number says if it is measured or modelled, and where it comes from. As a scientist I like that very much.
- The "1 in N species observations in New York, Oct" list is exactly the kind of answer I wanted for "what will I see here now".
- Metric units work well once they are on (cm, g, kg).
- The solid/dashed explanation is short and clear.
- There is a "World" names option. I didn't try it, but I would want IOC/AviList names like "Little Auk".

## E3. One change

A **search box at the top** that finds a bird by its English name (North American or world names), scientific name or code, with an "add to compare" button. And please make the compare page show the length, wingspan and weight as readable text under each bird on a phone.

## E4. How easy was it?

**2 out of 5.** The data is very good, but on a phone I had to work around the site, guessing codes and building addresses, to get answers. I would come back on a laptop to compare sizes, and maybe for the region and month list. On my phone, not unless there is search and the numbers are readable.
