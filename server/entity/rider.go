package entity

import (
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl32"
)

// Rider is an interface for entities that can ride other entities.
type Rider interface {
	// SeatPosition returns the Rider's current seat position.
	SeatPosition() mgl32.Vec3
	// RideEntity links the Rider to an entity if the entity is Rideable and if there is a seat available.
	RideEntity(e world.Entity)
	// DismountEntity unlinks the Rider from an entity.
	DismountEntity(e world.Entity)
	// CheckSeats moves a Rider to the seat corresponding to their current index within the slice of riders.
	// It is called whenever a Rider dismounts an entity.
	CheckSeats(e world.Entity)
	// Seat returns the index of a Rider within the slice of riders.
	Seat(e world.Entity) int
	// Riding returns the entity that the player is currently riding.
	Riding() world.Entity
	// SetRiding saves the entity the Rider is currently riding.
	SetRiding(e world.Entity)
}
