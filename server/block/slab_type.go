package block

import (
	"github.com/df-mc/dragonfly/server/world"
)

// encodeSlabBlock encodes the provided block in to an identifier and meta value that can be used to encode the slab.
// halfFlattened is a temporary hack for a stone_block_slab which has been flattened but double_stone_block_slab
// has not. This can be removed in 1.21.10 where they have flattened all slab types.
func encodeSlabBlock(block world.Block, double bool) (id string, suffix string) {
	suffix = "_slab"
	if double {
		suffix = "_double_slab"
	}

	switch block := block.(type) {
	case Andesite:
		if block.Polished {
			return "polished_andesite", suffix
		}
		return "andesite", suffix
	case Blackstone:
		if block.Type == NormalBlackstone() {
			return "blackstone", suffix
		} else if block.Type == PolishedBlackstone() {
			return "polished_blackstone", suffix
		}
	case Bricks:
		return "brick", suffix
	case Cobblestone:
		if block.Mossy {
			return "mossy_cobblestone", suffix
		}
		return "cobblestone", suffix
	case Copper:
		if block.Type == CutCopper() {
			suffix = "cut_copper_slab"
			if double {
				suffix = "double_" + suffix
			}
			var name string
			if block.Oxidation != UnoxidisedOxidation() {
				name = block.Oxidation.String() + "_"
			}
			if block.Waxed {
				name = "waxed_" + name
			}
			return name, suffix
		}
	case Deepslate:
		if block.Type == CobbledDeepslate() {
			return "cobbled_deepslate", suffix
		} else if block.Type == PolishedDeepslate() {
			return "polished_deepslate", suffix
		}
	case DeepslateBricks:
		if !block.Cracked {
			return "deepslate_brick", suffix
		}
	case DeepslateTiles:
		if !block.Cracked {
			return "deepslate_tile", suffix
		}
	case Diorite:
		if block.Polished {
			return "polished_diorite", suffix
		}
		return "diorite", suffix
	case EndBricks:
		return "end_stone_brick", suffix
	case Granite:
		if block.Polished {
			return "polished_granite", suffix
		}
		return "granite", suffix
	case MudBricks:
		return "mud_brick", suffix
	case NetherBricks:
		if block.Type == RedNetherBricks() {
			return "nether_brick", suffix
		}
		return "red_nether_brick", suffix
	case Planks:
		return block.Wood.String(), suffix
	case PolishedBlackstoneBrick:
		if !block.Cracked {
			return "polished_blackstone_brick", suffix
		}
	case PolishedTuff:
		return "polished_tuff", suffix
	case Prismarine:
		switch block.Type {
		case NormalPrismarine():
			return "prismarine", suffix
		case DarkPrismarine():
			return "dark_prismarine", suffix
		case BrickPrismarine():
			return "prismarine_brick", suffix
		}
		panic("invalid prismarine type")
	case Purpur:
		return "purpur", suffix
	case Quartz:
		if block.Smooth {
			return "smooth_quartz", suffix
		}
		return "quartz", suffix
	case ResinBricks:
		return "resin_brick", suffix
	case Sandstone:
		switch block.Type {
		case NormalSandstone():
			if block.Red {
				return "red_sandstone", suffix
			}
			return "sandstone", suffix
		case CutSandstone():
			if block.Red {
				return "cut_red_sandstone", suffix
			}
			return "cut_sandstone", suffix
		case SmoothSandstone():
			if block.Red {
				return "smooth_red_sandstone", suffix
			}
			return "smooth_sandstone", suffix
		}
		panic("invalid sandstone type")
	case Stone:
		if block.Smooth {
			return "smooth_stone", suffix
		}
		return "normal_stone", suffix
	case StoneBricks:
		if block.Type == MossyStoneBricks() {
			return "mossy_stone_brick", suffix
		}
		return "stone_brick", suffix
	case Tuff:
		if !block.Chiseled {
			return "tuff", suffix
		}
	case TuffBricks:
		if !block.Chiseled {
			return "tuff_brick", suffix
		}
	}
	panic("invalid block used for slab")
}

// SlabBlocks returns a list of all possible blocks for a slab.
func SlabBlocks() []world.Block {
	b := []world.Block{
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
		PolishedTuff{},
		Purpur{},
		Quartz{Smooth: true},
		Quartz{},
		ResinBricks{},
		StoneBricks{Type: MossyStoneBricks()},
		StoneBricks{},
		Stone{Smooth: true},
		Stone{},
		Tuff{},
		TuffBricks{},
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
	for _, o := range OxidationTypes() {
		b = append(b, Copper{Type: CutCopper(), Oxidation: o})
		b = append(b, Copper{Type: CutCopper(), Oxidation: o, Waxed: true})
	}
	return b
}
