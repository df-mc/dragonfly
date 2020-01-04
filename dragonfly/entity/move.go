package entity

import (
	"github.com/dragonfly-tech/dragonfly/dragonfly/world"
	"github.com/go-gl/mathgl/mgl32"
)

// Move moves an entity to a different position in the world that it is in. The entity is moved by the delta
// position passed, meaning it will move deltaPosition blocks from its current position.
func Move(e world.Entity, deltaPosition mgl32.Vec3) {
	e.World().MoveEntity(e, deltaPosition)
}

// Rotate rotates an entity to a different rotation. The entity is rotated by the delta yaw and pitch passed,
// meaning it will change relative to its current rotation.
func Rotate(e world.Entity, deltaYaw, deltaPitch float32) {
	e.World().RotateEntity(e, deltaYaw, deltaPitch)
}

// Teleport teleports an entity to a target position. The entity is immediately moved to the new position.
func Teleport(e world.Entity, position mgl32.Vec3) {
	e.World().TeleportEntity(e, position)
}
