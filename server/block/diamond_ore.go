package block

import (
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/tool"
)

// DiamondOre is a rare ore that generates underground.
type DiamondOre struct {
	solid
	bassDrum
}

// BreakInfo ...
func (d DiamondOre) BreakInfo() BreakInfo {
	// TODO: Silk touch.
	i := newBreakInfo(3, func(t tool.Tool) bool {
		return t.ToolType() == tool.TypePickaxe && t.HarvestLevel() >= tool.TierIron.HarvestLevel
	}, pickaxeEffective, oneOf(item.Diamond{}))
	i.XPDrops = XPDropRange{3, 7}
	return i
}

// EncodeItem ...
func (DiamondOre) EncodeItem() (name string, meta int16) {
	return "minecraft:diamond_ore", 0
}

// EncodeBlock ...
func (DiamondOre) EncodeBlock() (string, map[string]interface{}) {
	return "minecraft:diamond_ore", nil
}
