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

// WithSettings ...
func (b Blindness) WithSettings(d time.Duration, level int, ambient bool) entity.Effect {
	return Blindness{b.withSettings(d, level, ambient)}
}

// RGBA ...
func (Blindness) RGBA() color.RGBA {
	return color.RGBA{R: 0x1f, G: 0x1f, B: 0x23, A: 0xff}
}
