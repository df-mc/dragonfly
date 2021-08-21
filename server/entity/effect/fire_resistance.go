package effect

import (
	"image/color"
)

// FireResistance is a lasting effect that grants immunity to fire & lava damage.
type FireResistance struct {
	nopLasting
}

// RGBA ...
func (FireResistance) RGBA() color.RGBA {
	return color.RGBA{R: 0xe4, G: 0x9a, B: 0x3a, A: 0xff}
}
