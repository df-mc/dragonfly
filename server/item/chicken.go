package item

import (
	"github.com/df-mc/dragonfly/server/entity/effect"
	"github.com/df-mc/dragonfly/server/world"
	"math/rand"
	"time"
)

// Chicken is a food item obtained from chickens. It can be cooked in a furnace, smoker, or campfire.
type Chicken struct {
	defaultFood

	// Cooked is whether the chicken is cooked.
	Cooked bool
}

// Consume ...
func (ch Chicken) Consume(_ *world.World, c Consumer) Stack {
	if ch.Cooked {
		c.Saturate(6, 7.2)
	} else {
		c.Saturate(2, 1.2)
		if rand.Float64() < 0.3 {
			c.AddEffect(effect.New(effect.Hunger{}, 1, 30*time.Second))
		}
	}
	return Stack{}
}

// EncodeItem ...
func (ch Chicken) EncodeItem() (name string, meta int16) {
	if ch.Cooked {
		return "minecraft:cooked_chicken", 0
	}
	return "minecraft:chicken", 0
}
