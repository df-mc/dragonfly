package block

import (
	"git.jetbrains.space/dragonfly/dragonfly.git/dragonfly/item"
	"git.jetbrains.space/dragonfly/dragonfly.git/dragonfly/world"
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
