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

// StoneSlab2 is a half block that allows entities to walk up blocks without jumping.
type StoneSlab2 struct {
	noNBT

	// StoneSlab2 is the type of stone of the slabs. This field must have one of the values found in the material
	// package.
	StoneSlab2 stone_slabs.StoneSlab2
	// Top specifies if the slab is in the top part of the block.
	Top bool
	// Double specifies if the slab is a double slab. These double slabs can be made by placing another slab
	// on an existing slab.
	Double bool
}

// Model ...
func (s StoneSlab2) Model() world.BlockModel {
	return model.Slab{Double: s.Double, Top: s.Top}
}

// UseOnBlock handles the placement of slabs with relation to them being upside down or not and handles slabs
// being turned into double slabs.
func (s StoneSlab2) UseOnBlock(pos world.BlockPos, face world.Face, clickPos mgl64.Vec3, w *world.World, user item.User, ctx *item.UseContext) (used bool) {
	clickedBlock := w.Block(pos)
	if clickedSlab, ok := clickedBlock.(StoneSlab2); ok && !s.Double {
		if (face == world.FaceUp && !clickedSlab.Double && clickedSlab.StoneSlab2 == s.StoneSlab2 && !clickedSlab.Top) ||
			(face == world.FaceDown && !clickedSlab.Double && clickedSlab.StoneSlab2 == s.StoneSlab2 && clickedSlab.Top) {
			// A half slab of the same type was clicked at the top, so we can make it full.
			clickedSlab.Double = true

			place(w, pos, clickedSlab, user, ctx)
			return placed(ctx)
		}
	}
	if sideSlab, ok := w.Block(pos.Side(face)).(StoneSlab2); ok && !replaceableWith(w, pos, s) && !s.Double {
		// The block on the side of the one clicked was a slab and the block clicked was not replaceableWith, so
		// the slab on the side must've been half and may now be filled if the wood types are the same.
		if !sideSlab.Double && sideSlab.StoneSlab2 == s.StoneSlab2 {
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
func (s StoneSlab2) BreakInfo() BreakInfo {
	if s.StoneSlab2 == stone_slabs.PrismarineRough() || s.StoneSlab2 == stone_slabs.PrismarineDark() || s.StoneSlab2 == stone_slabs.PrismarineBrick() {
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
func (s StoneSlab2) LightDiffusionLevel() uint8 {
	if s.Double {
		return 15
	}
	return 0
}

// AABB ...
func (s StoneSlab2) AABB(world.BlockPos, *world.World) []physics.AABB {
	if s.Double {
		return []physics.AABB{physics.NewAABB(mgl64.Vec3{}, mgl64.Vec3{1, 1, 1})}
	}
	if s.Top {
		return []physics.AABB{physics.NewAABB(mgl64.Vec3{0, 0.5, 0}, mgl64.Vec3{1, 1, 1})}
	}
	return []physics.AABB{physics.NewAABB(mgl64.Vec3{}, mgl64.Vec3{1, 0.5, 1})}
}

// EncodeItem ...
func (s StoneSlab2) EncodeItem() (id int32, meta int16) {
	switch s.StoneSlab2 {
	case stone_slabs.RedSandstone():
		if s.Double {
			return 181, 0
		}
		return 182, 0
	case stone_slabs.Purpur():
		if s.Double {
			return 181, 1
		}
		return 182, 1
	case stone_slabs.PrismarineRough():
		if s.Double {
			return 181, 2
		}
		return 182, 2
	case stone_slabs.PrismarineDark():
		if s.Double {
			return 181, 3
		}
		return 182, 3
	case stone_slabs.PrismarineBrick():
		if s.Double {
			return 181, 4
		}
		return 182, 4
	case stone_slabs.MossyCobblestone():
		if s.Double {
			return 181, 5
		}
		return 182, 5
	case stone_slabs.SmoothSandstone():
		if s.Double {
			return 181, 6
		}
		return 182, 6
	case stone_slabs.RedNetherBrick():
		if s.Double {
			return 181, 7
		}
		return 182, 7
	}
	panic("invalid stone type")
}

// EncodeBlock ...
func (s StoneSlab2) EncodeBlock() (name string, properties map[string]interface{}) {
	if s.Double {
		return "minecraft:double_stone_slab2", map[string]interface{}{"top_slot_bit": s.Top, "stone_slab_type_2": s.StoneSlab2.String()}
	}
	return "minecraft:stone_slab2", map[string]interface{}{"top_slot_bit": s.Top, "stone_slab_type_2": s.StoneSlab2.String()}
}

// Hash ...
func (s StoneSlab2) Hash() uint64 {
	return hashStoneSlab2 | (uint64(boolByte(s.Top)) << 32) | (uint64(boolByte(s.Double)) << 33) | (uint64(s.StoneSlab2.Uint8()) << 34)
}

// CanDisplace ...
func (s StoneSlab2) CanDisplace(b world.Liquid) bool {
	_, ok := b.(Water)
	return !s.Double && ok
}

// SideClosed ...
func (s StoneSlab2) SideClosed(pos, side world.BlockPos, _ *world.World) bool {
	// Only returns true if the side is below the slab and if the slab is not upside down.
	return !s.Top && side[1] == pos[1]-1
}

// allStoneSlabs2 returns all states of wood slabs.
func allStoneSlabs2() (slabs []world.Block) {
	f := func(double bool, upsideDown bool) {
		slabs = append(slabs, StoneSlab2{Double: double, Top: upsideDown, StoneSlab2: stone_slabs.RedSandstone()})
		slabs = append(slabs, StoneSlab2{Double: double, Top: upsideDown, StoneSlab2: stone_slabs.Purpur()})
		slabs = append(slabs, StoneSlab2{Double: double, Top: upsideDown, StoneSlab2: stone_slabs.PrismarineRough()})
		slabs = append(slabs, StoneSlab2{Double: double, Top: upsideDown, StoneSlab2: stone_slabs.PrismarineDark()})
		slabs = append(slabs, StoneSlab2{Double: double, Top: upsideDown, StoneSlab2: stone_slabs.PrismarineBrick()})
		slabs = append(slabs, StoneSlab2{Double: double, Top: upsideDown, StoneSlab2: stone_slabs.MossyCobblestone()})
		slabs = append(slabs, StoneSlab2{Double: double, Top: upsideDown, StoneSlab2: stone_slabs.SmoothSandstone()})
		slabs = append(slabs, StoneSlab2{Double: double, Top: upsideDown, StoneSlab2: stone_slabs.RedNetherBrick()})
	}
	f(false, false)
	f(false, true)
	f(true, false)
	f(true, true)
	return
}
