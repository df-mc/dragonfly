package effect

import (
	"image/color"
)

// SlowFalling is a lasting effect that causes the affected entity to fall very slowly.
type SlowFalling struct {
	nopLasting
}

// RGBA ...
func (SlowFalling) RGBA() color.RGBA {
	return color.RGBA{R: 0xf7, G: 0xf8, B: 0xe0, A: 0xff}
}
