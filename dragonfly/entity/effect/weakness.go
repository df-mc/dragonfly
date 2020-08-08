package effect

import (
	"image/color"
	"time"
)

// Weakness is a lasting effect that reduces the damage dealt to other entities with melee attacks.
type Weakness struct {
	lastingEffect
}

// Multiplier returns the damage multiplier of the effect.
func (w Weakness) Multiplier() float64 {
	v := -0.2 * float64(w.Lvl)
	if v < -1 {
		v = -1
	}
	return v
}

// WithSettings ...
func (w Weakness) WithSettings(d time.Duration, level int, ambient bool) Effect {
	return Weakness{w.withSettings(d, level, ambient)}
}

// RGBA ...
func (Weakness) RGBA() color.RGBA {
	return color.RGBA{R: 0x48, G: 0x4d, B: 0x48, A: 0xff}
}
