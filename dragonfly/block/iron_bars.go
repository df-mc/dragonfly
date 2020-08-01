package block

import (
	"github.com/df-mc/dragonfly/dragonfly/item"
)

// IronBars are blocks that serve a similar purpose to glass panes, but made of iron instead of glass.
type IronBars struct {
	noNBT
	transparent
	thin
}

// BreakInfo ...
func (i IronBars) BreakInfo() BreakInfo {
	return BreakInfo{
		Hardness:    5,
		Harvestable: pickaxeHarvestable,
		Effective:   pickaxeEffective,
		Drops:       simpleDrops(item.NewStack(i, 1)),
	}
}

// EncodeItem ...
func (IronBars) EncodeItem() (id int32, meta int16) {
	return 101, 0
}

// EncodeBlock ...
func (i IronBars) EncodeBlock() (name string, properties map[string]interface{}) {
	return "minecraft:iron_bars", nil
}

// Hash ...
func (i IronBars) Hash() uint64 {
	return hashIronBars
}
