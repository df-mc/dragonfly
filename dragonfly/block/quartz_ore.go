package block

import "github.com/df-mc/dragonfly/dragonfly/item"

// Nether quartz ore is ore found in the Nether.
type QuartzOre struct {
	noNBT
	solid
}

// BreakInfo ...
func (q QuartzOre) BreakInfo() BreakInfo {
	return BreakInfo{
		Hardness:    3,
		Harvestable: pickaxeHarvestable,
		Effective:   pickaxeEffective,
		Drops:       simpleDrops(item.NewStack(item.NetherQuartz{}, 1)),
		XPDrops:     XPDropRange{0, 3},
	}
}

// EncodeItem ...
func (QuartzOre) EncodeItem() (id int32, meta int16) {
	return 153, 0
}

// EncodeBlock ...
func (QuartzOre) EncodeBlock() (name string, properties map[string]interface{}) {
	return "miencraft:quartz_ore", nil
}

// Hash ...
func (QuartzOre) Hash() uint64 {
	return hashQuartzOre
}
