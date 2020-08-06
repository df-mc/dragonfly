package effect

import (
	"github.com/df-mc/dragonfly/dragonfly/entity/damage"
	"github.com/df-mc/dragonfly/dragonfly/world"
	"image/color"
	"time"
)

// Wither is a lasting effect that causes an entity to take continuous damage that is capable of killing an
// entity.
type Wither struct {
	lastingEffect
}

// Apply ...
func (w Wither) Apply(e world.Entity) {
	interval := 80 >> w.Lvl
	if tickDuration(w.Dur)%interval == 0 {
		if living, ok := e.(living); ok {
			living.Hurt(1, damage.SourceWitherEffect{})
		}
	}
}

// WithSettings ...
func (w Wither) WithSettings(d time.Duration, level int, ambient bool) Effect {
	return Wither{w.withSettings(d, level, ambient)}
}

// RGBA ...
func (Wither) RGBA() color.RGBA {
	return color.RGBA{R: 0x35, G: 0x2a, B: 0x27, A: 0xff}
}
