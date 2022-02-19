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
	valid := func(pos cube.Pos) bool {
		b := w.Block(pos)
		_, isPortal := b.(Portal)
		_, isFrame := b.(Obsidian)
		return isPortal || isFrame
	}

	shouldKeep := true
	if pos.Y() < w.Range().Max()-1 {
		shouldKeep = shouldKeep && valid(pos.Add(cube.Pos{0, 1, 0}))
	}
	if pos.Y() > w.Range().Min() {
		shouldKeep = shouldKeep && valid(pos.Subtract(cube.Pos{0, 1, 0}))
	}

	if p.Axis == cube.X {
		shouldKeep = shouldKeep && valid(pos.Subtract(cube.Pos{1, 0, 0}))
		shouldKeep = shouldKeep && valid(pos.Add(cube.Pos{1, 0, 0}))
	} else {
		shouldKeep = shouldKeep && valid(pos.Subtract(cube.Pos{0, 0, 1}))
		shouldKeep = shouldKeep && valid(pos.Add(cube.Pos{0, 0, 1}))
	}

	if !shouldKeep {
		if n, ok := portal.NetherPortalFromPos(w, pos); ok {
			n.Deactivate()
		}
	}
}
