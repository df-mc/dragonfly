package block

import (
	"github.com/df-mc/dragonfly/server/item"
)

// Netherite is a precious mineral block made from 9 netherite ingots.
type Netherite struct {
	solid
	bassDrum
}

func (n Netherite) BreakInfo() BreakInfo {
	return newBreakInfo(50, func(t item.Tool) bool {
		return t.ToolType() == item.TypePickaxe && t.HarvestLevel() >= item.ToolTierDiamond.HarvestLevel
	}, pickaxeEffective, oneOf(n)).withBlastResistance(6000)
}

func (Netherite) PowersBeacon() bool {
	return true
}

func (Netherite) EncodeItem() (name string, meta int16) {
	return "minecraft:netherite_block", 0
}

func (Netherite) EncodeBlock() (string, map[string]any) {
	return "minecraft:netherite_block", nil
}
