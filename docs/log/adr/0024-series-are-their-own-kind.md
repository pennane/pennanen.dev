# 24. Series are their own kind

**Status:** accepted · 2026-11-14

## Context

Letterboxd lists TV series as films, and the `movie` kind took them in as
films. Searching TMDB's films for a series either found nothing or found the
wrong thing: _Ripley_ (2024) became _In the Realm of Ripley_. Many series
ap has seen were never on Letterboxd, so they have to be added by hand too.

TMDB numbers series and films separately; series 1398 isn't film 1398.
Letterboxd's film page links the right one, `themoviedb.org/movie/…` or
`themoviedb.org/tv/…`, even though its `data-tmdb-type` attribute says
`movie` for both.

## Decision

`series` is a kind of its own. It has the statuses of `movie` plus `dropped`,
the same axes and Letterboxd take, and fields of its own: creators, seasons
and episodes in place of directors and runtime. Its identity is a TMDB series id, kept by the
source `tmdb-tv` so it can't collide with a film's.

The Letterboxd claim records which TMDB entry a film is, `tmdbType` (`movie`
or `tv`) next to `tmdbId`. A film from Letterboxd with no TMDB id yet gets one
fetch of its film page. The link there decides its kind before an item is
created, so a series ap added by hand is joined rather than duplicated. Title
search stays only as the fallback for films whose page links nothing.

When a match says an item is the other kind, the item **moves**: it keeps its
entry, scores, tags, prose, takes and observations, and gets a new slug if its
own is taken there. Ranks are per kind, so it goes to the bottom of its
division in the new kind.

Seasons and episodes aren't tracked, and there's no `watching` status.

## Consequences

- ap adds series and films from a TMDB search on the add page, as games come
  from IGDB.
- Pasting a `themoviedb.org/tv/…` link on a film moves it to `series`.
- The nightly TMDB match moves the series it finds among unmatched films.
- Two items with the same identity aren't merged: the second match is
  refused.
- `series` gets its public page, filters and tier lists from ADR 23.
