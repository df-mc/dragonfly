package entity

// Flammable is an interface for entities that can be set on fire.
type Flammable interface {
	// FireProof is whether the entity is currently fire proof.
	FireProof() bool
	// FireDamage deals fire damage to the entity.
	FireDamage(amount float64)
	// LavaDamage deals lava damage to the entity.
	LavaDamage(amount float64)
	// FireTicks returns duration of fire in ticks.
	FireTicks() int
	// SetFireTicks sets the player on fire for the specified duration in ticks.
	SetFireTicks(ticks int)
}
