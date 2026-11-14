# 21. No catch-all kind

**Status:** accepted · 2026-11-14 · replaces the `thing` part of [ADR 5](./0005-kinds-defined-in-code.md)

## Context

ADR 5 added `thing`, a kind with free-form fields and nothing checked, so an
item that fit no kind could still be logged. ap never used it, and films
would have landed there before `movie` existed, as items no sync could find.

## Decision

`thing` is gone, and with it a kind's option to accept undeclared fields.
Every field belongs to a kind that declares it. Something that fits no kind
waits until a kind is defined for it.

## Consequences

- Migration 13 deletes any `thing` items and their settings.
- Verdicts and overrides only ever hold declared fields.
- A new sort of item costs a kind definition first.
