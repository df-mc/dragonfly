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

// Portal returns the End. The actual destination world is resolved by World.PortalDestination, which returns the
// Overworld when called from the End — providing the return path through the same block.
func (EndPortal) Portal() world.Dimension {
	return world.End
}

// EntityInside is called for players (and other EntityInsider-aware travellers). Ent travellers go through the
// Portal()-based path in Ent.checkPortalInsiders.
func (EndPortal) EntityInside(_ cube.Pos, tx *world.Tx, e world.Entity) {
	if t, ok := e.(portalTraveller); ok {
		t.TravelThroughPortal(tx, world.End)
	}
}

// EncodeBlock ...
func (EndPortal) EncodeBlock() (string, map[string]any) {
	return "minecraft:end_portal", nil
}
