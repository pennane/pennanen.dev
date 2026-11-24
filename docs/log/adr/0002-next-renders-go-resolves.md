# 2. Next renders, Go resolves

**Status:** accepted · 2026-11-15

## Context

pennanen.dev is a Next app on Vercel; the log's data lives on a small VPS
that may go down. Recording should be quick, often from a phone.

## Decision

- Admin lives at `/admin` in the Next app, installable as a PWA. Middleware
  guards it with a password and a signed HTTP-only cookie. Server actions
  call the VPS with a bearer token. The browser talks to the VPS only to
  upload images, with a short-lived ticket signed with the admin token,
  because the site's functions can't take a 10 MB body.
- Public pages `/log`, `/log/<kind>` and `/log/<kind>/<slug>` are ISR. They
  fetch unauthenticated public endpoints that return only what may show,
  already resolved for the page ([ADR 10](./0010-presentation-is-a-stored-projection.md)).
  They refresh once a day, when the VPS calls a Vercel revalidate hook with
  its own secret, or when ap presses publish in admin. Edits and syncs
  never refresh them: each refresh rebuilds every log page a visitor or
  crawler opens, and Vercel's free tier caps those writes.
- Pictures use `next/image` with `unoptimized` and the source's own sizes.

## Consequences

- Visitors never depend on the VPS being up; they get the last cached page.
- Private data never reaches Next through the public endpoints.
- Next holds no presentation rules: it lays out and renders what it is given.
- The VPS serves no frontend, only an API and images.
