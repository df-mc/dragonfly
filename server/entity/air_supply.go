package entity

import (
	"math/rand/v2"
	"time"

	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/enchantment"
)

// AirSupply keeps track of the air an entity has left while submerged, and whether it is currently able to
// breathe. It implements the vanilla drowning cadence: air depletes by a tick's worth every tick without
// air, a drowning hit lands once the supply is a second past empty, and air replenishes at five times the
// depletion rate.
type AirSupply struct {
	supply, max time.Duration
	breathing   bool
}

// NewAirSupply returns an AirSupply with the maximum air passed, full and breathing. A zero maximum uses
// the vanilla default of 15 seconds.
func NewAirSupply(max time.Duration) *AirSupply {
	if max <= 0 {
		max = time.Second * 15
	}
	return &AirSupply{supply: max, max: max, breathing: true}
}

// Supply returns the air the entity has left.
func (a *AirSupply) Supply() time.Duration {
	return a.supply
}

// SetSupply sets the air the entity has left.
func (a *AirSupply) SetSupply(d time.Duration) {
	a.supply = d
}

// Max returns the maximum air supply of the entity.
func (a *AirSupply) Max() time.Duration {
	return a.max
}

// SetMax sets the maximum air supply of the entity.
func (a *AirSupply) SetMax(d time.Duration) {
	a.max = d
}

// Breathing checks if the entity is currently breathing.
func (a *AirSupply) Breathing() bool {
	return a.breathing
}

// Tick progresses the air supply by one tick. The helmet passed may reduce air loss through its Respiration
// enchantment. Tick returns whether the entity takes a drowning hit this tick and whether the state changed
// in a way viewers should see.
func (a *AirSupply) Tick(canBreathe bool, helmet item.Stack) (drowned, changed bool) {
	if !canBreathe {
		if r, ok := helmet.Enchantment(enchantment.Respiration); ok && rand.Float64() < enchantment.Respiration.Chance(r.Level()) {
			return false, false
		}
		if a.supply -= time.Second / 20; a.supply <= -time.Second {
			a.supply = 0
			drowned = true
		}
		a.breathing = false
		return drowned, true
	}
	if !a.breathing && a.supply < a.max {
		a.supply = min(a.supply+time.Second/4, a.max)
		a.breathing = a.supply == a.max
		return false, true
	}
	return false, false
}
