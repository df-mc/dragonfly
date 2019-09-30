package encoder

import (
	"github.com/dragonfly-tech/dragonfly/dragonfly/block"
)

// stoneEncoder implements the encoding and decoding of stone type blocks.
type stoneEncoder struct{}

// BlocksHandled ...
func (stoneEncoder) BlocksHandled() []string {
	return []string{"minecraft:stone"}
}

// DecodeBlock ...
func (stoneEncoder) DecodeBlock(id string, meta int16, nbt []byte) Block {
	switch meta {
	case 1:
		return block.Granite{}
	case 2:
		return block.PolishedGranite{}
	case 3:
		return block.Diorite{}
	case 4:
		return block.PolishedDiorite{}
	case 5:
		return block.Andesite{}
	case 6:
		return block.PolishedAndesite{}
	default:
		return block.Stone{}
	}
}

// EncodeBlock ...
func (stoneEncoder) EncodeBlock(b Block) (id string, meta int16, nbt []byte) {
	switch b.(type) {
	case block.Granite:
		return "minecraft:stone", 1, nil
	case block.PolishedGranite:
		return "minecraft:stone", 2, nil
	case block.Diorite:
		return "minecraft:stone", 3, nil
	case block.PolishedDiorite:
		return "minecraft:stone", 4, nil
	case block.Andesite:
		return "minecraft:stone", 5, nil
	case block.PolishedAndesite:
		return "minecraft:stone", 6, nil
	default:
		return "minecraft:stone", 0, nil
	}
}
