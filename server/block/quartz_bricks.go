package block

import "github.com/df-mc/dragonfly/server/item"

// QuartzBricks is a mineral block used only for decoration.
type QuartzBricks struct {
	solid
	bassDrum
}

// BreakInfo ...
func (q QuartzBricks) BreakInfo() BreakInfo {
	return BreakInfo{
		Hardness:    0.8,
		Harvestable: pickaxeHarvestable,
		Effective:   pickaxeEffective,
		Drops:       simpleDrops(item.NewStack(q, 1)),
	}
}

// EncodeItem ...
func (QuartzBricks) EncodeItem() (id int32, name string, meta int16) {
	return -304, "minecraft:quartz_bricks", 0
}

// EncodeBlock ...
func (QuartzBricks) EncodeBlock() (string, map[string]interface{}) {
	return "minecraft:quartz_bricks", nil
}
