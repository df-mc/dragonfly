package effect

import (
	"image/color"
)

// JumpBoost is a lasting effect that causes the affected entity to be able to
// jump much higher, depending on the level of the effect.
var JumpBoost jumpBoost

type jumpBoost struct {
	nopLasting
}

// RGBA ...
func (jumpBoost) RGBA() color.RGBA {
	return color.RGBA{R: 0x22, G: 0xff, B: 0x4c, A: 0xff}
}
