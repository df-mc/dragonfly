package effect

import (
	"github.com/df-mc/dragonfly/dragonfly/entity"
	"image/color"
	"time"
)

// Hunger is a lasting effect that causes an affected player to gradually lose saturation and food.
type Hunger struct {
	lastingEffect
}

// Apply ...
func (h Hunger) Apply(e entity.Living) {
	v := float64(h.Lvl) * 0.005
	if i, ok := e.(interface {
		Exhaust(points float64)
	}); ok {
		i.Exhaust(v)
	}
}

// WithDurationAndLevel ...
func (h Hunger) WithDurationAndLevel(d time.Duration, level int) entity.Effect {
	return Hunger{h.withDurationAndLevel(d, level)}
}

// RGBA ...
func (Hunger) RGBA() color.RGBA {
	return color.RGBA{R: 0x58, G: 0x76, B: 0x53, A: 0xff}
}
