package block

import (
	"github.com/df-mc/dragonfly/dragonfly/item"
)

// Bricks are decorative building blocks.
type Bricks struct {
	solid
	bassDrum
}

// BreakInfo ...
func (b Bricks) BreakInfo() BreakInfo {
	return BreakInfo{
		Hardness:    2,
		Harvestable: pickaxeHarvestable,
		Effective:   pickaxeEffective,
		Drops:       simpleDrops(item.NewStack(b, 1)),
	}
}

// EncodeItem ...
func (Bricks) EncodeItem() (id int32, meta int16) {
	return 45, 0
}

// EncodeBlock ...
func (Bricks) EncodeBlock() (string, map[string]interface{}) {
	return "minecraft:brick_block", nil
}
