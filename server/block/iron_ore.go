package block

import (
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/tool"
)

// IronOre is a mineral block found underground.
type IronOre struct {
	solid
	bassDrum

	// Type is the type of iron ore.
	Type OreType
}

// BreakInfo ...
func (i IronOre) BreakInfo() BreakInfo {
	return newBreakInfo(i.Type.Hardness(), func(t tool.Tool) bool {
		return t.ToolType() == tool.TypePickaxe && t.HarvestLevel() >= tool.TierStone.HarvestLevel
	}, pickaxeEffective, silkTouchOneOf(item.RawIron{}, i))
}

// EncodeItem ...
func (i IronOre) EncodeItem() (name string, meta int16) {
	return "minecraft:" + i.Type.Prefix() + "iron_ore", 0
}

// EncodeBlock ...
func (i IronOre) EncodeBlock() (string, map[string]interface{}) {
	return "minecraft:" + i.Type.Prefix() + "iron_ore", nil
}
