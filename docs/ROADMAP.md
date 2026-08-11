# Product roadmap

Status: `v0.8.0` declared; publication is pending. The remaining offline
official-SDK capability gaps are allocated through `v0.12.0`, followed by the
`v1.0.0` contract release, updated 2026-08-09.

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

P9 through P11 close every remaining known official C SDK gap that can materially block
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

## P9 — complete sound — `v0.9.0`

P9 completes offline playback, synthesis, samples, routing, and device output.

### P9.1 — output and channel control

Development started with default-channel access, synchronous headphone/headset
state, and headphone/speaker activation on both native adapters. Deterministic
unit coverage exists. The official Windows SDK 3.1.1 Simulator visibly passed
scene initialization and the disconnected-headset state on 2026-08-09. The
conservative-GC hard-float build and COM3 deployment also passed; the user
confirmed physical headphone routing and live insertion/removal state changes.
Headset-microphone detection, simultaneous outputs, soak, memory growth, and
post-run device-log cleanliness remain unverified.

- Continue with output routing, output as a source, channel membership, and the
  missing useful volume, pan, and modulator controls.
- Define behavior for hardware states that cannot be exercised in Simulator.

### P9.2 — samples and players

Implemented owned sample buffers, packaged and caller-data loading, in-place
sample reload, sample attachment/replacement, borrowed data inspection and
bounded copying, decompression, playback ranges, loop callbacks, file-player
reload/buffering/loop ranges, and underrun status/control on both native
adapters. Caller slices are copied into native-owned storage; `SampleData` views
expire with their sample, and players borrow rather than own attached samples.
Deterministic unit, public-API, runtime, and generated-adapter coverage passes.
The official Windows SDK 3.1.1 build passed on 2026-08-09. The conservative-GC
hard-float device build also passed. The user confirmed visible initialization,
audible range playback with A, and stop with B in the official Windows
Simulator. Loop-callback behavior, final-package device deployment/execution,
and audible/visible device behavior remain unverified.

- Confirm range/loop behavior on a physical Playdate and exercise the loop
  callback and streaming underrun controls in acceptance scenes.
- Measure soak/memory growth and inspect post-run device logs when requested.

### P9.3 — synthesis, modulation, and sequencing

Development started with synth-owned envelope curvature, velocity sensitivity,
and note-range rate scaling on both native adapters. The isolated
`examples/synthesis` scene passed audible comparison of negative and positive
curvature in the official Windows SDK 3.1.1 Simulator and, after conservative
hard-float build and COM3 deployment, on a physical Playdate on 2026-08-11. The
accepted device artifact uses 280,776 bytes of static RAM and produces a
1,211,620-byte ELF and a 42,434-byte PDX. Soak, memory-growth measurement,
lifecycle stress, and post-run device-log inspection remain unverified.

The implementation now includes one-pole filtering, wavetable and custom
generator synthesis, remaining synth parameters and modulation edges,
signal/controller lookup, and sequence/track/note introspection on both native
adapters. Callback PCM uses four fixed 4,096-frame native rings. Custom synths
use eight fixed userdata/voice slots with independent 4,096-frame rings and
native polyphonic copy semantics. Both expose update-thread Go callbacks while
the native audio thread only consumes bounded rings. Raw C function and userdata
entry points remain intentionally omitted because the Go contracts preserve the
portable behavior.

The focused `examples/callbackpcm` and `examples/generatorsynth` scenes pass
unit tests and audible acceptance in the official Windows SDK 3.1.1 Simulator.
Both were then built with the conservative-GC hard-float device pipeline,
installed through COM3, launched, and audibly accepted on a physical Playdate
on 2026-08-11. The accepted callback scene uses 352,340 bytes of static RAM and
produces a 1,376,984-byte ELF and 57,677-byte PDX. The accepted generator scene
uses 416,012 bytes of static RAM and produces a 1,463,812-byte ELF and
61,781-byte PDX. P9.3 implementation and focused audible acceptance are
complete; soak, memory-growth measurement, lifecycle stress, and a final
post-run log check roll into P9.4 acceptance.

SDK 3.1.1 declares `setMP3StreamSource` in the C header but provides no matching
official documentation or example establishing a shipping-device contract.
P9.3 therefore does not expose it from declaration discovery alone; packaged
MP3 playback remains available through `FilePlayer`. Reconsider a bounded byte
stream only when official device behavior can be probed and ordinary
file/sample players cannot preserve the required game behavior.

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
