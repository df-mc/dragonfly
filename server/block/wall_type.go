package block

import "github.com/df-mc/dragonfly/server/world"

// encodeWallBlock encodes the provided block in to an identifier and meta value that can be used to encode the wall.
func encodeWallBlock(block world.Block) (string, int16) {
	switch block := block.(type) {
	case Andesite:
		if !block.Polished {
			return "andesite", 4
		}
	// TODO: Blackstone
	case Bricks:
		return "brick", 6
	// TODO: Cobbled Deepslate
	case Cobblestone:
		if block.Mossy {
			return "mossy_cobblestone", 1
		}
		return "cobblestone", 0
	// TODO: Deepslate Brick
	// TODO: Deepslate Tile
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

// calculateWallHardness calculates the appropriate hardness for a wall based on the provided block.
func calculateWallHardness(block world.Block) float64 {
	switch block := block.(type) {
	case Andesite:
		if !block.Polished {
			return 1.5
		}
	// TODO: Blackstone: 2
	case Bricks:
		return 2
	// TODO: Cobbled Deepslate: 3.5
	case Cobblestone:
		if block.Mossy {
			return 2
		}
		return 1.5
	// TODO: Deepslate Brick: 3.5
	// TODO: Deepslate Tile: 3.5
	case Diorite:
		if !block.Polished {
			return 1.5
		}
	case EndBricks:
		return 3
	case Granite:
		if !block.Polished {
			return 1.5
		}
	case MudBricks:
		return 1.5
	case NetherBricks:
		if block.Type == NormalNetherBricks() || block.Type == RedNetherBricks() {
			return 2
		}
	case Prismarine:
		if block.Type == NormalPrismarine() {
			return 1.5
		}
	case Sandstone:
		if block.Type == NormalSandstone() {
			return 0.8
		}
	case StoneBricks:
		if block.Type == NormalStoneBricks() || block.Type == MossyStoneBricks() {
			return 1.5
		}
	}
	panic("invalid block used for wall")
}

// WallBlocks returns a list of all possible blocks for a wall.
func WallBlocks() []world.Block {
	return []world.Block{
		Andesite{},
		// TODO: Blackstone
		Bricks{},
		// TODO: Cobbled Deepslate
		Cobblestone{},
		Cobblestone{Mossy: true},
		// TODO: Deepslate Brick
		// TODO: Deepslate Tile
		Diorite{},
		EndBricks{},
		Granite{},
		MudBricks{},
		NetherBricks{},
		NetherBricks{Type: RedNetherBricks()},
		// TODO: Polished Blackstone
		// TODO: Polished Blackstone brick
		// TODO: Polished Deepslate
		Prismarine{},
		Sandstone{},
		Sandstone{Red: true},
		StoneBricks{},
		StoneBricks{Type: MossyStoneBricks()},
	}
}
