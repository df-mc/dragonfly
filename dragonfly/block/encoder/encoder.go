package encoder

import (
	"fmt"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"reflect"
)

// BlockEncoder represents an object that can encode and decode specific blocks. It essentially forms the
// bridge between the raw ID, meta and NBT of a block to a friendly Go implementation of a block.
type BlockEncoder interface {
	// DecodeBlock decodes a block from the raw ID, meta value and NBT passed. Most blocks do not need NBT and
	// can ignore the value.
	DecodeBlock(id string, meta int16, nbt []byte) Block
	// EncodeBlock encodes a block to the raw ID, meta value and NBT. If the block has no NBT, nil can be
	// returned instead.
	EncodeBlock(b Block) (id string, meta int16, nbt []byte)
	// BlocksHandled returns the ID of all blocks that the encoder handles. For example, the log encoder uses
	// "minecraft:log" and "minecraft:log2".
	BlocksHandled() []string
}

// RegisterBlocks registers a list of blocks with a specific encoder. This encoder must be able to produce
// and encode the blocks passed.
func Register(encoder BlockEncoder, blocks ...Block) {
	for _, b := range blocks {
		blockEncoders[reflect.TypeOf(b)] = encoder
	}
	for _, id := range encoder.BlocksHandled() {
		blockEncodersIDs[id] = encoder
	}
}

// ByID returns a block encoder by a block ID passed. If no encoder was registered for the block, false is
// returned.
func ByID(id string) (BlockEncoder, bool) {
	v, ok := blockEncodersIDs[id]
	return v, ok
}

// ByBlock returns a block encoder by a block passed. If no encoder was registered for the block, false is
// returned.
func ByBlock(b Block) (BlockEncoder, bool) {
	v, ok := blockEncoders[reflect.TypeOf(b)]
	return v, ok
}

// Block represents a block that may be placed in the world. Blocks should always implement the world.Block
// interface in order to be placeable in the world.
type Block interface{}

var blockEncoders = map[reflect.Type]BlockEncoder{}
var blockEncodersIDs = map[string]BlockEncoder{}

// RuntimeID returns the runtime ID of a block passed. If for any reason the runtime ID of the block could not
// be found, the function panics.
func RuntimeID(b Block) uint32 {
	e, ok := ByBlock(b)
	if !ok {
		panic(fmt.Sprintf("encoder not found for block %T", b))
	}
	id, meta, _ := e.EncodeBlock(b)
	return RuntimeIDs[protocol.BlockEntry{
		Name: id,
		Data: meta,
	}]
}

// BasicEncoder implements a basic encoder, which is useful for blocks that do not have metadata or NBT, such
// as bedrock.
type BasicEncoder struct {
	ID    string
	Block Block
}

// DecodeBlock ...
func (e BasicEncoder) DecodeBlock(id string, meta int16, nbt []byte) Block {
	return e.Block
}

// EncodeBlock ...
func (e BasicEncoder) EncodeBlock(b Block) (id string, meta int16, nbt []byte) {
	return e.ID, 0, nil
}

// BlocksHandled ...
func (e BasicEncoder) BlocksHandled() []string {
	return []string{e.ID}
}
