package block

import (
	"github.com/df-mc/dragonfly/server/item"
)

// Lapis is a decorative mineral block that is crafted from lapis lazuli.
type Lapis struct {
	solid
}

// BreakInfo ...
func (l Lapis) BreakInfo() BreakInfo {
	return newBreakInfo(3, func(t item.Tool) bool {
		return t.ToolType() == item.TypePickaxe && t.HarvestLevel() >= item.ToolTierStone.HarvestLevel
	}, pickaxeEffective, oneOf(l))
}

// EncodeItem ...
func (Lapis) EncodeItem() (name string, meta int16) {
	return "minecraft:lapis_block", 0
}

// EncodeBlock ...
func (Lapis) EncodeBlock() (string, map[string]any) {
	return "minecraft:lapis_block", nil
}
