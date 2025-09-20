package block

import (
	"github.com/df-mc/dragonfly/server/item"
)

// RawGold is a raw metal block equivalent to nine raw gold.
type RawGold struct {
	solid
	bassDrum
}

func (g RawGold) BreakInfo() BreakInfo {
	return newBreakInfo(5, func(t item.Tool) bool {
		return t.ToolType() == item.TypePickaxe && t.HarvestLevel() >= item.ToolTierIron.HarvestLevel
	}, pickaxeEffective, oneOf(g)).withBlastResistance(30)
}

func (RawGold) EncodeItem() (name string, meta int16) {
	return "minecraft:raw_gold_block", 0
}

func (RawGold) EncodeBlock() (string, map[string]any) {
	return "minecraft:raw_gold_block", nil
}
