package block

import (
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world/sound"
)

// Iron is a precious metal block made from 9 iron ingots.
type Iron struct {
	solid
}

// Instrument ...
func (i Iron) Instrument() sound.Instrument {
	return sound.IronXylophone()
}

// BreakInfo ...
func (i Iron) BreakInfo() BreakInfo {
	return newBreakInfo(5, func(t item.Tool) bool {
		return t.ToolType() == item.TypePickaxe && t.HarvestLevel() >= item.ToolTierStone.HarvestLevel
	}, pickaxeEffective, oneOf(i))
}

// PowersBeacon ...
func (Iron) PowersBeacon() bool {
	return true
}

// EncodeItem ...
func (Iron) EncodeItem() (name string, meta int16) {
	return "minecraft:iron_block", 0
}

// EncodeBlock ...
func (Iron) EncodeBlock() (string, map[string]any) {
	return "minecraft:iron_block", nil
}
