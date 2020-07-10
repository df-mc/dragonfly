package effect

import (
	"github.com/df-mc/dragonfly/dragonfly/entity"
	"github.com/df-mc/dragonfly/dragonfly/entity/healing"
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

// Apply instantly heals the entity.Living passed for a bit of health, depending on the effect level and
// potency.
func (i InstantHealth) Apply(e entity.Living) {
	if i.Potency == 0 {
		// Potency of 1 by default.
		i.Potency = 1
	}
	base := 2 << i.Lvl
	e.Heal(float64(base)*i.Potency, healing.SourceInstantHealthEffect{})
}

// WithDuration ...
func (i InstantHealth) WithDuration(d time.Duration) entity.Effect {
	return i
}
