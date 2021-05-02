package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
)

// DirtPath is a decorative block that can be created by using a shovel on a dirt or grass block.
type DirtPath struct {
	tilledGrass
	transparent
}

// NeighbourUpdateTick handles the turning from dirt path into dirt if a block is placed on top of it.
func (p DirtPath) NeighbourUpdateTick(pos, _ cube.Pos, w *world.World) {
	up := pos.Add(cube.Pos{0, 1})
	if w.Block(up).Model().FaceSolid(up, cube.FaceDown, w) {
		// A block with a solid side at the bottom was placed onto this one.
		w.SetBlock(pos, Dirt{})
	}
}

// BreakInfo ...
func (p DirtPath) BreakInfo() BreakInfo {
	return BreakInfo{
		Hardness:    0.6,
		Harvestable: alwaysHarvestable,
		Effective:   shovelEffective,
		Drops:       simpleDrops(item.NewStack(Dirt{}, 1)), //TODO: Silk Touch
	}
}

// EncodeItem ...
func (DirtPath) EncodeItem() (id int32, name string, meta int16) {
	return 198, "minecraft:grass_path", 0
}

// EncodeBlock ...
func (DirtPath) EncodeBlock() (string, map[string]interface{}) {
	return "minecraft:grass_path", nil
}
