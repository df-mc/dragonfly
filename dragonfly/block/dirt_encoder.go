package block

import "github.com/dragonfly-tech/dragonfly/dragonfly/block/encoder"

// dirtEncoder implements the encoding of dirt and coarse dirt blocks.
type dirtEncoder struct{}

// DecodeBlock ...
func (dirtEncoder) DecodeBlock(id string, meta int16, nbt []byte) encoder.Block {
	switch meta {
	default:
		return Dirt{}
	case 1:
		return CoarseDirt{}
	}
}

// EncodeBlock ...
func (dirtEncoder) EncodeBlock(b encoder.Block) (id string, meta int16, nbt []byte) {
	switch b.(type) {
	default:
		return "minecraft:dirt", 0, nil
	case CoarseDirt:
		return "minecraft:dirt", 1, nil
	}
}

// BlocksHandled ...
func (dirtEncoder) BlocksHandled() []string {
	return []string{"minecraft:dirt"}
}
