package block

import (
	"github.com/df-mc/dragonfly/dragonfly/item"
	"github.com/df-mc/dragonfly/dragonfly/item/tool"
)

type EmeraldBlock struct{}

func (e EmeraldBlock) BreakInfo() BreakInfo {
	return BreakInfo{
		Hardness: 5,
		Harvestable: func(t tool.Tool) bool {
			return t.ToolType() == tool.TypePickaxe && t.HarvestLevel() >= tool.TierDiamond.HarvestLevel
		},
		Effective: pickaxeEffective,
		Drops:     simpleDrops(item.NewStack(e, 1)),
	}
}

// EncodeItem ...
func (e EmeraldBlock) EncodeItem() (id int32, meta int16) {
	return 133, 0
}

// EncodeBlock ...
func (e EmeraldBlock) EncodeBlock() (name string, properties map[string]interface{}) {
	return "minecraft:emerald_block", nil
}
