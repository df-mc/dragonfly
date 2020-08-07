package block

import (
	"github.com/df-mc/dragonfly/dragonfly/entity"
	"github.com/df-mc/dragonfly/dragonfly/entity/effect"
	"github.com/df-mc/dragonfly/dragonfly/item"
	"github.com/df-mc/dragonfly/dragonfly/world"
	"github.com/df-mc/dragonfly/dragonfly/world/sound"
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
	// this block. Some blocks, such as leaves, have this behaviour. A diffusion level of 15 means that all
	// light will be completely blocked when it passes through the block.
	LightDiffusionLevel() uint8
}

// Replaceable represents a block that may be replaced by another block automatically. An example is grass,
// which may be replaced by clicking it with another block.
type Replaceable interface {
	// ReplaceableBy returns a bool which indicates if the block is replaceableWith by another block.
	ReplaceableBy(b world.Block) bool
}

// BeaconSource represents a block which is capable of contributing to powering a beacon pyramid.
type BeaconSource interface {
	// PowersBeacon returns a bool which indicates whether this block can contribute to powering up a
	// beacon pyramid.
	PowersBeacon() bool
}

// BonemealAffected represents a block that is affected when bonemeal is used on it.
type BonemealAffected interface {
	// Bonemeal attempts to affect the block.
	Bonemeal(pos world.BlockPos, w *world.World) bool
}

// beaconAffected represents an entity that can be powered by a beacon. Only players will implement this.
type beaconAffected interface {
	// AddEffect adds a specific effect to the entity that implements this interface.
	AddEffect(e effect.Effect)

	// BeaconAffected returns whether this entity can be powered by a beacon.
	BeaconAffected() bool
}

// replaceableWith checks if the block at the position passed is replaceable with the block passed.
func replaceableWith(w *world.World, pos world.BlockPos, with world.Block) bool {
	if pos.OutOfBounds() {
		return false
	}
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
	if replaceableWith(w, pos, with) {
		// A replaceableWith block was clicked, so we can replace it. This will then be assumed to be placed on
		// the top face. (Torches, for example, will get attached to the floor when clicking tall grass.)
		return pos, world.FaceUp, true
	}
	side := pos.Side(face)
	if replaceableWith(w, side, with) {
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

// boolByte returns 1 if the bool passed is true, or 0 if it is false.
func boolByte(b bool) uint8 {
	if b {
		return 1
	}
	return 0
}

// noNBT may be embedded by blocks that have no NBT.
type noNBT struct{}

// HasNBT ...
func (noNBT) HasNBT() bool {
	return false
}

// nbt may be embedded by blocks that do have NBT.
type nbt struct{}

// HasNBT ...
func (nbt) HasNBT() bool {
	return true
}

// replaceable is a struct that may be embedded to make a block replaceable by any other block.
type replaceable struct{}

// ReplaceableBy ...
func (replaceable) ReplaceableBy(world.Block) bool {
	return true
}

// transparent is a struct that may be embedded to make a block transparent to light. Light will be able to
// pass through this block freely.
type transparent struct{}

// LightDiffusionLevel ...
func (transparent) LightDiffusionLevel() uint8 {
	return 0
}

// GravityAffected represents blocks affected by gravity.
type GravityAffected interface {
	// CanSolidify returns whether the falling block can return back to a normal block without being on the ground.
	CanSolidify(pos world.BlockPos, w *world.World) bool
}

// gravityAffected is a struct that may be embedded for blocks affected by gravity.
type gravityAffected struct{}

// CanSolidify ...
func (g gravityAffected) CanSolidify(world.BlockPos, *world.World) bool {
	return false
}

// fall spawns a falling block entity at the given position.
func (g gravityAffected) fall(b world.Block, pos world.BlockPos, w *world.World) {
	_, air := w.Block(pos.Side(world.FaceDown)).(Air)
	_, liquid := w.Liquid(pos.Side(world.FaceDown))
	if air || liquid {
		w.BreakBlock(pos)

		e := entity.NewFallingBlock(b, pos.Vec3())
		w.AddEntity(e)
	}
}
