package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/model"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// SandstoneStairs are blocks that allow entities to walk up blocks without jumping. They are crafted using sandstone.
type SandstoneStairs struct {
	transparent

	// Smooth will specify if the sandstone block should be smooth sandstone stairs or normal sandstone stairs.
	Smooth bool

	// Red specifies if the sandstone type is red or not. When set to true, the sandstone stairs type will represent its
	// red variant, for example red sandstone stairs.
	Red bool

	// UpsideDown specifies if the stairs are upside down. If set to true, the full side is at the top part
	// of the block.
	UpsideDown bool
	// Facing is the direction that the full side of the stairs is facing.
	Facing cube.Direction
}

// UseOnBlock handles the directional placing of stairs and makes sure they are properly placed upside down
// when needed.
func (s SandstoneStairs) UseOnBlock(pos cube.Pos, face cube.Face, clickPos mgl64.Vec3, w *world.World, user item.User, ctx *item.UseContext) (used bool) {
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
func (s SandstoneStairs) Model() world.BlockModel {
	return model.Stair{Facing: s.Facing, UpsideDown: s.UpsideDown}
}

// BreakInfo ...
func (s SandstoneStairs) BreakInfo() BreakInfo {
	return newBreakInfo(2, pickaxeHarvestable, pickaxeEffective, oneOf(s))
}

// EncodeItem ...
func (s SandstoneStairs) EncodeItem() (name string, meta int16) {
	if s.Smooth {
		if s.Red {
			return "minecraft:smooth_red_sandstone_stairs", 0
		}
		return "minecraft:smooth_sandstone_stairs", 0
	}
	if s.Red {
		return "minecraft:red_sandstone_stairs", 0
	}
	return "minecraft:sandstone_stairs", 0
}

// EncodeBlock ...
func (s SandstoneStairs) EncodeBlock() (name string, properties map[string]interface{}) {
	if s.Smooth {
		if s.Red {
			return "minecraft:smooth_red_sandstone_stairs", map[string]interface{}{"upside_down_bit": s.UpsideDown, "weirdo_direction": toStairsDirection(s.Facing)}
		}
		return "minecraft:smooth_sandstone_stairs", map[string]interface{}{"upside_down_bit": s.UpsideDown, "weirdo_direction": toStairsDirection(s.Facing)}
	}
	if s.Red {
		return "minecraft:red_sandstone_stairs", map[string]interface{}{"upside_down_bit": s.UpsideDown, "weirdo_direction": toStairsDirection(s.Facing)}
	}
	return "minecraft:sandstone_stairs", map[string]interface{}{"upside_down_bit": s.UpsideDown, "weirdo_direction": toStairsDirection(s.Facing)}

}

// CanDisplace ...
func (SandstoneStairs) CanDisplace(b world.Liquid) bool {
	_, ok := b.(Water)
	return ok
}

// SideClosed ...
func (s SandstoneStairs) SideClosed(pos, side cube.Pos, w *world.World) bool {
	return s.Model().FaceSolid(pos, pos.Face(side), w)
}

// allSandstoneStairs ...
func allSandstoneStairs() (stairs []world.Block) {
	f := func(upsideDown bool) {
		for direction := cube.Direction(0); direction <= 3; direction++ {
			stairs = append(stairs, SandstoneStairs{Facing: direction, Smooth: true, Red: false, UpsideDown: upsideDown})
			stairs = append(stairs, SandstoneStairs{Facing: direction, Smooth: true, Red: true, UpsideDown: upsideDown})
			stairs = append(stairs, SandstoneStairs{Facing: direction, Smooth: false, Red: true, UpsideDown: upsideDown})
			stairs = append(stairs, SandstoneStairs{Facing: direction, Smooth: false, Red: false, UpsideDown: upsideDown})
		}
	}
	f(false)
	f(true)
	return
}
