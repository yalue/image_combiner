// The image combiner program is a file which takes 3 (presumably grayscale)
// images and combines them into a singe image file, using each input as a
// different color channel.
package main

import (
	"flag"
	"fmt"
	_ "github.com/spakin/netpbm"
	_ "golang.org/x/image/bmp"
	"image"
	"image/color"
	_ "image/gif"
	"image/jpeg"
	_ "image/png"
	"os"
)

// Converts a given arbitrary RGB color to a single channel in grayscale.
func convertToGrayscale(c color.Color) uint32 {
	r, g, b, _ := c.RGBA()
	return (r + g + b) / 3
}

// Takes 3 image filenames and returns the maximum dimensions of all of them.
func getMaxDimensions(imageFiles []string) (int, int, error) {
	var maxW, maxH, w, h int
	var pic image.Image
	var e error
	var f *os.File
	for _, filename := range imageFiles {
		fmt.Printf("Getting dimensions for %s...\n", filename)
		f, e = os.Open(filename)
		if e != nil {
			return 0, 0, fmt.Errorf("Failed opening %s: %s", filename, e)
		}
		pic, _, e = image.Decode(f)
		if e != nil {
			f.Close()
			return 0, 0, fmt.Errorf("Failed decoding %s: %s", filename, e)
		}
		w = pic.Bounds().Dx()
		h = pic.Bounds().Dy()
		pic = nil
		f.Close()
		if w > maxW {
			maxW = w
		}
		if h > maxH {
			maxH = h
		}
	}
	return maxW, maxH, nil
}

func setChannel(dest *image.RGBA64, pic image.Image, channel int) {
	w := pic.Bounds().Dx()
	h := pic.Bounds().Dy()
	var c color.RGBA64
	var grayscale uint16
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			grayscale = uint16(convertToGrayscale(pic.At(x, y)))
			c = dest.RGBA64At(x, y)
			switch channel {
			case 0:
				c.R = grayscale
			case 1:
				c.G = grayscale
			case 2:
				c.B = grayscale
			default:
				panic("Bad channel")
			}
			dest.Set(x, y, c)
		}
	}
}

func combineImages(imageFiles []string) (image.Image, error) {
	var pic image.Image
	var f *os.File
	if len(imageFiles) > 3 {
		return nil, fmt.Errorf("Using more than 3 channels is unsupported.")
	}
	w, h, e := getMaxDimensions(imageFiles)
	if e != nil {
		return nil, fmt.Errorf("Failed getting image dimensions: %s", e)
	}
	fmt.Printf("Combining images into a %dx%d image.\n", w, h)
	combined := image.NewRGBA64(image.Rect(0, 0, w, h))
	for i, imageFile := range imageFiles {
		fmt.Printf("Setting channel %d using %s...\n", i+1, imageFile)
		f, e = os.Open(imageFile)
		if e != nil {
			return nil, fmt.Errorf("Failed opening file %s: %s", imageFile, e)
		}
		pic, _, e = image.Decode(f)
		if e != nil {
			f.Close()
			return nil, fmt.Errorf("Failed decoding image %s: %s", imageFile,
				e)
		}
		setChannel(combined, pic, i)
		pic = nil
		f.Close()
	}
	return combined, nil
}

func run() int {
	var redName string
	var greenName string
	var blueName string
	var outputName string
	flag.StringVar(&redName, "r", "", "The image to use for the R channel.")
	flag.StringVar(&greenName, "g", "", "The image to use for the G channel.")
	flag.StringVar(&blueName, "b", "", "The image to use for the B channel.")
	flag.StringVar(&outputName, "output", "", "The filename to create. "+
		"Output files will be JPEG-format.")
	flag.Parse()
	if (redName == "") || (greenName == "") || (blueName == "") {
		fmt.Printf("An image must be supplied for every color.\n")
		fmt.Printf("Run with -h for more information.\n")
		return 1
	}
	if outputName == "" {
		fmt.Printf("An output filename is required.\n")
		fmt.Printf("Run with -h for more information.\n")
		return 1
	}
	outputImage, e := combineImages([]string{redName, greenName, blueName})
	if e != nil {
		fmt.Printf("Error combining images: %s\n", e)
		return 1
	}
	outputFile, e := os.Create(outputName)
	if e != nil {
		fmt.Printf("Error opening output file: %s\n", e)
		return 1
	}
	defer outputFile.Close()
	options := jpeg.Options{
		Quality: 100,
	}
	e = jpeg.Encode(outputFile, outputImage, &options)
	if e != nil {
		fmt.Printf("Failed creating output JPEG image: %s\n", e)
		return 1
	}
	return 0
}

func main() {
	os.Exit(run())
}
