package effect

import (
	"github.com/df-mc/dragonfly/dragonfly/entity/healing"
	"github.com/df-mc/dragonfly/dragonfly/world"
	"time"
)

// InstantHealth is an instant effect that causes the player that it is applied to to immediately regain some
// health. The amount of health regained depends on the effect level and potency.
type InstantHealth struct {
	instantEffect
	// Potency specifies the potency of the instant health. By default this value is 1, which means 100% of
	// the instant health will be applied to an entity. A lingering health potion, for example, has a potency
	// of 0.5: It heals 1 heart (per tick) instead of 2.
	Potency float64
}

// Apply instantly heals the world.Entity passed for a bit of health, depending on the effect level and
// potency.
func (i InstantHealth) Apply(e world.Entity) {
	if i.Potency == 0 {
		// Potency of 1 by default.
		i.Potency = 1
	}
	base := 2 << i.Lvl
	if living, ok := e.(living); ok {
		living.Heal(float64(base)*i.Potency, healing.SourceInstantHealthEffect{})
	}
}

// WithSettings ...
func (i InstantHealth) WithSettings(_ time.Duration, level int, _ bool) Effect {
	i.Lvl = level
	return i
}
