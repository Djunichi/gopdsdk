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

## Unreleased: video streams

The experimental `VideoStream`, `Videos.NewVideoStream`, and stream-specific
errors were removed after SDK 3.1.1 acceptance showed that ordinary PDV files
produce no frames and the first stream interface method dispatch crashed under
TinyGo 0.41.1 on physical hardware. Use `Videos.LoadVideo` and an explicitly
owned `VideoPlayer` for packaged PDV playback. There is no offline replacement
for the undocumented stream container; the capability is deferred to
post-v1.0 `v1.1 networking` research.

## Published-module verification

After publication, remove any local `replace`, require
`github.com/Djunichi/gopdsdk v0.10.0`, and
repeat the clean module-proxy check from [RELEASING.md](RELEASING.md).
