package block

import "github.com/df-mc/dragonfly/server/item"

// NetherQuartzOre is ore found in the Nether.
type NetherQuartzOre struct {
	solid
	bassDrum
}

// BreakInfo ...
func (q NetherQuartzOre) BreakInfo() BreakInfo {
	return newBreakInfo(3, pickaxeHarvestable, pickaxeEffective, silkTouchOneOf(item.NetherQuartz{}, q)).withXPDropRange(0, 3)
}

// EncodeItem ...
func (NetherQuartzOre) EncodeItem() (name string, meta int16) {
	return "minecraft:quartz_ore", 0
}

// EncodeBlock ...
func (NetherQuartzOre) EncodeBlock() (string, map[string]any) {
	return "minecraft:quartz_ore", nil
}
