package block

import "git.jetbrains.space/dragonfly/dragonfly.git/dragonfly/world"

// Light is an invisible block that can produce any light level.
type Light struct {
	// Level is the light level that the light block produces. It is a number from 0-15, where 15 is the
	// brightest and 0 is no light at all.
	Level uint8
}

// EncodeItem ...
func (l Light) EncodeItem() (id int32, meta int16) {
	return -215, int16(l.Level)
}

// LightEmissionLevel ...
func (l Light) LightEmissionLevel() uint8 {
	return l.Level
}

// LightDiffusionLevel ...
func (l Light) LightDiffusionLevel() uint8 {
	return 0
}

// EncodeBlock ...
func (l Light) EncodeBlock() (name string, properties map[string]interface{}) {
	return "minecraft:light_block", map[string]interface{}{"block_light_level": l.Level}
}

// allLight returns all possible light blocks.
func allLight() []world.Block {
	m := make([]world.Block, 0, 16)
	for i := uint8(0); i < 16; i++ {
		m = append(m, Light{Level: i})
	}
	return m
}
