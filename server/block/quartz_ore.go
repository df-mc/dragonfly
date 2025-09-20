package block

import "github.com/df-mc/dragonfly/server/item"

// NetherQuartzOre is ore found in the Nether.
type NetherQuartzOre struct {
	solid
	bassDrum
}

func (q NetherQuartzOre) BreakInfo() BreakInfo {
	return newBreakInfo(3, pickaxeHarvestable, pickaxeEffective, silkTouchOneOf(item.NetherQuartz{}, q)).withXPDropRange(0, 3)
}

func (NetherQuartzOre) SmeltInfo() item.SmeltInfo {
	return newOreSmeltInfo(item.NewStack(item.NetherQuartz{}, 1), 0.2)
}

func (NetherQuartzOre) EncodeItem() (name string, meta int16) {
	return "minecraft:quartz_ore", 0
}

func (NetherQuartzOre) EncodeBlock() (string, map[string]any) {
	return "minecraft:quartz_ore", nil
}
