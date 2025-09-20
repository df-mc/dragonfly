package block

import (
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world/sound"
)

// Emerald is a precious mineral block crafted using 9 emeralds.
type Emerald struct {
	solid
}

func (e Emerald) Instrument() sound.Instrument {
	return sound.Bit()
}

func (e Emerald) BreakInfo() BreakInfo {
	return newBreakInfo(5, func(t item.Tool) bool {
		return t.ToolType() == item.TypePickaxe && t.HarvestLevel() >= item.ToolTierIron.HarvestLevel
	}, pickaxeEffective, oneOf(e)).withBlastResistance(30)
}

func (Emerald) PowersBeacon() bool {
	return true
}

func (Emerald) EncodeItem() (name string, meta int16) {
	return "minecraft:emerald_block", 0
}

func (Emerald) EncodeBlock() (string, map[string]any) {
	return "minecraft:emerald_block", nil
}
