package world

import (
	"github.com/dragonfly-tech/dragonfly/dragonfly/world/chunk"
	"github.com/go-gl/mathgl/mgl32"
)

// Viewer is a viewer in the world. It can view changes that are made in the world, such as the addition of
// entities and the changes of blocks.
type Viewer interface {
	// ViewEntity views the entity passed. It is called for every entity that the viewer may encounter in the
	// world, either by moving entities or by moving the viewer using a world.Loader.
	ViewEntity(e Entity)
	// HideEntity stops viewing the entity passed. It is called for every entity that leaves the viewing range
	// of the viewer, either by its movement or the movement of the viewer using a world.Loader.
	HideEntity(e Entity)
	// ViewEntityMovement views the movement of an entity. The entity is moved with a delta position, yaw and
	// pitch, which, when applied to values of the entity, will result in the final values.
	ViewEntityMovement(e Entity, deltaPos mgl32.Vec3, deltaYaw, deltaPitch float32)
	// ViewEntityTeleport views the teleportation of an entity. The entity is immediately moved to a different
	// target position.
	ViewEntityTeleport(e Entity, position mgl32.Vec3)
	// ViewChunk views the chunk passed at a particular position. It is called for every chunk loaded using
	// the world.Loader.
	ViewChunk(pos ChunkPos, c *chunk.Chunk)
	// ViewTime views the time of the world. It is called every time the time is changed or otherwise every
	// second.
	ViewTime(time int)
}
