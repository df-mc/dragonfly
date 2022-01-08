package skin

import (
	"image"
	"image/color"
)

// Skin holds the data of a skin that a player has equipped. It includes geometry data, the texture and the
// cape, if one is present.
// Skin implements the image.Image interface to ease working with the value as an image.
type Skin struct {
	w, h int
	// Persona specifies if the skin uses the persona skin system.
	Persona   bool
	PlayFabID string

	// Pix holds the raw pixel data of the skin. This is an RGBA byte slice, meaning that every first byte is
	// a Red value, the second a Green value, the third a Blue value and the fourth an Alpha value.
	Pix []uint8

	// ModelConfig specifies how the Model field below should be used to form the total skin.
	ModelConfig ModelConfig
	// Model holds the raw JSON data that represents the model of the skin. If empty, it means the skin holds
	// the standard skin data (geometry.humanoid).
	// TODO: Write a full API for this. The model should be able to be easily modified or created runtime.
	Model []byte

	// Cape holds the cape of the skin. By default, an empty cape is set in the skin. Cape.Exists() may be
	// called to check if the cape actually has any data.
	Cape Cape

	// Animations holds a list of all animations that the skin has. These animations must be pointed to in the
	// ModelConfig, in order to display them on the skin.
	Animations []Animation
}

// New creates a new skin using the width and height passed. The dimensions passed must be either 64x32,
// 64x64 or 128x128. An error is returned if other dimensions are used.
// The skin pixels are initialised for the skin, and a random skin ID is picked. The model name and model is
// left empty.
func New(width, height int) Skin {
	return Skin{
		w:   width,
		h:   height,
		Pix: make([]uint8, width*height*4),
	}
}

// Bounds returns the bounds of the skin. These are either 64x32, 64x64 or 128, depending on the bounds of the
// skin of the player.
func (s Skin) Bounds() image.Rectangle {
	return image.Rectangle{
		Max: image.Point{X: s.w, Y: s.h},
	}
}

// ColorModel returns color.RGBAModel.
func (s Skin) ColorModel() color.Model {
	return color.RGBAModel
}

// At returns the colour at a given position in the skin. The concrete value of the colour returned is a color.RGBA
// value.
// If the x or y values exceed the bounds of the skin, At will panic.
func (s Skin) At(x, y int) color.Color {
	if x < 0 || y < 0 || x >= s.w || y >= s.h {
		panic("pixel coordinates out of bounds")
	}
	offset := x*4 + s.w*y*4
	return color.RGBA{
		R: s.Pix[offset],
		G: s.Pix[offset+1],
		B: s.Pix[offset+2],
		A: s.Pix[offset+3],
	}
}
