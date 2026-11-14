# 23. Public pages follow field types; admin only overrides

**Status:** accepted · 2026-11-14 · extends [ADR 11](./0011-presentation-settings.md)

## Context

The public browser was written for games: a playtime filter, search by
developer, a "last played" column, the overview's top five by rank. Movies
arrived and every one of those needed a hand-made twin, and each new kind or
field would too. ap wants new fields to show up on their own, and to choose
in admin what each kind's pages offer.

## Decision

- Every kind field has a **type** and a **reach**: for everyone, or only ap.
  Fields only ap sees (Steam playtime) never leave the log's API.
- An outside take declares types for the values worth listing (Letterboxd
  rating, last watched). Its other values, like the review, show as the take
  itself on the entry's page, never as list fields.
- The public presentation lists each kind's **public fields** with their
  type. A field's type decides what it can do by default: dates and numbers
  filter, sort and show as columns; lists filter and search; text searches;
  ids and links do nothing.
- Per kind, admin overrides: each field's uses (filter, column, page) and
  public name, which axes get a **tier list** and which a list column, the
  **default view** (tiers,
  list, or tiers once something is scored), and the **overview**: which
  sort picks the entries on the log's front page, and how many.
- A field nobody has set keeps its type's defaults, so a new field appears
  without anyone touching admin.

## Consequences

- No public page code names a field of a particular kind.
- Number filters band themselves from the data; date filters go by year when
  a field spans two decades or fewer, by decade otherwise.
- The overview's automatic sort is rank when anything is ranked, else the
  first axis or number or date most shown entries have.
- Steam playtime no longer reaches public pages, as the glossary always said.
