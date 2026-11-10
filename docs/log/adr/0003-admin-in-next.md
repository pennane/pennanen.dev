# 3. Admin lives at /admin in the Next app

**Status:** accepted · 2026-11-09

## Context

Recording should be quick and fun, often from a phone. A separate admin SPA
served from the VPS was considered.

## Decision

`/admin` in the existing Next app on Vercel, installable as a PWA. Middleware
guards `/admin/*` with a password (env var) and a signed HTTP-only cookie.
Server actions call the VPS with a bearer token; the browser never talks to
the VPS API directly.

## Consequences

- Reuses the site's components, deploys and preview URLs.
- The VPS serves no frontend, only an API.
- The site is no longer purely static; that's accepted.
- No offline support in v1: optimistic UI with background retry instead.
