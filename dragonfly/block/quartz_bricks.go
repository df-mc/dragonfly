package block

import "github.com/df-mc/dragonfly/dragonfly/item"

// QuartzBricks is a mineral block used only for decoration.
type QuartzBricks struct {
	noNBT
	solid
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
func (QuartzBricks) EncodeItem() (id int32, meta int16) {
	return -304, 0
}

// EncodeBlock ...
func (QuartzBricks) EncodeBlock() (name string, properties map[string]interface{}) {
	return "minecraft:quartz_bricks", nil
}

// Hash ...
func (QuartzBricks) Hash() uint64 {
	return hashQuartzBricks
}
