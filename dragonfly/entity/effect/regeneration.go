package effect

import (
	"github.com/df-mc/dragonfly/dragonfly/entity/healing"
	"github.com/df-mc/dragonfly/dragonfly/world"
	"image/color"
	"time"
)

// Regeneration is an effect that causes the entity that it is added to to slowly regenerate health. The level
// of the effect influences the speed with which the entity regenerates.
type Regeneration struct {
	lastingEffect
}

// Apply applies health to the world.Entity passed if the duration of the effect is at the right tick.
func (r Regeneration) Apply(e world.Entity) {
	interval := 50 >> r.Lvl
	if tickDuration(r.Dur)%interval == 0 {
		if living, ok := e.(living); ok {
			living.Heal(1, healing.SourceRegenerationEffect{})
		}
	}
}

// WithSettings ...
func (r Regeneration) WithSettings(d time.Duration, level int, ambient bool) Effect {
	return Regeneration{r.withSettings(d, level, ambient)}
}

// RGBA ...
func (Regeneration) RGBA() color.RGBA {
	return color.RGBA{R: 0xcd, G: 0x5c, B: 0xab, A: 0xff}
}
