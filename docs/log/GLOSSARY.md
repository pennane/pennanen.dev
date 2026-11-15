# things as ap saw them — glossary

The words used for the log, in code and in conversation. If a word here
disagrees with the code, one of them is wrong. The words for layers 2–5
follow ADRs 3–10, and the code is still moving to them; where a word
replaces older ones still in code, its entry says which.

## Layers

**Layer** — one of five, each reading only those below: records, world,
takes, rules, presentation ([ADR 3](./adr/0003-the-layered-model.md)).

**Record** — what a source said, kept as received: a claim or an
observation. Layer 1. Sources write nothing else.

**World** — what the log knows about things: items, their kinds, fields and
links. Layer 2, derived from records.

**Take** — ap's opinion of a thing. Layer 3. ap writes takes, and corrects
the world through `ap` records; nothing else.

**Rule** — a function from records and the world to proposals. Layer 4.

**Presentation** — what public pages show and how. Layer 5, stored as the
projection.

**Derive** — work out a derived layer from the ones below it. Each source
has a derive from its readings to items, fields and links.

**Dirty** — touched by a new record or take and waiting to be derived again,
with its one-step neighbours through links.

**Rebuild** — derive everything from records and takes again. Runs at start
when a derive, reader, rule or projection version changes.

## Records

**Source** — somewhere data comes from: Steam, IGDB, TMDB, Letterboxd,
Spotify, or `ap` by hand. A source writes records three ways: put a claim,
remove a claim (which leaves an observation), observe.

**`ap` record** — ap correcting the world: a field override, an item added
by hand, a match answer, a link edit, a merge answer. Read by the world like
any source's records, and as irreplaceable as takes.

**Identity** — `(source, external id)`: what a record is about. _`igdb:1942`,
`tmdb:603`, `isrc:GBUM71029604`._

**Claim** — what one source says now about one identity, exactly as
received. Never edited; a re-sync replaces it whole. Its `changed_at` is
set on put-claim when the new reading differs from the stored one;
re-reading never moves it. _Steam claims 412h._

**Observation** — something a source said at one moment, kept for good:
_Pendulum in the 4-week top 5 on 2026-03-14_, _watched Heat again on
2026-02-01_. Claims say how things are now; observations say what happened
when.

**Reading** — what a source's reader takes from one claim, stored beside it
and tagged with the readers version. Everything above records reads
readings, not claims.

**Outside take** — ap's take recorded on another site, kept as claims from
its own source (`letterboxd-take`, `steam-take`), apart from what the site
says about the thing. Shown alongside the entry, never merged into it.

## World

**Item** — something that can be seen: a game, a film, an album. The records
that share an identity. It has a stable id and slug, a kind, and a set of
identities; opinions live elsewhere.

**Kind** — what sort of item it is. A kind decides its fields, statuses,
identity source, unranked order and rules. Kinds are defined in code; their
**title**, **description** and axes are set in admin. There is no catch-all
kind: something that fits no kind waits until it gets one.

**Field** — one fact about an item, declared by its kind, with a **field
type** (text, int, number, partial date, string list, url, external id,
picture), a **reach**
(everyone, or **only ap**: Steam playtime) and, for a picture, a shape.

**Verdict** — the value that wins for a field. Sources are ranked per field;
the highest-ranked one with a value wins, and ap's override ranks first,
always. _Steam claims 412h, IGDB claims nothing, so the verdict is 412h._

**Inherited** — a field whose verdict, when no source of the item has a
value, comes from the item at the other end of a link of a declared type,
one step away. _An album with no genres shows its main artist's._

**Match** — linking records from different sources to one item. A sure match
is derived; an unsure one is a suggesting proposal; ap's answer, a pick or
**no match**, is an `ap` record.

**Move** — an item's kind changing because its derivation changed. _A
Letterboxd film whose TMDB link says `tv` moves to `series`._ It keeps its
id and entry and goes to the bottom of its division.

**Merge** — two items found to share an identity becoming one. The one whose
entry was edited last keeps its id and slug; the other entry is a **parked
entry**, proposed for merging in triage. Yes folds it in; no keeps the items
apart for good. _Replaces refusing the second match._

**Link** — `from, to, type, origin`, plus a position for ordered types. The
graph of items. _Replaces relation, part and mention._
- **by** — an album is by its **main artist**, the first one credited.
  Spotify's "Various Artists" is never a main artist.
- **with** — every other artist credited on an album.
- **part-of** — ordered: a track on an album at its position.
- **mentions** — a link written in ap's prose. _Realm of the Mad God goes
  with Pendulum._

**Origin** — where a link comes from: a source, `ap`, or `prose`.

**Position** — where a link of an ordered type sits among its whole's
links. Only ordered types have one.

**Part / whole** — an item that is one piece of another, at a position the
whole decides. One recording can be a part of an album and a single, at a
different position in each. A part is an item like any other.

## Takes

**Entry** — ap's take on one item, at most one per item: visibility, scores,
rank, blurb, prose, time spent.

**Logged / known** — an item with an entry is logged; one without is only
known: a sync came across it, but nobody put it in the log. Logging is
one-way: a source that logged an item never unlogs it.

**Status** — where ap is with the item, regardless of what ap thought of it.
Each kind has its own set; most start at `untriaged`. A status is ap's
choice or a proposal's, whichever word is newest.

**Status choice** — a status ap picked, with when. Kept if the item moves to
a kind without that status, and set aside until it moves back.

**Dismissal** — "this doesn't belong in my log", like Wallpaper Engine in a
Steam library. A take, not a status: only ap applies it (a hint may suggest it), and
nothing overrides it. The
item is kept so syncs don't bring it back, but never shows. _Replaces the
`dismissed` status._

**Tag** — a free-form label on an entry: `favorite`, `comfort`,
`formative`. Nest with a slash: `favorite/all-time`. What an item _is_
(`co-op`) is a field, not a tag. A tag is there by ap's **tag choice** or a
proposal's, whichever word is newest.
- `another-shot` — ap wants to try it again, whatever its status.
- `idled` — a game ap ran for trading cards, whether or not ap also played it.

**Link edit** — ap adding a link or suppressing one; an `ap` record.

**Match answer** — ap's pick of the item a record is about, or no match, so
nobody asks again; an `ap` record.

**Time spent** — ap's word on how long it took: `not-enough`, `a-bit`,
`just-right`, `a-lot`, `too-much`. A judgment, not a measurement.

**Blurb** — public text, written for others, shown on the item's page.

**Prose** — what ap writes about an entry: one Markdown document, with images
and links to other items, public exactly when the entry is. Pages call it
_visions_; each passage is a vision.

**Visibility** — whether an entry shows: `auto` (by its kind's shown
statuses), `public` or `private`. Prose follows its entry.

## Rules

**Proposal** — what a rule says an item should have: a status, a tag, an
entry, a match or a merge, with who says it and when its value last
changed. Its time comes from records, never from when it was derived.
_Replaces `status_set_by`, tag `added_by`, `StatusComputedBy`, hints,
suggestions and the match queue._

**Applying / suggesting** — a proposal's strength. Applying takes effect on
its own; suggesting waits in triage for ap's yes.

**Newest word** — how a status, tag or link is decided: among ap's choice and
the applying proposals, the one that changed most recently wins. Saying the
same again is not a change; a proposal with no known time is older than
ap's choice ([ADR 7](./adr/0007-proposals-and-the-newest-word.md)).

**Withdraw** — a source no longer saying something removes its proposal; the
next word down decides.

**Hint** — a suggesting proposal for a game during triage: a status
(`likely-played`, `likely-never-played`), a tag (`likely-idled`), or
dismissal (`likely-not-a-game`), which only ap's yes applies.

**Suggestion** — a suggesting proposal to log a known item, such as an album
by an artist ap has logged.

## Scoring

**Axis** — one dimension something can be scored on. Made, named and ordered
in admin and shared by every kind that uses it; each kind picks its axes. A
key, set once, and a name ap can change.

**Division** — a named slice of an axis, in order from best. Axes start with
`S`, `A`, `B`, `C`, `D`, `F`. There is no number behind a division.

**Score** — the division an entry sits in on one axis. Always optional.

**Prestige** — the axis that orders things: how high something stands for
ap, all told. Not quality: a masterpiece can bore, trash can delight. The
one axis with ranks.

**Quality** — the axis for how well made something is, whatever ap thought.

**Rank** — an entry's place in its kind's prestige order, across all
divisions. Strict: no ties. Never across kinds. Public pages order by it and
never show the number.

**Unranked order** — the number field a kind orders its unranked entries by,
highest first, after every ranked one: playtime for games, listening for
artists and albums. Public pages get the place, never the value.

**Derived axis** — an axis read from a number field through a **floor** per
division, best first: the lowest value that lands there. Never set by hand,
never ranked. One read from an outside take suggests a division while ap
rates. _Replaces outside, crowd and activity axes._

**Listening** — how much ap listens to an artist or album, from 0 to 100,
worked out nightly by a rule from observations. Public. _Replaces the
`spotify-listening` claim._

## Presentation

**Projection** — what each public page shows, resolved in Go and stored.
_Replaces `/public/presentation` and the site's own resolving._

**Shown statuses** — the statuses of a kind whose `auto` entries show, set
in admin.

**Tier list** — a kind's shown entries laid out by division on one axis,
best first; entries with no score on that axis are left off.

**Overview** — the log's front page: a few shown entries per kind, picked
by the kind's overview sort.

**Picture** — a field that is an image, with a **shape**: `tall` (a poster
or box art), `wide` (a Steam cover, a film's backdrop) or `square` (an
album cover, an artist). A **tile** is the picture a tier list shows, tall
first; a **banner** the one a row or an item page shows, wide first.

**Tag title** — what public pages show for a tag; the tag itself is its id.

## Kinds

**game** — identity IGDB. Statuses `untriaged`, `never-played`,
`maybe-played` (the honest answer when ap can't remember), `playing`,
`played`, `finished`, `dropped`. Pictures: Steam cover (wide), IGDB poster
(tall). Unranked order: playtime. No description field: the log shows ap's
words, not a summary.

**movie** — identity TMDB film. Statuses `untriaged`, `want`, `watched`.
Pictures: poster (tall), backdrop (wide).

**series** — a TV series as a whole; identity TMDB series (`tmdb-tv`), which
TMDB numbers apart from films. Statuses `untriaged`, `want`, `watching`,
`watched`, `dropped`.

**artist** — identity Spotify artist. Statuses `listening`, `listened`.
Picture: square. Unranked order: listening.

**album** — identity Spotify album. Statuses `untriaged`, `want`,
`listened`. Cover: square. Inherits genres through `by`. Unranked order:
listening.

**track** — a recording, the same item on every release it appears on,
identity its ISRC, or its Spotify id when it has none. Remasters and live
versions have their own ISRCs and stay apart. Statuses `untriaged`,
`listened`.

## Sources

**Steam** — the library, nightly. Every owned game is logged. Playtime is
only ap's. Achievements are fetched once at first import and then only for
games whose playtime changed. Reviews come from the community reviews pages
as `steam-take`: recommendation (`recommended`, `not-recommended`), review,
date, and hours at review time, which are only ap's. Reviews posted this year
lack the year; it's taken as the current one. A review of a game not in the
library is a sync problem, never a new item.

**IGDB** — game identity. Matches Steam app ids through IGDB's external
games; that match is sure. Quick-add searches it, leaving out DLC, bundles,
packs, updates and editions; remakes, remasters and ports are separate
games. Gives title, release date, genres, developers, platforms, pictures.

**TMDB** — film and series identity. A Letterboxd film is matched by the id
in its diary entry, then by the TMDB link on its Letterboxd page (which also
tells a series from a film), then by title and year when exactly one film
has both. The rest are tried again each night.

**Letterboxd** — ap's films. Watched → `watched`; watchlist → `want` for a
film not yet watched: applying proposals. A watched film on the watchlist
stays `watched` and proposes `another-shot` instead, withdrawn after a
later diary watch or an export without it. Nightly syncs
see only page one of the watchlist; an uploaded export ZIP settles the rest.
Ratings, likes and reviews are `letterboxd-take`.

**Spotify** — music, nightly. Top artists and tracks per range per day,
plays, saves and unsaves are observations.
- Every artist ever in a top-artist list, and the main artist of every saved
  album, is logged. An artist is `listening` while it is in the 4-week list
  of the newest day any top-artist list was recorded, `listened` otherwise.
- Every saved album is logged, `want` until any of its tracks has been in a
  top-track list or played, then `listened`. Unsaving withdraws nothing.
- Listening scores each top-list appearance at rank r of N as
  (N − r + 1) / N, times 3 for all-time, 2 for six months, 1 for four
  weeks. An album sums its tracks' appearances plus 0.1 per play; an artist
  sums its own plus every album it is `by`. The stored value is the
  percentile among the kind's logged entries.
- Tracklists are fetched for logged albums only, in batches of 50 when out
  of step.
- The refresh token is a stored secret.

## Doing

**Triage** — one card at a time through suggesting proposals and untriaged
entries, for every kind, with bulk actions.

**Rating** — giving a shown entry its scores, one card at a time: pick a
division, then answer _better or worse than X?_ until it has a rank.

**Quick-add** — adding an item by searching its kind's catalogue (IGDB for
games, TMDB for films and series), or by title alone when there is nothing.

**Sync** — fetching fresh records from a source, nightly and on demand. Its
last **sync report** (state, counts, problems, last success) survives
restarts; admin warns when a sync failed or hasn't succeeded in two days.

**Stored secret** — a credential that rotates (the Spotify refresh token),
kept in the database and replaced from admin or by the sync, not set in env.
