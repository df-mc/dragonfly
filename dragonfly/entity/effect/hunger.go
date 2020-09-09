package effect

import (
	"github.com/df-mc/dragonfly/dragonfly/world"
	"image/color"
	"time"
)

// Hunger is a lasting effect that causes an affected player to gradually lose saturation and food.
type Hunger struct {
	lastingEffect
}

// Apply ...
func (h Hunger) Apply(e world.Entity) {
	v := float64(h.Lvl) * 0.005
	if i, ok := e.(interface {
		Exhaust(points float64)
	}); ok {
		i.Exhaust(v)
	}
}

// WithSettings ...
func (h Hunger) WithSettings(d time.Duration, level int, ambient bool) Effect {
	return Hunger{h.withSettings(d, level, ambient)}
}

// RGBA ...
func (Hunger) RGBA() color.RGBA {
	return color.RGBA{R: 0x58, G: 0x76, B: 0x53, A: 0xff}
}
