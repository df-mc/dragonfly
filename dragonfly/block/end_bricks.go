package block

import "github.com/df-mc/dragonfly/dragonfly/item"

// EndBricks is a block made from combining four endstone blocks together.
type EndBricks struct {
	noNBT
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
func (c EndBricks) EncodeItem() (id int32, meta int16) {
	return 206, 0
}
