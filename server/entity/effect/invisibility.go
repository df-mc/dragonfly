package effect

import (
	"github.com/df-mc/dragonfly/server/world"
	"image/color"
)

// Invisibility is a lasting effect that causes the affected entity to turn
// invisible. While invisible, the entity's armour is still visible and effect
// particles will still be displayed.
var Invisibility invisibility

type invisibility struct {
	nopLasting
}

// Start ...
func (invisibility) Start(e world.Entity, _ int) {
	if i, ok := e.(interface {
		SetInvisible()
		SetVisible()
	}); ok {
		i.SetInvisible()
	}
}

// End ...
func (invisibility) End(e world.Entity, _ int) {
	if i, ok := e.(interface {
		SetInvisible()
		SetVisible()
	}); ok {
		i.SetVisible()
	}
}

// RGBA ...
func (invisibility) RGBA() color.RGBA {
	return color.RGBA{R: 0xf6, G: 0xf6, B: 0xf6, A: 0xff}
}
