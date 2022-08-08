package block

import "github.com/df-mc/dragonfly/server/world"

// encodeWallBlock encodes the provided block in to an identifier and meta value that can be used to encode the wall.
func encodeWallBlock(block world.Block) (string, int16) {
	switch block := block.(type) {
	case Andesite:
		if !block.Polished {
			return "andesite", 4
		}
	case Blackstone:
		if block.Type == NormalBlackstone() {
			return "blackstone", 0
		} else if block.Type == PolishedBlackstone() {
			return "polished_blackstone", 0
		}
	case Bricks:
		return "brick", 6
	case Cobblestone:
		if block.Mossy {
			return "mossy_cobblestone", 1
		}
		return "cobblestone", 0
	case Deepslate:
		if block.Type == CobbledDeepslate() {
			return "cobbled_deepslate", 0
		} else if block.Type == PolishedDeepslate() {
			return "polished_deepslate", 0
		}
	case DeepslateBricks:
		if !block.Cracked {
			return "deepslate_brick", 0
		}
	case DeepslateTiles:
		if !block.Cracked {
			return "deepslate_tile", 0
		}
	case Diorite:
		if !block.Polished {
			return "diorite", 3
		}
	case EndBricks:
		return "end_brick", 10
	case Granite:
		if !block.Polished {
			return "granite", 2
		}
	case MudBricks:
		return "mud_brick", 0
	case NetherBricks:
		if block.Type == NormalNetherBricks() {
			return "nether_brick", 9
		} else if block.Type == RedNetherBricks() {
			return "red_nether_brick", 13
		}
	case PolishedBlackstoneBrick:
		if !block.Cracked {
			return "polished_blackstone_brick", 0
		}
	case Prismarine:
		if block.Type == NormalPrismarine() {
			return "prismarine", 11
		}
	case Sandstone:
		if block.Type == NormalSandstone() {
			if block.Red {
				return "red_sandstone", 12
			}
			return "sandstone", 5
		}
	case StoneBricks:
		if block.Type == NormalStoneBricks() {
			return "stone_brick", 7
		} else if block.Type == MossyStoneBricks() {
			return "mossy_stone_brick", 8
		}
	}
	panic("invalid block used for wall")
}

// WallBlocks returns a list of all possible blocks for a wall.
func WallBlocks() []world.Block {
	return []world.Block{
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
		Diorite{},
		EndBricks{},
		Granite{},
		MudBricks{},
		NetherBricks{Type: RedNetherBricks()},
		NetherBricks{},
		PolishedBlackstoneBrick{},
		Prismarine{},
		Sandstone{Red: true},
		Sandstone{},
		StoneBricks{Type: MossyStoneBricks()},
		StoneBricks{},
	}
}
