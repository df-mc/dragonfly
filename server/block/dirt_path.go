package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
)

// DirtPath is a decorative block that can be created by using a shovel on a dirt or grass block.
type DirtPath struct {
	tilledGrass
	transparent
}

// Till ...
func (p DirtPath) Till() (world.Block, bool) {
	return Farmland{}, true
}

// NeighbourUpdateTick handles the turning from dirt path into dirt if a block is placed on top of it.
func (p DirtPath) NeighbourUpdateTick(pos, _ cube.Pos, w *world.World) {
	up := pos.Add(cube.Pos{0, 1})
	if w.Block(up).Model().FaceSolid(up, cube.FaceDown, w) {
		// A block with a solid side at the bottom was placed onto this one.
		w.SetBlock(pos, Dirt{}, nil)
	}
}

// BreakInfo ...
func (p DirtPath) BreakInfo() BreakInfo {
	return newBreakInfo(0.6, alwaysHarvestable, shovelEffective, silkTouchOneOf(Dirt{}, p))
}

// EncodeItem ...
func (DirtPath) EncodeItem() (name string, meta int16) {
	return "minecraft:grass_path", 0
}

// EncodeBlock ...
func (DirtPath) EncodeBlock() (string, map[string]any) {
	return "minecraft:grass_path", nil
}
