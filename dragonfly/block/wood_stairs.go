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
	Facing world.Face
}

// UseOnBlock handles the directional placing of stairs and makes sure they are properly placed upside down
// when needed.
func (s WoodStairs) UseOnBlock(pos world.BlockPos, face world.Face, clickPos mgl64.Vec3, w *world.World, user item.User, ctx *item.UseContext) (used bool) {
	pos, face, used = firstReplaceable(w, pos, face, s)
	if !used {
		return
	}
	s.Facing = user.Facing()
	if face == world.Down || (clickPos[1] > 0.5 && face != world.Up) {
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
func (s WoodStairs) AABB() []physics.AABB {
	// TODO: Account for stair curving.
	b := []physics.AABB{physics.NewAABB(mgl64.Vec3{}, mgl64.Vec3{1, 0.5, 1})}
	if s.UpsideDown {
		b[0] = physics.NewAABB(mgl64.Vec3{0, 0.5, 0}, mgl64.Vec3{1, 1, 1})
	}
	switch s.Facing {
	case world.North:
		b = append(b, physics.NewAABB(mgl64.Vec3{0, 0, 0.5}, mgl64.Vec3{1, 0.5, 1}))
	case world.South:
		b = append(b, physics.NewAABB(mgl64.Vec3{}, mgl64.Vec3{1, 0.5, 0.5}))
	case world.East:
		b = append(b, physics.NewAABB(mgl64.Vec3{0.5, 0, 0}, mgl64.Vec3{1, 0.5, 1}))
	case world.West:
		b = append(b, physics.NewAABB(mgl64.Vec3{}, mgl64.Vec3{0.5, 0.5, 1}))
	}

	if s.UpsideDown {
		b[0] = b[0].Translate(mgl64.Vec3{0, 0.5})
	} else {
		b[1] = b[1].Translate(mgl64.Vec3{0, 0.5})
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
func toStairsDirection(v world.Face) int32 {
	return int32(3 - (v - 2))
}

// CanDisplace ...
func (WoodStairs) CanDisplace(b world.Liquid) bool {
	_, ok := b.(Water)
	return ok
}

// SideClosed ...
func (s WoodStairs) SideClosed(pos, side world.BlockPos) bool {
	if !s.UpsideDown && side[1] == pos[1]-1 {
		// Non-upside down stairs have a closed side at the bottom.
		return true
	}
	// TODO: Implement stairs rotation calculations.
	if pos.Side(s.Facing) == side {
		return true
	}
	return false
}

// allWoodStairs returns all states of wood stairs.
func allWoodStairs() (stairs []world.Block) {
	f := func(facing world.Face, upsideDown bool) {
		stairs = append(stairs, WoodStairs{Facing: facing, UpsideDown: upsideDown, Wood: wood.Oak()})
		stairs = append(stairs, WoodStairs{Facing: facing, UpsideDown: upsideDown, Wood: wood.Spruce()})
		stairs = append(stairs, WoodStairs{Facing: facing, UpsideDown: upsideDown, Wood: wood.Birch()})
		stairs = append(stairs, WoodStairs{Facing: facing, UpsideDown: upsideDown, Wood: wood.Jungle()})
		stairs = append(stairs, WoodStairs{Facing: facing, UpsideDown: upsideDown, Wood: wood.Acacia()})
		stairs = append(stairs, WoodStairs{Facing: facing, UpsideDown: upsideDown, Wood: wood.DarkOak()})
	}
	for i := world.Face(2); i <= 5; i++ {
		f(i, true)
		f(i, false)
	}
	return
}
