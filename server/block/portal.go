package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/model"
	"github.com/df-mc/dragonfly/server/world"
	worldportal "github.com/df-mc/dragonfly/server/world/portal"
)

// Portal is the active Nether portal block.
type Portal struct {
	transparent

	// Axis is the axis normal to the portal plane.
	Axis cube.Axis
}

// Model returns the Nether portal model.
func (p Portal) Model() world.BlockModel {
	return model.Portal{Axis: p.Axis}
}

// Portal returns the destination dimension of the portal.
func (Portal) Portal() world.Dimension {
	return world.Nether
}

// EntityInside marks an entity as being inside a Nether portal block.
func (p Portal) EntityInside(pos cube.Pos, _ *world.Tx, e world.Entity) {
	if traveler, ok := e.(interface{ EnterNetherPortal(cube.Pos, cube.Axis) }); ok {
		traveler.EnterNetherPortal(pos, p.Axis)
	}
}

// HasLiquidDrops ...
func (Portal) HasLiquidDrops() bool {
	return false
}

// EncodeBlock ...
func (p Portal) EncodeBlock() (string, map[string]any) {
	return "minecraft:portal", map[string]any{"portal_axis": p.Axis.String()}
}

// NeighbourUpdateTick deactivates portal blocks if their frame is broken.
func (p Portal) NeighbourUpdateTick(pos, _ cube.Pos, tx *world.Tx) {
	if n, ok := worldportal.NetherPortalFromPos(tx, pos); ok {
		if !n.Framed() || !n.Activated() {
			n.Deactivate()
		}
		return
	}
	tx.SetBlock(pos, nil, nil)
}
