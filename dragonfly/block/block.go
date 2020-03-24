package block

import (
	"git.jetbrains.space/dragonfly/dragonfly.git/dragonfly/world"
)

// Activatable represents a block that may be activated by a viewer of the world. When activated, the block
// will execute some specific logic.
type Activatable interface {
	// Activate activates the block at a specific block position. The face clicked is passed, as well as the
	// world in which the block was activated and the viewer that activated it.
	Activate(pos world.BlockPos, clickedFace world.Face, w *world.World, e world.Entity)
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
