package block

import (
	"github.com/df-mc/dragonfly/dragonfly/item"
)

// Bricks are decorative building blocks.
type Bricks struct {
	noNBT
	solid
	bassdrum
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
func (b Bricks) EncodeBlock() (name string, properties map[string]interface{}) {
	return "minecraft:brick_block", nil
}

// Hash ...
func (b Bricks) Hash() uint64 {
	return hashBricks
}
