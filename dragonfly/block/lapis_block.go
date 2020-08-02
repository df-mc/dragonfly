package block

import (
	"github.com/df-mc/dragonfly/dragonfly/item"
	"github.com/df-mc/dragonfly/dragonfly/item/tool"
)

// LapisBlock is a decorative mineral block that is crafted from lapis lazuli.
type LapisBlock struct {
	noNBT
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
func (LapisBlock) EncodeItem() (id int32, meta int16) {
	return 22, 0
}

// EncodeBlock ...
func (LapisBlock) EncodeBlock() (name string, properties map[string]interface{}) {
	return "minecraft:lapis_block", nil
}

// Hash ...
func (l LapisBlock) Hash() uint64 {
	return hashLapisBlock
}
