package effect

import (
	"image/color"
)

// WaterBreathing is a lasting effect that allows the affected entity to breath underwater until the effect
// expires.
type WaterBreathing struct {
	nopLasting
}

// RGBA ...
func (WaterBreathing) RGBA() color.RGBA {
	return color.RGBA{R: 0x2e, G: 0x52, B: 0x99, A: 0xff}
}
