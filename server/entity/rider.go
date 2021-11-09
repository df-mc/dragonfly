package entity

import (
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl32"
)

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
	// GetSeat returns the index of a Rider within the slice of riders.
	GetSeat(e world.Entity) int
	// Riding returns the runtime ID of the entity the Rider is riding.
	Riding() uint64
	// SetRiding saves the runtime ID of the entity the Rider is riding.
	SetRiding(id uint64)
}
