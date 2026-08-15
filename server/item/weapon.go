package item

// KineticWeapon represents a weapon that deals damage based on the kinetic energy of the user's movement,
// such as a spear.
type KineticWeapon interface {
	// KineticWeaponInfo returns the kinetic weapon information of the item.
	KineticWeaponInfo() KineticWeaponInfo
}

// KineticWeaponInfo is a struct returned by items that implement KineticWeapon. It contains the information
// required for the client to handle the item as a kinetic weapon.
type KineticWeaponInfo struct {
	// Reach is the minimum and maximum reach of the weapon.
	Reach [2]float64
	// CreativeReach is the minimum and maximum reach of the weapon in creative mode.
	CreativeReach [2]float64
	// HitboxMargin is the margin of the hitbox of the weapon.
	HitboxMargin float64
	// DamageMultiplier is the multiplier applied to the damage dealt by the weapon.
	DamageMultiplier float64
	// DamageModifier is the modifier applied to the damage dealt by the weapon.
	DamageModifier float64
	// Delay is the delay in ticks before the weapon can be used again.
	Delay int
	// DamageConditions are the conditions required for the weapon to deal its full damage.
	DamageConditions WeaponConditions
	// KnockbackConditions are the conditions required for the weapon to knock back its target.
	KnockbackConditions WeaponConditions
	// DismountConditions are the conditions required for the weapon to dismount its target.
	DismountConditions WeaponConditions
}

// WeaponConditions is a struct returned by items that implement KineticWeapon. It contains the conditions
// required for the weapon to apply certain effects.
type WeaponConditions struct {
	// MaxDuration is the maximum duration in ticks for which the condition applies.
	MaxDuration int
	// MinSpeed is the minimum speed of the user for which the condition applies.
	MinSpeed float64
	// MinRelativeSpeed is the minimum relative speed of the user for which the condition applies.
	MinRelativeSpeed float64
}

// PiercingWeapon represents a weapon that can pierce through targets, such as a spear.
type PiercingWeapon interface {
	// PiercingWeaponInfo returns the piercing weapon information of the item.
	PiercingWeaponInfo() PiercingWeaponInfo
}

// PiercingWeaponInfo is a struct returned by items that implement PiercingWeapon. It contains the information
// required for the client to handle the item as a piercing weapon.
type PiercingWeaponInfo struct {
	// Reach is the minimum and maximum reach of the weapon.
	Reach [2]float64
	// CreativeReach is the minimum and maximum reach of the weapon in creative mode.
	CreativeReach [2]float64
	// HitboxMargin is the margin of the hitbox of the weapon.
	HitboxMargin float64
}
