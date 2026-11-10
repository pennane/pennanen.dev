# 16. Items can be known without being logged

**Status:** accepted · 2026-11-11

## Context

Every item made by a sync got an untriaged entry. With games that cost a
look per game, and the hints that should have saved it were right about
60% of the time. Music would be far worse: every album in a top list would
land in triage.

## Decision

- An item with an entry is **logged**. An item with no entry is only
  **known**. No new state: whether an entry exists is the whole difference.
- Each source decides per kind whether what it finds is logged or only
  known. Steam keeps logging every owned game, since a library is a log.
- A **suggestion** is a known item a rule thinks ap might log: for albums,
  a saved album, several tracks in the top tracks, or an album by an artist
  ap has logged. Suggestions wait in admin; they never fill a queue.
- Dismissing a suggestion logs the item as `dismissed`, so it isn't
  suggested again. This may be revisited.
- Open: whether Spotify logs artists from follows and the long-term top list,
  or only suggests them. Decided when music is built.

## Consequences

- Public pages and counts only ever see logged items.
- Logging a known item creates its entry; nothing else moves.
