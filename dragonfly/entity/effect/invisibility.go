package effect

import (
	"github.com/df-mc/dragonfly/dragonfly/world"
	"image/color"
	"time"
)

// Invisibility is a lasting effect that causes the affected entity to turn invisible. While invisible, the
// entity's armour is still visible and effect particles will still be displayed.
type Invisibility struct {
	lastingEffect
}

// Start ...
func (Invisibility) Start(e world.Entity) {
	if i, ok := e.(interface {
		SetInvisible()
		SetVisible()
	}); ok {
		i.SetInvisible()
	}
}

// End ...
func (Invisibility) End(e world.Entity) {
	if i, ok := e.(interface {
		SetInvisible()
		SetVisible()
	}); ok {
		i.SetVisible()
	}
}

// WithSettings ...
func (i Invisibility) WithSettings(d time.Duration, level int, ambient bool) Effect {
	return Invisibility{i.withSettings(d, level, ambient)}
}

// RGBA ...
func (Invisibility) RGBA() color.RGBA {
	return color.RGBA{R: 0x7f, G: 0x83, B: 0x92, A: 0xff}
}
