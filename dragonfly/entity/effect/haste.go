package effect

import (
	"github.com/df-mc/dragonfly/dragonfly/entity"
	"image/color"
	"time"
)

// Haste is a lasting effect that increases the mining speed of a player by 20% for each level of the effect.
type Haste struct {
	lastingEffect
}

// Multiplier returns the mining speed multiplier from this effect.
func (h Haste) Multiplier() float64 {
	v := 1 - float64(h.Lvl)*0.1
	if v < 0 {
		v = 0
	}
	return v
}

// WithDurationAndLevel ...
func (h Haste) WithDurationAndLevel(d time.Duration, level int) entity.Effect {
	return Haste{h.withDurationAndLevel(d, level)}
}

// RGBA ...
func (Haste) RGBA() color.RGBA {
	return color.RGBA{R: 0xd9, G: 0xc0, B: 0x43, A: 0xff}
}
