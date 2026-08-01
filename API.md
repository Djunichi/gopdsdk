# Public API

This document describes the public contract proposed for gopdsdk `v0.1.0`.
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

`playdate.Context` composes three smaller interfaces:

- `System` provides the wrapping monotonic millisecond clock.
- `InputReader` provides one immutable frame snapshot.
- `Graphics` provides clear/text and the accepted bitmap operations.
- `Sprites` creates explicitly owned sprites and runs the shared display-list
  update/draw pass.

Helpers should accept the narrowest interface they need. This keeps pure game
logic independent from runtime callbacks and makes deterministic testing
straightforward.

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

## Device Go subset

The accepted device profile is sequential Go with TinyGo conservative GC,
`scheduler none`, and fail-stop panic handling. Goroutines, channels, `select`,
`defer`/`recover`, finalizers, public reflection, cgo, and application runtime
hooks are unsupported. Applications must keep their live object graph bounded;
allocation inside `Update` is allowed but remains workload- and hardware-tested.

Simulator compilation alone does not prove device compatibility for a standard
library package. See [COMPATIBILITY.md](COMPATIBILITY.md) for the evidence
matrix and exact verified toolchain.
