# 26. Lists read claim readings

**Status:** accepted · 2026-11-15 · follows [ADR 25](./0025-steam-reviews-are-a-take.md)

## Context

Every list request decided its verdicts by parsing the whole source record of
every claim of every item on the page, about 33 MB of JSON. On the 1 vCPU,
458 MB VPS that takes three to six seconds a request, and Vercel gives up on
a page rebuild after ten.

Raising the VPS's costs was rejected, and so was dropping stored data: the
full record is what a later reader change replays from.

## Decision

Each claim has a **reading**, what its source's reader takes from it, stored
in `claim_reading` beside the claim and tagged with `verdict.ReadersVersion`.
`putClaim` writes the reading with the claim, on the same transaction. A claim
the reader rejects, or whose source has no reader, has no reading.

Lists and verdicts, takes and read scores are decided from readings. Values
are normalized by field type when a reading is loaded, so a reading decides
exactly as the claim does. The single-item view still loads whole claims.

At start, before serving, claims with no reading or an older version are
re-read in batches of 200, one transaction per batch.

The claim itself is never altered, trimmed or dropped.

## Consequences

- A change to any reader needs `ReadersVersion` bumped, or old readings stay.
- The database grows by the size of the readings, a small fraction of the
  records they come from.
- A claim no reader can read is looked at again on every start, as it never
  gets a reading to mark it done.
- A list no longer touches `claim.data`; the claim's tail pages stay unread.
