package block

import (
	"github.com/df-mc/dragonfly/dragonfly/item"
	"github.com/df-mc/dragonfly/dragonfly/item/tool"
)

type Farmland struct {
}

//TODO: Add Farmland wetness and planting functionality

func (f Farmland) BreakInfo() BreakInfo {
	return BreakInfo{
		Hardness: .6,
		Harvestable: func(t tool.Tool) bool {
			return t.ToolType() == tool.TypeNone
		},
		Effective: shovelEffective,
		Drops:     simpleDrops(item.NewStack(Dirt{}, 1)),
	}
}

func (f Farmland) EncodeBlock() (name string, properties map[string]interface{}) {
	return "minecraft:farmland", nil
}
