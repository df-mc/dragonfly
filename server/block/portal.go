package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/portal"
)

// Portal is the translucent part of the nether portal that teleports the player to and from the Nether.
type Portal struct {
	empty
	transparent

	// Axis is the axis which the chain faces.
	Axis cube.Axis
}

// EncodeBlock ...
func (p Portal) EncodeBlock() (string, map[string]interface{}) {
	return "minecraft:portal", map[string]interface{}{"portal_axis": p.Axis.String()}
}

// NeighbourUpdateTick ...
func (p Portal) NeighbourUpdateTick(pos, _ cube.Pos, w *world.World) {
	if n, ok := portal.NetherPortalFromPos(w, pos); ok && !n.Framed() {
		n.Deactivate()
	}
}
