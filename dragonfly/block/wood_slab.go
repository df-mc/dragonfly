package block

import (
	"github.com/df-mc/dragonfly/dragonfly/block/wood"
	"github.com/df-mc/dragonfly/dragonfly/entity/physics"
	"github.com/df-mc/dragonfly/dragonfly/item"
	"github.com/df-mc/dragonfly/dragonfly/item/tool"
	"github.com/df-mc/dragonfly/dragonfly/world"
	"github.com/go-gl/mathgl/mgl64"
)

// WoodSlab is a half block that allows entities to walk up blocks without jumping.
type WoodSlab struct {
	// Wood is the type of wood of the slabs. This field must have one of the values found in the material
	// package.
	Wood wood.Wood
	// UpsideDown specifies if the slabs are upside down.
	UpsideDown bool
	// Double specifies if the slab is a double slab. These double slabs can be made by placing another slab
	// on an existing slab.
	Double bool
}

// UseOnBlock handles the placement of slabs with relation to them being upside down or not and handles slabs
// being turned into double slabs.
func (s WoodSlab) UseOnBlock(pos world.BlockPos, face world.Face, clickPos mgl64.Vec3, w *world.World, user item.User, ctx *item.UseContext) (used bool) {
	clickedBlock := w.Block(pos)
	if clickedSlab, ok := clickedBlock.(WoodSlab); ok && !s.Double {
		if (face == world.FaceUp && !clickedSlab.Double && clickedSlab.Wood == s.Wood && !clickedSlab.UpsideDown) ||
			(face == world.FaceDown && !clickedSlab.Double && clickedSlab.Wood == s.Wood && clickedSlab.UpsideDown) {
			// A half slab of the same type was clicked at the top, so we can make it full.
			clickedSlab.Double = true

			place(w, pos, clickedSlab, user, ctx)
			return placed(ctx)
		}
	}
	if sideSlab, ok := w.Block(pos.Side(face)).(WoodSlab); ok && !replaceable(w, pos, s) && !s.Double {
		// The block on the side of the one clicked was a slab and the block clicked was not replaceable, so
		// the slab on the side must've been half and may now be filled if the wood types are the same.
		if !sideSlab.Double && sideSlab.Wood == s.Wood {
			sideSlab.Double = true

			place(w, pos.Side(face), sideSlab, user, ctx)
			return placed(ctx)
		}
	}
	pos, face, used = firstReplaceable(w, pos, face, s)
	if !used {
		return
	}
	if face == world.FaceDown || (clickPos[1] > 0.5 && face != world.FaceUp) {
		s.UpsideDown = true
	}

	place(w, pos, s, user, ctx)
	return placed(ctx)
}

// BreakInfo ...
func (s WoodSlab) BreakInfo() BreakInfo {
	return BreakInfo{
		Hardness:    2,
		Harvestable: alwaysHarvestable,
		Effective:   axeEffective,
		Drops: func(t tool.Tool) []item.Stack {
			if s.Double {
				s.Double = false
				// If the slab is double, it should drop two single slabs.
				return []item.Stack{item.NewStack(s, 2)}
			}
			return []item.Stack{item.NewStack(s, 1)}
		},
	}
}

// LightDiffusionLevel returns 0 if the slab is a half slab, or 15 if it is double.
func (s WoodSlab) LightDiffusionLevel() uint8 {
	if s.Double {
		return 15
	}
	return 0
}

// AABB ...
func (s WoodSlab) AABB(world.BlockPos, *world.World) []physics.AABB {
	if s.Double {
		return []physics.AABB{physics.NewAABB(mgl64.Vec3{}, mgl64.Vec3{1, 1, 1})}
	}
	if s.UpsideDown {
		return []physics.AABB{physics.NewAABB(mgl64.Vec3{0, 0.5, 0}, mgl64.Vec3{1, 1, 1})}
	}
	return []physics.AABB{physics.NewAABB(mgl64.Vec3{}, mgl64.Vec3{1, 0.5, 1})}
}

// EncodeItem ...
func (s WoodSlab) EncodeItem() (id int32, meta int16) {
	switch s.Wood {
	case wood.Oak():
		if s.Double {
			return 157, 0
		}
		return 158, 0
	case wood.Spruce():
		if s.Double {
			return 157, 1
		}
		return 158, 1
	case wood.Birch():
		if s.Double {
			return 157, 2
		}
		return 158, 2
	case wood.Jungle():
		if s.Double {
			return 157, 3
		}
		return 158, 3
	case wood.Acacia():
		if s.Double {
			return 157, 4
		}
		return 158, 4
	case wood.DarkOak():
		if s.Double {
			return 157, 5
		}
		return 158, 5
	}
	panic("invalid wood type")
}

// EncodeBlock ...
func (s WoodSlab) EncodeBlock() (name string, properties map[string]interface{}) {
	if s.Double {
		return "minecraft:double_wooden_slab", map[string]interface{}{"top_slot_bit": s.UpsideDown, "wood_type": s.Wood.String()}
	}
	return "minecraft:wooden_slab", map[string]interface{}{"top_slot_bit": s.UpsideDown, "wood_type": s.Wood.String()}
}

// CanDisplace ...
func (s WoodSlab) CanDisplace(b world.Liquid) bool {
	_, ok := b.(Water)
	return !s.Double && ok
}

// SideClosed ...
func (s WoodSlab) SideClosed(pos, side world.BlockPos, w *world.World) bool {
	// Only returns true if the side is below the slab and if the slab is not upside down.
	return !s.UpsideDown && side[1] == pos[1]-1
}

// FaceSolidTo ...
func (s WoodSlab) FaceSolidTo(_ world.BlockPos, _ world.Face, _ world.Block) bool {
	return s.Double
}

// allWoodSlabs returns all states of wood slabs.
func allWoodSlabs() (slabs []world.Block) {
	f := func(double bool, upsideDown bool) {
		slabs = append(slabs, WoodSlab{Double: double, UpsideDown: upsideDown, Wood: wood.Oak()})
		slabs = append(slabs, WoodSlab{Double: double, UpsideDown: upsideDown, Wood: wood.Spruce()})
		slabs = append(slabs, WoodSlab{Double: double, UpsideDown: upsideDown, Wood: wood.Birch()})
		slabs = append(slabs, WoodSlab{Double: double, UpsideDown: upsideDown, Wood: wood.Jungle()})
		slabs = append(slabs, WoodSlab{Double: double, UpsideDown: upsideDown, Wood: wood.Acacia()})
		slabs = append(slabs, WoodSlab{Double: double, UpsideDown: upsideDown, Wood: wood.DarkOak()})
	}
	f(false, false)
	f(false, true)
	f(true, false)
	f(true, true)
	return
}
