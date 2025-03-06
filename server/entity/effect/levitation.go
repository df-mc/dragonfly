package effect

import (
	"image/color"
)

// Levitation is a lasting effect that causes the affected entity to slowly
// levitate upwards. It is roughly the opposite of the slowFalling effect.
var Levitation levitation

type levitation struct {
	nopLasting
}

// RGBA ...
func (levitation) RGBA() color.RGBA {
	return color.RGBA{R: 0xce, G: 0xff, B: 0xff, A: 0xff}
}
