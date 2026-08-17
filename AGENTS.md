# Repository contract

Build `gopdsdk`, an independent cross-platform Go SDK for Playdate. Official SDK
headers, docs, examples, and observable tools are normative. pdgo is behavioral
reference only: never copy or mechanically translate its code, structure,
comments, tests, generated files, or tooling artifacts.

Use `$implement` for code changes and `$review` for reviews. Keep changes scoped,
preserve unrelated user edits, and do not commit or publish unless requested.

## Code

- Keep `cmd/<binary>` to parsing, dependency wiring, and exit codes. Put owned
  logic, models, and tests in `internal/features/<feature>`; extract to
  `internal/shared/<component>` only for a second real consumer. Never add
  `util`, `common`, `helpers`, or `misc` packages.
- Keep native public APIs in `playdate`, with declarations and sentinel errors
  grouped by feature-owned files. Add a subpackage only for a distinct
  higher-level layer such as `playdate/store`.
- Optional `playdate.Context` capabilities must exist in every native ABI
  context, forward through runtime `applicationContext`, and have regression
  coverage through `NewApplication`, not only ABI tests.
- Keep owned Simulator/device bridge sources as package-owned `go:embed` assets
  and update deterministic ABI tests with them. Resolve official headers, setup
  sources, and linker scripts from the installed SDK.
- Put the package comment in the primary implementation file, not a comment-only
  `doc.go`. Accept `context.Context` at process/I/O boundaries; represent commands
  and paths structurally; never build business-logic shell strings; avoid global
  mutable configuration.
- Default to the standard library and justify dependencies. Test platform logic
  separately for Windows, macOS, and Linux. Readiness requires the relevant probe,
  not executable discovery.

## Evidence

Name the actual verification level: unit, external-consumer CLI, native CI, SDK
integration, or physical device. CI, cross-compilation, dry-runs, and Docker do
not prove SDK, Simulator, USB, or hardware readiness. Physical-device acceptance
requires a connected Playdate; inspect its post-run logs only when requested.

## Documentation

Keep `README.md` product/example-oriented rather than an implementation/evidence
log, `API.md` for public contracts, `CHANGELOG.md` for unreleased changes and
evidence, `docs/ROADMAP.md` for active scope, and `COMPATIBILITY.md` for released
compatibility rather than unreleased status. On release, move durable evidence
to the changelog or compatibility matrix and remove the completed scope from the
roadmap.
