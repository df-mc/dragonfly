package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/model"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"time"
)

// WoodFence are blocks similar to Walls, which cannot normally be jumped over. Unlike walls however,
// they allow the player (but not mobs) to see through them, making for excellent barriers.
type WoodFence struct {
	transparent
	bass
	sourceWaterDisplacer

	// Wood is the type of wood of the fence. This field must have one of the values found in the wood
	// package.
	Wood WoodType
	// Connections holds the sides that the fence connects to.
	Connections Connections
}

// NeighbourUpdateTick ...
func (w WoodFence) NeighbourUpdateTick(pos, _ cube.Pos, tx *world.Tx) {
	if connections := calculateConnections(w.Model().(connector), tx, pos); connections != w.Connections {
		w.Connections = connections
		tx.SetBlock(pos, w, nil)
	}
}

// BreakInfo ...
func (w WoodFence) BreakInfo() BreakInfo {
	return newBreakInfo(2, alwaysHarvestable, axeEffective, oneOf(w)).withBlastResistance(3)
}

// SideClosed ...
func (WoodFence) SideClosed(cube.Pos, cube.Pos, *world.Tx) bool {
	return false
}

// FlammabilityInfo ...
func (w WoodFence) FlammabilityInfo() FlammabilityInfo {
	if !w.Wood.Flammable() {
		return newFlammabilityInfo(0, 0, false)
	}
	return newFlammabilityInfo(5, 20, true)
}

// FuelInfo ...
func (w WoodFence) FuelInfo() item.FuelInfo {
	if !w.Wood.Flammable() {
		return item.FuelInfo{}
	}
	return newFuelInfo(time.Second * 15)
}

// EncodeBlock ...
func (w WoodFence) EncodeBlock() (name string, properties map[string]any) {
	return "minecraft:" + w.Wood.String() + "_fence", w.Connections.properties()
}

// Model ...
func (w WoodFence) Model() world.BlockModel {
	return model.Fence{Wood: true}
}

// EncodeItem ...
func (w WoodFence) EncodeItem() (name string, meta int16) {
	return "minecraft:" + w.Wood.String() + "_fence", 0
}

// allFence ...
func allFence() (fence []world.Block) {
	for _, w := range WoodTypes() {
		for _, c := range allConnections() {
			fence = append(fence, WoodFence{Wood: w, Connections: c})
		}
	}
	return
}
