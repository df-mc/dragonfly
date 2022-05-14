package block

import (
	"github.com/df-mc/dragonfly/server/item"
)

// CoalBlock is a precious mineral block made from 9 coal.
type CoalBlock struct {
	solid
	bassDrum
}

// FlammabilityInfo ...
func (c CoalBlock) FlammabilityInfo() FlammabilityInfo {
	return newFlammabilityInfo(5, 5, false)
}

// BreakInfo ...
func (c CoalBlock) BreakInfo() BreakInfo {
	return newBreakInfo(5, func(t item.Tool) bool {
		return t.ToolType() == item.TypePickaxe && t.HarvestLevel() >= item.ToolTierWood.HarvestLevel
	}, pickaxeEffective, oneOf(c), XPDropRange{})
}

// EncodeItem ...
func (CoalBlock) EncodeItem() (name string, meta int16) {
	return "minecraft:coal_block", 0
}

// EncodeBlock ...
func (CoalBlock) EncodeBlock() (name string, properties map[string]any) {
	return "minecraft:coal_block", nil
}
