# Migrating to v0.7.0

`v0.7.0` adds advanced drawing state, owned bitmap data and masks, text layout
and font metrics, and display introspection. Existing v0.6 games can update
their module requirement and continue unchanged; all new capabilities are
additive.

## Drawing and bitmap data

Use the expanded `Graphics` surface for filled polygons, rounded rectangles,
line caps, background color, and screen-coordinate clipping. Owned bitmap data
is exposed only inside its callback lifetime; mark modified data dirty before
the callback returns. Keep borrowed mask views and glyph bitmaps within the
lifetime of their owning bitmap or font.

## Text, fonts, and display

Use bounded text drawing and wrapping-height measurement for rectangle layout.
Glyph metrics and kerning support custom renderers while packaged `.fnt` files
remain the portable custom-font path. The expanded `Display` capability reports
logical dimensions, nominal refresh rate, and measured FPS.

## Published-module verification

Remove any local `replace`, require `github.com/Djunichi/gopdsdk v0.7.0`, and
repeat the clean module-proxy check from [RELEASING.md](RELEASING.md).
