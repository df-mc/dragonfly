package block

import (
	"github.com/df-mc/dragonfly/server/world"
)

// Air is the block present in otherwise empty space.
type Air struct {
	empty
	replaceable
	transparent
}

// CanDisplace ...
func (Air) CanDisplace(world.Liquid) bool {
	return true
}

// HasLiquidDrops ...
func (Air) HasLiquidDrops() bool {
	return false
}

// EncodeItem ...
func (Air) EncodeItem() (name string, meta int16) {
	return "minecraft:air", 0
}

// EncodeBlock ...
func (Air) EncodeBlock() (string, map[string]interface{}) {
	return "minecraft:air", nil
}
