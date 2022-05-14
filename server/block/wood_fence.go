package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/model"
	"github.com/df-mc/dragonfly/server/world"
)

// WoodFence are blocks similar to Walls, which cannot normally be jumped over. Unlike walls however,
// they allow the player (but not mobs) to see through them, making for excellent barriers.
type WoodFence struct {
	transparent
	bass

	// Wood is the type of wood of the fence. This field must have one of the values found in the wood
	// package.
	Wood WoodType
}

// BreakInfo ...
func (w WoodFence) BreakInfo() BreakInfo {
	return newBreakInfo(2, alwaysHarvestable, axeEffective, oneOf(w), XPDropRange{})
}

// CanDisplace ...
func (WoodFence) CanDisplace(b world.Liquid) bool {
	_, ok := b.(Water)
	return ok
}

// SideClosed ...
func (WoodFence) SideClosed(cube.Pos, cube.Pos, *world.World) bool {
	return false
}

// FlammabilityInfo ...
func (w WoodFence) FlammabilityInfo() FlammabilityInfo {
	if !w.Wood.Flammable() {
		return newFlammabilityInfo(0, 0, false)
	}
	return newFlammabilityInfo(5, 20, true)
}

// EncodeBlock ...
func (w WoodFence) EncodeBlock() (name string, properties map[string]interface{}) {
	if w.Wood == CrimsonWood() || w.Wood == WarpedWood() {
		return "minecraft:" + w.Wood.String() + "_fence", nil
	}
	return "minecraft:fence", map[string]interface{}{"wood_type": w.Wood.String()}
}

// Model ...
func (w WoodFence) Model() world.BlockModel {
	return model.Fence{Wooden: true}
}

// EncodeItem ...
func (w WoodFence) EncodeItem() (name string, meta int16) {
	switch w.Wood {
	case CrimsonWood():
		return "minecraft:crimson_fence", 0
	case WarpedWood():
		return "minecraft:warped_fence", 0
	default:
		return "minecraft:fence", int16(w.Wood.Uint8())
	}
}

// allFence ...
func allFence() (fence []world.Block) {
	for _, w := range WoodTypes() {
		fence = append(fence, WoodFence{Wood: w})
	}
	return
}
