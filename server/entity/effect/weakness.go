package effect

import (
	"image/color"
)

// Weakness is a lasting effect that reduces the damage dealt to other entities with melee attacks.
type Weakness struct {
	nopLasting
}

// Multiplier returns the damage multiplier of the effect.
func (Weakness) Multiplier(lvl int) float64 {
	v := 0.2 * float64(lvl)
	if v > 1 {
		v = 1
	}
	return v
}

// RGBA ...
func (Weakness) RGBA() color.RGBA {
	return color.RGBA{R: 0x48, G: 0x4d, B: 0x48, A: 0xff}
}
