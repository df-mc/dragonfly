package block

import (
	"github.com/df-mc/dragonfly/server/item/tool"
)

// RawCopperBlock is a raw metal block equivalent to nine raw copper.
type RawCopperBlock struct {
	solid
	bassDrum
}

// BreakInfo ...
func (r RawCopperBlock) BreakInfo() BreakInfo {
	return newBreakInfo(5, func(t tool.Tool) bool {
		return t.ToolType() == tool.TypePickaxe && t.HarvestLevel() >= tool.TierStone.HarvestLevel
	}, pickaxeEffective, oneOf(r))
}

// EncodeItem ...
func (RawCopperBlock) EncodeItem() (name string, meta int16) {
	return "minecraft:raw_copper_block", 0
}

// EncodeBlock ...
func (RawCopperBlock) EncodeBlock() (string, map[string]interface{}) {
	return "minecraft:raw_copper_block", nil
}
