package block

import "github.com/df-mc/dragonfly/dragonfly/item"

// NetherQuartzOre is ore found in the Nether.
type NetherQuartzOre struct {
	noNBT
	solid
}

// BreakInfo ...
func (q NetherQuartzOre) BreakInfo() BreakInfo {
	return BreakInfo{
		Hardness:    3,
		Harvestable: pickaxeHarvestable,
		Effective:   pickaxeEffective,
		Drops:       simpleDrops(item.NewStack(item.NetherQuartz{}, 1)),
		XPDrops:     XPDropRange{0, 3},
	}
}

// EncodeItem ...
func (NetherQuartzOre) EncodeItem() (id int32, meta int16) {
	return 153, 0
}

// EncodeBlock ...
func (NetherQuartzOre) EncodeBlock() (name string, properties map[string]interface{}) {
	return "minecraft:quartz_ore", nil
}

// Hash ...
func (NetherQuartzOre) Hash() uint64 {
	return hashQuartzOre
}
