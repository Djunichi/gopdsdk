---
name: release
description: Prepare, verify, document, and optionally publish gopdsdk releases. Use for release readiness audits, release candidates, version documentation, acceptance evidence, tags, GitHub releases, and post-tag module-proxy verification.
---

# Release

Follow `AGENTS.md` and `RELEASING.md`. Keep readiness evidence precise and do
not publish unless the user explicitly requests the publishing action.

## Workflow

1. Identify the intended version and release scope from `docs/ROADMAP.md`.
   Inspect status, the diff from the previous tag, local and remote tags, and
   current CI for the exact commit.
2. Run the local gates with workspace `.cache` values for `GOCACHE` and
   `GOMODCACHE`: formatting, `go test ./...`, `go vet ./...`,
   `git diff --check`, and `gopdsdk doctor`.
3. Run `doctor --probe` only with the official SDK. Record Simulator, device
   build, USB, and physical-device results independently; never promote CI,
   cross-builds, Docker, or dry-runs to SDK or hardware evidence.
4. Confirm the release-specific Simulator, physical-device, performance,
   bounded-memory, soak, and requested post-run log gates. Treat user-reported
   physical results as evidence when the user explicitly confirms them; do not
   invent measurements or skipped interactions.
5. Update `CHANGELOG.md`, `API.md`, `MIGRATING.md`, `COMPATIBILITY.md`,
   `README.md`, and `RELEASING.md` consistently. When the release is declared,
   move durable evidence to the changelog or compatibility matrix and remove
   the completed scope row and milestone section from `docs/ROADMAP.md` so it
   contains only future work.
6. Re-run proportional checks and inspect the final diff for unrelated edits,
   stale version references, overstated evidence, and platform claims.

## Publishing boundary

Creating commits, tags, pushes, pull requests, or hosted releases requires an
explicit user request. A documentation-only declaration does not authorize any
of those actions; report their actual state separately.

After an explicit publish request, verify the reviewed commit, create the exact
stable tag, push only the requested refs, and verify the hosted release. Then
run the clean post-tag module-proxy and external-consumer checks without a
`replace` directive. Do not claim module availability before that succeeds.

## Handoff

Report completed gates, failed or skipped gates, exact evidence level, changed
documentation, tag and publication state, and the next required action. Name
every earlier sandbox or tooling failure that could affect interpretation.

