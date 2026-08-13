---
name: implement
description: Implement scoped gopdsdk Go changes and tests under its architecture and portability contract.
---

# Implement

Follow `AGENTS.md`.

## Workflow

1. Inspect status, relevant code, callers, tests, and the package comment.
2. State the feature/package boundary, then implement the smallest complete
   behavior and error paths.
3. Add deterministic tests; abstract I/O or processes only at test/platform
   boundaries.
4. Verify proportionally and inspect the diff for unrelated edits, portability,
   provenance, and overstated readiness.

Do not add dependencies, generated bindings, public API, or shared packages
speculatively.

## Checks

Use workspace `.cache` for `GOCACHE` and `GOMODCACHE`. Report all failures and
skipped capabilities:

```powershell
gofmt -w cmd internal
go test ./...
go vet ./...
git diff --check
go run ./cmd/gopdsdk doctor
```

Use `$release` for release-candidate or publishing work.
