package block

import (
	"github.com/df-mc/dragonfly/server/item/tool"
)

// RawGoldBlock is a raw metal block equivalent to nine raw gold.
type RawGoldBlock struct {
	solid
	bassDrum
}

// BreakInfo ...
func (g RawGoldBlock) BreakInfo() BreakInfo {
	return newBreakInfo(5, func(t tool.Tool) bool {
		return t.ToolType() == tool.TypePickaxe && t.HarvestLevel() >= tool.TierIron.HarvestLevel
	}, pickaxeEffective, oneOf(g), XPDropRange{})
}

// EncodeItem ...
func (RawGoldBlock) EncodeItem() (name string, meta int16) {
	return "minecraft:raw_gold_block", 0
}

// EncodeBlock ...
func (RawGoldBlock) EncodeBlock() (string, map[string]interface{}) {
	return "minecraft:raw_gold_block", nil
}
