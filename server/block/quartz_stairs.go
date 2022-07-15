package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/model"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// QuartzStairs are blocks that allow entities to walk up blocks without jumping. They are crafted using end bricks.
type QuartzStairs struct {
	transparent

	// UpsideDown specifies if the stairs are upside down. If set to true, the full side is at the top part
	// of the block.
	UpsideDown bool
	// Facing is the direction that the full side of the stairs is facing.
	Facing cube.Direction

	// Smooth indicates if it's smooth or not.
	Smooth bool
}

// UseOnBlock handles the directional placing of stairs and makes sure they are properly placed upside down
// when needed.
func (s QuartzStairs) UseOnBlock(pos cube.Pos, face cube.Face, clickPos mgl64.Vec3, w *world.World, user item.User, ctx *item.UseContext) (used bool) {
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
func (s QuartzStairs) Model() world.BlockModel {
	return model.Stair{Facing: s.Facing, UpsideDown: s.UpsideDown}
}

// BreakInfo ...
func (s QuartzStairs) BreakInfo() BreakInfo {
	return newBreakInfo(2, pickaxeHarvestable, pickaxeEffective, oneOf(s))
}

// EncodeItem ...
func (s QuartzStairs) EncodeItem() (name string, meta int16) {
	if s.Smooth {
		return "minecraft:smooth_quartz_stairs", 0
	}
	return "minecraft:quartz_stairs", 0
}

// EncodeBlock ...
func (s QuartzStairs) EncodeBlock() (name string, properties map[string]any) {
	if s.Smooth {
		return "minecraft:smooth_quartz_stairs", map[string]any{"upside_down_bit": s.UpsideDown, "weirdo_direction": toStairsDirection(s.Facing)}
	}
	return "minecraft:quartz_stairs", map[string]any{"upside_down_bit": s.UpsideDown, "weirdo_direction": toStairsDirection(s.Facing)}
}

// CanDisplace ...
func (QuartzStairs) CanDisplace(b world.Liquid) bool {
	_, ok := b.(Water)
	return ok
}

// SideClosed ...
func (s QuartzStairs) SideClosed(pos, side cube.Pos, w *world.World) bool {
	return s.Model().FaceSolid(pos, pos.Face(side), w)
}

// allQuartzStairs ...
func allQuartzStairs() (stairs []world.Block) {
	for direction := cube.Direction(0); direction <= 3; direction++ {
		stairs = append(stairs, QuartzStairs{Facing: direction, UpsideDown: true, Smooth: true})
		stairs = append(stairs, QuartzStairs{Facing: direction, UpsideDown: false, Smooth: true})
		stairs = append(stairs, QuartzStairs{Facing: direction, UpsideDown: true, Smooth: false})
		stairs = append(stairs, QuartzStairs{Facing: direction, UpsideDown: false, Smooth: false})
	}
	return
}
