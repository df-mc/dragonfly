package effect

import (
	"github.com/df-mc/dragonfly/server/entity/damage"
	"github.com/df-mc/dragonfly/server/world"
	"image/color"
)

// Absorption is a lasting effect that increases the health of an entity over the maximum. Once this extra
// health is lost, it will not regenerate.
type Absorption struct {
	nopLasting
}

// Absorbs checks if Absorption absorbs the damage source passed.
func (Absorption) Absorbs(src damage.Source) bool {
	switch src.(type) {
	case damage.SourceWitherEffect, damage.SourceInstantDamageEffect, damage.SourcePoisonEffect, damage.SourceStarvation:
		return true
	}
	return false
}

// Start ...
func (Absorption) Start(e world.Entity, lvl int) {
	if i, ok := e.(interface {
		SetAbsorption(health float64)
	}); ok {
		i.SetAbsorption(4 * float64(lvl))
	}
}

// End ...
func (Absorption) End(e world.Entity, _ int) {
	if i, ok := e.(interface {
		SetAbsorption(health float64)
	}); ok {
		i.SetAbsorption(0)
	}
}

// RGBA ...
func (Absorption) RGBA() color.RGBA {
	return color.RGBA{R: 0x25, G: 0x52, B: 0xa5, A: 0xff}
}
