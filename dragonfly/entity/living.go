package entity

import (
	"git.jetbrains.space/dragonfly/dragonfly.git/dragonfly/entity/damage"
	"github.com/go-gl/mathgl/mgl32"
)

// Living represents an entity that is alive and that has health. It is able to take damage and will die upon
// taking fatal damage.
type Living interface {
	// AttackImmune checks if the entity is currently immune to entity attacks. Entities typically turn
	// immune for half a second after being attacked.
	AttackImmune() bool
	// Hurt hurts the entity for a given amount of damage. The source passed represents the cause of the
	// damage, for example damage.SourceEntityAttack if the entity is attacked by another entity.
	// If the final damage exceeds the health that the player currently has, the entity is killed.
	Hurt(damage float32, source damage.Source)
	// KnockBack knocks the entity back with a given force and height. A source is passed which indicates the
	// source of the knockback, typically the position of an attacking entity. The source is used to calculate
	// the direction which the entity should be knocked back in.
	KnockBack(src mgl32.Vec3, force, height float32)
}
