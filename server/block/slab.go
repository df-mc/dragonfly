package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/model"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/go-gl/mathgl/mgl64"
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
func (s Slab) UseOnBlock(pos cube.Pos, face cube.Face, clickPos mgl64.Vec3, tx *world.Tx, user item.User, ctx *item.UseContext) (used bool) {
	id, meta := s.EncodeItem()
	clickedBlock := tx.Block(pos)
	if clickedSlab, ok := clickedBlock.(Slab); ok && !s.Double {
		clickedId, clickedMeta := clickedSlab.EncodeItem()
		if !clickedSlab.Double && id == clickedId && meta == clickedMeta && ((face == cube.FaceUp && !clickedSlab.Top) || (face == cube.FaceDown && clickedSlab.Top)) {
			// A half slab of the same type was clicked at the top, so we can make it full.
			clickedSlab.Double = true

			place(tx, pos, clickedSlab, user, ctx)
			return placed(ctx)
		}
	}
	if sideSlab, ok := tx.Block(pos.Side(face)).(Slab); ok && !replaceableWith(tx, pos, s) && !s.Double {
		sideId, sideMeta := sideSlab.EncodeItem()
		// The block on the side of the one clicked was a slab and the block clicked was not replaceableWith, so
		// the slab on the side must've been half and may now be filled if the wood types are the same.
		if !sideSlab.Double && id == sideId && meta == sideMeta {
			sideSlab.Double = true

			place(tx, pos.Side(face), sideSlab, user, ctx)
			return placed(ctx)
		}
	}
	pos, face, used = firstReplaceable(tx, pos, face, s)
	if !used {
		return
	}
	if face == cube.FaceDown || (clickPos[1] > 0.5 && face != cube.FaceUp) {
		s.Top = true
	}

	place(tx, pos, s, user, ctx)
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
	if flammable, ok := s.Block.(Flammable); ok {
		return flammable.FlammabilityInfo()
	}
	return newFlammabilityInfo(0, 0, false)
}

// FuelInfo ...
func (s Slab) FuelInfo() item.FuelInfo {
	if fuel, ok := s.Block.(item.Fuel); ok {
		return fuel.FuelInfo()
	}
	return item.FuelInfo{}
}

// CanDisplace ...
func (s Slab) CanDisplace(b world.Liquid) bool {
	water, ok := b.(Water)
	return !s.Double && ok && water.Depth == 8
}

// SideClosed ...
func (s Slab) SideClosed(pos, side cube.Pos, _ *world.Tx) bool {
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
	case Stone, Sandstone, Quartz, Purpur, Blackstone, PolishedBlackstoneBrick:
	// These slab types do not match their block's hardness or blast resistance
	case StoneBricks:
		if block.Type == MossyStoneBricks() {
			hardness = 1.5
		}
	case Breakable:
		breakInfo := block.BreakInfo()
		hardness, blastResistance, harvestable, effective = breakInfo.Hardness, breakInfo.BlastResistance, breakInfo.Harvestable, breakInfo.Effective
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
	name, suffix := encodeSlabBlock(s.Block, false)
	return "minecraft:" + name + suffix, 0
}

// EncodeBlock ...
func (s Slab) EncodeBlock() (string, map[string]any) {
	side := "bottom"
	if s.Top {
		side = "top"
	}
	name, suffix := encodeSlabBlock(s.Block, s.Double)
	return "minecraft:" + name + suffix, map[string]any{"minecraft:vertical_half": side}
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
