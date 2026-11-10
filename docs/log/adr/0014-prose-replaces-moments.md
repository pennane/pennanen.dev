# 14. Prose replaces moments

**Status:** accepted · 2026-11-10 · amends [ADR 1](./0001-sqlite-on-the-vps.md),
[ADR 5](./0005-kinds-defined-in-code.md) and [ADR 6](./0006-isr-with-revalidate.md)

## Context

Moments were dated plain-text notes, one record each, with their own
visibility. Writing about a game felt like filling in a form, there was no
room for images or links, and the private ones were never used.

## Decision

- An entry has one **prose** document, stored as Markdown. Pages call it
  _visions_. A date is something written inside it, not a field.
- Markdown is the stored form whatever the editor. The admin edits it in a
  WYSIWYG editor; swapping the editor never migrates data.
- Prose has no visibility of its own. It is public exactly when its entry is.
- Images are uploaded to the log service, stored next to the database and
  named by a hash of their content. Each is copied to R2 at upload, so the
  nightly database backup and the image copies together restore everything.
  Prose refers to them as `/log/images/<name>`, which the site serves from
  the log service, so stored text never names a host.
- The browser uploads images straight to the log service with a short-lived
  ticket signed with the admin token, because the site's functions cannot
  take a 10 MB body.
- Links to other items are ordinary Markdown links to their public pages.
- Moments that showed publicly become dated paragraphs of prose, newest
  first. Every moment and its references stay in archive tables.

## Consequences

- Public pages render Markdown on the server and ignore raw HTML in it.
- "Mentioned in" backlinks can come later by reading links out of prose.
- Unused images are never deleted for now.
