# things as ap saw them — glossary

The words used for the log, in code and in conversation. If a word here
disagrees with the code, one of them is wrong.

## Things

**Item** — something that can be seen: a game, a wine, a region, a person.
It has a kind and a title, and that's about it. Opinions live elsewhere.

**Kind** — what sort of item it is (`game`, `thing`, later `tabletop`, `wine`…).
A kind decides the item's fields, its axes, its statuses, and which source
gives it an identity. Kinds are defined in code. `thing` is the catch-all
kind: free-form fields, nothing checked, so an odd item never blocks you.

**Relation** — a named link between two items. *Margaux 2015 is a
`vintage-of` Château Margaux, which is `part-of` Margaux.* Follow the links
and the bottle is in Margaux.

## Takes

**Entry** — ap's take on one item. At most one per item. Holds a status,
scores, tags, notes, a blurb, and a visibility.

**Status** — where ap is with the item, regardless of what ap thought of it.
Each kind has its own set. Every kind starts at `untriaged` and can end at
`dismissed`. For `game`: `untriaged`, `never-played`, `maybe-played`,
`playing`, `played`, `finished`, `dropped`, `dismissed`. `maybe-played` is the
honest answer when ap can't remember. Each kind also lists the statuses it
**shows** publicly; for `game` those are the ones where ap played it.

**Dismissed** — the status for something that doesn't belong in the log at all,
like Wallpaper Engine in a Steam library. The item is kept so sync won't bring
it back, but it never shows on public pages.

**Axis** — one dimension something can be scored on: `enjoyment` (did I like
it), `quality` (is it well made). They are independent: a masterpiece can
bore, trash can delight. Each kind lists its axes; a kind with no axes can't
be scored.

**Score** — a number from 0 to 100 on one axis. Always optional; nothing is owed.

**Tag** — a free-form label ap puts on an entry: `favorite`, `comfort`,
`formative`. Nest with a slash for precision: `favorite/all-time`. Tags are
ap's, so they live on entries; what an item *is* (`co-op`, `jam-game`) is a
claim, not a tag. `another-shot` marks something ap wants to try again,
whatever its status. `idled` marks a game ap ran for trading cards, whether or
not ap also played it.

**Ranking** — ap's hand-picked, ordered list of items: "games I want to single
out". Any kinds may mix. Always public, at `/log/ranking/<slug>`. Separate from
scores; neither changes the other.

**Tier** — a named group inside a ranking (`S`, `all-time`). A ranking is
either `numbered` (one plain order) or `tiered` (tiers in order, each holding
its own order).

**Placement** — one item's spot in a ranking, with an optional one-line note
about why it sits there.

**Notes** — private markdown. Only ap sees it.

**Blurb** — public text, written for others, shown on the item's page.

**Moment** — a dated note on an entry: *replayed 2024, still great*. May
carry a score (which becomes the entry's score unless overridden) and may
point at other items.

## Sources

**Source** — somewhere item data comes from: Steam, IGDB, or ap by hand.

**Claim** — what one source says about an item, kept exactly as received.
Never edited; a re-sync replaces only that source's claim.

**Verdict** — the claim that wins for a field. Sources are ranked per field;
the highest-ranked one with a value wins. ap's own override ranks first,
always. *Steam claims 412h, IGDB claims nothing, so the verdict is 412h.*

**Identity** — the external ID that says two claims are about the same item.
For games, the IGDB ID. Optional: an item with no identity is just ap's.

**Match** — giving an item its identity. Steam games match by app id on their
own; the rest wait in the match queue until ap picks a game or answers
**no match**, which IGDB's claim then records so nobody asks again.

**Quick-add** — adding an item by searching IGDB, or by title alone when IGDB
has nothing.

**Cover / poster** — an item's wide picture and its tall box art.

## Doing

**Triage** — moving an entry out of `untriaged`. The fast, one-at-a-time
screen where imports get sorted into played, never played, and so on.

**Rating** — giving a shown entry its scores, one card at a time, after triage.

**Hint** — the importer's guess during triage (`likely-idled`,
`likely-played`, `likely-never-played`, `likely-not-a-game`). Hints suggest;
they never set a status.

**Sync** — pulling fresh claims from a source. Steam syncs nightly and on demand.

**Visibility** — who can see an entry: `public` (default) or `private`;
later maybe `unlisted` (reachable by link, not listed). A moment is `inherit`
(default, same as its entry) or `private`, so it can never be more visible
than its entry.
