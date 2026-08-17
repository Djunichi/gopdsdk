# Product roadmap

Status: `v0.10.0` declared; publication is pending. The remaining offline
official-SDK capability gaps are allocated through `v0.12.0`, followed by the
`v1.0.0` contract release, updated 2026-08-17.

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
- **In progress** means at least one vertical slice is implemented while the
  complete scope boundary and release gates remain open.
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
