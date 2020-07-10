package effect

import (
	"github.com/df-mc/dragonfly/dragonfly/entity"
	"image/color"
	"math"
	"time"
)

// MiningFatigue is a lasting effect that decreases the mining speed of a player by 10% for each level of the
// effect.
type MiningFatigue struct {
	lastingEffect
}

// Multiplier returns the mining speed multiplier from this effect.
func (m MiningFatigue) Multiplier() float64 {
	return math.Pow(3, float64(m.Lvl))
}

// WithDuration ...
func (m MiningFatigue) WithDuration(d time.Duration) entity.Effect {
	return MiningFatigue{m.withDuration(d)}
}

// RGBA ...
func (MiningFatigue) RGBA() color.RGBA {
	return color.RGBA{R: 0x4a, G: 0x42, B: 0x17, A: 0xff}
}
