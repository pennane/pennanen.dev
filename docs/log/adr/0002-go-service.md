# 2. The log service is written in Go

**Status:** accepted · 2026-11-09

## Context

The VPS has 1 vCPU and 458 MB RAM, already dipping into swap with five Node
services. Nothing may be built on the box.

## Decision

A Go service in `log/` in this repo, using `modernc.org/sqlite` (pure Go, no
cgo). GitHub Actions cross-compiles it and deploys like the other services:
`scp` → `/opt/log` → `log.service` (systemd, `EnvironmentFile=/opt/log/.env`).
Caddy routes `api.log.pennanen.dev` → `localhost:3005`; DNS A record in Netlify.

## Consequences

- ~15 MB resident instead of ~60+ MB for another Node process.
- One static binary; no runtime to install.
- A second language in the repo. Kind definitions live in Go only (see 0005)
  so nothing is duplicated in TypeScript.
