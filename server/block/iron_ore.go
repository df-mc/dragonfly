package block

import (
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/tool"
)

// IronOre is a mineral block found underground.
type IronOre struct {
	solid
	bassDrum
}

// BreakInfo ...
func (i IronOre) BreakInfo() BreakInfo {
	return newBreakInfo(3, func(t tool.Tool) bool {
		return t.ToolType() == tool.TypePickaxe && t.HarvestLevel() >= tool.TierStone.HarvestLevel
	}, pickaxeEffective, oneOf(item.RawIron{})) //TODO: Silk Touch
}

// EncodeItem ...
func (IronOre) EncodeItem() (name string, meta int16) {
	return "minecraft:iron_ore", 0
}

// EncodeBlock ...
func (IronOre) EncodeBlock() (string, map[string]interface{}) {
	return "minecraft:iron_ore", nil
}
