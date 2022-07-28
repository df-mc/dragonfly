package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
)

type RedstoneBlock struct{ solid }

// EncodeItem ...
func (b RedstoneBlock) EncodeItem() (name string, meta int16) {
	return "minecraft:redstone_block", 0
}

// EncodeBlock ...
func (b RedstoneBlock) EncodeBlock() (string, map[string]any) {
	return "minecraft:redstone_block", nil
}

// WeakPower ...
func (b RedstoneBlock) WeakPower(cube.Pos, cube.Face, *world.World) int {
	return 15
}

// StrongPower ...
func (b RedstoneBlock) StrongPower(cube.Pos, cube.Face, *world.World) int {
	return 0
}
