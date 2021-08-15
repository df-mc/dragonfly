package effect

import (
	"github.com/df-mc/dragonfly/server/world"
	"image/color"
)

// Invisibility is a lasting effect that causes the affected entity to turn invisible. While invisible, the
// entity's armour is still visible and effect particles will still be displayed.
type Invisibility struct {
	nopLasting
}

// Start ...
func (Invisibility) Start(e world.Entity, _ int) {
	if i, ok := e.(interface {
		SetInvisible()
		SetVisible()
	}); ok {
		i.SetInvisible()
	}
}

// End ...
func (Invisibility) End(e world.Entity, _ int) {
	if i, ok := e.(interface {
		SetInvisible()
		SetVisible()
	}); ok {
		i.SetVisible()
	}
}

// RGBA ...
func (Invisibility) RGBA() color.RGBA {
	return color.RGBA{R: 0x7f, G: 0x83, B: 0x92, A: 0xff}
}
