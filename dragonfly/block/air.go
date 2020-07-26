package block

import (
	"github.com/df-mc/dragonfly/dragonfly/world"
)

// Air is the block present in otherwise empty space.
type Air struct {
	noNBT
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
func (Air) EncodeItem() (id int32, meta int16) {
	return 0, 0
}

// EncodeBlock ...
func (Air) EncodeBlock() (name string, properties map[string]interface{}) {
	return "minecraft:air", nil
}

// Hash ...
func (Air) Hash() uint64 {
	return hashAir
}
