package effect

import (
	"github.com/df-mc/dragonfly/dragonfly/entity/damage"
	"github.com/df-mc/dragonfly/dragonfly/world"
	"time"
)

// InstantDamage is an instant effect that causes a living entity to immediately take some damage, depending
// on the level and the potency of the effect.
type InstantDamage struct {
	instantEffect
	// Potency specifies the potency of the instant damage. By default this value is 1, which means 100% of
	// the instant damage will be applied to an entity. A lingering damage potion, for example, has a potency
	// of 0.5: It deals 1.5 hearts damage (per tick) instead of 3.
	Potency float64
}

// Apply ...
func (i InstantDamage) Apply(e world.Entity) {
	if i.Potency == 0 {
		// Potency of 1 by default.
		i.Potency = 1
	}
	base := 3 << i.Lvl
	if living, ok := e.(living); ok {
		living.Hurt(float64(base)*i.Potency, damage.SourceInstantDamageEffect{})
	}
}

// WithSettings ...
func (i InstantDamage) WithSettings(_ time.Duration, level int, _ bool) Effect {
	i.Lvl = level
	return i
}
