package entity

import (
	"github.com/df-mc/dragonfly/server/entity/damage"
	"github.com/df-mc/dragonfly/server/entity/healing"
	"github.com/df-mc/dragonfly/server/world"
)

// Hurtable represents an entity that can be hurt, such as a Living entity or an Item entity.
type Hurtable interface {
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
}
