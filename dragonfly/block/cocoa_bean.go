package block

import (
	"github.com/df-mc/dragonfly/dragonfly/block/model"
	"github.com/df-mc/dragonfly/dragonfly/block/wood"
	"github.com/df-mc/dragonfly/dragonfly/item"
	"github.com/df-mc/dragonfly/dragonfly/item/tool"
	"github.com/df-mc/dragonfly/dragonfly/world"
	"github.com/go-gl/mathgl/mgl64"
	"math/rand"
)

// CocoaBean is a crop block found in Jungle biomes.
type CocoaBean struct {
	noNBT
	transparent

	// Facing is the direction from the cocoa bean to the log.
	Facing world.Direction
	// Age is the stage of the cocoa bean's growth. 2 is fully grown.
	Age int
}

// Bonemeal ...
func (c CocoaBean) Bonemeal(pos world.BlockPos, w *world.World) bool {
	if c.Age == 2 {
		return false
	}
	c.Age++
	w.PlaceBlock(pos, c)
	return true
}

// HasLiquidDrops ...
func (c CocoaBean) HasLiquidDrops() bool {
	return true
}

// NeighbourUpdateTick ...
func (c CocoaBean) NeighbourUpdateTick(pos, _ world.BlockPos, w *world.World) {
	if log, ok := w.Block(pos.Side(c.Facing.Face())).(Log); !ok || log.Wood != wood.Jungle() || log.Stripped {
		w.BreakBlock(pos)
	}
}

// UseOnBlock ...
func (c CocoaBean) UseOnBlock(pos world.BlockPos, face world.Face, _ mgl64.Vec3, w *world.World, user item.User, ctx *item.UseContext) bool {
	pos, _, used := firstReplaceable(w, pos, face, c)
	if !used {
		return false
	}

	if face == world.FaceUp || face == world.FaceDown {
		return false
	}
	if log, ok := w.Block(pos.Side(face.Opposite())).(Log); ok {
		if log.Wood == wood.Jungle() && !log.Stripped {
			c.Facing = face.Opposite().Direction()
			ctx.IgnoreAABB = true

			place(w, pos, c, user, ctx)
			return placed(ctx)
		}
	}

	return false
}

// RandomTick ...
func (c CocoaBean) RandomTick(pos world.BlockPos, w *world.World, r *rand.Rand) {
	if c.Age < 2 && r.Intn(5) == 0 {
		c.Age++
		w.PlaceBlock(pos, c)
	}
}

// BreakInfo ...
func (c CocoaBean) BreakInfo() BreakInfo {
	return BreakInfo{
		Hardness:    0.2,
		Harvestable: alwaysHarvestable,
		Effective:   axeEffective,
		Drops: func(t tool.Tool) []item.Stack {
			if c.Age == 2 {
				return []item.Stack{item.NewStack(c, rand.Intn(2)+2)}
			}
			return []item.Stack{item.NewStack(c, 1)}
		},
	}
}

// EncodeItem ...
func (c CocoaBean) EncodeItem() (id int32, meta int16) {
	return 351, 3
}

// EncodeBlock ...
func (c CocoaBean) EncodeBlock() (name string, properties map[string]interface{}) {
	direction := 2
	switch c.Facing {
	case world.South:
		direction = 0
	case world.West:
		direction = 1
	case world.East:
		direction = 3
	}

	return "minecraft:cocoa", map[string]interface{}{"age": int32(c.Age), "direction": int32(direction)}
}

// Hash ...
func (c CocoaBean) Hash() uint64 {
	return hashCocoa | (uint64(c.Age) << 32) | (uint64(c.Facing) << 34)
}

// Model ...
func (c CocoaBean) Model() world.BlockModel {
	return model.CocoaBean{Facing: c.Facing, Age: c.Age}
}

// allCocoaBeans ...
func allCocoaBeans() (cocoa []world.Block) {
	for i := world.Direction(0); i <= 3; i++ {
		cocoa = append(cocoa, CocoaBean{Facing: i, Age: 0})
		cocoa = append(cocoa, CocoaBean{Facing: i, Age: 1})
		cocoa = append(cocoa, CocoaBean{Facing: i, Age: 2})
	}
	return
}
