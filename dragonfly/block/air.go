package block

import (
	"git.jetbrains.space/dragonfly/dragonfly.git/dragonfly/entity/physics"
	"git.jetbrains.space/dragonfly/dragonfly.git/dragonfly/item"
	"git.jetbrains.space/dragonfly/dragonfly.git/dragonfly/world"
)

// Air is the block present in otherwise empty space.
type Air struct{}

// Drops returns an empty slice.
func (Air) Drops() []item.Stack {
	return nil
}

// EncodeItem ...
func (Air) EncodeItem() (id int32, meta int16) {
	return 0, 0
}

// EncodeBlock ...
func (Air) EncodeBlock() (name string, properties map[string]interface{}) {
	return "minecraft:air", nil
}

// AABB returns an empty Axis Aligned Bounding Box (as nothing can collide with air).
func (Air) AABB() []physics.AABB {
	return nil
}

// ReplaceableBy always returns true.
func (Air) ReplaceableBy(b world.Block) bool {
	return true
}
