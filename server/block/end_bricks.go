package block

import "github.com/df-mc/dragonfly/server/item"

// EndBricks is a block made from combining four endstone blocks together.
type EndBricks struct {
	solid
	bassDrum
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
func (EndBricks) EncodeItem() (id int32, name string, meta int16) {
	return 206, "minecraft:end_bricks", 0
}

// EncodeBlock ...
func (EndBricks) EncodeBlock() (string, map[string]interface{}) {
	return "minecraft:end_bricks", nil
}
