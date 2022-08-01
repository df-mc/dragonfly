package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/model"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/go-gl/mathgl/mgl64"
	"time"
)

// Slab is a half block that allows entities to walk up blocks without jumping.
type Slab struct {
	// Block is the block to use for the type of slab.
	Block world.Block
	// Top specifies if the slab is in the top part of the block.
	Top bool
	// Double specifies if the slab is a double slab. These double slabs can be made by placing another slab
	// on an existing slab.
	Double bool
}

// UseOnBlock handles the placement of slabs with relation to them being upside down or not and handles slabs
// being turned into double slabs.
func (s Slab) UseOnBlock(pos cube.Pos, face cube.Face, clickPos mgl64.Vec3, w *world.World, user item.User, ctx *item.UseContext) (used bool) {
	id, meta := s.EncodeItem()
	clickedBlock := w.Block(pos)
	if clickedSlab, ok := clickedBlock.(Slab); ok && !s.Double {
		clickedId, clickedMeta := clickedSlab.EncodeItem()
		if !clickedSlab.Double && id == clickedId && meta == clickedMeta && ((face == cube.FaceUp && !clickedSlab.Top) || (face == cube.FaceDown && clickedSlab.Top)) {
			// A half slab of the same type was clicked at the top, so we can make it full.
			clickedSlab.Double = true

			place(w, pos, clickedSlab, user, ctx)
			return placed(ctx)
		}
	}
	if sideSlab, ok := w.Block(pos.Side(face)).(Slab); ok && !replaceableWith(w, pos, s) && !s.Double {
		sideId, sideMeta := sideSlab.EncodeItem()
		// The block on the side of the one clicked was a slab and the block clicked was not replaceableWith, so
		// the slab on the side must've been half and may now be filled if the wood types are the same.
		if !sideSlab.Double && id == sideId && meta == sideMeta {
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

// Instrument ...
func (s Slab) Instrument() sound.Instrument {
	if _, ok := s.Block.(Planks); ok {
		return sound.Bass()
	}
	return sound.BassDrum()
}

// FlammabilityInfo ...
func (s Slab) FlammabilityInfo() FlammabilityInfo {
	if w, ok := s.Block.(Planks); ok && w.Wood.Flammable() {
		return newFlammabilityInfo(5, 20, true)
	}
	return newFlammabilityInfo(0, 0, false)
}

// FuelInfo ...
func (s Slab) FuelInfo() item.FuelInfo {
	if w, ok := s.Block.(Planks); ok && w.Wood.Flammable() {
		return newFuelInfo(time.Second * 15)
	}
	return item.FuelInfo{}
}

// CanDisplace ...
func (s Slab) CanDisplace(b world.Liquid) bool {
	water, ok := b.(Water)
	return !s.Double && ok && water.Depth == 8
}

// SideClosed ...
func (s Slab) SideClosed(pos, side cube.Pos, _ *world.World) bool {
	// Only returns true if the side is below the slab and if the slab is not upside down.
	return !s.Top && side[1] == pos[1]-1
}

// LightDiffusionLevel returns 0 if the slab is a half slab, or 15 if it is double.
func (s Slab) LightDiffusionLevel() uint8 {
	if s.Double {
		return 15
	}
	return 0
}

// BreakInfo ...
func (s Slab) BreakInfo() BreakInfo {
	hardness, blastResistance, harvestable, effective := 2.0, 30.0, pickaxeHarvestable, pickaxeEffective

	switch block := s.Block.(type) {
	// TODO: Copper
	// TODO: Deepslate
	case EndBricks:
		hardness = 3.0
		blastResistance = 45.0
	case StoneBricks:
		if block.Type == MossyStoneBricks() {
			hardness = 1.5
		}
	case Andesite:
		if block.Polished {
			hardness = 1.5
		}
	case Diorite:
		if block.Polished {
			hardness = 1.5
		}
	case Granite:
		if block.Polished {
			hardness = 1.5
		}
	case Prismarine:
		hardness = 1.5
	case Planks:
		harvestable = alwaysHarvestable
		effective = axeEffective
		blastResistance = 15.0
	}
	return newBreakInfo(hardness, harvestable, effective, func(tool item.Tool, enchantments []item.Enchantment) []item.Stack {
		if s.Double {
			return []item.Stack{item.NewStack(s, 2)}
		}
		return []item.Stack{item.NewStack(s, 1)}
	}).withBlastResistance(blastResistance)
}

// Model ...
func (s Slab) Model() world.BlockModel {
	return model.Slab{Double: s.Double, Top: s.Top}
}

// EncodeItem ...
func (s Slab) EncodeItem() (string, int16) {
	id, slabType, meta := encodeSlabBlock(s.Block)
	if slabType != "" {
		return "minecraft:" + encodeLegacySlabId(slabType), meta
	}
	return "minecraft:" + id + "_slab", meta
}

// EncodeBlock ...
func (s Slab) EncodeBlock() (string, map[string]any) {
	id, slabType, _ := encodeSlabBlock(s.Block)
	properties := map[string]any{"top_slot_bit": s.Top}
	if slabType != "" {
		properties[slabType] = id
		id = encodeLegacySlabId(slabType)
		if s.Double {
			id = "double_" + id
		}
	} else if s.Double {
		id = id + "_double_slab"
	} else {
		id = id + "_slab"
	}
	return "minecraft:" + id, properties
}

// allSlabs ...
func allSlabs() (b []world.Block) {
	for _, s := range SlabBlocks() {
		b = append(b, Slab{Block: s, Double: true})
		b = append(b, Slab{Block: s, Top: true, Double: true})
		b = append(b, Slab{Block: s, Top: true})
		b = append(b, Slab{Block: s})
	}
	return
}
