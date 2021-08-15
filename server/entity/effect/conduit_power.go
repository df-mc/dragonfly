package effect

import (
	"image/color"
)

// ConduitPower is a lasting effect that grants the affected entity the ability to breathe underwater and
// allows the entity to break faster when underwater or in the rain. (Similarly to haste.)
type ConduitPower struct {
	nopLasting
}

// Multiplier returns the mining speed multiplier from this effect.
func (ConduitPower) Multiplier(lvl int) float64 {
	v := 1 - float64(lvl)*0.1
	if v < 0 {
		v = 0
	}
	return v
}

// RGBA ...
func (ConduitPower) RGBA() color.RGBA {
	return color.RGBA{R: 0x1d, G: 0xc2, B: 0xd1, A: 0xff}
}
