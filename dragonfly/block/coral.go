package block

import (
	"github.com/df-mc/dragonfly/dragonfly/block/coral"
	"github.com/df-mc/dragonfly/dragonfly/item"
	"github.com/df-mc/dragonfly/dragonfly/item/tool"
	"github.com/df-mc/dragonfly/dragonfly/world"
	"github.com/go-gl/mathgl/mgl64"
	"time"
)

// Coral is a non solid block that comes in 5 variants.
type Coral struct {
	noNBT
	empty
	transparent

	// Type is the type of coral of the block.
	Type coral.Coral
	// Dead is whether the coral is dead.
	Dead bool
}

// UseOnBlock ...
func (c Coral) UseOnBlock(pos world.BlockPos, face world.Face, clickPos mgl64.Vec3, w *world.World, user item.User, ctx *item.UseContext) bool {
	pos, _, used := firstReplaceable(w, pos, face, c)
	if !used {
		return false
	}
	if !w.Block(pos.Side(world.FaceDown)).Model().FaceSolid(pos.Side(world.FaceDown), world.FaceUp, w) {
		return false
	}
	if liquid, ok := w.Liquid(pos); ok {
		if water, ok := liquid.(Water); ok {
			if water.Depth != 8 {
				return false
			}
		}
	}

	place(w, pos, c, user, ctx)
	return placed(ctx)
}

// HasLiquidDrops ...
func (c Coral) HasLiquidDrops() bool {
	return false
}

// CanDisplace ...
func (c Coral) CanDisplace(b world.Liquid) bool {
	_, water := b.(Water)
	return water
}

// SideClosed ...
func (c Coral) SideClosed(pos, side world.BlockPos, w *world.World) bool {
	return false
}

// NeighbourUpdateTick ...
func (c Coral) NeighbourUpdateTick(pos, changedNeighbour world.BlockPos, w *world.World) {
	if !w.Block(pos.Side(world.FaceDown)).Model().FaceSolid(pos.Side(world.FaceDown), world.FaceUp, w) {
		w.BreakBlock(pos)
		return
	}
	if c.Dead {
		return
	}
	w.ScheduleBlockUpdate(pos, time.Second*5/2)
}

// ScheduledTick ...
func (c Coral) ScheduledTick(pos world.BlockPos, w *world.World) {
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
func (c Coral) BreakInfo() BreakInfo {
	return BreakInfo{
		Hardness: 0,
		Harvestable: func(t tool.Tool) bool {
			return false //TODO: Silk touch
		},
		Effective: nothingEffective,
		Drops:     simpleDrops(),
	}
}

// EncodeBlock ...
func (c Coral) EncodeBlock() (name string, properties map[string]interface{}) {
	return "minecraft:coral", map[string]interface{}{"coral_color": c.Type.Colour.String(), "dead_bit": c.Dead}
}

// Hash ...
func (c Coral) Hash() uint64 {
	return hashCoral | (uint64(boolByte(c.Dead)) << 32) | (uint64(c.Type.Uint8()) << 33)
}

// EncodeItem ...
func (c Coral) EncodeItem() (id int32, meta int16) {
	if c.Dead {
		return -131, int16(c.Type.Uint8() | 8)
	}
	return -131, int16(c.Type.Uint8())
}

// allCoral returns a list of all coral block variants
func allCoral() (c []world.Block) {
	f := func(dead bool) {
		c = append(c, Coral{Type: coral.Tube(), Dead: dead})
		c = append(c, Coral{Type: coral.Brain(), Dead: dead})
		c = append(c, Coral{Type: coral.Bubble(), Dead: dead})
		c = append(c, Coral{Type: coral.Fire(), Dead: dead})
		c = append(c, Coral{Type: coral.Horn(), Dead: dead})
	}
	f(true)
	f(false)
	return
}
