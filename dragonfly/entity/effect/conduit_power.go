package effect

import (
	"image/color"
	"time"
)

// ConduitPower is a lasting effect that grants the affected entity the ability to breathe underwater and
// allows the entity to break faster when underwater or in the rain. (Similarly to haste.)
type ConduitPower struct {
	lastingEffect
}

// Multiplier returns the mining speed multiplier from this effect.
func (c ConduitPower) Multiplier() float64 {
	v := 1 - float64(c.Lvl)*0.1
	if v < 0 {
		v = 0
	}
	return v
}

// WithSettings ...
func (c ConduitPower) WithSettings(d time.Duration, level int, ambient bool) Effect {
	return ConduitPower{c.withSettings(d, level, ambient)}
}

// RGBA ...
func (c ConduitPower) RGBA() color.RGBA {
	return color.RGBA{R: 0x1d, G: 0xc2, B: 0xd1, A: 0xff}
}
