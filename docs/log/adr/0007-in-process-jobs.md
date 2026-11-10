# 7. Scheduled jobs run inside the service

**Status:** accepted · 2026-11-09

## Context

The VPS has no crontab and no `sqlite3` CLI. Steam sync and backups run nightly.

## Decision

The Go service runs its own tickers: nightly Steam sync (library each night;
achievements only for games whose playtime changed, plus once at first import)
and nightly backup (`VACUUM INTO` → R2). Admin has a "sync now" button.

## Consequences

- One unit to deploy and watch; nothing else to configure on the box.
- A crash at job time skips that night's run; acceptable, the next one catches up.
