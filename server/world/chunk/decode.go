package chunk

import (
	"bytes"
	"fmt"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/sandertv/gophertunnel/minecraft/nbt"
)

// StateToRuntimeID must hold a function to convert a name and its state properties to a runtime ID.
var StateToRuntimeID func(name string, properties map[string]interface{}) (runtimeID uint32, found bool)

// NetworkDecode decodes the network serialised data passed into a Chunk if successful. If not, the chunk
// returned is nil and the error non-nil.
// The sub chunk count passed must be that found in the LevelChunk packet.
//noinspection GoUnusedExportedFunction
func NetworkDecode(air uint32, data []byte, subChunkCount int) (*Chunk, error) {
	var (
		c   = New(air)
		buf = bytes.NewBuffer(data)
		err error
	)
	for y := 0; y < subChunkCount; y++ {
		index := uint8(y)
		c.sub[index], err = decodeSubChunk(buf, air, &index, NetworkEncoding)
		if err != nil {
			return nil, err
		}
	}
	if _, err := buf.Read(c.biomes[:]); err != nil {
		return nil, fmt.Errorf("error reading biomes: %w", err)
	}
	_, _ = buf.ReadByte()

	dec := nbt.NewDecoder(buf)
	for buf.Len() != 0 {
		var m map[string]interface{}
		if err := dec.Decode(&m); err != nil {
			return nil, fmt.Errorf("error decoding block entity: %w", err)
		}
		c.SetBlockNBT(cube.Pos{int(m["x"].(int32)), int(m["y"].(int32)), int(m["z"].(int32))}, m)
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

	var (
		c   = New(air)
		err error
	)
	if len(data.Data2D) >= 512+256 {
		copy(c.biomes[:], data.Data2D[512:])
	}

	for y, sub := range data.SubChunks {
		if len(sub) == 0 {
			// No data for this sub chunk.
			continue
		}
		index := uint8(y)
		c.sub[index], err = decodeSubChunk(bytes.NewBuffer(sub), air, &index, DiskEncoding)
		if err != nil {
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
		storage, err := decodeBlockStorage(buf, e)
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
			*index, err = buf.ReadByte()
			if err != nil {
				return nil, fmt.Errorf("error reading subchunk index: %w", err)
			}
		}
		sub.storages = make([]*BlockStorage, storageCount)

		for i := byte(0); i < storageCount; i++ {
			sub.storages[i], err = decodeBlockStorage(buf, e)
			if err != nil {
				return nil, err
			}
		}
	}
	return sub, nil
}

// decodeBlockStorage decodes a block storage from a bytes.Buffer. The Encoding passed is used to read either a network
// or disk block storage.
func decodeBlockStorage(buf *bytes.Buffer, e Encoding) (*BlockStorage, error) {
	blockSize, err := buf.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("error reading block size: %w", err)
	}
	blockSize >>= 1

	// blocksPerUint32 is the amount of blocks that may be stored in a single uint32.
	blocksPerUint32 := 32 / int(blockSize)

	// uint32Count is the amount of uint32s required to store all blocks: 4096 blocks need to be stored in
	// total.
	uint32Count := 4096 / blocksPerUint32

	if paletteSize(blockSize).padded() {
		// We've got one of the padded sizes, so the block storage has another uint32 to be able to store
		// every block.
		uint32Count++
	}

	uint32s := make([]uint32, uint32Count)
	byteCount := uint32Count * 4

	data := buf.Next(byteCount)
	if len(data) != byteCount {
		return nil, fmt.Errorf("cannot read block storage: not enough block data present: expected %v bytes, got %v", uint32Count*4, len(data))
	}
	for i := 0; i < uint32Count; i++ {
		// Explicitly don't use the binary package to greatly improve performance of reading the uint32s.
		uint32s[i] = uint32(data[i*4]) | uint32(data[i*4+1])<<8 | uint32(data[i*4+2])<<16 | uint32(data[i*4+3])<<24
	}
	p, err := e.decodePalette(buf, paletteSize(blockSize))
	return newBlockStorage(uint32s, p), err
}
