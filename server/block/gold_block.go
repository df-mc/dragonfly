package block

import (
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world/sound"
)

// GoldBlock is a precious metal block crafted from 9 gold ingots.
type GoldBlock struct {
	solid
}

// Instrument ...
func (g GoldBlock) Instrument() sound.Instrument {
	return sound.Bell()
}

// BreakInfo ...
func (g GoldBlock) BreakInfo() BreakInfo {
	return newBreakInfo(5, func(t item.Tool) bool {
		return t.ToolType() == item.TypePickaxe && t.HarvestLevel() >= item.TierIron.HarvestLevel
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
