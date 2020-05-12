package block

import (
	"git.jetbrains.space/dragonfly/dragonfly.git/dragonfly/block/material"
	"git.jetbrains.space/dragonfly/dragonfly.git/dragonfly/entity/physics"
	"git.jetbrains.space/dragonfly/dragonfly.git/dragonfly/item"
	"git.jetbrains.space/dragonfly/dragonfly.git/dragonfly/item/tool"
	"git.jetbrains.space/dragonfly/dragonfly.git/dragonfly/world"
	"github.com/go-gl/mathgl/mgl32"
)

// WoodSlab is a half block that allows entities to walk up blocks without jumping.
type WoodSlab struct {
	// Wood is the type of wood of the slabs. This field must have one of the values found in the material
	// package.
	Wood material.Wood
	// UpsideDown specifies if the slabs are upside down.
	UpsideDown bool
	// Double specifies if the slab is a double slab. These double slabs can be made by placing another slab
	// on an existing slab.
	Double bool
}

// UseOnBlock handles the placement of slabs with relation to them being upside down or not and handles slabs
// being turned into double slabs.
func (w WoodSlab) UseOnBlock(pos world.BlockPos, face world.Face, clickPos mgl32.Vec3, wo *world.World, user item.User, ctx *item.UseContext) bool {
	clickedBlock := wo.Block(pos)
	if clickedSlab, ok := clickedBlock.(WoodSlab); ok && !w.Double {
		if face == world.Up && !clickedSlab.Double && clickedSlab.Wood == w.Wood && !clickedSlab.UpsideDown {
			// A half slab of the same type was clicked at the top, so we can make it full.
			clickedSlab.Double = true
			wo.PlaceBlock(pos, clickedSlab)

			ctx.SubtractFromCount(1)
			return true
		} else if face == world.Down && !clickedSlab.Double && clickedSlab.Wood == w.Wood && clickedSlab.UpsideDown {
			// A half slab of the same type was clicked at the bottom, so we can make it full.
			clickedSlab.Double = true
			wo.PlaceBlock(pos, clickedSlab)

			ctx.SubtractFromCount(1)
			return true
		}
	}
	if sideSlab, ok := wo.Block(pos.Side(face)).(WoodSlab); ok && !replaceable(wo, pos, w) && !w.Double {
		// The block on the side of the one clicked was a slab and the block clicked was not replaceable, so
		// the slab on the side must've been half and may now be filled if the wood types are the same.
		if !sideSlab.Double && sideSlab.Wood == w.Wood {
			sideSlab.Double = true
			wo.PlaceBlock(pos.Side(face), sideSlab)

			ctx.SubtractFromCount(1)
			return true
		}
	}
	if replaceable(wo, pos.Side(face), w) {
		if face == world.Down || (clickPos[1] > 0.5 && face != world.Up) {
			w.UpsideDown = true
		}

		wo.PlaceBlock(pos.Side(face), w)

		ctx.SubtractFromCount(1)
		return true
	}
	return false
}

// BreakInfo ...
func (w WoodSlab) BreakInfo() BreakInfo {
	return BreakInfo{
		Hardness:    2,
		Harvestable: alwaysHarvestable,
		Effective:   axeEffective,
		Drops: func(t tool.Tool) []item.Stack {
			if w.Double {
				w.Double = false
				// If the slab is double, it should drop two single slabs.
				return []item.Stack{item.NewStack(w, 2)}
			}
			return []item.Stack{item.NewStack(w, 1)}
		},
	}
}

// AABB ...
func (w WoodSlab) AABB() []physics.AABB {
	if w.Double {
		return []physics.AABB{physics.NewAABB(mgl32.Vec3{}, mgl32.Vec3{1, 1, 1})}
	}
	if w.UpsideDown {
		return []physics.AABB{physics.NewAABB(mgl32.Vec3{0, 0.5, 0}, mgl32.Vec3{1, 1, 1})}
	}
	return []physics.AABB{physics.NewAABB(mgl32.Vec3{0, 0, 0}, mgl32.Vec3{1, 0.5, 1})}
}

// EncodeItem ...
func (w WoodSlab) EncodeItem() (id int32, meta int16) {
	switch w.Wood {
	case material.OakWood():
		if w.Double {
			return 157, 0
		}
		return 158, 0
	case material.SpruceWood():
		if w.Double {
			return 157, 1
		}
		return 158, 1
	case material.BirchWood():
		if w.Double {
			return 157, 2
		}
		return 158, 2
	case material.JungleWood():
		if w.Double {
			return 157, 3
		}
		return 158, 3
	case material.AcaciaWood():
		if w.Double {
			return 157, 4
		}
		return 158, 4
	case material.DarkOakWood():
		if w.Double {
			return 157, 5
		}
		return 158, 5
	}
	panic("invalid wood type")
}

// EncodeBlock ...
func (w WoodSlab) EncodeBlock() (name string, properties map[string]interface{}) {
	if w.Double {
		return "double_wooden_slab", map[string]interface{}{"top_slot_bit": w.UpsideDown, "wood_type": w.Wood.String()}
	}
	return "wooden_slab", map[string]interface{}{"top_slot_bit": w.UpsideDown, "wood_type": w.Wood.String()}
}

// allWoodSlabs returns all states of wood slabs.
func allWoodSlabs() (stairs []world.Block) {
	f := func(double bool, upsideDown bool) {
		stairs = append(stairs, WoodSlab{Double: double, UpsideDown: upsideDown, Wood: material.OakWood()})
		stairs = append(stairs, WoodSlab{Double: double, UpsideDown: upsideDown, Wood: material.SpruceWood()})
		stairs = append(stairs, WoodSlab{Double: double, UpsideDown: upsideDown, Wood: material.BirchWood()})
		stairs = append(stairs, WoodSlab{Double: double, UpsideDown: upsideDown, Wood: material.JungleWood()})
		stairs = append(stairs, WoodSlab{Double: double, UpsideDown: upsideDown, Wood: material.AcaciaWood()})
		stairs = append(stairs, WoodSlab{Double: double, UpsideDown: upsideDown, Wood: material.DarkOakWood()})
	}
	f(false, false)
	f(false, true)
	f(true, false)
	f(true, true)
	return
}
