package effect

import (
	"github.com/df-mc/dragonfly/server/world"
	"image/color"
)

// Speed is a lasting effect that increases the speed of an entity by 20% for each level that the effect has.
type Speed struct {
	nopLasting
}

// Start ...
func (Speed) Start(e world.Entity, lvl int) {
	speed := 1 + float64(lvl)*0.2
	if l, ok := e.(living); ok {
		l.SetSpeed(l.Speed() * speed)
	}
}

// End ...
func (Speed) End(e world.Entity, lvl int) {
	speed := 1 + float64(lvl)*0.2
	if l, ok := e.(living); ok {
		l.SetSpeed(l.Speed() / speed)
	}
}

// RGBA ...
func (Speed) RGBA() color.RGBA {
	return color.RGBA{R: 0x7c, G: 0xaf, B: 0xc6, A: 0xff}
}
