package block

import (
	"github.com/df-mc/dragonfly/dragonfly/item"
	"github.com/df-mc/dragonfly/dragonfly/item/tool"
)

// NetheriteBlock is a precious mineral block made from 9 netherite ingots.
type NetheriteBlock struct {
	solid
	bassDrum
}

// BreakInfo ...
func (n NetheriteBlock) BreakInfo() BreakInfo {
	return BreakInfo{
		Hardness: 50,
		Harvestable: func(t tool.Tool) bool {
			return t.ToolType() == tool.TypePickaxe && t.HarvestLevel() >= tool.TierDiamond.HarvestLevel
		},
		Effective: pickaxeEffective,
		Drops:     simpleDrops(item.NewStack(n, 1)),
	}
}

// PowersBeacon ...
func (NetheriteBlock) PowersBeacon() bool {
	return true
}

// EncodeItem ...
func (NetheriteBlock) EncodeItem() (id int32, meta int16) {
	return -270, 0
}

// EncodeBlock ...
func (NetheriteBlock) EncodeBlock() (string, map[string]interface{}) {
	return "minecraft:netherite_block", nil
}
