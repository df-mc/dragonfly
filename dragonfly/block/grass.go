package block

import (
	"git.jetbrains.space/dragonfly/dragonfly.git/dragonfly/item"
)

// Grass blocks generate abundantly across the surface of the world.
type Grass struct{}

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
func (Grass) EncodeItem() (id int32, meta int16) {
	return 2, 0
}

// EncodeBlock ...
func (Grass) EncodeBlock() (name string, properties map[string]interface{}) {
	return "minecraft:grass", nil
}
