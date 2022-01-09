package entity

import "time"

// Flammable is an interface for entities that can be set on fire.
type Flammable interface {
	// FireProof is whether the entity is currently fireproof.
	FireProof() bool
	// OnFireDuration returns duration of fire in ticks.
	OnFireDuration() time.Duration
	// SetOnFire sets the entity on fire for the specified duration.
	SetOnFire(duration time.Duration)
	// Extinguish extinguishes the entity.
	Extinguish()
}
