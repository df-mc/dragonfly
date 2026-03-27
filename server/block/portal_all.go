package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
)

func allPortals() (b []world.Block) {
	for _, axis := range []cube.Axis{cube.X, cube.Z} {
		b = append(b, Portal{Axis: axis})
	}
	return
}
