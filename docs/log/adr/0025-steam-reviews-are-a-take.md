# 25. Steam reviews are an outside take on games

**Status:** accepted · 2026-11-14 · follows [ADR 17](./0017-outside-takes-set-statuses.md)

## Context

ap reviews some games on Steam: a thumb up or down, a few words, the date, and
the hours played when writing. The Web API has no call for a user's own
reviews, but the public reviews page on the Steam community lists them all,
ten to a page.

## Decision

A Steam review is an outside take from the source `steam-take`, read by the
Steam sync from the reviews pages after the library. It records the
recommendation (`recommended` or `not-recommended`), the review, the date it
was posted and the hours at review time. Hours are private like playtime.

A review is joined to the game with that Steam app id. A review of a game not
in the library is a problem in the sync report; it never creates an item.

A review sets no status. Game statuses tell played, finished and dropped
apart, and a thumb doesn't.

## Consequences

- A game's page shows the thumb and the review beside ap's prose, as a film's
  shows Letterboxd stars.
- A thumb isn't a number, so no axis can read it.
- Steam leaves the year off reviews posted this year; the sync takes it as the
  current one.
