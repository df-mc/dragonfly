package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
)

// EndPortal is the active End portal block.
type EndPortal struct {
	transparent
	empty
}

// Portal returns the destination dimension of the portal.
func (EndPortal) Portal() world.Dimension {
	return world.End
}

// EntityInside marks an entity as being inside an End portal block.
func (EndPortal) EntityInside(pos cube.Pos, _ *world.Tx, e world.Entity) {
	if traveler, ok := e.(interface{ EnterEndPortal(cube.Pos) }); ok {
		traveler.EnterEndPortal(pos)
	}
}

// HasLiquidDrops ...
func (EndPortal) HasLiquidDrops() bool {
	return false
}

// LightEmissionLevel ...
func (EndPortal) LightEmissionLevel() uint8 {
	return 15
}

// EncodeBlock ...
func (EndPortal) EncodeBlock() (string, map[string]any) {
	return "minecraft:end_portal", nil
}
