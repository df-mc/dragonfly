package block

import (
	"github.com/df-mc/dragonfly/server/world"
)

// encodeSlabBlock encodes the provided block in to an identifier and meta value that can be used to encode the slab.
func encodeSlabBlock(block world.Block) (id, slabType string, meta int16) {
	switch block := block.(type) {
	case Planks:
		switch block.Wood {
		case OakWood(), SpruceWood(), BirchWood(), JungleWood(), AcaciaWood(), DarkOakWood():
			return block.Wood.String(), "wood_type", int16(block.Wood.Uint8())
		default:
			return block.Wood.String(), "", 0
		}
	case Stone:
		if block.Smooth {
			return "smooth_stone", "stone_slab_type", 0
		}
		return "stone", "stone_slab_type_4", 2
	case Cobblestone:
		if block.Mossy {
			return "mossy_cobblestone", "stone_slab_type_2", 5
		}
		return "cobblestone", "stone_slab_type", 3
	case StoneBricks:
		if block.Type == MossyStoneBricks() {
			return "mossy_stone_brick", "stone_slab_type_4", 0
		}
		return "stone_brick", "stone_slab_type", 5
	case Andesite:
		if block.Polished {
			return "polished_andesite", "stone_slab_type_3", 2
		}
		return "andesite", "stone_slab_type_3", 3
	case Diorite:
		if block.Polished {
			return "polished_diorite", "stone_slab_type_3", 5
		}
		return "diorite", "stone_slab_type_3", 4
	case Granite:
		if block.Polished {
			return "polished_granite", "stone_slab_type_3", 7
		}
		return "granite", "stone_slab_type_3", 6
	case Sandstone:
		switch block.Type {
		case NormalSandstone():
			if block.Red {
				return "red_sandstone", "stone_slab_type_2", 0
			}
			return "sandstone", "stone_slab_type", 1
		case CutSandstone():
			if block.Red {
				return "cut_red_sandstone", "stone_slab_type_4", 4
			}
			return "cut_sandstone", "stone_slab_type_4", 3
		case SmoothSandstone():
			if block.Red {
				return "smooth_red_sandstone", "stone_slab_type_3", 1
			}
			return "smooth_sandstone", "stone_slab_type_2", 6
		}
		panic("invalid sandstone type")
	case Bricks:
		return "brick", "stone_slab_type", 4
	case Prismarine:
		switch block.Type {
		case NormalPrismarine():
			return "prismarine_rough", "stone_slab_type_2", 2
		case DarkPrismarine():
			return "prismarine_dark", "stone_slab_type_2", 3
		case BrickPrismarine():
			return "prismarine_brick", "stone_slab_type_2", 4
		}
		panic("invalid prismarine type")
	case NetherBricks:
		if block.Type == RedNetherBricks() {
			return "nether_brick", "stone_slab_type", 7
		}
		return "red_nether_brick", "stone_slab_type_2", 7
	case Quartz:
		if block.Smooth {
			return "smooth_quartz", "stone_slab_type_4", 1
		}
		return "quartz", "stone_slab_type", 6
	case Purpur:
		return "purpur", "stone_slab_type_2", 1
	case EndBricks:
		return "end_stone_brick", "stone_slab_type_3", 0
	// TODO: Blackstone
	// TODO: Copper
	// TODO: Deepslate,
	case MudBricks:
		return "mud_brick", "", 0
	}
	panic("invalid block used for slab")
}

// encodeLegacySlabId encodes a legacy slab type to its identifier.
func encodeLegacySlabId(slabType string) string {
	switch slabType {
	case "wood_type":
		return "wooden_slab"
	case "stone_slab_type":
		return "stone_block_slab"
	case "stone_slab_type_2":
		return "stone_block_slab2"
	case "stone_slab_type_3":
		return "stone_block_slab3"
	case "stone_slab_type_4":
		return "stone_block_slab4"
	}
	panic("invalid slab type")
}

// SlabBlocks returns a list of all possible blocks for a slab.
func SlabBlocks() []world.Block {
	b := []world.Block{
		Stone{},
		Stone{Smooth: true},
		Cobblestone{},
		Cobblestone{Mossy: true},
		StoneBricks{},
		StoneBricks{Type: MossyStoneBricks()},
		Andesite{},
		Andesite{Polished: true},
		Diorite{},
		Diorite{Polished: true},
		Granite{},
		Granite{Polished: true},
		Bricks{},
		NetherBricks{},
		NetherBricks{Type: RedNetherBricks()},
		Quartz{},
		Quartz{Smooth: true},
		Purpur{},
		EndBricks{},
		// TODO: Blackstone
		// TODO: Copper
		// TODO: Deepslate,
		MudBricks{},
	}
	for _, p := range PrismarineTypes() {
		b = append(b, Prismarine{Type: p})
	}
	for _, s := range SandstoneTypes() {
		if s != ChiseledSandstone() {
			b = append(b, Sandstone{Type: s})
			b = append(b, Sandstone{Type: s, Red: true})
		}
	}
	for _, w := range WoodTypes() {
		b = append(b, Planks{Wood: w})
	}
	return b
}
