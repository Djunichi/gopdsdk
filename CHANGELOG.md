# Changelog

All notable release changes are documented here. The module remains pre-v1;
minor releases may intentionally change the public API when release notes call
that out.

## v0.3.0 - unreleased

### Added

- An optional `playdate.Launcher` capability that forwards the official
  `exitToLauncher` operation through Simulator and device adapters; Playdate
  delivers normal termination cleanup before returning to the Launcher.
- A deterministic `examples/navigation` scene covering `Play`, an in-process
  return from gameplay to the main menu, and `Exit` through `Launcher`.
- Documented launcher artwork packaging through `pdxinfo.imagePath` and the
  existing `resources/` staging boundary.
- A copied-data tile layer and clamped camera with per-frame work bounded to
  visible cells, observable draw statistics, and separate static tile overlap.
- A deterministic `examples/tilemap` P3.3 acceptance scene using owned bitmap
  tiles without coupling tile geometry to sprite collision.

- Optional immediate-mode line, rectangle, ellipse, and triangle
  drawing on Simulator and device adapters.
- Value-owned solid, XOR, and 8x8 pattern paints, plus clipping, draw offsets,
  and bitmap draw modes through narrow graphics capability interfaces.
- Portable sentinel errors for invalid primitive geometry, colors, and draw
  modes.
- A deterministic `examples/primitives` acceptance scene covering every P3.1
  drawing and state operation.

### Changed

- `playdate` sentinel errors remain centralized in `errors.go` but are now
  classified by bitmap, graphics, framebuffer, offscreen, tile-map, animation,
  sprite, audio, and font domains instead of reusing the bitmap error type.
- P3.4 retained scene-local single ownership after its audit found no two real
  consumers sharing loading, caching, rollback, transition, and shutdown
  semantics; no speculative resource manager or reference counting was added.

### Compatibility

- The P3 navigation scene passed Windows SDK 3.1.1 Simulator Play/menu-return/
  Exit interaction. Its launcher artwork packaged successfully; conservative-GC
  hard-float build, USB deployment on COM3, and device launch passed. Matching
  Play/menu-return/Exit interaction and correct card, icon, and launch-image
  display were confirmed on the physical Playdate.
- The P3.3 scene passed Windows SDK 3.1.1 Simulator visual acceptance and a
  conservative-GC physical-device build, USB deployment, controls, jump,
  collision, camera, and matching scene execution. Soak remains unverified.
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
