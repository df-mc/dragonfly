package effect

import (
	"github.com/df-mc/dragonfly/server/world"
	"image/color"
)

// Resistance is a lasting effect that reduces the damage taken from any
// sources except for void damage or custom damage.
var Resistance resistance

type resistance struct {
	nopLasting
}

// Multiplier returns a damage multiplier for the damage source passed.
func (resistance) Multiplier(e world.DamageSource, lvl int) float64 {
	if !e.ReducedByResistance() {
		return 1
	}
	if v := 1 - 0.2*float64(lvl); v >= 0 {
		return v
	}
	return 0
}

// RGBA ...
func (resistance) RGBA() color.RGBA {
	return color.RGBA{R: 0x91, G: 0x46, B: 0xf0, A: 0xff}
}
