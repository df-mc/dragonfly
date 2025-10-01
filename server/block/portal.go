package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/model"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/portal"
)

// Portal is the translucent part of the nether portal that teleports the player to and from the Nether.
type Portal struct {
	transparent

	// Axis is the axis which the portal faces.
	Axis cube.Axis
}

// Model ...
func (p Portal) Model() world.BlockModel {
	return model.Portal{Axis: p.Axis}
}

// Portal ...
func (Portal) Portal() world.Dimension {
	return world.Nether
}

// HasLiquidDrops ...
func (p Portal) HasLiquidDrops() bool {
	return false
}

// EncodeBlock ...
func (p Portal) EncodeBlock() (string, map[string]interface{}) {
	return "minecraft:portal", map[string]interface{}{"portal_axis": p.Axis.String()}
}

// NeighbourUpdateTick ...
func (p Portal) NeighbourUpdateTick(pos, _ cube.Pos, tx *world.Tx) {
	if n, ok := portal.NetherPortalFromPos(tx, pos); ok && (!n.Framed() || !n.Activated()) {
		n.Deactivate()
	}
}
