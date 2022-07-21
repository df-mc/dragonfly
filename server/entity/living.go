package entity

import (
	"github.com/df-mc/dragonfly/server/entity/damage"
	"github.com/df-mc/dragonfly/server/entity/effect"
	"github.com/df-mc/dragonfly/server/entity/healing"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// Living represents an entity that is alive and that has health. It is able to take damage and will die upon
// taking fatal damage.
type Living interface {
	world.Entity
	// Health returns the health of the entity.
	Health() float64
	// MaxHealth returns the maximum health of the entity.
	MaxHealth() float64
	// SetMaxHealth changes the maximum health of the entity to the value passed.
	SetMaxHealth(v float64)
	// Dead checks if the entity is considered dead. True is returned if the health of the entity is equal to or
	// lower than 0.
	Dead() bool
	// AttackImmune checks if the entity is currently immune to entity attacks. Entities typically turn
	// immune for half a second after being attacked.
	AttackImmune() bool
	// Hurt hurts the entity for a given amount of damage. The source passed represents the cause of the
	// damage, for example damage.SourceEntityAttack if the entity is attacked by another entity.
	// If the final damage exceeds the health that the entity currently has, the entity is killed.
	// Hurt returns the final amount of damage dealt to the Living entity and returns whether the Living entity
	// was vulnerable to the damage at all.
	Hurt(damage float64, src damage.Source) (n float64, vulnerable bool)
	// Heal heals the entity for a given amount of health. The source passed represents the cause of the
	// healing, for example healing.SourceFood if the entity healed by having a full food bar. If the health
	// added to the original health exceeds the entity's max health, Heal may not add the full amount.
	Heal(health float64, src healing.Source)
	// KnockBack knocks the entity back with a given force and height. A source is passed which indicates the
	// source of the velocity, typically the position of an attacking entity. The source is used to calculate
	// the direction which the entity should be knocked back in.
	KnockBack(src mgl64.Vec3, force, height float64)
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
