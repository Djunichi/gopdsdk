# Product roadmap

Status: `v0.11.0` is released. Bounded device Go enablement is allocated to
`v0.12.0`; the stable contract then ships as `v1.0.0`, updated 2026-08-18.

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
| P12 | `v0.12.0` | Bounded device Go enablement and replacements accepted by the completed audit | Planned |
| P13 | `v1.0.0` | API freeze, compatibility contract, and release evidence | Planned |

## Remaining capability milestones

- After P11, every materially useful offline official-SDK capability should be
  implemented or replaced without loss of game functionality.
- P12 implements only enablement and bounded replacements accepted by the
  completed device Go-profile audit.
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

## Device Go-profile decisions

The completed device audit is retained as a decision record rather than a release
milestone. Implementation work appears in P12 only where the audit accepted
enablement or a bounded replacement.

| Audit track | Probe | Accepted decision | Implementation consequence |
| --- | --- | --- | --- |
| A.0 | Baseline | `supported`; control ELF/PDX passed Simulator and physical-device acceptance | Complete; use this sequential `scheduler none` result as the control for every remaining track. |
| A.1 | Goroutine | `replaced`; no safe scheduler profile fits the Playdate callback lifecycle and bounded frame execution | Complete; P12.1 `playdate/schedule` owns the bounded stackless replacement. |
| A.2 | Channel | `replaced`; an immediately completing buffered subset passed physical-device acceptance, but any blocking operation fail-stops without a scheduler | Complete; P12.1 owns explicit bounded queue and task-coordination semantics without hidden blocking. |
| A.3 | `select` | `replaced`; ready/default operations passed physical-device acceptance, but blocking fail-stops and TinyGo always favors the first ready case | Complete; P12.1 owns explicit multi-queue polling with a documented deterministic policy. |
| A.4 | `defer` | `supported`; normal-return cleanup passed physical-device acceptance with the conservative SCB shadow, while fail-stop panic does not unwind deferred calls | Complete; P12.2 enables the verified subset and documents dynamic-defer allocation and panic boundaries. |
| A.5 | Panic | `supported`; an intentional physical-device panic produced the verified fail-stop trap and a symbolized crash record | Complete; panic is a terminal trap without message output, stack unwinding, deferred cleanup, or recovery. |
| A.6 | `recover` | `unsupported`; the accepted trap profile cannot unwind, while an experimental print profile silently diverged from Go for indirect recovery and `panic(nil)` | Complete; keep the runtime-symbol rejection, use explicit error returns, and reconsider only after the upstream runtime provides compatible semantics. |
| A.7 | Public reflection | `supported` for the explicitly verified inspection, extraction, conversion, and mutation subset; dynamic invocation and construction remain `unsupported` | Complete; P12.3 replaces the blanket symbol rejection with exact gates, documents the footprint, and proves allocation and memory-growth bounds. |
| A.8 | Finalizer | `unsupported`; conservative-GC `runtime.SetFinalizer` is a verified no-op and never delivers the callback | Complete; preserve the symbol rejection and require explicit resource ownership and lifecycle cleanup. |
| A.9 | cgo | `unsupported` for gopdsdk game packages; isolated `package main` compiles, but the generated-main application model reproducibly crashes the TinyGo interpreter | Complete; retain isolated compilation as diagnostic evidence and use gopdsdk-owned native bridges rather than an application cgo contract. |
| A.10 | `runtime.GC` | `supported`; the repository GC stress scene proved collection, heap reuse, bounded live memory, and sub-budget pauses in a 60-second physical-device soak | Complete; retain the existing production profile and document that pause and memory bounds remain workload-specific. |
| A.11 | `time` fixture | `supported` for the verified pure duration/value subset; `time.Now`, `Sleep`, timers, and tickers are `unsupported` because the inherited QEMU clock is synthetic and scheduler timers fail-stop | Complete; P12.1 owns Playdate-clock timed tasks and P12.2 enables the verified formatting path that retains normal-return defer. |
| A.12 | `fmt` fixture | `replaced`; basic print paths passed physically, but the dynamic `any` API retains unusable panic recovery and reflection at substantially higher cost than the verified `strconv` alternative | Complete; keep the symbol rejection and use typed `strconv`, bounded buffers, builders, and writers. |
| A.13 | `encoding/json` fixture | `replaced`; typed marshal/unmarshal passed physically, but the package retains unusable panic recovery and broad reflection at far higher cost than `playdate/json` | Complete; keep the standard-package symbol rejection and use the existing bounded reflection-free replacement. |

## P12 — bounded device Go enablement — `v0.12.0`

P12 implements only facilities that the completed device Go-profile audit either
accepted for enablement or identified as a material game-development need with a
bounded replacement. Every item must trace back to an accepted audit decision;
speculative language or standard-library expansion remains out of scope.

### P12.1 — `playdate/schedule`

Complete.

Provide a stackless cooperative task scheduler for work that must be spread
across update frames. It is a replacement for the useful game-development
cases of goroutines, channels, and `select`, not an implementation of those Go
facilities, parallelism, preemption, or transparent async functions.

- Run user task steps only from the application update boundary; native audio,
  network, and system callbacks may enqueue bounded records but never execute
  scheduled user code directly.
- Require tasks to yield explicitly by returning from a step. Bound queue
  capacity and work per frame, define deterministic ordering, cancellation,
  completion, overflow, and application-termination behavior, and avoid
  per-frame allocation in the scheduler itself.
- Expose non-blocking bounded queue operations with explicit full and empty
  results; never translate queue occupancy into a hidden wait or panic.
- Support polling across multiple queues with an explicit nothing-ready result
  and a documented deterministic priority or round-robin policy. Never inherit
  source-order bias accidentally or claim Go `select` fairness.
- Treat an elapsed-time budget as an optional secondary guard rather than a
  deterministic substitute for a step bound. Document that one task step can
  still overrun a frame and must remain small.
- Provide wrap-safe delayed and deadline task scheduling driven by
  `Context.CurrentTimeMilliseconds`; it replaces the useful device cases of
  `time.Sleep`, timers, and tickers without blocking the update callback or
  depending on TinyGo's synthetic QEMU clock.
- Prove the public contract with a repository-owned incremental-work consumer,
  pure-Go ordering and lifecycle tests, Simulator integration, device build,
  and physical-device frame-time and bounded-memory evidence.

Implemented on 2026-08-18 with fixed-capacity task storage, FIFO frame ordering,
generation-bearing cancellation, explicit completion and yielding, wrap-safe
delays/deadlines/repeats, a secondary elapsed-time guard, non-blocking bounded
queues, and deterministic round-robin polling. Pure-Go tests, the official
Windows SDK 3.1.1 Simulator build and launch, and the TinyGo 0.41.1 conservative
device build, COM3 installation, and launch pass for `examples/schedule`. The
final device artifact uses 283,068 bytes of static RAM and produces a
1,188,160-byte ELF and a 48,729-byte PDX. On 2026-08-18 the user confirmed
physical-device `PASS` markers for a two-step peak, exactly 40 task steps,
completion in 20 frames, and equal 30-item progress across four tasks. The
physical frame-time, bounded-memory, memory-growth, soak, and unchanged post-run
device-log gates also passed by user confirmation; exact measurement values
were not recorded.

### P12.2 — normal-return `defer`

Complete. The conservative-only linked-symbol gate, deterministic semantic
fixtures, A.11 duration parse/format fixture, and repository-owned cleanup and
repeated-defer soak consumer are implemented. Pure-Go tests, official Windows
SDK 3.1.1 Simulator build and launch, and the TinyGo 0.41.1 conservative device
build pass. On 2026-08-18 a user-provided Simulator screenshot confirmed
`PASS` for defer semantics, 1,288 cleanup calls, duration parsing and
formatting, and the current memory bound; its `Soak` marker was not yet
complete. The same conservative artifact was installed on COM3 and launched on
a physical Playdate on 2026-08-18. The user confirmed all five physical `PASS`
markers for defer semantics, repeated resource cleanup, duration parsing and
formatting, bounded memory growth, and the completed 60-second soak. The user
also confirmed that the post-run `crashlog.txt` and `errorlog.txt` remained
unchanged.

Enable the physically verified normal-return `defer` subset for conservative
device builds without enabling `recover` or promising panic unwinding.

- Remove the blanket `runtime.setupDeferFrame` rejection only when the device
  bootstrap provides and initializes the verified SCB shadow; continue to
  reject `runtime._recover` and unsupported runtime combinations.
- Preserve normal return, early return, LIFO, argument-evaluation, named-result,
  and repeated-defer behavior with deterministic unit and linked-symbol tests.
- Document that `-panic trap` does not run deferred cleanup and that dynamic
  defer registration, including defer in loops, can allocate. Advise against
  relying on dynamic defer in update hot paths without measurement.
- Add linked-symbol and physical-device coverage for the A.11 duration
  formatting/parsing fixture, documenting its code-size and allocation cost;
  this does not enable `time.Now`, sleep, timers, or tickers.
- Add a repository-owned resource-cleanup consumer, Simulator integration,
  device build, physical-device cleanup behavior, repeated-defer memory-growth
  soak, and unchanged post-run log evidence before declaring enablement ready.

### P12.3 — bounded public reflection subset

Enable only the A.7 operations physically verified under the accepted
device profile; do not claim compatibility with the complete `reflect` package.

- Replace the blanket `reflect.` linked-symbol rejection with exact allowed and
  forbidden operation gates. Continue to reject `Value.Call`, `CallSlice`,
  `MakeFunc`, reflected methods and channels, unsupported function Type APIs,
  and any newly linked unimplemented TinyGo reflection path.
- Cover metadata inspection, field and tag access, `Interface`, the accepted
  numeric conversion, and struct, slice, and map mutation with deterministic
  fixtures and negative linked-symbol tests.
- Document the verified operation list, direct typed and generated alternatives,
  and the measured code-size cost. Do not use public reflection internally where
  a static implementation is practical.
- Measure allocation behavior per accepted operation, repeat representative
  work across update frames, and require Simulator integration, device build,
  physical-device memory-growth soak, and unchanged post-run logs before
  declaring enablement ready.

Additional P12 tracks require another accepted device Go-profile decision.

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
