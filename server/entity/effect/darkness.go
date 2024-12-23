package effect

import (
	"image/color"
)

// Darkness is a lasting effect that causes the player's vision to dim
// occasionally.
var Darkness darkness

type darkness struct {
	nopLasting
}

// RGBA ...
func (darkness) RGBA() color.RGBA {
	return color.RGBA{R: 0x29, G: 0x27, B: 0x21, A: 0xff}
}
