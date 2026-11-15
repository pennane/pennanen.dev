# 29. Listening orders music, and activity is its own axis sort

**Status:** accepted · 2026-11-15 · amends [ADR 12](./0012-divisions-and-ranks.md),
[ADR 18](./0018-axis-sorts.md) and [ADR 19](./0019-observations.md)

## Context

ADR 12 has an entry without a rank join the bottom "in an order the kind
chooses (playtime for games)", but no kind ever chose one in code, and
public lists fell back to title. Artists and albums arrive by the hundred
from Spotify with no rank, so their pages read as random. Spotify's top
lists and plays say how much ap listens, but only as observations. Playtime
sat under crowd axes, which ADR 18 meant for a site's public average.

## Decision

- A kind names its **unranked order**: a number field, highest first, used
  after every ranked entry. Games use playtime; artists and albums use
  listening. The public list gets each entry's place in that order, never the
  value, so private fields stay private (ADR 23).
- **Listening** is a claim from the derived source `spotify-listening`,
  rewritten at the end of every nightly `spotify` sync from all observations:
  - a top-list appearance at rank r of N scores (N − r + 1) / N, times 3 for
    the all-time list, 2 for six months, 1 for four weeks;
  - an album sums its tracks' appearances, plus 0.1 per play;
  - an artist sums its own appearances, plus every album it is `by`.
  - The stored value is the percentile, 0 to 100, among the kind's logged
    entries, so axis floors keep their meaning as history grows.
- A fourth axis sort, **activity**, reads fields about what ap did
  (playtime, achievements unlocked, listening). Crowd keeps only a site's
  averages.
- ADR 19's "albums aren't judged by play counts" holds for statuses; plays
  do weigh into listening.

## Consequences

- The formula can change and rerun over all history, since observations are
  never thrown away.
- An axis that read playtime as crowd must be saved again as activity.
- Ties in listening, mostly saved albums never heard, fall back to title.
