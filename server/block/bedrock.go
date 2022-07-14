package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world/explosion"
)

// Bedrock is a block that is indestructible in survival.
type Bedrock struct {
	solid
	transparent
	bassDrum

	// InfiniteBurning specifies if the bedrock block is set aflame and will burn forever. This is the case
	// for bedrock found under end crystals on top of the end pillars.
	InfiniteBurning bool
}

// BlastResistance ...
func (Bedrock) BlastResistance() float64 {
	return 3600000
}

// Explode ...
func (b Bedrock) Explode(cube.Pos, explosion.Config) {}

// EncodeItem ...
func (Bedrock) EncodeItem() (name string, meta int16) {
	return "minecraft:bedrock", 0
}

// EncodeBlock ...
func (b Bedrock) EncodeBlock() (name string, properties map[string]any) {
	//noinspection SpellCheckingInspection
	return "minecraft:bedrock", map[string]any{"infiniburn_bit": b.InfiniteBurning}
}
