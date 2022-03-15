package block

import "github.com/df-mc/dragonfly/server/world"

// Light is an invisible block that can produce any light level.
type Light struct {
	empty
	replaceable
	transparent

	// Level is the light level that the light block produces. It is a number from 0-15, where 15 is the
	// brightest and 0 is no light at all.
	Level int
}

// EncodeItem ...
func (l Light) EncodeItem() (name string, meta int16) {
	return "minecraft:light_block", int16(l.Level)
}

// LightEmissionLevel ...
func (l Light) LightEmissionLevel() uint8 {
	return uint8(l.Level)
}

// EncodeBlock ...
func (l Light) EncodeBlock() (name string, properties map[string]any) {
	return "minecraft:light_block", map[string]any{"block_light_level": int32(l.Level)}
}

// allLight returns all possible light blocks.
func allLight() []world.Block {
	m := make([]world.Block, 0, 16)
	for i := 0; i < 16; i++ {
		m = append(m, Light{Level: i})
	}
	return m
}
