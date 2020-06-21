package block

import (
	"git.jetbrains.space/dragonfly/dragonfly.git/dragonfly/block/wood"
	"git.jetbrains.space/dragonfly/dragonfly.git/dragonfly/entity/physics"
	"git.jetbrains.space/dragonfly/dragonfly.git/dragonfly/item"
	"git.jetbrains.space/dragonfly/dragonfly.git/dragonfly/world"
	"github.com/go-gl/mathgl/mgl64"
)

// WoodStairs are blocks that allow entities to walk up blocks without jumping. They are crafted using planks.
type WoodStairs struct {
	// Wood is the type of wood of the stairs. This field must have one of the values found in the material
	// package.
	Wood wood.Wood
	// UpsideDown specifies if the stairs are upside down. If set to true, the full side is at the top part
	// of the block.
	UpsideDown bool
	// Facing is the direction that the full side of the stairs is facing.
	Facing world.Direction
}

// UseOnBlock handles the directional placing of stairs and makes sure they are properly placed upside down
// when needed.
func (s WoodStairs) UseOnBlock(pos world.BlockPos, face world.Face, clickPos mgl64.Vec3, w *world.World, user item.User, ctx *item.UseContext) (used bool) {
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
func (s WoodStairs) BreakInfo() BreakInfo {
	return BreakInfo{
		Hardness:    2,
		Harvestable: alwaysHarvestable,
		Effective:   axeEffective,
		Drops:       simpleDrops(item.NewStack(s, 1)),
	}
}

// LightDiffusionLevel always returns 0.
func (WoodStairs) LightDiffusionLevel() uint8 {
	return 0
}

// AABB ...
func (s WoodStairs) AABB(pos world.BlockPos, w *world.World) []physics.AABB {
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
func (s WoodStairs) EncodeItem() (id int32, meta int16) {
	switch s.Wood {
	case wood.Oak():
		return 53, 0
	case wood.Spruce():
		return 134, 0
	case wood.Birch():
		return 135, 0
	case wood.Jungle():
		return 136, 0
	case wood.Acacia():
		return 163, 0
	case wood.DarkOak():
		return 164, 0
	}
	panic("invalid wood type")
}

// EncodeBlock ...
func (s WoodStairs) EncodeBlock() (name string, properties map[string]interface{}) {
	switch s.Wood {
	case wood.Oak():
		return "minecraft:oak_stairs", map[string]interface{}{"upside_down_bit": s.UpsideDown, "weirdo_direction": toStairsDirection(s.Facing)}
	case wood.Spruce():
		return "minecraft:spruce_stairs", map[string]interface{}{"upside_down_bit": s.UpsideDown, "weirdo_direction": toStairsDirection(s.Facing)}
	case wood.Birch():
		return "minecraft:birch_stairs", map[string]interface{}{"upside_down_bit": s.UpsideDown, "weirdo_direction": toStairsDirection(s.Facing)}
	case wood.Jungle():
		return "minecraft:jungle_stairs", map[string]interface{}{"upside_down_bit": s.UpsideDown, "weirdo_direction": toStairsDirection(s.Facing)}
	case wood.Acacia():
		return "minecraft:acacia_stairs", map[string]interface{}{"upside_down_bit": s.UpsideDown, "weirdo_direction": toStairsDirection(s.Facing)}
	case wood.DarkOak():
		return "minecraft:dark_oak_stairs", map[string]interface{}{"upside_down_bit": s.UpsideDown, "weirdo_direction": toStairsDirection(s.Facing)}
	}
	panic("invalid wood type")
}

// toStairDirection converts a facing to a stairs direction for Minecraft.
func toStairsDirection(v world.Direction) int32 {
	return int32(3 - v)
}

// CanDisplace ...
func (WoodStairs) CanDisplace(b world.Liquid) bool {
	_, ok := b.(Water)
	return ok
}

// SideClosed ...
func (s WoodStairs) SideClosed(pos, side world.BlockPos, w *world.World) bool {
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

const (
	noCorner = iota
	cornerRightInner
	cornerLeftInner
	cornerRightOuter
	cornerLeftOuter
)

// cornerType returns the type of the corner that the stairs form, or 0 if it does not form a corner with any
// other stairs.
func (s WoodStairs) cornerType(pos world.BlockPos, w *world.World) uint8 {
	// TODO: Make stairs of all types curve.
	rotatedFacing := s.Facing.Rotate90()
	if closedSide, ok := w.Block(pos.Side(s.Facing.Face())).(WoodStairs); ok && closedSide.UpsideDown == s.UpsideDown {
		if closedSide.Facing == rotatedFacing {
			return cornerLeftOuter
		} else if closedSide.Facing == rotatedFacing.Opposite() {
			// This will only form a corner if there is not a stair on the right of this one with the same
			// direction.
			if side, ok := w.Block(pos.Side(s.Facing.Rotate90().Face())).(WoodStairs); !ok || side.Facing != s.Facing || side.UpsideDown != s.UpsideDown {
				return cornerRightOuter
			}
			return noCorner
		}
	}
	if openSide, ok := w.Block(pos.Side(s.Facing.Opposite().Face())).(WoodStairs); ok && openSide.UpsideDown == s.UpsideDown {
		if openSide.Facing == rotatedFacing {
			// This will only form a corner if there is not a stair on the right of this one with the same
			// direction.
			if side, ok := w.Block(pos.Side(s.Facing.Rotate90().Face())).(WoodStairs); !ok || side.Facing != s.Facing || side.UpsideDown != s.UpsideDown {
				return cornerRightInner
			}
		} else if openSide.Facing == rotatedFacing.Opposite() {
			return cornerLeftInner
		}
	}
	return noCorner
}

// allWoodStairs returns all states of wood stairs.
func allWoodStairs() (stairs []world.Block) {
	f := func(facing world.Direction, upsideDown bool) {
		stairs = append(stairs, WoodStairs{Facing: facing, UpsideDown: upsideDown, Wood: wood.Oak()})
		stairs = append(stairs, WoodStairs{Facing: facing, UpsideDown: upsideDown, Wood: wood.Spruce()})
		stairs = append(stairs, WoodStairs{Facing: facing, UpsideDown: upsideDown, Wood: wood.Birch()})
		stairs = append(stairs, WoodStairs{Facing: facing, UpsideDown: upsideDown, Wood: wood.Jungle()})
		stairs = append(stairs, WoodStairs{Facing: facing, UpsideDown: upsideDown, Wood: wood.Acacia()})
		stairs = append(stairs, WoodStairs{Facing: facing, UpsideDown: upsideDown, Wood: wood.DarkOak()})
	}
	for i := world.Direction(0); i <= 3; i++ {
		f(i, true)
		f(i, false)
	}
	return
}
