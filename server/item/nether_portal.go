package item

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	worldportal "github.com/df-mc/dragonfly/server/world/portal"
)

func activateNetherPortalAt(tx *world.Tx, pos cube.Pos) bool {
	n, ok := worldportal.NetherPortalFromPos(tx, pos)
	if !ok || !n.Framed() {
		return false
	}
	n.Activate()
	return true
}
