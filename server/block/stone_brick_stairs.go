package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/model"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// StoneBrickStairs are blocks that allow entities to walk up blocks without jumping. They are crafted using stone bricks.
type StoneBrickStairs struct {
	transparent

	// Mossy specifies if the stairs are mossy.
	Mossy bool
	// UpsideDown specifies if the stairs are upside down. If set to true, the full side is at the top part
	// of the block.
	UpsideDown bool
	// Facing is the direction that the full side of the stairs is facing.
	Facing cube.Direction
}

// UseOnBlock handles the directional placing of stairs and makes sure they are properly placed upside down
// when needed.
func (s StoneBrickStairs) UseOnBlock(pos cube.Pos, face cube.Face, clickPos mgl64.Vec3, w *world.World, user item.User, ctx *item.UseContext) (used bool) {
	pos, face, used = firstReplaceable(w, pos, face, s)
	if !used {
		return
	}
	s.Facing = user.Facing()
	if face == cube.FaceDown || (clickPos[1] > 0.5 && face != cube.FaceUp) {
		s.UpsideDown = true
	}

	place(w, pos, s, user, ctx)
	return placed(ctx)
}

// Model ...
func (s StoneBrickStairs) Model() world.BlockModel {
	return model.Stair{Facing: s.Facing, UpsideDown: s.UpsideDown}
}

// BreakInfo ...
func (s StoneBrickStairs) BreakInfo() BreakInfo {
	return newBreakInfo(1.5, pickaxeHarvestable, pickaxeEffective, oneOf(s))
}

// EncodeItem ...
func (s StoneBrickStairs) EncodeItem() (name string, meta int16) {
	if s.Mossy {
		return "minecraft:mossy_stone_brick_stairs", 0
	}
	return "minecraft:stone_brick_stairs", 0
}

// EncodeBlock ...
func (s StoneBrickStairs) EncodeBlock() (name string, properties map[string]any) {
	if s.Mossy {
		return "minecraft:mossy_stone_brick_stairs", map[string]any{"upside_down_bit": s.UpsideDown, "weirdo_direction": toStairsDirection(s.Facing)}
	}
	return "minecraft:stone_brick_stairs", map[string]any{"upside_down_bit": s.UpsideDown, "weirdo_direction": toStairsDirection(s.Facing)}
}

// CanDisplace ...
func (StoneBrickStairs) CanDisplace(b world.Liquid) bool {
	_, ok := b.(Water)
	return ok
}

// SideClosed ...
func (s StoneBrickStairs) SideClosed(pos, side cube.Pos, w *world.World) bool {
	return s.Model().FaceSolid(pos, pos.Face(side), w)
}

// allStoneBrickStairs ...
func allStoneBrickStairs() (stairs []world.Block) {
	for direction := cube.Direction(0); direction <= 3; direction++ {
		stairs = append(stairs, StoneBrickStairs{Facing: direction, UpsideDown: true, Mossy: true})
		stairs = append(stairs, StoneBrickStairs{Facing: direction, UpsideDown: false, Mossy: true})
		stairs = append(stairs, StoneBrickStairs{Facing: direction, UpsideDown: true, Mossy: false})
		stairs = append(stairs, StoneBrickStairs{Facing: direction, UpsideDown: false, Mossy: false})
	}
	return
}
