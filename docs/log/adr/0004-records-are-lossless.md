# 4. Records are lossless

**Status:** accepted · 2026-11-15

## Context

Sources disagree, change their answers and forget. A re-sync must never
clobber ap's fixes, and the model must be free to change later, replaying
what sources once said. Some of it can never be fetched again: Spotify's top
lists only show today, and Letterboxd blocks paginated pages.

## Decision

- A **claim** is what one source says now about one thing, kept exactly as
  received. It is keyed by **identity**, `(source, external id)`, not by
  item. A re-sync replaces it wholesale; nothing edits it. The store sets
  its `changed_at` when its reading differs from the one before.
- An **observation** is what a source said at one moment: _Pendulum in the
  4-week top 5 on 2026-03-14_. Appended, never changed or replaced.
- A **reading** is what a source's reader takes from a claim. It is stored
  beside the claim, written in the same transaction and tagged with the
  readers version; at start, claims with no reading or an older version
  are read again. Everything above records works from readings, never from
  claim bodies.
- ap's own input is the source `ap`. An **outside take**, ap's rating or
  review on another site, is a claim from that site's take source
  (`letterboxd-take`, `steam-take`), apart from what the site says about
  the thing.

## Consequences

- Data is stored redundantly on purpose. Changing a source ranking, a reader
  or the whole model needs no migration of records, only a rebuild.
- Records are never trimmed. They and takes are what the backup is for.
- Observations grow forever; one row per signal per day is small enough.
- A claim no reader can read is read again on every start.
