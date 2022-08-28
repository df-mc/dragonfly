package entity

import (
	"github.com/df-mc/dragonfly/server/entity/effect"
	"github.com/go-gl/mathgl/mgl64"
)

// Living represents an entity that is alive and that has health. It is able to take damage and will die upon
// taking fatal damage.
type Living interface {
	Hurtable
	// KnockBack knocks the entity back with a given force and height. A source is passed which indicates the
	// source of the velocity, typically the position of an attacking entity. The source is used to calculate
	// the direction which the entity should be knocked back in.
	KnockBack(src mgl64.Vec3, force, height float64)
	// Velocity returns the players current velocity.
	Velocity() mgl64.Vec3
	// SetVelocity updates the entity's velocity.
	SetVelocity(velocity mgl64.Vec3)
	// AddEffect adds an entity.Effect to the entity. If the effect is instant, it is applied to the entity
	// immediately. If not, the effect is applied to the entity every time the Tick method is called.
	// AddEffect will overwrite any effects present if the level of the effect is higher than the existing one, or
	// if the effects' levels are equal and the new effect has a longer duration.
	AddEffect(e effect.Effect)
	// RemoveEffect removes any effect that might currently be active on the entity.
	RemoveEffect(e effect.Type)
	// Effects returns any effect currently applied to the entity. The returned effects are guaranteed not to have
	// expired when returned.
	Effects() []effect.Effect
	// Speed returns the current speed of the living entity. The default value is different for each entity.
	Speed() float64
	// SetSpeed sets the speed of an entity to a new value.
	SetSpeed(float64)
}
