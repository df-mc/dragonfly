package block

import (
	"github.com/df-mc/dragonfly/dragonfly/block/coral"
	"github.com/df-mc/dragonfly/dragonfly/item"
	"github.com/df-mc/dragonfly/dragonfly/world"
	"time"
)

// CoralBlock is a solid block that comes in 5 variants.
type CoralBlock struct {
	noNBT
	solid
	bassDrum

	// Type is the type of coral of the block.
	Type coral.Coral
	// Dead is whether the coral block is dead.
	Dead bool
}

// NeighbourUpdateTick ...
func (c CoralBlock) NeighbourUpdateTick(pos, changedNeighbour world.BlockPos, w *world.World) {
	if c.Dead {
		return
	}
	w.ScheduleBlockUpdate(pos, time.Second*5/2)
}

// ScheduledTick ...
func (c CoralBlock) ScheduledTick(pos world.BlockPos, w *world.World) {
	if c.Dead {
		return
	}

	adjacentWater := false
	pos.Neighbours(func(neighbour world.BlockPos) {
		if liquid, ok := w.Liquid(neighbour); ok {
			if _, ok := liquid.(Water); ok {
				adjacentWater = true
			}
		}
	})
	if !adjacentWater {
		c.Dead = true
		w.PlaceBlock(pos, c)
	}
}

// BreakInfo ...
func (c CoralBlock) BreakInfo() BreakInfo {
	return BreakInfo{
		Hardness:    1.5,
		Harvestable: alwaysHarvestable,
		Effective:   pickaxeEffective,
		Drops:       simpleDrops(item.NewStack(CoralBlock{Type: c.Type, Dead: true}, 1)), //TODO: Not dead coral blocks should drop itself if mined with silk touch
	}
}

// EncodeBlock ...
func (c CoralBlock) EncodeBlock() (name string, properties map[string]interface{}) {
	return "minecraft:coral_block", map[string]interface{}{"coral_color": c.Type.Colour.String(), "dead_bit": c.Dead}
}

// Hash ...
func (c CoralBlock) Hash() uint64 {
	return hashCoralBlock | (uint64(boolByte(c.Dead)) << 32) | (uint64(c.Type.Uint8()) << 33)
}

// EncodeItem ...
func (c CoralBlock) EncodeItem() (id int32, meta int16) {
	if c.Dead {
		return -132, int16(c.Type.Uint8() | 8)
	}
	return -132, int16(c.Type.Uint8())
}

// allCoralBlocks returns a list of all coral block variants
func allCoralBlocks() (c []world.Block) {
	f := func(dead bool) {
		c = append(c, CoralBlock{Type: coral.Tube(), Dead: dead})
		c = append(c, CoralBlock{Type: coral.Brain(), Dead: dead})
		c = append(c, CoralBlock{Type: coral.Bubble(), Dead: dead})
		c = append(c, CoralBlock{Type: coral.Fire(), Dead: dead})
		c = append(c, CoralBlock{Type: coral.Horn(), Dead: dead})
	}
	f(true)
	f(false)
	return
}
