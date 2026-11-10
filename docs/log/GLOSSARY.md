# things as ap saw them — glossary

The words used for the log, in code and in conversation. If a word here
disagrees with the code, one of them is wrong.

## Things

**Item** — something that can be seen: a game, a wine, a region, a person.
It has a kind and a title, and that's about it. Opinions live elsewhere.

**Kind** — what sort of item it is (`game`, `thing`, later `tabletop`, `wine`…).
A kind decides the item's fields, its statuses, and which source gives it
an identity. Kinds are defined in code; which axes a kind uses, its **title**
(`games`) and its **description** are set in admin. `thing` is the catch-all
kind: free-form fields, nothing checked, so an odd item never blocks you.

**Relation** — a named link between two items. _Margaux 2015 is a
`vintage-of` Château Margaux, which is `part-of` Margaux._ Follow the links
and the bottle is in Margaux.

## Takes

**Entry** — ap's take on one item. At most one per item. Holds a status,
scores, tags, a blurb, and a visibility.

**Status** — where ap is with the item, regardless of what ap thought of it.
Each kind has its own set. Every kind starts at `untriaged` and can end at
`dismissed`. For `game`: `untriaged`, `never-played`, `maybe-played`,
`playing`, `played`, `finished`, `dropped`, `dismissed`. `maybe-played` is the
honest answer when ap can't remember. Each kind has **shown statuses**, set
in admin: an `auto` entry shows publicly when its status is one of them.

**Dismissed** — the status for something that doesn't belong in the log at all,
like Wallpaper Engine in a Steam library. The item is kept so sync won't bring
it back, but it never shows on public pages.

**Axis** — one dimension something can be scored on, such as `prestige` and
`quality`. Axes are their own thing: made, named and ordered in admin, and
shared by every kind that uses them. Each kind picks its axes; a kind with no
axes can't be scored. An axis has a key, set once, and a name, which ap can
change.

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

**Tag** — a free-form label ap puts on an entry: `favorite`, `comfort`,
`formative`. Nest with a slash for precision: `favorite/all-time`. Tags are
ap's, so they live on entries; what an item _is_ (`co-op`, `jam-game`) is a
claim, not a tag. `another-shot` marks something ap wants to try again,
whatever its status. `idled` marks a game ap ran for trading cards, whether or
not ap also played it. A tag's **title** is what public pages show for it;
the tag itself is its id.

**Time spent** — ap's word on how long it took: `not-enough`, `a-bit`,
`just-right`, `a-lot`, `too-much`. A judgment, not a measurement: Steam's
playtime stays private, and public pages show only this, when ap says it.

**Blurb** — public text, written for others, shown on the item's page.

**Prose** — what ap writes about an entry: one formatted document per entry,
with images and links to other items, shown on the entry's public page
exactly when the entry is. Pages call it _visions_; each passage in it is a
vision, such as a dated paragraph _2024-03 replayed, still great_.

## Sources

**Source** — somewhere item data comes from: Steam, IGDB, or ap by hand.

**Claim** — what one source says about an item, kept exactly as received.
Never edited; a re-sync replaces only that source's claim.

**Verdict** — the claim that wins for a field. Sources are ranked per field;
the highest-ranked one with a value wins. ap's own override ranks first,
always. _Steam claims 412h, IGDB claims nothing, so the verdict is 412h._

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

**Rating** — giving a shown entry its scores, one card at a time, after triage:
pick a division, then answer _better or worse than X?_ until it has a rank.

**Hint** — the importer's guess during triage (`likely-idled`,
`likely-played`, `likely-never-played`, `likely-not-a-game`). Hints suggest;
they never set a status.

**Sync** — pulling fresh claims from a source. Steam syncs nightly and on demand.

**Visibility** — whether an entry shows publicly: `auto` (default, by its
status), `public` (always) or `private` (never). Prose has no visibility of
its own; it follows its entry.

**Tier list** — a kind's shown entries laid out by division on one axis,
best division first; entries with no score on that axis are left off.
