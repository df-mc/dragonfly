package world

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world/chunk"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/google/uuid"
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
	// pitch, which, when applied to the respective values of the entity, will result in the final values.
	ViewEntityMovement(e Entity, pos mgl64.Vec3, yaw, pitch float64, onGround bool)
	// ViewEntityVelocity views the velocity of an entity. It is called right before a call to
	// ViewEntityMovement so that the Viewer may interpolate the movement itself.
	ViewEntityVelocity(e Entity, vel mgl64.Vec3)
	// ViewEntityTeleport views the teleportation of an entity. The entity is immediately moved to a different
	// target position.
	ViewEntityTeleport(e Entity, pos mgl64.Vec3)
	// ViewChunk views the chunk passed at a particular position. It is called for every chunk loaded using
	// the world.Loader.
	ViewChunk(pos ChunkPos, c *chunk.Chunk, blockEntities map[cube.Pos]Block)
	// ViewTime views the time of the world. It is called every time the time is changed or otherwise every
	// second.
	ViewTime(t int)
	// ViewEntityItems views the items currently held by an entity that is able to equip items.
	ViewEntityItems(e Entity)
	// ViewEntityArmour views the items currently equipped as armour by the entity.
	ViewEntityArmour(e Entity)
	// ViewEntityAction views an action performed by an entity. Available actions may be found in the `action`
	// package, and include things such as swinging an arm.
	ViewEntityAction(e Entity, a EntityAction)
	// ViewEntityState views the current state of an entity. It is called whenever an entity changes its
	// physical appearance, for example when sprinting.
	ViewEntityState(e Entity)
	// ViewParticle views a particle spawned at a given position in the world. It is called when a particle,
	// for example a block breaking particle, is spawned near the player.
	ViewParticle(pos mgl64.Vec3, p Particle)
	// ViewSound is called when a sound is played in the world.
	ViewSound(pos mgl64.Vec3, s Sound)
	// ViewBlockUpdate views the updating of a block. It is called when a block is set at the position passed
	// to the method.
	ViewBlockUpdate(pos cube.Pos, b Block, layer int)
	// ViewBlockAction views an action performed by a block. Available actions may be found in the `action`
	// package, and include things such as a chest opening.
	ViewBlockAction(pos cube.Pos, a BlockAction)
	// ViewEmote views an emote being performed by another entity.
	ViewEmote(e Entity, emote uuid.UUID)
	// ViewSkin views the current skin of a player.
	ViewSkin(e Entity)
	// ViewWorldSpawn views the current spawn location of the world.
	ViewWorldSpawn(pos cube.Pos)
	// ViewWeather views the weather of the world, including rain and thunder.
	ViewWeather(raining, thunder bool)
}

// NopViewer is a Viewer implementation that does not implement any behaviour. It may be embedded by other structs to
// prevent having to implement all of Viewer's methods.
type NopViewer struct{}

// Compile time check to make sure NopViewer implements Viewer.
var _ Viewer = NopViewer{}

func (NopViewer) ViewEntity(Entity)                                             {}
func (NopViewer) HideEntity(Entity)                                             {}
func (NopViewer) ViewEntityMovement(Entity, mgl64.Vec3, float64, float64, bool) {}
func (NopViewer) ViewEntityVelocity(Entity, mgl64.Vec3)                         {}
func (NopViewer) ViewEntityTeleport(Entity, mgl64.Vec3)                         {}
func (NopViewer) ViewChunk(ChunkPos, *chunk.Chunk, map[cube.Pos]Block)          {}
func (NopViewer) ViewTime(int)                                                  {}
func (NopViewer) ViewEntityItems(Entity)                                        {}
func (NopViewer) ViewEntityArmour(Entity)                                       {}
func (NopViewer) ViewEntityAction(Entity, EntityAction)                         {}
func (NopViewer) ViewEntityState(Entity)                                        {}
func (NopViewer) ViewParticle(mgl64.Vec3, Particle)                             {}
func (NopViewer) ViewSound(mgl64.Vec3, Sound)                                   {}
func (NopViewer) ViewBlockUpdate(cube.Pos, Block, int)                          {}
func (NopViewer) ViewBlockAction(cube.Pos, BlockAction)                         {}
func (NopViewer) ViewEmote(Entity, uuid.UUID)                                   {}
func (NopViewer) ViewSkin(Entity)                                               {}
func (NopViewer) ViewWorldSpawn(cube.Pos)                                       {}
func (NopViewer) ViewWeather(bool, bool)                                        {}
