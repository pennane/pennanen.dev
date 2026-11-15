# 30. Tracks, seasons and episodes are parts, and parts are items

**Status:** accepted · 2026-11-15 · amends [ADR 24](./0024-series-are-their-own-kind.md)

## Context

An album page should list its tracks, and later rank them by listening and
let ap rate them. A series has seasons ap rates on their own ("season 2 fell
off") and episodes inside them. A tracklist, or a series' seasons, has an
order that belongs to the whole: the same recording can be track 3 on the
album and track 1 on the single. Spotify gives each release of a recording
its own track id, but the recording code (ISRC) is the same.

## Decision

- `track`, `season` and `episode` are kinds. Their items are known until ap
  logs one; a logged one takes statuses, axes and ranks like any item, and
  ranks only against its own kind.
- A **part** is an item that is one piece of a **whole**, at a position: a
  track of an album, a season of a series, an episode of a season. Parts are
  kept apart from relations because they are ordered and the order is the
  whole's. One item can be a part of several wholes, at a position in each.
- The whole's source says what its parts are and in what order. A sync
  replaces a whole's parts all at once, as it replaces a claim.
- A track is a recording: its identity is its ISRC, so a single and the
  album it is on share the track. A Spotify track with no ISRC is its own
  recording, keyed by its Spotify id. Remasters and live versions have their
  own ISRCs and stay apart.
- Parts are fetched for logged wholes only, never for dismissed ones.
  Spotify's album tracklists carry no ISRC, so a whole whose parts are
  missing or out of step fetches its tracks in batches of 50; one with its
  parts in step costs no calls.
- A season of specials is listed after the numbered seasons.
- A whole's public page lists its parts in order. Parts have no public list
  of their own yet; a logged part shows its scores on its whole's page when
  its entry is shown.

## Consequences

- ADR 24's "seasons and episodes aren't tracked" no longer holds.
- Top-track and play observations stay on albums for now. Moving them onto
  tracks needs no new mechanism, but it is its own decision.
- About two thousand tracks become known items on the first sync. Lists show
  logged items only, so they don't slow down.
