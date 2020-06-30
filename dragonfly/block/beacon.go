package block

import (
	"github.com/df-mc/dragonfly/dragonfly/item"
	"github.com/df-mc/dragonfly/dragonfly/item/tool"
)

type Beacon struct{}

func (b Beacon) BreakInfo() BreakInfo {
	return BreakInfo{
		Hardness: 3,
		Harvestable: func(t tool.Tool) bool {
			return t.ToolType() == tool.TypePickaxe && t.HarvestLevel() >= tool.TierDiamond.HarvestLevel
		},
		Effective: pickaxeEffective,
		Drops:     simpleDrops(item.NewStack(b, 1)),
	}
}

// EncodeItem ...
func (b Beacon) EncodeItem() (id int32, meta int16) {
	return 138, 0
}

// EncodeBlock ...
func (b Beacon) EncodeBlock() (name string, properties map[string]interface{}) {
	return "minecraft:beacon", nil
}
