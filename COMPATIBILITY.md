# Compatibility and evidence

The released `v0.3.0` baseline uses one exact verified toolchain profile.
Other versions are not rejected only because their version differs, but remain
`UNVERIFIED` until the relevant probe and acceptance level passes.

## Verified toolchain

| Component | Verified version | Evidence |
| --- | --- | --- |
| Go | 1.26.5 | Native tests, vet, CLI and external consumers |
| Playdate SDK | 3.1.1 | Windows Simulator, packaging, USB deployment and hardware runs |
| TinyGo | 0.41.1 with LLVM 20.1.1 | Conservative GC, runtime validation and physical-device soak |
| Arm GNU Toolchain | GCC 15.3.1 from 15.3.Rel1 | Hard-float compile, link, ELF validation and packaging |

## Host and target matrix

| Host | Pure Go and CLI | Official Simulator | Device build | Physical device |
| --- | --- | --- | --- | --- |
| Windows | VERIFIED | VERIFIED with SDK 3.1.1 | VERIFIED | VERIFIED |
| macOS | CI-tested | UNVERIFIED | UNVERIFIED | UNVERIFIED |
| Linux | CI-tested, including race detector | UNVERIFIED | UNVERIFIED | UNVERIFIED |

`CI-tested` covers native Go behavior, path policy, CLI composition, the
external-module workflow, and deterministic target plans. It does not imply
that official SDK tools, GUI Simulator execution, Arm compilation, USB, or
physical hardware work on that host. Docker does not promote those levels.

## Accepted application evidence

- P1.1: common lifecycle, buttons, crank, dock transitions and frame delta on
  Windows Simulator and physical Playdate.
- P1.2: packaged bitmap load/create/draw/scale and explicit ownership cleanup
  on both accepted targets.
- P1.4 candidate: Windows `doctor --probe` promoted Simulator and device-build
  to `READY`; the versioned external consumer repeated physical deployment and
  a 65-second soak with both device logs unchanged.
- P2.1 candidate: sprite creation, display-list membership, bitmap assignment,
  movement, visibility, z-index, explicit close, and rollback are unit-tested
  through both generated adapters. Simulator/device acceptance remains
  unverified for this release candidate.
- P2.2 candidate: collision rectangles, response modes, resolved movement, and
  point/rectangle/overlap queries are unit-tested. Simulator/device collision
  parity remains unverified for this release candidate.
- P2.3 candidate: owned bitmap tables, borrowed frames, delta-time animation,
  fixed frames, pause/resume, validation, and cleanup are unit-tested.
  Simulator/device animation parity remains unverified for this candidate.
- P2.4 candidate: portable sound-effect and single-file-player ownership,
  repeated playback, state, volume, lifecycle pause/resume, rollback, and close
  are unit-tested. Simulator/device parity and the 10-minute physical-device
  soak remain acceptance work until run and observed.
- P2.5 candidate: custom-font loading, selected-font drawing, native text
  measurement, explicit close, and deterministic HUD/pause/game-over/restart
  plans are unit-tested. Windows Simulator/device visual parity passed; a
  longer physical-device memory soak remains acceptance work.
- P3.1 implemented: immediate-mode lines, outlined/filled rectangles,
  ellipses, outlined/filled triangles, solid/XOR/8x8 pattern paint, clipping,
  draw offset, draw modes, forwarding through the application context, and
  portable validation errors are unit-tested through both generated adapters.
  `examples/primitives` passed Windows SDK 3.1.1 Simulator compilation and
  visual execution, hard-float device build, USB deployment on COM3, and
  matching physical Playdate execution on 2026-08-02. The measured artifact
  used 266,556 bytes of static RAM and produced a 29,160-byte PDX. A
  conservative-GC soak and memory-growth measurement remain unverified.
- P3.2 implemented: callback-scoped zero-copy framebuffer access, checked
  callback lifetime, explicit dirty-row aggregation, and drawing into owned
  offscreen bitmaps with context restoration are covered by deterministic
  pure-Go and generated-adapter tests. Crank Caverns exercises both public
  capabilities in its integrated scene. Separate SDK visual execution and
  physical-device acceptance for `examples/framebuffer` remain unverified.
- P3.3 implemented: camera clamping, visible-range tile traversal, bitmap
  placement, configuration ownership, and independent static tile overlap are
  covered by deterministic pure-Go tests. Windows SDK 3.1.1 Simulator visual
  acceptance and conservative-GC physical-device build, USB deployment, input,
  jump, collision, camera, and matching scene execution passed on 2026-08-02.
  The device artifact used 270,908 bytes of static RAM and produced a
  37,440-byte PDX. Soak and memory-growth measurement remain unverified.
- P3.4 assessed: existing acceptance scenes prove explicit single-owner
  rollback and shutdown, but no two real consumers yet share caching and
  scene-transition semantics. Resource ownership therefore remains local to
  each scene; no new runtime, Simulator, or device capability is claimed.
- P3.5 foundation: the `Launcher` capability and its `NewApplication` forwarding
  path are unit-tested in both generated adapters. The `examples/navigation`
  Play/menu-return/Exit scene passed Windows SDK 3.1.1 Simulator interaction,
  including return to the Launcher, on 2026-08-02. Its 1-bit 350x155 card,
  32x32 icon, and 400x240 launch image were packaged through `imagePath`; a
  conservative-GC hard-float build, USB deployment on COM3, and launch command
  passed the same day. Physical Play/menu-return/Exit interaction and correct
  card, icon, and launch-image display in the Launcher were then confirmed on
  the connected Playdate. The OS returned to the Launcher through the official
  exit path; resource-cleanup behavior inside `LifecycleTerminate` was not
  separately observed because this scene owns no native resources.
- P3.5 integrated consumer: the external
  [Crank Caverns](https://github.com/Djunichi/gopdsdkgame) game, currently in a
  private repository, uses only public API and combines lifecycle/input, owned
  graphics resources, sprites and collisions, animation, audio, fonts,
  primitives and graphics state, offscreen drawing, direct framebuffer access,
  tile map and camera, menu transitions, and Launcher exit. Its deterministic
  gameplay and render plans are unit-tested. Fixed frame-time, bounded-heap,
  extended physical-device soak, cleanup observation on termination, and
  post-run device-log comparison remain unverified for the integrated game.
- P4.1 implemented: optional filesystem capability forwarding, owned file
  lifetime, read/write/seek/flush/close behavior, path and mode validation,
  metadata, listing, and directory mutations are covered by deterministic
  runtime and generated Simulator/device adapter tests. The focused
  `examples/filesystem` flow passed Windows SDK 3.1.1 Simulator and physical
  Playdate execution on 2026-08-02 after a device-only borrowed listing string
  was fixed by copying it before bridge-memory release. The repeated device
  flow displayed `P4.1 filesystem OK`; multi-session durability, interrupted
  writes, soak, and memory-growth measurement remain unverified.
- P4.2 implemented: the portable `playdate/store` layer provides bounded
  versioned values, checksummed corruption detection, explicit stepwise
  migrations, and recoverable temporary/backup replacement. Pure-Go tests cover
  first save, replacement, future versions, interrupted rename, short writes,
  device-style non-overwriting rename, backup recovery, and preservation of the
  last valid value. The `examples/persistence` flow passed Windows SDK 3.1.1
  Simulator and physical Playdate save, migration, replacement, and reload on
  2026-08-02. The first hardware run exposed non-overwriting device `rename`;
  the backup-swap correction then passed the conservative-GC gate at 267,116
  bytes of static RAM and a 30,269-byte PDX, USB deployment, and physical
  execution with `P4.2 STORE OK`. Cross-launch durability, injected power loss,
  soak, and memory-growth measurement remain unverified.
- P4.3 implemented: optional owned action, checkmark, and option System Menu
  items and localization are covered by deterministic ownership, forwarding,
  generated Simulator/device adapter, and lifecycle cleanup tests. The focused
  `examples/systemmenu` consumer persists localized menu values through P4.2.
  On 2026-08-02 it passed Windows SDK 3.1.1 Simulator execution, the
  conservative device gate at 268,932 bytes of static RAM and a 35,674-byte
  PDX, USB deployment, and physical Playdate execution. Both menu callbacks
  changed their settings and the values survived a game restart. Extended
  conservative-GC soak, memory-growth measurement, and post-run crashlog
  inspection remain unverified.
- P4.4 implemented: optional accelerometer, power-monitor, and read-only system-
  preference capabilities are forwarded through `NewApplication`; termination
  disables an enabled accelerometer, and the existing `Launcher` capability is
  covered in the same deterministic flow. `examples/systemstatus` passed SDK
  3.1.1 Simulator build and launch and the conservative device gate on
  2026-08-02 at 283,884 bytes of static RAM and a 55,193-byte PDX, including USB
  deployment and physical execution. Hardware confirmed accelerometer, battery,
  volume, timezone, clock format, reduce-flashing, and `NONE`/`USB` power states.
  The device run exposed and verified the correction for direct float-return ABI
  corruption across TinyGo/C. `CHARGE`, `SCREWS`, soak, memory growth, and
  post-run crashlog inspection remain unverified.

Run `gopdsdk doctor` for discovery and `gopdsdk doctor --probe` for current-host
SDK integration. A successful probe applies only to the capability and host it
actually exercised.
