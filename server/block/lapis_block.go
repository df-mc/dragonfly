package block

import (
	"github.com/df-mc/dragonfly/server/item/tool"
)

// LapisBlock is a decorative mineral block that is crafted from lapis lazuli.
type LapisBlock struct {
	solid
}

// BreakInfo ...
func (l LapisBlock) BreakInfo() BreakInfo {
	return newBreakInfo(3, func(t tool.Tool) bool {
		return t.ToolType() == tool.TypePickaxe && t.HarvestLevel() >= tool.TierStone.HarvestLevel
	}, pickaxeEffective, oneOf(l), XPDropRange{})
}

// EncodeItem ...
func (LapisBlock) EncodeItem() (name string, meta int16) {
	return "minecraft:lapis_block", 0
}

// EncodeBlock ...
func (LapisBlock) EncodeBlock() (string, map[string]interface{}) {
	return "minecraft:lapis_block", nil
}
