package block

import (
	"github.com/df-mc/dragonfly/server/item"
)

// RawGoldBlock is a raw metal block equivalent to nine raw gold.
type RawGoldBlock struct {
	solid
	bassDrum
}

// BreakInfo ...
func (g RawGoldBlock) BreakInfo() BreakInfo {
	return newBreakInfo(5, func(t item.Tool) bool {
		return t.ToolType() == item.TypePickaxe && t.HarvestLevel() >= item.ToolTierIron.HarvestLevel
	}, pickaxeEffective, oneOf(g)).withExplosionInfo(6, false)
}

// EncodeItem ...
func (RawGoldBlock) EncodeItem() (name string, meta int16) {
	return "minecraft:raw_gold_block", 0
}

// EncodeBlock ...
func (RawGoldBlock) EncodeBlock() (string, map[string]any) {
	return "minecraft:raw_gold_block", nil
}
