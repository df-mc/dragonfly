package block

import (
	"github.com/df-mc/dragonfly/server/item"
)

// Diamond is a block which can only be gained by crafting it.
type Diamond struct {
	solid
}

// BreakInfo ...
func (d Diamond) BreakInfo() BreakInfo {
	return newBreakInfo(5, func(t item.Tool) bool {
		return t.ToolType() == item.TypePickaxe && t.HarvestLevel() >= item.ToolTierIron.HarvestLevel
	}, pickaxeEffective, oneOf(d))
}

// PowersBeacon ...
func (Diamond) PowersBeacon() bool {
	return true
}

// EncodeItem ...
func (Diamond) EncodeItem() (name string, meta int16) {
	return "minecraft:diamond_block", 0
}

// EncodeBlock ...
func (Diamond) EncodeBlock() (string, map[string]any) {
	return "minecraft:diamond_block", nil
}
