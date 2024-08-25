package block

import (
	"github.com/df-mc/dragonfly/server/world"
)

// encodeSlabBlock encodes the provided block in to an identifier and meta value that can be used to encode the slab.
// halfFlattened is a temporary hack for a stone_block_slab which has been flattened but double_stone_block_slab
// has not. This can be removed in 1.21.10 where they have flattened all slab types.
func encodeSlabBlock(block world.Block) (id string) {
	switch block := block.(type) {
	// TODO: Copper
	case Andesite:
		if block.Polished {
			return "polished_andesite"
		}
		return "andesite"
	case Blackstone:
		if block.Type == NormalBlackstone() {
			return "blackstone"
		} else if block.Type == PolishedBlackstone() {
			return "polished_blackstone"
		}
	case Bricks:
		return "brick"
	case Cobblestone:
		if block.Mossy {
			return "mossy_cobblestone"
		}
		return "cobblestone"
	case Deepslate:
		if block.Type == CobbledDeepslate() {
			return "cobbled_deepslate"
		} else if block.Type == PolishedDeepslate() {
			return "polished_deepslate"
		}
	case DeepslateBricks:
		if !block.Cracked {
			return "deepslate_brick"
		}
	case DeepslateTiles:
		if !block.Cracked {
			return "deepslate_tile"
		}
	case Diorite:
		if block.Polished {
			return "polished_diorite"
		}
		return "diorite"
	case EndBricks:
		return "end_stone_brick"
	case Granite:
		if block.Polished {
			return "polished_granite"
		}
		return "granite"
	case MudBricks:
		return "mud_brick"
	case NetherBricks:
		if block.Type == RedNetherBricks() {
			return "nether_brick"
		}
		return "red_nether_brick"
	case Planks:
		return block.Wood.String()
	case PolishedBlackstoneBrick:
		if !block.Cracked {
			return "polished_blackstone_brick"
		}
	case Prismarine:
		switch block.Type {
		case NormalPrismarine():
			return "prismarine"
		case DarkPrismarine():
			return "dark_prismarine"
		case BrickPrismarine():
			return "prismarine_brick"
		}
		panic("invalid prismarine type")
	case Purpur:
		return "purpur"
	case Quartz:
		if block.Smooth {
			return "smooth_quartz"
		}
		return "quartz"
	case Sandstone:
		switch block.Type {
		case NormalSandstone():
			if block.Red {
				return "red_sandstone"
			}
			return "sandstone"
		case CutSandstone():
			if block.Red {
				return "cut_red_sandstone"
			}
			return "cut_sandstone"
		case SmoothSandstone():
			if block.Red {
				return "smooth_red_sandstone"
			}
			return "smooth_sandstone"
		}
		panic("invalid sandstone type")
	case Stone:
		if block.Smooth {
			return "smooth_stone"
		}
		return "normal_stone"
	case StoneBricks:
		if block.Type == MossyStoneBricks() {
			return "mossy_stone_brick"
		}
		return "stone_brick"
	case Tuff:
		return "tuff"
	}
	panic("invalid block used for slab")
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
		Tuff{},
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
