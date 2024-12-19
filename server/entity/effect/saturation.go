package effect

import (
	"github.com/df-mc/dragonfly/server/world"
	"image/color"
)

// Saturation is a lasting effect that causes the affected player to regain food and saturation.
type Saturation struct {
	nopLasting
}

// Apply ...
func (Saturation) Apply(e world.Entity, eff Effect) {
	if i, ok := e.(interface {
		Saturate(food int, saturation float64)
	}); ok {
		i.Saturate(eff.Level(), 2*float64(eff.Level()))
	}
}

// RGBA ...
func (Saturation) RGBA() color.RGBA {
	return color.RGBA{R: 0xf8, G: 0x24, B: 0x23, A: 0xff}
}
