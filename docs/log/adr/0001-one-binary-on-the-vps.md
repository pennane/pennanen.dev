# 1. One Go binary on the VPS

**Status:** accepted · 2026-11-15

## Context

The log lives on the $5 VPS: 1 vCPU and 458 MB RAM, already dipping into swap
with five Node services. Nothing may be built on the box, and it has no
crontab and no `sqlite3` CLI. Cost should stay at zero. Pages are rebuilt on
Vercel, which gives up on a rebuild after ten seconds.

## Decision

- One Go service in `apps/log`, using `modernc.org/sqlite` (pure Go, no cgo),
  owning one SQLite file. GitHub Actions cross-compiles it: `scp` →
  `/opt/log` → `log.service` (systemd, `EnvironmentFile=/opt/log/.env`).
  Caddy routes `api.log.pennanen.dev` → `localhost:3005`.
- Jobs run on the service's own tickers: every sync nightly, and a nightly
  `VACUUM INTO` snapshot uploaded to Cloudflare R2. Admin can run any of
  them now.
- **Performance budget**: a request reads stored tables and never derives
  anything. A public request stays well under a second on the box.
  Derivation happens when records or takes change, not when pages are read
  ([ADR 3](./0003-the-layered-model.md)).

## Consequences

- About 15 MB resident, one static binary, one unit to deploy and watch.
- A crash at job time skips that night's run; the next one catches up.
- Single writer, single box. Fine for one user.
- Backups hold records, takes and images
  ([ADR 6](./0006-takes-are-irreplaceable.md)), because they are all that
  can't be derived again.
