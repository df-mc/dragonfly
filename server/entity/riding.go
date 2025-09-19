package entity

import (
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl32"
)

// Rider is an interface for entities that can ride other entities.
type Rider interface {
	world.Entity
	// RidingEntity returns the entity the player is currently riding.
	RidingEntity() Rideable
	// SeatPosition returns the position of where the player is sitting.
	SeatPosition() mgl32.Vec3
	// MountEntity mounts the Rider to an entity if the entity is Rideable and if there is a seat available.
	MountEntity(e Rideable, position mgl32.Vec3, driver bool)
	// DismountEntity dismounts the rider from the entity they are currently riding.
	DismountEntity()
}

// Rideable is an interface for entities that can be ridden.
type Rideable interface {
	world.Entity
	Driver() Rider
	Move(vector mgl32.Vec2, yaw, pitch float32)
}
