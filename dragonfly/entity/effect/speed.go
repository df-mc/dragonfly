package effect

import (
	"github.com/df-mc/dragonfly/dragonfly/entity"
	"image/color"
	"time"
)

// Speed is a lasting effect that increases the speed of an entity by 20% for each level that the effect has.
type Speed struct {
	lastingEffect
}

// Start ...
func (s Speed) Start(e entity.Living) {
	speed := 1 + float64(s.Lvl)*0.2
	e.SetSpeed(e.Speed() * speed)
}

// End ...
func (s Speed) End(e entity.Living) {
	speed := 1 + float64(s.Lvl)*0.2
	e.SetSpeed(e.Speed() / speed)
}

// WithDuration ...
func (s Speed) WithDuration(d time.Duration) entity.Effect {
	return Speed{s.withDuration(d)}
}

// RGBA ...
func (s Speed) RGBA() color.RGBA {
	return color.RGBA{R: 0x7c, G: 0xaf, B: 0xc6, A: 0xff}
}
