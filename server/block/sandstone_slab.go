package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/model"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// SandstoneSlab are blocks that allow entities to walk up blocks without jumping. They are crafted using sandstone.
type SandstoneSlab struct {
	transparent

	// Type is the type of sandstone of the block.
	Type SandstoneType

	// Red specifies if the sandstone type is red or not. When set to true, the sandstone slab type will represent its
	// red variant, for example red sandstone slab.
	Red bool

	// Top specifies if the slab is in the top part of the block.
	Top bool
	// Double specifies if the slab is a double slab. These double slabs can be made by placing another slab
	// on an existing slab.
	Double bool
}

// UseOnBlock handles the directional placing of slab and makes sure they are properly placed upside down
// when needed.
func (s SandstoneSlab) UseOnBlock(pos cube.Pos, face cube.Face, clickPos mgl64.Vec3, w *world.World, user item.User, ctx *item.UseContext) (used bool) {
	clickedBlock := w.Block(pos)
	if clickedSlab, ok := clickedBlock.(SandstoneSlab); ok && !s.Double {
		if (face == cube.FaceUp && !clickedSlab.Double && clickedSlab.Type == s.Type && clickedSlab.Red == s.Red && !clickedSlab.Top) ||
			(face == cube.FaceDown && !clickedSlab.Double && clickedSlab.Type == s.Type && clickedSlab.Red == s.Red && clickedSlab.Top) {
			// A half slab of the same type was clicked at the top, so we can make it full.
			clickedSlab.Double = true

			place(w, pos, clickedSlab, user, ctx)
			return placed(ctx)
		}
	}
	if sideSlab, ok := w.Block(pos.Side(face)).(SandstoneSlab); ok && !replaceableWith(w, pos, s) && !s.Double {
		// The block on the side of the one clicked was a slab and the block clicked was not replaceableWith, so
		// the slab on the side must've been half and may now be filled if the stone types are the same.
		if !sideSlab.Double && sideSlab.Type == s.Type && sideSlab.Red == s.Red {
			sideSlab.Double = true

			place(w, pos.Side(face), sideSlab, user, ctx)
			return placed(ctx)
		}
	}
	pos, face, used = firstReplaceable(w, pos, face, s)
	if !used {
		return
	}
	if face == cube.FaceDown || (clickPos[1] > 0.5 && face != cube.FaceUp) {
		s.Top = true
	}

	place(w, pos, s, user, ctx)
	return placed(ctx)
}

// Model ...
func (s SandstoneSlab) Model() world.BlockModel {
	return model.Slab{Double: s.Double, Top: s.Top}
}

// BreakInfo ...
func (s SandstoneSlab) BreakInfo() BreakInfo {
	return newBreakInfo(s.Type.Hardness(), alwaysHarvestable, axeEffective, func(item.Tool, []item.Enchantment) []item.Stack {
		if s.Double {
			s.Double = false
			// If the slab is double, it should drop two single slabs.
			return []item.Stack{item.NewStack(s, 2)}
		}
		return []item.Stack{item.NewStack(s, 1)}
	})
}

// EncodeItem ...
func (s SandstoneSlab) EncodeItem() (name string, meta int16) {
	prefix := "minecraft:stone_block_slab"
	if s.Double {
		prefix = "minecraft:double_stone_block_slab"
	}

	if s.Type == CutSandstone() {
		if s.Red {
			return prefix + "4", 1
		}
		return prefix + "4", 0
	} else if s.Type == SmoothSandstone() {
		if s.Red {
			return prefix + "3", 1
		}
		return prefix + "2", 1
	}
	if s.Red {
		return prefix + "2", 0
	}
	return prefix, 0
}

// EncodeBlock ...
func (s SandstoneSlab) EncodeBlock() (name string, properties map[string]any) {
	prefix := "minecraft:stone_block_slab"
	if s.Double {
		prefix = "minecraft:double_stone_block_slab"
	}

	if s.Type == CutSandstone() {
		if s.Red {
			return prefix + "4", map[string]any{"top_slot_bit": s.Top, "stone_slab_type_4": "cut_red_sandstone"}
		}
		return prefix + "4", map[string]any{"top_slot_bit": s.Top, "stone_slab_type_4": "cut_sandstone"}
	} else if s.Type == SmoothSandstone() {
		if s.Red {
			return prefix + "3", map[string]any{"top_slot_bit": s.Top, "stone_slab_type_3": "smooth_red_sandstone"}
		}
		return prefix + "2", map[string]any{"top_slot_bit": s.Top, "stone_slab_type_2": "smooth_sandstone"}
	}
	if s.Red {
		return prefix + "2", map[string]any{"top_slot_bit": s.Top, "stone_slab_type_2": "red_sandstone"}
	}
	return prefix, map[string]any{"top_slot_bit": s.Top, "stone_slab_type": "sandstone"}

}

// CanDisplace ...
func (s SandstoneSlab) CanDisplace(b world.Liquid) bool {
	_, ok := b.(Water)
	return !s.Double && ok
}

// SideClosed ...
func (s SandstoneSlab) SideClosed(pos, side cube.Pos, w *world.World) bool {
	return !s.Top && side[1] == pos[1]-1
}

// allSandstoneSlab ...
func allSandstoneSlabs() (slabs []world.Block) {
	f := func(red bool) {
		for _, t := range SandstoneTypes() {
			if t.SlabAble() {
				slabs = append(slabs, SandstoneSlab{Double: false, Top: false, Type: t, Red: red})
				slabs = append(slabs, SandstoneSlab{Double: false, Top: true, Type: t, Red: red})
				slabs = append(slabs, SandstoneSlab{Double: true, Top: false, Type: t, Red: red})
				slabs = append(slabs, SandstoneSlab{Double: true, Top: true, Type: t, Red: red})
			}
		}
	}
	f(true)
	f(false)
	return
}
