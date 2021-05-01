package skin

import (
	"image"
	"image/color"
)

const (
	// AnimationHead is an animation that is played over the head part of the skin.
	AnimationHead AnimationType = iota
	// AnimationBody32x32 is an animation that is played over the body of a skin with a 32x32(/64) size. This
	// is the usual animation type for body animations.
	AnimationBody32x32
	// AnimationBody128x128 is an animation that is played over a body of a skin with a 128x128 size. This is
	// the animation type for body animations with high resolution.
	AnimationBody128x128
)

// AnimationType represents a type of the animation. It is one of the constants above, and specifies to what
// part of the body it is assigned.
type AnimationType int

// Animation represents an animation that plays over the skin every so often. It is assigned to a particular
// part of the skin, which is represented by one of the constants above.
type Animation struct {
	w, h  int
	aType AnimationType

	// Pix holds skin data for every frame of the animation. This is an RGBA byte slice, meaning that every
	// first byte is a Red value, the second a Green value, the third a Blue value and the fourth an Alpha
	// value.
	Pix []uint8

	// FrameCount is the amount of frames that the animation plays for. Exactly this amount of frames should
	// be present in the Pix animation data.
	FrameCount int

	// AnimationExpression is the player's animation expression.
	AnimationExpression int
}

// NewAnimation returns a new animation using the width and height passed, with the type specifying what part
// of the body to display it on.
// NewAnimation fills out the Pix field adequately and sets FrameCount to 1 by default.
func NewAnimation(width, height int, expression int, animationType AnimationType) Animation {
	return Animation{
		w:                   width,
		h:                   height,
		aType:               animationType,
		Pix:                 make([]uint8, width*height*4),
		FrameCount:          1,
		AnimationExpression: expression,
	}
}

// Type returns the type of the animation, which is one of the constants above.
func (a Animation) Type() AnimationType {
	return a.aType
}

// ColorModel ...
func (a Animation) ColorModel() color.Model {
	return color.RGBAModel
}

// Bounds ...
func (a Animation) Bounds() image.Rectangle {
	return image.Rectangle{
		Max: image.Point{X: a.w, Y: a.h},
	}
}

// At returns the colour at a given position in the animation data, provided the X and Y are within the bounds
// of the animation passed during construction.
// The concrete type returned by At is a color.RGBA value.
func (a Animation) At(x, y int) color.Color {
	if x < 0 || y < 0 || x >= a.w || y >= a.h {
		panic("pixel coordinates out of bounds")
	}
	offset := x*4 + a.w*y*4
	return color.RGBA{
		R: a.Pix[offset],
		G: a.Pix[offset+1],
		B: a.Pix[offset+2],
		A: a.Pix[offset+3],
	}
}
