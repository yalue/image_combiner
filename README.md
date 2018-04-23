Image Combiner
==============

This program is intended to merge several input images, with each input image
corresponding to one color channel in an output image. It supports reading
gif, jpeg, png, and netpbm input images, but always creates outputs in the .jpg
format.

Usage
-----

The program can be either compiled from source code (see the Compilation
section below), or downloaded from [the releases page](https://github.com/yalue/image_combiner/releases).

The program takes pairs of command-line arguments representing input images and
colors to assign in the output image. The final argument is the name of the
output image, which should be a .jpg value.  Colors may either be specified
using an SVG color name, 6 hex digits (for 24-bit color), or 12 hex digits
(for 48-bit color).

```bash
# The output image, image_abc.jpg, will contain a red 'A', a green 'B', and a
# purple 'C'.
./image_combiner \
    examples/image_a.png red \
    examples/image_b.gif green \
    examples/image_c.pgm blueviolet \
    image_abc.jpg

# Alternatively, using hex digits instead of some color names, including 48-bit
# purple (rrrrggggbbbb) for image_c.pgm.
./image_combiner \
    examples/image_a.png red \
    examples/image_b.gif 008000 \
    examples/image_c.pgm 77770999cccc \
    image_abc.jpg
```

Colors can be specified as names or hex values with either 6 or 12 digits for
24-bit or 48-bit RGB (respectively).

If the resolutions of the input images don't match, then the output image will
have the resolution of the largest input, with all smaller images placed at the
top left.

Compilation
-----------

This program was created using `go`:

```bash
go get -u github.com/yalue/image_combiner
go install github.com/yalue/image_combiner
```
