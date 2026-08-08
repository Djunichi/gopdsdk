# Repository contract

Build `gopdsdk`, an independent cross-platform Go SDK for Playdate. Official SDK
headers, docs, examples, and observable tools are normative. pdgo is behavioral
reference only: never copy or mechanically translate its code, structure,
comments, tests, generated files, scripts, patches, or linker files.

Use `$implement` for code changes and `$review` for reviews. Keep changes scoped,
preserve unrelated user edits, and do not commit or publish unless requested.

## Architecture

- `cmd/<binary>`: input parsing, dependency composition, and exit codes only.
- `internal/features/<feature>`: feature-owned logic, models, and tests.
- `internal/shared/<component>`: only after two real feature consumers exist.
- Optional `playdate.Context` capabilities must exist in every native ABI
  context, forward through runtime `applicationContext`, and have regression
  coverage through `NewApplication` (ABI-only tests are insufficient).
- Keep project-owned Simulator and device bridge sources as package-owned
  `go:embed` assets. Update their deterministic ABI tests in the same change;
  continue resolving official headers, setup sources, and linker scripts from
  the installed Playdate SDK.

Prefer feature cohesion; never add `util`, `common`, `helpers`, or `misc`.
Default to the standard library; justify new dependencies.

## Go contract

- Put `// Package <name> ...` in the primary implementation file; do not add a
  comment-only `doc.go`.
- Accept `context.Context` at process and I/O boundaries.
- Represent commands and paths structurally; never build business-logic shell
  strings.
- Avoid global mutable configuration.
- Test platform-specific behavior separately for Windows, macOS, and Linux.
- Do not claim readiness from executable discovery alone; require the relevant
  probe.

Name the actual verification level: unit, external-consumer CLI, native CI, SDK
integration, or physical device. CI, cross-compilation, dry-runs, and Docker do
not prove SDK, Simulator, USB, or hardware readiness. Physical-device acceptance
requires a connected Playdate; check its post-run crashlog only when requested.
