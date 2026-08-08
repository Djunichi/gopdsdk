# Contributing to gopdsdk

gopdsdk is an independent implementation. Derive behavior from official
Playdate SDK headers, documentation, examples, and observed official-tool
behavior. Do not copy or mechanically translate pdgo or other third-party code,
tests, generated bindings, scripts, comments, layout, patches, or linker files.

Keep changes feature-cohesive under `internal/features/<feature>`. Move code to
`internal/shared/<component>` only after two real consumers exist. The
`cmd/gopdsdk` package is a composition root, and public `playdate` API additions
require a concrete example.

Importable application packages keep `pdxinfo` beside their Go source and put
all packageable assets below `resources/`. The contents of that directory are
staged at the PDX root; for example, `resources/images/player.png` is loaded as
`images/player`. Files outside `resources/` must never be copied implicitly.

Every package needs a `// Package <name>` comment in its primary implementation
file; do not add a comment-only `doc.go`. Prefer the standard library and record
the reason for every dependency.

Project-owned Simulator and device bridge sources are package-owned `go:embed`
assets. Keep changes to those assets together with their deterministic ABI
tests. Official Playdate headers, setup sources, and linker scripts must remain
external inputs resolved from the installed SDK.

Before opening a pull request, run:

```sh
gofmt -w cmd examples internal playdate
go test ./...
go vet ./...
git diff --check
go run ./cmd/gopdsdk doctor
```

Describe the official SDK version/source used, the independently stated
requirement, and the verification level reached. Distinguish unit/CI evidence
from official SDK, Simulator, and physical-device evidence. Do not include
generated build artifacts or internal `docs/` notes in a pull request.

Public API and compatibility changes must update [API.md](API.md) and
[COMPATIBILITY.md](COMPATIBILITY.md). Release candidates follow
[RELEASING.md](RELEASING.md); tagging and publishing require explicit approval.
