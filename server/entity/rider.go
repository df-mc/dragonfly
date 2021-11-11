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
	// RidingEntity returns the entity the player is currently riding and the player's seat index.
	RidingEntity() (world.Entity, int)
}
