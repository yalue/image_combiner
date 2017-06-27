Image Combiner
==============

This program is intended to merge several input images, with each input image
corresponding to one color channel in an output image. It supports reading
gif, jpeg, png, and netpbm input images, but always creates outputs in the .jpg
format.

Usage
-----

The `-r`, `-g`, and `-b` flags provide the paths to the images used for the
red, green, and blue color channels (respectively) in the output image.

```
image_combiner -r red_channel.png -g green_channel.pbm -b blue_channel.jpg \
  -output combined_image.jpg
```

Compilation
-----------

This program was created using `go`:

```bash
go install github.com/yalue/image_combiner
```
