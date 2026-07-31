# Repository agent guide

## Mission and sources

Build an independent, cross-platform Go SDK for Playdate. Official Playdate SDK
headers, documentation, examples, and observable tool behavior are normative.
pdgo is only a product and behavioral reference; do not copy or mechanically
translate its code, layout, comments, tests, generated bindings, scripts,
patches, or linker files.

## Workflow

Use `$implement` for code changes and `$review` for review tasks. For an
implementation:

1. Inspect the relevant code, repository status, and this guide.
2. State the intended package and feature boundary before editing.
3. Implement the smallest complete vertical slice.
4. Add or update deterministic tests.
5. Review the diff for correctness, architecture, portability, and accidental
   third-party transfer.
6. Run the required checks and report every capability left unverified.

Do not mix unrelated cleanup into feature work. Preserve user changes in a dirty
worktree. Do not publish or commit unless explicitly requested.

## Verification levels

- Unit tests are deterministic and require neither the Playdate SDK nor a
  connected device.
- CLI acceptance tests build `gopdsdk`, create a separate Go module, compile
  that consumer, and inspect Simulator/device dry-run plans.
- Native CI runs unit and CLI acceptance tests on Windows, macOS, and Linux.
- SDK integration requires the official SDK and executes `doctor --probe` or a
  target-specific probe.
- Hardware acceptance requires a connected Playdate and a post-run crashlog
  check. CI, cross-compilation, and Docker do not substitute for it.

Never promote native CI evidence to Simulator, SDK-tool, or device readiness.
Docker may add reproducible Linux coverage, but it does not verify Windows,
macOS, GUI Simulator behavior, or USB deployment.

## Project layout

```text
cmd/<binary>/                 composition roots only
internal/features/<feature>/ feature-specific logic, models, and tests
internal/shared/<component>/ code with at least two real feature consumers
.agents/skills/              repository workflows for Codex
.vscode/launch.json          shared debug configurations
docs/                        internal P0 notes (gitignored)
```

Prefer feature cohesion over technical layers. Keep code in its feature until a
second consumer proves it is shared. Never create generic `util`, `common`,
`helpers`, or `misc` packages.

`cmd` packages parse process inputs, compose dependencies, and select exit
codes; they do not contain business logic.

## Go conventions

- Every package has a comment beginning exactly with `// Package <name>` and a
  short purpose statement in its primary implementation file. Do not create a
  separate `doc.go` solely for the package comment.
- Keep package names short, lowercase, and singular where natural.
- Use the standard library by default; record the reason for each dependency.
- Accept `context.Context` at boundaries that may execute processes or perform
  I/O.
- Represent commands and paths structurally; do not assemble shell strings in
  business logic.
- Do not use global mutable configuration.
- Do not mark a capability ready merely because executables exist; require the
  relevant probe.
- Test Windows, macOS, and Linux behavior independently.
- Use `gopdsdk` as the canonical project and binary name.

## Required checks

When required, point `GOCACHE` and `GOMODCACHE` at `.cache` in the workspace.
Run:

```powershell
gofmt -w cmd internal
go test ./...
go vet ./...
git diff --check
go run ./cmd/gopdsdk doctor
```

Before a P0 release candidate, also run the external-consumer acceptance test
on the current host and inspect the GitHub Actions matrix. Run
`doctor --probe` only where the official SDK/toolchains are installed; report
other hosts as `CI-tested, SDK integration unverified`.

Never hide an earlier failed command behind a later successful command.
