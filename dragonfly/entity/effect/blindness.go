package effect

import (
	"github.com/df-mc/dragonfly/dragonfly/entity"
	"image/color"
	"time"
)

// Blindness is a lasting effect that greatly reduces the vision range of the entity affected.
type Blindness struct {
	lastingEffect
}

// WithDurationAndLevel ...
func (b Blindness) WithDurationAndLevel(d time.Duration, level int) entity.Effect {
	return Blindness{b.withDurationAndLevel(d, level)}
}

// RGBA ...
func (Blindness) RGBA() color.RGBA {
	return color.RGBA{R: 0x1f, G: 0x1f, B: 0x23, A: 0xff}
}
