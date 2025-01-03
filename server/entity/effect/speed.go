package effect

import (
	"github.com/df-mc/dragonfly/server/world"
	"image/color"
)

// Speed is a lasting effect that increases the speed of an entity by 20% for
// each level that the effect has.
var Speed speed

type speed struct {
	nopLasting
}

// Start ...
func (speed) Start(e world.Entity, lvl int) {
	speed := 1 + float64(lvl)*0.2
	if l, ok := e.(living); ok {
		l.SetSpeed(l.Speed() * speed)
	}
}

// End ...
func (speed) End(e world.Entity, lvl int) {
	speed := 1 + float64(lvl)*0.2
	if l, ok := e.(living); ok {
		l.SetSpeed(l.Speed() / speed)
	}
}

// RGBA ...
func (speed) RGBA() color.RGBA {
	return color.RGBA{R: 0x33, G: 0xeb, B: 0xff, A: 0xff}
}
