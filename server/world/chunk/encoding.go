package chunk

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/sandertv/gophertunnel/minecraft/nbt"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// Encoding is an encoding type used for Chunk encoding. Implementations of this interface are DiskEncoding and
// NetworkEncoding, which can be used to encode a Chunk to an intermediate disk or network representation respectively.
type Encoding interface {
	encoding() nbt.Encoding
	encodePalette(buf *bytes.Buffer, p *Palette)
	decodePalette(buf *bytes.Buffer, blockSize paletteSize) (*Palette, error)
	network() byte
	data2D(c *Chunk) []byte
}

// DiskEncoding is the Encoding for writing a Chunk to disk. It writes block palettes using NBT and does not use
// varints.
var DiskEncoding diskEncoding

// NetworkEncoding is the Encoding used for sending a Chunk over network. It does not use NBT and writes varints.
var NetworkEncoding networkEncoding

// diskEncoding implements the Chunk encoding for writing to disk.
type diskEncoding struct{}

func (diskEncoding) network() byte          { return 0 }
func (diskEncoding) encoding() nbt.Encoding { return nbt.LittleEndian }
func (diskEncoding) data2D(c *Chunk) []byte { return append(emptyHeightMap, c.biomes[:]...) }
func (diskEncoding) encodePalette(buf *bytes.Buffer, p *Palette) {
	_ = binary.Write(buf, binary.LittleEndian, uint32(p.Len()))
	blocks := make([]blockEntry, p.Len())
	for index, runtimeID := range p.values {
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
func (diskEncoding) decodePalette(buf *bytes.Buffer, blockSize paletteSize) (*Palette, error) {
	// The next 4 bytes are an LE int32, but we simply read it and decode the int32 ourselves, as it's much
	// faster here.
	data := buf.Next(4)
	if len(data) != 4 {
		return nil, fmt.Errorf("cannot read palette entry count: expected 4 bytes, got %v", len(data))
	}
	var (
		paletteCount = binary.LittleEndian.Uint32(data)
		palette      = newPalette(blockSize, make([]uint32, paletteCount))
		dec          = nbt.NewDecoderWithEncoding(buf, nbt.LittleEndian)
		e            blockEntry
		ok           bool
	)
	for i := uint32(0); i < paletteCount; i++ {
		if err := dec.Decode(&e); err != nil {
			return nil, fmt.Errorf("error decoding block: %w", err)
		}
		palette.values[i], ok = StateToRuntimeID(e.Name, e.State)
		if !ok {
			return nil, fmt.Errorf("cannot get runtime ID of block state %v{%+v}", e.Name, e.State)
		}
	}
	return palette, nil
}

// networkEncoding implements the Chunk encoding for sending over network.
type networkEncoding struct{}

func (networkEncoding) network() byte          { return 1 }
func (networkEncoding) encoding() nbt.Encoding { return nbt.NetworkLittleEndian }
func (networkEncoding) data2D(c *Chunk) []byte { return append(c.biomes[:], 0) }
func (networkEncoding) encodePalette(buf *bytes.Buffer, p *Palette) {
	_ = protocol.WriteVarint32(buf, int32(p.Len()))
	for _, runtimeID := range p.values {
		_ = protocol.WriteVarint32(buf, int32(runtimeID))
	}
}
func (networkEncoding) decodePalette(buf *bytes.Buffer, blockSize paletteSize) (*Palette, error) {
	var paletteCount int32
	if err := protocol.Varint32(buf, &paletteCount); err != nil {
		return nil, fmt.Errorf("error reading palette entry count: %w", err)
	}
	if paletteCount <= 0 {
		return nil, fmt.Errorf("invalid palette entry count %v", paletteCount)
	}

	blocks, temp := make([]uint32, paletteCount), int32(0)
	for i := int32(0); i < paletteCount; i++ {
		if err := protocol.Varint32(buf, &temp); err != nil {
			return nil, fmt.Errorf("error decoding palette entry: %w", err)
		}
		blocks[i] = uint32(temp)
	}
	return &Palette{values: blocks, size: blockSize}, nil
}
