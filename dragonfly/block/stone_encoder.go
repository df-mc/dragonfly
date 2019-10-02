package block

import (
	"github.com/dragonfly-tech/dragonfly/dragonfly/block/encoder"
)

// stoneEncoder implements the encoding and decoding of stone type blocks.
type stoneEncoder struct{}

// BlocksHandled ...
func (stoneEncoder) BlocksHandled() []string {
	return []string{"minecraft:stone"}
}

// DecodeBlock ...
func (stoneEncoder) DecodeBlock(id string, meta int16, nbt []byte) encoder.Block {
	switch meta {
	case 1:
		return Granite{}
	case 2:
		return Granite{Polished: true}
	case 3:
		return Diorite{}
	case 4:
		return Diorite{Polished: true}
	case 5:
		return Andesite{}
	case 6:
		return Andesite{Polished: true}
	default:
		return Stone{}
	}
}

// EncodeBlock ...
func (stoneEncoder) EncodeBlock(b encoder.Block) (id string, meta int16, nbt []byte) {
	switch block := b.(type) {
	case Granite:
		if !block.Polished {
			return "minecraft:stone", 1, nil
		}
		return "minecraft:stone", 2, nil
	case Diorite:
		if !block.Polished {
			return "minecraft:stone", 3, nil
		}
		return "minecraft:stone", 4, nil
	case Andesite:
		if !block.Polished {
			return "minecraft:stone", 5, nil
		}
		return "minecraft:stone", 6, nil
	default:
		return "minecraft:stone", 0, nil
	}
}
