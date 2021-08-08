package chunk

import (
	"bytes"
	"fmt"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/sandertv/gophertunnel/minecraft/nbt"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// StateToRuntimeID must hold a function to convert a name and its state properties to a runtime ID.
var StateToRuntimeID func(name string, properties map[string]interface{}) (runtimeID uint32, found bool)

// NetworkDecode decodes the network serialised data passed into a Chunk if successful. If not, the chunk
// returned is nil and the error non-nil.
// The sub chunk count passed must be that found in the LevelChunk packet.
//noinspection GoUnusedExportedFunction
func NetworkDecode(airRuntimeId uint32, data []byte, subChunkCount int) (*Chunk, error) {
	c, buf := New(airRuntimeId), bytes.NewBuffer(data)
	for y := 0; y < subChunkCount; y++ {
		ver, err := buf.ReadByte()
		if err != nil {
			return nil, fmt.Errorf("error reading version: %w", err)
		}
		c.sub[y] = NewSubChunk(airRuntimeId)
		switch ver {
		default:
			return nil, fmt.Errorf("unknown sub chunk version %v: can't decode", ver)
		case 1:
			// Version 1 only has one layer for each sub chunk, but uses the format with palettes.
			storage, err := networkDecodeBlockStorage(buf)
			if err != nil {
				return nil, err
			}
			c.sub[y].storages = append(c.sub[y].storages, storage)
		case 8:
			// Version 8 allows up to 256 layers for one sub chunk.
			storageCount, err := buf.ReadByte()
			if err != nil {
				return nil, fmt.Errorf("error reading storage count: %w", err)
			}
			c.sub[y].storages = make([]*BlockStorage, storageCount)

			for i := byte(0); i < storageCount; i++ {
				c.sub[y].storages[i], err = networkDecodeBlockStorage(buf)
				if err != nil {
					return nil, err
				}
			}
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

	c := New(air)
	copy(c.biomes[:], data.Data2D[512:])

	for y, sub := range data.SubChunks {
		if len(sub) == 0 {
			// No data for this sub chunk.
			continue
		}
		buf := bytes.NewBuffer(sub)
		ver, err := buf.ReadByte()
		if err != nil {
			return nil, fmt.Errorf("error reading version: %w", err)
		}
		c.sub[y] = NewSubChunk(air)
		switch ver {
		default:
			return nil, fmt.Errorf("unknown sub chunk version %v: can't decode", ver)
		case 1:
			// Version 1 only has one layer for each sub chunk, but uses the format with palettes.
			storage, err := diskDecodeBlockStorage(buf)
			if err != nil {
				return nil, err
			}
			c.sub[y].storages = append(c.sub[y].storages, storage)
		case 8:
			// Version 8 allows up to 256 layers for one sub chunk.
			storageCount, err := buf.ReadByte()
			if err != nil {
				return nil, fmt.Errorf("error reading storage count: %w", err)
			}
			c.sub[y].storages = make([]*BlockStorage, storageCount)

			for i := byte(0); i < storageCount; i++ {
				c.sub[y].storages[i], err = diskDecodeBlockStorage(buf)
				if err != nil {
					return nil, err
				}
			}
		}
	}
	return c, nil
}

// networkDecodeBlockStorage decodes a block storage from the buffer passed, assuming it holds data for a
// network encoded block storage, and returns it if successful.
func networkDecodeBlockStorage(buf *bytes.Buffer) (*BlockStorage, error) {
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
	data := buf.Next(uint32Count * 4)
	if len(data) != uint32Count*4 {
		return nil, fmt.Errorf("cannot read block storage: not enough block data present: expected %v bytes, got %v", uint32Count*4, len(data))
	}
	for i := 0; i < uint32Count; i++ {
		// Explicitly don't use the binary package to greatly improve performance of reading the uint32s.
		uint32s[i] = uint32(data[i*4]) | uint32(data[i*4+1])<<8 | uint32(data[i*4+2])<<16 | uint32(data[i*4+3])<<24
	}

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
	return newBlockStorage(uint32s, &Palette{blockRuntimeIDs: blocks, size: paletteSize(blockSize)}), nil
}

// diskDecodeBlockStorage decodes a block storage from the buffer passed. If not successful, an error is
// returned.
func diskDecodeBlockStorage(buf *bytes.Buffer) (*BlockStorage, error) {
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

	data := buf.Next(uint32Count * 4)
	if len(data) != uint32Count*4 {
		return nil, fmt.Errorf("cannot read block storage: not enough block data present: expected %v bytes, got %v", uint32Count*4, len(data))
	}

	for i := 0; i < uint32Count; i++ {
		// Explicitly don't use the binary package to greatly improve performance of reading the uint32s.
		uint32s[i] = uint32(data[i*4]) | uint32(data[i*4+1])<<8 | uint32(data[i*4+2])<<16 | uint32(data[i*4+3])<<24
	}

	// The next 4 bytes are an LE int32, but we simply read it and decode the int32 ourselves, as it's much
	// faster here.
	data = buf.Next(4)
	if len(data) != 4 {
		return nil, fmt.Errorf("cannot read palette entry count: expected 4 bytes, got %v", len(data))
	}
	paletteCount := uint32(data[0]) | uint32(data[1])<<8 | uint32(data[2])<<16 | uint32(data[3])<<24

	blocks := make([]blockEntry, paletteCount)

	dec := nbt.NewDecoderWithEncoding(buf, nbt.LittleEndian)

	// There are paletteCount NBT tags that represent unique blocks.
	for i := uint32(0); i < paletteCount; i++ {
		if err := dec.Decode(&blocks[i]); err != nil {
			return nil, fmt.Errorf("error decoding block: %w", err)
		}
	}

	palette := newPalette(paletteSize(blockSize), make([]uint32, paletteCount))
	for i, b := range blocks {
		var ok bool
		palette.blockRuntimeIDs[i], ok = StateToRuntimeID(b.Name, b.State)
		if !ok {
			return nil, fmt.Errorf("cannot get runtime ID of block state %v{%+v}", b.Name, b.State)
		}
	}
	return newBlockStorage(uint32s, palette), nil
}
