package block

import "github.com/df-mc/dragonfly/dragonfly/item"

// DiamondBlock is a block which can only be gained by crafting it.
type DiamondBlock struct{}

// BreakInfo ...
func (d DiamondBlock) BreakInfo() BreakInfo {
	return BreakInfo{
		Hardness:    5,
		Harvestable: pickaxeHarvestable,
		Effective:   pickaxeEffective,
		Drops:       simpleDrops(item.NewStack(d, 1)),
	}
}

// EncodeItem ...
func (d DiamondBlock) EncodeItem() (id int32, meta int16) {
	return 57, 0
}

// EncodeBlock ...
func (d DiamondBlock) EncodeBlock() (name string, properties map[string]interface{}) {
	return "minecraft:diamond_block", nil
}
