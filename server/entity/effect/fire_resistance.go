package effect

import (
	"image/color"
	"time"
)

// FireResistance is a lasting effect that grants immunity to fire & lava damage.
type FireResistance struct {
	lastingEffect
}

// WithSettings ...
func (f FireResistance) WithSettings(d time.Duration, level int, ambient bool) Effect {
	return FireResistance{f.withSettings(d, level, ambient)}
}

// RGBA ...
func (f FireResistance) RGBA() color.RGBA {
	return color.RGBA{R: 0xe4, G: 0x9a, B: 0x3a, A: 0xff}
}
