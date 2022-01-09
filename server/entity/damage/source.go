package damage

import "github.com/df-mc/dragonfly/server/world"

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

// SourceInstantDamageEffect is used for damage caused by an effect.InstantDamage applied to an entity.
type SourceInstantDamageEffect struct{}

// SourceVoid is used for damage caused by an entity being in the void.
type SourceVoid struct{}

// SourcePoisonEffect is used for damage caused by an effect.Poison or effect.FatalPoison applied to an
// entity.
type SourcePoisonEffect struct {
	// Fatal specifies if the damage was caused by effect.FatalPoison or not.
	Fatal bool
}

// SourceWitherEffect is used for damage caused by an effect.Wither applied to an entity.
type SourceWitherEffect struct{}

// SourceFire is used for damage caused by being in fire.
type SourceFire struct{}

// SourceFireTick is used for damage caused by being on fire.
type SourceFireTick struct{}

// SourceLava is used for damage caused by being in lava.
type SourceLava struct{}

// SourceFall is used for damage caused by falling.
type SourceFall struct{}

// SourceLightning is used for damage caused by being struck by lightning.
type SourceLightning struct{}

// SourceProjectile is used for damage caused by a projectile.
type SourceProjectile struct {
	// Projectile and Owner are the world.Entity that dealt the damage and the one that fired the projectile
	// respectively.
	Projectile, Owner world.Entity
}

// SourceCustom is a cause used for dealing any kind of custom damage. Armour reduces damage to this source,
// but otherwise no enchantments have an additional effect.
type SourceCustom struct{}

// ReducedByArmour ...
func (SourceFall) ReducedByArmour() bool {
	return false
}

// ReducedByArmour ...
func (SourceLightning) ReducedByArmour() bool {
	return true
}

// ReducedByArmour ...
func (SourceEntityAttack) ReducedByArmour() bool {
	return true
}

// ReducedByArmour ...
func (SourceStarvation) ReducedByArmour() bool {
	return false
}

// ReducedByArmour ...
func (SourceInstantDamageEffect) ReducedByArmour() bool {
	return false
}

// ReducedByArmour ...
func (SourceCustom) ReducedByArmour() bool {
	return false
}

// ReducedByArmour ...
func (SourceVoid) ReducedByArmour() bool {
	return false
}

// ReducedByArmour ...
func (SourcePoisonEffect) ReducedByArmour() bool {
	return false
}

// ReducedByArmour ...
func (SourceWitherEffect) ReducedByArmour() bool {
	return false
}

// ReducedByArmour ...
func (SourceFire) ReducedByArmour() bool {
	return true
}

// ReducedByArmour ...
func (SourceFireTick) ReducedByArmour() bool {
	return false
}

// ReducedByArmour ...
func (SourceLava) ReducedByArmour() bool {
	return true
}

// ReducedByArmour ...
func (SourceProjectile) ReducedByArmour() bool {
	return true
}
