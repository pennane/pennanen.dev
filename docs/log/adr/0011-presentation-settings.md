# 11. What shows, and how, is set in admin

**Status:** accepted · 2026-11-10 · refines [ADR 9](./0009-public-by-default.md)

## Context
ADR 9 fixed in code which statuses each kind shows, and the score axes showed
under their code names. ap wants to change these by mood: show
`never-played` some days, hide `maybe-played` on others, or show a single
never-played game. ap also wants to name the scores, show a word in place of
the number, and give tags titles apart from their ids.

## Decision
- Per kind, admin sets the **shown statuses**, an **axis label** per axis,
  the **score display** (`number`, `word` or `both`) and the **score words**:
  a list of `{from, word}` where a score takes the word with the highest
  `from` at or below it. Code holds only the defaults.
- Tags get **titles** in admin. The id (`another-shot`) stays in the data;
  public pages show the title ("worth another shot") when there is one.
- Settings are JSON rows in a `setting` table, keyed `kind:<name>` and
  `tag-titles`.
- **Visibility** has three values: `auto` (default; follows the shown
  statuses), `public` (always shown) and `private` (never shown). Existing
  `public` entries became `auto`. Dismissed entries never show, whatever
  their visibility.
- Public pages read the settings from `GET /public/presentation`.

## Consequences
- Axis keys (`enjoyment`, `quality`) stay as stored names; renaming one in
  admin changes only what pages say, so no score is migrated.
- Public pages refetch the log every five minutes, so a settings change shows
  within five minutes.
