package block

import "github.com/df-mc/dragonfly/dragonfly/item"

// EndStone is a block found in The End.
type EndStone struct {
	noNBT
	solid
}

// BreakInfo ...
func (e EndStone) BreakInfo() BreakInfo {
	return BreakInfo{
		Hardness:    3,
		Harvestable: pickaxeHarvestable,
		Effective:   pickaxeEffective,
		Drops:       simpleDrops(item.NewStack(e, 1)),
	}
}

// EncodeItem ...
func (e EndStone) EncodeItem() (id int32, meta int16) {
	return 121, 0
}

// EncodeBlock ...
func (e EndStone) EncodeBlock() (name string, properties map[string]interface{}) {
	return "minecraft:end_stone", nil
}

// Hash ...
func (e EndStone) Hash() uint64 {
	return hashEndStone
}
