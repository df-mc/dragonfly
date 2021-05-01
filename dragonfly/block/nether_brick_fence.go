package block

import (
	"github.com/df-mc/dragonfly/dragonfly/block/cube"
	"github.com/df-mc/dragonfly/dragonfly/block/model"
	"github.com/df-mc/dragonfly/dragonfly/item"
	"github.com/df-mc/dragonfly/dragonfly/world"
)

// NetherBrickFence is the nether brick variant of the fence block.
type NetherBrickFence struct {
	transparent
}

// BreakInfo ...
func (n NetherBrickFence) BreakInfo() BreakInfo {
	return BreakInfo{
		Hardness:    2,
		Harvestable: pickaxeHarvestable,
		Effective:   pickaxeEffective,
		Drops:       simpleDrops(item.NewStack(n, 1)),
	}
}

// CanDisplace ...
func (NetherBrickFence) CanDisplace(b world.Liquid) bool {
	_, ok := b.(Water)
	return ok
}

// SideClosed ...
func (NetherBrickFence) SideClosed(cube.Pos, cube.Pos, *world.World) bool {
	return false
}

// Model ...
func (n NetherBrickFence) Model() world.BlockModel {
	return model.Fence{}
}

// EncodeItem ...
func (NetherBrickFence) EncodeItem() (id int32, name string, meta int16) {
	return 113, "minecraft:nether_brick_fence", 0
}

// EncodeBlock ...
func (NetherBrickFence) EncodeBlock() (string, map[string]interface{}) {
	return "minecraft:nether_brick_fence", nil
}
