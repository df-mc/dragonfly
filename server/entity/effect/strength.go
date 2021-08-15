package effect

import (
	"image/color"
)

// Strength is a lasting effect that increases the damage dealt with melee attacks when applied to an entity.
type Strength struct {
	nopLasting
}

// Multiplier returns the damage multiplier of the effect.
func (Strength) Multiplier(lvl int) float64 {
	return 0.3 * float64(lvl)
}

// RGBA ...
func (Strength) RGBA() color.RGBA {
	return color.RGBA{R: 0x93, G: 0x24, B: 0x23, A: 0xff}
}
