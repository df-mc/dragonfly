package effect

import (
	"image/color"
)

// Blindness is a lasting effect that greatly reduces the vision range of the entity affected.
type Blindness struct {
	nopLasting
}

// RGBA ...
func (Blindness) RGBA() color.RGBA {
	return color.RGBA{R: 0x1f, G: 0x1f, B: 0x23, A: 0xff}
}
