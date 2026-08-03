# Migrating to v0.5.0

`v0.5.0` adds optional advanced-audio and microphone capabilities without
widening `playdate.Context`. Existing v0.4 games can update their module
requirement and continue unchanged.

## Advanced playback and graphs

Capability-assert only what a game uses: `SamplePlayers`, `AudioClock`,
`AudioChannels`, `Synthesizers`, `Sequencers`, or `AudioEffects`. Players,
channels, synths, signals, instruments, tracks, sequences, effects, delay taps,
and PCM players are explicitly owned. Close children and detach graph edges
before their parents; do not infer ownership transfer from attachment.

Streaming `FilePlayer` rates must remain positive. Reverse playback is limited
to PCM `SamplePlayer` assets; negative streaming rates return
`ErrAudioReverseUnsupported`. Completion callbacks are replaceable and are
released on stop/close according to their interface contract.

## Microphone input

Assert `Microphones`, request permission explicitly, and handle
pending/denied/granted states. `MicrophoneSamples` is valid only during its
callback. Copy into a bounded caller buffer and do not retain the scoped view.
Close the owned recorder on replacement or lifecycle termination.

Microphone failures use the separate `ErrMicrophone*` domain rather than
`ErrAudio*`. `PCMPlayers.NewPCMPlayer` copies mono signed 16-bit PCM into
native-owned storage; its input slice is not retained.

## Release-candidate boundary

Before the `v0.5.0` tag exists, use a local `replace` for candidate testing.
After tagging, remove the replacement and repeat the clean module-proxy check
from [RELEASING.md](RELEASING.md).
