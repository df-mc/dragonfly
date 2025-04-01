package block

import (
	"github.com/df-mc/dragonfly/server/item"
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
	return newBreakInfo(i.Type.Hardness(), func(t item.Tool) bool {
		return t.ToolType() == item.TypePickaxe && t.HarvestLevel() >= item.ToolTierStone.HarvestLevel
	}, pickaxeEffective, silkTouchOneOf(item.RawIron{}, i)).withBlastResistance(15)
}

// SmeltInfo ...
func (IronOre) SmeltInfo() item.SmeltInfo {
	return newOreSmeltInfo(item.NewStack(item.IronIngot{}, 1), 0.7)
}

// EncodeItem ...
func (i IronOre) EncodeItem() (name string, meta int16) {
	return "minecraft:" + i.Type.Prefix() + "iron_ore", 0
}

// EncodeBlock ...
func (i IronOre) EncodeBlock() (string, map[string]any) {
	return "minecraft:" + i.Type.Prefix() + "iron_ore", nil
}
