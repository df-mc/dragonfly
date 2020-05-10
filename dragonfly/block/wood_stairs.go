package block

import (
	"git.jetbrains.space/dragonfly/dragonfly.git/dragonfly/block/material"
	"git.jetbrains.space/dragonfly/dragonfly.git/dragonfly/entity/physics"
	"git.jetbrains.space/dragonfly/dragonfly.git/dragonfly/item"
	"git.jetbrains.space/dragonfly/dragonfly.git/dragonfly/world"
	"github.com/go-gl/mathgl/mgl32"
)

// WoodStairs are blocks that allow entities to walk up blocks without jumping. They are crafted using planks.
type WoodStairs struct {
	// Wood is the type of wood of the stairs. This field must have one of the values found in the material
	// package.
	Wood material.Wood
	// UpsideDown specifies if the stairs are upside down. If set to true, the full side is at the top part
	// of the block.
	UpsideDown bool
	// Facing is the direction that the stairs are facing.
	Facing world.Face
}

// UseOnBlock handles the directional placing of stairs and makes sure they are properly placed upside down
// when needed.
func (w WoodStairs) UseOnBlock(pos world.BlockPos, face world.Face, clickPos mgl32.Vec3, wo *world.World, user item.User, ctx *item.UseContext) bool {
	if replaceable(wo, pos.Side(face), w) {
		w.Facing = user.Facing()

		if face == world.Up {
			w.UpsideDown = false
		} else if face == world.Down || clickPos[1] > 0.5 {
			w.UpsideDown = true
		}

		wo.PlaceBlock(pos.Side(face), w)

		ctx.SubtractFromCount(1)
		return true
	}
	return false
}

// BreakInfo ...
func (w WoodStairs) BreakInfo() BreakInfo {
	return BreakInfo{
		Hardness:    2,
		Harvestable: alwaysHarvestable,
		Effective:   axeEffective,
		Drops:       simpleDrops(item.NewStack(w, 1)),
	}
}

// AABB ...
func (w WoodStairs) AABB() []physics.AABB {
	// TODO: Account for stair curving.
	b := []physics.AABB{physics.NewAABB(mgl32.Vec3{0, 0, 0}, mgl32.Vec3{1, 0.5, 1})}
	if w.UpsideDown {
		b[0] = physics.NewAABB(mgl32.Vec3{0, 0.5, 0}, mgl32.Vec3{1, 1, 1})
	}
	switch w.Facing {
	case world.North:
		b = append(b, physics.NewAABB(mgl32.Vec3{0, 0, 0.5}, mgl32.Vec3{1, 0.5, 1}))
	case world.South:
		b = append(b, physics.NewAABB(mgl32.Vec3{0, 0, 0}, mgl32.Vec3{1, 0.5, 0.5}))
	case world.East:
		b = append(b, physics.NewAABB(mgl32.Vec3{0.5, 0, 0}, mgl32.Vec3{1, 0.5, 1}))
	case world.West:
		b = append(b, physics.NewAABB(mgl32.Vec3{0, 0, 0}, mgl32.Vec3{0.5, 0.5, 1}))
	}

	if w.UpsideDown {
		b[0] = b[0].Translate(mgl32.Vec3{0, 0.5})
	} else {
		b[1] = b[1].Translate(mgl32.Vec3{0, 0.5})
	}
	return b
}

// EncodeItem ...
func (w WoodStairs) EncodeItem() (id int32, meta int16) {
	switch w.Wood {
	case material.OakWood():
		return 53, 0
	case material.SpruceWood():
		return 134, 0
	case material.BirchWood():
		return 135, 0
	case material.JungleWood():
		return 136, 0
	case material.AcaciaWood():
		return 163, 0
	case material.DarkOakWood():
		return 164, 0
	}
	panic("invalid wood type")
}

// EncodeBlock ...
func (w WoodStairs) EncodeBlock() (name string, properties map[string]interface{}) {
	switch w.Wood {
	case material.OakWood():
		return "minecraft:oak_stairs", map[string]interface{}{"upside_down_bit": w.UpsideDown, "weirdo_direction": toStairsDirection(w.Facing)}
	case material.SpruceWood():
		return "minecraft:spruce_stairs", map[string]interface{}{"upside_down_bit": w.UpsideDown, "weirdo_direction": toStairsDirection(w.Facing)}
	case material.BirchWood():
		return "minecraft:birch_stairs", map[string]interface{}{"upside_down_bit": w.UpsideDown, "weirdo_direction": toStairsDirection(w.Facing)}
	case material.JungleWood():
		return "minecraft:jungle_stairs", map[string]interface{}{"upside_down_bit": w.UpsideDown, "weirdo_direction": toStairsDirection(w.Facing)}
	case material.AcaciaWood():
		return "minecraft:acacia_stairs", map[string]interface{}{"upside_down_bit": w.UpsideDown, "weirdo_direction": toStairsDirection(w.Facing)}
	case material.DarkOakWood():
		return "minecraft:dark_oak_stairs", map[string]interface{}{"upside_down_bit": w.UpsideDown, "weirdo_direction": toStairsDirection(w.Facing)}
	}
	panic("invalid wood type")
}

// toStairDirection converts a facing to a stairs direction for Minecraft.
func toStairsDirection(v world.Face) int32 {
	return int32(3 - (v - 2))
}

// allWoodStairs returns all states of wood stairs.
func allWoodStairs() (stairs []world.Block) {
	f := func(facing world.Face, upsideDown bool) {
		stairs = append(stairs, WoodStairs{Facing: facing, UpsideDown: upsideDown, Wood: material.OakWood()})
		stairs = append(stairs, WoodStairs{Facing: facing, UpsideDown: upsideDown, Wood: material.SpruceWood()})
		stairs = append(stairs, WoodStairs{Facing: facing, UpsideDown: upsideDown, Wood: material.BirchWood()})
		stairs = append(stairs, WoodStairs{Facing: facing, UpsideDown: upsideDown, Wood: material.JungleWood()})
		stairs = append(stairs, WoodStairs{Facing: facing, UpsideDown: upsideDown, Wood: material.AcaciaWood()})
		stairs = append(stairs, WoodStairs{Facing: facing, UpsideDown: upsideDown, Wood: material.DarkOakWood()})
	}
	for i := world.Face(2); i <= 5; i++ {
		f(i, true)
		f(i, false)
	}
	return
}
