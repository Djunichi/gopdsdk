# gopdsdk

An independent Go SDK and toolchain for building Playdate applications.

The latest declared release is **`v0.8.0`**; publication remains pending until
its tag and hosted release exist. It completes the offline sprite and collision
facilities with presentation controls, native tilemaps, detailed queries,
display-list operations, and bounded callbacks. P8 acceptance, soak,
bounded-memory, and post-run device-log checks passed on the verified Windows
Simulator and physical-device profile.
The public API is snapshot-tested and documented,
but remains pre-v1. Hardware evidence varies by feature and is
reported without promotion in [COMPATIBILITY.md](COMPATIBILITY.md). The official Playdate C API is the
normative source; third-party
projects, including pdgo, may be studied only as behavioral and product
references. Their implementation is not copied.

The target matrix is Windows, macOS, and Linux. Host policy selects `.dll`,
`.dylib`, or `.so`, official SDK tool names, native compiler candidates, and
Simulator layout. GitHub Actions executes the pure Go and
external-consumer CLI suite natively on all three hosts. Windows is additionally verified with the
official SDK, Simulator, GNU Arm toolchain, and a physical Playdate. macOS and
Linux SDK/Simulator/device execution remain explicitly unverified.

The exact `v0.8.0` verified toolchain profile is Go 1.26.5, Playdate SDK 3.1.1,
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
- TinyGo 0.41.1 and GNU Arm Embedded 15.3.1 for the verified device build.

Set `PLAYDATE_SDK_PATH` when the SDK is outside its conventional host location.
TinyGo and the Arm toolchain are unnecessary for Simulator-only development.

## Install v0.8.0 after publication

After the `v0.8.0` tag is published and the module-proxy check passes, run the
released CLI directly at that version:

```sh
go run github.com/Djunichi/gopdsdk/cmd/gopdsdk@v0.8.0 doctor
go run github.com/Djunichi/gopdsdk/cmd/gopdsdk@v0.8.0 init --module example.com/my-game ./my-game
cd my-game
go mod tidy
```

The tagged CLI creates a project requiring the same module version without a
local `replace`. A development CLI built from a checkout intentionally writes a
local replacement instead; use the checkout workflow below when developing
from source. See [API.md](API.md), [COMPATIBILITY.md](COMPATIBILITY.md),
[MIGRATING.md](MIGRATING.md), and [RELEASING.md](RELEASING.md). The remaining
path to `v1.0.0`, including post-release multiplayer research, is recorded in
[docs/ROADMAP.md](docs/ROADMAP.md).

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
Deployment and physical-device execution are proven on the verified Windows setup.
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

Connect and unlock the Playdate, then mount its data disk and print
`crashlog.txt` directly to the console with:

```sh
go run ./cmd/gopdsdk crashlog --sdk /path/to/PlaydateSDK
```

Retrieve `errorlog.txt` through the same flow with:

```sh
go run ./cmd/gopdsdk errorlog --sdk /path/to/PlaydateSDK
```

Both commands read but do not modify the selected log. Log contents are written
to stdout, so they can be redirected to a file, and the resolved source path is
written to stderr. Mounting changes the connected Playdate into data-disk mode;
neither command proves that a game ran successfully.

## External-game measurements

`playdate/diagnostics` aggregates a bounded run without allocating while
samples are recorded. External games use the runtime-provided frame delta,
sample live heap from their Go runtime, and count the native resources they
explicitly own:

```go
collector, _ := diagnostics.New(1800) // one minute at 30 FPS
// update and render the frame
var memory runtime.MemStats
runtime.ReadMemStats(&memory)
_ = collector.Record(diagnostics.Sample{
	FrameMilliseconds: uint32(ctx.Input().DeltaSeconds*1000 + 0.5),
	HeapBytes:          memory.HeapAlloc,
	NativeResources:    ownedResources,
})
```

After the interval, `Report` returns frame-time mean/p50/p95/p99/max, heap
start/end/max/growth, and resource start/end/min/max. Render or persist that
report after measurement so diagnostic I/O does not distort the samples. Use
`gopdsdk build device` for static RAM, ELF, and packaged PDX sizes, and label
Simulator and physical-device measurements separately.

## Build a Simulator application

Create an independent starter project with the CLI using:

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
on macOS and Linux remains unverified; device builds use the verified
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

The CLI carries its project-owned Simulator and device ABI bridge sources as
package-owned `go:embed` assets. It materializes those version-matched sources
only inside the temporary workspace. Official `pd_api.h`, `setup.c`, and
`link_map.ld` remain external inputs read from the selected Playdate SDK.

Build and launch the example, replacing its previous build artifact, with:

```sh
go run ./cmd/gopdsdk run --sdk /path/to/PlaydateSDK ./examples/hello
```

The command leaves the Simulator process running. `Process.Kill` is used only by
the automated probe command.

## Lifecycle and input parity

Lifecycle and input parity are complete on the verified Windows profile. The same game code received
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

## Bitmap acceptance

Bitmap acceptance is complete on the verified Windows profile. The same public bitmap API
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

Launcher artwork uses the same staging rule. Set, for example,
`imagePath=images/launcher` in `pdxinfo`, then provide:

```text
resources/images/launcher/
  card.png        # 350x155
  icon.png        # 32x32
  launchImage.png # 400x240
```

The compiler places these files at the PDX paths named by `imagePath`. The
official format also permits highlighted/pressed artwork and launch animation
directories; these can be added under the same resource directory.

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

For pixel editors, masks, collision maps, transformed assets, and persistent
screen captures, assert `playdate.BitmapDataGraphics`. Its bitmap-data view is
callback-scoped, while copies, rotations, tables, and display snapshots return
owned resources that the game closes explicitly.

`examples/bitmapdata` is the complete P7.2 acceptance scene. It edits owned
pixels with dirty tracking, copies and reloads bitmaps, creates and reloads a
table, attaches and inspects a mask, performs flipped mask collision, creates a
rotated bitmap, and retains a display-buffer snapshot:

```sh
go run ./cmd/gopdsdk run --sdk /path/to/PlaydateSDK ./examples/bitmapdata
go run ./cmd/gopdsdk run device --memory conservative --install --sdk /path/to/PlaydateSDK ./examples/bitmapdata
```

On 2026-08-08 this scene passed deterministic tests, official Windows SDK
3.1.1 Simulator build and visual execution, conservative hard-float device
build, USB installation through COM3, launch, and visual execution on a
physical Playdate. The accepted device artifact uses 275,580 bytes of static
RAM and produces an 872,964-byte ELF and a 38,768-byte PDX. Extended soak,
memory-growth measurement, and post-run device-log inspection passed on
2026-08-08.

For complete text layout and custom renderers, assert `playdate.TextGraphics`.
It adds bounded aligned text, character and word wrapping, tracking, leading,
wrapping-height measurement, and font-owned borrowed glyph bitmaps with advance
and kerning metrics. Packaged `.fnt` resources remain the portable offline font
source on both Simulator and device.

The public package remains a single `playdate` import but is organized by
domain (`application`, `lifecycle`, `input`, `graphics`, `bitmap`, `audio`, and
`errors`). `playdate.Context` composes narrower capabilities so application
helpers can depend on only the API surface they use.

## Sprites

`examples/sprites` exercises explicitly owned sprites, bitmap assignment,
position and relative movement, visibility, z-index, idempotent display-list
membership, the shared update/draw pass, display presentation controls, and
live logical dimensions, nominal refresh rate, and measured FPS introspection.
Sprites must be closed before the bitmaps they reference.

On 2026-08-08 the P7.4 scene passed matching visual acceptance in the official
Windows SDK 3.1.1 Simulator and on a physical Playdate after conservative-GC
hard-float build and USB installation through COM3. The accepted device
artifact uses 292,016 bytes of static RAM and produces a 69,505-byte PDX.
Bounded-memory, soak, and post-run device-log gates passed on 2026-08-08;
performance regression evidence remains unverified.

```sh
go run ./cmd/gopdsdk run --sdk /path/to/PlaydateSDK ./examples/sprites
go run ./cmd/gopdsdk run device --memory conservative --sdk /path/to/PlaydateSDK ./examples/sprites
```

`examples/spritepresentation` is the P8 sprite acceptance scene. Its P8.1
coverage includes sprite
center, bounds, position and state getters, all image flips, draw mode, opacity,
stencil patterns and images, clipping, draw-offset policy, and independent
update/collision enable state. It also attaches an explicitly owned native
tilemap to a sprite and alternates one tile while reporting its native getter.
At startup it additionally self-checks the P8.2 line and detailed-hit queries,
non-mutating collision checks, sprite count, bulk add/remove, remove-all state,
and collision-world reset. It also renders a procedural sprite through a draw
callback, counts ordered update callbacks, and verifies a pair-specific bounce
response before displaying `P8.3 PASS`.
Press A or B to toggle the enable states, Up to
clear or restore both stencils, Down to clear or restore clipping, and Left or
Right to demonstrate the difference between offset-following and
offset-ignoring sprites.

On 2026-08-08 this scene passed visual acceptance in the official Windows SDK
3.1.1 Simulator and on a physical Playdate. The conservative-GC hard-float
device artifact uses 281,880 bytes of static RAM and produces a 1,304,340-byte
ELF and 61,792-byte PDX. The final P8 device acceptance covers the required
conservative-GC soak, bounded memory growth, and unchanged post-run device
logs.

On 2026-08-09 the expanded P8.2 startup self-check and visible PASS status
succeeded in the official Windows SDK 3.1.1 Simulator and on a physical
Playdate. The conservative-GC hard-float device artifact uses 282,512 bytes of
static RAM and produces a 1,398,940-byte ELF and 69,893-byte PDX. The final P8
device acceptance covers the required conservative-GC soak, bounded memory
growth, and unchanged post-run device logs.

On 2026-08-09 the P8.3 build displayed `P8.3 PASS` in the official Windows SDK
3.1.1 Simulator. Its startup pair-specific collision-response check succeeded,
and the visible update-callback counter increased continuously. The conservative
hard-float device build uses 283,104 bytes of static RAM and produces a
1,479,780-byte ELF and a 73,656-byte PDX. USB installation through COM3 and
launch passed; user-confirmed physical execution showed matching `P8.3 PASS`, a
continuously increasing update counter, and the procedural draw. A
user-confirmed 60-second conservative-GC physical-device soak passed on
2026-08-09. The user also confirmed bounded memory growth and unchanged
post-run `crashlog.txt` and `errorlog.txt` for the accepted device run.

```sh
go run ./cmd/gopdsdk run --sdk /path/to/PlaydateSDK ./examples/spritepresentation
go run ./cmd/gopdsdk run device --memory conservative --install --sdk /path/to/PlaydateSDK ./examples/spritepresentation
```

## Collisions

`examples/collision` is a deterministic collision scene using collide
rectangles, slide/freeze/overlap/bounce responses, resolved movement, and
point/rectangle/overlap queries. The SDK additionally exposes non-mutating
collision checks, simple and detailed line queries, sprite count, bulk
display-list membership, remove-all, and collision-world reset. Portable
results contain ordered collision or line-intersection geometry without
exposing native pointers.

```sh
go run ./cmd/gopdsdk run --sdk /path/to/PlaydateSDK ./examples/collision
go run ./cmd/gopdsdk run device --memory conservative --sdk /path/to/PlaydateSDK ./examples/collision
```

## Bitmap-table animation

`examples/animation` loads a packaged bitmap table and selects borrowed frames
with the allocation-free `Animation` helper. It supports delta-time looping,
fixed frames, and pause/resume while retaining partial-frame time.

```sh
go run ./cmd/gopdsdk run --sdk /path/to/PlaydateSDK ./examples/animation
go run ./cmd/gopdsdk run device --memory conservative --sdk /path/to/PlaydateSDK ./examples/animation
```

## Base audio

The base audio implementation established two deliberately narrow vertical APIs: a memory-backed short
sound effect and one streaming file/music player. Both expose stereo volume,
stopped/playing/paused status, lifecycle pause/resume, and explicit close.
`examples/audio` now retains those paths while also serving as the advanced sample
advanced-audio acceptance game described below.

Run the same package on both targets:

```sh
go run ./cmd/gopdsdk run --sdk /path/to/PlaydateSDK ./examples/audio
go run ./cmd/gopdsdk run device --memory conservative --sdk /path/to/PlaydateSDK ./examples/audio
```

The original implementation and generated ABI remain regression-tested.
Advanced synthesis and microphone input remain optional capabilities outside
that base slice.

## Fonts and game UI

`examples/fontsui` is the complete P7.3 acceptance scene. It loads
`resources/fonts/gopdsdk-ui.fnt` as `fonts/gopdsdk-ui`, configures tracking and
leading, measures wrapped height, reads glyph advance and pair kerning, and
draws centered word-wrapped text inside a bounded rectangle. Its pure
`LayoutPlan` produces stable HUD, score, pause, and game-over commands; A
increments score or restarts after game over, B ends the run, and lifecycle
pause/resume selects the pause screen. Glyph bitmaps borrow the font lifetime,
and the owned font is closed on termination.

Run the same example on either target with:

```sh
go run ./cmd/gopdsdk run --sdk /path/to/PlaydateSDK ./examples/fontsui
go run ./cmd/gopdsdk run device --memory conservative --install --sdk /path/to/PlaydateSDK ./examples/fontsui
```

On 2026-08-08 deterministic tests, Windows SDK 3.1.1 Simulator visual
acceptance, and physical-device installation, launch, and visual acceptance
passed with bounded wrapping, matching custom-font HUD, score, pause,
game-over, and restart screens. The conservative hard-float artifact uses
278,688 bytes of static RAM and produces an 826,340-byte ELF and a 41,203-byte
PDX. Device soak, memory-growth measurement, and post-run log inspection passed
on 2026-08-08.

## Drawing primitives and state

`examples/primitives` is the consumer-driven acceptance scene for immediate
mode lines, outlined and filled rectangles, ellipses, outlined and filled
triangles and polygons, rounded rectangles, solid/XOR/8x8 pattern paint, line
caps, background color, local and screen clipping, draw offset, and bitmap draw
modes. Games opt into the narrow `playdate.PrimitiveGraphics` and
`playdate.GraphicsState` capabilities instead of expanding every `Graphics`
fake.

Run the same scene on either target with:

```sh
go run ./cmd/gopdsdk run --sdk /path/to/PlaydateSDK ./examples/primitives
go run ./cmd/gopdsdk run device --memory conservative --sdk /path/to/PlaydateSDK ./examples/primitives
```

Deterministic unit tests, Windows SDK 3.1.1 Simulator compilation and visual
execution, hard-float device build, USB deployment, and matching physical
Playdate execution passed on 2026-08-02. The accepted device artifact used
266,556 bytes of static RAM and produced a 29,160-byte PDX. Conservative-GC
soak and memory-growth measurement remain unverified.

## Framebuffer and offscreen drawing

Native contexts now expose narrow `playdate.FramebufferGraphics` and
`playdate.OffscreenGraphics` capabilities. Framebuffer access is callback
scoped and zero-copy, combines explicitly reported dirty rows, and rejects
checked access after the callback. Direct byte mutations require an explicit
`MarkDirtyRows` call. `DrawInto` accepts owned bitmaps only and always restores
the previous drawing context before returning.

The portable pixel layout, validation, dirty-range aggregation, generated
Simulator bridge, and both device memory profiles are unit-tested. SDK
integration, Simulator visual execution, device build, USB deployment, and
physical-device execution for this capability remain unverified.

## Tile map and camera

`playdate.TileMap` renders only cells intersecting an integer-pixel
`playdate.Camera`; `TileDrawStats` makes that per-frame bound observable.
Tile zero is empty and other values index caller-owned bitmaps. Static solid
tile overlap is available through `IntersectsSolid` and deliberately remains
separate from sprite collision and movement response.

The deterministic unit suite covers camera clamping, visible-range work,
screen-space placement, copied configuration, validation, draw failures, and
static collision edges. `examples/tilemap` is the consumer-driven vertical
slice. Windows SDK 3.1.1 Simulator visual acceptance and conservative-GC
physical-device build, USB deployment, controls, jump, collision, camera, and
matching scene execution passed on 2026-08-02. The accepted device artifact
used 270,908 bytes of static RAM and produced a 37,440-byte PDX. Soak and
memory-growth measurement remain unverified.

```sh
go run ./cmd/gopdsdk run --sdk /path/to/PlaydateSDK ./examples/tilemap
```

## Repeated resource ownership

Resources remain explicitly owned by the scene that loads or creates them.
Initialization closes already acquired resources when a later step fails, and
termination closes each owned resource once in dependency-safe order. Borrowed
handles, such as bitmap-table frames, are never independently closed.

The resource-ownership audit found no two real consumers with matching cache and transition
semantics, so the SDK does not yet expose a resource manager or reference
counting API. The integrated acceptance scene is the next consumer evidence; an
abstraction will be added only after another real scene repeats its complete
loading, caching, rollback, transition, and shutdown policy.

## Integrated acceptance game

The external [Crank Caverns](https://github.com/Djunichi/gopdsdkgame) game
completes the integrated consumer slice. Its repository is currently private, but the
link is retained as the canonical acceptance-game reference. It uses only the
public `gopdsdk` API and integrates lifecycle and input, crank control, owned
bitmaps, sprites and collision queries, bitmap-table animation, sound effect
and streaming music, custom-font UI, primitives and graphics state, offscreen
drawing, direct framebuffer pixels, tile map and camera.

The game begins at an in-game menu with `Play` and `Exit`.
Gameplay can transition back to that menu without restarting the application.
`Exit` asserts `playdate.Launcher` and calls `ExitToLauncher`; Playdate sends
`LifecycleTerminate` before returning to the Launcher, preserving normal owned
resource cleanup. Both generated native adapters implement and forward this
capability.

`examples/navigation` is the deterministic acceptance scene for this slice:

```sh
go run ./cmd/gopdsdk run --sdk /path/to/PlaydateSDK ./examples/navigation
```

The example includes generated 1-bit launcher artwork at the three baseline
sizes under `resources/images/launcher` and selects it with
`imagePath=images/launcher`.

Windows SDK 3.1.1 Simulator and physical-device acceptance passed on
2026-08-02: `Play`, B-button return to the menu, `Exit` back to the Launcher,
and the packaged card, icon, and launch image all behaved as intended. The
device path used a conservative-GC hard-float build and USB deployment on COM3.
Crank Caverns deterministically tests its gameplay state and render plans and
owns every native resource explicitly. The complete game establishes the
integrated product boundary. Later diagnostics recorded frame-time
distributions, live-heap growth, and stable native-resource counts over bounded
1,800-frame Simulator
and physical-device runs. The `v0.6.0` regression run subsequently covered extended
soak, termination cleanup observation, and post-run device-log comparison; that
later evidence is not implied by the v0.3.0 release.

## Owned filesystem

Games can optionally assert `playdate.FileSystem` from their callback context.
It exposes owned files with Go `Read`, `Write`, `Seek`, `Flush`, and `Close`
behavior, plus `Stat`, non-recursive `List`, `Mkdir`, `Remove`, and `Rename`.
Read options preserve the official distinction between packaged PDX files,
Data files, and Data-first fallback:

```go
files, ok := any(context).(playdate.FileSystem)
if !ok {
	return playdate.ErrFileUnavailable
}

file, err := files.OpenFile(
	"save.bin",
	playdate.FileReadData|playdate.FileReadPackage,
)
if err != nil {
	return err
}
defer file.Close()
```

Paths are game-relative. Native diagnostics are copied into
`playdate.FileOperationError`; callers can use `errors.Is(err,
playdate.ErrFileIO)` without parsing diagnostic text. `rename` follows the
official overwrite behavior and is the atomic replacement primitive used by
the versioned store. The focused `examples/filesystem` flow passed Windows SDK 3.1.1 Simulator
and physical Playdate execution on 2026-08-02, including write, flush, close,
rename, Data read, stat, list, and recursive remove. Multi-session durability,
interrupted replacement, soak, and memory-growth evidence remain unverified.

## Versioned store

`playdate/store` adds bounded, versioned persistence above `FileSystem`. It
writes a binary envelope with a payload checksum to a sibling temporary file,
flushes and closes it, and first attempts the documented overwriting rename.
On hardware that rejects an existing destination, it falls back to a recoverable
`final → backup`, `temporary → final` swap. `Load` recovers the backup if power
was lost between those operations, and never selects a stale temporary file.

```go
save, err := store.New(files, store.Config{
	Path:        "save.bin",
	Version:     2,
	MaximumSize: 4096,
	Migrations: []store.VersionMigration{
		{From: 1, Migrate: migrateVersion1},
	},
})
if err != nil {
	return err
}
payload, err := save.Load()
```

Each migration advances exactly one version and a successfully migrated value
is atomically rewritten at the current version before it is returned. Callers
choose their own payload encoding; the SDK does not impose JSON. Pure-Go unit
tests cover first save, replacement, migrations, future versions, corruption,
size bounds, interrupted rename, short writes, and preservation of the last
valid value.

The `examples/persistence` flow passed Windows SDK 3.1.1 Simulator and physical
Playdate execution on 2026-08-02, displaying `STORE OK` after save,
migration, replacement, and reload. The first hardware run exposed that device
`rename` rejected an existing destination despite the documented overwrite
contract; it preserved both the valid final file and completed temporary file.
The backup-swap fallback was then added, passed the conservative-GC device gate
at 267,116 bytes of static RAM and a 30,269-byte PDX, and passed repeated USB
deployment and physical execution. Cross-launch durability, power-loss injection,
soak, and memory-growth evidence remain unverified.

## System menu and localization

Games can optionally assert `playdate.SystemMenu` to add owned action,
checkmark, and options items. Items expose their SDK title and value, support
idempotent removal, retain callbacks until removal, and are automatically
removed after `LifecycleTerminate`. `playdate.Localization` exposes only the
system language and `.strings` lookup; a missing key is reported explicitly so
fallback text stays in game code.

The `examples/systemmenu` consumer combines these capabilities with the versioned store
store, persisting its checkmark and option values while localizing the visible
labels. On 2026-08-02 it passed Windows SDK 3.1.1 Simulator execution, the
conservative device gate at 268,932 bytes of static RAM and a 35,674-byte PDX,
USB deployment, and physical Playdate execution. The menu callbacks changed
both settings and their values survived a game restart. Extended
conservative-GC soak, memory-growth measurement, and post-run device-log
inspection remain unverified.

## Device and system status

Games can optionally assert `playdate.Accelerometer`, `playdate.PowerMonitor`,
and `playdate.SystemPreferences`. Accelerometer sampling requires explicit
enablement and is disabled automatically after lifecycle termination. Power and
preference access is read-only. The existing `playdate.Launcher` remains the
only exit-to-launcher API.

The `examples/systemstatus` consumer displays motion, battery, power, volume,
reduce-flashing, timezone, and clock-format values and exits through the
Launcher when B is pressed. On 2026-08-02 it passed deterministic tests, SDK
3.1.1 Simulator build and launch, and the conservative device gate at 283,884
bytes of static RAM and a 55,193-byte PDX, including USB deployment and physical
Playdate execution. Hardware observation confirmed accelerometer, battery,
volume, timezone, clock format, reduce-flashing, and `NONE`/`USB` power states.
The first device run exposed direct float-return ABI corruption for battery and
volume; passing their IEEE-754 bits across the TinyGo/C boundary corrected it.
`CHARGE` and `SCREWS` power states, extended conservative-GC soak, memory-growth
measurement, and post-run device-log inspection remain unverified.

## Optional online and debug facilities

`playdate.Scoreboards` provides bounded asynchronous board discovery, score
submission, and personal-best retrieval without implying general networking or
multiplayer support. `playdate.DebugMessages` separately provides a bounded
FIFO for Simulator `!msg` and device serial `msg` diagnostics. Both callbacks
copy SDK-owned data and suppress delivery after lifecycle termination.

The focused consumers pass deterministic adapter tests, Simulator builds, and
conservative device packaging. Serial `msg` delivery was confirmed on physical
hardware; configured-board and live online-scoreboard behavior remain
unverified.

## Integrated persistence acceptance

The external Crank Caverns consumer now uses only public API to combine the released gameplay and persistence capabilities.
Its versioned checkpoint stores progress, score, best time, sound, and
difficulty; migrations cover four schema generations. The title and pause menus
exercise explicit new/save/load/continue flows, System Menu settings use
localized lookup with game-owned fallback, the HUD reads battery status, and
Exit uses `playdate.Launcher`.

Deterministic tests cover round trips, migration, corrupt payload rejection,
failed-write retry, reload, and new-run reset. Windows Simulator interaction,
the conservative-GC device gate, USB installation, and the device launch
command passed on 2026-08-02. The device artifact used 277,524 bytes of static
RAM and produced a 948,032-byte PDX. Physical multi-session restart/update,
injected power loss, corrupt-save recovery, soak, memory-growth measurement,
and post-run device-log inspection remain unverified.

## Advanced sample playback

Games can capability-assert `playdate.SamplePlayers` from their context and
load an explicitly owned `SamplePlayer`. It preserves the base sound-effect
controls and adds bounded repeats, forward or reverse rates, sample duration,
and playback-position control. Streaming `FilePlayer` values optionally expose
`VariableRatePlayer` for positive pitch/speed changes. Negative file-player
rates return `ErrAudioReverseUnsupported`, matching the official streaming API.
Sample reverse is available for PCM assets, but not ADPCM. `Close` releases
native ownership.

`examples/audio` maps A to three sample repeats, Left/Right to forward and
reverse sample rates, B to streaming-music start/stop, and Up/Down to music
rates from 0.25x through 2x. It redraws only on input or playback-state changes;
reverse sample playback seeks to the sample end before starting.

The runtime and generated Simulator/device ABI paths are regression-tested. On
2026-08-02 the example passed audible interaction in Windows Simulator and on a
physical Playdate after conservative hard-float build, USB installation on
COM3, and launch. The device artifact used 268,940 bytes of static RAM and
produced a 125,340-byte PDX. Extended soak, memory-growth measurement, lifecycle
stress, and post-run device-log inspection remain unverified.

## Timed fades and completion

Owned sample and file players optionally expose `CompletionPlayer`; replacing
or clearing its callback releases the previous registration, and `Close`
detaches it before freeing native ownership. Streaming players additionally
expose `FadingPlayer`, whose duration is measured in 44.1 kHz audio frames.
`AudioClock.CurrentAudioTime` returns the same wrapping frame clock through an
optional context capability, allowing fades and game scheduling to share an
exact timebase.

`examples/audio` keeps callbacks bounded to counters and redraw flags. A plays
the repeated sample, while B cycles streaming music through play, a half-second
fade, and stop; the scene displays completion counters and the clock sampled on
input. Unit tests cover callback replacement/lifetime, one-shot fade callbacks,
validation, clock forwarding, and generated Simulator/device bridge symbols.
Official Windows Simulator and hard-float device builds pass; the device
artifact uses 269,876 bytes of static RAM and produces a 130,127-byte PDX.
On 2026-08-02 the scene passed audible sample completion, the half-second music
fade, `Done S/F` callback counters, and an advancing audio-clock display in
Windows Simulator and on a physical Playdate. Installation through COM3 and
device launch succeeded. Extended soak, memory-growth measurement, lifecycle
stress, and post-run device-log inspection remain unverified.

## Routing, synthesizers, and signals

Games can capability-assert `AudioChannels`, `AudioOutputs`, and `Synthesizers`
without widening the base `Context`. `AudioOutputs` exposes the borrowed default
channel, current headphone/headset-microphone state, and headphone/speaker
activation. Explicitly owned channels route sample, file, and synth sources and
expose channel volume and pan. Synths support native waveforms,
ADSR parameters, transpose, audio-clock note scheduling, and frequency or
amplitude modulation by owned LFOs, envelopes, and control-signal timelines.

Routing and modulation edges do not transfer endpoint ownership. Closing a
source or signal detaches every retained edge before freeing it; closing a
channel detaches its sources without closing them. Unit tests cover duplicate
attachments, close ordering, invalid parameters, graph forwarding through
`NewApplication`, and generated Simulator/device bridge symbols.

The audio acceptance scene polls `AudioOutputState` while running and displays
`Output H/M` as connection state. This makes headphone insertion/removal visible
without requiring an audio-thread callback.

`examples/audio` exercises the complete audio graph. A starts an indefinite
scheduled synth note and release schedules `NoteOff`; B plays the routed sample
and A+B controls routed file music. Left/Right cycles all eight synth waveforms,
Up/Down cycles no modulation, amplitude/frequency LFO, envelope, and control
signal. A+Left/Right cycles all seven LFO shapes, B+Left/Right cycles transpose,
B+Up/Down changes channel volume, and the crank controls shared channel pan.

On 2026-08-02 this matrix passed audible interaction in the official Windows
Simulator and on a physical Playdate. The conservative hard-float device
artifact uses 273,456 bytes of static RAM and produces a 146,771-byte PDX; USB
installation through COM3 and launch succeeded. macOS/Linux SDK integration,
extended soak, memory-growth measurement, lifecycle stress, and post-run
device-log inspection remain unverified.

## Instruments, sequences, and effects

Games can capability-assert `Sequencers` and `AudioEffects`. Instruments retain
voice-range attachments without taking ownership of synths; tracks likewise do
not own instruments, and sequence slots do not own tracks. Tracks support note
and MIDI-controller events, while sequences support MIDI loading, tempo, loops,
time, dynamic track construction, and bounded one-shot completion callbacks.

Channels accept owned two-pole filters, bit crushers, ring modulators, delay
lines, and overdrive processors. Effect mix and parameter modulators are
explicit. Closing either endpoint detaches its graph edges before releasing the
native handle.

`examples/audio` creates a dynamic four-note sequence and routes its instrument,
sample player, streaming file player, and synth through one channel containing
all five effect types. B starts or fades music, B+Up plays the completion sample,
B+Left/Right selects the active effect, A+Up/Down starts or stops the sequence,
Up/Down selects modulation, A+Left/Right selects the LFO, and A+B+Left/Right
selects the synth waveform. Arpeggiator LFOs use the explicit `SetArpeggiation`
steps `0, 4, 7, 12` and select frequency modulation.

Unit tests and the official Windows Simulator and conservative hard-float device
builds pass. On 2026-08-03 the full sequence, completion-counter, routing,
effect, synth-waveform, and modulation matrix passed audible interaction in the
Windows Simulator and on a physical Playdate after USB installation through
COM3. The accepted device artifact uses 278,900 bytes of static RAM and produces
a 168,558-byte PDX. Device audio-thread completions enter a bounded native FIFO
and are delivered to Go on the next update frame. Extended soak, memory-growth
measurement, lifecycle stress, post-run device-log inspection, and macOS/Linux SDK
integration remain unverified.

## Microphone input

Games capability-assert `Microphones` without widening `Context`. Permission is
explicitly pending, denied, or granted; recording selects automatic, internal,
or headset input and returns an owned recorder. Native sample views expire when
their callback returns. `MicrophoneSamples.CopyTo` copies only into the game's
bounded destination and never retains that buffer.

`examples/microphone` requests permission, starts recording, and displays the
live peak and delivered-block count. A stops or restarts the recorder. B saves
up to one second as `microphone.wav` in Data and audibly plays the same capture
through native-owned copied PCM:

```sh
go run ./cmd/gopdsdk run --sdk /path/to/PlaydateSDK ./examples/microphone
go run ./cmd/gopdsdk run device --memory conservative --sdk /path/to/PlaydateSDK ./examples/microphone
```

On 2026-08-03 permission, changing peak/block counters, stop/start, WAV save,
and audible playback passed in the official Windows Simulator and on a physical
Playdate installed through COM3. Device input reaches Go through a bounded
native FIFO on update frames. The accepted hard-float artifact uses 282,824
bytes of static RAM and produces a 50,601-byte PDX. Denial/revocation,
long-run overflow/memory measurement, lifecycle stress, post-run device-log
inspection, and macOS/Linux SDK integration remain unverified.

## Bitmap composition

Games capability-assert `BitmapCompositor` for rotated/scaled bitmap drawing
and callback-scoped stencils. Transforms reject non-finite values and
non-positive scales. `WithStencil` borrows a live bitmap only until its callback
returns, clears the native stencil on callback errors, rejects nesting, and
requires tiled stencil widths to be multiples of 32 pixels.

`examples/composition` builds an owned 64×64 source and a screen-aligned 400×240
stencil bitmap through `OffscreenGraphics`. Stencils use framebuffer
coordinates, so the mask is positioned around the right-hand draw target. SDK
3.1.1 rendered a direct stencil-plus-rotation path only at cardinal angles in
both the Go scene and an equivalent official Lua diagnostic. The accepted
portable path rotates into a transparent 400×240 offscreen canvas, then draws
that canvas through the screen stencil.

```sh
go run ./cmd/gopdsdk run --sdk /path/to/PlaydateSDK ./examples/composition
go run ./cmd/gopdsdk run device --memory conservative --sdk /path/to/PlaydateSDK ./examples/composition
```

The portable contract and generated Simulator/device bridge paths are
unit-tested. Official Windows SDK 3.1.1 Simulator compilation and conservative
hard-float device build pass. Visual interaction passed in the Simulator and on
a physical Playdate after USB deployment through COM3 on 2026-08-08. The
accepted device build used 275,312 bytes of static RAM and produced a
36,970-byte PDX. Performance, bounded-memory, soak, and post-run device-log
regression checks pass on the verified Windows profile.

## Video

`examples/video` owns a generated four-second, 48-frame PDV fixture and matching
audio track. It exercises metadata, validated frame playback, explicit video
and audio cleanup, and screen/offscreen render-target switching. A pauses, B
changes the target, and Left/Right step through frames.

On 2026-08-08 the complete interaction passed visual and audible acceptance in
the official Windows Simulator and on a physical Playdate. The accepted
hard-float device build used 279,880 bytes of static RAM and produced a
976,532-byte ELF and a 227,458-byte PDX. Performance, bounded-memory, soak, and
post-run device-log regression checks pass on the verified Windows profile.

```sh
go run ./cmd/gopdsdk run --sdk /path/to/PlaydateSDK ./examples/video
go run ./cmd/gopdsdk run device --memory conservative --sdk /path/to/PlaydateSDK ./examples/video
```

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

Docker is not part of the supported-host matrix. It would provide another Linux environment,
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
- Graphics cover clear/text, bitmaps, sprites, animation, custom fonts,
  callback-scoped framebuffer access, and drawing into owned bitmaps. Audio
  covers sound effects/file players, advanced sample controls,
  timed fades/completion callbacks, owned routing, waveform synths and
  modulation signals, instruments/sequencing/effects, and bounded
  microphone input with Simulator and physical-device acceptance.
