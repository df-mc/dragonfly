package damage

import "github.com/df-mc/dragonfly/dragonfly/world"

// Source represents the source of the damage dealt to an entity. This source may be passed to the Hurt()
// method of an entity in order to deal damage to an entity with a specific source.
type Source interface {
	// ReducedByArmour checks if the source of damage may be reduced if the receiver of the damage is wearing
	// armour.
	ReducedByArmour() bool
}

// SourceEntityAttack is used for damage caused by other entities, for example when a player attacks another
// player.
type SourceEntityAttack struct {
	// Attacker holds the attacking entity. The entity may be a player or any other entity.
	Attacker world.Entity
}

// SourceStarvation is used for damage caused by a completely depleted food bar.
type SourceStarvation struct{}

// SourceCustom is a cause used for dealing any kind of custom damage. Armour reduces damage of this source,
// but otherwise no enchantments have an additional effect.
type SourceCustom struct{}

// ReducedByArmour ...
func (SourceEntityAttack) ReducedByArmour() bool {
	return true
}

// ReducedByArmour ...
func (SourceStarvation) ReducedByArmour() bool {
	return false
}

// ReducedByArmour ...
func (SourceCustom) ReducedByArmour() bool {
	return true
}
