package world

import (
	"git.jetbrains.space/dragonfly/dragonfly/dragonfly/block"
	"git.jetbrains.space/dragonfly/dragonfly/dragonfly/entity/action"
	"git.jetbrains.space/dragonfly/dragonfly/dragonfly/entity/state"
	"git.jetbrains.space/dragonfly/dragonfly/dragonfly/world/chunk"
	"git.jetbrains.space/dragonfly/dragonfly/dragonfly/world/particle"
	"git.jetbrains.space/dragonfly/dragonfly/dragonfly/world/sound"
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
	// ViewEntityItems views the items currently held by an entity that is able to equip items.
	ViewEntityItems(e CarryingEntity)
	// ViewEntityAction views an action performed by an entity. Available actions may be found in the `action`
	// package, and include things such as swinging an arm.
	ViewEntityAction(e Entity, a action.Action)
	// ViewEntityState views the current state of an entity. It is called whenever an entity changes its
	// physical appearance, for example when sprinting.
	ViewEntityState(e Entity, s []state.State)
	// ViewParticle views a particle spawned at a given position in the world. It is called when a particle,
	// for example a block breaking particle, is spawned near the player.
	ViewParticle(pos mgl32.Vec3, p particle.Particle)
	// ViewSound is called when a sound is played in the world.
	ViewSound(pos mgl32.Vec3, s sound.Sound)
	// ViewBlockUpdate views the updating of a block. It is called when a block is set at the position passed
	// to the method.
	ViewBlockUpdate(pos block.Position, b block.Block)
}
