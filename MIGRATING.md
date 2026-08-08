# Migrating to v0.6.0

`v0.6.0` adds optional bitmap composition, display and sprite-redraw controls,
owned PDV video, and bounded diagnostics without widening `playdate.Context`.
Existing v0.5 games can update their module requirement and continue unchanged.

## Graphics and redraw

Capability-assert `BitmapCompositor` for rotated/scaled bitmap drawing and
callback-scoped stencils. Pre-render non-cardinal rotation into a transparent
offscreen bitmap before drawing through a screen stencil. Capability-assert
`Display` and `SpriteRedraw` only when using presentation modes or global dirty
regions; use `Sprite.MarkDirty` or `Sprite.MarkDirtyRect` for per-sprite
invalidation.

## Video and diagnostics

Capability-assert `Videos` and close every owned `VideoPlayer`. A player target
borrows an owned bitmap; keep that bitmap open until changing the target or
closing the player. External games may use `playdate/diagnostics.Collector` for
bounded frame-time, heap, and owned-resource samples without frame-loop I/O.

## Published-module verification

Remove any local `replace`, require `github.com/Djunichi/gopdsdk v0.6.0`, and
repeat the clean module-proxy check from [RELEASING.md](RELEASING.md).
