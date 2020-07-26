package block

import (
	"github.com/df-mc/dragonfly/dragonfly/entity/physics"
	"github.com/df-mc/dragonfly/dragonfly/item"
	"github.com/df-mc/dragonfly/dragonfly/world"
	"github.com/go-gl/mathgl/mgl64"
)

// EndBrickStairs are blocks that allow entities to walk up blocks without jumping. They are crafted using end bricks.
type EndBrickStairs struct {
	noNBT
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

// BreakInfo ...
func (s EndBrickStairs) BreakInfo() BreakInfo {
	return BreakInfo{
		Hardness:    3,
		Harvestable: pickaxeHarvestable,
		Effective:   pickaxeEffective,
		Drops:       simpleDrops(item.NewStack(s, 1)),
	}
}

// LightDiffusionLevel always returns 0.
func (EndBrickStairs) LightDiffusionLevel() uint8 {
	return 0
}

// AABB ...
func (s EndBrickStairs) AABB(pos world.BlockPos, w *world.World) []physics.AABB {
	b := []physics.AABB{physics.NewAABB(mgl64.Vec3{}, mgl64.Vec3{1, 0.5, 1})}
	if s.UpsideDown {
		b[0] = physics.NewAABB(mgl64.Vec3{0, 0.5, 0}, mgl64.Vec3{1, 1, 1})
	}
	t := s.cornerType(pos, w)

	if t == noCorner || t == cornerRightInner || t == cornerRightOuter {
		b = append(b, physics.NewAABB(mgl64.Vec3{0.5, 0.5, 0.5}, mgl64.Vec3{0.5, 1, 0.5}).
			ExtendTowards(int(s.Facing), 0.5).
			ExtendTowards(int(s.Facing.Rotate90()), 0.5).
			ExtendTowards(int(s.Facing.Rotate90().Opposite()), 0.5))
	}
	if t == cornerRightOuter {
		b = append(b, physics.NewAABB(mgl64.Vec3{0.5, 0.5, 0.5}, mgl64.Vec3{0.5, 1, 0.5}).
			ExtendTowards(int(s.Facing), 0.5).
			ExtendTowards(int(s.Facing.Rotate90().Opposite()), 0.5))
	} else if t == cornerLeftOuter {
		b = append(b, physics.NewAABB(mgl64.Vec3{0.5, 0.5, 0.5}, mgl64.Vec3{0.5, 1, 0.5}).
			ExtendTowards(int(s.Facing), 0.5).
			ExtendTowards(int(s.Facing.Rotate90()), 0.5))
	} else if t == cornerRightInner {
		b = append(b, physics.NewAABB(mgl64.Vec3{0.5, 0.5, 0.5}, mgl64.Vec3{0.5, 1, 0.5}).
			ExtendTowards(int(s.Facing.Opposite()), 0.5).
			ExtendTowards(int(s.Facing.Rotate90().Opposite()), 0.5))
	} else if t == cornerLeftInner {
		b = append(b, physics.NewAABB(mgl64.Vec3{0.5, 0.5, 0.5}, mgl64.Vec3{0.5, 1, 0.5}).
			ExtendTowards(int(s.Facing.Opposite()), 0.5).
			ExtendTowards(int(s.Facing.Rotate90()), 0.5))
	}

	if s.UpsideDown {
		for i := range b[1:] {
			b[i] = b[i].Translate(mgl64.Vec3{0, -0.5})
		}
	}
	return b
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
	if !s.UpsideDown && side[1] == pos[1]-1 {
		// Non-upside down stairs have a closed side at the bottom.
		return true
	}
	t := s.cornerType(pos, w)
	if t == cornerRightOuter || t == cornerLeftOuter {
		// Small corner blocks, they do not block water flowing out horizontally.
		return false
	} else if t == noCorner {
		// Not a corner, so only block directly behind the stairs.
		return pos.Side(s.Facing.Face()) == side
	}
	if t == cornerRightInner {
		return side == pos.Side(s.Facing.Rotate90().Face()) || side == pos.Side(s.Facing.Face())
	}
	return side == pos.Side(s.Facing.Rotate90().Opposite().Face()) || side == pos.Side(s.Facing.Face())
}

// cornerType returns the type of the corner that the stairs form, or 0 if it does not form a corner with any
// other stairs.
func (s EndBrickStairs) cornerType(pos world.BlockPos, w *world.World) uint8 {
	// TODO: Make stairs of all types curve.
	rotatedFacing := s.Facing.Rotate90()
	if closedSide, ok := w.Block(pos.Side(s.Facing.Face())).(EndBrickStairs); ok && closedSide.UpsideDown == s.UpsideDown {
		if closedSide.Facing == rotatedFacing {
			return cornerLeftOuter
		} else if closedSide.Facing == rotatedFacing.Opposite() {
			// This will only form a corner if there is not a stair on the right of this one with the same
			// direction.
			if side, ok := w.Block(pos.Side(s.Facing.Rotate90().Face())).(EndBrickStairs); !ok || side.Facing != s.Facing || side.UpsideDown != s.UpsideDown {
				return cornerRightOuter
			}
			return noCorner
		}
	}
	if openSide, ok := w.Block(pos.Side(s.Facing.Opposite().Face())).(EndBrickStairs); ok && openSide.UpsideDown == s.UpsideDown {
		if openSide.Facing == rotatedFacing {
			// This will only form a corner if there is not a stair on the right of this one with the same
			// direction.
			if side, ok := w.Block(pos.Side(s.Facing.Rotate90().Face())).(EndBrickStairs); !ok || side.Facing != s.Facing || side.UpsideDown != s.UpsideDown {
				return cornerRightInner
			}
		} else if openSide.Facing == rotatedFacing.Opposite() {
			return cornerLeftInner
		}
	}
	return noCorner
}

// allEndBrickStairs returns all states of endbrick stairs.
func allEndBrickStairs() (stairs []world.Block) {
	for i := world.Direction(0); i <= 3; i++ {
		stairs = append(stairs, EndBrickStairs{Facing: i, UpsideDown: true})
		stairs = append(stairs, EndBrickStairs{Facing: i, UpsideDown: false})
	}
	return
}
