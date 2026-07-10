package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/model"
	"github.com/df-mc/dragonfly/server/world"
)

// EndPortal is the translucent block that teleports the player to and from the End. It is created by inserting an Eye
// of Ender into all twelve frame blocks of an End portal frame ring.
type EndPortal struct {
	transparent
}

// Model ...
func (EndPortal) Model() world.BlockModel {
	return model.Empty{}
}

// LightEmissionLevel returns 15.
func (EndPortal) LightEmissionLevel() uint8 {
	return 15
}

// HasLiquidDrops ...
func (EndPortal) HasLiquidDrops() bool {
	return false
}

// Portal returns the End dimension. The same block leads back to the Overworld when entered from the End.
func (EndPortal) Portal() world.Dimension {
	return world.End
}

// EntityInside ...
func (EndPortal) EntityInside(_ cube.Pos, tx *world.Tx, e world.Entity) {
	if t, ok := e.(portalTraveller); ok {
		t.TravelThroughPortal(tx, world.End)
	}
}

// EncodeBlock ...
func (EndPortal) EncodeBlock() (string, map[string]any) {
	return "minecraft:end_portal", nil
}
