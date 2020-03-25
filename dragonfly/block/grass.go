package block

import (
	"git.jetbrains.space/dragonfly/dragonfly.git/dragonfly/item"
)

// Grass blocks generate abundantly across the surface of the world.
type Grass struct{}

// Drops returns a dirt item.
func (g Grass) Drops() []item.Stack {
	return []item.Stack{item.NewStack(Dirt{}, 1)}
}

// EncodeItem ...
func (Grass) EncodeItem() (id int32, meta int16) {
	return 2, 0
}

// EncodeBlock ...
func (Grass) EncodeBlock() (name string, properties map[string]interface{}) {
	return "minecraft:grass", nil
}
