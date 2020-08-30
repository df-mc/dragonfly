package entity

import "time"

// Flammable is an interface for entities that can be set on fire.
type Flammable interface {
	// FireProof is whether the entity is currently fire proof.
	FireProof() bool
	// FireDamage deals fire damage to the entity.
	FireDamage(amount float64)
	// LavaDamage deals lava damage to the entity.
	LavaDamage(amount float64)
	// OnFireDuration returns duration of fire in ticks.
	OnFireDuration() time.Duration
	// SetOnFire sets the entity on fire for the specified duration.
	SetOnFire(duration time.Duration)
	// Extinguish extinguishes the entity.
	Extinguish()
}
