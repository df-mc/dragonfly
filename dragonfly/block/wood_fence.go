package block

import (
	"github.com/df-mc/dragonfly/dragonfly/block/model"
	"github.com/df-mc/dragonfly/dragonfly/block/wood"
	"github.com/df-mc/dragonfly/dragonfly/item"
	"github.com/df-mc/dragonfly/dragonfly/world"
)

// WoodFence are blocks similar to Walls, which cannot normally be jumped over. Unlike walls however,
// they allow the player (but not mobs) to see through them, making for excellent barriers.
type WoodFence struct {
	noNBT
	transparent
	bass

	// Wood is the type of wood of the fence. This field must have one of the values found in the wood
	// package.
	Wood wood.Wood
}

// BreakInfo ...
func (w WoodFence) BreakInfo() BreakInfo {
	return BreakInfo{
		Hardness:    2,
		Harvestable: alwaysHarvestable,
		Effective:   axeEffective,
		Drops:       simpleDrops(item.NewStack(w, 1)),
	}
}

// CanDisplace ...
func (WoodFence) CanDisplace(b world.Liquid) bool {
	_, ok := b.(Water)
	return ok
}

// SideClosed ...
func (WoodFence) SideClosed(world.BlockPos, world.BlockPos, *world.World) bool {
	return false
}

// FlammabilityInfo ...
func (w WoodFence) FlammabilityInfo() FlammabilityInfo {
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
func (w WoodFence) EncodeBlock() (name string, properties map[string]interface{}) {
	if w.Wood == wood.Crimson() || w.Wood == wood.Warped() {
		return "minecraft:" + w.Wood.String() + "_fence", nil
	}
	return "minecraft:fence", map[string]interface{}{"wood_type": w.Wood.String()}
}

// Hash ...
func (w WoodFence) Hash() uint64 {
	return hashWoodFence | (uint64(w.Wood.Uint8()) << 32)
}

// Model ...
func (w WoodFence) Model() world.BlockModel {
	return model.Fence{Wooden: true}
}

// EncodeItem ...
func (w WoodFence) EncodeItem() (id int32, meta int16) {
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
func allFence() (fence []WoodFence) {
	for _, w := range wood.All() {
		fence = append(fence, WoodFence{Wood: w})
	}
	return
}
