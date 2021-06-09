package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
)

// RootedDirt is a natural decorative block.
type RootedDirt struct {
	solid
}

// BoneMeal ...
func (r RootedDirt) BoneMeal(pos cube.Pos, w *world.World) bool {
	if _, ok := w.Block(pos.Side(cube.FaceDown)).(Air); ok {
		w.SetBlock(pos.Side(cube.FaceDown), HangingRoots{})
		return true
	}
	return false
}

// BreakInfo ...
func (r RootedDirt) BreakInfo() BreakInfo {
	return newBreakInfo(0.5, alwaysHarvestable, nothingEffective, oneOf(r))
}

// EncodeItem ...
func (r RootedDirt) EncodeItem() (name string, meta int16) {
	return "minecraft:dirt_with_roots", 0
}

// EncodeBlock ...
func (r RootedDirt) EncodeBlock() (string, map[string]interface{}) {
	return "minecraft:dirt_with_roots", nil
}
