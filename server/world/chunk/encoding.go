package chunk

import (
	"bytes"
	"github.com/sandertv/gophertunnel/minecraft/nbt"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// Encoding is an encoding type used for Chunk encoding. Implementations of this interface are DiskEncoding and
// NetworkEncoding, which can be used to encode a Chunk to an intermediate disk or network representation respectively.
type Encoding interface {
	encoding() nbt.Encoding
	encodePalette(buf *bytes.Buffer, p *Palette)
	network() byte
	data2D(c *Chunk) []byte
}

// DiskEncoding is the Encoding for writing a Chunk to disk. It writes block palettes using NBT and does not use
// varints.
var DiskEncoding diskEncoding

// NetworkEncoding is the Encoding used for sending a Chunk over network. It does not use NBT and writes varints.
var NetworkEncoding networkEncoding

type diskEncoding struct{}

func (diskEncoding) network() byte          { return 0 }
func (diskEncoding) encoding() nbt.Encoding { return nbt.LittleEndian }
func (diskEncoding) data2D(c *Chunk) []byte { return append(c.biomes[:], 0) }
func (diskEncoding) encodePalette(buf *bytes.Buffer, p *Palette) {
	blocks := make([]blockEntry, p.Len())
	for index, runtimeID := range p.blockRuntimeIDs {
		// Get the block state registered with the runtime IDs we have in the palette of the block storage
		// as we need the name and data value to store.
		name, props, _ := RuntimeIDToState(runtimeID)
		blocks[index] = blockEntry{Name: name, State: props, Version: CurrentBlockVersion}
	}
	// Marshal the slice of block states into NBT and add it to the byte slice.
	enc := nbt.NewEncoderWithEncoding(buf, nbt.LittleEndian)
	for _, b := range blocks {
		_ = enc.Encode(b)
	}
}

type networkEncoding struct{}

func (networkEncoding) network() byte          { return 1 }
func (networkEncoding) encoding() nbt.Encoding { return nbt.NetworkLittleEndian }
func (networkEncoding) data2D(c *Chunk) []byte { return append(emptyHeightMap, c.biomes[:]...) }
func (networkEncoding) encodePalette(buf *bytes.Buffer, p *Palette) {
	_ = protocol.WriteVarint32(buf, int32(p.Len()))
	for _, runtimeID := range p.blockRuntimeIDs {
		_ = protocol.WriteVarint32(buf, int32(runtimeID))
	}
}
