package skin

import (
	"image"
	"image/color"
)

// Cape represents the cape that a skin may additionally have. A skin is of a fixed size (always 32x64 bytes)
// and may be either empty or of that size.
type Cape struct {
	w, h int

	// Pix holds the colour data of the cape in an RGBA byte array, similarly to the way that the pixels of
	// a Skin are stored.
	// The size of Pix is always 32 * 64 * 4 bytes.
	Pix []uint8
}

// NewCape initialises a new Cape using the width and height passed. The pixels are pre-allocated so that the
// Cape may be used immediately.
func NewCape(width, height int) Cape {
	return Cape{w: width, h: height, Pix: make([]uint8, width*height*4)}
}

// ColorModel ...
func (c Cape) ColorModel() color.Model {
	return color.RGBAModel
}

// Bounds returns the bounds of the cape, which is always 32x64 or 0x0, depending on if the cape has any data
// in it.
func (c Cape) Bounds() image.Rectangle {
	return image.Rectangle{
		Max: image.Point{X: c.w, Y: c.h},
	}
}

// At returns the colour at a given position in the cape, provided the X and Y are within the bounds of the
// cape passed during construction.
// The concrete type returned by At is a color.RGBA value.
func (c Cape) At(x, y int) color.Color {
	if x < 0 || y < 0 || x >= c.w || y >= c.h {
		panic("pixel coordinates out of bounds")
	}
	offset := x*4 + c.w*y*4
	return color.RGBA{
		R: c.Pix[offset],
		G: c.Pix[offset+1],
		B: c.Pix[offset+2],
		A: c.Pix[offset+3],
	}
}
