package block

import "github.com/df-mc/dragonfly/dragonfly/item"

// EndBricks is a block made from combining four endstone blocks together
type EndBricks struct {
	noNBT
	solid
}

// BreakInfo ...
func (c EndBricks) BreakInfo() BreakInfo {
	return BreakInfo{
		Hardness:    0.8,
		Harvestable: pickaxeHarvestable,
		Effective:   pickaxeEffective,
		Drops:       simpleDrops(item.NewStack(c, 1)),
	}
}

// EncodeItem ...
func (c EndBricks) EncodeItem() (id int32, meta int16) {
	return 206, 0
}

// EncodeBlock ...
func (c EndBricks) EncodeBlock() (name string, properties map[string]interface{}) {
	return "minecraft:end_bricks", nil
}

// Hash ...
func (c EndBricks) Hash() uint64 {
	return hashEndBricks
}
