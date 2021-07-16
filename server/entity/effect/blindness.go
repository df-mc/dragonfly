package effect

import (
	"image/color"
	"time"
)

// Blindness is a lasting effect that greatly reduces the vision range of the entity affected.
type Blindness struct {
	lastingEffect
}

// WithSettings ...
func (b Blindness) WithSettings(d time.Duration, level int, ambient bool) Effect {
	return Blindness{b.withSettings(d, level, ambient)}
}

// RGBA ...
func (Blindness) RGBA() color.RGBA {
	return color.RGBA{R: 0x1f, G: 0x1f, B: 0x23, A: 0xff}
}
