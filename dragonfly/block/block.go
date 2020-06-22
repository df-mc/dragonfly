package block

import (
	"git.jetbrains.space/dragonfly/dragonfly.git/dragonfly/entity/physics"
	"git.jetbrains.space/dragonfly/dragonfly.git/dragonfly/item"
	"git.jetbrains.space/dragonfly/dragonfly.git/dragonfly/world"
	"git.jetbrains.space/dragonfly/dragonfly.git/dragonfly/world/sound"
)

// Activatable represents a block that may be activated by a viewer of the world. When activated, the block
// will execute some specific logic.
type Activatable interface {
	// Activate activates the block at a specific block position. The face clicked is passed, as well as the
	// world in which the block was activated and the viewer that activated it.
	Activate(pos world.BlockPos, clickedFace world.Face, w *world.World, u item.User)
}

// LightEmitter represents a block that emits light when placed. Blocks such as torches or lanterns implement
// this interface.
type LightEmitter interface {
	// LightEmissionLevel returns the light emission level of the block, a number from 0-15 where 15 is the
	// brightest and 0 means it doesn't emit light at all.
	LightEmissionLevel() uint8
}

// LightDiffuser represents a block that diffuses light. This means that a specific amount of light levels
// will be subtracted when light passes through the block.
// Blocks that do not implement LightDiffuser will be assumed to be solid: Light will not be able to pass
// through these blocks.
type LightDiffuser interface {
	// LightDiffusionLevel returns the amount of light levels that is subtracted when light passes through
	// this block. Some locks, such as leaves, have this behaviour. A diffusion level of 15 means that all
	// light will be completely blocked when light passes through the block.
	LightDiffusionLevel() uint8
}

// Replaceable represents a block that may be replaced by another block automatically. An example is grass,
// which may be replaced by clicking it with another block.
type Replaceable interface {
	// ReplaceableBy returns a bool which indicates if the block is replaceable by another block.
	ReplaceableBy(b world.Block) bool
}

// replaceable checks if the block at the position passed is replaceable with the block passed.
func replaceable(w *world.World, pos world.BlockPos, with world.Block) bool {
	b := w.Block(pos)
	if replaceable, ok := b.(Replaceable); ok {
		return replaceable.ReplaceableBy(with)
	}
	return false
}

// firstReplaceable finds the first replaceable block position eligible to have a block placed on it after
// clicking on the position and face passed.
// If none can be found, the bool returned is false.
func firstReplaceable(w *world.World, pos world.BlockPos, face world.Face, with world.Block) (world.BlockPos, world.Face, bool) {
	if replaceable(w, pos, with) {
		// A replaceable block was clicked, so we can replace it. This will then be assumed to be placed on
		// the top face. (Torches, for example, will get attached to the floor when clicking tall grass.)
		return pos, world.FaceUp, true
	}
	side := pos.Side(face)
	if replaceable(w, side, with) {
		return side, face, true
	}
	return pos, face, false
}

// place places the block passed at the position passed. If the user implements the block.Placer interface, it
// will use its PlaceBlock method. If not, the block is placed without interaction from the user.
func place(w *world.World, pos world.BlockPos, b world.Block, user item.User, ctx *item.UseContext) {
	if placer, ok := user.(Placer); ok {
		placer.PlaceBlock(pos, b, ctx)
		return
	}
	w.PlaceBlock(pos, b)
	w.PlaySound(pos.Vec3(), sound.BlockPlace{Block: b})
}

// placed checks if an item was placed with the use context passed.
func placed(ctx *item.UseContext) bool {
	return ctx.CountSub > 0
}

// AABBer represents a block that has one or multiple specific Axis Aligned Bounding Boxes. These boxes are
// used to calculate collision.
type AABBer interface {
	// AABB returns all the axis aligned bounding boxes of the block.
	AABB(pos world.BlockPos, w *world.World) []physics.AABB
}
