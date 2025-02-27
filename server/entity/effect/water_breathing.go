package effect

import (
	"image/color"
)

// WaterBreathing is a lasting effect that allows the affected entity to breath
// underwater until the effect expires.
var WaterBreathing waterBreathing

type waterBreathing struct {
	nopLasting
}

// RGBA ...
func (waterBreathing) RGBA() color.RGBA {
	return color.RGBA{R: 0x98, G: 0xda, B: 0xc0, A: 0xff}
}
