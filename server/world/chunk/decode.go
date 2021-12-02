package chunk

import (
	"bytes"
	"fmt"
)

// StateToRuntimeID must hold a function to convert a name and its state properties to a runtime ID.
var StateToRuntimeID func(name string, properties map[string]interface{}) (runtimeID uint32, found bool)

// NetworkDecode decodes the network serialised data passed into a Chunk if successful. If not, the chunk
// returned is nil and the error non-nil.
// The sub chunk count passed must be that found in the LevelChunk packet.
//noinspection GoUnusedExportedFunction
func NetworkDecode(air uint32, data []byte, count int) (*Chunk, error) {
	var (
		c   = New(air)
		buf = bytes.NewBuffer(data)
		err error
	)
	for i := 0; i < count; i++ {
		index := uint8(i)
		c.sub[index], err = decodeSubChunk(buf, air, &index, NetworkEncoding)
		if err != nil {
			return nil, err
		}
	}
	var last *PalettedStorage
	for i := 0; i < subChunkCount; i++ {
		b, err := decodePalettedStorage(buf, NetworkEncoding, BiomePaletteEncoding)
		if err != nil {
			return nil, err
		}
		// b == nil means this paletted storage had the flag pointing to the previous one. It basically means we should
		// inherit whatever palette we decoded last.
		if i == 0 && b == nil {
			// This should never happen and there is no way to handle this.
			return nil, fmt.Errorf("first biome storage pointed to previous one")
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
	return c, nil
}

// DiskDecode decodes the data from a SerialisedData object into a chunk and returns it. If the data was
// invalid, an error is returned.
func DiskDecode(data SerialisedData) (*Chunk, error) {
	air, ok := StateToRuntimeID("minecraft:air", nil)
	if !ok {
		panic("cannot find air runtime ID")
	}

	c := New(air)

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
		if c.sub[index], err = decodeSubChunk(bytes.NewBuffer(sub), c.air, &index, DiskEncoding); err != nil {
			return nil, err
		}
	}
	return c, nil
}

// decodeSubChunk decodes a SubChunk from a bytes.Buffer. The Encoding passed defines how the block storages of the
// SubChunk are decoded.
func decodeSubChunk(buf *bytes.Buffer, air uint32, index *byte, e Encoding) (*SubChunk, error) {
	ver, err := buf.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("error reading version: %w", err)
	}
	sub := NewSubChunk(air)
	switch ver {
	default:
		return nil, fmt.Errorf("unknown sub chunk version %v: can't decode", ver)
	case 1:
		// Version 1 only has one layer for each sub chunk, but uses the format with palettes.
		storage, err := decodePalettedStorage(buf, e, BlockPaletteEncoding)
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
				return nil, fmt.Errorf("error reading subchunk index: %w", err)
			}
			*index = uint8(int8(uIndex) - minSubChunkY)
		}
		sub.storages = make([]*PalettedStorage, storageCount)

		for i := byte(0); i < storageCount; i++ {
			sub.storages[i], err = decodePalettedStorage(buf, e, BlockPaletteEncoding)
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
		for i := 0; i < subChunkCount; i++ {
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
	blockSize >>= 1
	if blockSize == 0x7f {
		return nil, nil
	}

	size := paletteSize(blockSize)
	uint32Count := size.uint32s()

	uint32s := make([]uint32, uint32Count)
	byteCount := uint32Count * 4

	data := buf.Next(byteCount)
	if len(data) != byteCount {
		return nil, fmt.Errorf("cannot read paletted storage (size=%v) %T: not enough block data present: expected %v bytes, got %v", blockSize, pe, byteCount, len(data))
	}
	for i := 0; i < uint32Count; i++ {
		// Explicitly don't use the binary package to greatly improve performance of reading the uint32s.
		uint32s[i] = uint32(data[i*4]) | uint32(data[i*4+1])<<8 | uint32(data[i*4+2])<<16 | uint32(data[i*4+3])<<24
	}
	p, err := e.decodePalette(buf, paletteSize(blockSize), pe)
	return newPalettedStorage(uint32s, p), err
}
