package effect

import (
	"github.com/df-mc/dragonfly/server/world"
	"image/color"
)

// HealthBoost causes the affected entity to have its maximum health changed for a specific duration.
type HealthBoost struct {
	nopLasting
}

// Start ...
func (HealthBoost) Start(e world.Entity, lvl int) {
	if l, ok := e.(living); ok {
		l.SetMaxHealth(l.MaxHealth() + 4*float64(lvl))
	}
}

// End ...
func (HealthBoost) End(e world.Entity, lvl int) {
	if l, ok := e.(living); ok {
		l.SetMaxHealth(l.MaxHealth() - 4*float64(lvl))
	}
}

// RGBA ...
func (HealthBoost) RGBA() color.RGBA {
	return color.RGBA{R: 0xf8, G: 0x7d, B: 0x23, A: 0xff}
}
