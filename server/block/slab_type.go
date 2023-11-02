package block

import (
	"github.com/df-mc/dragonfly/server/world"
)

// encodeSlabBlock encodes the provided block in to an identifier and meta value that can be used to encode the slab.
func encodeSlabBlock(block world.Block) (id, slabType string, meta int16) {
	switch block := block.(type) {
	// TODO: Copper
	case Andesite:
		if block.Polished {
			return "polished_andesite", "stone_slab_type_3", 2
		}
		return "andesite", "stone_slab_type_3", 3
	case Blackstone:
		if block.Type == NormalBlackstone() {
			return "blackstone", "", 0
		} else if block.Type == PolishedBlackstone() {
			return "polished_blackstone", "", 0
		}
	case Bricks:
		return "brick", "stone_slab_type", 4
	case Cobblestone:
		if block.Mossy {
			return "mossy_cobblestone", "stone_slab_type_2", 5
		}
		return "cobblestone", "stone_slab_type", 3
	case Deepslate:
		if block.Type == CobbledDeepslate() {
			return "cobbled_deepslate", "", 0
		} else if block.Type == PolishedDeepslate() {
			return "polished_deepslate", "", 0
		}
	case DeepslateBricks:
		if !block.Cracked {
			return "deepslate_brick", "", 0
		}
	case DeepslateTiles:
		if !block.Cracked {
			return "deepslate_tile", "", 0
		}
	case Diorite:
		if block.Polished {
			return "polished_diorite", "stone_slab_type_3", 5
		}
		return "diorite", "stone_slab_type_3", 4
	case EndBricks:
		return "end_stone_brick", "stone_slab_type_3", 0
	case Granite:
		if block.Polished {
			return "polished_granite", "stone_slab_type_3", 7
		}
		return "granite", "stone_slab_type_3", 6
	case MudBricks:
		return "mud_brick", "", 0
	case NetherBricks:
		if block.Type == RedNetherBricks() {
			return "nether_brick", "stone_slab_type", 7
		}
		return "red_nether_brick", "stone_slab_type_2", 7
	case Planks:
		switch block.Wood {
		case OakWood(), SpruceWood(), BirchWood(), JungleWood(), AcaciaWood(), DarkOakWood():
			return block.Wood.String(), "wood_type", int16(block.Wood.Uint8())
		default:
			return block.Wood.String(), "", 0
		}
	case PolishedBlackstoneBrick:
		if !block.Cracked {
			return "polished_blackstone_brick", "", 0
		}
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
	case Purpur:
		return "purpur", "stone_slab_type_2", 1
	case Quartz:
		if block.Smooth {
			return "smooth_quartz", "stone_slab_type_4", 1
		}
		return "quartz", "stone_slab_type", 6
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
	case Stone:
		if block.Smooth {
			return "smooth_stone", "stone_slab_type", 0
		}
		return "stone", "stone_slab_type_4", 2
	case StoneBricks:
		if block.Type == MossyStoneBricks() {
			return "mossy_stone_brick", "stone_slab_type_4", 0
		}
		return "stone_brick", "stone_slab_type", 5
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
		// TODO: Copper
		Andesite{Polished: true},
		Andesite{},
		Blackstone{Type: PolishedBlackstone()},
		Blackstone{},
		Bricks{},
		Cobblestone{Mossy: true},
		Cobblestone{},
		DeepslateBricks{},
		DeepslateTiles{},
		Deepslate{Type: CobbledDeepslate()},
		Deepslate{Type: PolishedDeepslate()},
		Diorite{Polished: true},
		Diorite{},
		EndBricks{},
		Granite{Polished: true},
		Granite{},
		MudBricks{},
		NetherBricks{Type: RedNetherBricks()},
		NetherBricks{},
		PolishedBlackstoneBrick{},
		Purpur{},
		Quartz{Smooth: true},
		Quartz{},
		StoneBricks{Type: MossyStoneBricks()},
		StoneBricks{},
		Stone{Smooth: true},
		Stone{},
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
