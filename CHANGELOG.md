# Changelog

All notable release changes are documented here. The module remains pre-v1;
minor releases may intentionally change the public API when release notes call
that out.

## Unreleased - v0.5.0

### Added

- Optional `playdate.SamplePlayers` and explicitly owned `SamplePlayer` APIs
  with bounded repeats, forward/reverse rates, duration, and playback offset.
- Optional `playdate.VariableRatePlayer` support for streaming `FilePlayer`
  pitch and speed control without widening the base `FilePlayer` contract;
  negative streaming rates return `ErrAudioReverseUnsupported`.
- Simulator and device ABI forwarding for advanced sample and file-player rate
  operations, including forwarding through `NewApplication`.
- A P5.1 audio acceptance scene with independent sample and music rate controls,
  reverse-start positioning, and state-driven redraws.
- Optional `CompletionPlayer`, streaming `FadingPlayer`, and `AudioClock`
  capabilities with explicit callback replacement, one-shot fade completion,
  close-time release, and 44.1 kHz frame timing.
- P5.2 Simulator/device ABI forwarding and acceptance-scene controls for
  play/fade/stop, completion counters, and sampled audio time.

### Fixed

- Reject reverse `FilePlayer` rates before entering the native API. The official
  streaming player supports positive rates only; reverse remains available to
  PCM `SamplePlayer` assets and is unsupported for ADPCM.

### Compatibility

- The full Go test and vet suites pass on Windows; Simulator and conservative
  hard-float device builds pass with the official SDK.
- Audible variable-rate sample and streaming-music interaction passed in
  Windows Simulator and on a physical Playdate on 2026-08-02. The device build
  used 268,940 bytes of static RAM and produced a 125,340-byte PDX; USB install
  on COM3 and launch succeeded.
- macOS/Linux native SDK integration, extended hardware soak, memory-growth,
  lifecycle stress, and post-run crashlog inspection remain pending.
- P5.2 is unit-tested and passes official Windows Simulator and hard-float
  device builds. Its device artifact uses 269,876 bytes of static RAM and
  produces a 130,127-byte PDX. Audible sample completion, the half-second music
  fade, completion counters, and advancing audio-clock display passed in
  Windows Simulator and on a physical Playdate on 2026-08-02; installation
  through COM3 and launch succeeded. Extended soak, memory-growth measurement,
  lifecycle stress, and post-run crashlog inspection remain pending.

## v0.4.0 - release candidate (2026-08-02)

### Added

- Owned Playdate filesystem operations and portable file-error categories.
- `playdate/store` bounded versioned persistence with checksums, atomic
  replacement, backup recovery, and stepwise migration.
- Owned System Menu items, localized string lookup, accelerometer, power and
  system-preference capabilities.
- Optional bounded scoreboards and debug-message capabilities.
- A P1-P4 Crank Caverns consumer with persisted settings, progress, score,
  explicit save/load checkpoints, migrations, battery status, and Launcher
  exit.

### Changed

- Tile-map draw failures retain both `ErrTileMapDraw` classification and their
  underlying cause without `errors.Join`, keeping the device runtime sequential.
- Store migration failures retain `ErrMigration` and the direct cause through
  `errors.Is` without formatting arbitrary errors in device code.
- Device validation distinguishes unsupported channel operations from
  reflectlite-retained `runtime.chanLen`/`runtime.chanCap` query helpers.

### Compatibility

- The full Go test and vet suites pass on the Windows candidate checkout.
- Crank Caverns passed Windows Simulator interaction, conservative-GC hard-
  float device build at 277,524 bytes of static RAM, USB installation on COM3,
  and the device launch command. Physical multi-session durability, soak,
  memory-growth, crashlog, native CI, and post-tag proxy checks remain pending.
- See [MIGRATING.md](MIGRATING.md) and [COMPATIBILITY.md](COMPATIBILITY.md) for
  consumer guidance and evidence limits.

## v0.3.0 - 2026-08-02

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
- Callback-scoped zero-copy framebuffer access with explicit dirty-row
  reporting, plus drawing into explicitly owned offscreen bitmaps with drawing
  context restoration.
- A deterministic `examples/framebuffer` scene covering portable pixel layout,
  dirty-range aggregation, callback lifetime, and owned offscreen drawing.
- The external [Crank Caverns](https://github.com/Djunichi/gopdsdkgame)
  acceptance game, currently in a private repository, integrating every
  implemented P1-P3 gameplay slice through public `gopdsdk` API.

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
- P3.2 framebuffer and offscreen contracts are unit-tested through the portable
  implementation and generated adapters. The complete Crank Caverns consumer
  exercises both capabilities; separate visual and physical-device evidence for
  `examples/framebuffer` remains unverified.
- Crank Caverns completes the integrated P3 consumer boundary and has portable
  deterministic gameplay and render-plan coverage. Fixed frame-time,
  bounded-heap, extended soak, and post-run device-log measurements for the
  complete game remain unverified.

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
