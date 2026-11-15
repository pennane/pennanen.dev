# 8. Scoring

**Status:** accepted · 2026-11-15

## Context

Scores were numbers from 0 to 100, which ap found too fine and always
guessed at. ap thinks in "above or below this one", wants a clear
best-to-worst order within a kind, and thinks of that order as prestige,
not enjoyment or quality. Other sites hold ap's ratings and the crowd's,
and some numbers say what ap did: playtime, listening.

## Decision

- An **axis** is data: a key set once, a name ap can change, and its
  **divisions** in order from best, by default `S A B C D F`. Kinds pick
  their axes in admin; the first leads on list rows. A score is the
  division an entry sits in on one axis. There is no number behind it.
- **Prestige** is fixed in code: renamable, never deletable, the one
  ranked axis. Its **rank** is a strict order of a kind's entries across
  divisions: no ties, never across kinds.
- Rating picks a division, then asks _better or worse than X?_ until the
  entry has a rank, and can stop early.
- An entry with no rank joins the bottom of its division in its kind's
  **unranked order**: a number field, highest first. Public pages get the
  place, never the value.
- A **derived axis** reads a number field, from a take, a describing source
  or a field about ap, through a floor per division, best first. Nobody
  sets it by hand and it never ranks. One that reads an outside take
  suggests a division while ap rates; the rest only show.

## Consequences

- Reordering is a drag or a comparison, never a typed number.
- Renaming an axis or a division changes only what pages say.
- Changing a derived axis's floors rescores every entry at once.
- Removing an axis from a kind is refused while entries hold scores on it.
