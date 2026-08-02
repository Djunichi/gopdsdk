# Product roadmap

Status: accepted direction after the `v0.3.0` release, updated 2026-08-02.

This document is the canonical high-level roadmap from the released `v0.3.0`
baseline to `v1.0.0`. The detailed P0, P1, and P2 documents retain their
historical plans, decisions, and evidence. Completing a scope means its product
boundary is implemented; Simulator, SDK integration, and physical-device
claims still require the evidence named in `COMPATIBILITY.md`.

## Product destination

`v1.0.0` means an external Go module can create, test, build, optimize, and
release a commercially realistic offline Playdate game using only public
gopdsdk API, without a game-owned C bridge or imports from `internal` packages.

This does not promise unlimited game complexity. Games remain bounded by the
Playdate hardware, its one-bit display, the official SDK, and the documented
sequential TinyGo subset. Full coverage of every official C API function is not
a prerequisite for `v1.0.0`; each public addition must still be justified by a
complete consumer-driven vertical slice.

The roadmap contains exactly P0 through P7. Work after `v1.0.0` is tracked by
feature and release rather than by reserving speculative P8 or later scopes.

## Status vocabulary

- **Planned** means the product boundary is accepted but no readiness evidence
  is implied.
- **Implemented** means the scoped behavior and deterministic tests exist on
  both adapters.
- **Release candidate** means the documented native CI, SDK integration,
  external-consumer, and applicable physical-device gates have been run.
- **Released** means the reviewed commit is tagged and published. Known
  limitations remain labeled by their actual evidence level.

Scope completion and evidence completion are deliberately separate. A missing
hardware run does not erase implemented behavior, and a successful dry-run or
cross-build does not promote a capability to device-ready.

## Release sequence

| Scope | Target | Outcome | Status |
| --- | --- | --- | --- |
| P0 | foundation | Cross-platform CLI, Simulator and device build foundations | Complete |
| P1 | `v0.1.x` | Lifecycle, input, bitmap graphics, resources, device runtime | Complete |
| P2 | `v0.2.0` | API guard, sprites, collisions, animation, base audio, fonts/UI | Released; evidence limits documented |
| P3 | `v0.3.0` | Production-capable 2D rendering and game worlds | Released; integrated consumer complete, evidence limits documented |
| P4 | `v0.4.0` | Persistence and Playdate system integration | Planned |
| P5 | `v0.5.0` | Advanced audio and music | Planned |
| P6 | `v0.6.0` | Advanced graphics, media, and performance facilities | Planned |
| P7 | `v1.0.0` | Production hardening through a real external game | Planned |

## Capability milestones

- After P3, the SDK should support complete conventional 2D gameplay and the
  framebuffer foundation required by custom software renderers.
- After P4, games should support durable progress, configuration, localization,
  and normal Playdate system integration.
- After P5, games should support production sound design and dynamic music.
- After P6, advanced graphics and a measured software-3D proof should remove
  the remaining rendering-class limitations of the SDK.
- After P7, the combined SDK should be proven through a releasable external game
  and become the stable `v1.0.0` contract.

## P3 — 2D rendering and game worlds

P3 removes the main conventional 2D rendering limitations without introducing
a generic game engine, ECS, scene graph, widget framework, or physics engine.

### P3.0 — baseline audit

- Repeat the full `v0.2.0` unit, CLI, Simulator, and device-build gates before
  growing the API.
- Treat toolchain-cache failures and P2 regressions as `v0.2.x` maintenance or
  explicitly documented environment limitations, not as P3 product features.
- Record the accepted P3 frame-time, heap, resource, and artifact-size budgets.

### P3.1 — drawing primitives and state — implemented

- Add only the drawing primitives required by the acceptance scene: lines,
  rectangles, ellipses, triangles or polygons, and their filled variants.
- Add clipping, draw offset, patterns, and draw modes through narrow capability
  interfaces rather than expanding every consumer-facing mock at once.
- Specify invalid geometry, color, and state behavior with portable errors.

The accepted vertical slice exposes optional `PrimitiveGraphics` and
`GraphicsState` capabilities, value-owned solid/XOR/pattern paint, and portable
validation errors. `examples/primitives` passed deterministic unit tests,
Windows SDK 3.1.1 Simulator compilation and visual execution, device
hard-float build, USB deployment, and matching physical Playdate execution on
2026-08-02. Conservative-GC soak and memory-growth measurement remain unrun;
P3.1 evidence does not complete P3.0 or later P3 scopes.

### P3.2 — framebuffer and offscreen drawing — implemented

- Add scoped framebuffer access with explicit lifetime, no hidden copy,
  dirty-row reporting, and deterministic pure-Go pixel tests.
- Make retention beyond the callback or frame invalid by contract.
- Add drawing into owned bitmaps only when the framebuffer or world consumer
  proves the required context and cleanup model.

The accepted vertical slice provides callback-scoped zero-copy framebuffer
access, checked lifetime, explicit dirty-row aggregation, and drawing into
owned bitmaps with drawing-context restoration. Deterministic portable and
generated-adapter tests cover the contract; `examples/framebuffer` provides the
focused scene and Crank Caverns exercises both capabilities in its integrated
consumer. Separate focused-scene SDK visual and physical-device evidence
remains unverified.

### P3.3 — tile map and camera — implemented

- Build a tile-map and camera vertical slice with visible-range work bounded per
  frame and simple static collision geometry.
- Keep tile collision separate from the P2 sprite collision contract and avoid
  a general physics abstraction.

The portable camera and tile layer copy their configuration, clamp world-space
viewports, bound each draw pass to visible cells, and expose static tile overlap
queries independently of sprites. Deterministic pure-Go tests and
`examples/tilemap` cover the vertical slice. Windows SDK 3.1.1 Simulator visual
acceptance and conservative-GC physical-device build, USB deployment, input,
jump, collision, camera, and matching scene execution passed on 2026-08-02.
Soak and memory-growth measurement remain unverified.

### P3.4 — repeated resource ownership — assessed

- Introduce a higher-level resource owner only after two real consumers expose
  the same loading, caching, rollback, transition, and shutdown behavior.
- Do not require reference counting if a simpler single-owner scene contract is
  sufficient.

The P1–P3 acceptance scenes repeat explicit initialization rollback and
idempotent termination cleanup, but they do not yet have two consumers with the
same cache keys, scene-transition lifetime, or heterogeneous loading policy.
P3.4 therefore keeps resources directly owned by each scene and adds no public
manager or reference counting. P3.5 must use that contract for its first
integrated scene; a higher-level owner is justified only if a second real
consumer duplicates the complete behavior above.

### P3.5 — integrated acceptance game — complete

- Build one external scene using primitives, framebuffer or offscreen drawing,
  camera, tiles, sprites, collisions, animation, audio, fonts, and lifecycle.
- Include an in-game main menu with `Play` and `Exit`, a gameplay action that
  returns to that menu without restarting the process, and `Exit` backed by
  the official `exitToLauncher` lifecycle path.
- Package the official launcher artwork baseline through `pdxinfo.imagePath`:
  a 350x155 `card.png`, 32x32 `icon.png`, and 400x240 `launchImage.png`.
- Pass fixed frame-time, bounded-memory, cleanup, Simulator parity, physical
  device, and soak gates without promoting unrun host capabilities.

The navigation foundation is accepted: `examples/navigation` covers `Play`, an
in-process B-button return to the main menu, and `ExitToLauncher`. Matching
interaction passed Windows SDK 3.1.1 Simulator and physical Playdate on
2026-08-02. The generated 1-bit card, icon, and launch image also displayed
correctly in the device Launcher.

The external [Crank Caverns](https://github.com/Djunichi/gopdsdkgame) consumer,
currently in a private repository, completes the integrated product slice using
only public `gopdsdk` API. It combines every implemented P1-P3 gameplay
capability in one playable scene and retains explicit scene ownership and
rollback. Portable deterministic gameplay and render-plan tests pass. Fixed
frame-time, bounded-heap, extended physical-device soak, termination cleanup
observation, and post-run device-log comparison remain unverified evidence and
are not promoted by scope completion.

### P3.6 — `v0.3.0` release — complete

- Review the exported snapshot and ownership contracts.
- Update API, compatibility, examples, migration notes, and release checks.
- Verify a published external consumer without a local `replace` after tagging.

The reviewed API and ownership contracts, compatibility matrix, examples,
release notes, and release procedure were updated for `v0.3.0`. The clean
published-module consumer check necessarily follows tag publication and remains
a post-release verification gate.

P3 should make platformers, puzzle games, arcade games, top-down games,
software-rendered pseudo-3D, and similar offline gameplay practical. P3 only
provides the low-level framebuffer capability; measured software-3D acceptance
belongs to P6 and does not become a general 3D-engine API.

## P4 — persistence and system integration

- Owned filesystem handles and portable file errors.
- Save/load with atomic replacement and explicit format migration.
- JSON support only where a real save/configuration consumer requires it.
- System menu items and callbacks.
- Localization and system-language information.
- Accelerometer, power status, system volume, and exit-to-launcher capabilities.
- Official scoreboards may be added as an independent optional online service;
  they do not imply general networking or multiplayer support.
- Simulator-only debug input and serial messaging kept separate from portable
  gameplay contracts.
- Acceptance through a multi-session external game that preserves settings and
  progress across restart, update, failure, and corrupted-save scenarios.

## P5 — advanced audio and music

- Samples and sample players beyond the P2 sound-effect convenience API.
- Channels, routing, fades, completion callbacks, and audio-clock timing.
- Synthesizers, instruments, sequences, signals, and effects only as separate
  vertical slices with explicit graph ownership.
- Bounded callback work and zero-allocation steady-state playback where the
  official audio callback requires it.
- Acceptance through dynamic music and sound scenes on Simulator and hardware.

Microphone input remains optional unless a concrete game requires it and the
permission, privacy, callback, and memory contracts can be tested end to end.

## P6 — advanced graphics, media, and performance

- Masks, stencils, transforms, drawing contexts, and display controls selected
  by real consumers.
- Advanced sprite redraw and dirty-region behavior.
- Video and other specialized media as independent optional slices.
- Frame-time, heap, native-resource, and artifact-size diagnostics that work for
  an external game rather than only repository fixtures.
- Hardware acceptance for at least one external software renderer with no
  frame-loop allocation: a representative raycaster, wireframe, or bounded
  polygon experiment with explicit frame-time and memory budgets.

P6 expands what games can render; it does not promise a GPU, a general 3D
engine, or performance beyond Playdate hardware limits. Passing the renderer
fixture proves that gopdsdk exposes the required primitives; the renderer stays
consumer code unless two real games justify a reusable package.

## P7 — production hardening and `v1.0.0`

P7 proves the combined SDK rather than adding broad speculative API.

- Build a commercially realistic external game using only public API.
- Exercise multiple scenes, resource transitions, saves, audio, lifecycle,
  pause/lock/low-power paths, and bounded long-running gameplay.
- Define and pass frame-time, heap, native-resource, and artifact-size budgets.
- Run extended physical-device soak and regression suites with unchanged device
  logs.
- Verify reproducible release builds and the published-module workflow without
  a local `replace`.
- Require native CI and external-consumer CLI acceptance on Windows, macOS, and
  Linux. Run official SDK, Simulator, device-build, and physical deployment on
  every host advertised at that evidence level; otherwise keep that host
  explicitly unverified instead of inferring support from CI or cross-builds.
- Repeat the P6 software-renderer fixture as a regression gate, without making
  3D a requirement for the final external game's design.
- Freeze the reviewed public contract and publish migration and compatibility
  policy for the `v1.x` line.

## Post-1.0 theory — networking and multiplayer

Networking is deliberately not a `v1.0.0` gate. The
[official Playdate 3.1.1 C API](https://sdk.play.date/3.1.1/Inside%20Playdate%20with%20C.html#_networking)
exposes permission-gated HTTP and outbound TCP connections, so online
multiplayer is technically possible, but it adds server operations, protocol
design, latency, reconnection, security, privacy, and long-lived compatibility
obligations unrelated to the offline SDK foundation.

A future networking track may investigate, without committing it to a release:

1. An HTTP leaderboard or asynchronous turn/ghost exchange with offline queue
   and explicit network permission.
2. A small authoritative server and a two-device TCP game with sequence
   numbers, bounded messages, timeouts, reconnection, and state interpolation.
3. Physical acceptance with two Playdates across disconnect, pause, lock, and
   server-version transitions.

The likely topology is two outbound clients connected to an external server;
the current C API does not provide a general inbound listen/server socket for
hosting a peer session on one Playdate. USB serial messaging remains a
development facility, not a consumer multiplayer transport.

This is a feasibility hypothesis, not a promised feature. No public networking
API should be designed until an end-to-end prototype has
measured permission flow, callback behavior, memory, bandwidth, latency, and
failure recovery against the normative official C API and hardware.

## Roadmap rules

- Official Playdate headers, documentation, examples, and observable tools are
  normative; third-party implementations remain behavioral references only.
- Every scope grows through narrow vertical slices on both runtime adapters.
- Exported API changes update the snapshot, `API.md`, compatibility evidence,
  and release notes in the same change.
- Resource ownership is explicit and deterministic; finalizers never carry
  correctness.
- A capability is not ready because its executable or symbol was discovered.
  Evidence is always labeled unit, external-consumer CLI, native CI, SDK
  integration, Simulator, device build, or physical device.
- Scope may change when a real consumer or hardware measurement contradicts
  the plan; the reason and affected release boundary must be documented here.
