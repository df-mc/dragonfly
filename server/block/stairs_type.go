package block

import (
	"github.com/df-mc/dragonfly/server/world"
)

// encodeStairsBlock encodes the provided block in to an identifier and meta value that can be used to encode the stairs.
func encodeStairsBlock(block world.Block) string {
	switch block := block.(type) {
	case Planks:
		return block.Wood.String()
	case Stone:
		if !block.Smooth {
			return "normal_stone"
		}
	case Cobblestone:
		if block.Mossy {
			return "mossy_cobblestone"
		}
		return "stone"
	case StoneBricks:
		if block.Type == MossyStoneBricks() {
			return "mossy_stone_brick"
		}
		return "stone_brick"
	case Andesite:
		if block.Polished {
			return "polished_andesite"
		}
		return "andesite"
	case Diorite:
		if block.Polished {
			return "polished_diorite"
		}
		return "diorite"
	case Granite:
		if block.Polished {
			return "polished_granite"
		}
		return "granite"
	case Sandstone:
		switch block.Type {
		case NormalSandstone():
			if block.Red {
				return "red_sandstone"
			}
			return "sandstone"
		case SmoothSandstone():
			if block.Red {
				return "smooth_red_sandstone"
			}
			return "smooth_sandstone"
		}
		panic("invalid sandstone type")
	case Bricks:
		return "brick"
	case Prismarine:
		switch block.Type {
		case NormalPrismarine():
			return "prismarine"
		case DarkPrismarine():
			return "dark_prismarine"
		case BrickPrismarine():
			return "prismarine_bricks"
		}
		panic("invalid prismarine type")
	case NetherBricks:
		if block.Type == RedNetherBricks() {
			return "nether_brick"
		}
		return "red_nether_brick"
	case Quartz:
		if block.Smooth {
			return "smooth_quartz"
		}
		return "quartz"
	case Purpur:
		return "purpur"
	case EndBricks:
		return "end_brick"
	// TODO: Blackstone
	// TODO: Copper
	// TODO: Deepslate,
	case MudBricks:
		return "mud_brick"
	}
	panic("invalid block used for slab")
}

// StairsBlocks returns a list of all possible blocks for stairs.
func StairsBlocks() []world.Block {
	b := []world.Block{
		// TODO: Blackstone
		// TODO: Copper
		// TODO: Deepslate,
		Andesite{Polished: true},
		Andesite{},
		Bricks{},
		Cobblestone{Mossy: true},
		Cobblestone{},
		Diorite{Polished: true},
		Diorite{},
		EndBricks{},
		Granite{Polished: true},
		Granite{},
		MudBricks{},
		NetherBricks{Type: RedNetherBricks()},
		NetherBricks{},
		Purpur{},
		Quartz{Smooth: true},
		Quartz{},
		StoneBricks{Type: MossyStoneBricks()},
		StoneBricks{},
		Stone{},
	}
	for _, p := range PrismarineTypes() {
		b = append(b, Prismarine{Type: p})
	}
	for _, s := range SandstoneTypes() {
		if s != CutSandstone() && s != ChiseledSandstone() {
			b = append(b, Sandstone{Type: s})
			b = append(b, Sandstone{Type: s, Red: true})
		}
	}
	for _, w := range WoodTypes() {
		b = append(b, Planks{Wood: w})
	}
	return b
}
