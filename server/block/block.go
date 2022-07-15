package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/entity"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
)

// Activatable represents a block that may be activated by a viewer of the world. When activated, the block
// will execute some specific logic.
type Activatable interface {
	// Activate activates the block at a specific block position. The face clicked is passed, as well as the
	// world in which the block was activated and the viewer that activated it.
	// Activate returns a bool indicating if activating the block was used successfully.
	Activate(pos cube.Pos, clickedFace cube.Face, w *world.World, u item.User) bool
}

// Pickable represents a block that may give a different item then the block itself when picked.
type Pickable interface {
	// Pick returns the item that is picked when the block is picked.
	Pick() item.Stack
}

// Punchable represents a block that may be punched by a viewer of the world. When punched, the block
// will execute some specific logic.
type Punchable interface {
	// Punch punches the block at a specific block position. The face clicked is passed, as well as the
	// world in which the block was punched and the viewer that punched it.
	Punch(pos cube.Pos, clickedFace cube.Face, w *world.World, u item.User)
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

// EntityLander represents a block that reacts to an entity landing on it after falling.
type EntityLander interface {
	// EntityLand is called when an entity lands on the block.
	EntityLand(pos cube.Pos, w *world.World, e world.Entity, distance *float64)
}

// EntityInsider represents a block that reacts to an entity going inside its 1x1x1 axis
// aligned bounding box.
type EntityInsider interface {
	// EntityInside is called when an entity goes inside the block's 1x1x1 axis aligned bounding box.
	EntityInside(pos cube.Pos, w *world.World, e world.Entity)
}

// Frictional represents a block that may have a custom friction value, friction is used for entity drag when the
// entity is on ground. If a block does not implement this interface, it should be assumed that its friction is 0.6.
type Frictional interface {
	// Friction returns the block's friction value.
	Friction() float64
}

func calculateFace(user item.User, placePos cube.Pos) cube.Face {
	userPos := user.Position()
	pos := cube.PosFromVec3(userPos)
	if abs(pos[0]-placePos[0]) < 2 && abs(pos[2]-placePos[2]) < 2 {
		y := userPos[1]
		if eyed, ok := user.(entity.Eyed); ok {
			y += eyed.EyeHeight()
		}

		if y-float64(placePos[1]) > 2.0 {
			return cube.FaceUp
		} else if float64(placePos[1])-y > 0.0 {
			return cube.FaceDown
		}
	}
	return user.Facing().Opposite().Face()
}

func abs(x int) int {
	if x > 0 {
		return x
	}
	return -x
}

// replaceableWith checks if the block at the position passed is replaceable with the block passed.
func replaceableWith(w *world.World, pos cube.Pos, with world.Block) bool {
	if pos.OutOfBounds(w.Range()) {
		return false
	}
	b := w.Block(pos)
	if replaceable, ok := b.(Replaceable); ok {
		return replaceable.ReplaceableBy(with) && b != with
	}
	return false
}

// firstReplaceable finds the first replaceable block position eligible to have a block placed on it after
// clicking on the position and face passed.
// If none can be found, the bool returned is false.
func firstReplaceable(w *world.World, pos cube.Pos, face cube.Face, with world.Block) (cube.Pos, cube.Face, bool) {
	if replaceableWith(w, pos, with) {
		// A replaceableWith block was clicked, so we can replace it. This will then be assumed to be placed on
		// the top face. (Torches, for example, will get attached to the floor when clicking tall grass.)
		return pos, cube.FaceUp, true
	}
	side := pos.Side(face)
	if replaceableWith(w, side, with) {
		return side, face, true
	}
	return pos, face, false
}

// place places the block passed at the position passed. If the user implements the block.Placer interface, it
// will use its PlaceBlock method. If not, the block is placed without interaction from the user.
func place(w *world.World, pos cube.Pos, b world.Block, user item.User, ctx *item.UseContext) {
	if placer, ok := user.(Placer); ok {
		placer.PlaceBlock(pos, b, ctx)
		return
	}
	w.SetBlock(pos, b, nil)
	w.PlaySound(pos.Vec3(), sound.BlockPlace{Block: b})
}

// horizontalDirection returns the horizontal direction of the given direction. This is a legacy type still used in
// various blocks.
func horizontalDirection(d cube.Direction) cube.Direction {
	switch d {
	case cube.South:
		return cube.North
	case cube.West:
		return cube.South
	case cube.North:
		return cube.West
	case cube.East:
		return cube.East
	}
	panic("invalid direction")
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

// gravityAffected is a struct that may be embedded for blocks affected by gravity.
type gravityAffected struct{}

// Solidifies ...
func (g gravityAffected) Solidifies(cube.Pos, *world.World) bool {
	return false
}

// fall spawns a falling block entity at the given position.
func (g gravityAffected) fall(b world.Block, pos cube.Pos, w *world.World) {
	_, air := w.Block(pos.Side(cube.FaceDown)).(Air)
	_, liquid := w.Liquid(pos.Side(cube.FaceDown))
	if air || liquid {
		w.SetBlock(pos, nil, nil)
		w.AddEntity(entity.NewFallingBlock(b, pos.Vec3Middle()))
	}
}

// Flammable is an interface for blocks that can catch on fire.
type Flammable interface {
	// FlammabilityInfo returns information about a block's behavior involving fire.
	FlammabilityInfo() FlammabilityInfo
}

// FlammabilityInfo contains values related to block behaviors involving fire.
type FlammabilityInfo struct {
	// Encouragement is the chance a block will catch on fire during attempted fire spread.
	Encouragement int
	// Flammability is the chance a block will burn away during a fire block tick.
	Flammability int
	// LavaFlammable returns whether it can catch on fire from lava.
	LavaFlammable bool
}

// newFlammabilityInfo creates a FlammabilityInfo struct with the properties passed.
func newFlammabilityInfo(encouragement, flammability int, lavaFlammable bool) FlammabilityInfo {
	return FlammabilityInfo{
		Encouragement: encouragement,
		Flammability:  flammability,
		LavaFlammable: lavaFlammable,
	}
}

// bass is a struct that may be embedded for blocks that create a bass sound.
type bass struct{}

// Instrument ...
func (bass) Instrument() sound.Instrument {
	return sound.Bass()
}

// snare is a struct that may be embedded for blocks that create a snare drum sound.
type snare struct{}

// Instrument ...
func (snare) Instrument() sound.Instrument {
	return sound.Snare()
}

// clicksAndSticks is a struct that may be embedded for blocks that create a clicks and sticks sound.
type clicksAndSticks struct{}

// Instrument ...
func (clicksAndSticks) Instrument() sound.Instrument {
	return sound.ClicksAndSticks()
}

// bassDrum is a struct that may be embedded for blocks that create a bass drum sound.
type bassDrum struct{}

// Instrument ...
func (bassDrum) Instrument() sound.Instrument {
	return sound.BassDrum()
}
