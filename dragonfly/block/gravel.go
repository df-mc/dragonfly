package block

import "github.com/df-mc/dragonfly/dragonfly/item"

// Gravel is a block that is affected by gravity.
type Gravel struct{}

// BreakInfo ...
func (g Gravel) BreakInfo() BreakInfo {
	return BreakInfo{
		Hardness:    0.6,
		Harvestable: alwaysHarvestable,
		Effective:   shovelEffective,
		Drops:       simpleDrops(item.NewStack(g, 1)),
		// TODO: Add Flint and drop it here.
	}
}

// EncodeItem ...
func (g Gravel) EncodeItem() (id int32, meta int16) {
	return 13, 0
}

// EncodeBlock ...
func (g Gravel) EncodeBlock() (name string, properties map[string]interface{}) {
	return "minecraft:gravel", nil
}
