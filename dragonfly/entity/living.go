package entity

import (
	"github.com/df-mc/dragonfly/dragonfly/entity/damage"
	"github.com/df-mc/dragonfly/dragonfly/entity/healing"
	"github.com/df-mc/dragonfly/dragonfly/world"
	"github.com/go-gl/mathgl/mgl64"
)

// Living represents an entity that is alive and that has health. It is able to take damage and will die upon
// taking fatal damage.
type Living interface {
	world.Entity
	// Health returns the health of the entity.
	Health() float64
	// AttackImmune checks if the entity is currently immune to entity attacks. Entities typically turn
	// immune for half a second after being attacked.
	AttackImmune() bool
	// Hurt hurts the entity for a given amount of damage. The source passed represents the cause of the
	// damage, for example damage.SourceEntityAttack if the entity is attacked by another entity.
	// If the final damage exceeds the health that the player currently has, the entity is killed.
	Hurt(damage float64, source damage.Source)
	// Heal heals the entity for a given amount of health. The source passed represents the cause of the
	// healing, for example healing.SourceFood if the entity healed by having a full food bar. If the health
	// added to the original health exceeds the entity's max health, Heal may not add the full amount.
	Heal(health float64, source healing.Source)
	// KnockBack knocks the entity back with a given force and height. A source is passed which indicates the
	// source of the velocity, typically the position of an attacking entity. The source is used to calculate
	// the direction which the entity should be knocked back in.
	KnockBack(src mgl64.Vec3, force, height float64)
	// Speed returns the current speed of the living entity. The default value is different for each entity.
	Speed() float64
	// SetSpeed sets the speed of an entity to a new value.
	SetSpeed(float64)
}
