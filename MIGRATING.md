# Migrating to v0.10.0

`v0.10.0` adds optional offline system-control and system-environment
capabilities. Existing v0.9 games can update their module requirement and
continue unchanged; the new capabilities and lifecycle events are additive.

## System control

Capability-assert `playdate.SystemControls` only where launch arguments,
restart, menu-image ownership, auto-lock, crank-sound control, or ordered
button callbacks are needed. Applications implementing `LifecycleGame` may now
receive mirror-started and mirror-ended events. Retained menu images and system
settings are cleared or restored before termination cleanup.

## Clock and device information

Capability-assert `playdate.SystemEnvironment` for the offline Playdate epoch,
calendar conversion, the independent elapsed timer, and copied OS, language,
and PDX version information. `Input.DeltaSeconds` now uses the wrapping
monotonic millisecond clock and no longer resets the SDK elapsed timer.

## Published-module verification

After publication, remove any local `replace`, require
`github.com/Djunichi/gopdsdk v0.10.0`, and
repeat the clean module-proxy check from [RELEASING.md](RELEASING.md).
