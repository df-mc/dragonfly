package effect

import (
	"github.com/df-mc/dragonfly/dragonfly/entity"
	"image/color"
	"time"
)

// Saturation is a lasting effect that causes the affected player to regain food and saturation.
type Saturation struct {
	lastingEffect
}

// Apply ...
func (s Saturation) Apply(e entity.Living) {
	if i, ok := e.(interface {
		Saturate(food int, saturation float64)
	}); ok {
		i.Saturate(1*s.Lvl, 2*float64(s.Lvl))
	}
}

// WithDurationAndLevel ...
func (s Saturation) WithDurationAndLevel(d time.Duration, level int) entity.Effect {
	return Saturation{s.withDurationAndLevel(d, level)}
}

// RGBA ...
func (s Saturation) RGBA() color.RGBA {
	return color.RGBA{R: 0xf8, G: 0x24, B: 0x23, A: 0xff}
}
