package effect

import (
	"github.com/df-mc/dragonfly/dragonfly/world"
	"image/color"
	"time"
)

// Speed is a lasting effect that increases the speed of an entity by 20% for each level that the effect has.
type Speed struct {
	lastingEffect
}

// Start ...
func (s Speed) Start(e world.Entity) {
	speed := 1 + float64(s.Lvl)*0.2
	if living, ok := e.(living); ok {
		living.SetSpeed(living.Speed() * speed)
	}
}

// End ...
func (s Speed) End(e world.Entity) {
	speed := 1 + float64(s.Lvl)*0.2
	if living, ok := e.(living); ok {
		living.SetSpeed(living.Speed() / speed)
	}
}

// WithSettings ...
func (s Speed) WithSettings(d time.Duration, level int, ambient bool) Effect {
	return Speed{s.withSettings(d, level, ambient)}
}

// RGBA ...
func (Speed) RGBA() color.RGBA {
	return color.RGBA{R: 0x7c, G: 0xaf, B: 0xc6, A: 0xff}
}
