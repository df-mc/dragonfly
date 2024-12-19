package effect

import (
	"github.com/df-mc/dragonfly/server/world"
	"image/color"
)

// Saturation is a lasting effect that causes the affected player to regain
// food and saturation.
var Saturation saturation

type saturation struct {
	nopLasting
}

// Apply ...
func (saturation) Apply(e world.Entity, eff Effect) {
	if i, ok := e.(interface {
		Saturate(food int, saturation float64)
	}); ok {
		i.Saturate(eff.Level(), 2*float64(eff.Level()))
	}
}

// RGBA ...
func (saturation) RGBA() color.RGBA {
	return color.RGBA{R: 0xf8, G: 0x24, B: 0x23, A: 0xff}
}
