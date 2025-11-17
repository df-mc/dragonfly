package entity

import (
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// Rider is an interface for entities that can ride other entities.
type Rider interface {
	world.Entity
	// RidingEntity returns the entity that the rider is currently sitting on.
	RidingEntity() Rideable
	// SeatIndex returns the position of where the rider is sitting.
	SeatIndex() int
	// MountEntity mounts the Rider to an entity if the entity is Rideable and if there is a seat available.
	MountEntity(rideable Rideable, seatIndex int)
	// DismountEntity dismounts the rider from the entity they are currently riding.
	DismountEntity()
}

// Rideable is an interface for entities that can be ridden.
type Rideable interface {
	world.Entity
	// SeatPositions returns a map of seat indices to their positions relative to the entity's position.
	SeatPositions() []mgl64.Vec3
	// NextFreeSeatIndex returns the index of the next free seat and whether a free seat was found.
	NextFreeSeatIndex(clickPos mgl64.Vec3) (int, bool)
	// ControllingRider returns the rider that is controlling the entity, if any.
	ControllingRider() Rider
	// MoveInput moves the entity based on input from the controlling rider.
	MoveInput(vector mgl64.Vec2, yaw, pitch float32)
}
