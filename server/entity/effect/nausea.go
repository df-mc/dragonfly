package effect

import (
	"image/color"
)

// Nausea is a lasting effect that causes the screen to warp, similarly to when entering a nether portal.
type Nausea struct {
	nopLasting
}

// RGBA ...
func (Nausea) RGBA() color.RGBA {
	return color.RGBA{R: 0x55, G: 0x1d, B: 0x4a, A: 0xff}
}
