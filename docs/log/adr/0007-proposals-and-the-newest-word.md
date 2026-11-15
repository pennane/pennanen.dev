# 7. Proposals, and the newest word wins

**Status:** accepted · 2026-11-15

## Context

Sources know things about ap's statuses and tags: a film rated on
Letterboxd was watched, an artist in this month's top list is being
listened to. There used to be four ways of letting them say so, each coded
inside its source: undoing an action on the site sent a status back to
`untriaged`; artists had a computed status where only `dismissed` stuck;
saved albums kept their status whatever happened to the save; tags kept
track of who added them. Hints, suggestions and the match queue were three
more queues.

The computed status took control away from ap to solve "I might forget I
pinned something". Music changes weekly and shouldn't need unpinning;
films barely change, and ap's word should hold.

## Decision

- A **rule** is a pure function from readings, observations and the world
  to **proposals**. A proposal says what an item should have (a status, a
  tag, an entry, a match, a merge), who says so, when that value last
  changed, and its strength: **applying**, which takes effect on its own,
  or **suggesting**, which waits for ap's yes.
- **The newest word wins.** Among ap's choice and the applying proposals,
  the one whose value changed most recently decides. Saying the same thing
  again is not a change. With nothing, a status is `untriaged`.
- A proposal's time comes from records: a claim's `changed_at` and the times
  of observations. Never from when derivation ran, so a rebuild changes
  nothing. A proposal with no known time is older than any choice of ap's.
- A source that stops saying something withdraws its proposal; the next
  word down decides.
- **Triage** is one screen for every kind: suggesting proposals and
  untriaged entries, one card at a time, with bulk actions.

What the old rules became:

| Before | As proposals |
|---|---|
| Letterboxd watched / watchlist → `watched` / `want` | applying status proposals |
| Steam hints (`likely-idled`, …) | suggesting status (or tag `idled`) proposals |
| Artist computed status | applying `listening`/`listened` from top lists; no special flag |
| Saved album `want` → `listened` | applying, read from observations (unsave doesn't withdraw it) |
| Steam owned → logged; top artists → logged; saved album → logged | applying entry proposals (create the entry) |
| Album suggestions | suggesting entry proposals |
| Match queue | suggesting match proposals |
| `another-shot` from watchlist | applying tag proposal |
| Listening | a derived number field computed by a rule from observations, not a fake claim |

_ap sets an artist to `listened`. Next month it's back in the 4-week top
list, the proposal changes to `listening`, and the lists lead again. If the
lists say the same thing for a year, ap's choice stands._

_ap sets a film to `dropped`. Letterboxd has said `watched` since before,
unchanged, so ap's choice holds. A newly logged rewatch would move it._

## Consequences

- Fast sources lead and slow ones defer to ap, with no flag on any kind.
- Admin shows when ap's choice and the current proposal disagree:
  _your choice: listened · lists say listening_.
- Hints, suggestions and matches all wait in triage.
- A new rule is a function; it can't reach takes.
- A source's rules live with it and run after its sync; the newest word is a view, never stored.
