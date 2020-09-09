package effect

import (
	"github.com/df-mc/dragonfly/dragonfly/world"
	"image/color"
	"time"
)

// HealthBoost causes the affected entity to have its maximum health changed for a specific duration.
type HealthBoost struct {
	lastingEffect
}

// Start ...
func (h HealthBoost) Start(e world.Entity) {
	if living, ok := e.(living); ok {
		living.SetMaxHealth(living.MaxHealth() + 4*float64(h.Lvl))
	}
}

// End ...
func (h HealthBoost) End(e world.Entity) {
	if living, ok := e.(living); ok {
		living.SetMaxHealth(living.MaxHealth() - 4*float64(h.Lvl))
	}
}

// WithSettings ...
func (h HealthBoost) WithSettings(d time.Duration, level int, ambient bool) Effect {
	return HealthBoost{h.withSettings(d, level, ambient)}
}

// RGBA ...
func (HealthBoost) RGBA() color.RGBA {
	return color.RGBA{R: 0xf8, G: 0x7d, B: 0x23, A: 0xff}
}
