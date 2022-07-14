package block

import (
	"github.com/df-mc/dragonfly/server/item"
)

// Netherite is a precious mineral block made from 9 netherite ingots.
type Netherite struct {
	solid
	bassDrum
}

// BreakInfo ...
func (n Netherite) BreakInfo() BreakInfo {
	return newBreakInfo(50, func(t item.Tool) bool {
		return t.ToolType() == item.TypePickaxe && t.HarvestLevel() >= item.ToolTierDiamond.HarvestLevel
	}, pickaxeEffective, oneOf(n))
}

// PowersBeacon ...
func (Netherite) PowersBeacon() bool {
	return true
}

// EncodeItem ...
func (Netherite) EncodeItem() (name string, meta int16) {
	return "minecraft:netherite_block", 0
}

// EncodeBlock ...
func (Netherite) EncodeBlock() (string, map[string]any) {
	return "minecraft:netherite_block", nil
}
