package block

import (
	"github.com/df-mc/dragonfly/server/item"
)

// RawIron is a raw metal block equivalent to nine raw iron.
type RawIron struct {
	solid
	bassDrum
}

// BreakInfo ...
func (r RawIron) BreakInfo() BreakInfo {
	return newBreakInfo(5, func(t item.Tool) bool {
		return t.ToolType() == item.TypePickaxe && t.HarvestLevel() >= item.ToolTierStone.HarvestLevel
	}, pickaxeEffective, oneOf(r))
}

// EncodeItem ...
func (RawIron) EncodeItem() (name string, meta int16) {
	return "minecraft:raw_iron_block", 0
}

// EncodeBlock ...
func (RawIron) EncodeBlock() (string, map[string]any) {
	return "minecraft:raw_iron_block", nil
}
