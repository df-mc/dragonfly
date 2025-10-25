package chunk

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"strings"

	"github.com/df-mc/worldupgrader/blockupgrader"
	"github.com/sandertv/gophertunnel/minecraft/nbt"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

type (
	// Encoding is an encoding type used for Chunk encoding. Implementations of this interface are DiskEncoding and
	// NetworkEncoding, which can be used to encode a Chunk to an intermediate disk or network representation respectively.
	Encoding interface {
		encodePalette(buf *bytes.Buffer, p *Palette, e paletteEncoding)
		decodePalette(buf *bytes.Buffer, blockSize paletteSize, e paletteEncoding) (*Palette, error)
		network() byte
	}
	// paletteEncoding is an encoding type used for Chunk encoding. It is used to encode different types of palettes
	// (for example, blocks or biomes) differently.
	paletteEncoding interface {
		encode(buf *bytes.Buffer, v uint32)
		decode(buf *bytes.Buffer, e Encoding) (uint32, error)
	}
)

var (
	// DiskEncoding is the Encoding for writing a Chunk to disk. It writes block palettes using NBT and does not use
	// varints.
	DiskEncoding diskEncoding
	// NetworkEncoding is the Encoding used for sending a Chunk over network. It does not use NBT and writes varints.
	NetworkEncoding networkEncoding
	// NetworkPersistentEncoding is the Encoding used for sending a Chunk over network. It uses NBT, unlike NetworkEncoding.
	NetworkPersistentEncoding networkPersistentEncoding
	// BiomePaletteEncoding is the paletteEncoding used for encoding a palette of biomes.
	BiomePaletteEncoding biomePaletteEncoding
	// BlockPaletteEncoding is the paletteEncoding used for encoding a palette of block states encoded as NBT.
	BlockPaletteEncoding blockPaletteEncoding
)

// biomePaletteEncoding implements the encoding of biome palettes to disk.
type biomePaletteEncoding struct{}

func (biomePaletteEncoding) encode(buf *bytes.Buffer, v uint32) {
	_ = binary.Write(buf, binary.LittleEndian, v)
}

func (biomePaletteEncoding) decode(buf *bytes.Buffer, e Encoding) (uint32, error) {
	var v uint32
	return v, binary.Read(buf, binary.LittleEndian, &v)
}

// blockPaletteEncoding implements the encoding of block palettes to disk.
type blockPaletteEncoding struct{}

func (bpe blockPaletteEncoding) encode(buf *bytes.Buffer, v uint32) {
	_ = nbt.NewEncoderWithEncoding(buf, nbt.LittleEndian).Encode(bpe.EncodeBlockState(v))
}
func (bpe blockPaletteEncoding) decode(buf *bytes.Buffer, e Encoding) (uint32, error) {
	var m map[string]any

	var encoding nbt.Encoding = nbt.LittleEndian
	if e.network() == 1 {
		encoding = nbt.NetworkLittleEndian
	}

	if err := nbt.NewDecoderWithEncoding(buf, encoding).Decode(&m); err != nil {
		return 0, fmt.Errorf("error decoding block palette entry: %w", err)
	}
	return bpe.DecodeBlockState(m)
}

func (bpe blockPaletteEncoding) EncodeBlockState(v uint32) blockEntry {
	// Get the block state registered with the runtime IDs we have in the palette of the block storage
	// as we need the name and data value to store.
	name, props, _ := RuntimeIDToState(v)
	return blockEntry{Name: name, State: props, Version: CurrentBlockVersion}
}

func (bpe blockPaletteEncoding) DecodeBlockState(m map[string]any) (uint32, error) {
	// Decode the name and version of the block entry.
	name, _ := m["name"].(string)
	version, _ := m["version"].(int32)

	if !strings.ContainsRune(name, ':') {
		name = "minecraft:" + name
	}

	// Now check for a state field.
	stateI, ok := m["states"]
	if version < 17694723 {
		// This entry is a pre-1.13 block state, so decode the meta value instead.
		meta, _ := m["val"].(int16)

		// Upgrade the pre-1.13 state into a post-1.13 state.
		state, ok := upgradeLegacyEntry(name, meta)
		if !ok {
			return 0, fmt.Errorf("cannot find mapping for legacy block entry: %v, %v", name, meta)
		}

		// Update the name, state, and version.
		name = state.Name
		stateI = state.State
		version = state.Version
	} else if !ok {
		// The state is a post-1.13 block state, but the states field is missing, likely due to a broken world
		// conversion.
		stateI = make(map[string]any)
	}
	state, ok := stateI.(map[string]any)
	if !ok {
		return 0, fmt.Errorf("invalid state in block entry")
	}

	// Upgrade the block state if necessary.
	upgraded := blockupgrader.Upgrade(blockupgrader.BlockState{
		Name:       name,
		Properties: state,
		Version:    version,
	})

	v, ok := StateToRuntimeID(upgraded.Name, upgraded.Properties)
	if !ok {
		return 0, fmt.Errorf("cannot get runtime ID of block state %v{%+v} %v", upgraded.Name, upgraded.Properties, upgraded.Version)
	}
	return v, nil
}

// diskEncoding implements the Chunk encoding for writing to disk.
type diskEncoding struct{}

func (diskEncoding) network() byte { return 0 }
func (diskEncoding) encodePalette(buf *bytes.Buffer, p *Palette, e paletteEncoding) {
	if p.size != 0 {
		_ = binary.Write(buf, binary.LittleEndian, uint32(p.Len()))
	}
	for _, v := range p.values {
		e.encode(buf, v)
	}
}

func (diskEncoding) decodePalette(buf *bytes.Buffer, blockSize paletteSize, e paletteEncoding) (*Palette, error) {
	paletteCount := uint32(1)
	if blockSize != 0 {
		if err := binary.Read(buf, binary.LittleEndian, &paletteCount); err != nil {
			return nil, fmt.Errorf("error reading palette entry count: %w", err)
		}
	}

	var err error
	palette := newPalette(blockSize, make([]uint32, paletteCount))
	for i := uint32(0); i < paletteCount; i++ {
		palette.values[i], err = e.decode(buf, DiskEncoding)
		if err != nil {
			return nil, err
		}
	}
	if paletteCount == 0 {
		return palette, fmt.Errorf("invalid palette entry count: found 0, but palette with %v bits per block must have at least 1 value", blockSize)
	}
	return palette, nil
}

// networkEncoding implements the Chunk encoding for sending over network.
type networkEncoding struct{}

func (networkEncoding) network() byte { return 1 }
func (networkEncoding) encodePalette(buf *bytes.Buffer, p *Palette, _ paletteEncoding) {
	if p.size != 0 {
		_ = protocol.WriteVarint32(buf, int32(p.Len()))
	}
	for _, val := range p.values {
		_ = protocol.WriteVarint32(buf, int32(val))
	}
}
func (networkEncoding) decodePalette(buf *bytes.Buffer, blockSize paletteSize, _ paletteEncoding) (*Palette, error) {
	var paletteCount int32 = 1
	if blockSize != 0 {
		if err := protocol.Varint32(buf, &paletteCount); err != nil {
			return nil, fmt.Errorf("error reading palette entry count: %w", err)
		}
		if paletteCount <= 0 {
			return nil, fmt.Errorf("invalid palette entry count %v", paletteCount)
		}
	}

	palette, temp := newPalette(blockSize, make([]uint32, paletteCount)), int32(0)
	for i := int32(0); i < paletteCount; i++ {
		if err := protocol.Varint32(buf, &temp); err != nil {
			return nil, fmt.Errorf("error decoding palette entry: %w", err)
		}
		palette.values[i] = uint32(temp)
	}
	return palette, nil
}

// networkPersistentEncoding implements the Chunk encoding for sending over network with persistent storage.
type networkPersistentEncoding struct{}

func (networkPersistentEncoding) network() byte { return 1 }
func (npe networkPersistentEncoding) encodePalette(buf *bytes.Buffer, p *Palette, pe paletteEncoding) {
	if p.size != 0 {
		_ = protocol.WriteVarint32(buf, int32(p.Len()))
	}

	enc := nbt.NewEncoderWithEncoding(buf, nbt.NetworkLittleEndian)
	for _, val := range p.values {
		name, props, _ := RuntimeIDToState(val)
		_ = enc.Encode(blockEntry{Name: strings.TrimPrefix("minecraft:", name), State: props, Version: CurrentBlockVersion})
	}
}
func (npe networkPersistentEncoding) decodePalette(buf *bytes.Buffer, blockSize paletteSize, pe paletteEncoding) (*Palette, error) {
	var paletteCount int32 = 1
	if blockSize != 0 {
		err := protocol.Varint32(buf, &paletteCount)
		if err != nil {
			panic(err)
		}
		if paletteCount <= 0 {
			return nil, fmt.Errorf("invalid palette entry count %v", paletteCount)
		}
	}

	blocks := make([]blockEntry, paletteCount)
	dec := nbt.NewDecoderWithEncoding(buf, nbt.NetworkLittleEndian)
	for i := int32(0); i < paletteCount; i++ {
		if err := dec.Decode(&blocks[i]); err != nil {
			return nil, fmt.Errorf("error decoding block state: %w", err)
		}
	}

	// fix for servers where the palette is encoded wrong,
	// 1 or more duplicate blocks is added to the palette which is not counted to the palette count field in the encoding
	// however is still part of the encoding
	// this adds those blocks by checking if there is still nbt data after the last blockEntry
	// and then reading blocks until it finds none left
	for {
		p, _ := buf.ReadByte()
		buf.UnreadByte()
		if p != 0x0a {
			break
		}
		block := blockEntry{}
		if err := dec.Decode(&block); err != nil {
			return nil, fmt.Errorf("error decoding block state: %w", err)
		}
		blocks = append(blocks, block)
		paletteCount++
	}

	palette := newPalette(blockSize, make([]uint32, paletteCount))
	if err := upgradePalette(blocks, palette); err != nil {
		return nil, err
	}

	return palette, nil
}

func upgradePalette(blocks []blockEntry, palette *Palette) error {
	for i, b := range blocks {
		if !strings.Contains(b.Name, ":") {
			b.Name = "minecraft:" + b.Name
		}

		// Upgrade the block state if necessary.
		upgraded := blockupgrader.Upgrade(blockupgrader.BlockState{
			Name:       b.Name,
			Properties: b.State,
			Version:    b.Version,
		})

		temp, ok := StateToRuntimeID(upgraded.Name, upgraded.Properties)
		if !ok {
			return fmt.Errorf("cannot get runtime ID of block state %v{%+v}", upgraded.Name, upgraded.Properties)
		}
		palette.values[i] = temp
	}
	return nil
}
