package entity

import (
	"github.com/df-mc/dragonfly/server/world"
	"time"
)

// SwingArmAction is a world.EntityAction that makes an entity or player swing its arm.
type SwingArmAction struct{ action }

// HurtAction is a world.EntityAction that makes an entity display the animation for being hurt. The entity will be
// shown as red for a short duration.
type HurtAction struct{ action }

// CriticalHitAction is a world.EntityAction that makes an entity display critical hit particles. This will show stars
// around the entity.
type CriticalHitAction struct{ action }

// DeathAction is a world.EntityAction that makes an entity display the death animation. After this animation, the
// entity disappears from viewers watching it.
type DeathAction struct{ action }

// EatAction is a world.EntityAction that makes an entity display the eating particles at its mouth to viewers with the
// item in its hand being eaten.
type EatAction struct{ action }

// ArrowShakeAction makes an arrow entity display a shaking animation for the given duration.
type ArrowShakeAction struct {
	// Duration is the duration of the shake.
	Duration time.Duration

	action
}

// PickedUpAction is a world.EntityAction that makes an item get picked up by a collector. After this animation, the
// item disappears from viewers watching it.
type PickedUpAction struct {
	// Collector is the entity that collected the item.
	Collector world.Entity

	action
}

// action implements the Action interface. Structures in this package may embed it to gets its functionality
// out of the box.
type action struct{}

func (action) EntityAction() {}
