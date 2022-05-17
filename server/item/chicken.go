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
func (c Chicken) Consume(_ *world.World, co Consumer) Stack {
	if c.Cooked {
		co.Saturate(6, 7.2)
	} else {
		co.Saturate(2, 1.2)
		if rand.Float64() < 0.3 {
			co.AddEffect(effect.New(effect.Hunger{}, 1, 30*time.Second))
		}
	}
	return Stack{}
}

// SmeltInfo ...
func (c Chicken) SmeltInfo() SmeltInfo {
	if c.Cooked {
		return SmeltInfo{}
	}
	return SmeltInfo{
		Product:    NewStack(Chicken{Cooked: true}, 1),
		Experience: 0.35,
		Food:       true,
	}
}

// EncodeItem ...
func (c Chicken) EncodeItem() (name string, meta int16) {
	if c.Cooked {
		return "minecraft:cooked_chicken", 0
	}
	return "minecraft:chicken", 0
}
