package effect

import (
	"github.com/df-mc/dragonfly/dragonfly/world"
	"image/color"
	"time"
)

// Saturation is a lasting effect that causes the affected player to regain food and saturation.
type Saturation struct {
	lastingEffect
}

// Apply ...
func (s Saturation) Apply(e world.Entity) {
	if i, ok := e.(interface {
		Saturate(food int, saturation float64)
	}); ok {
		i.Saturate(1*s.Lvl, 2*float64(s.Lvl))
	}
}

// WithSettings ...
func (s Saturation) WithSettings(d time.Duration, level int, ambient bool) Effect {
	return Saturation{s.withSettings(d, level, ambient)}
}

// RGBA ...
func (s Saturation) RGBA() color.RGBA {
	return color.RGBA{R: 0xf8, G: 0x24, B: 0x23, A: 0xff}
}
