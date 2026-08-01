# gopdsdk

An independent Go SDK and toolchain for building Playdate applications.

The **P0 foundation and P1.0 through P1.3 playable vertical slices are
complete** on the accepted Windows profile. P1.4 is the `v0.1.0` release
candidate: its public API is reviewed and documented, but remains pre-v1. The official Playdate C API is the
normative source; third-party
projects, including pdgo, may be studied only as behavioral and product
references. Their implementation is not copied.

The target matrix is Windows, macOS, and Linux. Host policy selects `.dll`,
`.dylib`, or `.so`, official SDK tool names, native compiler candidates, and
Simulator layout. GitHub Actions executes the pure Go and
external-consumer CLI suite natively on all three hosts. Windows is additionally verified with the
official SDK, Simulator, GNU Arm toolchain, and a physical Playdate. macOS and
Linux SDK/Simulator/device execution remain explicitly unverified.

The exact P1.0 verified toolchain profile is Go 1.26.5, Playdate SDK 3.1.1,
TinyGo 0.41.1, and Arm GNU Toolchain GCC 15.3.1. Other versions are not rejected
solely by version number: `doctor` reports them as `UNVERIFIED` until the
relevant probe succeeds.

## Support status

| Host | Native CI | Official SDK/Simulator | Device build/deploy |
| --- | --- | --- | --- |
| Windows | passing | verified with SDK 3.1.1 | verified on physical hardware |
| macOS | passing | unverified | unverified |
| Linux | passing | unverified | unverified |

Native CI proves Go behavior, path policy, CLI composition, and external-module
consumption. It does not prove that an official Simulator starts or that USB
deployment works on that host.

## Requirements

- Go 1.26.x for development, CLI use, tests, and Simulator compilation.
- The official Playdate SDK for packaging, Simulator runs, and device tools.
- A native C compiler supported by `doctor` for Simulator builds.
- TinyGo 0.41.1 and GNU Arm Embedded 15.3.1 for the verified P1.0 device build.

Set `PLAYDATE_SDK_PATH` when the SDK is outside its conventional host location.
TinyGo and the Arm toolchain are unnecessary for Simulator-only development.

## Install the release candidate

After the `v0.1.0` tag is published, run the CLI directly at that version:

```sh
go run github.com/Djunichi/gopdsdk/cmd/gopdsdk@v0.1.0 doctor
go run github.com/Djunichi/gopdsdk/cmd/gopdsdk@v0.1.0 init --module example.com/my-game ./my-game
cd my-game
go mod tidy
```

The tagged CLI creates a project requiring the same module version without a
local `replace`. A development CLI built from a checkout intentionally writes a
local replacement instead. Until the tag exists, use the checkout workflow
below. See [API.md](API.md), [COMPATIBILITY.md](COMPATIBILITY.md), and
[RELEASING.md](RELEASING.md).

## Environment diagnostics

Run the read-only environment check:

```sh
go run ./cmd/gopdsdk doctor
```

If the Playdate SDK is not in a conventional location, set the official
`PLAYDATE_SDK_PATH` environment variable or provide it explicitly:

```sh
go run ./cmd/gopdsdk doctor --sdk /path/to/PlaydateSDK
```

The command assesses SDK discovery, development, Simulator compilation, device
compilation, and device deployment separately. A tool reported as `UNVERIFIED`
was found but has not yet passed an end-to-end probe.

Run the Simulator and device-build probes during diagnostics with:

```sh
go run ./cmd/gopdsdk doctor --probe
```

This verifies both packaging pipelines but intentionally does not install or run
anything on a connected Playdate; `device-deploy` remains a separate capability.

Run the native Simulator toolchain probe with:

```sh
go run ./cmd/gopdsdk probe simulator --sdk /path/to/PlaydateSDK
```

The probe builds a temporary Go shared library, verifies the required
`eventHandler` export, packages a `.pdx`, and removes all temporary artifacts.

Launch the packaged probe in Playdate Simulator and verify `kEventInit`, the
update callback, and a Hello World `drawText` call with:

```sh
go run ./cmd/gopdsdk probe simulator --run --sdk /path/to/PlaydateSDK
```

The automated probe intentionally terminates the Simulator process it launches
after verification succeeds or the timeout expires.

Verify the first device compilation stage with:

```sh
go run ./cmd/gopdsdk probe device
```

This currently proves a hard-float Cortex-M7 TinyGo object, the official
Playdate `setup.c` and `link_map.ld` link, an ELF32/ARM executable with
`eventHandlerShim`, a one-time TinyGo runtime bootstrap, no unresolved symbols,
and conversion of `pdex.elf` into a packaged `pdex.bin` with the official `pdc`.
Device builds default to TinyGo conservative GC with a checked 256 KiB heap,
bounded-memory validation, and deterministic fail-stop OOM behavior.
Deployment and physical-device execution are proven on the Windows P1.0 setup.
The marker was rendered after the Go event handler returned.

Build an importable application package for a physical Playdate with:

```sh
go run ./cmd/gopdsdk build device --sdk /path/to/PlaydateSDK ./examples/hello
```

The default output is `build/<package>.pdx`; `--output` selects another path and
`--force` replaces an existing output.

After the read-only connection probe succeeds, explicitly install the verified
probe package with:

```sh
go run ./cmd/gopdsdk probe device --install --sdk /path/to/PlaydateSDK
```

`--install` changes the connected device by copying the probe game. It does not
run the game; start `gopdsdk Device Probe` manually on the Playdate.

Build, install, and launch the verified device probe in one command with:

```sh
go run ./cmd/gopdsdk run device --sdk /path/to/PlaydateSDK ./examples/hello
```

This changes the connected device, then asks the official `pdutil` to launch
the installed package. The existing `gopdsdk run <package>` form continues to
build and launch a Simulator application.

Safely check whether `pdutil` can open a connected Playdate without mounting,
installing, or running anything:

```sh
go run ./cmd/gopdsdk probe connection --sdk /path/to/PlaydateSDK
```

Connect and unlock the Playdate over USB before running the probe. A successful
probe verifies communication only; it does not modify the device or prove that
the packaged game runs.

Mount the connected Playdate data disk and print its `crashlog.txt` directly to
the console with:

```sh
go run ./cmd/gopdsdk crashlog --sdk /path/to/PlaydateSDK
```

The crash log contents are written to stdout, so they can also be redirected to
a file. The resolved source path is written to stderr.

## Build a Simulator application

Create an independent starter project during P0 with:

```sh
go run ./cmd/gopdsdk init --module example.com/my-game --author "Your Name" --bundle-id com.example.my-game ./my-game
```

When run from a development checkout, the generated `go.mod` uses a local
`replace` directive. A tagged CLI uses its own published module version without
`replace`. Run `go mod tidy` once to resolve that dependency and create
`go.sum` before the first build. The command never overwrites an existing path.

An application is an importable Go package that provides
`New() playdate.Game` and an official Playdate `pdxinfo` file in the same
directory. Source assets live only below `resources/`; its contents become the
root of the packaged PDX. Build the included Hello World example on Windows with:

```sh
go run ./cmd/gopdsdk build --sdk /path/to/PlaydateSDK ./examples/hello
```

The default output is `build/hello.pdx`. Use `--output` to select another path.
The build command does not overwrite an existing artifact. Simulator execution
on macOS and Linux remains unverified; device builds use the accepted P1.0
TinyGo conservative runtime.

Inspect either build without running compilers or SDK tools:

```sh
go run ./cmd/gopdsdk build --dry-run --sdk /path/to/PlaydateSDK ./examples/hello
go run ./cmd/gopdsdk build device --dry-run --sdk /path/to/PlaydateSDK ./examples/hello
```

Dry-run output is a typed semantic plan with structured executable arguments,
portable `${WORK}` and `${PACKAGE}` tokens, and explicit artifact retention.
Temporary workspaces are marked `cleanup`; published `.pdx` outputs are marked
`preserve`. Cleanup rejects unresolved, relative, and filesystem-root paths.

Build and launch the example, replacing its previous build artifact, with:

```sh
go run ./cmd/gopdsdk run --sdk /path/to/PlaydateSDK ./examples/hello
```

The command leaves the Simulator process running. `Process.Kill` is used only by
the automated probe command.

## P1.1 lifecycle and input parity

P1.1 is complete on the verified Windows profile. The same game code received
matching button transitions, crank state, dock state, frame delta, and ordered
pause/resume and lock/unlock callbacks in Simulator and on a physical Playdate.
The device reported an approximately 33.40 ms frame delta and completed the
required 60-second regression soak. Low-power and terminate routing are covered
by deterministic pure-Go tests and the common ABI event path; those two events
were not separately induced during physical acceptance.

The `examples/lifecycleinput` game exercises the same public lifecycle and
per-frame input snapshots on Simulator and device. Build or run that single
package on both targets:

```sh
go run ./cmd/gopdsdk run --sdk /path/to/PlaydateSDK ./examples/lifecycleinput
go run ./cmd/gopdsdk run device --sdk /path/to/PlaydateSDK ./examples/lifecycleinput
```

The display reports an ordered lifecycle trace and counters; current, pressed,
released, held, and latched edge button masks; crank angle and change; dock
transitions; frame delta; and the soak marker. Pure-Go tests supply fixed input
and lifecycle sequences to this same game implementation.

## P1.2 bitmap acceptance

P1.2 is complete on the verified Windows profile. The same public bitmap API
loaded, created, filled, measured, drew, scaled, and explicitly closed native
resources in Simulator and on a physical Playdate. Owned and borrowed handles,
double-close, use-after-close, invalid arguments, and platform-independent
errors have deterministic pure-Go coverage. Device acceptance used the
hard-float bridge and produced no new `errorlog.txt` or crashlog entry.

Application packages keep source assets below a dedicated `resources`
directory. Its contents become the PDX resource root:

```text
game/
  game.go
  game_test.go
  pdxinfo
  resources/
    images/
    audio/
    fonts/
    data/
```

For example, `resources/images/player.png` is compiled into the PDX as
`images/player.pdi` and loaded with `LoadBitmap("images/player")`. Files outside
`resources` are never copied into the package.

The `examples/bitmap` game loads a packaged 64x64 bitmap, creates and fills an
owned bitmap, draws both bitmaps, draws a scaled copy, reports the loaded
dimensions, and closes both resources on termination. Run the same package on
both targets:

```sh
go run ./cmd/gopdsdk run --sdk /path/to/PlaydateSDK ./examples/bitmap
go run ./cmd/gopdsdk run device --memory conservative --sdk /path/to/PlaydateSDK ./examples/bitmap
```

The accepted display contains `PDX: 64x64`, a PASS line, the packaged icon at
two scales, and a solid square created at runtime. Pure-Go tests verify the
operation order and one-time ownership cleanup.

The public package remains a single `playdate` import but is organized by
domain (`application`, `lifecycle`, `input`, `graphics`, `bitmap`, and
`errors`). `playdate.Context` composes the narrower `System`, `Graphics`, and
`InputReader` capabilities so application helpers can depend on only the API
surface they use.

## P1.3 external playable consumer

The P1.3 acceptance fixture is copied into a temporary directory and compiled
as a separate Go module with a local `replace` to gopdsdk. It is a small
crank-controlled catch game using buttons for catch/nudge actions, lifecycle
pause/resume, frame delta, and two packaged `resources/images` bitmaps. Gameplay
is a pure-Go state machine and rendering is derived as a deterministic draw
plan. Tests verify the 60-second completion marker, partial-initialization
rollback, and one-time termination cleanup.

The external-consumer CLI acceptance test compiles the fixture and inspects
both Simulator and device dry-run plans. On the accepted Windows profile, that
same fixture has also been built and launched in the official Simulator and
built, installed, and launched on a physical Playdate. The device completed a
65-second conservative-GC soak; `errorlog.txt` remained 142 bytes with its
pre-run timestamp and `crashlog.txt` remained 4337 bytes with its pre-run
timestamp, so the run produced no new log entry.

Repeat the physical acceptance procedure after relevant runtime or toolchain
changes:

1. Run the consumer in Simulator and exercise crank, A, d-pad, pause, and resume.
2. Run that same package on device for at least 60 seconds until `PASS` appears.
3. Mount the data disk and confirm that `errorlog.txt` and `crashlog.txt` have no
   new entry from the acceptance run.

No public API or subsystem was added for P1.3.

## P1.4 release candidate

P1.4 prepares `v0.1.0`: a version-aware `init` workflow, reviewed public API
contract, compatibility/evidence matrix, and reproducible release gates. It
does not add sprites, audio, collision, animation, fonts, framebuffer access,
or a resource manager. The version becomes consumable without `replace` only
after the reviewed commit is tagged and the tag is published. The candidate
passed Windows Simulator/device-build probes and a repeated 65-second physical
device soak without changing either device log.

## Development and CI

Run the repository checks with:

```sh
gofmt -w cmd examples internal playdate
go test ./...
go vet ./...
git diff --check
go run ./cmd/gopdsdk doctor
```

`go test ./...` includes a CLI acceptance test that builds `gopdsdk`, creates a
standalone module in a path containing spaces, compiles it through its local
`replace`, and requests both Simulator and device dry-run plans from inside the
consumer module. GitHub Actions repeats this suite on Windows, macOS, and Linux;
Linux additionally runs the race detector.

Docker is not part of the P0 matrix. It would provide another Linux environment,
not Windows or macOS semantics. A pinned Linux image becomes useful when it can
legally receive the official SDK and exercise the real Simulator/device build
toolchain without pretending to verify GUI or USB behavior.

## Current limitations

- Device execution uses a single-threaded TinyGo subset: goroutines, channels,
  `select`, public reflection, and finalizers are rejected.
- OOM and panic are deterministic fail-stop traps, not recoverable errors.
- Device `defer`/`recover` is rejected because TinyGo 0.41.1 accesses an ARM
  system register unavailable to Playdate applications through that path.
- macOS and Linux official SDK integration remains unverified.
- Graphics currently cover clear/text and the P1.2 bitmap slice. Sprites,
  animation, fonts, arbitrary framebuffer access, and audio remain unavailable.
