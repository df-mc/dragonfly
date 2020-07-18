package block

import (
	"github.com/df-mc/dragonfly/dragonfly/item"
	"github.com/df-mc/dragonfly/dragonfly/item/tool"
)

// EmeraldBlock is a precious mineral block crafted using 9 emeralds.
type EmeraldBlock struct{}

// BreakInfo ...
func (e EmeraldBlock) BreakInfo() BreakInfo {
	return BreakInfo{
		Hardness: 5,
		Harvestable: func(t tool.Tool) bool {
			return t.ToolType() == tool.TypePickaxe && t.HarvestLevel() >= tool.TierIron.HarvestLevel
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

// Hash ...
func (EmeraldBlock) Hash() uint64 {
	return hashEmeraldBlock
}
