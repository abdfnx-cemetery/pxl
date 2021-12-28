package pxl

import (
    "os"
    "io"
    "fmt"
    "image"
    "image/color"

    "github.com/pkg/errors"
)

// FromFile func is a convenience function that converts a file to a formatted string.
// See FromImage() for more details.
func FromFile(filename string) (encoded string, err error) {
    f, err := os.Open(filename)

    if err != nil {
        return
    }

    defer f.Close()
    return FromReader(io.Reader(f))
}


// FromReader is a convenience function that converts an io.Reader to a formatted string.
func FromReader(reader io.Reader) (encoded string, err error) {
    img, _, err := image.Decode(reader)
    if err != nil {
        return
    }

    return FromImage(img)
}

// FromImage is the core function of `pxl`,
// It takes an image.Image and converts it to a string formatted for tview.
// The unicode half-block character (▀) with a fg & bg colour set will represent
// pixels in the returned string.
// Because each character represents two pixels, it is not possible to convert an
func FromImage(img image.Image) (encoded string, err error) {
    if (img.Bounds().Max.Y - img.Bounds().Min.Y) % 2 != 0 {
        err = errors.New("pixelview: Can't process image with uneven height")
        return
    }

    switch v := img.(type) {
		default:
			return FromImageGeneric(img)

		case *image.Paletted:
			return FromPaletted(v)

		case *image.NRGBA:
			return FromNRGBA(v)
    }
}

// FromImageGeneric is the fallback function for processing images.
// It will be used for more exotic image formats than png or gif.
func FromImageGeneric(img image.Image) (encoded string, err error) {
    for y := img.Bounds().Min.Y; y < img.Bounds().Max.Y; y += 2 {
        var prevfg, prevbg color.Color
        for x := img.Bounds().Min.X; x < img.Bounds().Max.X; x++ {
            fg := img.At(x, y)
            bg := img.At(x, y + 1)
            encoded += Encode(fg, bg, &prevfg, &prevbg)
        }

        encoded += "\n"
    }

    return
}

// FromPaletted saves a few μs when working with paletted images.
// These are what PNG8 images are decoded as.
func FromPaletted(img *image.Paletted) (encoded string, err error) {
    for y := img.Rect.Min.Y; y < img.Rect.Max.Y; y += 2 {
        var prevfg, prevbg color.Color

        for x := img.Rect.Min.X; x < img.Rect.Max.X; x++ {
            i := (y - img.Rect.Min.Y) * img.Stride + (x - img.Rect.Min.X)
            fg := img.Palette[img.Pix[i]]
            bg := img.Palette[img.Pix[i + img.Stride]]
            encoded += Encode(fg, bg, &prevfg, &prevbg)
        }

        encoded += "\n"
    }

    return
}

// FromNRGBA saves a handful of μs when working with NRGBA images.
// These are what PNG24 images are decoded as.
func FromNRGBA(img *image.NRGBA) (encoded string, err error) {
    for y := img.Rect.Min.Y; y < img.Rect.Max.Y; y += 2 {
        var prevfg, prevbg color.Color

        for x := img.Rect.Min.X; x < img.Rect.Max.X; x++ {
            i := (y - img.Rect.Min.Y) * img.Stride + (x - img.Rect.Min.X) * 4
            fg := color.NRGBA{img.Pix[i], img.Pix[i+1], img.Pix[i+2], img.Pix[i+3]}
            i += img.Stride
            bg := color.NRGBA{img.Pix[i], img.Pix[i+1], img.Pix[i+2], img.Pix[i+3]}
            encoded += Encode(fg, bg, &prevfg, &prevbg)
        }

        encoded += "\n"
    }

    return
}

// Encode converts a fg & bg colour into a formatted pair of 'pixels',
// using the prevfg & prevbg colours to perform something akin to run-length encoding
func Encode(fg, bg color.Color, prevfg, prevbg *color.Color) (encoded string) {
    if fg == *prevfg && bg == *prevbg {
        encoded = "▀"
        return
    }

    if fg == *prevfg {
        encoded = fmt.Sprintf(
            "[:%s]▀",
            ColorHex(bg),
        )

        *prevbg = bg
        return
    }

    if bg == *prevbg {
        encoded = fmt.Sprintf(
            "[%s:]▀",
            ColorHex(fg),
        )

        *prevfg = fg
        return
    }

    encoded = fmt.Sprintf(
        "[%s:%s]▀",
        ColorHex(fg),
        ColorHex(bg),
    )

    *prevfg = fg
    *prevbg = bg
    return
}

func ColorHex(c color.Color) string {
    r, g, b, _ := c.RGBA()
    return fmt.Sprintf("#%.2x%.2x%.2x", r >> 8, g >> 8, b >> 8)
}
