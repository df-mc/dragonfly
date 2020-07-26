package block

import "github.com/df-mc/dragonfly/dragonfly/item"

// Endstone is a block found in The End
type Endstone struct {
	noNBT
}

// BreakInfo ...
func (e Endstone) BreakInfo() BreakInfo {
	return BreakInfo{
		Hardness:    3,
		Harvestable: pickaxeHarvestable,
		Effective:   pickaxeEffective,
		Drops:       simpleDrops(item.NewStack(e, 1)),
	}
}

// EncodeItem ...
func (e Endstone) EncodeItem() (id int32, meta int16) {
	return 121, 0
}

// EncodeBlock ...
func (e Endstone) EncodeBlock() (name string, properties map[string]interface{}) {
	return "minecraft:end_stone", nil
}

// Hash ...
func (e Endstone) Hash() uint64 {
	return hashEndstone
}
