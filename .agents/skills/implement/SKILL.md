---
name: implement
description: Implement scoped gopdsdk Go features, fixes, refactors, and tests under the repository architecture and portability contract.
---

# Implement

Follow `AGENTS.md`.

## Workflow

1. Inspect `git status`, relevant code, callers, tests, and package comment.
2. State the feature/package boundary; keep `cmd` compositional and extract to
   `internal/shared` only for a second real consumer.
3. Implement the smallest complete behavior and error paths.
4. Add deterministic tests; abstract I/O or processes only at test/platform
   boundaries.
5. Run proportional checks and inspect the diff for unrelated edits,
   portability, provenance, and overstated readiness.

Do not add dependencies, generated bindings, public API, or shared packages
speculatively.

## Checks

Use workspace-local `.cache` for `GOCACHE` and `GOMODCACHE` when needed. Run
proportionally; report skipped/failed capabilities and every earlier failure:

```powershell
gofmt -w cmd internal
go test ./...
go vet ./...
git diff --check
go run ./cmd/gopdsdk doctor
```

For a P0 release candidate, also run host external-consumer acceptance and
inspect the CI matrix. Run `doctor --probe` only with the official SDK installed;
otherwise report `CI-tested, SDK integration unverified`.
