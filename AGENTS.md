# Repository contract

Build `gopdsdk`, an independent cross-platform Go SDK for Playdate. Official
Playdate SDK headers, documentation, examples, and observable tool behavior are
normative. pdgo is behavioral reference only: never copy or mechanically
translate its code, structure, comments, tests, generated files, scripts,
patches, or linker files.

Use `$implement` for code changes and `$review` for reviews. Keep changes scoped,
preserve unrelated user edits, and do not commit or publish unless requested.

## Architecture

- `cmd/<binary>`: input parsing, dependency composition, and exit codes only.
- `internal/features/<feature>`: feature-owned logic, models, and tests.
- `internal/shared/<component>`: only after two real feature consumers exist.
- When adding an optional `playdate.Context` capability, implement it in each
  native ABI context and forward it through runtime `applicationContext`.
  Regression-test the capability assertion and call path through
  `NewApplication`; testing the ABI context alone is insufficient.

Prefer feature cohesion; never add generic `util`, `common`, `helpers`, or
`misc` packages. Use the standard library by default and document why any new
dependency is needed.

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

Verification claims must name their actual level: unit, external-consumer CLI,
native CI, SDK integration, or physical device. CI, cross-compilation, dry-runs,
and Docker do not prove SDK, Simulator, USB, or hardware readiness.
Physical-device acceptance requires a connected Playdate. Check the post-run
crashlog only when the user explicitly requests it.
