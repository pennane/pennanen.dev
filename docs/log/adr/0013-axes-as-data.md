# 13. Axes are data, shared by kinds; prestige orders things

**Status:** accepted · 2026-11-10 · supersedes the axis labels of
[ADR 11](./0011-presentation-settings.md) and the axis words of
[ADR 12](./0012-divisions-and-ranks.md) · amends [ADR 5](./0005-kinds-defined-in-code.md)

## Context

Axes were words in each kind's Go definition. Adding one meant a deploy, and
games and things each carried their own labels and divisions for what is the
same axis. ap thinks of the ordering axis as prestige, not enjoyment, and of
divisions as tiers.

## Decision

- An **axis** is its own record: a key set once at creation, a name ap can
  change, an order, and its divisions. Kinds use axes; admin picks which and
  in what order. The first axis a kind uses leads on list rows.
- Divisions belong to the axis, so every kind using an axis shares them.
  Per-kind axis labels and per-kind divisions are gone.
- **Prestige** replaces enjoyment, key and all. It is the only ranked axis,
  fixed in code: it can be renamed, never deleted, and every kind that ranks
  uses it. Other axes have divisions only.
- Divisions default to tiers `S A B C D F`. Prestige and quality migrate
  loved/masterpiece → S, liked/great → A, fine/solid → B, meh/rough → C,
  disliked/broken → D; F starts empty. Ranks inside each tier are kept. Where
  games and things disagreed on a name, the game's wins, except that the
  ranked axis is named "prestige".
- Removing an axis from a kind, or deleting it, is refused while entries hold
  scores on it.
- Axes are only set by ap. An axis filled from source claims may come later
  as its own decision.

## Consequences

- A new axis is an admin action, no deploy.
- Renaming an axis or a division changes only what pages say; keys stay.
- Giving a kind its own variant of an axis means making a second axis.
