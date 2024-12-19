package block

import "github.com/df-mc/dragonfly/server/world"

// encodeWallBlock encodes the provided block in to an identifier and meta value that can be used to encode the wall.
func encodeWallBlock(block world.Block) string {
	switch block := block.(type) {
	case Andesite:
		if !block.Polished {
			return "andesite"
		}
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
		if !block.Polished {
			return "diorite"
		}
	case EndBricks:
		return "end_stone_brick"
	case Granite:
		if !block.Polished {
			return "granite"
		}
	case MudBricks:
		return "mud_brick"
	case NetherBricks:
		if block.Type == NormalNetherBricks() {
			return "nether_brick"
		} else if block.Type == RedNetherBricks() {
			return "red_nether_brick"
		}
	case PolishedBlackstoneBrick:
		if !block.Cracked {
			return "polished_blackstone_brick"
		}
	case PolishedTuff:
		return "polished_tuff"
	case Prismarine:
		if block.Type == NormalPrismarine() {
			return "prismarine"
		}
	case ResinBricks:
		return "resin_brick"
	case Sandstone:
		if block.Type == NormalSandstone() {
			if block.Red {
				return "red_sandstone"
			}
			return "sandstone"
		}
	case StoneBricks:
		if block.Type == NormalStoneBricks() {
			return "stone_brick"
		} else if block.Type == MossyStoneBricks() {
			return "mossy_stone_brick"
		}
	case Tuff:
		if !block.Chiseled {
			return "tuff"
		}
	case TuffBricks:
		if !block.Chiseled {
			return "tuff_brick"
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
		PolishedTuff{},
		Prismarine{},
		ResinBricks{},
		Sandstone{Red: true},
		Sandstone{},
		StoneBricks{Type: MossyStoneBricks()},
		StoneBricks{},
		Tuff{},
		TuffBricks{},
	}
}
