package effect

import (
	"image/color"
)

// Blindness is a lasting effect that greatly reduces the vision range of the
// entity affected.
var Blindness blindness

type blindness struct {
	nopLasting
}

// RGBA ...
func (blindness) RGBA() color.RGBA {
	return color.RGBA{R: 0x1f, G: 0x1f, B: 0x23, A: 0xff}
}
