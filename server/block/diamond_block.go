package block

import (
	"github.com/df-mc/dragonfly/server/item/tool"
)

// DiamondBlock is a block which can only be gained by crafting it.
type DiamondBlock struct {
	solid
}

// BreakInfo ...
func (d DiamondBlock) BreakInfo() BreakInfo {
	return newBreakInfo(5, func(t tool.Tool) bool {
		return t.ToolType() == tool.TypePickaxe && t.HarvestLevel() >= tool.TierIron.HarvestLevel
	}, pickaxeEffective, oneOf(d), XPDropRange{})
}

// PowersBeacon ...
func (DiamondBlock) PowersBeacon() bool {
	return true
}

// EncodeItem ...
func (DiamondBlock) EncodeItem() (name string, meta int16) {
	return "minecraft:diamond_block", 0
}

// EncodeBlock ...
func (DiamondBlock) EncodeBlock() (string, map[string]interface{}) {
	return "minecraft:diamond_block", nil
}
