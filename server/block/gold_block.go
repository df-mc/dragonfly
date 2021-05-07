package block

import (
	"github.com/df-mc/dragonfly/server/block/instrument"
	"github.com/df-mc/dragonfly/server/item/tool"
)

// GoldBlock is a precious metal block crafted from 9 gold ingots.
type GoldBlock struct {
	solid
}

// Instrument ...
func (g GoldBlock) Instrument() instrument.Instrument {
	return instrument.Bell()
}

// BreakInfo ...
func (g GoldBlock) BreakInfo() BreakInfo {
	return newBreakInfo(5, func(t tool.Tool) bool {
		return t.ToolType() == tool.TypePickaxe && t.HarvestLevel() >= tool.TierIron.HarvestLevel
	}, pickaxeEffective, oneOf(g))
}

// PowersBeacon ...
func (GoldBlock) PowersBeacon() bool {
	return true
}

// EncodeItem ...
func (GoldBlock) EncodeItem() (name string, meta int16) {
	return "minecraft:gold_block", 0
}

// EncodeBlock ...
func (GoldBlock) EncodeBlock() (string, map[string]interface{}) {
	return "minecraft:gold_block", nil
}
