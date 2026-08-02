# Public API

This document describes the public contract proposed for gopdsdk `v0.2.0`.
The module is still pre-v1: minor releases may make intentional breaking
changes, which must be called out in release notes. Patch releases preserve the
documented API and behavior.

Applications import only `github.com/Djunichi/gopdsdk/playdate`. Packages below
`internal/`, generated runtime bridges, CLI build plans, and example internals
are not public API.

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

Games that draw immediate-mode geometry assert `PrimitiveGraphics`; games that
change clipping, draw offset, or bitmap compositing assert `GraphicsState`.
These optional capabilities keep existing graphics fakes small. Primitive
dimensions and stroke widths must be positive, and ellipse angles must be
finite. Invalid values return the corresponding graphics sentinel before a
native call. The runtime's per-frame context forwards both optional slices to
the platform adapter; an adapter without the requested slice returns
`ErrGraphicsUnavailable`.

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

`SolidPaint`, `XORPaint`, and `PatternPaint` create primitive paint values.
Patterns contain eight image rows followed by eight mask rows and are copied by
value; neither runtime adapter retains a pointer after the drawing call.
`DrawMode` mirrors the official copy, transparent, fill, XOR/NXOR, and inverted
bitmap compositing modes. Restore offsets, clipping, and draw mode after a
localized drawing pass because these values are native graphics state.

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

## Sprites and display list

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

This vertical API intentionally excludes synthesis, microphone input, and the
rest of the Playdate sound binding.

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

## Device Go subset

The accepted device profile is sequential Go with TinyGo conservative GC,
`scheduler none`, and fail-stop panic handling. Goroutines, channels, `select`,
`defer`/`recover`, finalizers, public reflection, cgo, and application runtime
hooks are unsupported. Applications must keep their live object graph bounded;
allocation inside `Update` is allowed but remains workload- and hardware-tested.

Simulator compilation alone does not prove device compatibility for a standard
library package. See [COMPATIBILITY.md](COMPATIBILITY.md) for the evidence
matrix and exact verified toolchain.
