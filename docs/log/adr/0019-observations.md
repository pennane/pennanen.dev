# 19. Observations keep what claims forget

**Status:** accepted · 2026-11-11 · amends [ADR 4](./0004-claims-and-verdicts.md)

## Context

A claim is replaced wholesale on every sync, so it only says how things are
now. Some signals only mean something over time: an artist in the 4-week
top 5 all through March, the day a Vivino rating first appeared (Vivino only
shows "17 days ago"), each Letterboxd diary entry and rewatch.

## Decision

- An **observation** is something a source said at one moment, appended and
  never changed or replaced.
- Observations hold: Spotify top lists per range per day, first-seen dates
  of outside takes, and Letterboxd diary entries.
- They are evidence, not verdicts: rules (suggestions, dates) read them;
  nothing shows them raw by default.

## Consequences

- Observations grow forever; one row per signal per day is small enough.
- Data for later graphs piles up from the first sync, even before anything
  reads it.
- Spotify's last-50 limit on recently played doesn't matter: albums aren't
  judged by play counts.
