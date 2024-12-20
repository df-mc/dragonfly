package effect

import (
	"github.com/df-mc/dragonfly/server/world"
	"image/color"
)

// Regeneration is an effect that causes the entity that it is added to slowly
// regenerate health. The level of the effect influences the speed with which
// the entity regenerates.
var Regeneration regeneration

type regeneration struct {
	nopLasting
}

// Apply applies health to the world.Entity passed if the duration of the effect is at the right tick.
func (regeneration) Apply(e world.Entity, eff Effect) {
	interval := max(50>>(eff.Level()-1), 1)
	if eff.Tick()%interval == 0 {
		if l, ok := e.(living); ok {
			l.Heal(1, RegenerationHealingSource{})
		}
	}
}

// RGBA ...
func (regeneration) RGBA() color.RGBA {
	return color.RGBA{R: 0xcd, G: 0x5c, B: 0xab, A: 0xff}
}

// RegenerationHealingSource is a healing source used when an entity regenerates
// health from an effect.regeneration.
type RegenerationHealingSource struct{}

func (RegenerationHealingSource) HealingSource() {}
