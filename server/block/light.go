package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"strconv"
)

// Light is an invisible block that can produce any light level.
type Light struct {
	empty
	replaceable
	transparent
	flowingWaterDisplacer

	// Level is the light level that the light block produces. It is a number from 0-15, where 15 is the
	// brightest and 0 is no light at all.
	Level int
}

// SideClosed ...
func (Light) SideClosed(cube.Pos, cube.Pos, *world.Tx) bool {
	return false
}

// EncodeItem ...
func (l Light) EncodeItem() (name string, meta int16) {
	return "minecraft:light_block_" + strconv.Itoa(l.Level), 0
}

// LightEmissionLevel ...
func (l Light) LightEmissionLevel() uint8 {
	return uint8(l.Level)
}

// EncodeBlock ...
func (l Light) EncodeBlock() (name string, properties map[string]any) {
	return "minecraft:light_block_" + strconv.Itoa(l.Level), nil
}

// allLight returns all possible light blocks.
func allLight() []world.Block {
	m := make([]world.Block, 0, 16)
	for i := 0; i < 16; i++ {
		m = append(m, Light{Level: i})
	}
	return m
}
