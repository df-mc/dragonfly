package effect

import (
	"github.com/df-mc/dragonfly/dragonfly/entity"
	"image/color"
	"time"
)

// HealthBoost causes the affected entity to have its maximum health changed for a specific duration.
type HealthBoost struct {
	lastingEffect
}

// Start ...
func (h HealthBoost) Start(e entity.Living) {
	e.SetMaxHealth(e.MaxHealth() + 4*float64(h.Lvl))
}

// End ...
func (h HealthBoost) End(e entity.Living) {
	e.SetMaxHealth(e.MaxHealth() - 4*float64(h.Lvl))
}

// WithDurationAndLevel ...
func (h HealthBoost) WithDurationAndLevel(d time.Duration, level int) entity.Effect {
	return HealthBoost{h.withDurationAndLevel(d, level)}
}

// RGBA ...
func (HealthBoost) RGBA() color.RGBA {
	return color.RGBA{R: 0xf8, G: 0x7d, B: 0x23, A: 0xff}
}
