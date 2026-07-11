package chunk

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/sandertv/gophertunnel/minecraft/nbt"
)

// NetworkDecode decodes the network serialised data passed into a Chunk if successful. If not, the chunk
// returned is nil and the error non-nil.
// The sub chunk count passed must be that found in the LevelChunk packet.
// NetworkDecode creates a new buffer and calls NetworkDecodeBuffer.
//
// The BlockRegistry passed must be finalized and must correspond to the runtime IDs used in the chunk data.
// noinspection GoUnusedExportedFunction
func NetworkDecode(br BlockRegistry, data []byte, count int, r cube.Range) (*Chunk, error) {
	return NetworkDecodeBuffer(br, bytes.NewBuffer(data), count, r)
}

// NetworkDecodeBuffer decodes the network serialised data from buf passed into a Chunk if successful. If not, the chunk
// returned is nil and the error non-nil.
// The sub chunk count passed must be that found in the LevelChunk packet.
// noinspection GoUnusedExportedFunction
func NetworkDecodeBuffer(br BlockRegistry, buf *bytes.Buffer, count int, r cube.Range) (*Chunk, error) {
	c := New(br, r)
	if count < 0 || count > len(c.sub) {
		return nil, fmt.Errorf("invalid sub-chunk count %d: chunk range has %d sub-chunks", count, len(c.sub))
	}
	for i := 0; i < count; i++ {
		index := uint8(i)
		sub, err := decodeSubChunk(buf, c, &index, NetworkEncoding)
		if err != nil {
			return nil, err
		}
		if int(index) >= len(c.sub) {
			return nil, fmt.Errorf("invalid sub-chunk index %d: chunk range has %d sub-chunks", index, len(c.sub))
		}
		c.sub[index] = sub
	}
	var last *PalettedStorage
	for i := 0; i < len(c.sub); i++ {
		b, err := decodePalettedStorage(buf, NetworkEncoding, BiomePaletteEncoding)
		if err != nil {
			// Some non-conformant servers encode the biome sections wrong (e.g. an unsigned palette
			// entry count instead of the spec's signed zigzag VarInt, or fewer sections than the
			// dimension height). Biomes are cosmetic, so keep the successfully decoded block
			// sub-chunks and leave the remaining biomes at their defaults rather than dropping an
			// otherwise-valid chunk. Discard the rest of the payload so the trailing border-block
			// and block-entity reads don't parse the misaligned tail as garbage.
			buf.Next(buf.Len())
			return c, nil
		}
		if b == nil {
			// b == nil means this paletted storage had the flag pointing to the previous one. It basically means we should
			// inherit whatever palette we decoded last.
			if i == 0 {
				// This should never happen and there is no way to handle this.
				return nil, fmt.Errorf("first biome storage pointed to previous one")
			}
			b = last
		} else {
			last = b
		}
		c.biomes[i] = b
	}
	return c, nil
}

// NetworkDecodeWithBlockNBTs decodes a network serialised Chunk and any trailing block entity NBT data.
// The sub chunk count passed must be that found in the LevelChunk packet.
// noinspection GoUnusedExportedFunction
func NetworkDecodeWithBlockNBTs(br BlockRegistry, data []byte, count int, r cube.Range) (*Chunk, []map[string]any, error) {
	return NetworkDecodeBufferWithBlockNBTs(br, bytes.NewBuffer(data), count, r)
}

// NetworkDecodeBufferWithBlockNBTs decodes a network serialised Chunk and any trailing block entity NBT data.
// The sub chunk count passed must be that found in the LevelChunk packet.
// noinspection GoUnusedExportedFunction
func NetworkDecodeBufferWithBlockNBTs(br BlockRegistry, buf *bytes.Buffer, count int, r cube.Range) (*Chunk, []map[string]any, error) {
	c, err := NetworkDecodeBuffer(br, buf, count, r)
	if err != nil {
		return nil, nil, err
	}

	// The LevelChunk payload may include extra data right after biomes. If there are no remaining bytes,
	// there are no extras to decode.
	borderBlocks, err := buf.ReadByte()
	if err != nil {
		// bytes.Buffer only errors on ReadByte when empty: treat that as "no extras".
		if buf.Len() == 0 {
			return c, nil, nil
		}
		return nil, nil, fmt.Errorf("error reading border blocks byte: %w", err)
	}
	if borderBlocks > 0 {
		skipped := buf.Next(int(borderBlocks))
		if len(skipped) != int(borderBlocks) {
			return nil, nil, fmt.Errorf("not enough border blocks data present: expected %d bytes, got %d", borderBlocks, len(skipped))
		}
	}

	blockNBTs, err := DecodeBlockNBTs(buf)
	if err != nil {
		return nil, nil, err
	}
	return c, blockNBTs, nil
}

// NetworkDecodeWithBlockEntities decodes a network serialised Chunk and any trailing block entities, returning them in
// the canonical chunk.BlockEntity type.
// The sub chunk count passed must be that found in the LevelChunk packet.
// noinspection GoUnusedExportedFunction
func NetworkDecodeWithBlockEntities(br BlockRegistry, data []byte, count int, r cube.Range) (*Chunk, []BlockEntity, error) {
	return NetworkDecodeBufferWithBlockEntities(br, bytes.NewBuffer(data), count, r)
}

// NetworkDecodeBufferWithBlockEntities decodes a network serialised Chunk and any trailing block entities, returning
// them in the canonical chunk.BlockEntity type.
// The sub chunk count passed must be that found in the LevelChunk packet.
// noinspection GoUnusedExportedFunction
func NetworkDecodeBufferWithBlockEntities(br BlockRegistry, buf *bytes.Buffer, count int, r cube.Range) (*Chunk, []BlockEntity, error) {
	c, blockNBTs, err := NetworkDecodeBufferWithBlockNBTs(br, buf, count, r)
	if err != nil {
		return nil, nil, err
	}
	blockEntities := make([]BlockEntity, 0, len(blockNBTs))
	for _, blockNBT := range blockNBTs {
		x, okX := blockNBT["x"].(int32)
		y, okY := blockNBT["y"].(int32)
		z, okZ := blockNBT["z"].(int32)
		// If x/y/z are missing (or have an unexpected type), keep the entry out: it can't be indexed into the world.
		if !okX || !okY || !okZ {
			continue
		}
		blockEntities = append(blockEntities, BlockEntity{
			Pos:  cube.Pos{int(x), int(y), int(z)},
			Data: blockNBT,
		})
	}
	return c, blockEntities, nil
}

// DecodeBlockNBTs decodes a list of NBT compounds from buf until it is fully consumed.
// noinspection GoUnusedExportedFunction
func DecodeBlockNBTs(buf *bytes.Buffer) ([]map[string]any, error) {
	if buf.Len() == 0 {
		return nil, nil
	}

	var blockNBTs []map[string]any

	dec := nbt.NewDecoderWithEncoding(buf, nbt.NetworkLittleEndian)
	dec.AllowZero = true
	for buf.Len() > 0 {
		blockNBT := make(map[string]any)
		if err := dec.Decode(&blockNBT); err != nil {
			return nil, err
		}
		if len(blockNBT) > 0 {
			blockNBTs = append(blockNBTs, blockNBT)
		}
	}

	return blockNBTs, nil
}

// DiskDecode decodes the data from a SerialisedData object into a chunk and returns it. If the data was invalid,
// an error is returned.
//
// The BlockRegistry passed must be finalized and must correspond to the runtime IDs used in the chunk data.
func DiskDecode(br BlockRegistry, data SerialisedData, r cube.Range) (*Chunk, error) {
	c := New(br, r)

	err := decodeBiomes(bytes.NewBuffer(data.Biomes), c, DiskEncoding)
	if err != nil {
		return nil, err
	}
	for i, sub := range data.SubChunks {
		if len(sub) == 0 {
			// No data for this sub chunk.
			continue
		}
		index := uint8(i)
		if c.sub[index], err = decodeSubChunk(bytes.NewBuffer(sub), c, &index, DiskEncoding); err != nil {
			return nil, err
		}
	}
	return c, nil
}

// decodeSubChunk decodes a SubChunk from a bytes.Buffer. The Encoding passed defines how the block storages of the
// SubChunk are decoded.
func decodeSubChunk(buf *bytes.Buffer, c *Chunk, index *byte, e Encoding) (*SubChunk, error) {
	ver, err := buf.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("error reading version: %w", err)
	}
	sub := NewSubChunk(c.air)
	switch ver {
	default:
		return nil, fmt.Errorf("unknown sub chunk version %v: can't decode", ver)
	case 1:
		// Version 1 only has one layer for each sub chunk, but uses the format with palettes.
		storage, err := decodePalettedStorage(buf, e, BlockPaletteEncoding{Blocks: c.br})
		if err != nil {
			return nil, err
		}
		sub.storages = append(sub.storages, storage)
	case 8, 9:
		// Version 8 allows up to 256 layers for one sub chunk.
		storageCount, err := buf.ReadByte()
		if err != nil {
			return nil, fmt.Errorf("error reading storage count: %w", err)
		}
		if ver == 9 {
			uIndex, err := buf.ReadByte()
			if err != nil {
				return nil, fmt.Errorf("error reading sub-chunk index: %w", err)
			}
			// The index as written here isn't the actual index of the sub-chunk within the chunk. Rather, it is the Y
			// value of the sub-chunk. This means that we need to translate it to an index.
			*index = uint8(int8(uIndex) - int8(c.r[0]>>4))
		}
		sub.storages = make([]*PalettedStorage, storageCount)

		for i := byte(0); i < storageCount; i++ {
			sub.storages[i], err = decodePalettedStorage(buf, e, BlockPaletteEncoding{Blocks: c.br})
			if err != nil {
				return nil, err
			}
		}
	}
	return sub, nil
}

// decodeBiomes reads the paletted storages holding biomes from buf and stores it into the Chunk passed.
func decodeBiomes(buf *bytes.Buffer, c *Chunk, e Encoding) error {
	var last *PalettedStorage
	if buf.Len() != 0 {
		for i := 0; i < len(c.sub); i++ {
			b, err := decodePalettedStorage(buf, e, BiomePaletteEncoding)
			if err != nil {
				return err
			}
			// b == nil means this paletted storage had the flag pointing to the previous one. It basically means we should
			// inherit whatever palette we decoded last.
			if i == 0 && b == nil {
				// This should never happen and there is no way to handle this.
				return fmt.Errorf("first biome storage pointed to previous one")
			}
			if b == nil {
				// This means this paletted storage had the flag pointing to the previous one. It basically means we should
				// inherit whatever palette we decoded last.
				b = last
			} else {
				last = b
			}
			c.biomes[i] = b
		}
	}
	return nil
}

// decodePalettedStorage decodes a PalettedStorage from a bytes.Buffer. The Encoding passed is used to read either a
// network or disk block storage.
func decodePalettedStorage(buf *bytes.Buffer, e Encoding, pe paletteEncoding) (*PalettedStorage, error) {
	blockSize, err := buf.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("error reading block size: %w", err)
	}
	_, isNetwork := e.(networkEncoding)
	_, isBlocks := pe.(BlockPaletteEncoding)
	if isNetwork && isBlocks && blockSize&1 != 1 {
		e = NetworkPersistentEncoding
	}

	blockSize >>= 1
	if blockSize == 0x7f {
		return nil, nil
	}

	size := paletteSize(blockSize)
	if size > 32 {
		return nil, fmt.Errorf("cannot read paletted storage (size=%v) %T: size too large", blockSize, pe)
	}
	uint32Count := size.uint32s()
	byteCount := uint32Count * 4

	data := buf.Next(byteCount)
	if len(data) != byteCount {
		return nil, fmt.Errorf("cannot read paletted storage (size=%v) %T: not enough block data present: expected %v bytes, got %v", blockSize, pe, byteCount, len(data))
	}
	if _, isPersistent := e.(networkPersistentEncoding); !isPersistent {
		if paletteCount, ok := peekPaletteCount(buf, size, e); ok && paletteCount == 1 {
			p, err := e.decodePalette(buf, size, pe)
			if err != nil {
				return nil, err
			}
			// Some servers encode a single-value palette using non-zero bits per block. Canonicalise it to a 0-bit
			// storage and skip allocating index words.
			p.size = 0
			return newPalettedStorage(nil, p), nil
		}
	}

	uint32s := make([]uint32, uint32Count)
	for i, j := 0, 0; i < uint32Count; i, j = i+1, j+4 {
		// Explicitly don't use the binary package to greatly improve performance of reading the uint32s.
		uint32s[i] = uint32(data[j]) | uint32(data[j+1])<<8 | uint32(data[j+2])<<16 | uint32(data[j+3])<<24
	}
	p, err := e.decodePalette(buf, size, pe)
	if err != nil {
		return nil, err
	}
	return newPalettedStorage(uint32s, p), nil
}

// peekPaletteCount peeks the amount of palette entries that follow in buf for this encoding and block size without
// consuming the buffer.
func peekPaletteCount(buf *bytes.Buffer, size paletteSize, e Encoding) (int32, bool) {
	if size == 0 {
		return 1, true
	}

	switch e.(type) {
	case diskEncoding:
		b := buf.Bytes()
		if len(b) < 4 {
			return 0, false
		}
		return int32(binary.LittleEndian.Uint32(b[:4])), true
	case networkEncoding:
		return peekVarint32(buf.Bytes())
	default:
		return 0, false
	}
}

func peekVarint32(b []byte) (int32, bool) {
	var v uint32
	for i := uint(0); i < 35; i += 7 {
		index := int(i / 7)
		if index >= len(b) {
			return 0, false
		}
		c := b[index]
		v |= uint32(c&0x7f) << i
		if c&0x80 == 0 {
			x := int32(v >> 1)
			if v&1 != 0 {
				x = ^x
			}
			return x, true
		}
	}
	return 0, false
}
