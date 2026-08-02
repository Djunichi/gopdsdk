# Migrating to v0.4.0

`v0.4.0` adds optional persistence and Playdate system-integration capabilities
without making them mandatory on `playdate.Context`. Existing v0.3 games can
upgrade their module requirement and continue unchanged.

## Persistence

Assert `playdate.FileSystem` only where persistence is required. Close every
opened file and use `errors.Is(err, playdate.ErrFileIO)` for classification;
human-readable native diagnostics are not a machine contract.

Prefer `playdate/store` for small saves and settings that need bounded size,
checksum validation, atomic replacement, backup recovery, and versioned
migration. Each `VersionMigration` upgrades exactly one version. Applications
still own payload encoding, defaults, future-version policy, and the decision
to recover or preserve corrupt data.

Migration failures continue to satisfy `errors.Is(err, store.ErrMigration)`.
The direct migration cause is also classifiable with `errors.Is`, but v0.4.0 no
longer appends an arbitrary cause string to the public error text. Do not parse
error messages.

## System capabilities

System Menu, localization, motion, power, preferences, scoreboards, debug
messages, and Launcher exit are narrow optional interfaces. Assert only the
capabilities a game needs and keep game-owned fallback behavior when one is
unavailable. Owned menu items and enabled accelerometer state are cleaned up at
termination; games remain responsible for their other owned resources.

## Tile-map draw errors

Tile-map drawing still satisfies both `errors.Is(err, playdate.ErrTileMapDraw)`
and `errors.Is(err, cause)`. Its text is now stable rather than concatenating
the underlying error text. Code that parsed the old string must switch to
`errors.Is` or a concrete cause assertion.

## Release-candidate boundary

Before the `v0.4.0` tag exists, use a local `replace` for candidate testing.
After tagging, remove the replacement and repeat the clean module-proxy check
from [RELEASING.md](RELEASING.md).
