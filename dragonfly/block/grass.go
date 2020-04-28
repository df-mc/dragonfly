package block

import (
	"git.jetbrains.space/dragonfly/dragonfly.git/dragonfly/item"
)

// Grass blocks generate abundantly across the surface of the world.
type Grass struct {
	// Path specifies if the grass was made into a path or not. If true, the block will have only 15/16th of
	// the height of a full block.
	Path bool
}

// BreakInfo ...
func (g Grass) BreakInfo() BreakInfo {
	return BreakInfo{
		Hardness:    0.6,
		Harvestable: alwaysHarvestable,
		Effective:   shovelEffective,
		Drops:       simpleDrops(item.NewStack(Dirt{}, 1)),
	}
}

// EncodeItem ...
func (g Grass) EncodeItem() (id int32, meta int16) {
	if g.Path {
		return 198, 0
	}
	return 2, 0
}

// EncodeBlock ...
func (g Grass) EncodeBlock() (name string, properties map[string]interface{}) {
	if g.Path {
		return "minecraft:grass_path", nil
	}
	return "minecraft:grass", nil
}
