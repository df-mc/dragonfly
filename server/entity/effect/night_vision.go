package effect

import (
	"image/color"
)

// NightVision is a lasting effect that causes the affected entity to see in
// dark places as though they were fully lit up.
var NightVision nightVision

type nightVision struct {
	nopLasting
}

// RGBA ...
func (nightVision) RGBA() color.RGBA {
	return color.RGBA{R: 0xc2, G: 0xff, B: 0x66, A: 0xff}
}
