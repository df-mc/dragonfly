package block

import "github.com/df-mc/dragonfly/server/item"

// EndStone is a block found in The End.
type EndStone struct {
	solid
	bassDrum
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
func (EndStone) EncodeItem() (name string, meta int16) {
	return "minecraft:end_stone", 0
}

// EncodeBlock ...
func (EndStone) EncodeBlock() (string, map[string]interface{}) {
	return "minecraft:end_stone", nil
}
