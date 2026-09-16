package entity

import (
	"time"

	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
)

// Flammable is an interface for entities that can be set on fire.
type Flammable interface {
	// OnFireDuration returns duration of fire in ticks.
	OnFireDuration() time.Duration
	// SetOnFire sets the entity on fire for the specified duration.
	SetOnFire(duration time.Duration)
	// Extinguish extinguishes the entity.
	Extinguish()
}

// TickOnFire progresses the burning of an entity by one tick: rain extinguishes it and every full second
// FlammableEntity is a Hurtable that can catch fire, which is what the shared
// fire tick works on.
type FlammableEntity interface {
	Hurtable
	Flammable
}

// of remaining fire time deals a point of fire damage.
func TickOnFire(e FlammableEntity, tx *world.Tx) {
	if e.OnFireDuration() <= 0 {
		return
	}
	if tx.RainingAt(cube.PosFromVec3(e.Position())) {
		e.Extinguish()
		return
	}
	if e.OnFireDuration()%time.Second == 0 {
		e.Hurt(1, block.FireDamageSource{})
	}
}
