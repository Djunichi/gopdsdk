# PlayDateSDK

An independent Go SDK and toolchain for building Playdate applications.

The project is currently in **P0: foundation and feasibility**. No public API is
stable yet. The official Playdate C API is the normative source; third-party
projects, including pdgo, may be studied only as behavioral and product
references. Their implementation is not copied.

The target matrix is Windows, macOS, and Linux for both Simulator and
physical-device workflows. Host policy selects `.dll`, `.dylib`, or `.so`, the
official SDK tool names, native compiler candidates, and Simulator layout.
Windows is verified end to end; macOS and Linux execution remains unverified
until run with their official SDK distributions. Public documentation will be
added as the API takes shape; design documents remain internal during P0.

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

Run the Windows Simulator toolchain probe with:

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
Allocation uses `playdate->system->realloc` in a non-collecting P0 mode.
Deployment and physical-device execution are proven on the Windows P0 setup.
The marker was rendered after the Go event handler returned.

Build an importable application package for a physical Playdate with:

```powershell
go run ./cmd/gopdsdk build device --sdk /path/to/PlaydateSDK ./examples/hello
```

The default output is `build/<package>.pdx`; `--output` selects another path and
`--force` replaces an existing output.

After the read-only connection probe succeeds, explicitly install the verified
probe package with:

```powershell
go run ./cmd/gopdsdk probe device --install --sdk /path/to/PlaydateSDK
```

`--install` changes the connected device by copying the probe game. It does not
run the game; start `gopdsdk Device Probe` manually on the Playdate.

Build, install, and launch the verified device probe in one command with:

```powershell
go run ./cmd/gopdsdk run device --sdk /path/to/PlaydateSDK ./examples/hello
```

This changes the connected device, then asks the official `pdutil` to launch
the installed package. The existing `gopdsdk run <package>` form continues to
build and launch a Simulator application.

Safely check whether `pdutil` can open a connected Playdate without mounting,
installing, or running anything:

```powershell
go run ./cmd/gopdsdk probe connection --sdk /path/to/PlaydateSDK
```

Connect and unlock the Playdate over USB before running the probe. A successful
probe verifies communication only; it does not modify the device or prove that
the packaged game runs.

Mount the connected Playdate data disk and print its `crashlog.txt` directly to
the console with:

```powershell
go run ./cmd/gopdsdk crashlog --sdk /path/to/PlaydateSDK
```

The crash log contents are written to stdout, so they can also be redirected to
a file. The resolved source path is written to stderr.

## Build a Simulator application

Create an independent starter project during P0 with:

```sh
go run ./cmd/gopdsdk init --module example.com/my-game --author "Your Name" --bundle-id com.example.my-game ./my-game
```

The generated `go.mod` uses a local `replace` directive until gopdsdk has a
versioned module release. The command never overwrites an existing path.

An application is an importable Go package that provides
`New() playdate.Game` and an official Playdate `pdxinfo` file in the same
directory. Build the included Hello World example on Windows with:

```sh
go run ./cmd/gopdsdk build --sdk /path/to/PlaydateSDK ./examples/hello
```

The default output is `build/hello.pdx`. Use `--output` to select another path.
The build command does not overwrite an existing artifact. Simulator execution
on macOS and Linux remains unverified; device builds use the P0 TinyGo runtime.

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
