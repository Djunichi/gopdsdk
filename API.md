# Public API

This document describes the published `v0.6.0` public contract, including
bitmap composition, display and sprite-redraw controls, optional video, and the
bounded diagnostics package. The module is still pre-v1: minor releases may
make intentional breaking changes, which must be called out in release notes.
Patch releases preserve the documented API and behavior.

Applications import the native contract from
`github.com/Djunichi/gopdsdk/playdate`. Applications that need the optional
bounded persistence layer additionally import
`github.com/Djunichi/gopdsdk/playdate/store`. External games collecting bounded
performance evidence import
`github.com/Djunichi/gopdsdk/playdate/diagnostics`. Packages below `internal/`,
generated runtime bridges, CLI build plans, and example internals are not
public API.

## Bounded diagnostics

`playdate/diagnostics.Collector` records a fixed number of samples without
allocating in the frame loop. A game supplies elapsed frame milliseconds, live
heap bytes, and its count of currently owned native resources. The report
contains mean, p50, p95, p99, and maximum frame time; heap start, end, maximum,
and signed growth; and native-resource start, end, minimum, and maximum. Frame
times of 255 ms or more share a bounded tail bucket; the exact observed maximum
is retained. The 36,000-frame ceiling is twenty minutes at 30 FPS.

The collector deliberately does not discover runtime memory or resource
ownership and does not perform frame-loop I/O. Consumers obtain heap data from
their target runtime, count the resources they explicitly own, and write or
render the final report outside the measured interval. Device artifact static
RAM, ELF size, and packaged PDX size continue to come from `gopdsdk build
device`; those build metrics are not evidence of Simulator or hardware timing.

## Application contract

An importable application package provides:

```go
func New() playdate.Game
```

`Game.Init` runs before updates. `Game.Update` receives the common `Context` and
returns whether the display should refresh. An error from a game callback
permanently stops that application instance.

Implement `playdate.LifecycleGame` when the game needs pause, resume, lock,
unlock, terminate, or low-power events. Events retain platform order. Release
owned native resources on `LifecycleTerminate`; initialization must roll back
resources already acquired when a later acquisition fails.

## Context capabilities

`playdate.Context` composes five smaller interfaces:

- `System` provides the wrapping monotonic millisecond clock.
- `InputReader` provides one immutable frame snapshot.
- `Graphics` provides clear/text and the accepted bitmap operations.
- `Sprites` creates explicitly owned sprites and runs the shared display-list
  update/draw pass.
- `Audio` loads explicitly owned short sound effects and one-file streaming
  players.

Helpers should accept the narrowest interface they need. This keeps pure game
logic independent from runtime callbacks and makes deterministic testing
straightforward.

Games that offer an explicit quit action assert the optional `Launcher`
capability and call `ExitToLauncher`. The official runtime sends
`LifecycleTerminate` before starting the Launcher, so cleanup remains in the
normal lifecycle path. An in-game title or pause menu is application state:
switching from gameplay back to that menu does not terminate or restart the
process.

Games that sample motion assert `Accelerometer`, explicitly enable it, and read
`AccelerometerXYZ`; reads before enablement return zero and the runtime disables
the peripheral after `LifecycleTerminate`. `PowerMonitor` exposes the power
flags, battery percentage, and voltage. `SystemPreferences` exposes only the
read-only system volume, reduce-flashing preference, timezone offset in seconds,
and 12/24-hour preference. These capabilities do not add setters for global
user preferences.

Games that draw immediate-mode geometry assert `PrimitiveGraphics`; games that
change clipping, draw offset, or bitmap compositing assert `GraphicsState`.
These optional capabilities keep existing graphics fakes small. Primitive
dimensions and stroke widths must be positive, and ellipse angles must be
finite. Invalid values return the corresponding graphics sentinel before a
native call. The runtime's per-frame context forwards both optional slices to
the platform adapter; an adapter without the requested slice returns
`ErrGraphicsUnavailable`.

`PrimitiveGraphics` includes filled polygons expressed as `[]GraphicsPoint`
with non-zero or even-odd fill rules, plus stroked and filled rounded
rectangles. A polygon requires at least three vertices and rounded-rectangle
radii cannot be negative. `GraphicsState` also controls butt, square, or round
line caps, the solid display background color, and clipping in absolute screen
coordinates. `SetClipRect` remains drawing-coordinate-relative;
`SetScreenClipRect` is unaffected by the current drawing offset.

Games that need direct 1-bit pixels assert `FramebufferGraphics` and use
`WithFramebuffer`. The callback receives a zero-copy view of the 400×240
working framebuffer with a 52-byte row stride. `Pixel` and `SetPixel` use
MSB-first bits; a set bit is white. `SetPixel` marks its row dirty, while code
that mutates `Bytes` must call `MarkDirtyRows`. Playdate receives the combined
inclusive dirty range when the callback returns, including when it returns an
error. The view and its byte slice must not be retained; checked operations on
the view return `ErrFramebufferExpired` after the callback.

Games that render through normal graphics operations into an owned bitmap
assert `OffscreenGraphics` and use `DrawInto`. The previous native drawing
context is restored before `DrawInto` returns, including callback errors.
Borrowed table frames are rejected with `ErrBitmapBorrowed`; callbacks and
bitmap cleanup remain explicitly owned by the game.

Games that rotate or stencil bitmaps assert `BitmapCompositor`.
`DrawRotatedBitmap` accepts a finite angle and center proportions plus positive,
finite scales. `WithStencil` borrows a live bitmap only for its synchronous
callback and clears the native stencil before returning, including callback
errors. Stencil callbacks cannot be nested. A tiled stencil must have a width
that is a multiple of 32 pixels, matching the official SDK constraint.
Playdate SDK 3.1.1 does not reliably combine an active screen stencil with a
direct rotated-bitmap draw at non-cardinal angles. Render the transform into a
transparent offscreen bitmap first, then draw that bitmap through the stencil.

`SolidPaint`, `XORPaint`, and `PatternPaint` create primitive paint values.
Patterns contain eight image rows followed by eight mask rows and are copied by
value; neither runtime adapter retains a pointer after the drawing call.
`DrawMode` mirrors the official copy, transparent, fill, XOR/NXOR, and inverted
bitmap compositing modes. Restore offsets, clipping, and draw mode after a
localized drawing pass because these values are native graphics state.

Games that need PDV playback assert the optional `Videos` capability.
`LoadVideo` returns an explicitly owned `VideoPlayer`; close it exactly once.
`Info` reports dimensions, rate, frame count, and decoder position.
`RenderFrame` validates the frame index and reports native decoder errors as
`VideoOperationError`. `SetContext` borrows a live owned bitmap without taking
ownership; keep it open until selecting the screen or another bitmap.
The four-second `examples/video` consumer passed visual and audible acceptance
in the official Windows Simulator and on a physical Playdate on 2026-08-08.
This evidence covers synchronized companion audio, pause/resume, looping,
stepping, screen/offscreen targets, and the later performance,
bounded-memory, soak, and post-run device-log regression checks on the verified
Windows profile.

## Input

`Input.Buttons` is the current button set. `Pressed` and `Released` are edges
for the current frame; `Held` contains buttons that remained down across the
frame boundary. `Buttons.Has` tests whether every requested bit is set.

`CrankAngle`, `CrankDelta`, dock state/transitions, and `DeltaSeconds` use the
same model in Simulator and device builds. Game state should advance from the
snapshot rather than branching on the build target.

## Bitmap ownership and errors

`LoadBitmap` and `NewBitmap` return owned bitmap handles. Close each successful
handle exactly once. A failed partial initialization must close earlier handles.
Do not rely on finalizers.

After a successful close, bitmap operations return `ErrBitmapClosed`.
Invalid sizes, colors, and scales use the exported sentinel errors. A failed
resource load returns `BitmapLoadError`, which retains the Playdate diagnostic.
Callers should use `errors.Is` for sentinels and `errors.As` for the typed load
error rather than matching error text.

Games that need bitmap storage and mask operations assert
`BitmapDataGraphics`. `WithBitmapData` accepts only an owned bitmap and exposes
its MSB-first image and optional mask bytes for one synchronous callback;
checked access after return fails with `ErrBitmapDataExpired`. `SetPixel`
accepts black or white, updates the native image bytes immediately, and marks
the view dirty. Call `MarkDirty` after direct slice writes; `Dirty` lets an
editor propagate the change to sprites through their redraw API.

`CopyBitmap`, `RotatedBitmap`, and `CopyDisplayBuffer` return new owned handles.
The integer returned with a rotated bitmap is the native allocation size for
budget accounting. `LoadIntoBitmap` and `LoadIntoBitmapTable` require owned
destinations and preserve the handle identity. A bitmap table created with
`NewBitmapTable` owns its frames; returned frames remain borrowed from it.

`SetBitmapMask` requires equal bitmap dimensions and keeps both Go handles live
without transferring ownership. `BitmapMask` returns an owned native view tied
to the masked bitmap's storage: close the view before its parent. A mask cannot
be closed while another bitmap retains it. `ClearBitmapMask` removes the
association without closing either bitmap. Mask collision accepts only the four
`BitmapFlip` values and an integer test rectangle.

## Sprites and display list

Games that need physical-display presentation controls assert the optional
`Display` capability. Refresh rates must be finite and between 0 and 50 FPS;
scale is limited to 1, 2, 4, or 8; mosaic components are limited to 0 through
3. Inversion, flipping, and display offset remain explicit presentation state.

Games that control global redraw policy assert `SpriteRedraw` and use
`SetAlwaysRedraw` or `AddDirtyRect`. Individual sprites use `MarkDirty` or
`MarkDirtyRect`; invalid rectangles are rejected before native calls. The
display and redraw acceptance scene passed dirty/full redraw switching,
partial invalidation, display effects, comparative measurements, and reset
behavior in the official
Windows Simulator and on physical Playdate hardware.

`NewSprite` returns an owned sprite. Configure its bitmap, position, visibility,
and z-index, then call `Add`. `Add` and `Remove` are idempotent. Each frame,
move game objects and call `UpdateAndDrawSprites` once to update and render the
global Playdate display list.

`Close` removes an added sprite before freeing it and is always explicit; close
sprites before closing bitmaps referenced by them. If initialization fails,
close every sprite already created, followed by its bitmap. Sprite movement and
crank input use the same float32 contract in Simulator and device adapters.

Source resources live below `resources/`; that directory becomes the PDX root.
For example, load `resources/images/player.png` as `images/player`.

## Collision queries

Sprite collision rectangles opt an owned sprite into collision queries.
`MoveWithCollisions` resolves a goal position and returns the actual position
plus ordered contacts using slide, freeze, overlap, or bounce responses.
Point, rectangle, and overlap queries expose only sprites owned by the current
runtime; a foreign or closed sprite passed to an operation returns the
corresponding sprite error.

## Bitmap tables and animation

`LoadBitmapTable` returns an owned table whose `Frame` method returns borrowed
bitmap handles. Close the table only after its animations and sprites no longer
use those frames; borrowed frames cannot be closed independently.

`NewAnimation` creates a looping, allocation-free frame selector over a bounded
table range. Advance it with `Input.DeltaSeconds`, or select a fixed frame.
Pause and resume retain partial-frame time. Invalid construction returns
`ErrAnimationConfig`; an invalid table frame returns `ErrBitmapFrameRange`.

## Audio ownership and lifecycle

`LoadSoundEffect` loads a short, memory-backed sample/player pair;
`LoadFilePlayer` loads one streaming file player. Both expose play, stop,
stereo volume, stopped/playing/paused state, pause/resume, and explicit close.
Repeated `SoundEffect.Play` restarts the accepted effect without allocating a
new native player.

Pause players on pause/lock lifecycle events, resume them on resume/unlock,
and close file players before sound effects on termination. `Close` stops
playback before releasing native handles. A failed sound-effect construction
frees its partially acquired sample/player, and game initialization must close
earlier successful audio loads if a later load or configuration step fails.
After close, operations return `ErrAudioClosed`; invalid volume returns
`ErrAudioVolume`, playback rejection returns `ErrAudioPlay`, and load failures
return `AudioLoadError`.

Advanced games may separately capability-assert `AudioChannels`,
`Synthesizers`, `Sequencers`, and `AudioEffects`; these optional slices do not
widen the base `Context`.

Microphone games capability-assert `Microphones`. Request permission before
recording, handle pending/denied/granted explicitly, and close the owned
`MicrophoneRecorder`. A `MicrophoneSamples` value is callback-scoped; copy only
the needed samples into a bounded caller-owned `[]int16`. The view expires when
the callback returns and the runtime never retains the destination. Microphone
failures use the separate `ErrMicrophone*` domain rather than audio errors.
On device, microphone input first enters a bounded native FIFO and is delivered
to Go during update frames rather than re-entering Go from the audio thread.
`PCMPlayers.NewPCMPlayer` synchronously copies mono signed 16-bit PCM into
native-owned storage and never retains the caller slice.

`LFO.SetArpeggiation` requires at least one finite half-step offset and configures
the native arpeggiator sequence; an empty or non-finite step list returns
`ErrAudioParameter`. Audio completion callbacks are retained by ID and delivered
on an update frame. On device, native audio callbacks first enter a bounded FIFO
instead of re-entering Go from the audio thread.

## Fonts and deterministic UI

Native contexts also implement the narrow `FontGraphics` capability. Assert it
only in games that need custom fonts, load packaged `resources/fonts/name.fnt`
as `fonts/name`, and close every successful `Font` exactly once. `TextWidth`
and `Height` use the same native font metrics as drawing; closed fonts return
`ErrFontClosed`, and foreign handles return `ErrFontInvalid`.

Keep game UI as state plus a deterministic draw plan: measure strings, derive
coordinates, then execute `DrawTextFont` commands. HUD, pause, and game-over
screens do not require a generic widget tree, focus model, event router, or
layout engine. Add shared UI abstractions only after a second real game exposes
the same repeated contract.

## Filesystem and versioned persistence

Games assert the optional `FileSystem` capability for owned files, metadata,
non-recursive listing, and directory mutations. Paths are game-relative;
read-source flags distinguish packaged resources from the Data directory.
Every successful `OpenFile` result must be closed. Native failures are copied
into `FileOperationError` and classify as `ErrFileIO`.

The separate `playdate/store` package provides bounded checksummed envelopes,
atomic temporary-file replacement, backup recovery, and explicit one-version
migrations. Serialization and corrupt-save policy remain application-owned.
Migration failures classify as `store.ErrMigration` and preserve their direct
cause for `errors.Is`; callers must not parse error strings.

Native capability sentinels remain members of the common `playdate` package and
are grouped by API area in the source tree. `playdate/store` owns its separate
error domain because it is a higher-level package rather than a native context
capability. Source-file placement does not change error identity or import
paths.

## System integration

`SystemMenu` owns at most three action, checkmark, or options items. Items retain
their callback until removal, may be removed idempotently, and are cleaned up
after termination. `Localization` exposes system language and copied `.strings`
lookups; missing-key fallback remains game policy.

`Accelerometer` requires explicit enablement and is disabled at termination.
`PowerMonitor` and `SystemPreferences` expose read-only device/global state.
`Launcher` remains the single exit-to-launcher capability.

`Scoreboards` is an optional bounded asynchronous service, not general
networking. `DebugMessages` is a separate bounded diagnostic FIFO for Simulator
and serial input. Neither capability may re-enter game code from a native
callback, and callbacks after termination are suppressed.

## Device Go subset

The accepted device profile is sequential Go with TinyGo conservative GC,
`scheduler none`, and fail-stop panic handling. Goroutines, channels, `select`,
`defer`/`recover`, finalizers, public reflection, cgo, and application runtime
hooks are unsupported. Applications must keep their live object graph bounded;
allocation inside `Update` is allowed but remains workload- and hardware-tested.

Simulator compilation alone does not prove device compatibility for a standard
library package. See [COMPATIBILITY.md](COMPATIBILITY.md) for the evidence
matrix and exact verified toolchain.
