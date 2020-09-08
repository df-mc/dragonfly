package block

import (
	"github.com/df-mc/dragonfly/dragonfly/block/model"
	"github.com/df-mc/dragonfly/dragonfly/block/wood"
	"github.com/df-mc/dragonfly/dragonfly/world"
)

// WoodenFence are blocks similar to Walls, which cannot normally be jumped over. Unlike walls however,
// they allow the player (but not mobs) to see through them, making for excellent barriers.
type WoodenFence struct {
	noNBT
	transparent

	// Wood is the type of wood of the fence. This field must have one of the values found in the material
	// package.
	Wood wood.Wood
}

// FlammabilityInfo ...
func (w WoodenFence) FlammabilityInfo() FlammabilityInfo {
	if w.Wood.Flammable() {
		return FlammabilityInfo{}
	}
	return FlammabilityInfo{
		Encouragement: 5,
		Flammability:  20,
		LavaFlammable: true,
	}
}

// EncodeBlock ...
func (w WoodenFence) EncodeBlock() (name string, properties map[string]interface{}) {
	if w.Wood == wood.Crimson() || w.Wood == wood.Warped() {
		return "minecraft:" + w.Wood.String() + "_fence", nil
	}
	return "minecraft:fence", map[string]interface{}{"wood_type": w.Wood.String()}
}

// Hash ...
func (w WoodenFence) Hash() uint64 {
	return hashWoodFence | (uint64(w.Wood.Uint8()) << 32)
}

// Model ...
func (w WoodenFence) Model() world.BlockModel {
	return model.Fence{Wooden: true}
}

// EncodeItem ...
func (w WoodenFence) EncodeItem() (id int32, meta int16) {
	switch w.Wood {
	case wood.Crimson():
		return -256, 0
	case wood.Warped():
		return -257, 0
	default:
		return 85, int16(w.Wood.Uint8())
	}
}

// allFence ...
func allFence() (fence []world.Block) {
	for _, w := range wood.All() {
		fence = append(fence, WoodenFence{Wood: w})
	}
	return
}
