---
name: release
description: Audit, prepare, document, or explicitly publish gopdsdk releases and post-tag verification.
---

# Release

Follow `AGENTS.md` and `RELEASING.md`. Do not publish without an explicit request.

## Workflow

1. Identify the intended version and release scope from `docs/ROADMAP.md`.
   Inspect status, the previous-tag diff, local/remote tags, and CI for the exact
   commit.
2. Run the `RELEASING.md` gates with workspace `.cache` for `GOCACHE` and
   `GOMODCACHE`. Run `doctor --probe` only with the official SDK and keep
   Simulator, device-build, USB, and physical evidence distinct. Accept explicit
   user-confirmed physical results; never invent measurements or interactions.
3. Synchronize the documents named by `AGENTS.md`, plus `MIGRATING.md` and
   `RELEASING.md` when applicable. Re-run proportional checks and inspect for
   unrelated edits, stale versions, and overstated claims.

## Publishing boundary

Commits, tags, pushes, pull requests, and hosted releases each require explicit
authorization; a documentation declaration authorizes none of them.

When authorized, verify the reviewed commit, create the exact stable tag, push
only requested refs, verify the hosted release, then run clean post-tag proxy and
external-consumer checks without `replace`. Claim module availability only after
they pass.

## Handoff

Report gates, evidence level, changed docs, tag/publication state, next action,
and every tooling failure that could affect interpretation.

