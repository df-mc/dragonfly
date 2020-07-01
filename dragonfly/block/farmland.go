package block

import (
	"github.com/df-mc/dragonfly/dragonfly/item"
)

type Farmland struct {
}

//TODO: Add Farmland wetness and planting functionality

func (f Farmland) BreakInfo() BreakInfo {
	return BreakInfo{
		Hardness:    0.6,
		Harvestable: alwaysHarvestable,
		Effective:   shovelEffective,
		Drops:       simpleDrops(item.NewStack(Dirt{}, 1)),
	}
}

func (f Farmland) EncodeBlock() (name string, properties map[string]interface{}) {
	return "minecraft:farmland", nil
}
