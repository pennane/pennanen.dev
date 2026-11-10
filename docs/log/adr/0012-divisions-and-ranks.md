# 12. Scores are divisions; enjoyment is a rank

**Status:** accepted · 2026-11-10 · supersedes [ADR 8](./0008-rankings.md) and
the score words and score display of [ADR 11](./0011-presentation-settings.md)

## Context
Scores were numbers from 0 to 100. Triage could only give four of them
(92, 80, 60, 35), so 200 rated games sat in four ties and in an order ap
disagreed with. Moving one meant guessing a number. ap thinks in "above or
below this one", finds 0–100 too fine, yet wants a clear best-to-worst order
of games. Quality is an annotation; enjoyment is what orders things.

## Decision
- An axis is an ordered list of named **divisions** (`loved` > `liked` >
  `fine` > `meh` > `disliked`), set in admin per axis, as many as it needs.
  A score is the division an entry sits in. No numbers are stored.
- Enjoyment also has a **rank**: a strict order of a kind's entries across
  its divisions. No ties. Other axes have divisions only.
- Rating picks a division, then asks *better or worse than X?* to find the
  rank, and can stop early. An entry given a division without a rank joins
  the bottom of it, in an order the kind chooses (playtime for games).
- Public pages show the division and, for enjoyment, the rank among public
  entries: *loved · #14 of 214 games*.
- Score words and the number/word/both display setting are gone; divisions
  replace them.
- Rankings (ADR 8) are removed, unbuilt in admin and on public pages.
  Cross-kind or per-tag lists may come back later as their own decision.
- Moments no longer carry scores.
- Old scores migrate once: the nearest of 92/80/60/35 picks the division,
  and within it the higher old number ranks first, then playtime. The old
  numbers stay in `number_score_archive` until nothing needs them.

## Consequences
- Reordering is a drag or a comparison, never a typed number.
- Adding an entry shifts the ranks below it; the division stays.
- A rank is a position inside its division; a move renumbers that division,
  which is a few hundred rows at most.
- Ranks never span kinds.
