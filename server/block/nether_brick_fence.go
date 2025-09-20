package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/model"
	"github.com/df-mc/dragonfly/server/world"
)

// NetherBrickFence is the nether brick variant of the fence block.
type NetherBrickFence struct {
	transparent
	sourceWaterDisplacer
}

func (n NetherBrickFence) BreakInfo() BreakInfo {
	return newBreakInfo(2, pickaxeHarvestable, pickaxeEffective, oneOf(n)).withBlastResistance(30)
}

func (NetherBrickFence) SideClosed(cube.Pos, cube.Pos, *world.Tx) bool {
	return false
}

func (n NetherBrickFence) Model() world.BlockModel {
	return model.Fence{}
}

func (NetherBrickFence) EncodeItem() (name string, meta int16) {
	return "minecraft:nether_brick_fence", 0
}

func (NetherBrickFence) EncodeBlock() (string, map[string]any) {
	return "minecraft:nether_brick_fence", nil
}
