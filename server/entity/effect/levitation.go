package effect

import (
	"image/color"
)

// Levitation is a lasting effect that causes the affected entity to slowly levitate upwards. It is roughly
// the opposite of the SlowFalling effect.
type Levitation struct {
	nopLasting
}

// RGBA ...
func (Levitation) RGBA() color.RGBA {
	return color.RGBA{R: 0xce, G: 0xff, B: 0xff, A: 0xff}
}
