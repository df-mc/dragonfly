package world

import (
	"time"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world/chunk"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/google/uuid"
)

// Viewer is a viewer in the world. It can view changes that are made in the world, such as the addition of
// entities and the changes of blocks.
type Viewer interface {
	// ViewEntity views the Entity passed. It is called for every Entity that the viewer may encounter in the
	// world, either by moving entities or by moving the viewer using a world.Loader.
	ViewEntity(e Entity)
	// HideEntity stops viewing the Entity passed. It is called for every Entity that leaves the viewing range
	// of the viewer, either by its movement or the movement of the viewer using a world.Loader.
	HideEntity(e Entity)
	// ViewEntityGameMode views the game mode of the Entity passed. This is necessary for game-modes like spectator,
	// which may update how the Entity is viewed for others.
	ViewEntityGameMode(e Entity)
	// ViewEntityMovement views the movement of an Entity. The Entity is moved with a delta position, yaw and
	// pitch, which, when applied to the respective values of the Entity, will result in the final values.
	ViewEntityMovement(e Entity, pos mgl64.Vec3, rot cube.Rotation, onGround bool)
	// ViewEntityVelocity views the velocity of an Entity. It is called right before a call to
	// ViewEntityMovement so that the Viewer may interpolate the movement itself.
	ViewEntityVelocity(e Entity, vel mgl64.Vec3)
	// ViewEntityTeleport views the teleportation of an Entity. The Entity is immediately moved to a different
	// target position.
	ViewEntityTeleport(e Entity, pos mgl64.Vec3)
	// ViewFurnaceUpdate updates a furnace for the associated session based on previous times.
	ViewFurnaceUpdate(prevCookTime, cookTime, prevRemainingFuelTime, remainingFuelTime, prevMaxFuelTime, maxFuelTime time.Duration)
	// ViewBrewingUpdate updates a brewing stand for the associated session based on previous times.
	ViewBrewingUpdate(prevBrewTime, brewTime time.Duration, prevFuelAmount, fuelAmount, prevFuelTotal, fuelTotal int32)
	// ViewChunk views the chunk passed at a particular position. It is called for every chunk loaded using
	// the world.Loader.
	ViewChunk(pos ChunkPos, dim Dimension, blockEntities map[cube.Pos]Block, c *chunk.Chunk)
	// ViewTime views the time of the world. It is called every time the time is changed or otherwise every
	// second.
	ViewTime(t int)
	// ViewTimeCycle controls the automatic time-of-day cycle (day and night) in the world for this viewer.
	ViewTimeCycle(doDayLightCycle bool)
	// ViewEntityItems views the items currently held by an Entity that is able to equip items.
	ViewEntityItems(e Entity)
	// ViewEntityArmour views the items currently equipped as armour by the Entity.
	ViewEntityArmour(e Entity)
	// ViewEntityAction views an action performed by an Entity. Available actions may be found in the `action`
	// package, and include things such as swinging an arm.
	ViewEntityAction(e Entity, a EntityAction)
	// ViewEntityState views the current state of an Entity. It is called whenever an Entity changes its
	// physical appearance, for example when sprinting.
	ViewEntityState(e Entity)
	// ViewEntityAnimation starts viewing an animation performed by an Entity.
	ViewEntityAnimation(e Entity, a EntityAnimation)
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
	// ViewEmote views an emote being performed by another Entity.
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

func (NopViewer) ViewEntity(Entity)                                                          {}
func (NopViewer) HideEntity(Entity)                                                          {}
func (NopViewer) ViewEntityGameMode(Entity)                                                  {}
func (NopViewer) ViewEntityMovement(Entity, mgl64.Vec3, cube.Rotation, bool)                 {}
func (NopViewer) ViewEntityVelocity(Entity, mgl64.Vec3)                                      {}
func (NopViewer) ViewEntityTeleport(Entity, mgl64.Vec3)                                      {}
func (NopViewer) ViewChunk(ChunkPos, Dimension, map[cube.Pos]Block, *chunk.Chunk)            {}
func (NopViewer) ViewTime(int)                                                               {}
func (NopViewer) ViewTimeCycle(bool)                                                         {}
func (NopViewer) ViewEntityItems(Entity)                                                     {}
func (NopViewer) ViewEntityArmour(Entity)                                                    {}
func (NopViewer) ViewEntityAction(Entity, EntityAction)                                      {}
func (NopViewer) ViewEntityState(Entity)                                                     {}
func (NopViewer) ViewEntityAnimation(Entity, EntityAnimation)                                {}
func (NopViewer) ViewParticle(mgl64.Vec3, Particle)                                          {}
func (NopViewer) ViewSound(mgl64.Vec3, Sound)                                                {}
func (NopViewer) ViewBlockUpdate(cube.Pos, Block, int)                                       {}
func (NopViewer) ViewBlockAction(cube.Pos, BlockAction)                                      {}
func (NopViewer) ViewEmote(Entity, uuid.UUID)                                                {}
func (NopViewer) ViewSkin(Entity)                                                            {}
func (NopViewer) ViewWorldSpawn(cube.Pos)                                                    {}
func (NopViewer) ViewWeather(bool, bool)                                                     {}
func (NopViewer) ViewBrewingUpdate(time.Duration, time.Duration, int32, int32, int32, int32) {}
func (NopViewer) ViewFurnaceUpdate(time.Duration, time.Duration, time.Duration, time.Duration, time.Duration, time.Duration) {
}
