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

// PortalInterior returns true if air may occupy the inside of a portal frame before activation for the target dimension.
func (Air) PortalInterior(target world.Dimension) bool {
	return target == world.Nether
}

// EncodeItem ...
func (Air) EncodeItem() (name string, meta int16) {
	return "minecraft:air", 0
}

// EncodeBlock ...
func (Air) EncodeBlock() (string, map[string]any) {
	return "minecraft:air", nil
}
