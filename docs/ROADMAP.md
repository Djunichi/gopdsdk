# Product roadmap

Status: `v0.7.0` released; the remaining offline official-SDK capability gaps
are allocated through `v0.12.0`, followed by the `v1.0.0` contract release,
updated 2026-08-08.

This is the only planning document under `docs/` and the canonical roadmap from
the released foundation to `v1.0.0`. Completed-scope evidence lives in
`COMPATIBILITY.md`; public contracts live in `API.md`; current implementation
rules live in `AGENTS.md`. Completing a scope means its product boundary is
implemented; Simulator, SDK integration, and physical-device claims still
require the evidence named in `COMPATIBILITY.md`.

## Product destination

`v1.0.0` means an external Go module can create, test, build, optimize, and
release an offline Playdate game using only public gopdsdk API, without a
game-owned C bridge or imports from `internal` packages. Every official C SDK
capability that can materially enable an offline game design must either have a
public Go equivalent or a documented, functionality-preserving replacement.

This does not promise unlimited game complexity. Games remain bounded by the
Playdate hardware, its one-bit display, the official SDK, and the documented
device Go profile. Symbol-for-symbol coverage is not required: deprecated
entry points, duplicate representations, implementation plumbing, Lua
integration, and C facilities replaced without loss by Go may be omitted.
Release scopes use repository-owned acceptance consumers; validation through a
commercially realistic external game begins after `v1.0.0` and is not a gate
for that release.

The roadmap allocates the remaining offline capability work through P13. Work
after `v1.0.0` is tracked by feature and release.

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
| P8 | `v0.8.0` | Complete offline sprite and collision facilities | Planned |
| P9 | `v0.9.0` | Complete offline sound, sample, and output facilities | Planned |
| P10 | `v0.10.0` | Complete offline system and lifecycle facilities | Planned |
| P11 | `v0.11.0` | Remaining offline filesystem, scoreboards, and media gaps | Planned |
| P12 | `v0.12.0` | Final device Go-profile audit and runtime hardening | Planned |
| P13 | `v1.0.0` | API freeze, compatibility contract, and release evidence | Planned |

## Remaining capability milestones

- After P11, every materially useful offline official-SDK capability should be
  implemented or replaced without loss of game functionality.
- P12 is the final feature stage before `v1.0.0` and owns the device Go-profile
  limitations rather than allowing them to shape earlier capability releases.
- P13 freezes and releases the stable `v1.0.0` contract. Real-game validation
  begins afterward.

## Offline official-SDK completion policy

P8 through P11 close every remaining known official C SDK gap that can materially block
an offline game. Selection is based on lost capability, not anticipated
popularity: a feature is in scope even when uncommon if omitting it prevents a
game design that the official C SDK supports.

A C entry point may be intentionally omitted only when at least one of these is
documented and tested:

- the public Go API already provides equivalent behavior;
- a portable Go implementation preserves the useful behavior and target
  constraints;
- the entry point is deprecated, duplicate ABI compatibility, allocator or
  formatting plumbing, Lua integration, or unsafe userdata with no capability
  beyond Go-owned state;
- it is HTTP/TCP networking, which is reserved for `v1.1.0`;
- it is Simulator-only diagnostics with no effect on a shipped game.

Every included slice requires ownership and error contracts, forwarding through
both native contexts, deterministic tests, a repository-owned acceptance
consumer, official Simulator integration, device build, and applicable physical
device behavior. Callback APIs additionally require bounded queues, no unsafe
re-entry into Go, termination suppression, and overflow behavior.

## P8 — complete sprites and collisions — `v0.8.0`

P8 removes restrictions that prevent representing an official native sprite
configuration or collision query.

### P8.1 — sprite geometry and presentation

Status: complete. Geometry, presentation, useful getters, update/collision
enable state, native tilemap ownership and attachment, both native adapters,
deterministic tests, and Simulator/device visual acceptance are complete.

- Add center, bounds, image flip, draw mode, opacity, stencil image/pattern,
  clip rectangle, draw-offset policy, tilemap attachment when it is not
  equivalent to the portable tile layer, and the corresponding useful getters.
- Add update-enabled and collision-enabled state with deterministic lifecycle
  behavior.

### P8.2 — queries and display-list control

- Add line queries and their detailed hit information, collision checks,
  sprite count, bulk add/remove, remove-all, and collision-world reset.
- Omit bulk functions only where repeated typed Go operations are behaviorally
  identical, preserve ordering, and cannot introduce partial-failure behavior.

### P8.3 — bounded callbacks

- Add per-sprite draw, update, and collision-response callbacks.
- Define callback ordering, frame lifetime, panic/error policy, mutation rules,
  cleanup, queue bounds, and behavior after termination on both ABIs.
- Do not expose raw C userdata; Go-owned callback identity and state provide the
  same useful capability.

### P8.4 — release

- Exercise procedural sprites, dynamic pair-specific collision responses,
  line queries, stencils, flips, clipping, enable transitions, and cleanup in a
  repository-owned acceptance scene before releasing `v0.8.0`.

## P9 — complete sound — `v0.9.0`

P9 completes offline playback, synthesis, samples, routing, and device output.

### P9.1 — output and channel control

- Add the default channel, headphone/headset state, output routing, output as a
  source, channel membership, output activation, and the missing useful volume,
  pan, and modulator controls.
- Define behavior for hardware states that cannot be exercised in Simulator.

### P9.2 — samples and players

- Add owned sample buffers, packaged and caller-data loading, loading into an
  existing sample/player, copying, decompression, sample data inspection,
  playback ranges, loop callbacks, buffering, and underrun status/control.
- Preserve bounded memory and explicitly distinguish borrowed sample views from
  native-owned copies.

### P9.3 — synthesis, modulation, and sequencing

- Add one-pole filtering, wavetable synthesis, remaining envelope/synth
  parameters, rate/pan/volume/effect modulation, signal/controller lookup, and
  sequence/track/note introspection that enables behavior unavailable through
  the current write-only API.
- Add custom synth generators and callback audio sources through a bounded
  device-safe callback path. Raw C function and userdata entry points remain
  intentionally omitted once the Go callback contract is equivalent.
- Treat MP3 callback streaming as required only if the installed official SDK
  supports it as a shipping-device capability and existing file/sample players
  cannot preserve its behavior.

### P9.4 — release

- Run audible Simulator and physical-device acceptance for routing, headphones,
  samples, wavetable/custom synthesis, callbacks, underruns, lifecycle cleanup,
  soak, memory bounds, and unchanged device logs before releasing `v0.9.0`.

## P10 — complete system integration — `v0.10.0`

P10 completes system facilities that can affect a shipped offline game.

### P10.1 — launch and lifecycle control

- Add launch arguments, restart, menu image, auto-lock control, crank-sound
  control, and official mirror lifecycle events when delivered to games.
- Add button callbacks only if they provide behavior that frame snapshots cannot
  preserve; otherwise document and test the snapshot equivalence.

### P10.2 — clock, calendar, and device information

- Add epoch/calendar conversion, current epoch time, elapsed-time measurement,
  and system information with portable owned values.
- Keep server time with networking in `v1.1.0` when it requires online access.

### P10.3 — intentional system omissions and release

- Omit delay, allocator, instruction-cache, printf-style formatting/parsing,
  fatal-error, and raw userdata entry points from public API because they are
  runtime plumbing or have direct Go equivalents.
- Keep console logging and FPS drawing as development facilities unless a
  repository diagnostic proves that they need public Simulator-only contracts.
- Exercise restart, launch arguments, menu image, time conversion, auto-lock,
  crank sound, lifecycle cleanup, and system information before releasing
  `v0.10.0`.

## P11 — remaining offline services and media — `v0.11.0`

P11 closes residual subsystem gaps after the focused graphics, sprite, sound,
and system releases.

### P11.1 — filesystem completion

- Add current-position reporting if `io.Seeker` cannot preserve the official
  behavior and expose any remaining filesystem diagnostic needed for correct
  error classification.
- Do not expose `geterr` as an independently retained string; continue copying
  its transient diagnostic at the failing operation boundary.

### P11.2 — scoreboards acceptance

- Retain scoreboards as the sole pre-1.0 online exception already present in
  the SDK and complete live configured-service Simulator and physical-device
  acceptance, failure, cancellation, termination, and bounded-callback gates.
- This scope does not introduce general network access or multiplayer.

### P11.3 — video, JSON, and residual audit

- Audit the installed official video API for buffering, byte/progress/error
  diagnostics, and file/network-source operations. Add every offline operation
  that provides behavior unavailable through the existing `VideoPlayer`.
- Provide or qualify a device-safe Go JSON codec and omit the C callback JSON
  API when the Go replacement supports bounded decode/encode without loss of
  offline functionality.
- Reconcile every official header entry against `implemented`, `equivalent`,
  `intentionally omitted`, or `v1.1 networking`; no unexplained entries may
  remain before releasing `v0.11.0`.

## P12 — device Go profile and runtime hardening — `v0.12.0`

P12 is deliberately the last feature stage before `v1.0.0`. Earlier releases
close official-SDK capability gaps without expanding the accepted Go language
subset; this stage then audits whether the remaining runtime restrictions block
otherwise supported offline game architectures.

- Reassess goroutines, channels, `select`, `defer`, panic/recover, reflection,
  finalizers, cgo, standard-library coverage, scheduler choice, GC strategy,
  and application runtime hooks against the verified TinyGo/device toolchain.
- Implement and prove every safely supportable facility whose absence causes a
  material offline-game limitation. Keep restrictions that are inherent,
  unsafe, unbounded, or replaceable without loss, with precise compile-time or
  probe diagnostics and documented alternatives.
- Stress callback queues introduced by P8 and P9 together with GC, allocation,
  lifecycle transitions, long-running updates, panic policy, and native
  resource cleanup.
- Run extended physical-device soak, memory-growth, overflow, and unchanged-log
  gates before releasing `v0.12.0`.

## P13 — stable contract and `v1.0.0`

P13 adds no planned feature family. It releases the accumulated offline SDK as
a stable contract.

- Freeze and review the exported API, ownership, error, callback, compatibility,
  and device Go-profile contracts.
- Verify reproducible release builds and the published-module workflow without
  a local `replace` using minimal external fixture modules, not a real game.
- Require native CI and external-fixture CLI acceptance on Windows, macOS, and
  Linux. Run official SDK, Simulator, device-build, USB, and physical-device
  gates on every host advertised at that level; otherwise label the host
  explicitly unverified.
- Run the complete repository-owned acceptance-scene regression suite, extended
  hardware soak, resource and heap bounds, lifecycle cleanup, and unchanged
  device-log checks.
- Publish migration, semantic-versioning, support, deprecation, compatibility,
  and evidence policies for the `v1.x` line.

## Post-1.0 real-game validation

After `v1.0.0`, validate the stable SDK through one or more commercially
realistic external games using only the published public module. Exercise
multi-scene transitions, long-lived resource ownership, saves and recovery,
audio, callbacks, lifecycle paths, performance budgets, packaging, and release
operations. Findings become normal `v1.x` compatibility fixes or evidence;
they do not retroactively gate `v1.0.0`.

## `v1.1.0` — networking and multiplayer feasibility

Networking is deliberately not a `v1.0.0` gate. The
[official Playdate 3.1.1 C API](https://sdk.play.date/3.1.1/Inside%20Playdate%20with%20C.html#_networking)
exposes permission-gated HTTP and outbound TCP connections, so online
multiplayer is technically possible, but it adds server operations, protocol
design, latency, reconnection, security, privacy, and long-lived compatibility
obligations unrelated to the offline SDK foundation.

The `v1.1.0` networking track may investigate before committing a stable public
network API:

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
