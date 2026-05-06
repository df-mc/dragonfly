package block

import "github.com/df-mc/dragonfly/server/world"

// Air is the block present in otherwise empty space.
type Air struct {
	empty
	replaceable
	transparent
}

// HasLiquidDrops ...
func (Air) HasLiquidDrops() bool {
	return false
}

// PortalInterior returns true if air may occupy the inside of a portal frame before activation.
func (Air) PortalInterior(dimension world.Dimension) bool {
	return dimension == world.Nether
}

// EncodeItem ...
func (Air) EncodeItem() (name string, meta int16) {
	return "minecraft:air", 0
}

// EncodeBlock ...
func (Air) EncodeBlock() (string, map[string]any) {
	return "minecraft:air", nil
}
