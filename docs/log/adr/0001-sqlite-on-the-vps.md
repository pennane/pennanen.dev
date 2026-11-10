# 1. SQLite on the VPS

**Status:** accepted · 2026-11-09 · prose: [ADR 14](./0014-prose-replaces-moments.md)

## Context

The log holds ap's entries, scores and notes, plus raw claims from sources.
Triage means hundreds of tiny writes. Backends already live on the $5 VPS;
cost should stay at zero.

## Decision

One SQLite file on the VPS, owned by the log service. Nightly snapshot via
`VACUUM INTO`, uploaded to Cloudflare R2 (free tier).

## Consequences

- No hosted database, no extra bill.
- Git-as-database was rejected: every triage tap would be a commit.
- Only entries, scores, notes and moments are irreplaceable; claims can be
  re-synced. Backups exist for the former.
- Single writer, single box. Fine for one user.
