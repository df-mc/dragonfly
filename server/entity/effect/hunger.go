package effect

import (
	"github.com/df-mc/dragonfly/server/world"
	"image/color"
	"time"
)

// Hunger is a lasting effect that causes an affected player to gradually lose saturation and food.
type Hunger struct {
	nopLasting
}

// Apply ...
func (Hunger) Apply(e world.Entity, lvl int, _ time.Duration) {
	if i, ok := e.(interface {
		Exhaust(points float64)
	}); ok {
		i.Exhaust(float64(lvl) * 0.005)
	}
}

// RGBA ...
func (Hunger) RGBA() color.RGBA {
	return color.RGBA{R: 0x58, G: 0x76, B: 0x53, A: 0xff}
}
