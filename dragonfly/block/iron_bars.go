package block

import (
	"github.com/df-mc/dragonfly/dragonfly/item"
	"github.com/df-mc/dragonfly/dragonfly/world"
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

// CanDisplace ...
func (i IronBars) CanDisplace(b world.Liquid) bool {
	_, water := b.(Water)
	return water
}

// SideClosed ...
func (i IronBars) SideClosed(world.BlockPos, world.BlockPos, *world.World) bool {
	return false
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
