# Usability test: Maya, new birder, phone

24 commands in all.

## T1. First look

**Answer:** It's a site that draws birds as stick figures to scale so you can compare their sizes. First I'd tap "the birds around you now", or just type a bird I know into the search box.

**Confidence:** 4

**Where:** Front page, first screenful. The one-line intro and the "New here? Try…" links are clear. Below them are five collapsed panels and a block of chips ("Section step 15%"?), which mean nothing to me yet.

**Visited:** `http://localhost:8080/` (screenshot)

## T2. Merlin and Blue Jay

**Answer:** A Merlin is **bigger** than a Blue Jay, though about the same length:
- Length: Merlin 11.5 in vs Blue Jay 10.7 in, so about 0.8 in longer.
- Wingspan: 23.8 in vs 15.2 in, so about 8.6 in wider (roughly half again as wide).
- Weight: 6.6 oz vs 3.1 oz, so a little over twice as heavy.

**Confidence:** 4. The numbers are right there on the search results. The Merlin page says its total length is "modeled / estimated", though, so I'm slightly unsure about length.

**Where:** I searched each name and read the Length / Wingspan / Mass line on the bird card. I also opened the Merlin page.

**Visited:**
- `http://localhost:8080/?q=merlin`
- `http://localhost:8080/?q=blue+jay`
- `http://localhost:8080/species/merlin?q=merlin`

## T3. New York in October

**Answer:** Blue Jay (1 in 24 observations), American Robin (1 in 35), American Crow (1 in 36). Right behind them: Black-capped Chickadee, White-throated Sparrow and Canada Goose.

**Confidence:** 4 that this is what the site says. It took me a while to get there.

**Where:** "Regions and months" panel, then "Order by: Frequency" in Controls.
- The month and region pickers start with **everything selected**, so tapping "Oct" or "NY" *removes* it. You have to tap "All" first to clear, and only then tap what you want. I nearly got this backwards.
- In the end I tapped Month "Now" (it's October), then Region "All" to clear, then "NY New York", then Order by "Frequency".
- The region list is huge on a phone.

**Visited:**
- `http://localhost:8080/` (Regions panel opened, several times while hunting for the month and region controls)
- `http://localhost:8080/?month=`
- `http://localhost:8080/now` (redirected to `?month=Oct`)
- `http://localhost:8080/?month=Oct&region=`
- `http://localhost:8080/?month=Oct&region=NY` (Controls panel)
- `http://localhost:8080/?by=frequency&month=Oct&region=NY` (screenshot)

## T4. Three ducks

**Answer:**
- **Widest wingspan: Mallard**, 34.6 in. Wood Duck is 28.0 in; Canvasback is in between.
- **Heaviest: Canvasback**, 2.6 lb. Mallard is a bit lighter; Wood Duck is 23.2 oz, about 1.5 lb.

**Confidence:** 4 for wingspan, 3–4 for weight.
- On the compare screen, the Mallard's green wing line is clearly the longest.
- The numbers on the drawings are microscopic on my phone; I couldn't read them.
- So I tapped into Canvasback and Mallard details for the actual numbers. I never got a clean readable Mallard weight, only a squint at the tiny label.

**Where:**
1. Searched Mallard and opened its page.
2. Tapped "Compare with another bird…".
3. Searched each other duck and tapped "+ Pick", then "Compare".
4. On the compare page ("Compared to one scale"), tapped through to "Species details".

**Visited:**
- `http://localhost:8080/?q=mallard`
- `http://localhost:8080/species/mallar3?q=mallard`
- `http://localhost:8080/?q=wood+duck&species=mallar3`
- `http://localhost:8080/?q=canvasback&species=mallar3&species=wooduc`
- `http://localhost:8080/compare?species=mallar3&species=wooduc&species=canvas` (text + screenshot)
- `http://localhost:8080/species/canvas?species=mallar3&species=wooduc&species=canvas`
- `http://localhost:8080/species/mallar3?species=mallar3&species=wooduc&species=canvas`

## T5. Solid and dashed

**Answer:** Solid lines are measured; dashed lines are estimated, and so is the hollow head. On the wing, the solid ends are the measured "wing chord", which I don't know, and little ticks mark where it ends. The cube is the bird's weight drawn as water.

**Confidence:** 5 for solid vs dashed. The wing chord part went over my head.

**Where:** The note right under the drawing on a bird's page (Merlin), and again under the compare page.

**Visited:**
- `http://localhost:8080/species/merlin?q=merlin`
- `http://localhost:8080/compare?species=mallar3&species=wooduc&species=canvas`

## T6. Sharp-shinned Hawk

**Answer:** **4.6 oz.** For fun: same length as a Merlin (11.5 in), but lighter.

**Confidence:** 5

**Where:** Searched "sharp-shinned" and read the bird card.

**Visited:** `http://localhost:8080/?q=sharp-shinned`

---

## E1. What confused or annoyed me

- **Month and region filters work backwards.** Everything starts selected, so tapping "NY" would have *hidden* New York.
  - I had to read the small print to learn I needed "All" first, to clear.
  - On a phone, the region list is a wall of buttons, and it appears twice ("grouped two ways"). That's too much.
- **Compare screen labels are unreadably tiny on a phone.** I could see which wing line was longest, but couldn't read any numbers off the drawings.
- **Drawings in the frequency list are minuscule.** When I sorted New York birds by frequency, the scale switched to "Global" by itself. The Blue Jay became a speck.
- **Jargon and clutter.** "Section step 15%", "Length Tier 10", "Per section", "wing chord", "modeled estimated", species codes like "mallar3". There's a lot of science-y stuff I don't need.
- **Odd numbers.** One bird said "0 in 89,525,032". The Merlin's total length is "estimated" even though it's a common bird.

## E2. What I liked

- Search is great. I typed a name and got one card with length, wingspan and weight right there. That covered T2 and T6 in seconds.
- Seeing "1 in 24 observations in New York, Oct" for Blue Jay is cool. Once I got the filter set, it felt like a list of what I'd actually see on my walk.
- "Compare with another bird…" then "+ Pick" is a nice idea, and the plain "solid = measured, dashed = estimated" note is easy to get.
- The "small / medium / large / huge birds" starter links are friendly.

## E3. One change I'd make

Make the month and region pickers work the way I expect: tap "October" and "New York" to show *only* those. Add a big "Birds near me this month" button that does it in one tap, sorted by most common, and keeps the drawings big enough to see.

## E4. Ease of use

**3 / 5.**
- Looking up one bird is very easy.
- Anything with filters or comparing took real effort on my phone, and I'd have given up on the subway.
- Would I come back? Maybe, to check a bird's size after Merlin IDs it. Probably not to browse.
