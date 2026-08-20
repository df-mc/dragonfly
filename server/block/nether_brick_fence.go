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

	// Connections holds the sides that the fence connects to.
	Connections Connections
}

// NeighbourUpdateTick ...
func (n NetherBrickFence) NeighbourUpdateTick(pos, _ cube.Pos, tx *world.Tx) {
	if connections := calculateConnections(n.Model().(connector), tx, pos); connections != n.Connections {
		n.Connections = connections
		tx.SetBlock(pos, n, nil)
	}
}

// BreakInfo ...
func (n NetherBrickFence) BreakInfo() BreakInfo {
	return newBreakInfo(2, pickaxeHarvestable, pickaxeEffective, oneOf(n)).withBlastResistance(6)
}

// SideClosed ...
func (NetherBrickFence) SideClosed(cube.Pos, cube.Pos, *world.Tx) bool {
	return false
}

// Model ...
func (n NetherBrickFence) Model() world.BlockModel {
	return model.Fence{}
}

// EncodeItem ...
func (NetherBrickFence) EncodeItem() (name string, meta int16) {
	return "minecraft:nether_brick_fence", 0
}

// EncodeBlock ...
func (n NetherBrickFence) EncodeBlock() (string, map[string]any) {
	return "minecraft:nether_brick_fence", n.Connections.properties()
}

// allNetherBrickFence ...
func allNetherBrickFence() (fence []world.Block) {
	for _, c := range allConnections() {
		fence = append(fence, NetherBrickFence{Connections: c})
	}
	return
}
