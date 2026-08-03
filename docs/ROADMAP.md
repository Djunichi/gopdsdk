# Product roadmap

Status: `v0.4.0` release candidate preparation, updated 2026-08-02.

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
| P4 | `v0.4.0` | Persistence and Playdate system integration | Release candidate; final gates pending |
| P5 | `v0.5.0` | Advanced audio and music | In progress; P5.1-P5.2 device-accepted |
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

P4 adds durable game state and the normal Playdate system facilities required
by an external offline game. Persistence remains layered: the native adapters
provide owned files, while save schemas, migrations, and recovery policy are
portable Go behavior driven by the acceptance consumer.

### P4.0 — baseline and persistence contract — complete

- Repeat the released `v0.3.0` unit, CLI, native CI, SDK integration, and
  applicable physical-device gates before growing the API.
- Specify data-directory path rules, bundled-resource read behavior, file
  ownership, close and flush semantics, and portable error categories against
  the official filesystem contract.
- Define the failure matrix for interrupted writes, missing files, short I/O,
  incompatible versions, and corrupt saves before choosing the public save
  API.

The accepted filesystem contract follows the official SDK rather than host OS
semantics. Paths are game-relative. Read modes are flags: packaged PDX only,
Data only, or Data-first with packaged fallback; write and append always target
Data. Handles are closed rather than freed, with the official limit of 64 open
handles. Directory listing is non-recursive and identifies child directories
with a trailing slash. `mkdir` does not create parents, while `rename`
overwrites its destination and therefore becomes the P4.2 atomic-replacement
primitive.

P4.1 must copy the transient `geterr()` diagnostic at the failing call boundary
and wrap stable portable error categories without treating human-readable text
as a machine contract. Zero-byte reads mean EOF; negative results mean failure;
short successful reads remain valid, while short writes surface explicitly.
Close invalidates the owned Go handle even when native close reports failure,
preventing a second close against an indeterminate native lifetime. Seek input
must fit the official signed 32-bit position contract. P4.2 owns interrupted
replacement, incompatible-version, migration, and corrupt-save policy rather
than embedding those decisions in the filesystem adapter.

The released `v0.3.0` evidence is accepted as the P4 baseline and was not
re-promoted. The current pure-Go unit suite and vet passed on Windows on
2026-08-02. No new Simulator, SDK integration, USB, or physical-device evidence
is claimed by P4.0.

### P4.1 — owned filesystem — implemented

- Add the smallest capability for owned open files, read, write, seek, flush,
  close, stat, list, directory creation, rename, and removal required by the
  persistence consumer.
- Keep native error text available for diagnosis while exposing portable
  sentinel categories suitable for `errors.Is` decisions.
- Regression-test capability assertion and calls through `NewApplication` on
  both native ABI contexts, including idempotent cleanup after partial setup.

The implemented vertical slice exposes optional `FileSystem`, owned `File`,
official read-source flags, metadata and non-recursive listing, and the native
directory mutation operations. Portable validation covers paths, modes,
offsets, closed handles, EOF, and short writes. Native failures retain a copied
diagnostic and support `errors.Is(err, ErrFileIO)`. Deterministic runtime,
application-forwarding, API-snapshot, and generated Simulator/device adapter
tests pass.

The focused `examples/filesystem` scene then passed Windows SDK 3.1.1
Simulator and physical Playdate execution on 2026-08-02. It exercised write,
flush, close, rename, Data read, stat, non-recursive list, and recursive remove
in one native flow. The first hardware run exposed a device-only borrowed-string
lifetime bug in listing; the adapter now copies each callback name before
freeing bridge memory, and the repeated hardware run displayed the expected
`P4.1 filesystem OK`. The device artifact used 266,904 bytes of static RAM and
produced a 28,737-byte PDX. Multi-session durability, interrupted writes, soak,
and memory-growth measurement remain P4.2 or later evidence.

### P4.2 — save and configuration store — implemented

- Implement versioned save/load above P4.1 with atomic temporary-file
  replacement, explicit migration, size bounds, and deterministic corrupt-save
  handling.
- Add JSON only if the external save or configuration consumer selects it;
  serialization does not belong in the native adapters.
- Test first save, replacement, migration, unsupported future versions,
  interrupted replacement, short I/O, and recovery without silently discarding
  the last valid state.

The implemented `playdate/store` package keeps serialization in the consumer
and adds a bounded binary envelope containing schema version, payload length,
and checksum. Writes use a sibling temporary file followed by flush and close,
then attempt the documented overwriting rename. If the target cannot be
overwritten, a backup swap preserves a recoverable valid generation across
each rename boundary. Loads ignore stale temporary files, recover an orphaned
backup, reject malformed or future-version data deterministically, and run an
explicit migration for every schema step before persisting the upgrade.

Deterministic pure-Go tests cover the required failure matrix, device-style
non-overwriting rename, backup recovery, and preservation of the last valid
value. `examples/persistence` passed Windows SDK 3.1.1 Simulator and physical
Playdate execution on 2026-08-02, displaying `P4.2 STORE OK` after save,
migration, replacement, and reload. The first hardware run preserved the valid
final and completed temporary file but showed that device `rename` rejected an
existing destination despite the documented overwrite contract. The corrected
fallback passed the conservative-GC device gate at 267,116 bytes of static RAM
and a 30,269-byte PDX, USB deployment, and physical execution. Cross-launch
durability, injected power loss, soak, and memory-growth evidence remain
unverified.

### P4.3 — system menu and localization — implemented

- Add owned action, checkmark, and option menu items with callback lifetime,
  removal, title, and value behavior defined across Simulator and device.
- Expose system language and localized-text lookup without inventing a general
  translation framework; game-owned fallback text remains portable consumer
  policy.
- Exercise menu-driven settings whose localized values persist through P4.2,
  including callback cleanup during lifecycle termination.

The implemented slice exposes optional `SystemMenu` and `Localization`
capabilities. Action, checkmark, and option items own their native handles and
callbacks, support title and value access, remove idempotently, and are removed
by `NewApplication` after lifecycle termination. Localized text is copied into
Go ownership before the SDK allocation is freed; missing-key fallback remains
consumer policy. Deterministic tests cover value validation, ownership,
forwarding, localization, and termination cleanup. `examples/systemmenu`
persists its localized menu settings through P4.2. On 2026-08-02 it passed
Windows SDK 3.1.1 Simulator execution, the conservative device gate at 268,932
bytes of static RAM and a 35,674-byte PDX, USB deployment, and physical
Playdate execution. The menu callbacks changed both settings and the values
survived a game restart. Extended conservative-GC soak, memory-growth
measurement, and post-run crashlog inspection remain unverified.

### P4.4 — device and system status — implemented

- Add opt-in accelerometer sampling with explicit peripheral enablement and
  lifecycle behavior.
- Add power status, battery percentage and voltage, system volume, reduce-
  flashing preference, timezone, and 12/24-hour preference only through narrow
  capabilities required by the acceptance consumer.
- Preserve `ExitToLauncher` as the existing optional lifecycle capability and
  include it in the integrated P4 flow rather than creating a second exit API.

The implemented slice exposes optional `Accelerometer`, `PowerMonitor`, and
`SystemPreferences` capabilities. Sampling requires explicit peripheral
enablement, returns zero before enablement, and is disabled after lifecycle
termination. Power and global preferences remain read-only, and the existing
`Launcher` API is forwarded in the same acceptance flow. Deterministic tests
cover capability assertions, forwarding, power flags, pre-enablement reads, and
termination cleanup; generated Simulator and both device adapters are checked
for every bridge. On 2026-08-02 `examples/systemstatus` passed SDK 3.1.1
Simulator build and launch and the conservative device gate at 283,884 bytes of
static RAM and a 55,193-byte PDX, including USB deployment and physical Playdate
execution. Hardware observation confirmed accelerometer, battery, volume,
timezone, clock format, reduce-flashing, and `NONE`/`USB` power states. The first
device run exposed direct float-return ABI corruption for battery and volume;
passing their IEEE-754 bits across the TinyGo/C boundary corrected it. `CHARGE`
and `SCREWS` power states, conservative-GC soak, memory-growth measurement, and
post-run crashlog inspection remain unverified.

### P4.5 — optional online and debug facilities — implemented, integration unverified

- Assess official scoreboards as an independent optional online service; add
  them only with a concrete consumer, bounded callbacks, failure handling, and
  no implication of general networking or multiplayer support.
- Keep Simulator-only debug input and serial messaging separate from portable
  gameplay contracts and omit them if no acceptance or diagnostic consumer
  justifies their callback and platform boundaries.

The implemented optional `Scoreboards` capability copies every SDK-owned result
before the native bridge releases it, preserves immediate and callback failure
diagnostics, and permits at most one pending request of each operation kind.
Callbacks arriving after lifecycle termination are suppressed.
`examples/scoreboards` exercises board discovery, score submission, and personal
best retrieval, but no configured board or live Playdate service has yet been
used; online SDK integration and physical-device behavior remain unverified.
This capability implies neither general networking nor multiplayer support.
On 2026-08-02 both P4.5 examples passed Windows SDK 3.1.1 Simulator builds and
conservative device packaging. The scoreboards device build used 271,452 bytes
of static RAM with a 35,288-byte PDX; deployment and execution were not run.

The implemented optional `DebugMessages` capability receives device `msg` and
Simulator `!msg` input into an eight-message FIFO, copies at most 256 bytes per
message, drops the oldest message when full, and exposes polling from the normal
game callback rather than re-entering game code from the native serial callback.
Termination drains queued input. `examples/debugmessages` is the diagnostic
consumer and generated-bridge tests cover its callback path. On 2026-08-02 its
device build used 267,652 bytes of static RAM with a 26,054-byte PDX; USB
deployment and physical Playdate execution passed. Sending `msg device-hello`
directly over COM3 reached the native serial callback and displayed
`device-hello` in the running example. Live Simulator `!msg` delivery,
conservative-GC soak, memory-growth measurement, and post-run crashlog
inspection remain unverified.

### P4.6 — integrated multi-session acceptance game — implemented, physical multi-session gates pending

- Extend an external game using only public `gopdsdk` API to preserve settings
  and progress across normal restart, application update, failed write,
  migration, and corrupted-save scenarios.
- Exercise localized system-menu settings, at least one justified P4.4 system
  capability, and exit-to-launcher without a game-owned C bridge or imports
  from `internal` packages.
- Run deterministic portable failure tests first, then matching Simulator and
  physical-device multi-session gates; label every unrun durability or hardware
  scenario by its actual evidence level.

Crank Caverns now persists settings, clears, best time, explicit checkpoints,
and run score through four schema versions. Its title and pause menus expose
continue, start-new, save, and load behavior; localized System Menu settings,
battery status, and Launcher exit use only public capabilities. Portable tests
cover failed writes, corrupt payloads, migration, reload, and new-run reset.
Windows Simulator interaction, conservative-GC device build (277,524 bytes
static RAM), USB installation, and device launch command passed on 2026-08-02.
Physical cross-restart/update and injected-failure scenarios, soak, memory
growth, and crashlog inspection remain unverified.

### P4.7 — `v0.4.0` release — release candidate preparation

- Review the exported persistence ownership, callback, migration, and recovery
  contracts.
- Update API, compatibility, examples, migration notes, changelog, and release
  procedure for `v0.4.0`.
- Verify a published external consumer without a local `replace` after tagging.

## P5 — advanced audio and music

### P5.1 — advanced sample playback

- Add an optional `SamplePlayers` capability without widening the base
  `Context` or breaking the P2 sound-effect API.
- Support bounded repeat counts, forward and reverse sample rates, duration,
  playback-position control, and variable-rate streaming music with explicit
  player ownership.
- Reject negative streaming rates with `ErrAudioReverseUnsupported`; reverse
  playback remains sample-only and requires PCM rather than ADPCM assets.
- Forward the capability through `NewApplication` and both native ABI contexts.
- Native bridge generation is unit-tested. On 2026-08-02 the acceptance game
  passed audible variable-rate sample and streaming-music interaction in
  Windows Simulator and on a physical Playdate after conservative build, USB
  installation, and launch. macOS/Linux native SDK integration, extended soak,
  memory-growth measurement, lifecycle stress, and crashlog inspection remain
  unverified.

### P5.2 — timed fades and completion

- Add optional completion callbacks to owned sample and file players, with
  replacement and close releasing retained Go callbacks.
- Add file-player linear fades measured against the native 44.1 kHz audio
  clock, with bounded callback work and one-shot fade completion.
- Forward the optional audio clock through `NewApplication` and both native ABI
  contexts. Callback ownership is unit-tested, and official Windows Simulator
  and hard-float device builds pass. On 2026-08-02 the acceptance scene passed
  audible sample completion, a half-second streaming fade, completion counters,
  and advancing audio-clock display in Windows Simulator and on a physical
  Playdate after installation and launch through COM3. Extended soak,
  memory-growth measurement, lifecycle stress, and crashlog inspection remain
  unverified.

### P5.3 — routing, synthesizers, and signals

- Add explicitly owned channels without widening the base `Context`.
- Route sample, file, and synth sources through owned graph edges; source and
  channel closure must detach edges before native resources are released.
- Support channel volume and pan and make modulator attachment explicit.
- Add owned waveform synthesizers, envelopes, LFOs, and control signals as one
  vertical slice, including scheduled note-on and note-off behavior.
- Keep custom render and signal callbacks out of the public contract until their
  bounded, zero-allocation steady-state behavior can be proven end to end.
- Detach every retained routing and modulation edge before either endpoint is
  released.
- Unit tests and official Windows Simulator and conservative hard-float device
  builds pass. On 2026-08-02 the full routing, waveform, modulation, envelope,
  control-signal, transpose, note-off, volume, and pan matrix passed audible
  interaction in Windows Simulator and on a physical Playdate after USB
  installation through COM3 and launch. macOS/Linux SDK integration, extended
  soak, memory-growth measurement, lifecycle stress, and post-run crashlog
  inspection remain unverified.

### P5.4 — instruments, sequences, and effects

- Add explicitly owned instruments, voice ranges, tracks, note and control
  events, and sequences with documented parent/child ownership.
- Support MIDI loading and programmatic dynamic-music construction without
  transferring ownership implicitly between tracks, instruments, and synths.
- Bound completion callbacks and release retained callbacks on replacement,
  stop, and close.
- Add typed filters, bit crushing, ring modulation, delay, and overdrive as
  separate owned effect values attached through channel graph edges.
- Keep effect parameters and signal modulators explicit; detach effects and
  delay taps before freeing either side of an edge.
- Accept the combined P5 graph through dynamic music and sound scenes in the
  Simulator and on hardware.

### P5.5 — microphone input

- Add microphone input as an optional `Context` capability without widening the
  base application contract.
- Expose explicit permission request and denial states, input-source selection,
  and owned start/stop recording lifetime.
- Keep native recording callbacks bounded and allocation-free, copy samples
  only into caller-provided bounded buffers, and never retain a game buffer
  after its call completes.
- Stop recording and release callbacks on lifecycle termination, permission
  revocation, replacement, and close.
- Test permission, privacy, callback, overflow, and memory behavior end to end;
  accept recording audibly in the Simulator and on physical hardware without
  claiming microphone readiness from build or discovery alone.

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
