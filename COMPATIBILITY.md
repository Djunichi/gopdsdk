# Compatibility and evidence

The proposed `v0.2.0` release supports one exact verified toolchain profile.
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
- P1.3: an external crank-controlled module using only public API, with real
  Simulator launch, physical-device deployment, a 65-second soak, and no new
  `errorlog.txt` or `crashlog.txt` entry.
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

Run `gopdsdk doctor` for discovery and `gopdsdk doctor --probe` for current-host
SDK integration. A successful probe applies only to the capability and host it
actually exercised.
