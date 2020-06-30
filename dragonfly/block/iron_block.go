package block

import (
	"github.com/df-mc/dragonfly/dragonfly/item"
	"github.com/df-mc/dragonfly/dragonfly/item/tool"
)

type IronBlock struct{}

func (i IronBlock) BreakInfo() BreakInfo {
	return BreakInfo{
		Hardness: 5,
		Harvestable: func(t tool.Tool) bool {
			return t.ToolType() == tool.TypePickaxe && t.HarvestLevel() >= tool.TierDiamond.HarvestLevel
		},
		Effective: pickaxeEffective,
		Drops:     simpleDrops(item.NewStack(i, 1)),
	}
}

// EncodeItem ...
func (i IronBlock) EncodeItem() (id int32, meta int16) {
	return 42, 0
}

// EncodeBlock ...
func (i IronBlock) EncodeBlock() (name string, properties map[string]interface{}) {
	return "minecraft:iron_block", nil
}
