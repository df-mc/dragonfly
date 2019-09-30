package encoder

import "github.com/dragonfly-tech/dragonfly/dragonfly/block"

// dirtEncoder implements the encoding of dirt and coarse dirt blocks.
type dirtEncoder struct{}

// DecodeBlock ...
func (dirtEncoder) DecodeBlock(id string, meta int16, nbt []byte) Block {
	switch meta {
	default:
		return block.Dirt{}
	case 1:
		return block.CoarseDirt{}
	}
}

// EncodeBlock ...
func (dirtEncoder) EncodeBlock(b Block) (id string, meta int16, nbt []byte) {
	switch b.(type) {
	default:
		return "minecraft:dirt", 0, nil
	case block.CoarseDirt:
		return "minecraft:dirt", 1, nil
	}
}

// BlocksHandled ...
func (dirtEncoder) BlocksHandled() []string {
	return []string{"minecraft:dirt"}
}
