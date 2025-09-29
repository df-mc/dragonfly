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

func (p DirtPath) Till() (world.Block, bool) {
	return Farmland{}, true
}

// NeighbourUpdateTick handles the turning from dirt path into dirt if a block is placed on top of it.
func (p DirtPath) NeighbourUpdateTick(pos, _ cube.Pos, tx *world.Tx) {
	up := pos.Side(cube.FaceUp)
	if tx.Block(up).Model().FaceSolid(up, cube.FaceDown, tx) {
		// A block with a solid side at the bottom was placed onto this one.
		tx.SetBlock(pos, Dirt{}, nil)
	}
}

func (p DirtPath) BreakInfo() BreakInfo {
	return newBreakInfo(0.65, alwaysHarvestable, shovelEffective, silkTouchOneOf(Dirt{}, p))
}

func (DirtPath) EncodeItem() (name string, meta int16) {
	return "minecraft:grass_path", 0
}

func (DirtPath) EncodeBlock() (string, map[string]any) {
	return "minecraft:grass_path", nil
}
