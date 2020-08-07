package block

import "github.com/df-mc/dragonfly/dragonfly/item"

// Brick can be obtained while crafting brick item
type Brick struct {
	solid
	noNBT
}

// BreakInfo ...
func (b Brick) BreakInfo() BreakInfo {
	return BreakInfo{
		Hardness:    2,
		Harvestable: pickaxeHarvestable,
		Effective:   pickaxeEffective,
		Drops:       simpleDrops(item.NewStack(item.Brick{}, 1)),
	}
}

// EncodeItem ...
func (b Brick) EncodeItem() (id int32, meta int16) {
	return 45, 0
}

// EncodeBlock ...
func (b Brick) EncodeBlock() (name string, properties map[string]interface{}) {
	return "minecraft:brick", nil
}

// Hash ...
func (b Brick) Hash() uint64 {
	return hashBrick
}
