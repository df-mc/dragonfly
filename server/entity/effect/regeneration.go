package effect

import (
	"github.com/df-mc/dragonfly/server/world"
	"image/color"
	"time"
)

// Regeneration is an effect that causes the entity that it is added to slowly regenerate health. The level
// of the effect influences the speed with which the entity regenerates.
type Regeneration struct {
	nopLasting
}

// Apply applies health to the world.Entity passed if the duration of the effect is at the right tick.
func (Regeneration) Apply(e world.Entity, lvl int, d time.Duration) {
	interval := 50 >> (lvl - 1)
	if interval < 1 {
		interval = 1
	}
	if tickDuration(d)%interval == 0 {
		if l, ok := e.(living); ok {
			l.Heal(1, RegenerationHealingSource{})
		}
	}
}

// RGBA ...
func (Regeneration) RGBA() color.RGBA {
	return color.RGBA{R: 0xcd, G: 0x5c, B: 0xab, A: 0xff}
}

// RegenerationHealingSource is a healing source used when an entity regenerates
// health from an effect.Regeneration.
type RegenerationHealingSource struct{}

func (RegenerationHealingSource) HealingSource() {}
