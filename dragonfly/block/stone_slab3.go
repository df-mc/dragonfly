package block

import (
	"github.com/df-mc/dragonfly/dragonfly/block/model"
	"github.com/df-mc/dragonfly/dragonfly/block/stone_slabs"
	"github.com/df-mc/dragonfly/dragonfly/entity/physics"
	"github.com/df-mc/dragonfly/dragonfly/item"
	"github.com/df-mc/dragonfly/dragonfly/item/tool"
	"github.com/df-mc/dragonfly/dragonfly/world"
	"github.com/go-gl/mathgl/mgl64"
)

// StoneSlab3 is a half block that allows entities to walk up blocks without jumping.
type StoneSlab3 struct {
	noNBT

	// StoneSlab3 is the type of stone of the slabs. This field must have one of the values found in the material
	// package.
	StoneSlab3 stone_slabs.StoneSlab3
	// Top specifies if the slab is in the top part of the block.
	Top bool
	// Double specifies if the slab is a double slab. These double slabs can be made by placing another slab
	// on an existing slab.
	Double bool
}

// Model ...
func (s StoneSlab3) Model() world.BlockModel {
	return model.Slab{Double: s.Double, Top: s.Top}
}

// UseOnBlock handles the placement of slabs with relation to them being upside down or not and handles slabs
// being turned into double slabs.
func (s StoneSlab3) UseOnBlock(pos world.BlockPos, face world.Face, clickPos mgl64.Vec3, w *world.World, user item.User, ctx *item.UseContext) (used bool) {
	clickedBlock := w.Block(pos)
	if clickedSlab, ok := clickedBlock.(StoneSlab3); ok && !s.Double {
		if (face == world.FaceUp && !clickedSlab.Double && clickedSlab.StoneSlab3 == s.StoneSlab3 && !clickedSlab.Top) ||
			(face == world.FaceDown && !clickedSlab.Double && clickedSlab.StoneSlab3 == s.StoneSlab3 && clickedSlab.Top) {
			// A half slab of the same type was clicked at the top, so we can make it full.
			clickedSlab.Double = true

			place(w, pos, clickedSlab, user, ctx)
			return placed(ctx)
		}
	}
	if sideSlab, ok := w.Block(pos.Side(face)).(StoneSlab3); ok && !replaceableWith(w, pos, s) && !s.Double {
		// The block on the side of the one clicked was a slab and the block clicked was not replaceableWith, so
		// the slab on the side must've been half and may now be filled if the wood types are the same.
		if !sideSlab.Double && sideSlab.StoneSlab3 == s.StoneSlab3 {
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
		s.Top = true
	}

	place(w, pos, s, user, ctx)
	return placed(ctx)
}

// BreakInfo ...
func (s StoneSlab3) BreakInfo() BreakInfo {
	switch s.StoneSlab3 { // todo: well... don't know if you like this. either this or a really long if or?
	case stone_slabs.EndStoneBrick():
		return BreakInfo{
			Hardness:    3,
			Harvestable: pickaxeHarvestable,
			Effective:   pickaxeEffective,
			Drops: func(t tool.Tool) []item.Stack {
				if s.Double {
					s.Double = false
					// If the slab is double, it should drop two single slabs.
					return []item.Stack{item.NewStack(s, 2)}
				}
				return []item.Stack{item.NewStack(s, 1)}
			},
		}
	case stone_slabs.Andesite():
	case stone_slabs.PolishedAndesite():
	case stone_slabs.Diorite():
	case stone_slabs.PolishedDiorite():
	case stone_slabs.Granite():
	case stone_slabs.PolishedGranite():
		return BreakInfo{
			Hardness:    1.5,
			Harvestable: pickaxeHarvestable,
			Effective:   pickaxeEffective,
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
	return BreakInfo{
		Hardness:    2,
		Harvestable: pickaxeHarvestable,
		Effective:   pickaxeEffective,
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
func (s StoneSlab3) LightDiffusionLevel() uint8 {
	if s.Double {
		return 15
	}
	return 0
}

// AABB ...
func (s StoneSlab3) AABB(world.BlockPos, *world.World) []physics.AABB {
	if s.Double {
		return []physics.AABB{physics.NewAABB(mgl64.Vec3{}, mgl64.Vec3{1, 1, 1})}
	}
	if s.Top {
		return []physics.AABB{physics.NewAABB(mgl64.Vec3{0, 0.5, 0}, mgl64.Vec3{1, 1, 1})}
	}
	return []physics.AABB{physics.NewAABB(mgl64.Vec3{}, mgl64.Vec3{1, 0.5, 1})}
}

// EncodeItem ...
func (s StoneSlab3) EncodeItem() (id int32, meta int16) {
	switch s.StoneSlab3 {
	case stone_slabs.EndStoneBrick():
		if s.Double {
			return -167, 0
		}
		return -162, 0
	case stone_slabs.SmoothRedSandstone():
		if s.Double {
			return -167, 1
		}
		return -162, 1
	case stone_slabs.PolishedAndesite():
		if s.Double {
			return -167, 2
		}
		return -162, 2
	case stone_slabs.Andesite():
		if s.Double {
			return -167, 3
		}
		return -162, 3
	case stone_slabs.Diorite():
		if s.Double {
			return -167, 4
		}
		return -162, 4
	case stone_slabs.PolishedDiorite():
		if s.Double {
			return -167, 5
		}
		return -162, 5
	case stone_slabs.Granite():
		if s.Double {
			return -167, 6
		}
		return -162, 6
	case stone_slabs.PolishedGranite():
		if s.Double {
			return -167, 7
		}
		return -162, 7
	}
	panic("invalid stone type")
}

// EncodeBlock ...
func (s StoneSlab3) EncodeBlock() (name string, properties map[string]interface{}) {
	if s.Double {
		return "minecraft:double_stone_slab3", map[string]interface{}{"top_slot_bit": s.Top, "stone_slab_type_3": s.StoneSlab3.String()}
	}
	return "minecraft:stone_slab3", map[string]interface{}{"top_slot_bit": s.Top, "stone_slab_type_3": s.StoneSlab3.String()}
}

// Hash ...
func (s StoneSlab3) Hash() uint64 {
	return hashStoneSlab3 | (uint64(boolByte(s.Top)) << 32) | (uint64(boolByte(s.Double)) << 33) | (uint64(s.StoneSlab3.Uint8()) << 34)
}

// CanDisplace ...
func (s StoneSlab3) CanDisplace(b world.Liquid) bool {
	_, ok := b.(Water)
	return !s.Double && ok
}

// SideClosed ...
func (s StoneSlab3) SideClosed(pos, side world.BlockPos, _ *world.World) bool {
	// Only returns true if the side is below the slab and if the slab is not upside down.
	return !s.Top && side[1] == pos[1]-1
}

// allStoneSlabs3 returns all states of wood slabs.
func allStoneSlabs3() (slabs []world.Block) {
	f := func(double bool, upsideDown bool) {
		slabs = append(slabs, StoneSlab3{Double: double, Top: upsideDown, StoneSlab3: stone_slabs.EndStoneBrick()})
		slabs = append(slabs, StoneSlab3{Double: double, Top: upsideDown, StoneSlab3: stone_slabs.SmoothRedSandstone()})
		slabs = append(slabs, StoneSlab3{Double: double, Top: upsideDown, StoneSlab3: stone_slabs.PolishedAndesite()})
		slabs = append(slabs, StoneSlab3{Double: double, Top: upsideDown, StoneSlab3: stone_slabs.Andesite()})
		slabs = append(slabs, StoneSlab3{Double: double, Top: upsideDown, StoneSlab3: stone_slabs.Diorite()})
		slabs = append(slabs, StoneSlab3{Double: double, Top: upsideDown, StoneSlab3: stone_slabs.PolishedDiorite()})
		slabs = append(slabs, StoneSlab3{Double: double, Top: upsideDown, StoneSlab3: stone_slabs.Granite()})
		slabs = append(slabs, StoneSlab3{Double: double, Top: upsideDown, StoneSlab3: stone_slabs.PolishedGranite()})
	}
	f(false, false)
	f(false, true)
	f(true, false)
	f(true, true)
	return
}
