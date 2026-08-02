# Changelog

All notable release changes are documented here. The module remains pre-v1;
minor releases may intentionally change the public API when release notes call
that out.

## v0.3.0 - unreleased

### Added

- Optional immediate-mode line, rectangle, ellipse, and triangle
  drawing on Simulator and device adapters.
- Value-owned solid, XOR, and 8x8 pattern paints, plus clipping, draw offsets,
  and bitmap draw modes through narrow graphics capability interfaces.
- Portable sentinel errors for invalid primitive geometry, colors, and draw
  modes.
- A deterministic `examples/primitives` acceptance scene covering every P3.1
  drawing and state operation.

### Compatibility

- Windows SDK 3.1.1 Simulator and physical-device visual acceptance passed for
  P3.1 with matching output. Hard-float build and USB deployment passed; soak
  and memory-growth evidence remain unverified.

## v0.2.0 - 2026-08-01

### Added

- Explicitly owned sprites with display-list membership, positioning,
  visibility, z-index, and deterministic cleanup.
- Collision rectangles, slide/freeze/overlap/bounce responses, resolved
  movement, and sprite point/rectangle/overlap queries.
- Owned bitmap tables, borrowed frames, and an allocation-free animation helper
  supporting delta time, fixed frames, and pause/resume.
- Explicitly owned short sound effects and streaming file players with volume,
  playback state, lifecycle pause/resume, rollback, and cleanup.
- Custom-font loading, native measurement, selected-font drawing, and a
  deterministic game-UI example.
- A snapshot test that makes every exported `playdate` API change explicit.

### Changed

- `playdate.Context` now composes sprite and audio capabilities in addition to
  the P1 system, input, and graphics capabilities. This is an intentional
  pre-v1 public API expansion.
- Simulator and device runtime adapters now expose the P2 graphics, sprite,
  collision, audio, and font slices.

### Compatibility

- The verified toolchain remains Go 1.26.5, Playdate SDK 3.1.1, TinyGo 0.41.1,
  and Arm GNU Toolchain GCC 15.3.1.
- Windows remains the only host with verified official SDK and physical-device
  execution. See [COMPATIBILITY.md](COMPATIBILITY.md) for per-feature evidence
  and the P2 acceptance work that remains unverified.

## v0.1.1

- Corrected release CI and version-aware external-consumer handling after the
  first public release.

## v0.1.0

- First public release, covering the P0 foundation and P1.0 through P1.4.
