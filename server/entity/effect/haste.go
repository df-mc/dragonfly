package effect

import (
	"image/color"
)

// Haste is a lasting effect that increases the mining speed of a player by 20% for each level of the effect.
type Haste struct {
	nopLasting
}

// Multiplier returns the mining speed multiplier from this effect.
func (Haste) Multiplier(lvl int) float64 {
	v := 1 - float64(lvl)*0.1
	if v < 0 {
		v = 0
	}
	return v
}

// RGBA ...
func (Haste) RGBA() color.RGBA {
	return color.RGBA{R: 0xd9, G: 0xc0, B: 0x43, A: 0xff}
}
