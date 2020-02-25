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
