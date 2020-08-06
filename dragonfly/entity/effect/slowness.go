package effect

import (
	"github.com/df-mc/dragonfly/dragonfly/world"
	"image/color"
	"time"
)

// Slowness is a lasting effect that decreases the movement speed of a living entity by 15% for each level
// that the effect has.
type Slowness struct {
	lastingEffect
}

// Start ...
func (s Slowness) Start(e world.Entity) {
	slowness := 1 - float64(s.Lvl)*0.15
	if slowness <= 0 {
		slowness = 0.00001
	}
	if living, ok := e.(living); ok {
		living.SetSpeed(living.Speed() * slowness)
	}
}

// Stop ...
func (s Slowness) Stop(e world.Entity) {
	slowness := 1 - float64(s.Lvl)*0.15
	if slowness <= 0 {
		slowness = 0.00001
	}
	if living, ok := e.(living); ok {
		living.SetSpeed(living.Speed() / slowness)
	}
}

// WithSettings ...
func (s Slowness) WithSettings(d time.Duration, level int, ambient bool) Effect {
	return Slowness{s.withSettings(d, level, ambient)}
}

// RGBA ...
func (Slowness) RGBA() color.RGBA {
	return color.RGBA{R: 0x5a, G: 0x6c, B: 0x81, A: 0xff}
}
