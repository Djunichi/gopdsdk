# P6.3 video acceptance

This example owns a four-second deterministic PDV fixture and matching WAV
generated from original frames and tones. PDV has no embedded audio, so the
game starts and loops a separate owned file player alongside it. Regenerate
both assets from this directory with:

```powershell
go run ./internal/generate
```

The generator uses only the Go standard library. Its minimal all-I-frame
container follows the public PDV format notes and is accepted against the
official SDK decoder; the official Playdate API remains the normative runtime
reference.

Controls: A pauses playback, B switches between the owned offscreen bitmap and
the direct screen context, and Left/Right step through frames.
