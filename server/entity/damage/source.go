package damage

import "github.com/df-mc/dragonfly/server/world"

type (
	// Source represents the source of the damage dealt to an entity. This source may be passed to the Hurt()
	// method of an entity in order to deal damage to an entity with a specific source.
	Source interface {
		// ReducedByArmour checks if the source of damage may be reduced if the receiver of the damage is wearing
		// armour.
		ReducedByArmour() bool
		// ReducedByResistance specifies if the Source is affected by the resistance effect. If false, damage dealt
		// to an entity with this source will not be lowered if the entity has the resistance effect.
		ReducedByResistance() bool
	}

	// SourceEntityAttack is used for damage caused by other entities, for example when a player attacks another
	// player.
	SourceEntityAttack struct {
		// Attacker holds the attacking entity. The entity may be a player or any other entity.
		Attacker world.Entity
	}

	// SourceStarvation is used for damage caused by a completely depleted food bar.
	SourceStarvation struct{}

	// SourceInstantDamageEffect is used for damage caused by an effect.InstantDamage applied to an entity.
	SourceInstantDamageEffect struct{}

	// SourceVoid is used for damage caused by an entity being in the void.
	SourceVoid struct{}

	// SourceSuffocation is used for damage caused by an entity suffocating in a block.
	SourceSuffocation struct{}

	// SourceDrowning is used for damage caused by an entity drowning in water.
	SourceDrowning struct{}

	// SourcePoisonEffect is used for damage caused by an effect.Poison or effect.FatalPoison applied to an
	// entity.
	SourcePoisonEffect struct {
		// Fatal specifies if the damage was caused by effect.FatalPoison or not.
		Fatal bool
	}

	// SourceWitherEffect is used for damage caused by an effect.Wither applied to an entity.
	SourceWitherEffect struct{}

	// SourceFire is used for damage caused by being in fire.
	SourceFire struct{}

	// SourceFireTick is used for damage caused by being on fire.
	SourceFireTick struct{}

	// SourceLava is used for damage caused by being in lava.
	SourceLava struct{}

	// SourceFall is used for damage caused by falling.
	SourceFall struct{}

	// SourceLightning is used for damage caused by being struck by lightning.
	SourceLightning struct{}

	// SourceProjectile is used for damage caused by a projectile.
	SourceProjectile struct {
		// Projectile and Owner are the world.Entity that dealt the damage and the one that fired the projectile
		// respectively.
		Projectile, Owner world.Entity
	}

	// SourceBlock is used for damage caused by a block, such as an anvil.
	SourceBlock struct {
		// Block is the block that caused the damage.
		Block world.Block
	}
)

func (SourceFall) ReducedByArmour() bool                    { return false }
func (SourceFall) ReducedByResistance() bool                { return true }
func (SourceLightning) ReducedByArmour() bool               { return true }
func (SourceLightning) ReducedByResistance() bool           { return true }
func (SourceEntityAttack) ReducedByArmour() bool            { return true }
func (SourceEntityAttack) ReducedByResistance() bool        { return true }
func (SourceStarvation) ReducedByArmour() bool              { return false }
func (SourceStarvation) ReducedByResistance() bool          { return false }
func (SourceInstantDamageEffect) ReducedByArmour() bool     { return false }
func (SourceInstantDamageEffect) ReducedByResistance() bool { return true }
func (SourceVoid) ReducedByResistance() bool                { return false }
func (SourceVoid) ReducedByArmour() bool                    { return false }
func (SourceSuffocation) ReducedByResistance() bool         { return false }
func (SourceSuffocation) ReducedByArmour() bool             { return false }
func (SourceDrowning) ReducedByResistance() bool            { return false }
func (SourceDrowning) ReducedByArmour() bool                { return false }
func (SourcePoisonEffect) ReducedByResistance() bool        { return true }
func (SourcePoisonEffect) ReducedByArmour() bool            { return false }
func (SourceWitherEffect) ReducedByResistance() bool        { return true }
func (SourceWitherEffect) ReducedByArmour() bool            { return false }
func (SourceFire) ReducedByResistance() bool                { return true }
func (SourceFire) ReducedByArmour() bool                    { return true }
func (SourceFireTick) ReducedByResistance() bool            { return true }
func (SourceFireTick) ReducedByArmour() bool                { return false }
func (SourceLava) ReducedByResistance() bool                { return true }
func (SourceLava) ReducedByArmour() bool                    { return true }
func (SourceProjectile) ReducedByResistance() bool          { return true }
func (SourceProjectile) ReducedByArmour() bool              { return true }
func (SourceBlock) ReducedByResistance() bool               { return true }
func (SourceBlock) ReducedByArmour() bool                   { return true }
