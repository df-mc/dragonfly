package block

import (
	"github.com/df-mc/dragonfly/server/item"
)

// DiamondOre is a rare ore that generates underground.
type DiamondOre struct {
	solid
	bassDrum

	// Type is the type of diamond ore.
	Type OreType
}

// BreakInfo ...
func (d DiamondOre) BreakInfo() BreakInfo {
	return newBreakInfo(d.Type.Hardness(), func(t item.Tool) bool {
		return t.ToolType() == item.TypePickaxe && t.HarvestLevel() >= item.ToolTierIron.HarvestLevel
	}, pickaxeEffective, silkTouchOneOf(item.Diamond{}, d)).withXPDropRange(3, 7)
}

// EncodeItem ...
func (d DiamondOre) EncodeItem() (name string, meta int16) {
	return "minecraft:" + d.Type.Prefix() + "diamond_ore", 0
}

// EncodeBlock ...
func (d DiamondOre) EncodeBlock() (string, map[string]any) {
	return "minecraft:" + d.Type.Prefix() + "diamond_ore", nil
}
