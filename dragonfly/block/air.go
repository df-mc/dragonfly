package block

import "git.jetbrains.space/dragonfly/dragonfly.git/dragonfly/entity/physics"

// Air is the block present in otherwise empty space.
type Air struct{}

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
