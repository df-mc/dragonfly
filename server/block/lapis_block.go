package block

import (
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/tool"
)

// LapisBlock is a decorative mineral block that is crafted from lapis lazuli.
type LapisBlock struct {
	solid
}

// BreakInfo ...
func (l LapisBlock) BreakInfo() BreakInfo {
	return BreakInfo{
		Hardness: 3,
		Harvestable: func(t tool.Tool) bool {
			return t.ToolType() == tool.TypePickaxe && t.HarvestLevel() >= tool.TierStone.HarvestLevel
		},
		Effective: pickaxeEffective,
		Drops:     simpleDrops(item.NewStack(l, 1)),
	}
}

// EncodeItem ...
func (LapisBlock) EncodeItem() (id int32, name string, meta int16) {
	return 22, "minecraft:lapis_block", 0
}

// EncodeBlock ...
func (LapisBlock) EncodeBlock() (string, map[string]interface{}) {
	return "minecraft:lapis_block", nil
}
