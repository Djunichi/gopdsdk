# Compatibility and evidence

The declared `v0.11.0` retains the exact toolchain profile accepted by
the published `v0.5.0` baseline.
Other versions are not rejected only because their version differs, but remain
`UNVERIFIED` until the relevant probe and acceptance level passes.

## Verified toolchain

| Component         | Verified version          | Evidence                                                       |
| ----------------- | ------------------------- | -------------------------------------------------------------- |
| Go                | 1.26.5                    | Native tests, vet, CLI and external consumers                  |
| Playdate SDK      | 3.1.1                     | Windows Simulator, packaging, USB deployment and hardware runs |
| TinyGo            | 0.41.1 with LLVM 20.1.1   | Conservative GC, runtime validation and physical-device soak   |
| Arm GNU Toolchain | GCC 15.3.1 from 15.3.Rel1 | Hard-float compile, link, ELF validation and packaging         |

## Host and target matrix

| Host    | Pure Go and CLI                    | Official Simulator      | Device build | Physical device |
| ------- | ---------------------------------- | ----------------------- | ------------ | --------------- |
| Windows | VERIFIED                           | VERIFIED with SDK 3.1.1 | VERIFIED     | VERIFIED        |
| macOS   | CI-tested                          | UNVERIFIED              | UNVERIFIED   | UNVERIFIED      |
| Linux   | CI-tested, including race detector | UNVERIFIED              | UNVERIFIED   | UNVERIFIED      |

`CI-tested` covers native Go behavior, path policy, CLI composition, the
external-module workflow, and deterministic target plans. It does not imply
that official SDK tools, GUI Simulator execution, Arm compilation, USB, or
physical hardware work on that host. Docker does not promote those levels.

The `crashlog` and `errorlog` command routing, supported filenames, mount-output
parsing, and missing-tool failures are unit- and external-consumer CLI-tested.
The shared implementation reads the requested root-level log without modifying
it.

The `playdate/diagnostics` collector is unit-tested for bounded frame
aggregation, percentiles, maximum, signed live-heap growth, owned
native-resource extrema, invalid limits, and completed-collection behavior. It
does not itself prove target timing or memory behavior.

Run `gopdsdk doctor` for discovery and `gopdsdk doctor --probe` for current-host
SDK integration. A successful probe applies only to the capability and host it
actually exercised.

The v0.11.0 release candidate passed formatting, the full Go suite and vet on
Windows, green Windows/macOS/Linux native CI for commit
`e879458213497eefc47c68d99ef59299d197742e`, the Linux race detector, and the
external-consumer workflow. Official Windows SDK 3.1.1 `doctor --probe` passed
Simulator compilation and packaging plus TinyGo conservative hard-float
compile, link, relocation, and packaging on 2026-08-18.

The JSON consumer passed Windows Simulator build and launch, conservative
device build, USB installation, and user-confirmed physical behavior covering
packaged-file reading, bounded decoding, schema lookup, mutation, and bounded
encoding. The scoreboard consumer passed user-confirmed Simulator and physical
failure and sequential-request interaction for all four operation paths.
Successful live responses and termination during a pending live request remain
unverified because they require external configuration for
`dev.gopdsdk.scoreboards`.

On 2026-08-18 the user confirmed the final P11 conservative-GC soak,
bounded-memory and memory-growth checks, and unchanged post-run
`crashlog.txt` and `errorlog.txt`. The residual SDK 3.1.1 header reconciliation
found no additional materially useful offline capability outside the public Go
equivalents and documented intentional omissions accumulated through P1-P11.

The v0.10.0 release passed the full Go suite and vet on Windows,
green Windows/macOS/Linux native CI, the Linux race detector, and official
Windows SDK 3.1.1 `doctor --probe` for Simulator compilation/packaging and
TinyGo conservative hard-float compile, link, relocation, and packaging. On
2026-08-17 both focused consumers passed user-confirmed Simulator interaction.
The current combined artifacts installed and ran on a physical Playdate and
passed the complete P10.1 and P10.2 interaction matrices, required
conservative-GC soak, bounded-memory check, and unchanged post-run
`crashlog.txt` and `errorlog.txt` checks.

The v0.9.0 sound acceptance matrix passed the official Windows SDK 3.1.1
Simulator and conservative hard-float device build. On 2026-08-12 the user
confirmed final physical-device acceptance for routing, headphone states,
samples, wavetable and custom synthesis, callbacks, underruns, lifecycle
cleanup, the required conservative-GC soak, bounded memory growth, and
unchanged post-run `crashlog.txt` and `errorlog.txt`. The release checkout also
passed `doctor --probe` for Simulator c-shared compilation and packaging and
for TinyGo hard-float compilation, link, relocation checks, and packaging.

The v0.8.0 sprite acceptance scene passed the official Windows SDK 3.1.1
Simulator, conservative hard-float device build, USB deployment, and physical
Playdate execution on 2026-08-09. The accepted device artifact uses 283,104
bytes of static RAM and produces a 1,479,780-byte ELF and 73,656-byte PDX. The
user-confirmed physical run covered procedural drawing, continuously delivered
update callbacks, pair-specific collision response, the complete presentation
and query matrix, the required 60-second conservative-GC soak, bounded memory
growth, and unchanged post-run `crashlog.txt` and `errorlog.txt`.
