package effect

import (
	"github.com/df-mc/dragonfly/dragonfly/entity"
	"github.com/df-mc/dragonfly/dragonfly/entity/damage"
	"image/color"
	"time"
)

// Absorption is a lasting effect that increases the health of an entity over the maximum. Once this extra
// health is lost, it will not regenerate.
type Absorption struct {
	lastingEffect
}

// Absorbs checks if Absorption absorbs the damage source passed.
func (a Absorption) Absorbs(src damage.Source) bool {
	switch src.(type) {
	case damage.SourceWitherEffect, damage.SourceInstantDamageEffect, damage.SourcePoisonEffect, damage.SourceStarvation:
		return true
	}
	return false
}

// Start ...
func (a Absorption) Start(e entity.Living) {
	if i, ok := e.(interface {
		SetAbsorption(health float64)
	}); ok {
		i.SetAbsorption(4 * float64(a.Lvl))
	}
}

// Stop ...
func (a Absorption) Stop(e entity.Living) {
	if i, ok := e.(interface {
		SetAbsorption(health float64)
	}); ok {
		i.SetAbsorption(0)
	}
}

// WithDuration ...
func (a Absorption) WithDuration(d time.Duration) entity.Effect {
	return Absorption{a.withDuration(d)}
}

// RGBA ...
func (a Absorption) RGBA() color.RGBA {
	return color.RGBA{R: 0x25, G: 0x52, B: 0xa5, A: 0xff}
}
