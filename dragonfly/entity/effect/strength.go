package effect

import (
	"github.com/df-mc/dragonfly/dragonfly/entity"
	"image/color"
	"time"
)

// Strength is a lasting effect that increases the damage dealt with melee attacks when applied to an entity.
type Strength struct {
	lastingEffect
}

// Multiplier returns the damage multiplier of the effect.
func (s Strength) Multiplier() float64 {
	return 0.3 * float64(s.Lvl)
}

// WithDuration ...
func (s Strength) WithDuration(d time.Duration) entity.Effect {
	return Strength{s.withDuration(d)}
}

// RGBA ...
func (Strength) RGBA() color.RGBA {
	return color.RGBA{R: 0x93, G: 0x24, B: 0x23, A: 0xff}
}
