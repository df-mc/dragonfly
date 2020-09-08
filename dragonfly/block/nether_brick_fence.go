package block

import (
	"github.com/df-mc/dragonfly/dragonfly/block/model"
	"github.com/df-mc/dragonfly/dragonfly/world"
)

// NetherBrickFence is the nether brick variant of the fence block.
type NetherBrickFence struct {
	noNBT
	transparent
}

// EncodeBlock ...
func (n NetherBrickFence) EncodeBlock() (name string, properties map[string]interface{}) {
	return "minecraft:nether_brick_fence", nil
}

// Hash ...
func (n NetherBrickFence) Hash() uint64 {
	return hashNetherBrickFence
}

// Model ...
func (n NetherBrickFence) Model() world.BlockModel {
	return model.Fence{}
}

// EncodeItem ...
func (n NetherBrickFence) EncodeItem() (id int32, meta int16) {
	return 113, 0
}
