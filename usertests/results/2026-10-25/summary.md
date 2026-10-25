# Simulated users, 2026-10-25

Four readers ([personas](../../personas.md)) on two revisions: **old** is `4f25193`, what the first
real readers saw; **new** is `b3b2f93`, with the phone layout, the prose pass and the search box.
Same [tasks](../../tasks.md), scored against [answers.md](../../answers.md). Reports are in
[old/](old/) and [new/](new/).

## Scores

Every reader answered T2–T6 correctly on both revisions, so correctness does not separate them;
effort and ease do. Old-Maya answered T4 from the sort order because the numbers were unreadable;
new-Maya could not read the Mallard's weight on the comparison and gave no number for it.

| Reader | Old ease | New ease | Old: how they reached birds | New |
| --- | --- | --- | --- | --- |
| Maya, new birder, phone | 2 | 3 | Find-in-page and guessed addresses | Search |
| Dan, field birder, laptop | 2 | 3 | Typed guessed species codes into the address | Search |
| Funmi, ornithologist, laptop | 3 | 3.5 | Herons-and-ducks shortcut, then picking | Search |
| Lena, German birder, phone | 2 | 3 | Guessed codes and built addresses | Search |
| **Mean** | **2.25** | **3.1** | | |

## Calibration: did the old run find what real readers found?

| Real readers said | Old simulated readers |
| --- | --- |
| No way to search (Ryan, #3) | All four, first; every E3 asked for a search box |
| Dense and intimidating (Ryan, Joe) | Maya ("the first screen is all settings") and Lena (jargon: Tier, Section step) |
| Text past its boxes on phones (Karen, Ryan, #2) | Weakly: Maya noted the page "slightly wider than my screen". The readers saw panels closed, where the overflow was |
| "All X start selected" reads oddly (Karen, #1) | All four, and more strongly: tapping a month turns it *off* |

Good enough to compare designs with, with two caveats. The readers are more capable than people:
they guessed species codes and built addresses by hand, which a real reader would not. And the
text tool shows panels closed by default, so layout faults inside a panel are seen only in
screenshots the reader chooses to take.

## What the new run found

Ranked by how many of the four raised it.

1. **Month and region toggles work backwards (4 of 4).** Everything starts selected, so tapping
   October or New York *removes* it; three of the four built the address by hand instead. Karen's wording fix
   did not reach the behavior.
2. **The comparison's numbers are unreadable on a phone (2 of 2 phone readers, both runs).** The
   labels inside the drawings are a few pixels high; both phone readers opened each bird's page
   to get the numbers, so the comparison did not answer "which is bigger".
3. **Jargon and provenance noise (3 of 4).** "Section step", "Tier", "wing chord", "AVONET2_eBird",
   "mass^0.37"; the provenance table repeats itself ("measured" tag, then "measured, from…";
   "modeled" tag, then "estimated: …").
4. **Units (2 of 4).** Inches first, cm/g hidden in Controls; Funmi saw units change between pages.
5. **Comparison sort highlights more than one button (2 of 4).** By design a metric reads as on
   when the birds already happen to be in its order, so Bill and Tail both lit up; readers took
   it for a fault.
6. **Frequency forces Global scale (Maya).** Sorted by frequency, the drawings were specks.
7. **The region list is a wall of buttons, shown twice (Maya).**
8. **Mass is one number (Dan, Funmi).** They wanted a range or a male/female split where the
   source has one.

What readers liked: search ("covered T2 and T6 in seconds"), the measured-versus-modeled honesty
and sources (all four, both runs), the New York in October frequency list, and the small, medium,
large and huge links.
