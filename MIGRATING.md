# Migrating to v0.9.0

`v0.9.0` completes the offline sound surface with owned samples, streaming
control, output routing, synthesis, modulation, sequencing, effects, and
bounded custom audio callbacks. Existing v0.8 games can update their module
requirement and continue unchanged; all new capabilities are additive.

## Samples, routing, and output

Capability-assert `playdate.AudioSamples`, `playdate.AudioChannels`, or
`playdate.AudioOutputs` only where needed. Samples own their native storage;
borrowed sample-data views expire with the sample, while players borrow an
attached sample and must release it before that sample can close. Channels do
not own their sources, signals, or effects.

## Custom audio and synthesis

`playdate.CallbackAudio` and `playdate.GeneratorSynthesizers` invoke Go render
callbacks during frame updates. Native audio callbacks consume fixed
4,096-frame rings and emit silence on underrun; they never re-enter Go. Custom
generator parameters use the official 1-based indices 1 through 8. Close
sources, synths, signals, effects, tracks, and sequences deterministically.

## Published-module verification

Remove any local `replace`, require `github.com/Djunichi/gopdsdk v0.9.0`, and
repeat the clean module-proxy check from [RELEASING.md](RELEASING.md).
