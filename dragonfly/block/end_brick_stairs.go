package block

import (
	"github.com/df-mc/dragonfly/dragonfly/block/model"
	"github.com/df-mc/dragonfly/dragonfly/item"
	"github.com/df-mc/dragonfly/dragonfly/world"
	"github.com/go-gl/mathgl/mgl64"
)

// EndBrickStairs are blocks that allow entities to walk up blocks without jumping. They are crafted using end bricks.
type EndBrickStairs struct {
	noNBT
	transparent

	// UpsideDown specifies if the stairs are upside down. If set to true, the full side is at the top part
	// of the block.
	UpsideDown bool
	// Facing is the direction that the full side of the stairs is facing.
	Facing world.Direction
}

// UseOnBlock handles the directional placing of stairs and makes sure they are properly placed upside down
// when needed.
func (s EndBrickStairs) UseOnBlock(pos world.BlockPos, face world.Face, clickPos mgl64.Vec3, w *world.World, user item.User, ctx *item.UseContext) (used bool) {
	pos, face, used = firstReplaceable(w, pos, face, s)
	if !used {
		return
	}
	s.Facing = user.Facing()
	if face == world.FaceDown || (clickPos[1] > 0.5 && face != world.FaceUp) {
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
func (s EndBrickStairs) EncodeItem() (id int32, meta int16) {
	return -178, 0
}

// EncodeBlock ...
func (s EndBrickStairs) EncodeBlock() (name string, properties map[string]interface{}) {
	return "minecraft:end_brick_stairs", map[string]interface{}{"upside_down_bit": s.UpsideDown, "weirdo_direction": toStairsDirection(s.Facing)}
}

// Hash ...
func (s EndBrickStairs) Hash() uint64 {
	return hashEndBrickStairs | (uint64(boolByte(s.UpsideDown)) << 32) | (uint64(s.Facing) << 33)
}

// CanDisplace ...
func (EndBrickStairs) CanDisplace(b world.Liquid) bool {
	_, ok := b.(Water)
	return ok
}

// SideClosed ...
func (s EndBrickStairs) SideClosed(pos, side world.BlockPos, w *world.World) bool {
	return s.Model().FaceSolid(pos, pos.Face(side), w)
}

// allEndBrickStairs returns all states of endbrick stairs.
func allEndBrickStairs() (stairs []world.Block) {
	for i := world.Direction(0); i <= 3; i++ {
		stairs = append(stairs, EndBrickStairs{Facing: i, UpsideDown: true})
		stairs = append(stairs, EndBrickStairs{Facing: i, UpsideDown: false})
	}
	return
}
