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
	}, pickaxeEffective, oneOf(item.RawIron{})) //TODO: Silk Touch
}

// EncodeItem ...
func (i IronOre) EncodeItem() (name string, meta int16) {
	switch i.Type {
	case StoneOre():
		return "minecraft:iron_ore", 0
	case DeepslateOre():
		return "minecraft:deepslate_iron_ore", 0
	}
	panic("unknown ore type")
}

// EncodeBlock ...
func (i IronOre) EncodeBlock() (string, map[string]interface{}) {
	switch i.Type {
	case StoneOre():
		return "minecraft:iron_ore", nil
	case DeepslateOre():
		return "minecraft:deepslate_iron_ore", nil
	}
	panic("unknown ore type")
}
