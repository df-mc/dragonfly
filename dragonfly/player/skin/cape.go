package skin

import (
	"errors"
	"image"
	"image/color"
)

// Cape represents the cape that a skin may additionally have. A skin is of a fixed size (always 32x64 bytes)
// and may be either empty or of that size.
type Cape struct {
	// Pix holds the colour data of the cape in an RGBA byte array, similarly to the way that the pixels of
	// a Skin are stored.
	// The size of Pix is always 32 * 64 * 4 bytes.
	Pix []uint8
}

// NewCape initialises a new Cape with the correct size of the Pix byte slice. The pixels themselves are
// filled with zero bytes.
func NewCape() Cape {
	return Cape{Pix: make([]uint8, 32*64*4)}
}

// NewCapeFromBytes creates a new Cape from the byte slice passed. The cape returned will be initialised with
// this data, provided it is of the correct size (either 64x32 or 0x0). If it is not, an error is returned.
func NewCapeFromBytes(p []uint8) (Cape, error) {
	if len(p) != 0 && len(p) != 64*32*4 {
		return Cape{}, errors.New("cape dimensions must be either 0x0 or 64x32")
	}
	return Cape{Pix: p}, nil
}

// Exists checks if the cape exists. In other words, it checks if the skin data is not empty.
func (c Cape) Exists() bool {
	return len(c.Pix) != 0
}

// ColorModel ...
func (c Cape) ColorModel() color.Model {
	return color.RGBAModel
}

// Bounds returns the bounds of the cape, which is always 32x64 or 0x0, depending on if the cape has any data
// in it.
func (c Cape) Bounds() image.Rectangle {
	if !c.Exists() {
		return image.Rectangle{}
	}
	return image.Rectangle{
		Max: image.Point{X: 32, Y: 64},
	}
}

// At returns the colour at a given position in the cape, provided the X and Y are within the bounds,
// 0 <= x < 32 and 0 <= y < 64, and the cape exists.
// The concrete type returned by At is a color.RGBA value.
func (c Cape) At(x, y int) color.Color {
	if x < 0 || y < 0 || x >= 32 || y >= 64 || !c.Exists() {
		panic("pixel coordinates out of bounds")
	}
	offset := x*4 + 32*y*4
	return color.RGBA{
		R: c.Pix[offset],
		G: c.Pix[offset+1],
		B: c.Pix[offset+2],
		A: c.Pix[offset+3],
	}
}
