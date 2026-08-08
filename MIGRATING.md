# Migrating to v0.8.0

`v0.8.0` completes the offline sprite and collision surface with geometry and
presentation controls, native sprite tilemaps, detailed queries and
display-list operations, and bounded per-sprite callbacks. Existing v0.7 games
can update their module requirement and continue unchanged; all new
capabilities are additive.

## Sprite presentation and tilemaps

Owned sprites now expose center, bounds, position getters, image flip, draw
mode, opacity, stencil, clip rectangle, draw-offset policy, and deterministic
update/collision enable state. Games needing official native sprite tilemaps
can assert `playdate.SpriteTileMaps`; retain attached bitmap tables until the
tilemap and its sprites no longer use them.

## Queries, display-list control, and callbacks

Assert `playdate.SpriteQueries` for line and detailed-hit queries and
`playdate.SpriteDisplayList` for count, bulk membership, remove-all, and
collision-world reset. Callback registration is bounded to 64 slots per
context. Draw and update callbacks run synchronously in display-list order;
collision callbacks select the response for the ordered sprite pair. Clear
callbacks or close their sprites to release registrations deterministically.

## Published-module verification

Remove any local `replace`, require `github.com/Djunichi/gopdsdk v0.8.0`, and
repeat the clean module-proxy check from [RELEASING.md](RELEASING.md).
