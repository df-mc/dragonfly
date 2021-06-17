package block

import (
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/tool"
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
	// TODO: Silk touch.
	i := newBreakInfo(d.Type.Hardness(), func(t tool.Tool) bool {
		return t.ToolType() == tool.TypePickaxe && t.HarvestLevel() >= tool.TierIron.HarvestLevel
	}, pickaxeEffective, oneOf(item.Diamond{}))
	i.XPDrops = XPDropRange{3, 7}
	return i
}

// EncodeItem ...
func (d DiamondOre) EncodeItem() (name string, meta int16) {
	return "minecraft:" + d.Type.Prefix() + "diamond_ore", 0
}

// EncodeBlock ...
func (d DiamondOre) EncodeBlock() (string, map[string]interface{}) {
	return "minecraft:" + d.Type.Prefix() + "diamond_ore", nil
}
