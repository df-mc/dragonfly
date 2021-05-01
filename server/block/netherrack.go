package block

import "github.com/df-mc/dragonfly/server/item"

// Netherrack is a block found in The Nether.
type Netherrack struct {
	solid
	bassDrum
}

// BreakInfo ...
func (e Netherrack) BreakInfo() BreakInfo {
	return BreakInfo{
		Hardness:    0.4,
		Harvestable: pickaxeHarvestable,
		Effective:   pickaxeEffective,
		Drops:       simpleDrops(item.NewStack(e, 1)),
	}
}

// EncodeItem ...
func (Netherrack) EncodeItem() (id int32, name string, meta int16) {
	return 87, "minecraft:netherrack", 0
}

// EncodeBlock ...
func (Netherrack) EncodeBlock() (string, map[string]interface{}) {
	return "minecraft:netherrack", nil
}
