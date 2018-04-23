// The image combiner program takes multiple images and an associated color for
// each. It multiplies the overall brightness for each pixel in each input
// image by the corresponding color for that image. All such colored pixels are
// added together in the output image.
package main

import (
	"fmt"
	_ "github.com/spakin/netpbm"
	_ "golang.org/x/image/bmp"
	"golang.org/x/image/colornames"
	"image"
	"image/color"
	_ "image/gif"
	"image/jpeg"
	_ "image/png"
	"os"
	"strconv"
	"strings"
)

// Implements the color interface, but uses floating-point colors for easier
// multiplication.
type floatColor struct {
	r float32
	g float32
	b float32
}

func (c floatColor) Add(toAdd color.Color) floatColor {
	converted := convertToFloatColor(toAdd)
	return floatColor{
		r: c.r + converted.r,
		g: c.g + converted.g,
		b: c.b + converted.b,
	}
}

func (c floatColor) Multiply(scale color.Color) floatColor {
	converted := convertToFloatColor(scale)
	return floatColor{
		r: c.r * converted.r,
		g: c.g * converted.g,
		b: c.b * converted.b,
	}
}

func (c floatColor) Scale(scale float32) floatColor {
	return floatColor{
		r: c.r * scale,
		g: c.g * scale,
		b: c.b * scale,
	}
}

func (c floatColor) RGBA() (r, g, b, a uint32) {
	var red, green, blue uint32
	if c.r >= 1.0 {
		red = 0xffff
	} else {
		red = uint32(c.r * float32(0xffff))
	}
	if c.g >= 1.0 {
		green = 0xffff
	} else {
		green = uint32(c.g * float32(0xffff))
	}
	if c.b >= 1.0 {
		blue = 0xffff
	} else {
		blue = uint32(c.b * float32(0xffff))
	}
	return red, green, blue, 0xffff
}

func (c floatColor) String() string {
	return fmt.Sprintf("%04x%04x%04x", uint16(c.r*0xffff), uint16(c.g*0xffff),
		uint16(c.b*0xffff))
}

func convertToFloatColor(c color.Color) floatColor {
	tryResult, ok := c.(floatColor)
	if ok {
		return tryResult
	}
	r, g, b, _ := c.RGBA()
	return floatColor{
		r: float32(r) / 0xffff,
		g: float32(g) / 0xffff,
		b: float32(b) / 0xffff,
	}
}

type floatColorImage struct {
	pixels []floatColor
	w, h   int
}

func (f *floatColorImage) Bounds() image.Rectangle {
	return image.Rect(0, 0, f.w, f.h)
}

func (f *floatColorImage) ColorModel() color.Model {
	return color.ModelFunc(func(c color.Color) color.Color {
		return convertToFloatColor(c)
	})
}

func (f *floatColorImage) At(x, y int) color.Color {
	if (x < 0) || (y < 0) || (x >= f.w) || (y >= f.h) {
		return color.Black
	}
	return f.pixels[(y*f.w)+x]
}

func (f *floatColorImage) Add(x, y int, toAdd color.Color) {
	if (x < 0) || (y < 0) || (x >= f.w) || (y >= f.h) {
		return
	}
	pixel := f.pixels[(y*f.w)+x]
	f.pixels[(y*f.w)+x] = pixel.Add(toAdd)
}

func newFloatColorImage(w, h int) (*floatColorImage, error) {
	if (w <= 0) || (h <= 0) {
		return nil, fmt.Errorf("Image bounds must be positive")
	}
	return &floatColorImage{
		w:      w,
		h:      h,
		pixels: make([]floatColor, w*h),
	}, nil
}

func parse24BitColor(value string) (floatColor, error) {
	parsed, e := strconv.ParseUint(value, 16, 32)
	if e != nil {
		return floatColor{}, fmt.Errorf("Couldn't parse color %s: %s", value,
			e)
	}
	return floatColor{
		r: float32((parsed>>16)&0xff) / 255.0,
		g: float32((parsed>>8)&0xff) / 255.0,
		b: float32(parsed&0xff) / 255.0,
	}, nil
}

func parse48BitColor(value string) (floatColor, error) {
	parsed, e := strconv.ParseUint(value, 16, 64)
	if e != nil {
		return floatColor{}, fmt.Errorf("Couldn't parse color %s: %s", value,
			e)
	}
	return floatColor{
		r: float32((parsed>>32)&0xffff) / 65535.0,
		g: float32((parsed>>16)&0xffff) / 65535.0,
		b: float32(parsed&0xffff) / 65535.0,
	}, nil
}

// Attempts to parse a color using an SVG color name. Returns false if a color
// with the given name wasn't found.
func parseNamedColor(name string) (floatColor, bool) {
	name = strings.ToLower(name)
	namedColor := colornames.Map[name]
	// Since a map returns a zero-value if the key doesn't exist, and no
	// visible will have a zero alpha value, we use an alpha value of zero to
	// detect that the given name wasn't in the colornames map.
	_, _, _, a := namedColor.RGBA()
	if a == 0 {
		return convertToFloatColor(namedColor), false
	}
	return convertToFloatColor(namedColor), true
}

// Parses an input hex string with either 24-bit or 48-bit RGB color as a float
// color. Returns an error if the input value is invalid.
func parseFloatColor(value string) (floatColor, error) {
	// First check if a named color was given.
	namedColor, nameOK := parseNamedColor(value)
	if nameOK {
		return namedColor, nil
	}
	// Allow hex color values starting with a single '#'
	value = strings.TrimPrefix(value, "#")
	if len(value) == 6 {
		return parse24BitColor(value)
	}
	if len(value) == 12 {
		return parse48BitColor(value)
	}
	return floatColor{}, fmt.Errorf("Need a 24- or 48-bit RGB color, got %s",
		value)
}

// This contains a filename and parsed color value, parsed from the command
// line arguments.
type imageInput struct {
	filename   string
	colorValue floatColor
}

// Converts a given arbitrary RGB color to a single brightness value.
func convertToBrightness(c color.Color) float32 {
	r, g, b, _ := c.RGBA()
	return float32(r+g+b) / (3.0 * 65535.0)
}

// Takes 3 image filenames and returns the maximum dimensions of all of them.
func getMaxDimensions(imageFiles []imageInput) (int, int, error) {
	var maxW, maxH, w, h int
	var pic image.Image
	var e error
	var f *os.File
	for _, inputPic := range imageFiles {
		filename := inputPic.filename
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

func addColor(dest *floatColorImage, pic image.Image, addColor floatColor) {
	w := pic.Bounds().Dx()
	h := pic.Bounds().Dy()
	var scale float32
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			scale = convertToBrightness(pic.At(x, y))
			dest.Add(x, y, addColor.Scale(scale))
		}
	}
}

func combineImages(imageFiles []imageInput) (image.Image, error) {
	var pic image.Image
	var f *os.File
	w, h, e := getMaxDimensions(imageFiles)
	if e != nil {
		return nil, fmt.Errorf("Failed getting image dimensions: %s", e)
	}
	fmt.Printf("Combining images into a %dx%d image.\n", w, h)
	combined, e := newFloatColorImage(w, h)
	if e != nil {
		return nil, fmt.Errorf("Failed creating new image: %s", e)
	}
	for _, imageFile := range imageFiles {
		fmt.Printf("Setting color %s using %s...\n", imageFile.colorValue,
			imageFile.filename)
		f, e = os.Open(imageFile.filename)
		if e != nil {
			return nil, fmt.Errorf("Failed opening file %s: %s", imageFile, e)
		}
		pic, _, e = image.Decode(f)
		if e != nil {
			f.Close()
			return nil, fmt.Errorf("Failed decoding image %s: %s", imageFile,
				e)
		}
		addColor(combined, pic, imageFile.colorValue)
		pic = nil
		f.Close()
	}
	return combined, nil
}

func printUsage() {
	fmt.Printf("Usage: %s <image 1 path> <image 1 color> <image 2> "+
		"<image 2 color> ... <output filename.jpg>\n\n"+
		"The image colors may an SVG color name, 6 hex digits, or 12 hex "+
		"digits (for 48-bit color).\n", os.Args[0])
}

// Parses the command line arguments. Returns an error if the arguments are
// invalid for any reason. Returns a slice of input images and colors, the
// output filename, or an error if one occurs.
func parseArguments() ([]imageInput, string, error) {
	var e error
	if len(os.Args) <= 2 {
		return nil, "", fmt.Errorf("Invalid arguments: at least one " +
			"image/color must be provided")
	}
	if (len(os.Args) % 2) != 0 {
		return nil, "", fmt.Errorf("Invalid arguments: each image must have " +
			"a corresponding color")
	}
	outputName := os.Args[len(os.Args)-1]
	// Subtract the program name and output filename from the args array to get
	// the number of image and color arguments. Divide by 2 to get # of pairs.
	toReturn := make([]imageInput, (len(os.Args)-2)/2)
	var parsedColor floatColor
	for i := range toReturn {
		toReturn[i].filename = os.Args[(i*2)+1]
		parsedColor, e = parseFloatColor(os.Args[(i*2)+2])
		if e != nil {
			return nil, "", fmt.Errorf("Invalid color for image %s: %s",
				toReturn[i].filename, e)
		}
		toReturn[i].colorValue = parsedColor
	}
	return toReturn, outputName, nil
}

func run() int {
	toCombine, outputName, e := parseArguments()
	if e != nil {
		fmt.Printf("Failed parsing arguments: %s\n", e)
		printUsage()
		return 1
	}
	outputImage, e := combineImages(toCombine)
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
