package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
)

// IronBars are blocks that serve a similar purpose to glass panes, but made of iron instead of glass.
type IronBars struct {
	transparent
	thin
	sourceWaterDisplacer

	// Connections holds the sides that the iron bars connects to.
	Connections Connections
}

// NeighbourUpdateTick ...
func (i IronBars) NeighbourUpdateTick(pos, _ cube.Pos, tx *world.Tx) {
	if connections := calculateConnections(i.Model().(connector), tx, pos); connections != i.Connections {
		i.Connections = connections
		tx.SetBlock(pos, i, nil)
	}
}

// BreakInfo ...
func (i IronBars) BreakInfo() BreakInfo {
	return newBreakInfo(5, pickaxeHarvestable, pickaxeEffective, oneOf(i)).withBlastResistance(6)
}

// SideClosed ...
func (i IronBars) SideClosed(cube.Pos, cube.Pos, *world.Tx) bool {
	return false
}

// EncodeItem ...
func (IronBars) EncodeItem() (name string, meta int16) {
	return "minecraft:iron_bars", 0
}

// EncodeBlock ...
func (i IronBars) EncodeBlock() (string, map[string]any) {
	return "minecraft:iron_bars", i.Connections.properties()
}

// allIronBars ...
func allIronBars() (bars []world.Block) {
	for _, c := range allConnections() {
		bars = append(bars, IronBars{Connections: c})
	}
	return
}
