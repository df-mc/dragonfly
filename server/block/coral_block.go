package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"math/rand"
	"time"
)

// CoralBlock is a solid block that comes in 5 variants.
type CoralBlock struct {
	solid
	bassDrum

	// Type is the type of coral of the block.
	Type CoralType
	// Dead is whether the coral block is dead.
	Dead bool
}

// NeighbourUpdateTick ...
func (c CoralBlock) NeighbourUpdateTick(pos, _ cube.Pos, w *world.World) {
	if c.Dead {
		return
	}
	w.ScheduleBlockUpdate(pos, time.Second*5/2)
}

// ScheduledTick ...
func (c CoralBlock) ScheduledTick(pos cube.Pos, w *world.World, _ *rand.Rand) {
	if c.Dead {
		return
	}

	adjacentWater := false
	pos.Neighbours(func(neighbour cube.Pos) {
		if liquid, ok := w.Liquid(neighbour); ok {
			if _, ok := liquid.(Water); ok {
				adjacentWater = true
			}
		}
	}, w.Range())
	if !adjacentWater {
		c.Dead = true
		w.SetBlock(pos, c, nil)
	}
}

// BreakInfo ...
func (c CoralBlock) BreakInfo() BreakInfo {
	return newBreakInfo(7, pickaxeHarvestable, pickaxeEffective, silkTouchOneOf(CoralBlock{Type: c.Type, Dead: true}, c))
}

// EncodeBlock ...
func (c CoralBlock) EncodeBlock() (name string, properties map[string]any) {
	return "minecraft:coral_block", map[string]any{"coral_color": c.Type.Colour().String(), "dead_bit": c.Dead}
}

// EncodeItem ...
func (c CoralBlock) EncodeItem() (name string, meta int16) {
	if c.Dead {
		return "minecraft:coral_block", int16(c.Type.Uint8() | 8)
	}
	return "minecraft:coral_block", int16(c.Type.Uint8())
}

// allCoralBlocks returns a list of all coral block variants
func allCoralBlocks() (c []world.Block) {
	f := func(dead bool) {
		for _, t := range CoralTypes() {
			c = append(c, CoralBlock{Type: t, Dead: dead})
		}
	}
	f(true)
	f(false)
	return
}
