# things as ap saw them — glossary

The words used for the log, in code and in conversation. If a word here
disagrees with the code, one of them is wrong.

## Things

**Item** — something that can be seen: a game, a wine, a region, a person.
It has a kind and a title, and that's about it. Opinions live elsewhere.

**Kind** — what sort of item it is (`game`, `movie`, `series`, later `album`, `wine`…).
A kind decides the item's fields, its statuses, and which source gives it
an identity. Kinds are defined in code; which axes a kind uses, its **title**
(`games`) and its **description** are set in admin. A field is for everyone
or **only ap** (Steam playtime); only ap's never leave the log. A field's
type decides how public pages use it, unless admin says otherwise. There is no catch-all
kind: something that fits no kind waits until it gets one.
A `wine` is what every year of it shares; a `vintage` is one bottling of a
wine, of a single year or **non-vintage** (blended across years, like most
Champagne). Every bottle ap drinks is a vintage, so all bottles rank
together; a wine itself is only ever logged as `want`.
A `series` is a TV series as a whole; seasons and episodes aren't items.
Letterboxd calls series films too, so a film found to be a series **moves**
to `series`, keeping its entry and going to the bottom of its division.

**Relation** — a named link between two items. _Margaux 2015 is a
`vintage-of` Château Margaux, which is `part-of` Margaux._ Follow the links
and the bottle is in Margaux.
An album is `by` its **main artist**, the first one credited, and `with` every
other credited artist. Spotify's "Various Artists" is never a main artist.

## Takes

**Entry** — ap's take on one item. At most one per item. Holds a status,
scores, tags, a blurb, and a visibility.

**Logged / known** — an item with an entry is logged: it's in ap's log. An
item with no entry is only known: a sync came across it (an album in a top
list, the wine behind a vintage) but nobody put it in the log. Not the same
as `dismissed`, which is logged and ruled out.

**Status** — where ap is with the item, regardless of what ap thought of it.
Each kind has its own set. Every kind starts at `untriaged` and can end at
`dismissed`. For `game`: `untriaged`, `never-played`, `maybe-played`,
`playing`, `played`, `finished`, `dropped`, `dismissed`. `maybe-played` is the
honest answer when ap can't remember. Each kind has **shown statuses**, set
in admin: an `auto` entry shows publicly when its status is one of them.
`want` is the status for something ap means to get to, from a watchlist,
to-read shelf or wishlist; it isn't shown by default.
A status is **set by** ap or by a source's sync. A sync changes only a
status it set itself; once ap sets one, no sync touches it again, except for
a **computed status**, where only `dismissed` sticks.

**Computed status** — a status a source works out instead of ap choosing it.
Artists have one: the source `spotify` sets it on every sync. Only
`dismissed` sticks; any other status ap picks goes back to the source, which
sets it again on the next sync.

A saved Spotify album is logged `want` until one of its tracks shows up in a
top list or a play, then `listened`. That isn't a computed status: once ap
picks one, it sticks, and unsaving the album changes nothing.

**Listening / listened** — an artist's computed statuses. `listening` means
being in the 4-week top list of the newest day any top-artist list was
recorded; `listened` is everything else that was ever in a top list, and the
main artist of every saved album.

**Dismissed** — the status for something that doesn't belong in the log at all,
like Wallpaper Engine in a Steam library. The item is kept so sync won't bring
it back, but it never shows on public pages.

**Axis** — one dimension something can be scored on, such as `prestige` and
`quality`. Axes are their own thing: made, named and ordered in admin, and
shared by every kind that uses them. Each kind picks its axes; a kind with no
axes can't be scored. An axis has a key, set once, and a name, which ap can
change. Axes come in four sorts. **Own** axes are ap's, and only ap
sets them. An **outside** axis is ap's rating on another site
(`letterboxd`, `vivino`). A **crowd** axis is a site's public average.
An **activity** axis is a number about what ap did: playtime, achievements
unlocked, listening. Outside, crowd and activity axes are read from claims, never set by hand and never
ranked. Each has a **reading**: the source and field it reads, and a
**floor** per division, best first, the lowest value that lands there. An
outside axis reads a number from a take, a crowd axis one from a describing
source, an activity axis one from a field about ap; admin offers only the
pairs some kind has. An outside axis suggests a division when ap rates on an own axis; a
crowd axis only shows.

**Prestige** — the axis that orders things: how high something stands for
ap, all told, from how much it gave to how it stuck. Not quality: a
masterpiece can bore, trash can delight. Prestige is the one axis with ranks,
and every kind that ranks uses it.

**Quality** — the axis for how well made something is, whatever ap thought
of it.

**Score** — the division an entry sits in on one axis. Always optional;
nothing is owed.

**Division** — a named slice of an axis, in order from best. An axis starts
with tiers: `S`, `A`, `B`, `C`, `D`, `F`. Each axis has its own divisions, as
many as it needs, renamed or reordered in admin. There is no number behind a
division.

**Rank** — an entry's place in its kind's prestige order, counted from the
top across all divisions. Strict: no two entries share a rank. Only prestige
has ranks; other axes have divisions alone. Games rank against games, never
against wine. Public pages never show the number; they order by it.

**Unranked order** — the field a kind orders its unranked entries by, highest
first, after every ranked one: playtime for games, listening for artists and
albums. Public pages get each entry's place in it, never the number.

**Listening** — how much ap listens to an artist or album, from 0 to 100: its
percentile among its kind's logged entries. Worked out nightly from every
top-list day and play. Public: it can filter, sort and show like any number.

**Tag** — a free-form label ap puts on an entry: `favorite`, `comfort`,
`formative`. Nest with a slash for precision: `favorite/all-time`. Tags are
ap's, so they live on entries; what an item _is_ (`co-op`, `jam-game`) is a
claim, not a tag. `another-shot` marks something ap wants to try again,
whatever its status. `idled` marks a game ap ran for trading cards, whether or
not ap also played it. A tag's **title** is what public pages show for it;
the tag itself is its id.
A tag is **added by** ap or by a source, like a status. A sync adds and
removes only the tags it added itself; once ap touches one, it's ap's. _A
watched film back on the Letterboxd watchlist gets `another-shot` from
Letterboxd, which takes it off again after the rewatch._

**Time spent** — ap's word on how long it took: `not-enough`, `a-bit`,
`just-right`, `a-lot`, `too-much`. A judgment, not a measurement: Steam's
playtime stays private, and public pages show only this, when ap says it.

**Blurb** — public text, written for others, shown on the item's page.

**Prose** — what ap writes about an entry: one formatted document per entry,
with images and links to other items, shown on the entry's public page
exactly when the entry is. Pages call it _visions_; each passage in it is a
vision, such as a dated paragraph _2024-03 replayed, still great_.

**Mention** — a link to another item written in prose. _Realm of the Mad God
goes with Pendulum._ Mentions are ap's links; relations are what sources say.
Together they make the graph of items.

## Sources

**Source** — somewhere item data comes from: Steam, IGDB, or ap by hand.

**Claim** — what one source says about an item, kept exactly as received.
Never edited; a re-sync replaces only that source's claim.

**Reading** — what a source's reader takes from one claim. Stored with the
claim and recomputed when the readers change. Verdicts are decided from
readings, not from claims.

**Outside take** — ap's take recorded on another site: Letterboxd stars, a
Goodreads shelf and review, a Vivino rating and tasting note. Kept as claims
from its own source (`letterboxd-take`, `steam-take`, `vivino-take`), apart from what the
site says about the thing itself. Shown alongside the entry, never merged into
it. Private parts, like the price paid, show only in admin.

**Observation** — something a source said at one moment, kept for good:
_Pendulum in the 4-week top 5 on 2026-03-14_, _watched Heat again on
2026-02-01_. Claims say how things are now; observations say what happened
when.

**Verdict** — the claim that wins for a field. Sources are ranked per field;
the highest-ranked one with a value wins. ap's own override ranks first,
always. _Steam claims 412h, IGDB claims nothing, so the verdict is 412h._
Some fields are **inherited**: when no source of the item itself has a
value, the verdict comes from the item it's related to, one step away. _A
vintage with no grapes of its own shows its wine's grapes._

**Identity** — the external ID that says two claims are about the same item.
For games, the IGDB ID; for films and series, the TMDB ID, which TMDB numbers
separately for each. Optional: an item with no identity is just ap's.

**Match** — giving an item its identity. Steam games match by app id on their
own; the rest wait in the match queue until ap picks a game or answers
**no match**, which IGDB's claim then records so nobody asks again.
Films match TMDB: by the id Letterboxd's diary gives, by the TMDB link on
the film's Letterboxd page, which also tells a series from a film, or by
searching title and year when exactly one film has both; the rest stay
unmatched and are tried again each night. ap can correct a match with a TMDB
link or id, or answer no match, which TMDB's claim then records.

**Quick-add** — adding an item by searching its kind's catalogue (IGDB for
games, TMDB for films and series), or by title alone when it has nothing.

**Cover / poster** — an item's wide picture and its tall box art.

## Doing

**Triage** — moving an entry out of `untriaged`. The fast, one-at-a-time
screen where imports get sorted into played, never played, and so on.

**Rating** — giving a shown entry its scores, one card at a time, after triage:
pick a division, then answer _better or worse than X?_ until it has a rank.

**Hint** — the importer's guess during triage (`likely-idled`,
`likely-played`, `likely-never-played`, `likely-not-a-game`). Hints suggest;
they never set a status.

**Suggestion** — a known item that ap might want to log, such as a saved
album, or an album by an artist ap has logged. Suggestions wait in admin
until ap looks; they never fill a queue. Dismissing one logs it as
`dismissed`, so it isn't suggested again.

**Sync** — pulling fresh claims from a source. Every sync runs nightly and on
demand: Steam, IGDB matching, and each new source. Its last **sync report**
(state, counts, problems, when it last succeeded) survives restarts; admin
warns when a sync failed or hasn't succeeded in two days.

**Stored secret** — a credential that rotates (the Spotify refresh token), kept in the database and replaced from admin or by the sync, not set in env.

**Visibility** — whether an entry shows publicly: `auto` (default, by its
status), `public` (always) or `private` (never). Prose has no visibility of
its own; it follows its entry.

**Tier list** — a kind's shown entries laid out by division on one axis,
best division first; entries with no score on that axis are left off. Admin
picks which axes get one.

**Overview** — the log's front page: a few shown entries per kind, picked by
the kind's overview sort, set in admin.
