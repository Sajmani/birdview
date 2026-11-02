# 2026-11-01: labels in the drawing against measurements on the tile

Four readers (Maya and Lena on a phone, Dan and Funmi on a laptop) each ran the standing tasks
plus T7 and T8 ([tasks.md](tasks.md)) against two revisions:

- **old**, `c989ff9`: the species page labels each part inside its drawing, with leader lines.
- **new**, `fdab1f9`: the species page shows the home page's tile, larger, naming every part and
  its measurement, with no text below 13px (decisions.textSizeFloor).

A first pass at the old revision is in `invalid-old/` and is not counted: a server left over from
the 2026-10-25 run still held its port, so those readers saw the classic view.

## Scores

| Task | old | new |
| --- | --- | --- |
| T7, the heron's bill and tail | 4 of 4 correct, confidence 5, 5, 5, 5 | 4 of 4 correct, confidence 5, 5, 4, 5 |
| T8, the wing chord: the part, and its length | 4 of 4 correct, confidence 3, 5, 5, 4 | 4 of 4 correct, confidence 3, 4, 4, 3 |
| E4, ease, 1 to 5 | 2, 3, 3, 3 | 3, 3, 3, 3 |

## What the comparison shows

**The labels were too small, and every old reader said so.** On a phone they were unreadable:
"far too small to read" (Maya), "I cannot read any of the labels" (Lena). Both took the numbers
from the table below the drawing instead. On a laptop they were readable with effort: "I had to
squint" (Funmi), "in very small print" (Dan). No new reader called the measurements too small;
all four read the bill and tail off the tile.

**The leader labels did one job the tile does not: they put "Wing chord" on the bar.** The two
laptop readers on the old revision answered T8 with confidence 5, reading the label on the left
outer segment. New readers found the chord from the note under the drawing and the ticks, and
said so: "the small ticks are the only thing that marks off the chord, and I would not have known
that without reading the text" (Dan). Where the wingspan is measured, the whole bar is solid, so
the ticks are all that separates the chord from the rest.

## Found on both revisions

- **The settings would not open for any reader.** This is the harness, not the site: `page --open`
  knows only the classic view's panel names, and `shot` cannot click. Every reader reached New York
  in October through the guide's example link instead.
- **Pinned birds stack in one column under tier headings**, one bird per tier, so "side by side"
  means scrolling (Dan, Funmi, Maya, Lena).
- **Mass changes unit between pages**: the Mallard is 29.8 oz on the home page and 1.9 lb when
  pinned (Dan, Maya, Lena, Funmi).
- **One mass per species**, no sex or range, for strongly dimorphic raptors (Dan, Funmi, Maya);
  already [sexSpecificDrawings](../../../river/todo.md#sexspecificdrawings).
- **The modeled total length runs short**: the heron's 34.7 in against about 46 in in field guides
  (Dan, Funmi).
- **Small birds are specks on a phone tile** at the page's shared scale (Maya, Lena).
