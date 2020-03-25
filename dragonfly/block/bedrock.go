package block

import "git.jetbrains.space/dragonfly/dragonfly.git/dragonfly/item"

// Bedrock is a block that is indestructible in survival.
type Bedrock struct {
	// InfiniteBurning specifies if the bedrock block is set aflame and will burn forever. This is the case
	// for bedrock found under end crystals on top of the end pillars.
	InfiniteBurning bool
}

// Drops returns an empty slice.
func (Bedrock) Drops() []item.Stack {
	return nil
}

// EncodeItem ...
func (Bedrock) EncodeItem() (id int32, meta int16) {
	return 7, 0
}

// EncodeBlock ...
func (b Bedrock) EncodeBlock() (name string, properties map[string]interface{}) {
	//noinspection SpellCheckingInspection
	return "minecraft:bedrock", map[string]interface{}{"infiniburn_bit": b.InfiniteBurning}
}
