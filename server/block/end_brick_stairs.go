package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/model"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// EndBrickStairs are blocks that allow entities to walk up blocks without jumping. They are crafted using end bricks.
type EndBrickStairs struct {
	transparent

	// UpsideDown specifies if the stairs are upside down. If set to true, the full side is at the top part
	// of the block.
	UpsideDown bool
	// Facing is the direction that the full side of the stairs is facing.
	Facing cube.Direction
}

// UseOnBlock handles the directional placing of stairs and makes sure they are properly placed upside down
// when needed.
func (s EndBrickStairs) UseOnBlock(pos cube.Pos, face cube.Face, clickPos mgl64.Vec3, w *world.World, user item.User, ctx *item.UseContext) (used bool) {
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
func (s EndBrickStairs) Model() world.BlockModel {
	return model.Stair{Facing: s.Facing, UpsideDown: s.UpsideDown}
}

// BreakInfo ...
func (s EndBrickStairs) BreakInfo() BreakInfo {
	return BreakInfo{
		Hardness:    3,
		Harvestable: pickaxeHarvestable,
		Effective:   pickaxeEffective,
		Drops:       simpleDrops(item.NewStack(s, 1)),
	}
}

// EncodeItem ...
func (s EndBrickStairs) EncodeItem() (id int32, name string, meta int16) {
	return -178, "minecraft:end_brick_stairs", 0
}

// EncodeBlock ...
func (s EndBrickStairs) EncodeBlock() (name string, properties map[string]interface{}) {
	return "minecraft:end_brick_stairs", map[string]interface{}{"upside_down_bit": s.UpsideDown, "weirdo_direction": toStairsDirection(s.Facing)}
}

// CanDisplace ...
func (EndBrickStairs) CanDisplace(b world.Liquid) bool {
	_, ok := b.(Water)
	return ok
}

// SideClosed ...
func (s EndBrickStairs) SideClosed(pos, side cube.Pos, w *world.World) bool {
	return s.Model().FaceSolid(pos, pos.Face(side), w)
}

// allEndBrickStairs ...
func allEndBrickStairs() (stairs []world.Block) {
	for direction := cube.Direction(0); direction <= 3; direction++ {
		stairs = append(stairs, EndBrickStairs{Facing: direction, UpsideDown: true})
		stairs = append(stairs, EndBrickStairs{Facing: direction, UpsideDown: false})
	}
	return
}
