package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
)

// Barrier is a transparent solid block used to create invisible boundaries.
type Barrier struct {
	transparent
	solid
}

// CanDisplace ...
func (Barrier) CanDisplace(b world.Liquid) bool {
	_, water := b.(Water)
	return water
}

// SideClosed ...
func (Barrier) SideClosed(cube.Pos, cube.Pos, *world.World) bool {
	return false
}

// EncodeItem ...
func (Barrier) EncodeItem() (name string, meta int16) {
	return "minecraft:barrier", 0
}

// EncodeBlock ...
func (Barrier) EncodeBlock() (string, map[string]any) {
	return "minecraft:barrier", nil
}
